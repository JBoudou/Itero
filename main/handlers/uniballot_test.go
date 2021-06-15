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
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"testing"

	"github.com/JBoudou/Itero/mid/db"
	dbt "github.com/JBoudou/Itero/mid/db/dbtest"
	srvt "github.com/JBoudou/Itero/mid/server/servertest"
)

func TestAllAlternatives(t *testing.T) {
	precheck(t)

	var env dbt.Env
	defer env.Close()

	pollInfo := PollInfo{NbChoices: 2}
	userId := env.CreateUser()
	pollInfo.Id = env.CreatePoll("AllAlternatives", userId, db.ElectorateAll)
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
	noVote        = uninominalBallot{Value: 0, State: uninominalBallotStateValid}
	yesVote       = uninominalBallot{Value: 1, State: uninominalBallotStateValid}
	blankVote     = uninominalBallot{State: uninominalBallotStateBlank}
	undefinedVote = uninominalBallot{State: uninominalBallotStateUndefined}
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

	userId := env.CreateUser()
	pollId := env.CreatePoll("UninominalBallotHandler", userId, db.ElectorateAll)
	mustt(t, env.Error)

	request := *makePollRequest(t, pollId, &userId)

	alternatives := []PollAlternative{
		{Id: 0, Name: "No", Cost: 1.},
		{Id: 1, Name: "Yes", Cost: 1.},
	}

	const qBlankVote = `DELETE FROM Ballots WHERE User = ? AND Round = ?`

	vote := func(alternative, round uint8) func(t *testing.T) {
		return func(t *testing.T) {
			env.Vote(pollId, round, userId, alternative)
			env.Must(t)
		}
	}

	blank := func(round uint8) func(t *testing.T) {
		return func(t *testing.T) {
			_, err := db.DB.Exec(qBlankVote, userId, round)
			mustt(t, err)
		}
	}

	s := func(s string) *string { return &s }

	tests := []srvt.Test{

		// Sequential tests first //
		// TODO make them all independent

		&srvt.T{
			Name:    "No Ballot",
			Request: request,
			Checker: srvt.CheckJSON{Body: &UninominalBallotAnswer{Alternatives: alternatives}},
		},
		&srvt.T{
			Name:    "Current ballot",
			Update:  vote(0, 0),
			Request: request,
			Checker: srvt.CheckJSON{Body: &UninominalBallotAnswer{
				Current:      noVote,
				Alternatives: alternatives,
			}},
		},
		&srvt.T{
			Name: "Previous ballot",
			Update: func(t *testing.T) {
				env.NextRound(pollId)
				env.Must(t)
			},
			Request: request,
			Checker: srvt.CheckJSON{Body: &UninominalBallotAnswer{
				Previous:     noVote,
				Alternatives: alternatives,
			}},
		},
		&srvt.T{
			Name:    "Both ballots",
			Update:  vote(1, 1),
			Request: request,
			Checker: srvt.CheckJSON{Body: &UninominalBallotAnswer{
				Previous:     noVote,
				Current:      yesVote,
				Alternatives: alternatives,
			}},
		},
		&srvt.T{
			Name:    "Previous Blank",
			Update:  blank(0),
			Request: request,
			Checker: srvt.CheckJSON{Body: &UninominalBallotAnswer{
				Previous:     blankVote,
				Current:      yesVote,
				Alternatives: alternatives,
			}},
		},
		&srvt.T{
			Name:    "Both Blank",
			Update:  blank(1),
			Request: request,
			Checker: srvt.CheckJSON{Body: &UninominalBallotAnswer{
				Previous:     blankVote,
				Current:      blankVote,
				Alternatives: alternatives,
			}},
		},

		// Independent tests //

		&pollTest{
			Name:       "No user public",
			Electorate: db.ElectorateAll,
			Vote:       []pollTestVote{{1, 0, 0}},
			UserType:   pollTestUserTypeNone,
			Checker: srvt.CheckJSON{Body: &UninominalBallotAnswer{
				Previous:     undefinedVote,
				Current:      undefinedVote,
				Alternatives: alternatives,
			}},
		},
		&pollTest{
			Name:       "No user Hidden",
			Electorate: db.ElectorateAll,
			Hidden:     true,
			Vote:       []pollTestVote{{1, 0, 0}},
			UserType:   pollTestUserTypeNone,
			Checker: srvt.CheckJSON{Body: &UninominalBallotAnswer{
				Previous:     undefinedVote,
				Current:      undefinedVote,
				Alternatives: alternatives,
			}},
		},
		&pollTest{
			Name:       "Same addr public",
			Electorate: db.ElectorateAll,
			Vote:       []pollTestVote{{1, 0, 0}},
			UserType:   pollTestUserTypeNone,
			Request:    srvt.Request{RemoteAddr: s("1.2.3.4:56")},
			Checker: srvt.CheckJSON{Body: &UninominalBallotAnswer{
				Previous:     undefinedVote,
				Current:      noVote,
				Alternatives: alternatives,
			}},
		},
		&pollTest{
			Name:       "Same addr Hidden",
			Electorate: db.ElectorateAll,
			Hidden:     true,
			Vote:       []pollTestVote{{1, 0, 0}},
			UserType:   pollTestUserTypeNone,
			Request:    srvt.Request{RemoteAddr: s("9.2.3.4:56")},
			Checker: srvt.CheckJSON{Body: &UninominalBallotAnswer{
				Previous:     undefinedVote,
				Current:      noVote,
				Alternatives: alternatives,
			}},
		},
		&pollTest{
			Name:       "Unlogged public",
			Electorate: db.ElectorateAll,
			Vote:       []pollTestVote{{1, 0, 0}},
			UserType:   pollTestUserTypeUnlogged,
			Checker: srvt.CheckJSON{Body: &UninominalBallotAnswer{
				Previous:     undefinedVote,
				Current:      noVote,
				Alternatives: alternatives,
			}},
		},
		&pollTest{
			Name:       "Unlogged hidden",
			Electorate: db.ElectorateAll,
			Hidden:     true,
			Vote:       []pollTestVote{{1, 0, 0}},
			UserType:   pollTestUserTypeUnlogged,
			Checker: srvt.CheckJSON{Body: &UninominalBallotAnswer{
				Previous:     undefinedVote,
				Current:      noVote,
				Alternatives: alternatives,
			}},
		},

		&pollTest{
			Name:       "No user registered",
			Electorate: db.ElectorateLogged,
			UserType:   pollTestUserTypeNone,
			Checker:    srvt.CheckStatus{http.StatusNotFound},
		},
		&pollTest{
			Name:       "No user hidden registered",
			Electorate: db.ElectorateLogged,
			Hidden:     true,
			UserType:   pollTestUserTypeNone,
			Checker:    srvt.CheckStatus{http.StatusNotFound},
		},
		&pollTest{
			Name:       "Unlogged registered",
			Electorate: db.ElectorateLogged,
			UserType:   pollTestUserTypeUnlogged,
			Checker:    srvt.CheckStatus{http.StatusNotFound},
		},
		&pollTest{
			Name:       "Unlogged hidden registered",
			Electorate: db.ElectorateLogged,
			Hidden:     true,
			UserType:   pollTestUserTypeUnlogged,
			Checker:    srvt.CheckStatus{http.StatusNotFound},
		},

		&pollTest{
			Name:       "Poll verified, User unverified",
			Electorate: db.ElectorateVerified,
			UserType:   pollTestUserTypeLogged,
			Checker:    srvt.CheckStatus{http.StatusNotFound},
		},
		&pollTest{
			Name:       "Poll verified, User verified",
			Electorate: db.ElectorateVerified,
			UserType:   pollTestUserTypeLogged,
			Verified:   true,
			Checker: srvt.CheckJSON{Body: &UninominalBallotAnswer{
				Previous:     undefinedVote,
				Current:      undefinedVote,
				Alternatives: alternatives,
			}},
		},
	}
	srvt.RunFunc(t, tests, UninominalBallotHandler)
}
