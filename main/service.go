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
	"log"
	"os"
	"time"
	"fmt"

	"github.com/JBoudou/Itero/alarm"
	"github.com/JBoudou/Itero/db"
	"github.com/JBoudou/Itero/events"
)

const (
	// The time to wait when there seems to be no forthcoming deadline.
	pollServiceDefaultWaitDuration = time.Hour
	// Run fullCheck instead of checkOne once every pollServiceFullCheckFreq steps.
	pollServiceFullCheckFreq = 5
)

// pollService is the base classes of some services like nextRound and closePoll.
type pollService struct {
	lastCheck time.Time
	adjust    time.Duration
	step      uint8

	// Logger to use for warning. Will be replaced by a leveled logger in near future.
	warn        *log.Logger

	serviceName string
	makeEvent   func(pollId uint32) events.Event
}

// newPollService creates a new pollService.
func newPollService(serviceName string, makeEvent func(pollId uint32) events.Event) pollService {
	return pollService{
		adjust:      10 * time.Minute,
		serviceName: serviceName,
		warn:        log.New(os.Stderr, serviceName, log.LstdFlags|log.Lshortfile|log.Lmsgprefix),
		makeEvent:   makeEvent,
	}
}

// fullCheck_helper helps to implement full checks.
//
// Query qSelect selects all the ids (uint32) on which an update is needed. This query takes one
// parameter which is lastCheck.
// Query qUpdate is then called on each collected id. It takes the id as only parameter.
// If the query is successful, an event constructed by makeEvent is emitted and step is reset to 0.
func (self *pollService) fullCheck_helper(qSelect, qUpdate string) error {
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

	closeSet := make(map[uint32]bool)
	if err := self.collectId(closeSet, tx, qSelect); err != nil {
		return err
	}
	if err := self.execUpdate(closeSet, tx, qUpdate); err != nil {
		return err
	}

	// Commit
	if err := tx.Commit(); err != nil {
		return err
	}
	commited = true

	// Send events
	for id := range closeSet {
		events.Send(self.makeEvent(id))
	}

	self.step = 0
	log.Print(self.serviceName, " fullCkeck terminated.")
	return nil
}

// collectId is an internal (private) method.
func (self pollService) collectId(set map[uint32]bool, tx *sql.Tx, query string) error {
	rows, err := tx.Query(query)
	if err != nil {
		return err
	}
	for rows.Next() {
		var key uint32
		if err := rows.Scan(&key); err != nil {
			return nil
		}
		set[key] = true
	}
	return nil
}

// execUpdate is an internal (private) method
func (self pollService) execUpdate(set map[uint32]bool, tx *sql.Tx, query string) error {
	stmt, err := tx.Prepare(query)
	if err != nil {
		return err
	}
	for id := range set {
		if _, err := stmt.Exec(id); err != nil {
			return err
		}
	}
	stmt.Close()
	return nil
}

// checkOne_helper helps to implement a checker for one identifier (usually a poll).
//
// The given query is executed with pollId as unique parameter. If the query affects exactly one
// row, an event constructed by makeEvent is emitted, and step is increased.
func (self *pollService) checkOne_helper(qUpdate string, pollId uint32) error {
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
		events.Send(self.makeEvent(pollId))
	}

	self.step++
	log.Printf("%s Check for poll %d terminated.", self.serviceName, pollId)
	return nil
}

// updateLastCheck update the lastCheck field.
// The new value is computed from the database.
func (self *pollService) updateLastCheck() {
	row := db.DB.QueryRow(`SELECT CURRENT_TIMESTAMP()`)
	if err := row.Scan(&self.lastCheck); err != nil {
		self.warn.Print(err)
	}
	self.adjust = (self.adjust + time.Since(self.lastCheck)) / 2
}

// nextAlarm_helper helps to implement a method providing events for an alarm.
//
// The query must take one parameter, which is set to lastCheck, and must return 3 values: the id
// (uint32) to inspect when the alarm will fire, the time the alarm must fire, and the current time
// for the database. This query must return zero or one row. If it returns zero row, the returned
// event has Time computed using pollServiceDefaultWaitDuration and Data nil.
func (self *pollService) nextAlarm_helper(qNext string) (ret alarm.Event) {
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
		ret.Time = time.Now().Add(pollServiceDefaultWaitDuration)
	}

	log.Printf("%s Next alarm at %v.", self.serviceName, ret.Time)
	return
}