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
	"net/http"
	"strconv"
	"testing"

	"github.com/JBoudou/Itero/mid/db"
	dbt "github.com/JBoudou/Itero/mid/db/dbtest"
	"github.com/JBoudou/Itero/mid/salted"
	srvt "github.com/JBoudou/Itero/mid/server/servertest"
)

func TestCountInfoHandler(t *testing.T) {
	precheck(t)

	const (
		qReport    = `UPDATE Polls SET ReportVote = TRUE WHERE Id = ?`
		qBlankVote = `INSERT INTO Participants (Poll, Round, User) VALUE (?,?,?)`
	)

	var env dbt.Env
	defer env.Close()
	var users [3]uint32
	for i := range users {
		users[i] = env.CreateUserWith(strconv.FormatInt(int64(i), 10))
	}

	alt := []string{"Ham", "Stram", "Gram"}
	altAns := [3]PollAlternative{
		{Id: 0, Name: "Ham", Cost: 1},
		{Id: 1, Name: "Stram", Cost: 1},
		{Id: 2, Name: "Gram", Cost: 1},
	}

	pollId := env.CreatePollWith("Test", users[0], db.ElectorateAll, alt)
	env.Vote(pollId, 0, users[0], 2)
	env.Vote(pollId, 0, users[1], 2)
	env.Vote(pollId, 0, users[2], 0)
	env.Must(t)

	request := *makePollRequest(t, pollId, &users[0])

	previousRoundRequest := func(t *testing.T, round uint8) srvt.Request {
		pollSegment := salted.Segment{Salt: 42, Id: pollId}
		encoded, err := pollSegment.Encode()
		mustt(t, err)
		target := "/a/test/" + strconv.FormatUint(uint64(round), 10) + "/" + encoded
		return srvt.Request{Target: &target, UserId: &users[0]}
	}

	// In each pair of the parameter,
	// the first value is the alternative index,
	// the second value is the number of votes.
	makeChecker := func(result [][2]uint32) srvt.Checker {
		entries := make([]CountInfoEntry, len(result))
		for i, val := range result {
			entries[i].Alternative = altAns[val[0]]
			entries[i].Count = val[1]
		}
		return srvt.CheckJSON{Body: CountInfoAnswer{Result: entries}}
	}

	tests := []srvt.Test{

		// Sequential tests first //
		// TODO make them all independent

		&srvt.T{
			Name:    "Round Zero",
			Request: request,
			Checker: srvt.CheckStatus{http.StatusBadRequest},
		},
		&srvt.T{
			Name: "All voted",
			Update: func(t *testing.T) {
				env.NextRound(pollId)
				env.Must(t)
			},
			Request: request,
			Checker: makeChecker([][2]uint32{{2, 2}, {0, 1}, {1, 0}}),
		},
		&srvt.T{
			Name: "One voted",
			Update: func(t *testing.T) {
				env.Vote(pollId, 1, users[0], 1)
				env.NextRound(pollId)
				env.Must(t)
			},
			Request: request,
			Checker: makeChecker([][2]uint32{{1, 1}, {0, 0}, {2, 0}}),
		},
		&srvt.T{
			Name: "Carry forward",
			Update: func(t *testing.T) {
				env.Vote(pollId, 2, users[1], 0)
				env.NextRound(pollId)
				// Current round is 3
				env.QuietExec(qReport, pollId)
				env.Must(t)
			},
			Request: request,
			// 0 voted 1 on round 1
			// 1 voted 0 on round 2
			// 2 voted 0 on round 0
			Checker: makeChecker([][2]uint32{{0, 2}, {1, 1}, {2, 0}}),
		},
		&srvt.T{
			Name: "Vote after carry forward result",
			Update: func(t *testing.T) {
				env.Vote(pollId, 3, users[2], 2)
				env.Must(t)
			},
			Request: request,
			Checker: makeChecker([][2]uint32{{0, 2}, {1, 1}, {2, 0}}),
		},
		&srvt.T{
			Name:    "Round 3",
			Request: previousRoundRequest(t, 3),
			Checker: srvt.CheckStatus{http.StatusBadRequest},
		},
		&srvt.T{
			Name:    "Round 2",
			Request: previousRoundRequest(t, 2),
			Checker: makeChecker([][2]uint32{{0, 2}, {1, 1}, {2, 0}}),
		},
		&srvt.T{
			Name:    "Round 1",
			Request: previousRoundRequest(t, 1),
			Checker: makeChecker([][2]uint32{{0, 1}, {1, 1}, {2, 1}}),
		},
		&srvt.T{
			Name:    "Round 0",
			Request: previousRoundRequest(t, 0),
			Checker: makeChecker([][2]uint32{{2, 2}, {0, 1}, {1, 0}}),
		},
		&srvt.T{
			Name: "Abstain",
			Update: func(t *testing.T) {
				_, err := db.DB.Exec(qBlankVote, pollId, 3, users[0])
				mustt(t, err)
				env.NextRound(pollId)
				env.Must(t)
			},
			Request: request,
			Checker: makeChecker([][2]uint32{{0, 1}, {2, 1}, {1, 0}}),
		},

		// Independent tests //

		&pollTest{
			Name:         "No user public",
			Electorate: db.ElectorateAll,
			Alternatives: alt,
			UserType:     pollTestUserTypeNone,
			Vote:         []pollTestVote{{2, 0, 0}, {3, 0, 0}, {4, 0, 2}},
			Round:        1,
			Checker:      makeChecker([][2]uint32{{0, 2}, {2, 1}, {1, 0}}),
		},
		&pollTest{
			Name:         "No user hidden",
			Electorate: db.ElectorateAll,
			Hidden: true,
			Alternatives: alt,
			UserType:     pollTestUserTypeNone,
			Vote:         []pollTestVote{{2, 0, 0}, {3, 0, 0}, {4, 0, 2}},
			Round:        1,
			Checker:      makeChecker([][2]uint32{{0, 2}, {2, 1}, {1, 0}}),
		},
		&pollTest{
			Name:         "Unlogged public",
			Electorate: db.ElectorateAll,
			Alternatives: alt,
			UserType:     pollTestUserTypeUnlogged,
			Vote:         []pollTestVote{{2, 0, 0}, {3, 0, 0}, {4, 0, 2}},
			Round:        1,
			Checker:      makeChecker([][2]uint32{{0, 2}, {2, 1}, {1, 0}}),
		},
		&pollTest{
			Name:         "Unlogged hidden",
			Electorate: db.ElectorateAll,
			Hidden: true,
			Alternatives: alt,
			UserType:     pollTestUserTypeUnlogged,
			Vote:         []pollTestVote{{2, 0, 0}, {3, 0, 0}, {4, 0, 2}},
			Round:        1,
			Checker:      makeChecker([][2]uint32{{0, 2}, {2, 1}, {1, 0}}),
		},

		&pollTest{
			Name:         "No user public registered",
			Electorate: db.ElectorateLogged,
			Alternatives: alt,
			UserType:     pollTestUserTypeNone,
			Vote:         []pollTestVote{{2, 0, 0}, {3, 0, 0}, {4, 0, 2}},
			Round:        1,
			Checker:      srvt.CheckStatus{http.StatusNotFound},
		},
		&pollTest{
			Name:         "No user hidden registered",
			Electorate: db.ElectorateLogged,
			Hidden: true,
			Alternatives: alt,
			UserType:     pollTestUserTypeNone,
			Vote:         []pollTestVote{{2, 0, 0}, {3, 0, 0}, {4, 0, 2}},
			Round:        1,
			Checker:      srvt.CheckStatus{http.StatusNotFound},
		},
		&pollTest{
			Name:         "Unlogged public registered",
			Electorate: db.ElectorateLogged,
			Alternatives: alt,
			UserType:     pollTestUserTypeUnlogged,
			Vote:         []pollTestVote{{2, 0, 0}, {3, 0, 0}, {4, 0, 2}},
			Round:        1,
			Checker:      srvt.CheckStatus{http.StatusNotFound},
		},
		&pollTest{
			Name:         "Unlogged hidden registered",
			Electorate: db.ElectorateLogged,
			Hidden: true,
			Alternatives: alt,
			UserType:     pollTestUserTypeUnlogged,
			Vote:         []pollTestVote{{2, 0, 0}, {3, 0, 0}, {4, 0, 2}},
			Round:        1,
			Checker:      srvt.CheckStatus{http.StatusNotFound},
		},
	}
	srvt.RunFunc(t, tests, CountInfoHandler)
}
