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

package alarm

import (
	"reflect"
	"time"
)

type Event struct {
	Time time.Time
	Data interface{}
	// Before resending the event, Remaining is set by Alarm to the number of remaining waiting events.
	Remaining int
}

// Alarm resends events at the requested time, or later.
// Closing the Send channel asks the alarm to terminate, which is notified by closing Receive.
type Alarm struct {
	Send    chan<- Event
	Receive <-chan Event

	discardLaterEvent     bool
	discardDuplicates     bool
	discardLateDuplicates bool
}

func allAfter(waiting map[Event]bool, evt Event) bool {
	for w := range waiting {
		if w.Time.Before(evt.Time) {
			return false
		}
	}
	return true
}

func noDuplicate(waiting map[Event]bool, evt Event) bool {
	for w := range waiting {
		if reflect.DeepEqual(w.Data, evt.Data) {
			return false
		}
	}
	return true
}

func noLateDuplicate(waiting map[Event]bool, evt Event) bool {
	for w := range waiting {
		if w.Time.Before(evt.Time) && reflect.DeepEqual(w.Data, evt.Data) {
			return false
		}
	}
	return true
}

func wait(send chan<- Event, evt Event) {
	duration := time.Until(evt.Time)
	if duration > 0 {
		time.Sleep(duration)
	}
	send <- evt
}

func (self Alarm) run(rcv <-chan Event, send chan<- Event) {
	waiting := make(map[Event]bool, 1)
	tick := make(chan Event)
	closing := false

mainLoop:
	for true {
		select {

		case evt, ok := <-rcv:
			if !ok {
				closing = true
			} else if (!self.discardLaterEvent || allAfter(waiting, evt)) &&
				(!self.discardDuplicates || noDuplicate(waiting, evt)) &&
				(! self.discardLateDuplicates || noLateDuplicate(waiting, evt)) {
				waiting[evt] = true
				go wait(tick, evt)
			}

		case evt := <-tick:
			delete(waiting, evt)
			evt.Remaining = len(waiting)
			send <- evt
			if closing && len(waiting) == 0 {
				break mainLoop
			}

		}
	}

	close(tick)
	close(send)
}

type Option func(evt *Alarm)

var (
	// Events received by Alarm with Time greater than any waiting event are ignored.
	DiscardLaterEvent Option = func(evt *Alarm) {
		evt.discardLaterEvent = true
	}

	// Event received with same Data as any waiting event are ignored.
	DiscardDuplicates Option = func(evt *Alarm) {
		evt.discardDuplicates = true
	}

	// Event received with same Data and later Time than some waiting event are ignored.
	DiscardLateDuplicates Option = func(evt *Alarm) {
		evt.discardLateDuplicates = true
	}
)

// New creates a new Alarm with the given size for Send.
// Receive is always unbuffered. Thus, settings chanSize to zero may result in deadlocks.
func New(chanSize int, opts ...Option) Alarm {
	in := make(chan Event, chanSize)
	out := make(chan Event)
	alarm := Alarm{Send: in, Receive: out}
	for _, opt := range opts {
		opt(&alarm)
	}
	go alarm.run(in, out)
	return alarm
}
