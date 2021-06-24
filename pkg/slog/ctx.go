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

package slog

import (
	"context"
)

type ctxKey int

const (
	ctxKeyLogger ctxKey = iota
)

// CtxSaveLogger creates a context containing the given Logger.
func CtxSaveLogger(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, ctxKeyLogger, logger)
}

// CtxLoadLogger retrieves a Logger from a context.
// The Logger must have been stored in the context by CtxSaveLogger.
func CtxLoadLogger(ctx context.Context) Logger {
	ret, ok := ctx.Value(ctxKeyLogger).(Logger)
	if !ok {
		return nil
	}
	return ret
}

// CtxLog calls Log on the Logger stored in the context.
func CtxLog(ctx context.Context, args ...interface{}) {
	CtxLoadLogger(ctx).Log(args...)
}

// CtxLog calls Log on the Logger stored in the context.
func CtxLogf(ctx context.Context, format string, args ...interface{}) {
	CtxLoadLogger(ctx).Logf(format, args...)
}

// CtxLog calls Error on the logger stored in the context.
// If the stored logger does not have interface Leveled, Log is called instead
// with "Error" as first argument.
func CtxError(ctx context.Context, args ...interface{}) {
	log := CtxLoadLogger(ctx)
	if lvl, ok := log.(Leveled); ok {
		lvl.Error(args...)
	} else {
		log.Log(append([]interface{}{"Error"}, args...)...)
	}
}

// CtxLog calls Errorf on the logger stored in the context.
// If the stored logger does not have interface Leveled, Logf is called instead
// with the format prefixed with "Error ".
func CtxErrorf(ctx context.Context, format string, args ...interface{}) {
	log := CtxLoadLogger(ctx)
	if lvl, ok := log.(Leveled); ok {
		lvl.Errorf(format, args...)
	} else {
		log.Logf("Error "+format, args...)
	}
}
