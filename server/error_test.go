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

package server

import (
	"errors"
	"net/http"
	"strings"
	"testing"
)

func TestNewHttpError(t *testing.T) {
	tests := []struct {
		name   string
		code   int
		msg    string
		detail string
	}{
		{
			name:   "404",
			code:   404,
			msg:    "Not found",
			detail: "For some reason",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewHttpError(tt.code, tt.msg, tt.detail)
			if err.Code != tt.code {
				t.Errorf("Got code %d. Expect %d.", err.Code, tt.code)
			}
			if err.Msg != tt.msg {
				t.Errorf(`Got msg "%s". Expect "%s".`, err.Msg, tt.msg)
			}
			if !strings.Contains(err.Error(), tt.detail) {
				t.Errorf(`Got Error() "%s". Expected to contain "%s".`, err.Error(), tt.detail)
			}
			if err.Unwrap() != nil {
				t.Errorf("Got %v. Expect nil.", err.Unwrap())
			}
		})
	}
}

func TestWrapError(t *testing.T) {
	tests := []struct {
		name string
		code int
		msg  string
		err  error
	}{
		{
			name: "Simple",
			code: 427,
			msg:  "wrapping",
			err:  errors.New("wrapped"),
		},
		{
			name: "Recursive",
			code: 400,
			msg:  "outter",
			err:  WrapError(401, "inner", errors.New("wrapped")),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WrapError(tt.code, tt.msg, tt.err)
			if result.Code != tt.code {
				t.Errorf("Got code %d. Expect %d.", result.Code, tt.code)
			}
			if result.Msg != tt.msg {
				t.Errorf(`Got msg "%s". Expect "%s".`, result.Msg, tt.msg)
			}
			if !strings.Contains(result.Error(), tt.err.Error()) {
				t.Errorf(`Error() gives "%s". Expect to contain "%s".`, result.Error(), tt.err.Error())
			}
			if !errors.Is(result, tt.err) {
				t.Errorf("Unwrap() gives %v. Expect %v.", result.Unwrap(), tt.err)
			}
		})
	}
}

func TestInternalHttpError(t *testing.T) {
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
			got := InternalHttpError(tt.err)
			if got.Code != http.StatusInternalServerError {
				t.Errorf("Got Code %d. Expect %d.", got.Code, http.StatusInternalServerError)
			}
			if got.Msg != InternalHttpErrorMsg {
				t.Errorf(`Got Msg "%s". Expect "%s".`, got.Msg, InternalHttpErrorMsg)
			}
			if !strings.Contains(got.Error(), tt.err.Error()) {
				t.Errorf(`Error() gives "%s". Expect to contain "%s".`, got.Error(), tt.err.Error())
			}
			if !errors.Is(got, tt.err) {
				t.Errorf("Unwrap() gives %v. Expect %v.", got.Unwrap(), tt.err)
			}
		})
	}
}

func TestUnauthorizedHttpError(t *testing.T) {
	tests := []struct {
		name string
		detail string
	}{
		{
			name: "Simple",
			detail: "Test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := UnauthorizedHttpError(tt.detail)
			if result.Code != http.StatusForbidden {
				t.Errorf("Got Code %d. Expect %d.", result.Code, http.StatusForbidden)
			}
			if result.Msg != UnauthorizedHttpErrorMsg {
				t.Errorf(`Got Msg "%s". Expect "%s".`, result.Msg, UnauthorizedHttpErrorMsg)
			}
			if !strings.Contains(result.Error(), tt.detail) {
				t.Errorf(`Got Error() "%s". Expected to contain "%s".`, result.Error(), tt.detail)
			}
			if result.Unwrap() != nil {
				t.Errorf("Got %v. Expect nil.", result.Unwrap())
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
			self: HttpError{Msg: "Not found", detail: "For some reason"},
			want: "For some reason",
		},
		{
			name: "wrapped",
			self: HttpError{Msg: "A", detail: "B", wrapped: errors.New("C")},
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
