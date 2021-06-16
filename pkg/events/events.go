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

// Package events provides types to implement busses of events in an application.
// Senders can send events on these busses and receiver receive the events. 
// What differenciate such a bus from a Go channel is that each event is dispatched to all
// receiver.
//
// By providing a central busses to different parts of the application, these parts can communicate
// together without knowing each other.
package events

import (
	"errors"
)

type Event interface {
}

// Receiver receives events from a Manager.
//
// It can be assumed that the methods of a given Receiver are always called from the same goroutine
// by the Manager.
type Receiver interface {
	Receive(Event)
	Close()
}

// Manager dispatches events from senders to Receivers.
//
// All its method must be goroutine-safe.
type Manager interface {

	// Send dispatch an event to all the Receivers of the Manager.
	// It takes ownership of the event.
	Send(Event) error
	
	// AddReceiver registers a Receiver to receive all events sent to the Manager.
	// If the Receiver is added to more than one Manager, it must be goroutine-safe.
	AddReceiver(Receiver) error
	
	// Close stops the Manager. Any subsequent Send will result in an error.
	// The Close method of all the Receivers of the Manager will eventually be called.
	Close() error
}

// Errors
var (
	ManagerClosedError = errors.New("The manager is closed")
)

const DefaultManagerChannelSize = 256

// DefaultManager is the manager used by direct functions like Send.
//
// It is initialised by a call to NewAsyncManager whith a channel size of DefaultManagerChannelSize.
// It shoudl not be assigned after any receiver has been added or any event has been sent.
var DefaultManager Manager

func init() {
	DefaultManager = NewAsyncManager(DefaultManagerChannelSize)
}

// Send dispatch an event to all the Receivers of DefaultManager.
// It takes ownership of the event.
func Send(evt Event) error {
	return DefaultManager.Send(evt)
}

// AddReceiver registers a Receiver to receive all events sent to DefaultManager.
// If the Receiver is added to another Manager, it must be goroutine-safe.
func AddReceiver(rcv Receiver) error {
	return DefaultManager.AddReceiver(rcv)
}

// ReceiverFunc turn a function into a stateless Receiver, whose Close method does nothing.
type ReceiverFunc func(Event)

func (self ReceiverFunc) Receive(evt Event) {
	self(evt)
}

func (self ReceiverFunc) Close() {
}

// AsyncForwarder is a simple Receiver that forward events through a channel.
type AsyncForwarder struct {
	// Filter returns true for events that must be forwarded through the channel.
	Filter func(Event) bool
	Chan chan<- Event
}

func (self AsyncForwarder) Receive(evt Event) {
	if self.Filter(evt) {
		self.Chan <- evt
	}
}

func (self AsyncForwarder) Close() {
	close(self.Chan)
}
