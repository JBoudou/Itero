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
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/JBoudou/Itero/db"
	dbt "github.com/JBoudou/Itero/db/dbtest"
	"github.com/JBoudou/Itero/events"
	"github.com/JBoudou/Itero/events/eventstest"
	"github.com/JBoudou/Itero/server"
	srvt "github.com/JBoudou/Itero/server/servertest"
)

type voteChecker struct {
	poll  uint32
	user  uint32
	round uint8

	originalEventManager events.Manager
	eventSent            bool
}

func (self *voteChecker) Before(t *testing.T) {
	self.eventSent = false
	self.originalEventManager = events.DefaultManager
	events.DefaultManager = &eventstest.ManagerMock{
		T: t,
		Send_: func(evt events.Event) error {
			if vEvt, ok := evt.(VoteEvent); ok && vEvt.Poll == self.poll {
				self.eventSent = true
			}
			return nil
		},
	}
}

func (self *voteChecker) Check(t *testing.T, response *http.Response, request *server.Request) {
	defer func() {
		events.DefaultManager = self.originalEventManager
	}()

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

	row := db.DB.QueryRow(qCheck, self.poll, self.user)
	var gotRound, gotAlternative sql.NullInt32
	err := row.Scan(&gotRound, &gotAlternative)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			t.Errorf("User %d does not participate in poll %d.", self.user, self.poll)
			return
		}
		t.Fatal(err)
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

	if !self.eventSent {
		t.Error("VoteEvent not sent")
	}
}

func TestUninominalVoteHandler(t *testing.T) {
	precheck(t)

	var env dbt.Env
	defer env.Close()
	userId := env.CreateUser()
	pollId := env.CreatePoll("Test", userId, db.PollPublicityPublicRegistered)
	forbiddenId := env.CreatePoll("Forbidden", userId, db.PollPublicityInvited)
	env.Must(t)

	makeRequest := func(userId *uint32, pollId uint32, vote UninominalVoteQuery) (req srvt.Request) {
		req = *makePollRequest(t, pollId, userId)
		b, err := json.Marshal(vote)
		mustt(t, err)
		req.Body = string(b)
		req.Method = "POST"
		return
	}

	tests := []srvt.Test{
		&srvt.T{
			Name:    "No user",
			Request: makeRequest(nil, pollId, UninominalVoteQuery{}),
			Checker: srvt.CheckStatus{http.StatusForbidden},
		},
		&srvt.T{
			Name:    "Wrong segment",
			Request: makeRequest(&userId, forbiddenId, UninominalVoteQuery{}),
			Checker: srvt.CheckStatus{http.StatusNotFound},
		},
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
			Name: "Previous round",
			Request: makeRequest(&userId, pollId, UninominalVoteQuery{Blank: true, Round: 0}),
			Checker: srvt.CheckError{Code: http.StatusLocked, Body: "Next round"},
		},
		&srvt.T{
			Name: "Next round",
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
	}

	srvt.RunFunc(t, tests, UninominalVoteHandler)
}
