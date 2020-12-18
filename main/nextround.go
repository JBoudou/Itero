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
	lastCheck time.Time
	adjust    time.Duration
	step      uint8
	warn      *log.Logger
}

func newNextRound() *nextRound {
	return &nextRound{
		adjust: 10 * time.Minute,
		warn:   log.New(os.Stderr, "nextRound", log.LstdFlags|log.Lshortfile|log.Lmsgprefix),
	}
}

func (self *nextRound) fullCheck() error {
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
		qSelectNext = `
		  SELECT p.Id
		    FROM Polls AS p LEFT OUTER JOIN Participants AS a ON p.Id = a.Poll
		  WHERE p.Active AND p.CurrentRound < p.MaxNbRounds
		  GROUP BY p.Id,
		        p.CurrentRoundStart, p.MaxRoundDuration, p.CurrentRound, p.Publicity, p.RoundThreshold
		 HAVING ADDTIME(p.CurrentRoundStart, p.MaxRoundDuration) <= CURRENT_TIMESTAMP()
		        OR ( (p.CurrentRound > 0 OR p.Publicity = ?)
		              AND ( (p.RoundThreshold = 0 AND SUM(a.LastRound = p.CurrentRound) > 0)
		                   OR ( p.RoundThreshold > 0
		                        AND SUM(a.LastRound = p.CurrentRound) / COUNT(a.LastRound) >= p.RoundThreshold )))
		    FOR UPDATE`
		qNextRound = `UPDATE Polls SET CurrentRound = CurrentRound + 1 WHERE Id = ?`
	)
	nextSet := make(map[uint32]bool)
	if err := collectUI32Id(nextSet, tx, qSelectNext, db.PollPublicityInvited); err != nil {
		return err
	}
	if err := execOnUI32Id(nextSet, tx, qNextRound); err != nil {
		return err
	}

	// Commit
	if err := tx.Commit(); err != nil {
		return err
	}
	commited = true

	// Send events
	for id := range nextSet {
		events.Send(NextRoundEvent{id})
	}

	log.Print("nextRound fullCkeck terminated.")
	return nil
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
	                 OR ( (p.CurrentRound > 0 OR p.Publicity = ?)
	                       AND ( (p.RoundThreshold = 0 AND SUM(a.LastRound = p.CurrentRound) > 0)
	                            OR ( p.RoundThreshold > 0
	                                 AND SUM(a.LastRound = p.CurrentRound) / COUNT(a.LastRound) >= p.RoundThreshold )))
	         )`

	result, err := db.DB.Exec(qUpdate, pollId, db.PollPublicityInvited)
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
		events.Send(NextRoundEvent{pollId})
	}

	log.Printf("nextRound check for poll %d terminated.", pollId)
	return nil
}

func (self *nextRound) updateLastCheck() {
	row := db.DB.QueryRow(`SELECT CURRENT_TIMESTAMP()`)
	if err := row.Scan(&self.lastCheck); err != nil {
		self.warn.Print(err)
	}
	self.adjust = (self.adjust + time.Since(self.lastCheck)) / 2
}

func (self *nextRound) nextAlarm() (ret alarm.Event) {
	const (
		qNext = `
		  SELECT Id, ADDTIME(CurrentRoundStart, MaxRoundDuration) AS Next, CURRENT_TIMESTAMP()
			  FROM Polls
		   WHERE Active HAVING Next >= ?
		   ORDER BY Next ASC LIMIT 1`
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
		ret.Time = time.Now().Add(nextRoundDefaultWaitDuration)
	}

	log.Printf("Next nextRound alarm at %v.", ret.Time)
	return
}

func (self *nextRound) run() {
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

			if makeFullCheck {
				self.step++
			} else {
				self.step = 0
			}

			at.Send <- self.nextAlarm()
		}
	}
}

func StartNextRound() {
	go newNextRound().run()
}
