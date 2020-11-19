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
	"errors"
	"reflect"
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
