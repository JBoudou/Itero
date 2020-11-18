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

package server

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestNewHttpError(t *testing.T) {
	type args struct {
		code   int
		msg    string
		detail string
	}
	tests := []struct {
		name string
		args args
		want HttpError
	}{
		{
			name: "unique",
			args: args{code: 404, msg: "Not found", detail: "For some reason"},
			want: HttpError{Code: 404, msg: "Not found", detail: "For some reason"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewHttpError(tt.args.code, tt.args.msg, tt.args.detail); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewHttpError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewInternalHttpError(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{
			name: "unique",
			err:  errors.New("Test"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewInternalHttpError(tt.err)
			if got.Code < 400 {
				t.Errorf("Wrong status code %d", got.Code)
			}
			if got.wrapped != tt.err {
				t.Errorf("Wrong wrapped. Got %v. Expect %v", got.wrapped, tt.err)
			}
		})
	}
}

func TestHttpError_Error(t *testing.T) {
	type fields struct {
		Code   int
		msg    string
		detail string
	}
	tests := []struct {
		name string
		self HttpError
		want string
	}{
		{
			name: "unwrapped",
			self: HttpError{msg: "Not found", detail: "For some reason"},
			want: "For some reason",
		},
		{
			name: "wrapped",
			self: HttpError{msg: "A", detail: "B", wrapped: errors.New("C")},
			want: "C",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.self.Error(); got != tt.want {
				t.Errorf("HttpError.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHandle(t *testing.T) {
	type args struct {
		pattern string
		handler Handler
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Handle(tt.args.pattern, tt.args.handler)
		})
	}
}

type handlerArgs struct {
	pattern string
	fct     func(ctx context.Context, resp Response, req *Request)
}
type reqFields struct {
	method string
	target string
	body   string
}
type handlerTestsStruct struct {
	name    string
	args    handlerArgs
	req     reqFields
	checker TestChecker
}

var handlerTests []handlerTestsStruct = []handlerTestsStruct{
	{
		name: "write",
		args: handlerArgs{
			pattern: "/t/write",
			fct: func(ctx context.Context, resp Response, req *Request) {
				msg := "bar"
				resp.SendJSON(ctx, msg)
			},
		},
		req: reqFields{
			method: "GET",
			target: "/t/write",
		},
		checker: TestCheckerJSONString(http.StatusOK, "bar"),
	},
	{
		name: "echo",
		args: handlerArgs{
			pattern: "/t/echo",
			fct: func(ctx context.Context, resp Response, req *Request) {
				var msg string
				if err := req.UnmarshalJSONBody(&msg); err != nil {
					resp.SendError(err)
					return
				}
				resp.SendJSON(ctx, msg)
			},
		},
		req: reqFields{
			method: "POST",
			target: "/t/echo",
			body:   `"Hello"`,
		},
		checker: TestCheckerJSONString(http.StatusOK, "Hello"),
	},
	{
		name: "error",
		args: handlerArgs{
			pattern: "/t/error",
			fct: func(ctx context.Context, resp Response, req *Request) {
				resp.SendError(NewHttpError(http.StatusPaymentRequired, "Flublu", "Test"))
			},
		},
		req: reqFields{
			method: "GET",
			target: "/t/error",
		},
		checker: TestCheckerRawString(http.StatusPaymentRequired, "Flublu"),
	},
	{
		name: "panic",
		args: handlerArgs{
			pattern: "/t/panic",
			fct: func(ctx context.Context, resp Response, req *Request) {
				panic(NewHttpError(http.StatusPaymentRequired, "Barbaz", "Test"))
			},
		},
		req: reqFields{
			method: "GET",
			target: "/t/panic",
		},
		checker: TestCheckerRawString(http.StatusPaymentRequired, "Barbaz"),
	},
	{
		name: "struct",
		args: handlerArgs{
			pattern: "/t/struct",
			fct: func(ctx context.Context, resp Response, req *Request) {
				resp.SendJSON(ctx, struct {A int; B string}{A: 42, B: "Foobar"})
			},
		},
		req: reqFields{
			method: "GET",
			target: "/t/struct",
		},
		checker: TestCheckerJSON(http.StatusOK, struct{A int; B string}{A: 42, B: "Foobar"}),
	},
}

func TestHandleFunc(t *testing.T) {
	for _, tt := range handlerTests {
		t.Run(tt.name, func(t *testing.T) {
			HandleFunc(tt.args.pattern, tt.args.fct)

			var body io.Reader
			if tt.req.body != "" {
				body = strings.NewReader(tt.req.body)
			}

			wr := httptest.NewRecorder()
			req := httptest.NewRequest(tt.req.method, tt.req.target, body)
			http.DefaultServeMux.ServeHTTP(wr, req)

			tt.checker(t, wr.Result(), req)
		})
	}
}

func TestTestHandlerFunc(t *testing.T) {
	for _, tt := range handlerTests {
		t.Run(tt.name, func(t *testing.T) {
			var body io.Reader
			if tt.req.body != "" {
				body = strings.NewReader(tt.req.body)
			}
			req := httptest.NewRequest(tt.req.method, tt.req.target, body)
			response := TestHandlerFunc(tt.args.pattern, tt.args.fct, req)
			tt.checker(t, response, req)
		})
	}
}
