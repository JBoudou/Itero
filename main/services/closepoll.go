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

package services

import (
	"time"

	"github.com/JBoudou/Itero/mid/db"
	"github.com/JBoudou/Itero/mid/service"
	"github.com/JBoudou/Itero/pkg/events"
	"github.com/JBoudou/Itero/pkg/slog"
)

type closePollService struct {
	logger     slog.Leveled
	evtManager events.Manager
}

// ClosePollService is the factory for the service that terminates polls when their last round is
// over.
func ClosePollService(evtManager events.Manager, log slog.StackedLeveled) *closePollService {
	return &closePollService{
		logger:     log.With("ClosePoll"),
		evtManager: evtManager,
	}
}

func (self *closePollService) ProcessOne(id uint32) error {
	const qUpdate = `
	  UPDATE Polls SET State = 'Terminated'
	   WHERE Id = ? AND State = 'Active'
	     AND ( CurrentRound >= MaxNbRounds
	           OR (CurrentRound >= MinNbRounds AND Deadline <= CURRENT_TIMESTAMP) )`

	if err := service.SQLProcessOne(qUpdate, id); err != nil {
		return err
	}
	return self.evtManager.Send(ClosePollEvent{id})
}

func (self *closePollService) CheckAll() service.Iterator {
	const qSelectClose = `
		  SELECT Id, COALESCE(LEAST(Deadline, CURRENT_TIMESTAMP), CURRENT_TIMESTAMP)
		    FROM Polls
		  WHERE State = 'Active'
		    AND ( CurrentRound >= MaxNbRounds
		          OR (CurrentRound >= MinNbRounds AND Deadline <= CURRENT_TIMESTAMP) )`
	return service.SQLCheckAll(qSelectClose)
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

func (self *closePollService) Interval() time.Duration {
	return 12 * time.Hour
}

func (self *closePollService) Logger() slog.Leveled {
	return self.logger
}

func (self *closePollService) FilterEvent(evt events.Event) bool {
	switch evt.(type) {
	case NextRoundEvent:
		return true
	}
	return false
}

func (self *closePollService) ReceiveEvent(evt events.Event, ctrl service.RunnerControler) {
	switch e := evt.(type) {
	case NextRoundEvent:
		ctrl.Schedule(e.Poll)
	}
}
