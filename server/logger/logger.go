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

// Package logger provides logging facilities with incremental prefixes stored in contexts.
//
// Calls to Print and Println print log to the Target, with all the prefixes stored in the context
// by previous calls to Push. This provides a convenient facility for log in servers, for which
// different components may add information to be logged about the current request. The function
// Constructor is a so-called "middleware" initiating this facility for net/http compatible servers.
package logger

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

// NoLogInfo is returned when the given context has no log prefix stack.
var NoLogInfo = errors.New("Log info not found in context")

// Printer is the type of the Target variable.
type Printer interface {
	Println(a ...interface{})
}

// Target is underlying object used to print the logs.
// By default it is a log.Logger object with the same behaviour as log's standard logger.
var Target Printer

func init() {
	Target = log.New(os.Stderr, "", log.LstdFlags)
}

type contextKey int

const (
	logInfoKey contextKey = iota
)

// New creates returns a copy of ctx with a log prefix stack.
func New(ctx context.Context) context.Context {
	info := make([]interface{}, 0, 4)
	return context.WithValue(ctx, logInfoKey, &info)
}

// Push add the given prefixes to the log prefix stack of the context.
func Push(ctx context.Context, info ...interface{}) error {
	stored, ok := ctx.Value(logInfoKey).(*[]interface{})
	if !ok {
		return NoLogInfo
	}
	*stored = append(*stored, info...)
	return nil
}

// Print send a log message with all the prefixes from the log stack of ctx.
func Print(ctx context.Context, msg ...interface{}) error {
	stored, ok := ctx.Value(logInfoKey).(*[]interface{})
	if !ok {
		return NoLogInfo
	}
	_log(stored, msg...)
	return nil
}

// Print send a log message with all the prefixes from the log stack of ctx.
func Printf(ctx context.Context, format string, msg ...interface{}) error {
	stored, ok := ctx.Value(logInfoKey).(*[]interface{})
	if !ok {
		return NoLogInfo
	}
	_log(stored, fmt.Sprintf(format, msg...))
	return nil
}

func _log(stored *[]interface{}, msg ...interface{}) {
	toPrint := make([]interface{}, 0, len(*stored) + len(msg))
	if len(*stored) > 0 {
		toPrint = append(toPrint, (*stored)...)
	}
	if len(msg) > 0 {
		toPrint = append(toPrint, msg...)
	}
	Target.Println(toPrint...)
}

type middleware struct {
	next http.Handler
}

func (self middleware) ServeHTTP(wr http.ResponseWriter, req *http.Request) {
	start := time.Now()
	info := make([]interface{}, 0, 4)
	ctx := context.WithValue(req.Context(), logInfoKey, &info)
	self.next.ServeHTTP(wr, req.WithContext(ctx))
	_log(&info, "in", time.Now().Sub(start).String())
}

// Constructor is a so-called net/http "middleware" that initialises a log prefix stack in the
// request's context and logs the time elapsed for the request.
func Constructor(next http.Handler) http.Handler {
	return middleware{next}
}
