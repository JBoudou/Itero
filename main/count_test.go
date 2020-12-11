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
	"testing"

	"github.com/JBoudou/Itero/db"
	dbt "github.com/JBoudou/Itero/db/dbtest"
	srvt "github.com/JBoudou/Itero/server/servertest"
)

func TestCountInfoHandler(t *testing.T) {
	precheck(t)

	var env dbt.Env
	var users [3]uint32
	for i, _ := range users {
		users[i] = env.CreateUserWith(string(i))
	}

	pollSegment := PollSegment{Salt: 42}
	pollSegment.Id = env.CreatePollWith("Test", users[0], db.PollPublicityPublic, []string{
		"Ham", "Stram", "Gram"})
	env.Vote(pollSegment.Id, 0, users[0], 2)
	env.Vote(pollSegment.Id, 0, users[1], 2)
	env.Vote(pollSegment.Id, 0, users[2], 0)
	env.Must(t)

	request := makePollRequest(t, pollSegment, &users[0])

	tests := []srvt.Test{
		{
			Name:    "Round Zero",
			Request: request,
			Checker: srvt.CheckStatus{http.StatusInternalServerError},
		},
		{
			Name: "Success",
			Update: func(t *testing.T) {
				env.NextRound(pollSegment.Id)
				env.Must(t)
			},
			Request: request,
			Checker: srvt.CheckJSON{Body: CountInfoAnswer{
				Result: []CountInfoEntry{
					{
						Alternative: PollAlternative{Id: 2, Name: "Gram", Cost: 1.},
						Count:       2,
					},
					{
						Alternative: PollAlternative{Id: 0, Name: "Ham", Cost: 1.},
						Count:       1,
					},
					{
						Alternative: PollAlternative{Id: 1, Name: "Stram", Cost: 1.},
						Count:       0,
					},
				},
			}},
		},
	}
	srvt.RunFunc(t, tests, CountInfoHandler)
}
