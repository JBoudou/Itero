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
	"net/http"

	"github.com/JBoudou/Itero/server/logger"

	"github.com/justinas/alice"
)

type Interceptor = alice.Constructor

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

type HandleFunction = func(ctx context.Context, response Response, request *Request)

// HandlerFunc wraps a function into a Handler.
type HandlerFunc HandleFunction

func (self HandlerFunc) Handle(ctx context.Context, response Response, request *Request) {
	self(ctx, response, request)
}

// HandlerWrapper wraps a Handler into an http.Handler.
type HandlerWrapper struct {
	pattern string
	handler Handler
}

func NewHandlerWrapper(pattern string, handler Handler) HandlerWrapper {
	return HandlerWrapper{pattern: pattern, handler: handler}
}

// ServeHTTP implements http.Handler.
func (self HandlerWrapper) ServeHTTP(wr http.ResponseWriter, original *http.Request) {
	ctx := original.Context()
	response := response{wr}
	request := NewRequest(self.pattern, original)

	defer func() {
		if thrown := recover(); thrown != nil {
			err := thrown.(error)
			var httpError HttpError
			if !errors.As(err, &httpError) {
				logger.Print(ctx, err)
				panic(err)
			}
			response.SendError(ctx, httpError)
		}
	}()

	self.handler.Handle(ctx, response, &request)
}

// Handle registers the handler for the given pattern.
// See http.ServeMux for a description of the pattern format.
func Handle(pattern string, handler Handler, interceptors... Interceptor) {
	packed := interceptorChain.Append(interceptors...).
		Then(HandlerWrapper{pattern: pattern, handler: handler})
	http.Handle(pattern, packed)
}

// Handle registers the handler function for the given pattern.
// See http.ServeMux for a description of the pattern format.
func HandleFunc(pattern string, fct HandleFunction, interceptors... Interceptor) {
	packed := interceptorChain.Append(interceptors...).
		Then(HandlerWrapper{pattern: pattern, handler: HandlerFunc(fct)})
	http.Handle(pattern, packed)
}
