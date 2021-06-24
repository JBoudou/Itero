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

// Package main/services contains the concrete services used by Itero middleware server.
package services

import (
	"log"
	"os"

	"github.com/JBoudou/Itero/mid/service"
	"github.com/JBoudou/Itero/pkg/alarm"
	"github.com/JBoudou/Itero/pkg/events"
	"github.com/JBoudou/Itero/pkg/ioc"
	"github.com/JBoudou/Itero/pkg/slog"
)

func init() {
	ioc.Root.Bind(func() service.AlarmInjector { return alarm.New })
	ioc.Root.Bind(func() events.Manager { return events.NewAsyncManager(events.DefaultManagerChannelSize) })

	// log
	ioc.Root.Bind(func() slog.StackedLeveled {
		return &slog.SimpleLeveled{
			Printer:  log.New(os.Stderr, "", log.LstdFlags|log.Lmicroseconds),
			LogStack: []interface{}{"I"},
			ErrStack: []interface{}{"W"},
		}
	})
	ioc.Root.Bind(func(l slog.StackedLeveled) slog.Leveled { return l })
	ioc.Root.Bind(func(l slog.StackedLeveled) slog.Stacked { return slog.AsStacked{l} })
	ioc.Root.Bind(func(l slog.StackedLeveled) slog.Logger { return l })
}

func serviceLogger(prefix string) func(l slog.StackedLeveled) slog.Leveled {
	return func(l slog.StackedLeveled) slog.Leveled {
		return l.With(prefix)
	}
}
