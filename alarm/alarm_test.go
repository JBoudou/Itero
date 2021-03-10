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
			durations: []string{"100ms"},
		},
		{
			name:      "Two",
			durations: []string{"100ms", "200ms"},
		},
		{
			name:      "One in the past",
			durations: []string{"200ms", "100ms", "300ms"},
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
			now := time.Now()
			for i, dur := range durations {
				times[i] = now.Add(dur)
			}

			for _, t := range times {
				alarm.Send <- Event{Time: t}
			}
			close(alarm.Send)

			sort.Slice(times, func(i, j int) bool {
				return durations[i] < durations[j]
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
				t.Errorf("Missing %d events.", len(times)-cur)
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
			durations: []string{"100ms"},
		},
		{
			name:      "Two",
			durations: []string{"100ms", "200ms"},
		},
		{
			name:      "One in the past",
			durations: []string{"200ms", "100ms", "300ms"},
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
			now := time.Now()
			for i, dur := range durations {
				times[i] = now.Add(dur)
			}

			for _, t := range times {
				alarm.Send <- Event{Time: t}
			}
			close(alarm.Send)

			filtered := make([]time.Time, 0, len(times))
			previous := now.Add(time.Hour)
			for _, t := range times {
				if t.Before(previous) {
					filtered = append(filtered, t)
					previous = t
				}
			}

			sort.Slice(filtered, func(i, j int) bool {
				return durations[i] < durations[j]
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
				t.Errorf("Missing %d events.", len(filtered)-cur)
			}
		})
	}
}

func TestAlarm_DiscardDuplicates(t *testing.T) {
	tests := []struct {
		name      string
		events [][2]string
		expect    []string
	}{
		{
			name:      "One",
			events: [][2]string{{"0", "100ms"}},
			expect:    []string{"0"},
		},
		{
			name:      "Two",
			events: [][2]string{{"0", "200ms"}, {"1", "100ms"}},
			expect:    []string{"1", "0"},
		},
		{
			name:      "Duplicate",
			events: [][2]string{{"0", "300ms"}, {"1", "200ms"}, {"0", "100ms"}},
			expect:    []string{"1", "0"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			durations := make([]time.Duration, len(tt.events))
			for i, evt := range tt.events {
				var err error
				durations[i], err = time.ParseDuration(evt[1])
				if err != nil {
					t.Fatal(err)
				}
			}

			alarm := New(0, DiscardDuplicates)

			times := make([]time.Time, len(durations))
			now := time.Now()
			for i, dur := range durations {
				times[i] = now.Add(dur)
			}

			for i, t := range times {
				alarm.Send <- Event{Time: t, Data: tt.events[i][0]}
			}
			close(alarm.Send)

			cur := 0
			for true {
				evt, ok := <-alarm.Receive
				if !ok {
					break
				}
				diff := time.Since(evt.Time)

				if evt.Data.(string) != tt.expect[cur] {
					t.Errorf("Wrong data. Got %s. Expect %s.", evt.Data.(string), tt.expect[cur])
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

			if cur < len(tt.expect) {
				t.Errorf("Missing %d events.", len(tt.expect)-cur)
			}
		})
	}
}

func TestAlarm_Remaining(t *testing.T) {
	alarm := New(0)

	const (
		durationStep = 2 * time.Millisecond
		nbEvents     = 10
	)

	for i, wait := 0, 100*time.Millisecond; i < nbEvents; i, wait = i+1, wait+durationStep {
		alarm.Send <- Event{Time: time.Now().Add(wait)}
	}

	for i := nbEvents; i > 0; i -= 1 {
		evt, ok := <-alarm.Receive
		if !ok {
			t.Errorf("Missing %d events", i)
			break
		}
		if evt.Remaining != i-1 {
			t.Errorf("Wrong Remaining. Got %d. Expect %d.", evt.Remaining, i-1)
		}
	}
}
