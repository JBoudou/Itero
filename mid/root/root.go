// Itero - Online iterative vote application
// Copyright (C) 2021 Joseph Boudou
//
// This program is free software: you can redistribute it and/or modify it under
// the terms of the GNU General Public License as published by the Free Software
// Foundation, either version 3 of the License, or (at your option) any later
// version.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS
// FOR A PARTICULAR PURPOSE. See the GNU General Public License for more
// details.
//
// You should have received a copy of the GNU General Public License along with
// this program. If not, see <https://www.gnu.org/licenses/>.

// Package root provides shared resources to the whole application.
//
// In particular, this package initialize IoC with binding for external packages.
// This package can be imported by all other application specific packages (main/* and mid/*).
// Therefore, it must not import any application specific package.
package root

import (
	"hash"
	"log"
	"os"

	"golang.org/x/crypto/blake2b"

	"github.com/JBoudou/Itero/pkg/config"
	"github.com/JBoudou/Itero/pkg/events"
	"github.com/JBoudou/Itero/pkg/ioc"
	"github.com/JBoudou/Itero/pkg/slog"
)

// IoC is the root locator for the application.
// Application packages providing services must Bind to it in their init function.
var IoC = ioc.New()

// Configured indicates whether the configuration file has successfully been read.
var Configured = false

// BaseDir is the path in which the configuration file has been found.
var BaseDir string

func init() {
	IoC.Bind(func() events.Manager { return events.NewAsyncManager(events.DefaultManagerChannelSize) })

	// log
	IoC.Bind(func() slog.Printer {
		return log.New(os.Stderr, "", log.LstdFlags|log.Lmicroseconds)
	})
	IoC.Bind(func(printer slog.Printer) slog.StackedLeveled {
		return &slog.SimpleLeveled{
			Printer:  printer,
			LogStack: []interface{}{"M"},
			ErrStack: []interface{}{"E"},
		}
	})
	IoC.Bind(func(l slog.StackedLeveled) slog.Leveled { return l })
	IoC.Bind(func(l slog.StackedLeveled) slog.Stacked { return slog.AsStacked{l} })
	IoC.Bind(func(l slog.StackedLeveled) slog.Logger { return l })

	// Config
	IoC.Inject(func(logger slog.Leveled) {
		var err error
		BaseDir, err = config.ReadFile(logger, "config.json", 2)
		if err == nil {
			Configured = true
		} else {
			logger.Errorf("Configuration error: %v", err)
		}
	})
}

// PasswdHash provides the hash function used for passwords.
func PasswdHash() (hash.Hash, error) {
	return blake2b.New256(nil)
}
