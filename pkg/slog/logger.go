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

// Package log provides simple log facilities.
// Loggers may have two levels (log and error) and a stack of prefixes.
//
// Example
//
// The following code
//     logger := slog.SimpleLeveled{
//         Printer: log.New(os.Stderr, "", 0),
//         ErrStack: []interface{}{"Error"},
//     }
//     logger.Push("Test")
//     logger.Log("Let's try something")
//     logger.Error("It has failed")
// must print
//     Test Let's try something
//     Error Test It has failed
package slog

// Logger is the base type for loggers.
type Logger interface {
	Log(args ...interface{})
	Logf(format string, args ...interface{})
}

// Stacked represents loggers with a stack of prefixes.
// All values on the stack are prepended to all messages.
type Stacked interface {
	Logger

	// Push appends values to the stack of prefixes.
	Push(args ...interface{})

	// With creates a new Stacked logger with the given values appended to the stack of the current
	// object.
	With(args ...interface{}) Stacked
}

// Leveled represents loggers with two levels: log and error.
type Leveled interface {
	Logger
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
}

// StackedLeveled represents two levels loggers with a stack of prefixes.
// All values on the stack are prepended to all messages.
type StackedLeveled interface {
	Leveled

	// Push appends values to the stack of prefixes.
	Push(args ...interface{})

	// With creates a new StackedLeveled logger with the given values appended to the stack of the
	// current object.
	With(args ...interface{}) StackedLeveled
}

//
// AsStacked
//

// AsStacked is a simple wrapper from StackedLeveled to Stacked.
type AsStacked struct {
	StackedLeveled
}

func (self AsStacked) With(args ...interface{}) Stacked {
	return AsStacked{self.StackedLeveled.With(args...)}
}

//
// SimpleLogger
//

// SimpleLogger is a simple stupid implementation of Stacked.
// There is no constructor and the zero value is not usable; you must provide a Printer.
type SimpleLogger struct {
	Printer Printer
	Stack   []interface{}
}

func (self SimpleLogger) Log(args ...interface{}) {
	Log(self.Printer, self.Stack, args...)
}

func (self SimpleLogger) Logf(format string, args ...interface{}) {
	Logf(self.Printer, self.Stack, format, args...)
}

func (self *SimpleLogger) Push(args ...interface{}) {
	self.Stack = append(self.Stack, args...)
}

func (self SimpleLogger) With(args ...interface{}) Stacked {
	return &SimpleLogger{
		Printer: self.Printer,
		Stack: append(self.Stack, args...),
	}
}

//
// SimpleLeveled
//

// SimpleLeveled is an implementation of StackedLeveled.
//
// There is no constructor and the zero value is not usable; you must provide a Printer.
//
// The two levels are distinguished by their stack only. In spite of that, both Push and With
// append the same values to both stacks. The distinction between the two stacks should be done when
// constructing a new SimpleLeveled.
type SimpleLeveled struct {
	Printer Printer
	LogStack []interface{}
	ErrStack []interface{}
}

func (self SimpleLeveled) Log(args ...interface{}) {
	Log(self.Printer, self.LogStack, args...)
}

func (self SimpleLeveled) Logf(format string, args ...interface{}) {
	Logf(self.Printer, self.LogStack, format, args...)
}

func (self SimpleLeveled) Error(args ...interface{}) {
	Log(self.Printer, self.ErrStack, args...)
}

func (self SimpleLeveled) Errorf(format string, args ...interface{}) {
	Logf(self.Printer, self.ErrStack, format, args...)
}

// Push appends values to the stacks of prefixes.
// The values are appended for both levels.
func (self *SimpleLeveled) Push(args ...interface{}) {
	self.LogStack = append(self.LogStack, args...)
	self.ErrStack = append(self.ErrStack, args...)
}

// With creates a new StackedLeveled logger with the given values appended to the stack of the
// current object.
// The values are appended for both levels.
func (self *SimpleLeveled) With(args ...interface{}) StackedLeveled {
	return &SimpleLeveled{
		Printer: self.Printer,
		LogStack: append(self.LogStack, args...),
		ErrStack: append(self.ErrStack, args...),
	}
}
