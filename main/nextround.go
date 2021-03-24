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
)

// The time to wait when there seems to be no forthcoming deadline.
const nextRoundDefaultWaitDuration = time.Hour

// NextRoundEvent is the type of events send when a new round starts.
type NextRoundEvent struct {
	Poll uint32
}

type nextRoundService struct{
	logger LevelLogger
}

var NextRoundService = &nextRoundService{logger: NewPrefixLogger("NextRound")}

func (self *nextRoundService) ProcessOne(id uint32) error {
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

	rows, err := db.DB.Query(qCheck, id, db.PollPublicityInvited)
	defer rows.Close()
	if err != nil {
		return err
	}

	if !rows.Next() {
		return NothingToDoYet
	}
	_, err = db.DB.Exec(qUpdate, id)
	if err != nil {
		return err
	}

	return events.Send(NextRoundEvent{id})
}

func (self *nextRoundService) CheckAll() IdAndDateIterator {
	const (
		qNext = `
		  SELECT Id, RoundDeadline(CurrentRoundStart, MaxRoundDuration, Deadline, CurrentRound, MinNbRounds) AS Next
		    FROM Polls
		   WHERE State = 'Active' AND CurrentRound < MaxNbRounds
		   ORDER BY Next ASC`
	)
	return SqlServiceCheckAll(qNext)
}

func (self *nextRoundService) CheckOne(id uint32) (ret time.Time) {
	const qCheck = `
	    SELECT RoundDeadline(p.CurrentRoundStart, p.MaxRoundDuration, p.Deadline, p.CurrentRound,
			                     p.MinNbRounds),
	           RoundDeadline(p.CurrentRoundStart, p.MaxRoundDuration, p.Deadline, p.CurrentRound,
			                     p.MinNbRounds) <= CURRENT_TIMESTAMP,
	           COALESCE(p.CurrentRound > 0 OR r.Count > 2, FALSE),
	           COALESCE(    (p.CurrentRound > 0 OR p.Publicity = ?)
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

	rows, err := db.DB.Query(qCheck, db.PollPublicityInvited, id)
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

func (self *nextRoundService) CheckInterval() time.Duration {
	return nextRoundDefaultWaitDuration
}

func (self *nextRoundService) Logger() LevelLogger {
	return self.logger
}

func (self *nextRoundService) FilterEvent(evt events.Event) bool {
	switch evt.(type) {
	case VoteEvent, CreatePollEvent, StartPollEvent:
		return true
	}
	return false
}

func (self *nextRoundService) ReceiveEvent(evt events.Event, ctrl ServiceRunnerControl) {
	switch e := evt.(type) {
	case VoteEvent:
		ctrl.Schedule(e.Poll)
	case CreatePollEvent:
		ctrl.Schedule(e.Poll)
	case StartPollEvent:
		ctrl.Schedule(e.Poll)
	}
}
