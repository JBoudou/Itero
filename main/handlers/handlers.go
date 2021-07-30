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

// Package handlers contains the handler for all the HTTP request on the server.
//
// Types and functions whose name ends with "Handler" are the handlers.
// Types whose name ends with "Query" are the types of the information received in the requests.
// Types whose name ends with "Answer" are the types of the information sent in the responses.
//
// Handlers are either handler functions (of type server.HandleFunction), or factories for handler
// objects (of type server.Handler).
package handlers

import (
	"errors"

	"github.com/JBoudou/Itero/mid/server"
)

// must ensures that err is nil. If it's not, the error is sent by panic, after being wrapped in a
// server.HttpError if it's not already one.
func must(err error) {
	if err != nil {
		var httpError server.HttpError
		if !errors.As(err, &httpError) {
			err = server.InternalHttpError(err)
		}
		panic(err)
	}
}
