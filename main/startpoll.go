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
	"database/sql"
	"log"
	"time"

	"github.com/JBoudou/Itero/alarm"
	"github.com/JBoudou/Itero/db"
	"github.com/JBoudou/Itero/events"
)

// StartPollEvent is send when a poll is started.
type StartPollEvent struct {
	Poll uint32
}

const (
	// The time to wait when there seems to be no waiting poll.
	startPollDefaultWaitDuration = time.Hour
	// Run fullCheck instead of checkOne once every startPollFullCheckFreq steps.
	startPollFullCheckFreq = 3
)

// startPoll represents a running startPoll service.
type startPoll struct {
	pollService
}

func newStartPoll() *startPoll {
	return &startPoll{
		pollService: newPollService("startPoll", func(pollId uint32) events.Event {
			return StartPollEvent{pollId}
		}),
	}
}

func (self *startPoll) fullCheck() error {
	const (
		qSelectStart = `
		  SELECT Id
		    FROM Polls
		   WHERE State  = 'Waiting'
 		     AND Start <= CURRENT_TIMESTAMP
		     FOR UPDATE`
		qStartPoll = `UPDATE Polls SET State = 'Active' WHERE Id = ?`
	)
	return self.fullCheck_helper(qSelectStart, qStartPoll)
}

func (self *startPoll) checkOne(pollId uint32) error {
	const qUpdate = `
	  UPDATE Polls SET State = 'Active'
	   WHERE Id = ? AND State = 'Waiting'
	     AND Start <= CURRENT_TIMESTAMP`
	return self.checkOne_helper(pollId, func() (sql.Result, error) {
		return db.DB.Exec(qUpdate, pollId)
	})
}

func (self *startPoll) nextAlarm() alarm.Event {
	const (
		qNext = `
		  SELECT Id, Start, CURRENT_TIMESTAMP() FROM Polls
		   WHERE State = 'Waiting' AND Start >= ?
		   ORDER BY Start ASC LIMIT 1`
	)
	return self.nextAlarm_helper(qNext, startPollDefaultWaitDuration)
}

func (self *startPoll) run(evtChan <-chan events.Event) {
	at := alarm.New(1, alarm.DiscardLaterEvent)

	self.updateLastCheck()
	self.fullCheck()
	at.Send <- self.nextAlarm()

	for {
		select {
		case evt, ok := <-at.Receive:
			if !ok {
				self.warn.Print("Alarm closed. Stopping.")
				break
			}

			self.updateLastCheck()

			var err error
			makeFullCheck := self.step >= startPollFullCheckFreq || evt.Data == nil
			if makeFullCheck {
				err = self.fullCheck()
			} else {
				err = self.checkOne(evt.Data.(uint32))
			}
			if err != nil {
				self.warn.Print(err)
				continue
			}

			// Do not send if the channel has been closed.
			if evt.Remaining == 0 && evtChan != nil {
				at.Send <- self.nextAlarm()
			}

		case evt, ok := <-evtChan:
			if !ok {
				log.Print("Event manager closing makes startPoll to close too.")
				close(at.Send)
				evtChan = nil
				continue
			}

			switch evt.(type) {
			case CreatePollEvent:
				at.Send <- self.nextAlarm()
			}
		}
	}
}

// StartStartPoll starts the startPoll service.
//
// The startPoll service receives CreatePollEvent, start polls when they need to be, and send
// StartPollEvents.
func StartStartPoll() {
	ch := make(chan events.Event, 16)
	events.AddReceiver(events.AsyncForwarder{
		Filter: func(evt events.Event) bool {
			switch evt.(type) {
			case CreatePollEvent:
				return true
			}
			return false
		},
		Chan: ch,
	})
	go newStartPoll().run(ch)
}
