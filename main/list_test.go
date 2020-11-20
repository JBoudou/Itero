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

func TestPollSegment(t *testing.T) {
	tests := []struct {
		name    string
		segment PollSegment
	}{
		{
			name:    "Simple",
			segment: PollSegment{Id: 0x12345, Salt: 0x71234},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := tt.segment.Encode()
			if err != nil {
				t.Fatalf("Encode error: %s", err)
			}
			got, err := PollSegmentDecode(encoded)
			if err != nil {
				t.Fatalf("Decode error: %s", err)
			}
			if got != tt.segment {
				t.Errorf("Got %v. Expect %v", got, tt.segment)
			}
		})
	}
}

func TestListHandler(t *testing.T) {
	env := new(dbt.Env)
	defer env.Close()

	userId := env.CreateUser()
	if env.Error != nil {
		t.Fatalf("Env failed: %s", env.Error)
	}

	const (
		qCreatePoll = `INSERT INTO Polls(Title, Admin, Salt, NbChoices, Publicity, MaxNbRounds)
			 VALUE (?, ?, 42, 2, ?, 3)`
		qCreateAlternatives = `INSERT INTO Alternatives(Poll, Id, Name) VALUES (?, 0, 'No'), (?, 1, 'Yes')`
		qParticipate        = `INSERT INTO Participants(Poll, User) VALUE (?, ?)`

		qRemovePoll = `DELETE FROM Polls WHERE Id = ?`

		poll1Title = "Test 1"
		poll2Title = "Test 2"
	)

	var (
		poll1Id uint32
		poll2Id uint32
	)

	type maker = func(t *testing.T) listResponseEntry

	makePollEntry := func(title string, id *uint32) maker {
		return func(t *testing.T) listResponseEntry {
			segment, err := PollSegment{Id: *id, Salt: 42}.Encode()
			if err != nil {
				t.Fatal(err)
			}
			return listResponseEntry{Title: title, Segment: segment, CurrentRound: 0, MaxRound: 3}
		}
	}

	checker := func(polls []maker) srvt.Checker {
		return func(t *testing.T, response *http.Response, req *http.Request) {
			expect := make([]listResponseEntry, 0, len(polls))
			for _, maker := range polls {
				expect = append(expect, maker(t))
			}
			// TODO improve the checker such that it does not fail in production.
			srvt.CheckerJSON(http.StatusOK, expect)(t, response, req)
		}
	}

	tests := []srvt.Test{
		{
			Name: "No session",
			// TODO fix this test once implemented
			Checker: srvt.CheckerStatus(http.StatusNotImplemented),
		},
		{
			Name:    "No poll",
			Request: srvt.Request{UserId: &userId},
			// TODO fix this test such that it does not fail in production.
			Checker: srvt.CheckerJSON(http.StatusOK, []listResponseEntry{}),
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
			Checker: checker([]maker{makePollEntry(poll1Title, &poll1Id)}),
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
			Checker: checker([]maker{makePollEntry(poll1Title, &poll1Id)}),
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
				makePollEntry(poll1Title, &poll1Id),
				makePollEntry(poll2Title, &poll2Id),
			}),
		},
	}

	srvt.RunFunc(t, tests, ListHandler)
}
