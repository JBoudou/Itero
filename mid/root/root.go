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
package root

import (
	"log"
	"os"

	"github.com/JBoudou/Itero/pkg/config"
	"github.com/JBoudou/Itero/pkg/events"
	"github.com/JBoudou/Itero/pkg/ioc"
	"github.com/JBoudou/Itero/pkg/slog"
)

var IoC = ioc.New()

var Configured = false

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
		Configured = config.ReadConfigFile(logger, "config.json", 2)
	})
}
