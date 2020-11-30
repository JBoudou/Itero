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
			segment: PollSegment{Id: 0xF1234567, Salt: 0x312345},
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

func TestPollHandler(t *testing.T) {
	precheck(t)

	env := new(dbt.Env)
	defer env.Close()

	userId := env.CreateUser()
	if env.Error != nil {
		t.Fatalf("Env failed: %s", env.Error)
	}

	const (
		qParticipate = `INSERT INTO Participants (Poll, User) VALUE (?, ?)`
		qNextRound =
			`UPDATE Polls SET CurrentRound = CurrentRound + 1 AND CurrentRoundStart = Now()
			 WHERE Id = ?`
		qClosePoll = `UPDATE Polls SET Active = false WHERE Id = ?`
	)

	var (
		segment1 PollSegment
		segment2 PollSegment
		segment3 PollSegment

		target1 string
		target2 string
		target3 string

		target1wrong string
	)

	createPoll := func(segment *PollSegment, target *string, publicity uint8) func(t *testing.T) {
		segment.Salt = 42
		return func(t *testing.T) {
			segment.Id = env.CreatePoll("Test", userId, publicity)
			if env.Error != nil {
				t.Fatal(env.Error)
			}
			encoded, err := segment.Encode()
			if err != nil {
				t.Fatal(err)
			}
			*target = "/a/test/" + encoded
		}
	}

	tests := []srvt.Test{
		{
			Name: "No segment",
			Request: srvt.Request{ UserId: &userId },
			Checker: srvt.CheckerStatus(http.StatusBadRequest),
		},
		{
			Name: "No session",
			Update:  createPoll(&segment1, &target1, db.PollPublicityHiddenRegistered),
			Request: srvt.Request{ Target: &target1 },
			Checker: srvt.CheckerStatus(http.StatusForbidden),
		},
		{
			Name: "Wrong salt",
			Update: func(t *testing.T) {
				segment := PollSegment{ Id: segment1.Id, Salt: 9999 }
				encoded, err := segment.Encode()
				if err != nil {
					t.Fatal(err)
				}
				target1wrong = "/a/test/" + encoded
			},
			Request: srvt.Request{ Target: &target1wrong, UserId: &userId },
			Checker: srvt.CheckerStatus(http.StatusForbidden),
		},
		{
			Name: "Private poll",
			Update:  createPoll(&segment2, &target2, db.PollPublicityInvited),
			Request: srvt.Request{ Target: &target2, UserId: &userId },
			Checker: srvt.CheckerStatus(http.StatusForbidden),
		},
		{
			Name: "Late public poll",
			Update:  func(t *testing.T) {
				createPoll(&segment3, &target3, db.PollPublicityPublic)(t)
				_, err := db.DB.Exec(qNextRound, segment3.Id)
				if err != nil {
					t.Fatal(err)
				}
			},
			Request: srvt.Request{ Target: &target3, UserId: &userId },
			Checker: srvt.CheckerError(http.StatusForbidden, "Too late"),
		},
		{
			Name: "Ok Hidden Registered",
			Request: srvt.Request{ Target: &target1, UserId: &userId },
			Checker: srvt.CheckerJSON(http.StatusOK,
				PollAnswer{Ballot: BallotTypeUninomial, Information: InformationTypeNoneYet}),
		},
		{
			Name: "Ok Invited",
			Update: func(t *testing.T) {
				_, err := db.DB.Exec(qParticipate, segment2.Id, userId)
				if err != nil {
					t.Fatal(err)
				}
			},
			Request: srvt.Request{ Target: &target2, UserId: &userId },
			Checker: srvt.CheckerJSON(http.StatusOK,
				PollAnswer{Ballot: BallotTypeUninomial, Information: InformationTypeNoneYet}),
		},
		{
			Name: "Ok next round",
			Update: func(t *testing.T) {
				_, err := db.DB.Exec(qNextRound, segment1.Id)
				if err != nil {
					t.Fatal(err)
				}
			},
			Request: srvt.Request{ Target: &target1, UserId: &userId },
			Checker: srvt.CheckerJSON(http.StatusOK,
				PollAnswer{Ballot: BallotTypeUninomial, Information: InformationTypeCounts}),
		},
		{
			Name: "Ok closed",
			Update: func(t *testing.T) {
				_, err := db.DB.Exec(qClosePoll, segment1.Id)
				if err != nil {
					t.Fatal(err)
				}
			},
			Request: srvt.Request{ Target: &target1, UserId: &userId },
			Checker: srvt.CheckerJSON(http.StatusOK,
				PollAnswer{Ballot: BallotTypeClosed, Information: InformationTypeCounts}),
		},
	}
	srvt.RunFunc(t, tests, PollHandler)

	tests = []srvt.Test{
	}
	srvt.RunFunc(t, tests, PollHandler)
}
