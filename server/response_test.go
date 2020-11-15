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
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func canceledContext() (ret context.Context) {
	ret, cancelFct := context.WithCancel(context.Background())
	cancelFct()
	return
}

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

func TestHttpError_Error(t *testing.T) {
	type fields struct {
		Code   int
		msg    string
		detail string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "unique",
			fields: fields{Code: 404, msg: "Not found", detail: "For some reason"},
			want: "For some reason",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			self := HttpError{
				Code:   tt.fields.Code,
				msg:    tt.fields.msg,
				detail: tt.fields.detail,
			}
			if got := self.Error(); got != tt.want {
				t.Errorf("HttpError.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResponse_SendJSON(t *testing.T) {
	type args struct {
		ctx  context.Context
		data interface{}
	}

	checkSuccess := func (t *testing.T, mock *httptest.ResponseRecorder, args *args) {
		// Check status code
		result := mock.Result()
		if result.StatusCode < 200 || result.StatusCode >= 300 {
			t.Errorf("Wrong StatusCode %d", result.StatusCode)
		}

		// Read the body
		var buff bytes.Buffer
		if _, err := buff.ReadFrom(result.Body); err != nil {
			t.Fatalf("Error reading body: %s", err.Error())
		}
		body := buff.Bytes()

		// Compare the body
	  if !json.Valid(body) {
			t.Fatal("Invalid JSON in body")
		}
		buff.Reset()
		if err := json.Compact(&buff, body); err != nil {
			t.Fatal(err)
		}
		got := buff.Bytes()
		// We assume json.Marshal produce a compact representation
		expected, err := json.Marshal(args.data)
		if err != nil {
			t.Fatalf("Marshal error: %s", err.Error())
		}
		if !bytes.Equal(got, expected) {
			t.Errorf("Wrong body. Got %s. Expect %s", got, expected)
		}
	}
	checkFail := func (t *testing.T, mock *httptest.ResponseRecorder, args *args) {
		result := mock.Result()
		if result.StatusCode < 400 {
			t.Errorf("Wrong StatusCode %d", result.StatusCode)
		}
	}
	
	tests := []struct {
		name   string
		args   args
		check  func(t *testing.T, mock *httptest.ResponseRecorder, args *args)
	}{
		{
			name: "String",
			args: args{ctx: context.Background(), data: "foobar"},
			check: checkSuccess,
		},
		{
			name: "Number",
			args: args{ctx: context.Background(), data: 42},
			check: checkSuccess,
		},
		{
			name: "Slice",
			args: args{ctx: context.Background(), data: []string{ "a", "b", "c", "d" }},
			check: checkSuccess,
		},
		{
			name: "Struct",
			args: args{ctx: context.Background(), data: struct{foo string; bar int}{"foobar", 42}},
			check: checkSuccess,
		},
		{
			name: "Canceled",
			args: args{ctx: canceledContext(), data: 0},
			check: checkFail,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := httptest.NewRecorder()
			self := Response{
				Writer: mock,
			}
			self.SendJSON(tt.args.ctx, tt.args.data)
			tt.check(t, mock, &tt.args)
		})
	}
}

func TestResponse_SendError(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		expectedCode int // set to zero to disable
	}{
		{
			name: "403",
			err: NewHttpError(http.StatusForbidden, "Forbidden", "Test"),
			expectedCode: http.StatusForbidden,
		},
		{
			name: "explicit 500",
			err: NewHttpError(http.StatusInternalServerError, "Server error", "Test"),
			expectedCode: http.StatusInternalServerError,
		},
		{
			name: "implicit 500",
			err: errors.New("Internal error"),
			expectedCode: http.StatusInternalServerError,
		},
		{
			name: "context canceled",
			err: context.Canceled,
		},
		{
			name: "context deadline exceeded",
			err: context.DeadlineExceeded,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := httptest.NewRecorder()
			self := Response{
				Writer: mock,
			}
			self.SendError(tt.err)

			got := mock.Result().StatusCode
			if got < 400 {
				t.Errorf("Wrong status code %d", got)
			}
			if tt.expectedCode != 0 && got != tt.expectedCode {
				t.Errorf("Status code. Got %d. Expect %d", got, tt.expectedCode)
			}
		})
	}
}

func TestResponse_SendLoginAccepted(t *testing.T) {
	type args struct {
		ctx  context.Context
		user string
	}
	tests := []struct {
		name string
		self Response
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.self.SendLoginAccepted(tt.args.ctx, tt.args.user)
		})
	}
}
