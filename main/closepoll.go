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
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/JBoudou/Itero/alarm"
	"github.com/JBoudou/Itero/db"
	"github.com/JBoudou/Itero/events"
)

type ClosePollEvent struct {
	Poll uint32
}

const (
	// The time to wait when there seems to be no forthcoming deadline.
	closePollDefaultWaitDuration = time.Hour
	// Run fullCheck instead of checkOne once every closePollFullCheckFreq steps.
	closePollFullCheckFreq = 5
)

type closePoll struct {
	lastCheck time.Time
	adjust    time.Duration
	step      uint8
	warn      *log.Logger
}

func newClosePoll() *closePoll {
	return &closePoll{
		adjust: 10 * time.Minute,
		warn:   log.New(os.Stderr, "closePoll", log.LstdFlags|log.Lshortfile|log.Lmsgprefix),
	}
}

func (self *closePoll) fullCheck() error {
	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}
	commited := false
	defer func() {
		if !commited {
			tx.Rollback()
		}
	}()

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
	closeSet := make(map[uint32]bool)
	if err := collectUI32Id(closeSet, tx, qSelectClose); err != nil {
		return err
	}
	if err := execOnUI32Id(closeSet, tx, qClosePoll); err != nil {
		return err
	}

	// Commit
	if err := tx.Commit(); err != nil {
		return err
	}
	commited = true

	// Send events
	for id := range closeSet {
		events.Send(ClosePollEvent{id})
	}

	log.Print("closeSet fullCkeck terminated.")
	return nil
}

func (self *closePoll) checkOne(pollId uint32) error {
	const qUpdate = `
	  UPDATE Polls SET Active = false
	   WHERE Id = ? AND Active
	     AND ( CurrentRound >= MaxNbRounds
	           OR (CurrentRound >= MinNbRounds AND Deadline <= CURRENT_TIMESTAMP) )`

	result, err := db.DB.Exec(qUpdate, pollId)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected > 1 {
		return errors.New(fmt.Sprintf("More than one poll with Id %d. No event send.", pollId))
	}
	if affected == 1 {
		events.Send(ClosePollEvent{pollId})
	}

	log.Printf("closePoll check for poll %d terminated.", pollId)
	return nil
}

func (self *closePoll) updateLastCheck() {
	row := db.DB.QueryRow(`SELECT CURRENT_TIMESTAMP()`)
	if err := row.Scan(&self.lastCheck); err != nil {
		self.warn.Print(err)
	}
	self.adjust = (self.adjust + time.Since(self.lastCheck)) / 2
}

func (self *closePoll) nextAlarm() (ret alarm.Event) {
	const (
		qNext = `
		  SELECT Id, Deadline, CURRENT_TIMESTAMP() FROM Polls
		   WHERE Active AND Deadline >= ?
		   ORDER BY Deadline ASC LIMIT 1`
	)

	var pollId uint32
	var timestamp time.Time
	row := db.DB.QueryRow(qNext, self.lastCheck)
	switch err := row.Scan(&pollId, &ret.Time, &timestamp); {
	case err == nil:
		self.adjust = (self.adjust + time.Since(timestamp)) / 2
		ret.Time = ret.Time.Add(self.adjust)
		ret.Data = pollId
	default:
		if !errors.Is(err, sql.ErrNoRows) {
			self.warn.Print(err)
		}
		ret.Time = time.Now().Add(closePollDefaultWaitDuration)
	}

	log.Printf("Next closePoll alarm at %v.", ret.Time)
	return
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
			log.Printf("DEBUG closePoll received alarm %v.", evt)

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

			if makeFullCheck {
				self.step++
			} else {
				self.step = 0
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
			log.Printf("DEBUG closePoll received event %v.", evt)

			// TODO remove this useless check (should panic when the event has wrong type).
			var nextEvt NextRoundEvent
			if nextEvt, ok = evt.(NextRoundEvent); !ok {
				self.warn.Printf("Unhandled event %v.", evt)
				continue
			}

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

			if makeFullCheck {
				self.step++
			} else {
				self.step = 0
			}
		}
	}
}

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
