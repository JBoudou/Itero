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
	"net/http"
	"strconv"
	"testing"

	"github.com/JBoudou/Itero/db"
	dbt "github.com/JBoudou/Itero/db/dbtest"
	srvt "github.com/JBoudou/Itero/server/servertest"
)

func TestCountInfoHandler(t *testing.T) {
	precheck(t)

	const (
		qReport = `UPDATE Polls SET ReportVote = TRUE WHERE Id = ?`
		qSetRound = `UPDATE Participants SET LastRound = ? WHERE User = ?`
	)

	var env dbt.Env
	defer env.Close()
	var users [3]uint32
	for i := range users {
		users[i] = env.CreateUserWith(strconv.FormatInt(int64(i), 10))
	}

	pollSegment := PollSegment{Salt: 42}
	pollSegment.Id = env.CreatePollWith("Test", users[0], db.PollPublicityPublic, []string{
		"Ham", "Stram", "Gram"})
	env.Vote(pollSegment.Id, 0, users[0], 2)
	env.Vote(pollSegment.Id, 0, users[1], 2)
	env.Vote(pollSegment.Id, 0, users[2], 0)
	env.Must(t)

	request := makePollRequest(t, pollSegment, &users[0])

	alt := [3]PollAlternative{
		{Id: 0, Name: "Ham", Cost: 1},
		{Id: 1, Name: "Stram", Cost: 1},
		{Id: 2, Name: "Gram", Cost: 1},
	}

	// In each pair of the parameter,
	// the first value is the alternative index,
	// the second value is the number of votes.
	makeChecker := func(result [][2]uint32) srvt.Checker {
		entries := make([]CountInfoEntry, len(result))
		for i, val := range result {
			entries[i].Alternative = alt[val[0]]
			entries[i].Count = val[1]
		}
		return srvt.CheckJSON{Body: CountInfoAnswer{Result: entries}}
	}

	tests := []srvt.Test{
		// WARNING: Tests are sequential!
		{
			Name:    "Round Zero",
			Request: request,
			Checker: srvt.CheckStatus{http.StatusInternalServerError},
		},
		{
			Name: "All voted",
			Update: func(t *testing.T) {
				env.NextRound(pollSegment.Id)
				env.Must(t)
			},
			Request: request,
			Checker: makeChecker([][2]uint32{{2,2}, {0,1}, {1,0}}),
		},
		{
			Name: "One voted",
			Update: func(t *testing.T) {
				env.Vote(pollSegment.Id, 1, users[0], 1)
				env.NextRound(pollSegment.Id)
				env.Must(t)
			},
			Request: request,
			Checker: makeChecker([][2]uint32{{1,1},{0,0},{2,0}}),
		},
		{
			Name: "Carry forward",
			Update: func(t *testing.T) {
				env.Vote(pollSegment.Id, 2, users[1], 0)
				env.NextRound(pollSegment.Id)
				// Current round is 3
				env.QuietExec(qReport, pollSegment.Id)
				env.QuietExec(qSetRound, 1, users[0])
				env.QuietExec(qSetRound, 2, users[1])
				env.QuietExec(qSetRound, 0, users[2])
				env.Must(t)
			},
			Request: request,
			// 0 voted 1 on round 1
			// 1 voted 0 on round 2
			// 2 voted 0 on round 0
			Checker: makeChecker([][2]uint32{{0,2},{1,1},{2,0}}),
		},
		{
			Name: "Vote after carry forward result",
			Update: func(t *testing.T) {
				env.Vote(pollSegment.Id, 3, users[2], 2)
				env.QuietExec(qSetRound, 3, users[2])
				env.Must(t)
			},
			Request: request,
			Checker: makeChecker([][2]uint32{{0,2},{1,1},{2,0}}),
		},
	}
	srvt.RunFunc(t, tests, CountInfoHandler)
}
