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
	"net/http"
)

// A HttpError is an error that can be send as an HTTP response.
type HttpError struct {
	// HTTP status code for the error.
	Code    int
	msg     string
	detail  string
	wrapped error
}

// NewHttpError constructs a new HttpError.
//
// The code is to be sent as the HTTP code of the response. It should be a constant from the
// net/http package. The message (msg) is to be sent as body of the HTTP response. This is the
// public description of the error. The detail is the private description of the error, to be
// displayed in the logs.
func NewHttpError(code int, msg string, detail string) HttpError {
	return HttpError{Code: code, msg: msg, detail: detail}
}

// NewInternalHttpError wraps another error into an InternalServerError HttpError.
// This function is particularly usefull to panic inside an Handler, see Handler.
func NewInternalHttpError(err error) HttpError {
	return HttpError{Code: http.StatusInternalServerError, msg: "Internal error", wrapped: err}
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

// A Handler responds to an HTTP request.
//
// The Handle method should read the Request then use Response's methods to send the response.
// The Context must be checked for completion and transmitted to all called functions.
//
// As a convenience, if the Handle method panics with an HttpError then that error is send
// as response.
type Handler interface {
	Handle(ctx context.Context, response Response, request *Request)
}

// HandlerFunc wraps a function into a Handler.
type HandlerFunc func(ctx context.Context, response Response, request *Request)

func (self HandlerFunc) Handle(ctx context.Context, response Response, request *Request) {
	self(ctx, response, request)
}

type handlerWrapper struct {
	pattern string
	handler Handler
}

// ServeHTTP implements http.Handler.
func (self handlerWrapper) ServeHTTP(wr http.ResponseWriter, original *http.Request) {
	ctx := original.Context()
	response := Response{wr}
	request := newRequest(self.pattern, original)

	defer func() {
		if err := recover(); err != nil {
			response.SendError(err.(HttpError))
		}
	}()

	self.handler.Handle(ctx, response, &request)
}

// Handle registers the handler for the given pattern.
// See http.ServeMux for a description of the pattern format.
func Handle(pattern string, handler Handler) {
	http.Handle(pattern, handlerWrapper{pattern: pattern, handler: handler})
}

// Handle registers the handler function for the given pattern.
// See http.ServeMux for a description of the pattern format.
func HandleFunc(pattern string, fct func(ctx context.Context, resp Response, req *Request)) {
	http.Handle(pattern, handlerWrapper{pattern: pattern, handler: HandlerFunc(fct)})
}
