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
	pollsCreated map[uint32]bool
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
			SELECT Title, Description, Admin, Salt, MinNbRounds, MaxNbRounds, Deadline,
			       CurrentRoundStart, ADDTIME(CurrentRoundStart, MaxRoundDuration), RoundThreshold
			  FROM Polls
			 WHERE Id = ?`
		qCheckAlternative = `SELECT Id, Name FROM Alternatives WHERE Poll = ?`
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
	var roundStart, roundEnd time.Time
	mustt(t, row.Scan(
		&got.Title,
		&got.Description,
		&admin,
		&salt,
		&got.MinNbRounds,
		&got.MaxNbRounds,
		&got.Deadline,
		&roundStart,
		&roundEnd,
		&got.RoundThreshold,
	))
	got.MaxRoundDuration = uint64(roundEnd.Sub(roundStart).Milliseconds())
	got.Deadline = got.Deadline.Truncate(time.Second)
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
	altMap := make(map[uint8]string, len(query.Alternatives))
	for _, alt := range query.Alternatives {
		altMap[alt.Id] = alt.Name
	}
	for rows.Next() {
		var id uint8
		var name string
		mustt(t, rows.Scan(&id, &name))
		expect, ok := altMap[id]
		if !ok {
			t.Errorf("Extraneous or duplicated alternative %d -> %s.", id, name)
			continue
		}
		delete(altMap, id)
		if name != expect {
			t.Errorf("Wrong alternative %d. Got %s. Expect %s.", id, name, expect)
		}
	}
	if len(altMap) > 0 {
		t.Errorf("Missing %d alternatives", len(altMap))
	}

	// Check events
	_, ok := self.pollsCreated[pollSegment.Id]
	if !ok {
		t.Errorf("CreatePollEvent not sent");
	}
}


func TestCreateHandler(t *testing.T) {
	precheck(t)
	
	var env dbt.Env
	defer env.Close()
	userId := env.CreateUser()
	env.Must(t)

	makeRequest := func(user *uint32, innerBody string, alternatives []string) srvt.Request {
		pollAlternatives := make([]PollAlternative, len(alternatives))
		for id, name := range alternatives {
			pollAlternatives[id] = PollAlternative{ Id: uint8(id), Name: name, Cost: 1. }
		}
		encoded, err := json.Marshal(pollAlternatives)
		mustt(t, err)
		return srvt.Request{
			UserId: user,
			Method: "POST",
			Body: `{"Title":"Test",` + innerBody + `"Alternatives": ` + string(encoded) + "}",
		}
	}

	tests := []srvt.Test{
		{
			Name: "No user",
			Request: makeRequest(nil, "", []string{"No", "Yes"}),
			Checker: srvt.CheckStatus{http.StatusForbidden},
		},
		{
			Name: "GET",
			Request: srvt.Request {
				UserId: &userId,
				Method: "GET",
				Body: `{
					"Title": "Test",
					"Alternatives": [{"Id":0, "Name":"No", "Cost":1}, {"Id":1, "Name":"Yes", "Cost":1}]
				}`,
			},
			Checker: srvt.CheckStatus{http.StatusForbidden},
		},
		{
			Name: "Duplicate",
			Request: srvt.Request {
				UserId: &userId,
				Method: "POST",
				Body: `{
					"Title": "Test",
					"Alternatives": [{"Id":0, "Name":"Yip", "Cost":1}, {"Id":0, "Name":"Yop", "Cost":1}]
				}`,
			},
			Checker: srvt.CheckAnyErrorStatus,
		},
		{
			Name: "Success",
			Request: makeRequest(&userId, "", []string{"No", "Yes"}),
			Checker: &createPollChecker{user: userId},
		},
		{
			Name: "Unordered",
			Request: srvt.Request {
				UserId: &userId,
				Method: "POST",
				Body: `{
					"Title": "Test",
					"Alternatives": [{"Id":1, "Name":"Second", "Cost":1}, {"Id":0, "Name":"First", "Cost":1}]
				}`,
			},
			Checker: &createPollChecker{user: userId},
		},
	}
			
	srvt.RunFunc(t, tests, CreateHandler)
}
