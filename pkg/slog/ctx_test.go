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
	"context"
	"reflect"
	"testing"
)

func TestCtxSaveLoad(t *testing.T) {
	t.Parallel()

	original := &SimpleLeveled{}
	ctx := CtxSaveLogger(context.Background(), original)
	loaded := CtxLoadLogger(ctx)

	if loaded == nil {
		t.Errorf("loaded is nil")
	}
	if _, ok := loaded.(Logger); !ok {
		t.Errorf("loaded is not a Logger")
	}
	if _, ok := loaded.(StackedLeveled); !ok {
		t.Errorf("loaded is not a StackedLeveled")
	}
}

func TestCtxLogger(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		action    func(log context.Context)
		expect    [][]interface{}
		expectLvl [][]interface{} // Expectation on StackedLeveled. If nil expect is used instead.
	}{
		{
			name:   "nothing",
			action: func(ctx context.Context) {},
		},
		{
			name:   "Log",
			action: func(ctx context.Context) { CtxLog(ctx, 1, "2", 3) },
			expect: [][]interface{}{{1, "2", 3}},
		},
		{
			name:   "Logf",
			action: func(ctx context.Context) { CtxLogf(ctx, "%d %s", 1, "2") },
			expect: [][]interface{}{{"1 2"}},
		},
		{
			name:   "Error",
			action: func(ctx context.Context) { CtxError(ctx, 1, "2", 3) },
			expect: [][]interface{}{{"Error", 1, "2", 3}},
		},
		{
			name:      "Errorf",
			action:    func(ctx context.Context) { CtxErrorf(ctx, "%d %s", 1, "2") },
			expect:    [][]interface{}{{"Error 1 2"}},
			expectLvl: [][]interface{}{{"Error", "1 2"}},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run("Logger/"+tt.name, func(t *testing.T) {
			t.Parallel()

			recorder := printRecorder{}
			tt.action(CtxSaveLogger(context.Background(), &SimpleLogger{Printer: &recorder}))

			if !reflect.DeepEqual(recorder.records, tt.expect) {
				t.Errorf("Got %v. Expect %v.", recorder.records, tt.expect)
			}
		})

		t.Run("Leveled/"+tt.name, func(t *testing.T) {
			t.Parallel()

			recorder := printRecorder{}
			tt.action(CtxSaveLogger(context.Background(), &SimpleLeveled{
				Printer:  &recorder,
				ErrStack: []interface{}{"Error"},
			}))

			expect := tt.expectLvl
			if expect == nil {
				expect = tt.expect
			}
			if !reflect.DeepEqual(recorder.records, expect) {
				t.Errorf("Got %v. Expect %v.", recorder.records, expect)
			}
		})
	}
}
