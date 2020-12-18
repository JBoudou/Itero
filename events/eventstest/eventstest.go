// Itero - Online iterative vote application
// Copyright (C) 2020 Joseph Boudou
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

package eventstest

import (
	"testing"

	. "github.com/JBoudou/Itero/events"
)

type ManagerMock struct {
	T           *testing.T
	Send_        func(Event) error
	AddReceiver_ func(Receiver) error
	Close_      func() error
}

func (self *ManagerMock) Send(evt Event) error {
	if self.Send_ == nil {
		self.T.Fatal("Send unexpectedly called")
	}
	return self.Send_(evt)
}

func (self *ManagerMock) AddReceiver(rcv Receiver) error {
	if self.AddReceiver_ == nil {
		self.T.Fatal("AddReceiver unexpectedly called")
	}
	return self.AddReceiver_(rcv)
}

func (self *ManagerMock) Close() error {
	if self.Close_ == nil {
		self.T.Fatal("Close unexpectedly called")
	}
	return self.Close_()
}

type ReceiverMock struct {
	T *testing.T
	Receive_ func(Event)
	Close_ func()
}

func (self *ReceiverMock) Receive(evt Event) {
	if self.Receive_ == nil {
		self.T.Fatal("Receive unexpectedly called")
	}
	self.Receive_(evt)
}

func (self *ReceiverMock) Close() {
	if self.Close_ == nil {
		self.T.Fatal("Close unexpectedly called")
	}
	self.Close_()
}
