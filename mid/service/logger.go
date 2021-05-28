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

package service

import (
	"log"

	"github.com/JBoudou/Itero/pkg/ioc"
)

func init() {
	ioc.Root.Set(func () LevelLogger { return EasyLogger{} })
}

// LevelLogger is a temporary interface before the new logger facility is implemented.
type LevelLogger interface {
	Logf(format string, v ...interface{})
	Warnf(format string, v ...interface{})
	Errorf(format string, v ...interface{})
}

// EasyLogger is a temporary implementation of LevelLogger.
type EasyLogger struct{}

func (self EasyLogger) Logf(format string, v ...interface{}) {
	log.Printf(format, v...)
}

func (self EasyLogger) Warnf(format string, v ...interface{}) {
	log.Printf("Warn: "+format, v...)
}

func (self EasyLogger) Errorf(format string, v ...interface{}) {
	log.Printf("Err: "+format, v...)
}

type prefixLogger struct {
	logPrefix string
	warnPrefix string
	errorPrefix string
}

// NewPrefixLogger construct a single LevelLogger that simply flag logs.
func NewPrefixLogger(prefix string) LevelLogger {
	return &prefixLogger{
		logPrefix: prefix + " ",
		warnPrefix: prefix + " WARN ",
		errorPrefix: prefix + " ERROR ",
	}
}

func (self *prefixLogger) Logf(format string, v ...interface{}) {
	log.Printf(self.logPrefix + format, v...)
}

func (self *prefixLogger) Warnf(format string, v ...interface{}) {
	log.Printf(self.warnPrefix + format, v...)
}

func (self *prefixLogger) Errorf(format string, v ...interface{}) {
	log.Printf(self.errorPrefix + format, v...)
}
