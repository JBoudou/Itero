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

package services

import (
	"time"

	"github.com/JBoudou/Itero/pkg/events"
	"github.com/JBoudou/Itero/pkg/ioc"
)

//
// PollNotifications
//

type PollNotifAction uint8

const (
	PollNotifStart PollNotifAction = iota
	PollNotifNext
	PollNotifTerm
	PollNotifDelete
)

type PollNotification struct {
	Timestamp    time.Time
	Id           uint32
	Action       PollNotifAction
	Round        uint8
	Title        string
	Participants map[uint32]bool
}

func NewPollNotification(evt events.Event) (ret *PollNotification) {
	ret = &PollNotification{Timestamp: time.Now()}

	switch e := evt.(type) {
	case StartPollEvent:
		ret.Id = e.Poll
		ret.Action = PollNotifStart

	case NextRoundEvent:
		ret.Id = e.Poll
		ret.Action = PollNotifNext
		ret.Round = e.Round

	case ClosePollEvent:
		ret.Id = e.Poll
		ret.Action = PollNotifTerm

	case DeletePollEvent:
		ret.Id = e.Poll
		ret.Action = PollNotifDelete
		ret.Title = e.Title
		ret.Participants = e.Participants
	}

	return
}

//
// PollNotif service
//

// PollNotifChannel provides lists of recent notifications.
type PollNotifChannel = <-chan []*PollNotification

// PollNotifDelay is the default duration during which notifications are kept in the list.
// This duration must be strictly greater than the period between pulls of the frontend.
const PollNotifDelay = 20 * time.Second

func init() {
	ioc.Root.Bind(func(evtManager events.Manager) (PollNotifChannel, error) {
		return RunPollNotif(PollNotifDelay, evtManager)
	})
}

func RunPollNotif(delay time.Duration, evtManager events.Manager) (PollNotifChannel, error) {
	runner := newPollNotifRunner(delay)

	eventChan := make(chan events.Event, 64)
	err := evtManager.AddReceiver(events.AsyncForwarder{
		Filter: runner.filter,
		Chan:   eventChan,
	})
	if err != nil {
		return nil, err
	}

	notifChan := make(chan []*PollNotification)
	go runner.run(eventChan, notifChan)
	return notifChan, nil
}

type pollNotifRunner struct {
	toKeep *PollNotifList
	toSend *PollNotifList
}

func newPollNotifRunner(delay time.Duration) *pollNotifRunner {
	base := NewPollNotifList(delay)
	return &pollNotifRunner{toKeep: base, toSend: base.Copy()}
}

func (self *pollNotifRunner) filter(evt events.Event) bool {
	switch evt.(type) {
	case StartPollEvent, NextRoundEvent, ClosePollEvent, DeletePollEvent:
		return true
	}
	return false
}

func (self *pollNotifRunner) run(eventChan <-chan events.Event, notifChan chan<- []*PollNotification) {
	for {
		select {
		case notifChan <- self.toSend.Slice():
			self.toSend = self.toKeep.Copy()

		case evt, ok := <-eventChan:
			if !ok {
				close(notifChan)
				return
			}
			notif := NewPollNotification(evt)
			self.toKeep.Add(notif)
			self.toSend.Add(notif)
		}
	}
}

//
// PollNotifList
//

// PollNotifList is a list of PollNotification that removes too old elements.
type PollNotifList struct {
	data  []*PollNotification
	first int
	delay time.Duration
}

func NewPollNotifList(delay time.Duration) *PollNotifList {
	return &PollNotifList{
		data:  make([]*PollNotification, 0, 4),
		first: 0,
		delay: delay,
	}
}

// Add adds a notification to the list.
// The list may be tidied.
func (self *PollNotifList) Add(notif *PollNotification) {
	if len(self.data) == cap(self.data) {
		self.Tidy()
	}
	if len(self.data) == cap(self.data) {
		if 2*self.first < len(self.data) {
			*self = *(self.Copy())
		} else {
			nlen := len(self.data) - self.first
			copy(self.data[0:nlen], self.data[self.first:])
			self.data = self.data[0:nlen]
			self.first = 0
		}
	}
	self.data = append(self.data, notif)
}

// Tidy removes too old notifications from the list.
func (self *PollNotifList) Tidy() {
	l := len(self.data)
	if l == 0 {
		return
	}

	limit := time.Now().Add(-1 * self.delay)
	for ; self.first < l && self.data[self.first].Timestamp.Before(limit); self.first++ {
		self.data[self.first] = nil
	}

	if self.first == l {
		self.first = 0
		self.data = self.data[0:0]
	}
}

// Slice return the list of notifications.
// The list is always tidied first.
func (self *PollNotifList) Slice() []*PollNotification {
	self.Tidy()
	return self.data[self.first:]
}

// Copy construct a copy of the current list.
// The original list is never tidied but the constructed one always is.
func (self *PollNotifList) Copy() *PollNotifList {
	nlen := len(self.data) - self.first
	ret := &PollNotifList{
		data:  make([]*PollNotification, nlen, 2*nlen),
		first: 0,
		delay: self.delay,
	}
	copy(ret.data[0:nlen], self.data[self.first:])
	return ret
}
