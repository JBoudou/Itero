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

package events

import (
	"testing"
)

type managerMock struct {
	t           *testing.T
	send        func(Event) error
	addReceiver func(Receiver) error
	_close      func() error
}

func (self *managerMock) Send(evt Event) error {
	if self.send == nil {
		self.t.Fatal("Send unexpectedly called")
	}
	return self.send(evt)
}

func (self *managerMock) AddReceiver(rcv Receiver) error {
	if self.addReceiver == nil {
		self.t.Fatal("AddReceiver unexpectedly called")
	}
	return self.addReceiver(rcv)
}

func (self *managerMock) Close() error {
	if self._close == nil {
		self.t.Fatal("Close unexpectedly called")
	}
	return self._close()
}

type receiverMock struct {
	t *testing.T
	receive func(Event)
	_close func()
}

func (self *receiverMock) Receive(evt Event) {
	if self.receive == nil {
		self.t.Fatal("Receive unexpectedly called")
	}
	self.receive(evt)
}

func (self *receiverMock) Close() {
	if self._close == nil {
		self.t.Fatal("Close unexpectedly called")
	}
	self._close()
}

func TestSend(t *testing.T) {
	stack := make([]Event, 0, 1)
	spy := func(evt Event) error {
		stack = append(stack, evt)
		return nil
	}
	DefaultManager = &managerMock{t: t, send: spy}
	if err := Send(27); err != nil {
		t.Fatal(err)
	}
	if len(stack) != 1 {
		t.Fatalf("Wrong number of call to DefaultManager.Send (%d)", len(stack))
	}
	if i, ok := stack[0].(int); !ok || i != 27 {
		t.Errorf("Got %v. Expect 27.", stack[0])
	}
}

func TestAddReceiver(t *testing.T) {
	stack := make([]Receiver, 0, 1)
	spy := func(rcv Receiver) error {
		stack = append(stack, rcv)
		return nil
	}
	DefaultManager = &managerMock{t: t, addReceiver: spy}
	if err := AddReceiver(&receiverMock{t: t}); err != nil {
		t.Fatal(err)
	}
	if len(stack) != 1 {
		t.Fatalf("Wrong number of call to DefaultManager.Send (%d)", len(stack))
	}
	if r, ok := stack[0].(*receiverMock); !ok || r.t != t {
		t.Errorf("Got unexpected %v.", stack[0])
	}
}
