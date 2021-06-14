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
)

// The time to wait when there seems to be no forthcoming deadline.
const nextRoundDefaultWaitDuration = time.Hour

type nextRoundService struct {
	logger     service.LevelLogger
	evtManager events.Manager
}

func NextRoundService(evtManager events.Manager) *nextRoundService {
	return &nextRoundService{
		logger:     service.NewPrefixLogger("NextRound"),
		evtManager: evtManager,
	}
}

func (self *nextRoundService) ProcessOne(id uint32) error {
	const (
		qCheck = `
	    SELECT p.CurrentRound
	      FROM Polls AS p
	      LEFT OUTER JOIN Participants_Round_Count AS r ON (p.Id, p.CurrentRound) = (r.Poll, r.Round)
	      LEFT OUTER JOIN Participants_Poll_Count  AS a ON p.Id = a.Poll
	     WHERE p.Id = ? AND p.State = 'Active' AND p.CurrentRound < p.MaxNbRounds
	       AND (   ( RoundDeadline(p.CurrentRoundStart, p.MaxRoundDuration, p.Deadline,
	                               p.CurrentRound, p.MinNbRounds) <= CURRENT_TIMESTAMP()
	                 AND ( p.CurrentRound > 0 OR r.Count > 2 ))
	            OR (    p.CurrentRound > 0
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

	rows, err := db.DB.Query(qCheck, id)
	defer rows.Close()
	if err != nil {
		return err
	}

	if !rows.Next() {
		return service.NothingToDoYet
	}
	var round uint8
	err = rows.Scan(&round)
	if err != nil {
		return err
	}
	rows.Close()

	_, err = db.DB.Exec(qUpdate, id)
	if err != nil {
		return err
	}

	return self.evtManager.Send(NextRoundEvent{Poll: id, Round: round + 1})
}

func (self *nextRoundService) CheckAll() service.Iterator {
	const (
		qNext = `
		  SELECT Id, RoundDeadline(CurrentRoundStart, MaxRoundDuration, Deadline, CurrentRound, MinNbRounds) AS Next
		    FROM Polls
		   WHERE State = 'Active' AND CurrentRound < MaxNbRounds
		   ORDER BY Next ASC`
	)
	return service.SQLCheckAll(qNext)
}

func (self *nextRoundService) CheckOne(id uint32) (ret time.Time) {
	const qCheck = `
	    SELECT RoundDeadline(p.CurrentRoundStart, p.MaxRoundDuration, p.Deadline, p.CurrentRound,
			                     p.MinNbRounds),
	           RoundDeadline(p.CurrentRoundStart, p.MaxRoundDuration, p.Deadline, p.CurrentRound,
			                     p.MinNbRounds) <= CURRENT_TIMESTAMP,
	           COALESCE(p.CurrentRound > 0 OR r.Count > 2, FALSE),
	           COALESCE(    p.CurrentRound > 0
	                    AND (   (p.RoundThreshold = 0 AND r.Count > 0)
	                         OR ( p.RoundThreshold > 0
	                              AND r.Count / a.Count >= p.RoundThreshold ) )
	                    AND (   (p.CurrentRound + 1 < MinNbRounds)
	                         OR p.Deadline IS NULL
	                         OR (ADDTIME(CURRENT_TIMESTAMP(), p.MaxRoundDuration) < p.Deadline)
	                         OR (p.Deadline < CURRENT_TIMESTAMP()) ), FALSE)
	      FROM Polls AS p
	      LEFT OUTER JOIN Participants_Round_Count AS r ON (p.Id, p.CurrentRound) = (r.Poll, r.Round)
	      LEFT OUTER JOIN Participants_Poll_Count  AS a ON p.Id = a.Poll
	     WHERE p.Id = ? AND p.State = 'Active' AND p.CurrentRound < p.MaxNbRounds`

	rows, err := db.DB.Query(qCheck, id)
	defer rows.Close()
	if err != nil {
		self.Logger().Errorf("CheckOne query error: %v.", err)
		return
	}
	if !rows.Next() {
		return
	}

	var dueTime, easyRound, enoughBallots bool
	if err = rows.Scan(&ret, &dueTime, &easyRound, &enoughBallots); err != nil {
		self.Logger().Errorf("CheckOne scan error: %v.", err)
		return
	}

	if (dueTime && easyRound) || enoughBallots {
		return time.Now()
	}
	if dueTime {
		return time.Time{}
	}
	return
}

func (self *nextRoundService) Interval() time.Duration {
	return nextRoundDefaultWaitDuration
}

func (self *nextRoundService) Logger() service.LevelLogger {
	return self.logger
}

func (self *nextRoundService) FilterEvent(evt events.Event) bool {
	switch evt.(type) {
	case VoteEvent, CreatePollEvent, StartPollEvent:
		return true
	}
	return false
}

func (self *nextRoundService) ReceiveEvent(evt events.Event, ctrl service.RunnerControler) {
	switch e := evt.(type) {
	case VoteEvent:
		ctrl.Schedule(e.Poll)
	case CreatePollEvent:
		ctrl.Schedule(e.Poll)
	case StartPollEvent:
		ctrl.Schedule(e.Poll)
	}
}
