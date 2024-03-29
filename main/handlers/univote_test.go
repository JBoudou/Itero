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
	"database/sql"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/JBoudou/Itero/main/services"
	"github.com/JBoudou/Itero/mid/db"
	dbt "github.com/JBoudou/Itero/mid/db/dbtest"
	"github.com/JBoudou/Itero/mid/server"
	srvt "github.com/JBoudou/Itero/mid/server/servertest"
	"github.com/JBoudou/Itero/pkg/events"
)

type voteChecker struct {
	poll  uint32
	user  uint32
	round uint8
}

func (self *voteChecker) Check(t *testing.T, response *http.Response, request *server.Request) {
	srvt.CheckStatus{http.StatusOK}.Check(t, response, request)

	var query UninominalVoteQuery
	mustt(t, request.UnmarshalJSONBody(&query))

	const qCheck = `
		SELECT p.Round, b.Alternative
		  FROM (
						 SELECT Poll, User, MAX(Round) AS Round
						   FROM Participants
							GROUP BY Poll, User
		       ) AS p LEFT OUTER JOIN Ballots AS b
			  ON (p.Poll, p.User, p.Round) = (b.Poll, b.User, b.Round)
		 WHERE p.Poll = ? AND p.User = ?`

	rows, err := db.DB.Query(qCheck, self.poll, self.user)
	mustt(t, err)
	defer rows.Close()
	if !rows.Next() {
		t.Errorf("User %d does not participate in poll %d.", self.user, self.poll)
		return
	}
	var gotRound, gotAlternative sql.NullInt32
	mustt(t, rows.Scan(&gotRound, &gotAlternative))
	if rows.Next() {
		t.Errorf("More than one alternative for user %d on poll %d.", self.user, self.poll)
	}
	if !gotRound.Valid || gotRound.Int32 != int32(self.round) {
		t.Errorf("Wrong round. Got %v. Expect %d.", gotRound, self.round)
	}

	if query.Blank {
		if gotAlternative.Valid {
			t.Errorf("Non-blank vote %d.", gotAlternative.Int32)
		}
	} else {
		if !gotAlternative.Valid {
			t.Errorf("Got blank vote. Expect %d.", query.Alternative)
		} else if gotAlternative.Int32 != int32(query.Alternative) {
			t.Errorf("Wrong alternative. Got %d. Expect %d.", gotAlternative.Int32, query.Alternative)
		}
	}
}

func voteCheckerFactory(param PollTestCheckerFactoryParam) srvt.Checker {
	return &voteChecker{poll: param.PollId, user: param.UserId, round: param.Round}
}

func checkVoteEvent(param PollTestCheckerFactoryParam, evt events.Event) bool {
	converted, ok := evt.(services.VoteEvent)
	return ok && converted.Poll == param.PollId
}

