// Itero - Online iterative vote application
// Copyright (C) 2020 Joseph Boudou
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package alarm

import (
	"testing"
	"time"
)

const AllowedInterval = time.Millisecond

func TestAlarm(t *testing.T) {
	tests := []struct {
		name      string
		durations []string
	}{
		{
			name:      "One",
			durations: []string{ "5ms" },
		},
		{
			name:      "Two",
			durations: []string{ "5ms", "10ms" },
		},
		{
			name:      "One in the past",
			durations: []string{ "5ms", "4ms900µs", "10ms" },
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			durations := make([]time.Duration, len(tt.durations))
			for i, str := range tt.durations {
				var err error
				durations[i], err = time.ParseDuration(str)
				if err != nil {
					t.Fatal(err)
				}
			}

			times := make([]time.Time, len(durations))
			now := time.Now()
			for i, dur := range durations {
				times[i] = now.Add(dur)
			}

			alarm := New(len(durations))
			for _, t := range times {
				alarm.Send <- Event{ Time: t }
			}
			close(alarm.Send)

			cur := 0
			for true {
				evt, ok := <-alarm.Receive
				if !ok {
					break
				}
				diff := time.Since(evt.Time)

				if !evt.Time.Equal(times[cur]) {
					t.Errorf("Wrong time. Got %v. Expect %v.", evt.Time, times[cur])
					continue
				}
				cur++
				if diff < 0 {
					t.Errorf("Received in the past (%v).", diff)
				}
				if diff > AllowedInterval {
					t.Errorf("Received too late (%d).", diff)
				}
			}

			if cur < len(times) {
				t.Errorf("Missing %d events.", len(times) - cur)
			}
		})
	}
}
