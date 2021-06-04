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
	"github.com/JBoudou/Itero/pkg/events"
	"github.com/JBoudou/Itero/pkg/ioc"
)

func init() {
	// Events
	// TODO: do not have events.DefaultManager
	ioc.Root.Set(func() events.Manager { return events.DefaultManager })
}
