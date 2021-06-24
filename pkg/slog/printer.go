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
	"fmt"
)

// Printer represents backends for Logger.
type Printer interface {
	Println(a ...interface{})
}

// Log is a low level function to implement Logger.
// It uses the given Printer to display the stack and the arguments.
func Log(printer Printer, stack []interface{}, args ...interface{}) {
	toPrint := make([]interface{}, 0, len(stack) + len(args))
	if len(stack) > 0 {
		toPrint = append(toPrint, (stack)...)
	}
	if len(args) > 0 {
		toPrint = append(toPrint, args...)
	}
	printer.Println(toPrint...)
}

// Log is a low level function to implement Logger.
// It uses the given Printer to display the stack and the formatted arguments.
func Logf(printer Printer, stack []interface{}, format string, args ...interface{}) {
	Log(printer, stack, fmt.Sprintf(format, args...))
}
