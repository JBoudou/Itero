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
	"log"
	"time"

	"github.com/JBoudou/Itero/alarm"
	"github.com/JBoudou/Itero/events"
)

// ClosePollEvent is type of events send when a poll is closed.
type ClosePollEvent struct {
	Poll uint32
}

const (
	// The time to wait when there seems to be no forthcoming deadline.
	closePollDefaultWaitDuration = 12 * time.Hour
	// Run fullCheck instead of checkOne once every closePollFullCheckFreq steps.
	closePollFullCheckFreq = 7
)

// closePoll represents a running closePoll service.
type closePoll struct {
	pollService
}

func newClosePoll() *closePoll {
	return &closePoll{
		pollService: newPollService("closePoll", func(pollId uint32) events.Event {
			return ClosePollEvent{pollId};
		}),
	}
}

func (self *closePoll) fullCheck() error {
	const (
		qSelectClose = `
		  SELECT Id
		    FROM Polls
		  WHERE Active
		    AND ( CurrentRound >= MaxNbRounds
		          OR (CurrentRound >= MinNbRounds AND Deadline <= CURRENT_TIMESTAMP) )
		    FOR UPDATE`
		qClosePoll = `UPDATE Polls SET Active = false WHERE Id = ?`
	)
	return self.fullCheck_helper(qSelectClose, qClosePoll)
}

func (self *closePoll) checkOne(pollId uint32) error {
	const qUpdate = `
	  UPDATE Polls SET Active = false
	   WHERE Id = ? AND Active
	     AND ( CurrentRound >= MaxNbRounds
	           OR (CurrentRound >= MinNbRounds AND Deadline <= CURRENT_TIMESTAMP) )`
	return self.checkOne_helper(qUpdate, pollId)
}

func (self *closePoll) nextAlarm() alarm.Event {
	const (
		qNext = `
		  SELECT Id, Deadline, CURRENT_TIMESTAMP() FROM Polls
		   WHERE Active AND Deadline >= ?
		   ORDER BY Deadline ASC LIMIT 1`
	)
	return self.nextAlarm_helper(qNext, closePollDefaultWaitDuration)
}

func (self *closePoll) run(evtChan <-chan events.Event) {
	at := alarm.New(1)

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
			makeFullCheck := self.step >= closePollFullCheckFreq || evt.Data == nil
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
			if evtChan != nil {
				at.Send <- self.nextAlarm()
			}

		case evt, ok := <-evtChan:
			if !ok {
				log.Print("Event manager closing makes closePoll to close too.")
				close(at.Send)
				evtChan = nil
				continue
			}
			nextEvt := evt.(NextRoundEvent)

			var err error
			makeFullCheck := self.step >= closePollFullCheckFreq
			if makeFullCheck {
				err = self.fullCheck()
			} else {
				err = self.checkOne(nextEvt.Poll)
			}
			if err != nil {
				self.warn.Print(err)
				continue
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
			_, ok := evt.(NextRoundEvent)
			return ok
		},
		Chan: ch,
	})
	go newClosePoll().run(ch)
}
