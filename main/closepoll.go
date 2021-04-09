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
	"time"

	"github.com/JBoudou/Itero/db"
	"github.com/JBoudou/Itero/events"
	"github.com/JBoudou/Itero/service"
)

// ClosePollEvent is type of events send when a poll is closed.
type ClosePollEvent struct {
	Poll uint32
}

type closePollService struct {
	logger service.LevelLogger
}

var ClosePollService = &closePollService{logger: service.NewPrefixLogger("ClosePoll")}

func (self *closePollService) ProcessOne(id uint32) error {
	const qUpdate = `
	  UPDATE Polls SET State = 'Terminated'
	   WHERE Id = ? AND State = 'Active'
	     AND ( CurrentRound >= MaxNbRounds
	           OR (CurrentRound >= MinNbRounds AND Deadline <= CURRENT_TIMESTAMP) )`

	return service.SqlServiceProcessOne(qUpdate, id, ClosePollEvent{id})
}

func (self *closePollService) CheckAll() service.IdAndDateIterator {
	const qSelectClose = `
		  SELECT Id, COALESCE(LEAST(Deadline, CURRENT_TIMESTAMP), CURRENT_TIMESTAMP)
		    FROM Polls
		  WHERE State = 'Active'
		    AND ( CurrentRound >= MaxNbRounds
		          OR (CurrentRound >= MinNbRounds AND Deadline <= CURRENT_TIMESTAMP) )`
	return service.SqlServiceCheckAll(qSelectClose)
}

func (self *closePollService) CheckOne(id uint32) time.Time {
	const qCheckOne = `
	  SELECT 1 FROM Polls
	   WHERE Id = ? AND State = 'Active'
	     AND ( CurrentRound >= MaxNbRounds
	           OR (CurrentRound >= MinNbRounds AND Deadline <= CURRENT_TIMESTAMP) )`

	rows, err := db.DB.Query(qCheckOne, id)
	defer rows.Close()
	if err != nil {
		self.Logger().Errorf("CheckOne query error: %v", err)
		return time.Time{}
	}
	if !rows.Next() {
		return time.Time{}
	}
	return time.Now()
}

func (self *closePollService) CheckInterval() time.Duration {
	return 12 * time.Hour
}

func (self *closePollService) Logger() service.LevelLogger {
	return self.logger
}

func (self *closePollService) FilterEvent(evt events.Event) bool {
	switch evt.(type) {
	case NextRoundEvent:
		return true
	}
	return false
}

func (self *closePollService) ReceiveEvent(evt events.Event, ctrl service.ServiceRunnerControl) {
	switch e := evt.(type) {
	case NextRoundEvent:
		ctrl.Schedule(e.Poll)
	}
}
