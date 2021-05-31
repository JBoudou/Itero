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

package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/JBoudou/Itero/mid/db"
	"github.com/JBoudou/Itero/mid/server"
)

type CountInfoEntry struct {
	Alternative PollAlternative
	Count       uint32
}

type CountInfoAnswer struct {
	Result     []CountInfoEntry
}

func getPollRoundFromRequest(request *server.Request, fallback uint8) uint8 {
	remainingLength := len(request.RemainingPath)
	if remainingLength >= 2 {
		tmp, err := strconv.Atoi(request.RemainingPath[0])
		if err == nil {
			return uint8(tmp)
		}
	}
	return fallback
}

// CountInfoEntry sends the plurality result of the previous round.
func CountInfoHandler(ctx context.Context, response server.Response, request *server.Request) {
	pollInfo, err := checkPollAccess(ctx, request)
	must(err)

	// Get the round to return results of.
	round := getPollRoundFromRequest(request, pollInfo.CurrentRound - 1)
	if round >= pollInfo.CurrentRound {
		err = server.NewHttpError(http.StatusBadRequest, "Protocol error", "No result for this round")
		response.SendError(ctx, err)
		return
	}

	var answer CountInfoAnswer
	answer.Result = make([]CountInfoEntry, pollInfo.NbChoices)

	const (
		qReport = `SELECT ReportVote FROM Polls WHERE Id = ?`
		qCountAbstain = `
		  SELECT a.Id, a.Name, a.Cost, IFNULL(b.Count,0)
		    FROM (
		          SELECT *
		            FROM Alternatives
		           WHERE Poll = ?
		         ) AS a LEFT JOIN (
		           SELECT Poll, Alternative as Id, COUNT(*) as Count
		             FROM Ballots
		            WHERE Round = ?
		            GROUP BY Poll, Alternative
		         ) AS b ON (a.Poll, a.Id) = (b.Poll, b.Id)
		   ORDER BY b.Count DESC, a.Id ASC`
		qCountReport = `
		  SELECT a.Id, a.Name, a.Cost, IFNULL(b.Count,0)
		    FROM (
		          SELECT *
		            FROM Alternatives
		           WHERE Poll = ?
		         ) AS a LEFT JOIN (
		           SELECT b.Poll, b.Alternative as Id, COUNT(*) as Count
		             FROM Ballots AS b JOIN (
		                    SELECT User, Poll, MAX(Round) as M
		                      FROM Participants
		                     WHERE Round <= ?
		                     GROUP BY User, Poll
		             ) AS c ON (b.User, b.Poll, b.Round) = (c.User, c.Poll, c.M)
		            GROUP BY b.Poll, b.Alternative
		         ) AS b ON (a.Poll, a.Id) = (b.Poll, b.Id)
		   ORDER BY b.Count DESC, a.Id ASC`
	)

	row := db.DB.QueryRowContext(ctx, qReport, pollInfo.Id)
	var reportVote bool
	must(row.Scan(&reportVote))
	var query string
	if reportVote {
		query = qCountReport
	} else {
		query = qCountAbstain
	}

	rows, err := db.DB.QueryContext(ctx, query, pollInfo.Id, round)
	must(err)
	for i := 0; rows.Next(); i++ {
		must(rows.Scan(&answer.Result[i].Alternative.Id,
			&answer.Result[i].Alternative.Name,
			&answer.Result[i].Alternative.Cost,
			&answer.Result[i].Count))
	}

	response.SendJSON(ctx, answer)
	return
}
