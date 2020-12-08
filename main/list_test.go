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
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"reflect"
	"testing"

	"github.com/JBoudou/Itero/db"
	dbt "github.com/JBoudou/Itero/db/dbtest"
	srvt "github.com/JBoudou/Itero/server/servertest"
)

func TestListHandler(t *testing.T) {
	precheck(t)

	env := new(dbt.Env)
	defer env.Close()

	userId := env.CreateUser()
	if env.Error != nil {
		t.Fatalf("Env failed: %s", env.Error)
	}

	const (
		qParticipate = `INSERT INTO Participants(Poll, User) VALUE (?, ?)`

		poll1Title = "Test 1"
		poll2Title = "Test 2"
	)

	var (
		poll1Id uint32
		poll2Id uint32
	)

	type maker = func(t *testing.T) listAnswerEntry

	makePollEntry := func(title string, id *uint32, action string) maker {
		return func(t *testing.T) listAnswerEntry {
			segment, err := PollSegment{Id: *id, Salt: 42}.Encode()
			if err != nil {
				t.Fatal(err)
			}
			return listAnswerEntry{Title: title, Segment: segment, CurrentRound: 0, MaxRound: 3,
				Action: action}
		}
	}

	checker := func(include []maker, exclude []maker) srvt.Checker {
		return srvt.CheckerFun(func(t *testing.T, response *http.Response, req *http.Request) {
			if response.StatusCode != http.StatusOK {
				t.Errorf("Wrong status code. Got %d. Expect %d", response.StatusCode, http.StatusOK)
			}

			wanted := make(map[string]*listAnswerEntry, 2)
			for _, maker := range include {
				entry := maker(t)
				wanted[entry.Segment] = &entry
			}
			for _, maker := range exclude {
				entry := maker(t)
				wanted[entry.Segment] = nil
			}

			var got []listAnswerEntry
			var buff bytes.Buffer
			if _, err := buff.ReadFrom(response.Body); err != nil {
				t.Fatalf("Error reading body: %s", err)
			}
			if err := json.Unmarshal(buff.Bytes(), &got); err != nil {
				t.Fatalf("Error reading body: %s", err)
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
		})
	}

	tests := []srvt.Test{
		{
			Name: "No session",
			// TODO fix this test once implemented
			Checker: srvt.CheckStatus{http.StatusNotImplemented},
		},
		{
			Name: "PublicRegistered Poll",
			Update: func(t *testing.T) {
				poll1Id = env.CreatePoll(poll1Title, userId, db.PollPublicityPublicRegistered)
				if env.Error != nil {
					t.Fatalf("Env: %s", env.Error)
				}
			},
			Request: srvt.Request{UserId: &userId},
			Checker: checker([]maker{makePollEntry(poll1Title, &poll1Id, "Part")}, []maker{}),
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
			Checker: checker(
				[]maker{makePollEntry(poll1Title, &poll1Id, "Part")},
				[]maker{makePollEntry(poll2Title, &poll2Id, "Vote")}),
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
			Checker: checker([]maker{
				makePollEntry(poll1Title, &poll1Id, "Part"),
				makePollEntry(poll2Title, &poll2Id, "Vote"),
			}, []maker{}),
		},
	}

	srvt.RunFunc(t, tests, ListHandler)
}
