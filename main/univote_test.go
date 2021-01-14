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
		SELECT p.LastRound, b.Alternative
		  FROM Participants AS p LEFT OUTER JOIN Ballots AS b
			  ON (p.Poll, p.User, p.LastRound) = (b.Poll, b.User, b.Round)
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
	pollSegment := PollSegment{Salt: 42}
	pollSegment.Id = env.CreatePoll("Test", userId, db.PollPublicityPublicRegistered)
	forbiddenSegment := PollSegment{Salt: 42}
	forbiddenSegment.Id = env.CreatePoll("Forbidden", userId, db.PollPublicityInvited)
	env.Must(t)

	makeRequest := func(userId *uint32, pollSegment PollSegment,
		vote UninominalVoteQuery) (req srvt.Request) {

		req = makePollRequest(t, pollSegment, userId)
		b, err := json.Marshal(vote)
		mustt(t, err)
		req.Body = string(b)
		req.Method = "POST"
		return
	}

	tests := []srvt.Test{
		{
			Name:    "No user",
			Request: makeRequest(nil, pollSegment, UninominalVoteQuery{}),
			Checker: srvt.CheckStatus{http.StatusForbidden},
		},
		{
			Name:    "Forbidden",
			Request: makeRequest(&userId, forbiddenSegment, UninominalVoteQuery{}),
			Checker: srvt.CheckStatus{http.StatusForbidden},
		},
		{
			Name:    "First vote",
			Request: makeRequest(&userId, pollSegment, UninominalVoteQuery{Alternative: 1}),
			Checker: &voteChecker{poll: pollSegment.Id, user: userId, round: 0},
		},
		{
			Name:    "Change vote",
			Request: makeRequest(&userId, pollSegment, UninominalVoteQuery{Alternative: 0}),
			Checker: &voteChecker{poll: pollSegment.Id, user: userId, round: 0},
		},
		{
			Name:    "Change vote again",
			Request: makeRequest(&userId, pollSegment, UninominalVoteQuery{Alternative: 1}),
			Checker: &voteChecker{poll: pollSegment.Id, user: userId, round: 0},
		},
		{
			Name: "First blank",
			Update: func(t *testing.T) {
				env.NextRound(pollSegment.Id)
				env.Must(t)
			},
			Request: makeRequest(&userId, pollSegment, UninominalVoteQuery{Blank: true}),
			Checker: &voteChecker{poll: pollSegment.Id, user: userId, round: 1},
		},
		{
			Name:    "Change to non-blank",
			Request: makeRequest(&userId, pollSegment, UninominalVoteQuery{Alternative: 1}),
			Checker: &voteChecker{poll: pollSegment.Id, user: userId, round: 1},
		},
		{
			Name:    "Change to blank",
			Request: makeRequest(&userId, pollSegment, UninominalVoteQuery{Blank: true}),
			Checker: &voteChecker{poll: pollSegment.Id, user: userId, round: 1},
		},
		{
			Name: "Inactive",
			Update: func(t *testing.T) {
				const qInactivate = `UPDATE Polls SET Active = FALSE WHERE Id = ?`
				_, err := db.DB.Exec(qInactivate, pollSegment.Id)
				mustt(t, err)
			},
			Request: makeRequest(&userId, pollSegment, UninominalVoteQuery{}),
			Checker: srvt.CheckStatus{http.StatusBadRequest},
		},
	}

	srvt.RunFunc(t, tests, UninominalVoteHandler)
}
