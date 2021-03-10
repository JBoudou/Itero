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
	"reflect"
	"testing"

	"github.com/JBoudou/Itero/db"
	dbt "github.com/JBoudou/Itero/db/dbtest"
	srvt "github.com/JBoudou/Itero/server/servertest"
)

func TestAllAlternatives(t *testing.T) {
	precheck(t)

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
	noVote  = uninominalBallot{Value: 0, State: uninominalBallotStateValid}
	yesVote = uninominalBallot{Value: 1, State: uninominalBallotStateValid}
	blankVote = uninominalBallot{State: uninominalBallotStateBlank}
)

func TestUninominalBallotAnswer_MarshalJSON(t *testing.T) {
	precheck(t)

	tests := []struct {
		name        string
		answer      UninominalBallotAnswer
		expectValue string
		expectError error
	}{
		{
			name: "Full",
			answer: UninominalBallotAnswer{
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
			answer: UninominalBallotAnswer{
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
			answer: UninominalBallotAnswer{
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
			answer: UninominalBallotAnswer{
				Alternatives: []PollAlternative{
					{Id: 0, Name: "No", Cost: 1.},
					{Id: 1, Name: "Yes", Cost: 1.},
				},
			},
			expectValue: `{"Alternatives":[{"Id":0,"Name":"No","Cost":1},{"Id":1,"Name":"Yes","Cost":1}]}`,
		},
		{
			name: "Blank Previous",
			answer: UninominalBallotAnswer{
				Previous: blankVote,
				Alternatives: []PollAlternative{
					{Id: 0, Name: "No", Cost: 1.},
					{Id: 1, Name: "Yes", Cost: 1.},
				},
			},
			expectValue: `{"PreviousIsBlank":true,"Alternatives":[{"Id":0,"Name":"No","Cost":1},
				{"Id":1,"Name":"Yes","Cost":1}]}`,
		},
		{
			name: "Blank Current",
			answer: UninominalBallotAnswer{
				Current: blankVote,
				Alternatives: []PollAlternative{
					{Id: 0, Name: "No", Cost: 1.},
					{Id: 1, Name: "Yes", Cost: 1.},
				},
			},
			expectValue: `{"CurrentIsBlank":true,"Alternatives":[{"Id":0,"Name":"No","Cost":1},
				{"Id":1,"Name":"Yes","Cost":1}]}`,
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

func TestUninominalBallotHandler(t *testing.T) {
	precheck(t)

	env := new(dbt.Env)
	defer env.Close()

	pollSegment := PollSegment{Salt: 42}
	userId := env.CreateUser()
	pollSegment.Id = env.CreatePoll("UninominalBallotHandler", userId, db.PollPublicityPublic)
	mustt(t, env.Error)

	request := makePollRequest(t, pollSegment, &userId)

	alternatives := []PollAlternative{
		{Id: 0, Name: "No", Cost: 1.},
		{Id: 1, Name: "Yes", Cost: 1.},
	}

	const qBlankVote = `DELETE FROM Ballots WHERE User = ? AND Round = ?`;

	vote := func(alternative, round uint8) func(t *testing.T) {
		return func(t *testing.T) {
			env.Vote(pollSegment.Id, round, userId, alternative)
			env.Must(t)
		}
	}

	blank := func(round uint8) func(t *testing.T) {
		return func(t *testing.T) {
			_, err := db.DB.Exec(qBlankVote, userId, round)
			mustt(t, err)
		}
	}

	// WARNING: The tests are sequential.

	tests := []srvt.Test{
		{
			Name:    "No Ballot",
			Request: request,
			Checker: srvt.CheckJSON{Body: &UninominalBallotAnswer{Alternatives: alternatives}},
		},
		{
			Name:    "Current ballot",
			Update:  vote(0, 0),
			Request: request,
			Checker: srvt.CheckJSON{Body: &UninominalBallotAnswer{
				Current:      noVote,
				Alternatives: alternatives,
			}},
		},
		{
			Name: "Previous ballot",
			Update: func(t *testing.T) {
				env.NextRound(pollSegment.Id)
				env.Must(t)
			},
			Request: request,
			Checker: srvt.CheckJSON{Body: &UninominalBallotAnswer{
				Previous:     noVote,
				Alternatives: alternatives,
			}},
		},
		{
			Name:    "Both ballots",
			Update:  vote(1, 1),
			Request: request,
			Checker: srvt.CheckJSON{Body: &UninominalBallotAnswer{
				Previous:     noVote,
				Current:      yesVote,
				Alternatives: alternatives,
			}},
		},
		{
			Name:    "Previous Blank",
			Update:  blank(0),
			Request: request,
			Checker: srvt.CheckJSON{Body: &UninominalBallotAnswer{
				Previous:     blankVote,
				Current:      yesVote,
				Alternatives: alternatives,
			}},
		},
		{
			Name:    "Both Blank",
			Update:  blank(1),
			Request: request,
			Checker: srvt.CheckJSON{Body: &UninominalBallotAnswer{
				Previous:     blankVote,
				Current:      blankVote,
				Alternatives: alternatives,
			}},
		},
	}
	srvt.RunFunc(t, tests, UninominalBallotHandler)
}
