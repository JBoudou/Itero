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
	"time"

	"github.com/JBoudou/Itero/db"
	"github.com/JBoudou/Itero/events"
)

// The time to wait when there seems to be no waiting poll.
const	startPollDefaultWaitDuration = time.Hour

// StartPollEvent is send when a poll is started.
type StartPollEvent struct {
	Poll uint32
}

type startPollService struct {
	logger LevelLogger
}

var StartPollService = &startPollService{logger: NewPrefixLogger("StartPoll")}

func (self *startPollService) ProcessOne(id uint32) error {
	const qUpdate = `
	  UPDATE Polls SET State = 'Active'
	   WHERE Id = ? AND State = 'Waiting'
	     AND Start <= CURRENT_TIMESTAMP`
	
	result, err := db.DB.Exec(qUpdate, id)
	if err != nil {
		return err
	}
	
	var affected int64
	affected, err = result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return NothingToDoYet
	}

	return events.Send(StartPollEvent{id})
}

func (self *startPollService) CheckAll() IdAndDateIterator {
	const	qSelectStart = `
		  SELECT Id, Start
		    FROM Polls
		   WHERE State  = 'Waiting'
			 ORDER BY Start ASC`

	rows, err := db.DB.Query(qSelectStart)
	if err != nil {
		return ErrorIdDateIterator{err}
	} else {
		return NewRowsIdDateIterator(rows)
	}
}

func (self *startPollService) CheckOne(id uint32) time.Time {
	const qCheckOne = `SELECT Start FROM Polls WHERE Id = ? AND State = 'Waiting'`

	rows, err := db.DB.Query(qCheckOne, id)
	defer rows.Close()
	if err != nil {
		self.Logger().Errorf("CheckOne query error: %v", err)
		return time.Time{}
	}
	if !rows.Next() {
		return time.Time{}
	}
	
	var ret time.Time
	err = rows.Scan(&ret)
	if err != nil {
		self.Logger().Errorf("CheckOne scan error: %v", err)
		return time.Time{}
	}
	return ret
}

func (self *startPollService) CheckInterval() time.Duration {
	return startPollDefaultWaitDuration
}

func (self *startPollService) Logger() LevelLogger {
	return self.logger
}


func (self *startPollService) FilterEvent(evt events.Event) bool {
	switch evt.(type) {
	case CreatePollEvent:
		return true
	}
	return false
}

func (self *startPollService) ReceiveEvent(evt events.Event, ctrl ServiceRunnerControl) {
	switch e := evt.(type) {
	case CreatePollEvent:
		ctrl.Schedule(e.Poll)
	}
}
