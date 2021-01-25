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
	"net/http"
)

// A HttpError is an error that can be send as an HTTP response.
type HttpError struct {
	// HTTP status code for the error.
	Code    int
	
	// Message to send in the response.
	Msg     string

	detail  string
	wrapped error
}

func (self HttpError) Error() string {
	if self.wrapped == nil {
		return self.detail
	} else {
		return self.wrapped.Error()
	}
}

func (self HttpError) Unwrap() error {
	return self.wrapped
}

// NewHttpError constructs a new HttpError.
//
// The code is to be sent as the HTTP code of the response. It should be a constant from the
// net/http package. The message (msg) is to be sent as body of the HTTP response. This is the
// public description of the error. The detail is the private description of the error, to be
// displayed in the logs.
func NewHttpError(code int, msg string, detail string) HttpError {
	return HttpError{Code: code, Msg: msg, detail: detail}
}

// WrapError wraps an error into an HttpError.
// Detail of the resulting error is the Error() message of the wrapped error.
func WrapError(code int, msg string, err error) HttpError {
	return HttpError{Code: code, Msg: msg, wrapped: err}
}

const (
	InternalHttpErrorMsg = "Internal error"
	UnauthorizedHttpErrorMsg = "Unauthorized"
)

// InternalHttpError wraps another error into an InternalServerError HttpError.
// This function is particularly usefull to panic inside an Handler, see Handler.
func InternalHttpError(err error) HttpError {
	return WrapError(http.StatusInternalServerError, InternalHttpErrorMsg, err)
}

// UnauthorizedHttpError creates a preformatted HttpError notifying unauthorized request.
func UnauthorizedHttpError(detail string) HttpError {
	return NewHttpError(http.StatusForbidden, UnauthorizedHttpErrorMsg, detail)
}

// WrapUnauthorizedError wrap an error into a preformatted HttpError notifying unauthorized request.
func WrapUnauthorizedError(err error) HttpError {
	return WrapError(http.StatusForbidden, UnauthorizedHttpErrorMsg, err)
}
