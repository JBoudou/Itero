// Itero - Online iterative vote application
// Copyright (C) 2021 Joseph Boudou
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

package slog

import (
	"log"
	"os"
	"reflect"
	"testing"
)

type printRecorder struct {
	records [][]interface{}
}

func (self *printRecorder) Println(args ...interface{}) {
	self.records = append(self.records, args)
}

func (self *printRecorder) Log(args ...interface{}) {
	self.records = append(self.records, args)
}

func (self *printRecorder) Error(args ...interface{}) {
	self.records = append(self.records, args)
}

func TestSimpleLogger(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		action func(log *SimpleLogger)
		expect [][]interface{}
	}{
		{
			name:   "nothing",
			action: func(log *SimpleLogger) {},
		},
		{
			name:   "Log",
			action: func(log *SimpleLogger) { log.Log(1, "2", 3) },
			expect: [][]interface{}{{1, "2", 3}},
		},
		{
			name:   "Logf",
			action: func(log *SimpleLogger) { log.Logf("%d %s", 1, "2") },
			expect: [][]interface{}{{"1 2"}},
		},
		{
			name: "Push Log",
			action: func(log *SimpleLogger) {
				log.Push(1, 2)
				log.Log(3, 4)
			},
			expect: [][]interface{}{{1, 2, 3, 4}},
		},
		{
			name: "Push Logf",
			action: func(log *SimpleLogger) {
				log.Push(1, 2)
				log.Logf("%d %s", 3, "4")
			},
			expect: [][]interface{}{{1, 2, "3 4"}},
		},
		{
			name: "With Log",
			action: func(log *SimpleLogger) {
				log.Push(1, 2)
				w := log.With(3, 4)
				w.Log(5, 6)
				log.Log(7, 8)
			},
			expect: [][]interface{}{{1, 2, 3, 4, 5, 6}, {1, 2, 7, 8}},
		},
		{
			name: "With Logf",
			action: func(log *SimpleLogger) {
				log.Push(1, 2)
				w := log.With(3, 4)
				w.Logf("%d-%d", 5, 6)
				log.Logf("%d-%d", 7, 8)
			},
			expect: [][]interface{}{{1, 2, 3, 4, "5-6"}, {1, 2, "7-8"}},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			recorder := printRecorder{}
			tt.action(&SimpleLogger{Printer: &recorder})

			if !reflect.DeepEqual(recorder.records, tt.expect) {
				t.Errorf("Got %v. Expect %v.", recorder.records, tt.expect)
			}
		})
	}
}

func TestStackedLeveled(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		action func(log StackedLeveled)
		expect [][]interface{}
	}{
		{
			name:   "Log",
			action: func(log StackedLeveled) { log.Log(1, "2", 3) },
			expect: [][]interface{}{{1, "2", 3}},
		},
		{
			name:   "Logf",
			action: func(log StackedLeveled) { log.Logf("%d %s", 1, "2") },
			expect: [][]interface{}{{"1 2"}},
		},
		{
			name:   "Error",
			action: func(log StackedLeveled) { log.Error(1, "2", 3) },
			expect: [][]interface{}{{1, "2", 3}},
		},
		{
			name:   "Errorf",
			action: func(log StackedLeveled) { log.Errorf("%d %s", 1, "2") },
			expect: [][]interface{}{{"1 2"}},
		},
		{
			name: "Push Log",
			action: func(log StackedLeveled) {
				log.Push(1, 2)
				log.Log(3, 4)
			},
			expect: [][]interface{}{{1, 2, 3, 4}},
		},
		{
			name: "Push Logf",
			action: func(log StackedLeveled) {
				log.Push(1, 2)
				log.Logf("%d %s", 3, "4")
			},
			expect: [][]interface{}{{1, 2, "3 4"}},
		},
		{
			name: "Push Error",
			action: func(log StackedLeveled) {
				log.Push(1, 2)
				log.Error(3, 4)
			},
			expect: [][]interface{}{{1, 2, 3, 4}},
		},
		{
			name: "Push Errorf",
			action: func(log StackedLeveled) {
				log.Push(1, 2)
				log.Errorf("%d %s", 3, "4")
			},
			expect: [][]interface{}{{1, 2, "3 4"}},
		},
		{
			name: "With Log",
			action: func(log StackedLeveled) {
				log.Push(1, 2)
				w := log.With(3, 4)
				w.Log(5, 6)
				log.Log(7, 8)
			},
			expect: [][]interface{}{{1, 2, 3, 4, 5, 6}, {1, 2, 7, 8}},
		},
		{
			name: "With Logf",
			action: func(log StackedLeveled) {
				log.Push(1, 2)
				w := log.With(3, 4)
				w.Logf("%d-%d", 5, 6)
				log.Logf("%d-%d", 7, 8)
			},
			expect: [][]interface{}{{1, 2, 3, 4, "5-6"}, {1, 2, "7-8"}},
		},
		{
			name: "With Error",
			action: func(log StackedLeveled) {
				log.Push(1, 2)
				w := log.With(3, 4)
				w.Error(5, 6)
				log.Error(7, 8)
			},
			expect: [][]interface{}{{1, 2, 3, 4, 5, 6}, {1, 2, 7, 8}},
		},
		{
			name: "With Errorf",
			action: func(log StackedLeveled) {
				log.Push(1, 2)
				w := log.With(3, 4)
				w.Errorf("%d-%d", 5, 6)
				log.Errorf("%d-%d", 7, 8)
			},
			expect: [][]interface{}{{1, 2, 3, 4, "5-6"}, {1, 2, "7-8"}},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			recorder := printRecorder{}
			tt.action(&SimpleLeveled{Printer: &recorder})

			if !reflect.DeepEqual(recorder.records, tt.expect) {
				t.Errorf("Got %v. Expect %v.", recorder.records, tt.expect)
			}
		})
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			recorder := printRecorder{}
			tt.action(&WithStack{Target: &recorder})

			if !reflect.DeepEqual(recorder.records, tt.expect) {
				t.Errorf("Got %v. Expect %v.", recorder.records, tt.expect)
			}
		})
	}
}

// Examples

func Example() {
	logger := SimpleLeveled{
		Printer:  log.New(os.Stdout, "", 0),
		ErrStack: []interface{}{"Error"},
	}
	logger.Push("Test")
	logger.Log("Let's try something")
	logger.Error("It has failed")
	// Output:
	// Test Let's try something
	// Error Test It has failed
}
