// Itero - Online iterative vote application
// Copyright (C) 2021 Joseph Boudou
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
	"encoding/json"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/JBoudou/Itero/db"
	dbt "github.com/JBoudou/Itero/db/dbtest"
	"github.com/JBoudou/Itero/events"
	"github.com/JBoudou/Itero/events/eventstest"
	"github.com/JBoudou/Itero/server"
	srvt "github.com/JBoudou/Itero/server/servertest"
)

type createPollChecker struct {
	user uint32

	originalEventManager events.Manager
	pollsCreated         map[uint32]bool
}

func (self *createPollChecker) Before(t *testing.T) {
	self.pollsCreated = make(map[uint32]bool, 1)
	self.originalEventManager = events.DefaultManager
	events.DefaultManager = &eventstest.ManagerMock{
		T: t,
		Send_: func(evt events.Event) error {
			if vEvt, ok := evt.(CreatePollEvent); ok {
				self.pollsCreated[vEvt.Poll] = true
			}
			return nil
		},
	}
}

func (self *createPollChecker) Check(t *testing.T, response *http.Response, request *server.Request) {
	defer func() {
		events.DefaultManager = self.originalEventManager
	}()

	srvt.CheckStatus{http.StatusOK}.Check(t, response, request)

	const (
		qCheckPoll = `
			SELECT Title, Description, Admin, Salt, Publicity, MinNbRounds, MaxNbRounds, Deadline,
			       CurrentRoundStart, ADDTIME(CurrentRoundStart, MaxRoundDuration), RoundThreshold
			  FROM Polls
			 WHERE Id = ?`
		qCheckAlternative = `SELECT Name FROM Alternatives WHERE Poll = ? ORDER BY Id ASC`
		qCleanUp          = `DELETE FROM Polls WHERE Id = ?`
	)

	var answer string
	mustt(t, json.NewDecoder(response.Body).Decode(&answer))
	pollSegment, err := PollSegmentDecode(answer)
	mustt(t, err)
	defer func() {
		db.DB.Exec(qCleanUp, pollSegment.Id)
	}()

	query := defaultCreateQuery()
	mustt(t, request.UnmarshalJSONBody(&query))
	query.Deadline = query.Deadline.Truncate(time.Second)

	// Check Polls
	row := db.DB.QueryRow(qCheckPoll, pollSegment.Id)
	got := query
	var admin uint32
	var salt uint32
	var publicity uint8
	var roundStart, roundEnd time.Time
	mustt(t, row.Scan(
		&got.Title,
		&got.Description,
		&admin,
		&salt,
		&publicity,
		&got.MinNbRounds,
		&got.MaxNbRounds,
		&got.Deadline,
		&roundStart,
		&roundEnd,
		&got.RoundThreshold,
	))
	got.MaxRoundDuration = uint64(roundEnd.Sub(roundStart).Milliseconds())
	got.Deadline = got.Deadline.Truncate(time.Second)
	got.Hidden = (publicity == db.PollPublicityHidden) ||
		(publicity == db.PollPublicityHiddenRegistered)
	if salt != pollSegment.Salt {
		t.Errorf("Wrong salt. Got %d. Expect %d.", salt, pollSegment.Salt)
	}
	if admin != self.user {
		t.Errorf("Wrong admin. Got %d. Expect %d.", admin, self.user)
	}
	if !reflect.DeepEqual(got, query) {
		t.Errorf("Got %v. Expect %v.", got, query)
	}

	// Check Alternatives
	rows, err := db.DB.Query(qCheckAlternative, pollSegment.Id)
	mustt(t, err)
	for id, alt := range query.Alternatives {
		if (!rows.Next()) {
			t.Errorf("Premature end of the alternatives. Got %d. Expect %d.", id, len(query.Alternatives))
			break
		}
		var name string
		mustt(t, rows.Scan(&name))
		if name != alt.Name {
			t.Errorf("Wrong alternative %d. Got %s. Expect %s.", id, name, alt.Name)
		}
	}
	if (rows.Next()) {
		t.Errorf("Unexpected alternatives.")
		rows.Close()
	}

	// Check events
	_, ok := self.pollsCreated[pollSegment.Id]
	if !ok {
		t.Errorf("CreatePollEvent not sent")
	}
}

func TestCreateHandler(t *testing.T) {
	precheck(t)

	var env dbt.Env
	defer env.Close()
	userId := env.CreateUser()
	env.Must(t)

	makeRequest := func(user *uint32, innerBody string, alternatives []string) srvt.Request {
		pollAlternatives := make([]SimpleAlternative, len(alternatives))
		for id, name := range alternatives {
			pollAlternatives[id] = SimpleAlternative{Name: name, Cost: 1.}
		}
		encoded, err := json.Marshal(pollAlternatives)
		mustt(t, err)
		return srvt.Request{
			UserId: user,
			Method: "POST",
			Body:   `{"Title":"Test",` + innerBody + `"Alternatives": ` + string(encoded) + "}",
		}
	}

	tests := []srvt.Test{
		{
			Name:    "No user",
			Request: makeRequest(nil, "", []string{"No", "Yes"}),
			Checker: srvt.CheckStatus{http.StatusForbidden},
		},
		{
			Name: "GET",
			Request: srvt.Request{
				UserId: &userId,
				Method: "GET",
				Body: `{
					"Title": "Test",
					"Alternatives": [{"Name":"No", "Cost":1}, {"Name":"Yes", "Cost":1}]
				}`,
			},
			Checker: srvt.CheckStatus{http.StatusForbidden},
		},
		{
			Name: "Duplicate",
			Request: srvt.Request{
				UserId: &userId,
				Method: "POST",
				Body: `{
					"Title": "Test",
					"Alternatives": [{"Name":"Yip", "Cost":1}, {"Name":"Yip", "Cost":1}]
				}`,
			},
			Checker: srvt.CheckAnyErrorStatus,
		},
		{
			Name:    "Success",
			Request: makeRequest(&userId, "", []string{"No", "Yes"}),
			Checker: &createPollChecker{user: userId},
		},
		{
			Name: "Hidden",
			Request: srvt.Request{
				UserId: &userId,
				Method: "POST",
				Body: `{
					"Title": "Test",
					"Alternatives": [{"Name":"First", "Cost":1}, {"Name":"Second", "Cost":1}],
					"Hidden": true
				}`,
			},
			Checker: &createPollChecker{user: userId},
		},
	}

	srvt.RunFunc(t, tests, CreateHandler)
}
