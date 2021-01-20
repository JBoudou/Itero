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
	"sort"
	"testing"
	"time"
)

// AllowedInterval is the maximal authorized delay for events.
// Since the package intentionally does not give any guarantee of that delay,
// it is set to a quite long duration.
const AllowedInterval = 1 * time.Second

func TestAlarm(t *testing.T) {
	tests := []struct {
		name      string
		durations []string
	}{
		{
			name:      "One",
			durations: []string{ "10ms" },
		},
		{
			name:      "Two",
			durations: []string{ "10ms", "20ms" },
		},
		{
			name:      "One in the past",
			durations: []string{ "10ms", "5ms", "20ms" },
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

			alarm := New(len(durations))

			times := make([]time.Time, len(durations))
			now := time.Now().Add(10 * time.Millisecond)
			for i, dur := range durations {
				times[i] = now.Add(dur)
			}

			for _, t := range times {
				alarm.Send <- Event{ Time: t }
			}
			close(alarm.Send)

			sort.Slice(times, func (i, j int) bool {
				return durations[i] < durations[j];
			})
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
				if diff < 0 {
					t.Errorf("Received event %d too early (%v).", cur, diff)
				}
				if diff > AllowedInterval {
					t.Errorf("Received event %d too late (%v).", cur, diff)
				}
				cur++
			}

			if cur < len(times) {
				t.Errorf("Missing %d events.", len(times) - cur)
			}
		})
	}
}

func TestAlarm_DiscardLaterEvent(t *testing.T) {
	tests := []struct {
		name      string
		durations []string
	}{
		{
			name:      "One",
			durations: []string{ "10ms" },
		},
		{
			name:      "Two",
			durations: []string{ "10ms", "20ms" },
		},
		{
			name:      "One in the past",
			durations: []string{ "10ms", "5ms", "20ms" },
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

			alarm := New(0, DiscardLaterEvent)

			times := make([]time.Time, len(durations))
			now := time.Now().Add(100 * time.Millisecond)
			for i, dur := range durations {
				times[i] = now.Add(dur)
			}

			for _, t := range times {
				alarm.Send <- Event{ Time: t }
			}
			close(alarm.Send)

			filtered := make([]time.Time, 0, len(times))
			first := now.Add(time.Hour)
			for _, t := range times {
				if t.Before(first) {
					filtered = append(filtered, t)
					first = t
				}
			}

			sort.Slice(filtered, func (i, j int) bool {
				return durations[i] < durations[j];
			})
			cur := 0
			for true {
				evt, ok := <-alarm.Receive
				if !ok {
					break
				}
				diff := time.Since(evt.Time)

				if !evt.Time.Equal(filtered[cur]) {
					t.Errorf("Wrong time. Got %v. Expect %v.", evt.Time, filtered[cur])
					continue
				}
				if diff < 0 {
					t.Errorf("Received event %d too early (%v).", cur, diff)
				}
				if diff > AllowedInterval {
					t.Errorf("Received event %d too late (%v).", cur, diff)
				}
				cur++
			}

			if cur < len(filtered) {
				t.Errorf("Missing %d events.", len(filtered) - cur)
			}
		})
	}
}

func TestAlarm_Remaining(t *testing.T) {
	alarm := New(0)

	const (
		durationStep = 2 * time.Millisecond
		nbEvents = 10
	)

	for i, wait := 0, 100 * time.Millisecond; i < nbEvents; i, wait = i + 1, wait + durationStep {
		alarm.Send <- Event{ Time: time.Now().Add(wait) }
	}

	for i := nbEvents; i > 0; i -= 1 {
		evt, ok := <-alarm.Receive
		if !ok {
			t.Errorf("Missing %d events", i)
			break
		}
		if (evt.Remaining != i - 1) {
			t.Errorf("Wrong Remaining. Got %d. Expect %d.", evt.Remaining, i - 1)
		}
	}
}
