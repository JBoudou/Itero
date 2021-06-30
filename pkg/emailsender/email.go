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

// emailsender package provides email senders.
package emailsender

import (
	"errors"
	"text/template"
)

var (
	WrongEmailValue = errors.New("Wrong email value (missing recipient or nil template)")
)

type Email struct {
	To   []string
	Tmpl *template.Template
	Data interface{}
}

type Sender interface {

	// Send asks the sender to send an email. The email may be kept in memory and sent later.
	// Hence nil return value does not mean the email has been sent.
	Send(email Email) error

	// Close closes the sender. This may result in retained email to be sent.
	Close() error
}
