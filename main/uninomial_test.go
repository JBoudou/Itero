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
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"testing"

	"github.com/JBoudou/Itero/db"
	dbt "github.com/JBoudou/Itero/db/dbtest"
	srvt "github.com/JBoudou/Itero/server/servertest"
)

func mustt(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func TestAllAlternatives(t *testing.T) {
	var env dbt.Env
	defer env.Close()

	pollInfo := PollInfo{NbChoices: 2}
	userId := env.CreateUser()
	pollInfo.Id = env.CreatePoll("AllAlternatives", userId, db.PollPublicityPublic)
	mustt(t, env.Error)

	var got []PollAlternative
	allAlternatives(context.Background(), pollInfo, &got)

	expect := []PollAlternative{
		{Id: 0, Name: "No", Cost: 1.},
		{Id: 1, Name: "Yes", Cost: 1.},
	}
	if !reflect.DeepEqual(got, expect) {
		t.Errorf("Got %v. Expect %v", got, expect)
	}
}

var (
	noVote  = NullUInt8{Value: 0, Valid: true}
	yesVote = NullUInt8{Value: 1, Valid: true}
)

func TestUninomialBallotAnswer_MarshalJSON(t *testing.T) {
	tests := []struct {
		name        string
		answer      UninomialBallotAnswer
		expectValue string
		expectError error
	}{
		{
			name: "Full",
			answer: UninomialBallotAnswer{
				Previous: noVote,
				Current:  noVote,
				Alternatives: []PollAlternative{
					{Id: 0, Name: "No", Cost: 1.},
					{Id: 1, Name: "Yes", Cost: 1.},
				},
			},
			expectValue: `{"Previous":0,"Current":0,"Alternatives":[{"Id":0,"Name":"No","Cost":1},
				{"Id":1,"Name":"Yes","Cost":1}]}`,
		},
		{
			name: "No Previous",
			answer: UninomialBallotAnswer{
				Current: noVote,
				Alternatives: []PollAlternative{
					{Id: 0, Name: "No", Cost: 1.},
					{Id: 1, Name: "Yes", Cost: 1.},
				},
			},
			expectValue: `{"Current":0,"Alternatives":[{"Id":0,"Name":"No","Cost":1},
				{"Id":1,"Name":"Yes","Cost":1}]}`,
		},
		{
			name: "No Current",
			answer: UninomialBallotAnswer{
				Previous: noVote,
				Alternatives: []PollAlternative{
					{Id: 0, Name: "No", Cost: 1.},
					{Id: 1, Name: "Yes", Cost: 1.},
				},
			},
			expectValue: `{"Previous":0,"Alternatives":[{"Id":0,"Name":"No","Cost":1},
				{"Id":1,"Name":"Yes","Cost":1}]}`,
		},
		{
			name: "No Ballot",
			answer: UninomialBallotAnswer{
				Alternatives: []PollAlternative{
					{Id: 0, Name: "No", Cost: 1.},
					{Id: 1, Name: "Yes", Cost: 1.},
				},
			},
			expectValue: `{"Alternatives":[{"Id":0,"Name":"No","Cost":1},{"Id":1,"Name":"Yes","Cost":1}]}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotJSON, err := tt.answer.MarshalJSON()
			if err != tt.expectError {
				t.Errorf("Got error %s. Expect %s", err, tt.expectError)
			}
			if tt.expectError != nil {
				return
			}
			if !json.Valid(gotJSON) {
				t.Errorf("Invalid JSON: %s", gotJSON)
			}

			var got map[string]interface{}
			err = json.Unmarshal(gotJSON, &got)
			if err != nil {
				t.Fatal(err)
			}
			var expect map[string]interface{}
			err = json.Unmarshal([]byte(tt.expectValue), &expect)
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(got, expect) {
				t.Errorf("Got %v. Expect %v.", got, expect)
			}
		})
	}
}

func TestUninomialBallotHandler(t *testing.T) {
	precheck(t)

	env := new(dbt.Env)
	defer env.Close()

	pollSegment := PollSegment{Salt: 42}
	userId := env.CreateUser()
	pollSegment.Id = env.CreatePoll("UninomialBallotHandler", userId, db.PollPublicityPublic)
	mustt(t, env.Error)

	encoded, err := pollSegment.Encode()
	mustt(t, err)
	target := "/a/test/" + encoded
	request := srvt.Request{Target: &target, UserId: &userId}

	alternatives := []PollAlternative{
		{Id: 0, Name: "No", Cost: 1.},
		{Id: 1, Name: "Yes", Cost: 1.},
	}

	const (
		qVote      = `INSERT INTO Ballots(User, Poll, Alternative, Round) VALUE (?, ?, ?, ?)`
		qNextRound = `UPDATE Polls SET CurrentRound = CurrentRound + 1 WHERE Id = ?`
	)

	vote := func(alternative, round uint8) func(t *testing.T) {
		return func(t *testing.T) {
			_, err := db.DB.Exec(qVote, userId, pollSegment.Id, alternative, round)
			mustt(t, err)
		}
	}

	tests := []srvt.Test{
		{
			Name:    "No Ballot",
			Request: request,
			Checker: srvt.CheckJSON{http.StatusOK, &UninomialBallotAnswer{Alternatives: alternatives}},
		},
		{
			Name:    "Current ballot",
			Update:  vote(0, 0),
			Request: request,
			Checker: srvt.CheckJSON{http.StatusOK, &UninomialBallotAnswer{
				Current:      noVote,
				Alternatives: alternatives,
			}},
		},
		{
			Name: "Previous ballot",
			Update: func(t *testing.T) {
				_, err := db.DB.Exec(qNextRound, pollSegment.Id)
				mustt(t, err)
			},
			Request: request,
			Checker: srvt.CheckJSON{http.StatusOK, &UninomialBallotAnswer{
				Previous:     noVote,
				Alternatives: alternatives,
			}},
		},
		{
			Name:    "Both ballots",
			Update:  vote(1, 1),
			Request: request,
			Checker: srvt.CheckJSON{http.StatusOK, &UninomialBallotAnswer{
				Previous:     noVote,
				Current:      yesVote,
				Alternatives: alternatives,
			}},
		},
	}
	srvt.RunFunc(t, tests, UninomialBallotHandler)
}
