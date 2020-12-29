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
	"fmt"
	"time"
	"log"

	"github.com/JBoudou/Itero/alarm"
	"github.com/JBoudou/Itero/db"
	"github.com/JBoudou/Itero/events"
)

// NextRoundEvent is the type of events send when a new round starts.
type NextRoundEvent struct {
	Poll uint32
}

const (
	// The time to wait when there seems to be no forthcoming deadline.
	nextRoundDefaultWaitDuration = time.Hour
	// Run fullCheck instead of checkOne once every nextRoundFullCheckFreq steps.
	nextRoundFullCheckFreq = 5
)

type nextRound struct {
	pollService
}

func newNextRound() *nextRound {
	return &nextRound{
		pollService: newPollService("nextRound", func(pollId uint32) events.Event {
			return NextRoundEvent{pollId}
		}),
	}
}

func (self *nextRound) fullCheck() error {
	const (
		qSelectNext = `
		  SELECT p.Id
		    FROM Polls AS p LEFT OUTER JOIN Participants AS a ON p.Id = a.Poll
		  WHERE p.Active AND p.CurrentRound < p.MaxNbRounds
		  GROUP BY p.Id,
		        p.CurrentRoundStart, p.MaxRoundDuration, p.CurrentRound, p.Publicity, p.RoundThreshold
		 HAVING ADDTIME(p.CurrentRoundStart, p.MaxRoundDuration) <= CURRENT_TIMESTAMP()
		        OR ( (p.CurrentRound > 0 OR p.Publicity = %d)
		              AND ( (p.RoundThreshold = 0 AND SUM(a.LastRound = p.CurrentRound) > 0)
		                   OR ( p.RoundThreshold > 0
		                        AND SUM(a.LastRound = p.CurrentRound) / COUNT(a.LastRound) >= p.RoundThreshold )))
		    FOR UPDATE`
		qNextRound = `UPDATE Polls SET CurrentRound = CurrentRound + 1 WHERE Id = ?`
	)
	return self.fullCheck_helper(fmt.Sprintf(qSelectNext, db.PollPublicityInvited), qNextRound)
}

func (self *nextRound) checkOne(pollId uint32) error {
	const qUpdate = `
	  UPDATE Polls SET CurrentRound = CurrentRound + 1
	   WHERE Id IN (
	           SELECT p.Id
	             FROM Polls AS p LEFT OUTER JOIN Participants AS a ON p.Id = a.Poll
	           WHERE p.Id = ? AND p.Active AND p.CurrentRound < p.MaxNbRounds
	           GROUP BY p.Id,
	                 p.CurrentRoundStart, p.MaxRoundDuration, p.CurrentRound, p.Publicity, p.RoundThreshold
	          HAVING ADDTIME(p.CurrentRoundStart, p.MaxRoundDuration) <= CURRENT_TIMESTAMP()
	                 OR ( (p.CurrentRound > 0 OR p.Publicity = %d)
	                       AND ( (p.RoundThreshold = 0 AND SUM(a.LastRound = p.CurrentRound) > 0)
	                            OR ( p.RoundThreshold > 0
	                                 AND SUM(a.LastRound = p.CurrentRound) / COUNT(a.LastRound) >= p.RoundThreshold )))
	         )`
	return self.checkOne_helper(fmt.Sprintf(qUpdate, db.PollPublicityInvited), pollId)
}

func (self *nextRound) nextAlarm() alarm.Event {
	const (
		qNext = `
		  SELECT Id, ADDTIME(CurrentRoundStart, MaxRoundDuration) AS Next, CURRENT_TIMESTAMP()
			  FROM Polls
		   WHERE Active HAVING Next >= ?
		   ORDER BY Next ASC LIMIT 1`
	)
	return self.nextAlarm_helper(qNext, nextRoundDefaultWaitDuration)
}

func (self *nextRound) run(evtChan <-chan events.Event) {
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
			makeFullCheck := self.step >= nextRoundFullCheckFreq || evt.Data == nil
			if makeFullCheck {
				err = self.fullCheck()
			} else {
				err = self.checkOne(evt.Data.(uint32))
			}
			if err != nil {
				self.warn.Print(err)
				continue
			}

			at.Send <- self.nextAlarm()

		case evt, ok := <-evtChan:
			if !ok {
				log.Print("Event manager closing makes nextRound to close too.")
				close(at.Send)
				evtChan = nil
				continue
			}

			voteEvt := evt.(VoteEvent)
			if err := self.checkOne(voteEvt.Poll); err != nil {
				self.warn.Print(err)
				continue
			}
		}
	}
}

// StartNextRound starts the nextRound service.
func StartNextRound() {
	ch := make(chan events.Event, 64)
	events.AddReceiver(events.AsyncForwarder{
		Filter: func(evt events.Event) bool {
			_, ok := evt.(VoteEvent)
			return ok
		},
		Chan: ch,
	})
	go newNextRound().run(ch)
}
