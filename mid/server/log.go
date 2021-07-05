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
	"time"
	"net/http"
	
	"github.com/JBoudou/Itero/mid/root"
	"github.com/JBoudou/Itero/pkg/slog"
)


func addLogger(next http.Handler) http.Handler {
	var printer slog.Printer
	if err := root.IoC.Inject(&printer); err != nil {
		panic(err)
	}
	return loggerInterceptor{
		next: next,
		logger: &slog.SimpleLogger{
			Printer: printer,
			Stack:   []interface{}{"H"},
		},
	}
}

type loggerInterceptor struct {
	next   http.Handler
	logger slog.Stacked
}

func (self loggerInterceptor) ServeHTTP(wr http.ResponseWriter, req *http.Request) {
	start := time.Now()
	logger := self.logger.With(req.RemoteAddr, req.URL.Path)
	ctx := slog.CtxSaveLogger(req.Context(), logger)
	response := &responseWithStatus{ ResponseWriter: wr, status: http.StatusOK }
	self.next.ServeHTTP(response, req.WithContext(ctx))
	logger.Log(response.status, "in", time.Now().Sub(start).String())
}

type responseWithStatus struct {
	http.ResponseWriter
	status int
}

func (self *responseWithStatus) WriteHeader(statusCode int) {
	self.status = statusCode
	self.ResponseWriter.WriteHeader(statusCode)
}
