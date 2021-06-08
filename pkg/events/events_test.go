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

package events_test

import (
	"testing"

	. "github.com/JBoudou/Itero/pkg/events"
	. "github.com/JBoudou/Itero/pkg/events/eventstest"
)


func TestSend(t *testing.T) {
	stack := make([]Event, 0, 1)
	spy := func(evt Event) error {
		stack = append(stack, evt)
		return nil
	}
	DefaultManager = &ManagerMock{T: t, Send_: spy}
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
	DefaultManager = &ManagerMock{T: t, AddReceiver_: spy}
	if err := AddReceiver(&ReceiverMock{T: t}); err != nil {
		t.Fatal(err)
	}
	if len(stack) != 1 {
		t.Fatalf("Wrong number of call to DefaultManager.Send (%d)", len(stack))
	}
	if r, ok := stack[0].(*ReceiverMock); !ok || r.T != t {
		t.Errorf("Got unexpected %v.", stack[0])
	}
}
