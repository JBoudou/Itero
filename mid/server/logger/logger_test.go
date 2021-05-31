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

package logger

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

var printerStackEmpty = errors.New("Nothing to read from the printer")

type printer [][]interface{}

func (self *printer) Println(a ...interface{}) {
	*self = append(*self, a)
}

func (self *printer) Read() (v []interface{}, err error) {
	if len(*self) == 0 {
		return nil, printerStackEmpty
	}
	v = (*self)[0]
	*self = (*self)[1:]
	return
}

func (self *printer) Check(t *testing.T, expect ...interface{}) {
	t.Helper()
	got, err := self.Read()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, expect) {
		t.Errorf("Got %v. Expect %v.", got, expect)
	}
}

func TestPrint(t *testing.T) {
	ctx := context.Background()
	var target printer
	Target = &target

	tests := []struct {
		name   string
		update func(t *testing.T, ctx context.Context) context.Context
		args   []interface{}
		expect []interface{}
	}{
		{
			name: "New",
			update: func(t *testing.T, ctx context.Context) context.Context {
				return New(ctx)
			},
			args: []interface{}{ "Hello" },
			expect: []interface{}{ "Hello" },
		},
		{
			name: "Push",
			update: func(t *testing.T, ctx context.Context) context.Context {
				if err := Push(ctx, "Hello"); err != nil {
					t.Fatal(err)
				}
				return ctx
			},
			args: []interface{}{ "world" },
			expect: []interface{}{ "Hello", "world" },
		},
		{
			name: "two",
			update: func(t *testing.T, ctx context.Context) context.Context {
				if err := Push(ctx, "dear", "test"); err != nil {
					t.Fatal(err)
				}
				return ctx
			},
			args: []interface{}{ 1, 2 },
			expect: []interface{}{ "Hello", "dear", "test", 1, 2 },
		},
		{
			name: "WithCancel parent",
			update: func(t *testing.T, ctx context.Context) context.Context {
				child, _ := context.WithCancel(ctx)
				if err := Push(child, 1); err != nil {
					t.Fatal(err)
				}
				return ctx
			},
			args: []interface{}{ 2 },
			expect: []interface{}{ "Hello", "dear", "test", 1, 2 },
		},
		{
			name: "WithCancel child",
			update: func(t *testing.T, ctx context.Context) context.Context {
				child, _ := context.WithCancel(ctx)
				if err := Push(child, 2); err != nil {
					t.Fatal(err)
				}
				return child
			},
			args: []interface{}{ 3 },
			expect: []interface{}{ "Hello", "dear", "test", 1, 2, 3 },
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx = tt.update(t, ctx)
			if err := Print(ctx, tt.args...); err != nil {
				t.Errorf("Print() error = %v.", err)
			}
			target.Check(t, tt.expect...)
		})
	}
}

func TestPrintf(t *testing.T) {
	ctx := context.Background()
	var target printer
	Target = &target

	type args struct {
		format string
		values []interface{}
	}
	tests := []struct {
		name   string
		update func(t *testing.T, ctx context.Context) context.Context
		args   args
		expect []interface{}
	}{
		{
			name: "New",
			update: func(t *testing.T, ctx context.Context) context.Context {
				return New(ctx)
			},
			args: args{format: "Hello", values: []interface{}{}},
			expect: []interface{}{ "Hello" },
		},
		{
			name: "Push",
			update: func(t *testing.T, ctx context.Context) context.Context {
				if err := Push(ctx, "Hello"); err != nil {
					t.Fatal(err)
				}
				return ctx
			},
			args: args{format: "%s", values: []interface{}{ "world" }},
			expect: []interface{}{ "Hello", "world" },
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx = tt.update(t, ctx)
			if err := Printf(ctx, tt.args.format, tt.args.values...); err != nil {
				t.Errorf("Printf() error = %v.", err)
			}
			target.Check(t, tt.expect...)
		})
	}
}

func Test_middleware_ServeHTTP(t *testing.T) {
	var target printer
	Target = &target

	tests := []struct {
		name   string
		next func(t *testing.T)http.HandlerFunc
		expect [][]interface{}
		finalPrefixes []interface{}
	}{
		{
			name: "Basic",
			next: func(t *testing.T)http.HandlerFunc {
				return func(wr http.ResponseWriter, req *http.Request) {
					ctx := req.Context()
					if err := Push(ctx, req.Method); err != nil {
						t.Fatal(err)
					}
					if err := Print(ctx, "inside"); err != nil {
						t.Fatal(err)
					}
				}
			},
			expect: [][]interface{} { { "GET", "inside" } },
			finalPrefixes: []interface{} { "GET" },
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			self := middleware{
				next: tt.next(t),
			}
			self.ServeHTTP(nil, httptest.NewRequest("GET", "/foo", nil))
			for _, expect := range tt.expect {
				target.Check(t, expect...)
			}
			got, err := target.Read()
			if err != nil {
				t.Fatal(err)
			}
			for i, expect := range tt.finalPrefixes {
				if !reflect.DeepEqual(got[i], expect) {
					t.Errorf("Got %v. Expect %v.", got, expect)
				}
			}
		})
	}
}
