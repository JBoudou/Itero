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

package emailsendertest

import (
	"testing"

	"github.com/JBoudou/Itero/pkg/emailsender"
)

type SenderMock struct {
	T *testing.T
	Send_ func(emailsender.Email) error
	Close_ func() error
}

func (self SenderMock) Send(email emailsender.Email) error {
	if self.Send_ == nil {
		self.T.Error("Send unexpectedly called")
		return nil
	}
	return self.Send_(email)
}

func (self SenderMock) Close() error {
	if self.Close_ == nil {
		self.T.Error("Close unexpectedly called")
		return nil
	}
	return self.Close_()
}
