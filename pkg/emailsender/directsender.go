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

package emailsender

import (
	"net/smtp"
)

// DirectSender is a sender that directly sends emails to an SMTP server.
// The connection to the server is left open for multiple emails to be send in one session.
type DirectSender struct {
	Sender string // email address
	SMTP   string // host:port

	client *smtp.Client
}

// Send sends an email to the SMTP server. It opens a new connection if needed, otherwise it uses
// the previously opened connection.
func (self *DirectSender) Send(email Email) (err error) {
	if self.client == nil {
		err = self.connect()
		if err != nil {
			return
		}
	}

	err = self.client.Mail(self.Sender)
	if err != nil {
		return
	}

	for _, to := range email.To {
		err = self.client.Rcpt(to)
		if err != nil {
			return
		}
	}

	wr, err := self.client.Data()
	if err != nil {
		return
	}
	err = email.Tmpl.Execute(wr, email.Data)
	if err != nil {
		return
	}
	err = wr.Close()
	if err != nil {
		return
	}

	return
}

func (self *DirectSender) connect() (err error) {
	self.client, err = smtp.Dial(self.SMTP)
	return
}

func (self *DirectSender) Close() error {
	return self.Quit()
}

// Quit closes the currently opened connection, if any. It should be called as soon as there is no
// email to send.
func (self *DirectSender) Quit() error {
	if self.client == nil {
		return nil
	}
	defer func() { self.client = nil }()
	return self.client.Quit()
}
