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
	"net/http"
	"reflect"
	"testing"

	"github.com/JBoudou/Itero/db"
	dbt "github.com/JBoudou/Itero/db/dbtest"
	"github.com/JBoudou/Itero/server"
	srvt "github.com/JBoudou/Itero/server/servertest"
)

type listChecker struct {
	publicInc []listCheckerEntry
	publicExc []listCheckerEntry
	ownInc []listCheckerEntry
	ownExc []listCheckerEntry
}

type listCheckerEntry struct {
	title string
	id *uint32
	action uint8
}

func (self *listCheckerEntry) toListEntry(t *testing.T) *listAnswerEntry {
	segment, err := PollSegment{Id: *self.id, Salt: 42}.Encode()
	mustt(t, err)
	return &listAnswerEntry{
		Title: self.title,
		Segment: segment,
		CurrentRound: 0,
		MaxRound: 4,
		Action: self.action,
	}
}

func (self listChecker) Before(t *testing.T) {
}

func (self listChecker) Check(t *testing.T, response *http.Response, request *server.Request) {
	srvt.CheckStatus{http.StatusOK}.Check(t, response, request)
	
	var answer ListAnswer
	mustt(t, json.NewDecoder(response.Body).Decode(&answer))

	listCheckList(t, answer.Public, self.publicInc, self.publicExc)
	listCheckList(t, answer.Own   , self.ownInc   , self.ownExc   )
}

func listCheckList(t *testing.T, got []listAnswerEntry,
	include []listCheckerEntry, exclude []listCheckerEntry) {

	wanted := make(map[string]*listAnswerEntry, len(include) + len(exclude))
	for _, maker := range include {
		entry := maker.toListEntry(t)
		wanted[entry.Segment] = entry
	}
	for _, maker := range exclude {
		entry := maker.toListEntry(t)
		wanted[entry.Segment] = nil
	}

	for _, entry := range got {
		entry.Deadline = NuDate(sql.NullTime{Valid: false})
		want, ok := wanted[entry.Segment]
		if !ok {
			continue
		}
		if want == nil {
			t.Errorf("Unwanted %v", entry)
		} else if !reflect.DeepEqual(entry, *want) {
			t.Errorf("Got %v. Expect %v", entry, *want)
		}
		delete(wanted, entry.Segment)
	}

	for _, value := range wanted {
		if value != nil {
			t.Errorf("Missing %v", value)
		}
	}
}


func TestListHandler(t *testing.T) {
	// BEWARE! This test is sequential!
	precheck(t)

	env := new(dbt.Env)
	defer env.Close()

	userId := env.CreateUser()
	otherId := env.CreateUserWith("other")
	env.Must(t)

	const (
		qParticipate = `INSERT INTO Participants(Poll, User, Round) VALUE (?, ?, 0)`
		qTerminate = `UPDATE Polls SET State = 'Terminated' WHERE Id = ?`
		qWaiting = `
		  UPDATE Polls
			   SET State = 'Waiting', Start = ADDTIME(CURRENT_TIMESTAMP(), '1:00')
			 WHERE Id = ?`

		poll1Title = "Test-1"
		poll2Title = "Test-2"
		poll3Title = "Test-3"
	)

	var (
		poll1Id uint32
		poll2Id uint32
		poll3Id uint32
	)

	tests := []srvt.Test{
		// BEWARE! This test is sequential!
		{
			Name: "No session",
			// TODO fix this test once implemented
			Checker: srvt.CheckStatus{http.StatusNotImplemented},
		},
		{
			Name: "PublicRegistered Poll",
			Update: func(t *testing.T) {
				poll1Id = env.CreatePoll(poll1Title, userId, db.PollPublicityPublicRegistered)
				env.Must(t)
			},
			Request: srvt.Request{UserId: &userId},
			Checker: listChecker{
				ownInc: []listCheckerEntry{
					{title: poll1Title, id: &poll1Id, action: PollActionPart},
				},
			},
		},
		{
			Name: "Other participate",
			Update: func(t *testing.T) {
				_, err := db.DB.Exec(qParticipate, poll1Id, otherId)
				if err != nil {
					t.Fatal(err)
				}
			},
			Request: srvt.Request{UserId: &userId},
			Checker: listChecker{
				ownInc: []listCheckerEntry{
					{title: poll1Title, id: &poll1Id, action: PollActionPart},
				},
			},
		},
		{
			Name: "HiddenRegistered Poll",
			Update: func(t *testing.T) {
				poll2Id = env.CreatePoll(poll2Title, userId, db.PollPublicityHiddenRegistered)
				if env.Error != nil {
					t.Fatalf("Env: %s", env.Error)
				}
			},
			Request: srvt.Request{UserId: &userId},
			Checker: listChecker{
				ownInc: []listCheckerEntry{
					{title: poll1Title, id: &poll1Id, action: PollActionPart},
					{title: poll2Title, id: &poll2Id, action: PollActionPart},
				},
			},
		},
		{
			Name: "HiddenRegistered is hidden",
			Request: srvt.Request{UserId: &otherId},
			Checker: listChecker{
				publicInc: []listCheckerEntry{
					{title: poll1Title, id: &poll1Id, action: PollActionModif},
				},
				publicExc: []listCheckerEntry{
					{title: poll2Title, id: &poll2Id, action: PollActionPart},
				},
			},
		},
		{
			Name: "HiddenRegistered Poll Participate",
			Update: func(t *testing.T) {
				_, err := db.DB.Exec(qParticipate, poll2Id, userId)
				if err != nil {
					t.Fatal(err)
				}
			},
			Request: srvt.Request{UserId: &userId},
			Checker: listChecker{
				ownInc: []listCheckerEntry{
					{title: poll1Title, id: &poll1Id, action: PollActionPart},
					{title: poll2Title, id: &poll2Id, action: PollActionModif},
				},
			},
		},
		{
			Name: "Terminated",
			Update: func(t *testing.T) {
				_, err := db.DB.Exec(qTerminate, poll2Id)
				if err != nil {
					t.Fatal(err)
				}
			},
			Request: srvt.Request{UserId: &userId},
			Checker: listChecker{
				ownInc: []listCheckerEntry{
					{title: poll1Title, id: &poll1Id, action: PollActionPart},
					{title: poll2Title, id: &poll2Id, action: PollActionTerm},
				},
			},
		},
		{
			Name: "Waiting",
			Update: func(t *testing.T) {
				poll3Id = env.CreatePoll(poll3Title, userId, db.PollPublicityPublicRegistered)
				env.Must(t)
				_, err := db.DB.Exec(qWaiting, poll3Id)
				if err != nil {
					t.Fatal(err)
				}
			},
			Request: srvt.Request{UserId: &userId},
			Checker: listChecker{
				ownInc: []listCheckerEntry{ {title: poll3Title, id: &poll3Id, action: PollActionWait}, },
			},
		},
		{
			Name: "Waiting is hidden",
			Request: srvt.Request{UserId: &userId},
			Checker: listChecker{
				publicExc: []listCheckerEntry{ {title: poll3Title, id: &poll3Id, action: PollActionWait}, },
			},
		},
	}

	srvt.RunFunc(t, tests, ListHandler)
}
