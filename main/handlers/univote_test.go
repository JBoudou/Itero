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
	"github.com/JBoudou/Itero/pkg/events/eventstest"
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
			if vEvt, ok := evt.(services.VoteEvent); ok && vEvt.Poll == self.poll {
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

	if !self.eventSent {
		t.Error("VoteEvent not sent")
	}
}

func voteCheckerFactory(param PollTestCheckerFactoryParam) srvt.Checker {
	return &voteChecker{poll: param.PollId, user: param.UserId, round: param.Round}
}

func TestUninominalVoteHandler(t *testing.T) {
	precheck(t)

	var env dbt.Env
	defer env.Close()
	userId := env.CreateUser()
	pollId := env.CreatePoll("Test", userId, db.PollPublicityPublicRegistered)
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
			Name:       "No user public",
			Sequential: true,
			Publicity:  db.PollPublicityPublic,
			UserType:   pollTestUserTypeNone,
			Request:    fillRequest(UninominalVoteQuery{Alternative: 0}, srvt.Request{RemoteAddr: s("1.2.3.4:5")}),
			Checker:    voteCheckerFactory,
		},
		&pollTest{
			Name:       "No user hidden",
			Sequential: true,
			Publicity:  db.PollPublicityHidden,
			UserType:   pollTestUserTypeNone,
			Request:    fillRequest(UninominalVoteQuery{Alternative: 0}, srvt.Request{RemoteAddr: s("1.2.3.4:5")}),
			Checker:    voteCheckerFactory,
		},
		&pollTest{
			Name:       "Unlogged public",
			Sequential: true,
			Publicity:  db.PollPublicityPublic,
			UserType:   pollTestUserTypeUnlogged,
			Request:    fillRequest(UninominalVoteQuery{Alternative: 0}, srvt.Request{}),
			Checker:    voteCheckerFactory,
		},
		&pollTest{
			Name:       "Unlogged hidden",
			Sequential: true,
			Publicity:  db.PollPublicityHidden,
			UserType:   pollTestUserTypeUnlogged,
			Request:    fillRequest(UninominalVoteQuery{Alternative: 0}, srvt.Request{}),
			Checker:    voteCheckerFactory,
		},

		&pollTest{
			Name:       "No user public change",
			Sequential: true,
			Publicity:  db.PollPublicityPublic,
			Vote:       []pollTestVote{{User: 1, Alt: 1}},
			UserType:   pollTestUserTypeNone,
			Request:    fillRequest(UninominalVoteQuery{Alternative: 0}, srvt.Request{RemoteAddr: s("1.2.3.4:5")}),
			Checker:    voteCheckerFactory,
		},
		&pollTest{
			Name:       "No user hidden change",
			Sequential: true,
			Publicity:  db.PollPublicityHidden,
			Vote:       []pollTestVote{{User: 1, Alt: 1}},
			UserType:   pollTestUserTypeNone,
			Request:    fillRequest(UninominalVoteQuery{Alternative: 0}, srvt.Request{RemoteAddr: s("1.2.3.4:5")}),
			Checker:    voteCheckerFactory,
		},
		&pollTest{
			Name:       "Unlogged public change",
			Sequential: true,
			Publicity:  db.PollPublicityPublic,
			Vote:       []pollTestVote{{User: 1, Alt: 1}},
			UserType:   pollTestUserTypeUnlogged,
			Request:    fillRequest(UninominalVoteQuery{Alternative: 0}, srvt.Request{}),
			Checker:    voteCheckerFactory,
		},
		&pollTest{
			Name:       "Unlogged hidden change",
			Sequential: true,
			Publicity:  db.PollPublicityHidden,
			Vote:       []pollTestVote{{User: 1, Alt: 1}},
			UserType:   pollTestUserTypeUnlogged,
			Request:    fillRequest(UninominalVoteQuery{Alternative: 0}, srvt.Request{}),
			Checker:    voteCheckerFactory,
		},

		&pollTest{
			Name:      "No user registered",
			Publicity: db.PollPublicityPublicRegistered,
			UserType:  pollTestUserTypeNone,
			Request:   fillRequest(UninominalVoteQuery{Alternative: 0}, srvt.Request{}),
			Checker:   srvt.CheckStatus{http.StatusNotFound},
		},
		&pollTest{
			Name:      "No user hidden registered",
			Publicity: db.PollPublicityHiddenRegistered,
			UserType:  pollTestUserTypeNone,
			Request:   fillRequest(UninominalVoteQuery{Alternative: 0}, srvt.Request{}),
			Checker:   srvt.CheckStatus{http.StatusNotFound},
		},
		&pollTest{
			Name:      "Unlogged registered",
			Publicity: db.PollPublicityPublicRegistered,
			UserType:  pollTestUserTypeUnlogged,
			Request:   fillRequest(UninominalVoteQuery{Alternative: 0}, srvt.Request{}),
			Checker:   srvt.CheckStatus{http.StatusNotFound},
		},
		&pollTest{
			Name:      "Unlogged hidden registered",
			Publicity: db.PollPublicityHiddenRegistered,
			UserType:  pollTestUserTypeUnlogged,
			Request:   fillRequest(UninominalVoteQuery{Alternative: 0}, srvt.Request{}),
			Checker:   srvt.CheckStatus{http.StatusNotFound},
		},

		&pollTest{
			Name:      "No user public cookie",
			Publicity: db.PollPublicityPublic,
			UserType:  pollTestUserTypeNone,
			Request:   fillRequest(UninominalVoteQuery{Alternative: 0}, srvt.Request{RemoteAddr: s("1.2.3.4:5")}),
			Checker:   srvt.CheckCookieIsSet{Name: server.SessionUnlogged},
		},
		&pollTest{
			Name:      "No user hidden cookie",
			Publicity: db.PollPublicityHidden,
			UserType:  pollTestUserTypeNone,
			Request:   fillRequest(UninominalVoteQuery{Alternative: 0}, srvt.Request{RemoteAddr: s("1.2.3.4:5")}),
			Checker:   srvt.CheckCookieIsSet{Name: server.SessionUnlogged},
		},
	}

	srvt.RunFunc(t, tests, UninominalVoteHandler)
}
