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
	"net/http"

	"github.com/JBoudou/Itero/db"
	"github.com/JBoudou/Itero/server"
)

type CountInfoEntry struct {
	Alternative PollAlternative
	Count       uint32
}

type CountInfoAnswer struct {
	Result     []CountInfoEntry
}

func CountInfoHandler(ctx context.Context, response server.Response, request *server.Request) {
	pollInfo, err := checkPollAccess(ctx, request)
	must(err)
	if pollInfo.CurrentRound < 1 {
		err = server.NewHttpError(http.StatusInternalServerError, "Protocol error", "No previous round")
		response.SendError(ctx, err)
		return
	}

	var answer CountInfoAnswer
	answer.Result = make([]CountInfoEntry, pollInfo.NbChoices)

	const (
		qCount = `
		  SELECT a.Id, a.Name, IFNULL(b.Count,0)
		    FROM (
		          SELECT Poll, Id, Name
		            FROM Alternatives
		            WHERE Poll = ?
		        ) AS a LEFT JOIN (
		          SELECT Poll, Alternative as Id, COUNT(*) as Count
		            FROM Ballots
		            WHERE Round = ?
		            GROUP BY Poll, Alternative
		        ) AS b ON (a.Poll, a.Id) = (b.Poll, b.Id)
		  ORDER BY b.Count DESC`
	)

	rows, err := db.DB.QueryContext(ctx, qCount, pollInfo.Id, pollInfo.CurrentRound - 1)
	must(err)
	for i := 0; rows.Next(); i++ {
		must(rows.Scan(&answer.Result[i].Alternative.Id,
			&answer.Result[i].Alternative.Name,
			&answer.Result[i].Count))
		answer.Result[i].Alternative.Cost = 1.
	}

	response.SendJSON(ctx, answer)
	return
}
