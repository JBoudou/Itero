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
	"bytes"
	"context"
	"encoding/json"
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

type checkerFunction = func(t *testing.T, response *http.Response, req *http.Request)

func checkerJSON(expectCode int, expectBody string) checkerFunction {
	return func(t *testing.T, response *http.Response, req *http.Request) {
		if response.StatusCode != expectCode {
			t.Errorf("Wrong status code. Got %d. Expect %d", response.StatusCode, expectCode)
		}

		var body string
		var buff bytes.Buffer
		if _, err := buff.ReadFrom(response.Body); err != nil {
			t.Fatalf("Error reading body: %s", err)
		}
		if err := json.Unmarshal(buff.Bytes(), &body); err != nil {
			t.Fatalf("Error reading body: %s", err)
		}
		if body != expectBody {
			t.Errorf("Wrong body. Got %s. Expect %s", body, expectBody)
		}
	}
}

func checkerRaw(expectCode int, expectBody string) checkerFunction {
	return func(t *testing.T, response *http.Response, req *http.Request) {
		if response.StatusCode != expectCode {
			t.Errorf("Wrong status code. Got %d. Expect %d", response.StatusCode, expectCode)
		}

		var buff bytes.Buffer
		if _, err := buff.ReadFrom(response.Body); err != nil {
			t.Fatalf("Error reading body: %s", err)
		}
		if body := strings.TrimSpace(string(buff.Bytes())); body != expectBody {
			t.Errorf("Wrong body. Got %s. Expect %s", body, expectBody)
		}
	}
}

type hanlderArgs struct {
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
	args    hanlderArgs
	req     reqFields
	checker checkerFunction
}

var handlerTests []handlerTestsStruct = []handlerTestsStruct{
	{
		name: "write",
		args: hanlderArgs{
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
		checker: checkerJSON(http.StatusOK, "bar"),
	},
	{
		name: "echo",
		args: hanlderArgs{
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
		checker: checkerJSON(http.StatusOK, "Hello"),
	},
	{
		name: "error",
		args: hanlderArgs{
			pattern: "/t/error",
			fct: func(ctx context.Context, resp Response, req *Request) {
				resp.SendError(NewHttpError(http.StatusPaymentRequired, "Flublu", "Test"))
			},
		},
		req: reqFields{
			method: "GET",
			target: "/t/error",
		},
		checker: checkerRaw(http.StatusPaymentRequired, "Flublu"),
	},
	{
		name: "panic",
		args: hanlderArgs{
			pattern: "/t/panic",
			fct: func(ctx context.Context, resp Response, req *Request) {
				panic(NewHttpError(http.StatusPaymentRequired, "Barbaz", "Test"))
			},
		},
		req: reqFields{
			method: "GET",
			target: "/t/panic",
		},
		checker: checkerRaw(http.StatusPaymentRequired, "Barbaz"),
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
