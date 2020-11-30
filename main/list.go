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

// TODO Move PollSegment somewhere else.

// end PollSegment

type listAnswerEntry struct {
	Segment      string `json:"s"`
	Title        string `json:"t"`
	CurrentRound uint8  `json:"c"`
	MaxRound     uint8  `json:"m"`
	Deadline     string `json:"d"` // TODO Use (a variant of) time.Time
	Action       string `json:"a"` // TODO Use an "enum" ?
}

func ListHandler(ctx context.Context, response server.Response, request *server.Request) {
	reply := make([]listAnswerEntry, 0, 16)

	if request.User == nil {
		// TODO change that
		response.SendError(ctx, server.NewHttpError(http.StatusNotImplemented, "Unimplemented", ""))
		return
	}

	const query = `SELECT p.Id, p.Salt, p.Title, p.CurrentRound, p.MaxNbRounds,
		        addtime(p.CurrentRoundStart, p.MaxRoundDuration) AS Deadline,
		        CASE WHEN a.User IS NULL THEN 'Part'
		             WHEN a.LastRound >= p.CurrentRound THEN 'Modi'
		             ELSE 'Vote' END AS Action
		   FROM Polls AS p LEFT OUTER JOIN Participants AS a ON p.Id = a.Poll
		  WHERE p.Active
			  AND ((a.User IS NULL AND p.CurrentRound = 0 AND p.Publicity <= ?) OR a.User = ?)`
	rows, err := db.DB.QueryContext(ctx, query, db.PollPublicityPublicRegistered, request.User.Id)
	if err != nil {
		response.SendError(ctx, err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var listAnswerEntry listAnswerEntry
		var segment PollSegment
		var deadline sql.NullString

		err = rows.Scan(&segment.Id, &segment.Salt, &listAnswerEntry.Title,
			&listAnswerEntry.CurrentRound, &listAnswerEntry.MaxRound, &deadline,
			&listAnswerEntry.Action)
		if err != nil {
			response.SendError(ctx, err)
			return
		}

		listAnswerEntry.Segment, err = segment.Encode()
		if err != nil {
			response.SendError(ctx, err)
			return
		}
		if deadline.Valid {
			listAnswerEntry.Deadline = deadline.String
		} else {
			listAnswerEntry.Deadline = "⋅";
		}

		reply = append(reply, listAnswerEntry)
	}

	response.SendJSON(ctx, reply)
	return
}
