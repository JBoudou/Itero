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

// Package alarm defines the Alarm type that resend events at given times.
package alarm

import (
	"reflect"
	"time"
)

// Events is the structure that is sent to and received from Alarm.
type Event struct {
	// Time indicates when the Event must be resent by Alarms.
	Time time.Time
	// Data is untouched by Alarm.
	Data interface{}
	// Remaining is set by Alarm to the number of remaining waiting events, when the event is resent.
	Remaining int
}

// Alarm resends events at the requested time, or later.
// Closing the Send channel asks the alarm to terminate, which is notified by closing Receive.
type Alarm struct {
	Send    chan<- Event
	Receive <-chan Event
}

// Option is the type of options given as optional arguments to New.
type Option func(evt *alarmLogic)

var (
	// DiscardLaterEvent is an Option to ignore Events received by Alarm with Time greater than any
	// waiting event.
	DiscardLaterEvent Option = func(evt *alarmLogic) {
		evt.discardLaterEvent = true
	}

	// DiscardDuplicates is an Option to ignore Events received with tsame Data as any waiting event.
	DiscardDuplicates Option = func(evt *alarmLogic) {
		evt.discardDuplicates = true
	}

	// DiscardLateDuplicates is an Option to ignore Events received with same Data and later Time than
	// some waiting event.
	DiscardLateDuplicates Option = func(evt *alarmLogic) {
		evt.discardLateDuplicates = true
	}
)

// New creates a new Alarm with the given size for Send.
// Receive is always unbuffered. Thus, settings chanSize to zero may result in deadlocks.
func New(chanSize int, opts ...Option) Alarm {
	alarm, logic := newAlarmLogic(chanSize, opts...)
	go run(logic)
	return alarm
}

//
// Logic
//

// alarmLogic encapsulates all the logic of alarm, except the timing part.
type alarmLogic struct {
	receive <-chan Event
	send    chan<- Event

	discardLaterEvent     bool
	discardDuplicates     bool
	discardLateDuplicates bool

	waiting map[Event]bool
}

func newAlarmLogic(chanSize int, opts ...Option) (alarm Alarm, logic *alarmLogic) {
	in := make(chan Event, chanSize)
	out := make(chan Event)
	alarm = Alarm{Send: in, Receive: out}
	logic = &alarmLogic{receive: in, send: out, waiting: make(map[Event]bool)}
	for _, opt := range opts {
		opt(logic)
	}
	return
}

func (self *alarmLogic) allAfter(evt Event) bool {
	for w := range self.waiting {
		if !w.Time.After(evt.Time) {
			return false
		}
	}
	return true
}

func (self *alarmLogic) noDuplicate(evt Event) bool {
	for w := range self.waiting {
		if reflect.DeepEqual(w.Data, evt.Data) {
			return false
		}
	}
	return true
}

func (self *alarmLogic) noLateDuplicate(evt Event) bool {
	for w := range self.waiting {
		if !w.Time.After(evt.Time) && reflect.DeepEqual(w.Data, evt.Data) {
			return false
		}
	}
	return true
}

// AddEvent checks whether an incoming Event must be added to the waiting list.
// If it is the case, it is added to self.waiting.
func (self *alarmLogic) AddEvent(evt Event) bool {
	active, ok := self.waiting[evt]
	if (!ok || !active) &&
		(!self.discardLaterEvent || self.allAfter(evt)) &&
		(!self.discardDuplicates || self.noDuplicate(evt)) &&
		(!self.discardLateDuplicates || self.noLateDuplicate(evt)) {
		self.waiting[evt] = true
		return true
	}
	return false
}

// ResendEvent resend an event from the waiting list.
// It returns whether the event was active.
// If found, the event is removed from self.waiting.
func (self *alarmLogic) ResendEvent(evt Event) bool {
	active, ok := self.waiting[evt]
	if ok {
		delete(self.waiting, evt)
	}
	evt.Remaining = len(self.waiting)
	self.send <- evt
	return ok && active
}

func (self *alarmLogic) Close() {
	close(self.send)
}

//
// Runner
//

func wait(send chan<- Event, evt Event) {
	duration := time.Until(evt.Time)
	if duration > 0 {
		time.Sleep(duration)
	}
	send <- evt
}

func run(logic *alarmLogic) {
	tick := make(chan Event)
	closing := false

mainLoop:
	for true {
		select {

		case evt, ok := <-logic.receive:
			if !ok {
				closing = true
			} else if logic.AddEvent(evt) {
				go wait(tick, evt)
			}

		case evt := <-tick:
			logic.ResendEvent(evt)
			if closing && len(logic.waiting) == 0 {
				break mainLoop
			}

		}
	}

	close(tick)
	logic.Close()
}