func TestUninominalVoteHandler(t *testing.T) {
	precheck(t)

	var env dbt.Env
	defer env.Close()
	userId := env.CreateUser()
	pollId := env.CreatePoll("Test", userId, db.ElectorateLogged)
	env.Must(t)

	fillRequest := func(vote UninominalVoteQuery, req srvt.Request) srvt.Request {
		b, err := json.Marshal(vote)
		mustt(t, err)
		req.Body = string(b)
		req.Method = "POST"
		return req
	}

	makeRequest := func(userId *uint32, pollId uint32, vote UninominalVoteQuery) srvt.Request {
		return fillRequest(vote, *makePollRequest(t, pollId, userId))
	}

	s := func(s string) *string { return &s }

	tests := []srvt.Test{

		// Sequential tests first //
		// TODO make them all independent

		&srvt.T{
			Name:    "First vote",
			Request: makeRequest(&userId, pollId, UninominalVoteQuery{Alternative: 1}),
			Checker: &voteChecker{poll: pollId, user: userId, round: 0},
		},
		&srvt.T{
			Name:    "Change vote",
			Request: makeRequest(&userId, pollId, UninominalVoteQuery{Alternative: 0}),
			Checker: &voteChecker{poll: pollId, user: userId, round: 0},
		},
		&srvt.T{
			Name:    "Change vote again",
			Request: makeRequest(&userId, pollId, UninominalVoteQuery{Alternative: 1}),
			Checker: &voteChecker{poll: pollId, user: userId, round: 0},
		},
		&srvt.T{
			Name: "First blank",
			Update: func(t *testing.T) {
				env.NextRound(pollId)
				env.Must(t)
			},
			Request: makeRequest(&userId, pollId, UninominalVoteQuery{Blank: true, Round: 1}),
			Checker: &voteChecker{poll: pollId, user: userId, round: 1},
		},
		&srvt.T{
			Name:    "Change to non-blank",
			Request: makeRequest(&userId, pollId, UninominalVoteQuery{Alternative: 1, Round: 1}),
			Checker: &voteChecker{poll: pollId, user: userId, round: 1},
		},
		&srvt.T{
			Name:    "Change to blank",
			Request: makeRequest(&userId, pollId, UninominalVoteQuery{Blank: true, Round: 1}),
			Checker: &voteChecker{poll: pollId, user: userId, round: 1},
		},
		&srvt.T{
			Name:    "Previous round",
			Request: makeRequest(&userId, pollId, UninominalVoteQuery{Blank: true, Round: 0}),
			Checker: srvt.CheckError{Code: http.StatusLocked, Body: "Next round"},
		},
		&srvt.T{
			Name:    "Next round",
			Request: makeRequest(&userId, pollId, UninominalVoteQuery{Blank: true, Round: 2}),
			Checker: srvt.CheckStatus{Code: http.StatusBadRequest},
		},
		&srvt.T{
			Name: "Inactive",
			Update: func(t *testing.T) {
				const qInactivate = `UPDATE Polls SET State = 'Terminated' WHERE Id = ?`
				_, err := db.DB.Exec(qInactivate, pollId)
				mustt(t, err)
			},
			Request: makeRequest(&userId, pollId, UninominalVoteQuery{}),
			Checker: srvt.CheckStatus{http.StatusLocked},
		},

		// Independent tests //

		&pollTest{
			Name:           "No user public",
			Electorate:     db.ElectorateAll,
			UserType:       pollTestUserTypeNone,
			Request:        fillRequest(UninominalVoteQuery{Alternative: 0}, srvt.Request{RemoteAddr: s("1.2.3.4:5")}),
			EventPredicate: checkVoteEvent,
			EventCount:     1,
			Checker:        voteCheckerFactory,
		},
		&pollTest{
			Name:           "No user hidden",
			Electorate:     db.ElectorateAll,
			Hidden:         true,
			UserType:       pollTestUserTypeNone,
			Request:        fillRequest(UninominalVoteQuery{Alternative: 0}, srvt.Request{RemoteAddr: s("1.2.3.4:5")}),
			EventPredicate: checkVoteEvent,
			EventCount:     1,
			Checker:        voteCheckerFactory,
		},
		&pollTest{
			Name:           "Unlogged public",
			Electorate:     db.ElectorateAll,
			UserType:       pollTestUserTypeUnlogged,
			Request:        fillRequest(UninominalVoteQuery{Alternative: 0}, srvt.Request{}),
			EventPredicate: checkVoteEvent,
			EventCount:     1,
			Checker:        voteCheckerFactory,
		},
		&pollTest{
			Name:           "Unlogged hidden",
			Electorate:     db.ElectorateAll,
			Hidden:         true,
			UserType:       pollTestUserTypeUnlogged,
			Request:        fillRequest(UninominalVoteQuery{Alternative: 0}, srvt.Request{}),
			EventPredicate: checkVoteEvent,
			EventCount:     1,
			Checker:        voteCheckerFactory,
		},

		&pollTest{
			Name:           "No user public change",
			Electorate:     db.ElectorateAll,
			Vote:           []pollTestVote{{User: 1, Alt: 1}},
			UserType:       pollTestUserTypeNone,
			Request:        fillRequest(UninominalVoteQuery{Alternative: 0}, srvt.Request{RemoteAddr: s("1.2.3.4:5")}),
			EventPredicate: checkVoteEvent,
			EventCount:     1,
			Checker:        voteCheckerFactory,
		},
		&pollTest{
			Name:           "No user hidden change",
			Electorate:     db.ElectorateAll,
			Hidden:         true,
			Vote:           []pollTestVote{{User: 1, Alt: 1}},
			UserType:       pollTestUserTypeNone,
			Request:        fillRequest(UninominalVoteQuery{Alternative: 0}, srvt.Request{RemoteAddr: s("1.2.3.4:5")}),
			EventPredicate: checkVoteEvent,
			EventCount:     1,
			Checker:        voteCheckerFactory,
		},
		&pollTest{
			Name:           "Unlogged public change",
			Electorate:     db.ElectorateAll,
			Vote:           []pollTestVote{{User: 1, Alt: 1}},
			UserType:       pollTestUserTypeUnlogged,
			Request:        fillRequest(UninominalVoteQuery{Alternative: 0}, srvt.Request{}),
			EventPredicate: checkVoteEvent,
			EventCount:     1,
			Checker:        voteCheckerFactory,
		},
		&pollTest{
			Name:           "Unlogged hidden change",
			Electorate:     db.ElectorateAll,
			Hidden:         true,
			Vote:           []pollTestVote{{User: 1, Alt: 1}},
			UserType:       pollTestUserTypeUnlogged,
			Request:        fillRequest(UninominalVoteQuery{Alternative: 0}, srvt.Request{}),
			EventPredicate: checkVoteEvent,
			EventCount:     1,
			Checker:        voteCheckerFactory,
		},

		&pollTest{
			Name:           "No user registered",
			Electorate:     db.ElectorateLogged,
			UserType:       pollTestUserTypeNone,
			Request:        fillRequest(UninominalVoteQuery{Alternative: 0}, srvt.Request{}),
			EventPredicate: checkVoteEvent,
			Checker:        srvt.CheckError{Code: http.StatusForbidden, Body: "Unlogged"},
		},
		&pollTest{
			Name:           "No user hidden registered",
			Electorate:     db.ElectorateLogged,
			Hidden:         true,
			UserType:       pollTestUserTypeNone,
			Request:        fillRequest(UninominalVoteQuery{Alternative: 0}, srvt.Request{}),
			EventPredicate: checkVoteEvent,
			Checker:        srvt.CheckError{Code: http.StatusForbidden, Body: "Unlogged"},
		},
		&pollTest{
			Name:           "Unlogged registered",
			Electorate:     db.ElectorateLogged,
			UserType:       pollTestUserTypeUnlogged,
			Request:        fillRequest(UninominalVoteQuery{Alternative: 0}, srvt.Request{}),
			EventPredicate: checkVoteEvent,
			Checker:        srvt.CheckError{Code: http.StatusForbidden, Body: "Unlogged"},
		},
		&pollTest{
			Name:           "Unlogged hidden registered",
			Electorate:     db.ElectorateLogged,
			Hidden:         true,
			UserType:       pollTestUserTypeUnlogged,
			Request:        fillRequest(UninominalVoteQuery{Alternative: 0}, srvt.Request{}),
			EventPredicate: checkVoteEvent,
			Checker:        srvt.CheckError{Code: http.StatusForbidden, Body: "Unlogged"},
		},

		&pollTest{
			Name:           "No user public cookie",
			Electorate:     db.ElectorateAll,
			UserType:       pollTestUserTypeNone,
			Request:        fillRequest(UninominalVoteQuery{Alternative: 0}, srvt.Request{RemoteAddr: s("1.2.3.4:5")}),
			EventPredicate: checkVoteEvent,
			EventCount:     1,
			Checker:        srvt.CheckCookieIsSet{Name: server.SessionUnlogged},
		},
		&pollTest{
			Name:           "No user hidden cookie",
			Electorate:     db.ElectorateAll,
			Hidden:         true,
			UserType:       pollTestUserTypeNone,
			Request:        fillRequest(UninominalVoteQuery{Alternative: 0}, srvt.Request{RemoteAddr: s("1.2.3.4:5")}),
			EventPredicate: checkVoteEvent,
			EventCount:     1,
			Checker:        srvt.CheckCookieIsSet{Name: server.SessionUnlogged},
		},

		&pollTest{
			Name:           "Poll verified, User unverified",
			Electorate:     db.ElectorateVerified,
			UserType:       pollTestUserTypeLogged,
			Request:        fillRequest(UninominalVoteQuery{Alternative: 0}, srvt.Request{}),
			EventPredicate: checkVoteEvent,
			Checker:        srvt.CheckError{Code: http.StatusForbidden, Body: "Unverified"},
		},
		&pollTest{
			Name:           "Poll verified, User verified",
			Electorate:     db.ElectorateVerified,
			UserType:       pollTestUserTypeLogged,
			Verified:       true,
			Request:        fillRequest(UninominalVoteQuery{Alternative: 0}, srvt.Request{}),
			EventPredicate: checkVoteEvent,
			EventCount:     1,
			Checker:        voteCheckerFactory,
		},
	}

	srvt.Run(t, tests, UninominalVoteHandler)
}
