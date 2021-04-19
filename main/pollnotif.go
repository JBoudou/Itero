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

package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/JBoudou/Itero/db"
	"github.com/JBoudou/Itero/events"
	"github.com/JBoudou/Itero/server"
)

//
// PollNotifications
//

const (
	PollNotifStart = iota
	PollNotifNext
	PollNotifTerm
	PollNotifDelete
)

type PollNotification struct {
	Timestamp time.Time
	Id        uint32
	Round     uint8
	Action    uint8
}

func NewPollNotification(evt events.Event) (ret *PollNotification) {
	ret = &PollNotification{ Timestamp: time.Now() }

	switch e := evt.(type) {
	case CreatePollEvent:
		ret.Id = e.Poll
		ret.Action = PollNotifStart

	case NextRoundEvent:
		ret.Id = e.Poll
		ret.Round = e.Round
		ret.Action = PollNotifNext

	case ClosePollEvent:
		ret.Id = e.Poll
		ret.Action = PollNotifTerm

	case DeletePollEvent:
		ret.Id = e.Poll
		ret.Action = PollNotifDelete
	}

	return
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

//
// PollNotif "Service"
//

// PollNotifChannel provides lists of recent notifications.
var PollNotifChannel <-chan []*PollNotification

// PollNotifDelay is the time notifications are kept in the list.
// This duration must be strictly greater than the period between pulls of the frontend.
const PollNotifDelay = 20 * time.Second

func RunPollNotif(delay time.Duration) error {
	if PollNotifChannel != nil {
		return errors.New("RunPollNotif already called")
	}

	runner := newPollNotifRunner(delay)

	eventChan := make(chan events.Event, 64)
	err := events.AddReceiver(events.AsyncForwarder{
		Filter: runner.filter,
		Chan:   eventChan,
	})
	if err != nil {
		return err
	}

	notifChan := make(chan []*PollNotification)
	PollNotifChannel = notifChan

	go runner.run(eventChan, notifChan)
	return nil
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
	case CreatePollEvent, NextRoundEvent, ClosePollEvent, DeletePollEvent:
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
// PollNotifHandler
//

type PollNotifQuery struct {
	LastUpdate time.Time
}

type PollNotifAnswerEntry struct {
	Timestamp time.Time
	Segment   string
	Round     uint8
	Action    uint8
}

func PollNotifHandler(ctx context.Context, response server.Response, request *server.Request) {
	if request.User == nil {
		if request.SessionError != nil {
			must(request.SessionError)
		} else {
			panic(server.UnauthorizedHttpError("Unlogged user"))
		}
	}
	must(request.CheckPOST(ctx))
	
	var query PollNotifQuery
	err := request.UnmarshalJSONBody(&query)
	if err != nil {
		panic(server.WrapError(http.StatusBadRequest, "Bad request", err))
	}

	baseList := <-PollNotifChannel
	if len(baseList) == 0 {
		response.SendJSON(ctx, make([]PollNotifAnswerEntry, 0))
		return
	}

	const qCheck = `
	  SELECT Salt FROM Polls
		 WHERE Id = ?
		   AND (Admin = %[1]d OR Id IN ( SELECT Poll FROM Participants WHERE User = %[1]d ))`
	stmt, err := db.DB.PrepareContext(ctx, fmt.Sprintf(qCheck, request.User.Id))
	must(err)
	defer stmt.Close()

	answer := make([]PollNotifAnswerEntry, 0, len(baseList)/2)
	for _, notif := range baseList {
		if notif.Timestamp.Before(query.LastUpdate) {
			continue
		}

		rows, err := stmt.QueryContext(ctx, notif.Id)
		must(err)
		if !rows.Next() {
			continue
		}
		segment := PollSegment{Id: notif.Id}
		err = rows.Scan(&segment.Salt)
		rows.Close()
		must(err)
		encoded, err := segment.Encode()
		must(err)
		
		answer = append(answer, PollNotifAnswerEntry{
			Timestamp: notif.Timestamp,
			Segment: encoded,
			Round: notif.Round,
			Action: notif.Action,
		})
	}
			
	response.SendJSON(ctx, answer)
}
