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
	"github.com/JBoudou/Itero/db"
	"github.com/JBoudou/Itero/events"
)

type NextRoundEvent struct {
	Poll uint32
}

type PollClosedEvent struct {
	Poll uint32
}

func roundCheckAllPolls() error {
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

	collectId := func(set map[uint32]bool, query string, queryArgs ...interface{}) error {
		rows, err := tx.Query(query, queryArgs...)
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

	execOnSet := func(set map[uint32]bool, query string) error {
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

	// Next round
	const (
		qSelectNext = `
		  SELECT p.Id
		    FROM Polls AS p LEFT OUTER JOIN Participants AS a ON p.Id = a.Poll
		  WHERE p.Active
		  GROUP BY p.Id,
		        p.CurrentRoundStart, p.MaxRoundDuration, p.CurrentRound, p.Publicity, p.RoundThreshold,
		        p.MaxNbRounds
		 HAVING p.CurrentRound < p.MaxNbRounds
		    AND ( ADDTIME(p.CurrentRoundStart, p.MaxRoundDuration) <= CURRENT_TIMESTAMP()
		          OR ( (p.CurrentRound > 0 OR p.Publicity = ?)
		                AND ( (p.RoundThreshold = 0 AND SUM(a.LastRound = p.CurrentRound) > 0)
		                     OR ( p.RoundThreshold > 0
		                          AND SUM(a.LastRound = p.CurrentRound) / COUNT(a.LastRound) >= p.RoundThreshold ))))
		    FOR UPDATE`
		qNextRound = `UPDATE Polls SET CurrentRound = CurrentRound + 1 WHERE Id = ?`
	)
	nextSet := make(map[uint32]bool)
	if err := collectId(nextSet, qSelectNext, db.PollPublicityInvited); err != nil {
		return err
	}
	if err := execOnSet(nextSet, qNextRound); err != nil {
		return err
	}

	// Close
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
	if err := collectId(closeSet, qSelectClose); err != nil {
		return err
	}
	if err := execOnSet(closeSet, qClosePoll); err != nil {
		return err
	}

	// Commit
	if err := tx.Commit(); err != nil {
		return err
	}
	commited = true

	// Send events
	for id := range closeSet {
		events.Send(PollClosedEvent{id})
		delete(nextSet, id)
	}
	for id := range nextSet {
		events.Send(NextRoundEvent{id})
	}

	return nil
}
