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
	"fmt"
	"log"
	"time"

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
	      FROM Polls AS p
	      LEFT OUTER JOIN Participants_Round_Count AS r ON (p.Id, p.CurrentRound) = (r.Poll, r.Round)
	      LEFT OUTER JOIN Participants_Poll_Count  AS a ON p.Id = a.Poll
	     WHERE p.State = 'Active' AND p.CurrentRound < p.MaxNbRounds
	       AND (   ( RoundDeadline(p.CurrentRoundStart, p.MaxRoundDuration, p.Deadline,
	                               p.CurrentRound, p.MinNbRounds) <= CURRENT_TIMESTAMP()
	                 AND ( p.CurrentRound > 0 OR r.Count > 2 ))
	            OR (    (p.CurrentRound > 0 OR p.Publicity = %d)
	                AND (   (p.RoundThreshold = 0 AND r.Count > 0)
	                     OR ( p.RoundThreshold > 0
	                          AND r.Count / a.Count >= p.RoundThreshold ) )
	                AND (   (p.CurrentRound + 1 < MinNbRounds)
	                     OR p.Deadline IS NULL
	                     OR (ADDTIME(CURRENT_TIMESTAMP(), p.MaxRoundDuration) < p.Deadline)
	                     OR (p.Deadline < CURRENT_TIMESTAMP()) )))
	       FOR UPDATE`
		qNextRound = `UPDATE Polls SET CurrentRound = CurrentRound + 1 WHERE Id = ?`
	)
	return self.fullCheck_helper(fmt.Sprintf(qSelectNext, db.PollPublicityInvited), qNextRound)
}

func (self *nextRound) checkOne(pollId uint32) error {
	// MariaDB 5.5.68 does not allows UPDATE subqueries to reference the updated table.

	const (
		qCheck = `
	    SELECT p.Id
	      FROM Polls AS p
	      LEFT OUTER JOIN Participants_Round_Count AS r ON (p.Id, p.CurrentRound) = (r.Poll, r.Round)
	      LEFT OUTER JOIN Participants_Poll_Count  AS a ON p.Id = a.Poll
	     WHERE p.Id = ? AND p.State = 'Active' AND p.CurrentRound < p.MaxNbRounds
	       AND (   ( RoundDeadline(p.CurrentRoundStart, p.MaxRoundDuration, p.Deadline,
	                               p.CurrentRound, p.MinNbRounds) <= CURRENT_TIMESTAMP()
	                 AND ( p.CurrentRound > 0 OR r.Count > 2 ))
	            OR (    (p.CurrentRound > 0 OR p.Publicity = ?)
	                AND (   (p.RoundThreshold = 0 AND r.Count > 0)
	                     OR ( p.RoundThreshold > 0
	                          AND r.Count / a.Count >= p.RoundThreshold ) )
	                AND (   (p.CurrentRound + 1 < MinNbRounds)
	                     OR p.Deadline IS NULL
	                     OR (ADDTIME(CURRENT_TIMESTAMP(), p.MaxRoundDuration) < p.Deadline)
	                     OR (p.Deadline < CURRENT_TIMESTAMP()) )))
	       FOR UPDATE`
		qUpdate = `
	    UPDATE Polls SET CurrentRound = CurrentRound + 1
	     WHERE Id = ?`
	)

	rows, err := db.DB.Query(qCheck, pollId, db.PollPublicityInvited)

	if err == nil {
		err = self.checkOne_helper(pollId, func() (sql.Result, error) {
			return db.DB.Exec(qUpdate, sql.NullInt64{Int64: int64(pollId), Valid: rows.Next()})
		})
	}

	return err
}

func (self *nextRound) nextAlarm() alarm.Event {
	const (
		qNext = `
		  SELECT Id, RoundDeadline(CurrentRoundStart, MaxRoundDuration, Deadline, CurrentRound, MinNbRounds) AS Next,
		         CURRENT_TIMESTAMP()
		    FROM Polls
		   WHERE State = 'Active' HAVING Next >= ?
		   ORDER BY Next ASC LIMIT 1`
	)
	return self.nextAlarm_helper(qNext, nextRoundDefaultWaitDuration)
}

func (self *nextRound) run(evtChan <-chan events.Event) {
	// DiscardDuplicates is needed because otherwise an event may be added after each vote (because of
	// adjusted time).
	at := alarm.New(1, alarm.DiscardLaterEvent, alarm.DiscardDuplicates)

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

			if evt.Remaining == 0 && evtChan != nil {
				at.Send <- self.nextAlarm()
			}

		case evt, ok := <-evtChan:
			if !ok {
				log.Print("Event manager closing makes nextRound to close too.")
				close(at.Send)
				evtChan = nil
				continue
			}

			switch typed := evt.(type) {
			case VoteEvent:
				if err := self.checkOne(typed.Poll); err != nil {
					self.warn.Print(err)
				}
				at.Send <- self.nextAlarm()

			case CreatePollEvent, StartPollEvent:
				at.Send <- self.nextAlarm()
			}
		}
	}
}

// StartNextRound starts the nextRound service.
func StartNextRound() {
	ch := make(chan events.Event, 64)
	events.AddReceiver(events.AsyncForwarder{
		Filter: func(evt events.Event) bool {
			switch evt.(type) {
			case VoteEvent, CreatePollEvent, StartPollEvent:
				return true;
			}
			return false
		},
		Chan: ch,
	})
	go newNextRound().run(ch)
}
