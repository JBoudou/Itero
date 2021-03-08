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
	"context"
	"database/sql"
	"net/http"

	"github.com/JBoudou/Itero/db"
	"github.com/JBoudou/Itero/server"
)

type NuDate sql.NullTime

func (self NuDate) MarshalJSON() ([]byte, error) {
	if !self.Valid {
		return []byte(`"⋅"`), nil
	} else {
		return self.Time.MarshalJSON()
	}
}

func (self *NuDate) UnmarshalJSON(raw []byte) (err error) {
	if string(raw) == `"⋅"` {
		self.Valid = false
		return
	}
	err = self.Time.UnmarshalJSON(raw)
	self.Valid = err == nil
	return
}

const (
	PollActionVote = iota
	PollActionModif
	PollActionPart
	PollActionTerm
	PollActionWait
)

type ListAnswer struct {
	Public []listAnswerEntry
	Own []listAnswerEntry
}

type listAnswerEntry struct {
	Segment      string
	Title        string
	CurrentRound uint8 
	MaxRound     uint8 
	Deadline     NuDate
	Action       uint8
}

func ListHandler(ctx context.Context, response server.Response, request *server.Request) {
	if request.User == nil {
		// TODO change that
		response.SendError(ctx, server.NewHttpError(http.StatusNotImplemented, "Unimplemented", ""))
		return
	}

	const (
		qPublic = `
	    SELECT p.Id, p.Salt, p.Title, p.CurrentRound, p.MaxNbRounds,
	           RoundDeadline(p.CurrentRoundStart, p.MaxRoundDuration, p.Deadline) AS Deadline,
	           CASE WHEN p.State = 'Terminated' THEN 3
	                WHEN a.User IS NULL THEN 2
	                WHEN a.LastRound >= p.CurrentRound THEN 1
	                ELSE 0 END AS Action
	      FROM Polls AS p LEFT OUTER JOIN (
	               SELECT Poll, User, LastRound
	                FROM Participants
	               WHERE User = ?
	           ) AS a ON p.Id = a.Poll
	     WHERE ( (p.State != 'Waiting' AND p.CurrentRound = 0 AND p.Publicity <= ?)
		            OR a.User IS NOT NULL )
		     AND p.Admin != ?
	     ORDER BY Action ASC, Deadline ASC`
		qOwn = `
	    SELECT p.Id, p.Salt, p.Title, p.CurrentRound, p.MaxNbRounds,
	           CASE WHEN p.State = 'Waiting' THEN p.Start
						      ELSE RoundDeadline(p.CurrentRoundStart, p.MaxRoundDuration, p.Deadline) END AS Deadline,
	           CASE WHEN p.State = 'Waiting' THEN 4
						      WHEN p.State = 'Terminated' THEN 3
	                WHEN a.User IS NULL THEN 2
	                WHEN a.LastRound >= p.CurrentRound THEN 1
	                ELSE 0 END AS Action
	      FROM Polls AS p LEFT OUTER JOIN (
	               SELECT Poll, User, LastRound
	                FROM Participants
	               WHERE User = ?
	           ) AS a ON p.Id = a.Poll
		   WHERE p.Admin = ?
	     ORDER BY Action ASC, Deadline ASC`
	)

	var publicList []listAnswerEntry
	rows, err := db.DB.QueryContext(ctx, qPublic,
		request.User.Id, db.PollPublicityPublicRegistered, request.User.Id)
	must(err)
	publicList, err = makeListEntriesList(rows)
	must(err)

	var ownList []listAnswerEntry
	rows, err = db.DB.QueryContext(ctx, qOwn, request.User.Id, request.User.Id)
	must(err)
	ownList, err = makeListEntriesList(rows)
	must(err)

	response.SendJSON(ctx, ListAnswer{Public: publicList, Own: ownList})
}

func makeListEntriesList(rows *sql.Rows) (list []listAnswerEntry, err error) {
	list = make([]listAnswerEntry, 0, 4)
	defer rows.Close()

	for rows.Next() {
		var listAnswerEntry listAnswerEntry
		var segment PollSegment
		var deadline sql.NullTime

		err = rows.Scan(&segment.Id, &segment.Salt, &listAnswerEntry.Title,
			&listAnswerEntry.CurrentRound, &listAnswerEntry.MaxRound, &deadline,
			&listAnswerEntry.Action)
		if err != nil {
			return
		}

		listAnswerEntry.Deadline = NuDate(deadline)
		listAnswerEntry.Segment, err = segment.Encode()
		if err != nil {
			return
		}

		list = append(list, listAnswerEntry)
	}

	return
}
