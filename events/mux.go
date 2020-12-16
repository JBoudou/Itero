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
	"github.com/tevino/abool"
)

// mux is the default asynchronous Manager.
type mux struct {
	ch chan<- Event
	closed *abool.AtomicBool
}

func (self *mux) Send(evt Event) error {
	if self.closed.IsSet() {
		return ManagerClosedError
	}
	self.ch <- evt
	return nil
}

type addReceiverEvent Receiver

func (self *mux) AddReceiver(receiver Receiver) error {
	return self.Send(addReceiverEvent(receiver))
}

func (self *mux) Close() error {
	if self.closed.SetToIf(false, true) {
		close(self.ch)
	}
	return nil
}

// muxRun is the goroutine of mux.
func muxRun(ch <-chan Event) {
	receivers := make([]Receiver, 0, 4)
	
	// main loop
	for evt := range ch {
		switch typed := evt.(type) {
		case addReceiverEvent:
			receivers = append(receivers, Receiver(typed))
		default:
			for _, receiver := range receivers {
				receiver.Receive(evt)
			}
		}
	}

	// close loop
	for _, receiver := range receivers {
		receiver.Close()
	}
}

// NewAsyncManager creates a new asynchronous Manager.
//
// The returned Manager send the events over a channel of the size given as argument.
// All its Receivers receive the events in the same goroutine, in the order they've been added
// to the manager.
func NewAsyncManager(chanSize int) Manager {
	ch := make(chan Event, chanSize)
	go muxRun(ch)
	return &mux{ch: ch, closed: abool.New()}
}
