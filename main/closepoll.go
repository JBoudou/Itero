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

package main

import (
	"database/sql"

	"github.com/JBoudou/Itero/db"
	"github.com/JBoudou/Itero/events"
)

// ClosePollEvent is type of events send when a poll is closed.
type ClosePollEvent struct {
	Poll uint32
}

// closePoll represents a running closePoll service.
type closePoll struct {
	pollService
}

func newClosePoll() *closePoll {
	return &closePoll{
		pollService: newPollService("closePoll", func(pollId uint32) events.Event {
			return ClosePollEvent{pollId}
		}),
	}
}

func (self *closePoll) fullCheck() error {
	const (
		qSelectClose = `
		  SELECT Id
		    FROM Polls
		  WHERE State = 'Active'
		    AND ( CurrentRound >= MaxNbRounds
		          OR (CurrentRound >= MinNbRounds AND Deadline <= CURRENT_TIMESTAMP) )
		    FOR UPDATE`
		qClosePoll = `UPDATE Polls SET State = 'Terminated' WHERE Id = ?`
	)
	return self.fullCheck_helper(qSelectClose, qClosePoll)
}

func (self *closePoll) checkOne(pollId uint32) error {
	const qUpdate = `
	  UPDATE Polls SET State = 'Terminated'
	   WHERE Id = ? AND State = 'Active'
	     AND ( CurrentRound >= MaxNbRounds
	           OR (CurrentRound >= MinNbRounds AND Deadline <= CURRENT_TIMESTAMP) )`
	return self.checkOne_helper(pollId, func() (sql.Result, error) {
		return db.DB.Exec(qUpdate, pollId)
	})
}

func (self *closePoll) run(evtChan <-chan events.Event) {
	self.updateLastCheck()
	self.fullCheck()

	for {
		select {
		case evt, ok := <-evtChan:
			if !ok {
				break
			}

			switch typed := evt.(type) {
			case NextRoundEvent:
				if err := self.checkOne(typed.Poll); err != nil {
					self.warn.Print(err)
				}
			}
		}
	}
}

// StartClosePoll starts the closePoll service.
//
// The closePoll service receives NextRoundEvents, closes polls when they need to be, and send
// ClosePollEvents.
func StartClosePoll() {
	ch := make(chan events.Event, 16)
	events.AddReceiver(events.AsyncForwarder{
		Filter: func(evt events.Event) bool {
			switch evt.(type) {
			case NextRoundEvent:
				return true
			}
			return false
		},
		Chan: ch,
	})
	go newClosePoll().run(ch)
}
