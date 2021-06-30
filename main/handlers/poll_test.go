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
	"net/http"
	"testing"

	"github.com/JBoudou/Itero/mid/db"
	dbt "github.com/JBoudou/Itero/mid/db/dbtest"
	"github.com/JBoudou/Itero/mid/salted"
	srvt "github.com/JBoudou/Itero/mid/server/servertest"
	"github.com/JBoudou/Itero/pkg/ioc"
)

type partialPollAnswer struct {
	Title        string
	Description  string
	Admin        string
	CurrentRound uint8
	Ballot       BallotType
	Information  InformationType
}

func TestPollHandler(t *testing.T) {
	precheck(t)

	env := new(dbt.Env)
	defer env.Close()

	userId := env.CreateUser()
	env.Must(t)

	const (
		qParticipate = `INSERT INTO Participants (Poll, User, Round) VALUE (?, ?, 0)`
		qClosePoll   = `UPDATE Polls SET State = 'Terminated' WHERE Id = ?`
	)

	var (
		segment1 salted.Segment

		target1 string

		target1wrong string
	)

	createPoll := func(segment *salted.Segment, target *string, electorate db.Electorate) func(t *testing.T) {
		segment.Salt = 42
		return func(t *testing.T) {
			segment.Id = env.CreatePoll("Test", userId, electorate)
			env.Must(t)
			encoded, err := segment.Encode()
			if err != nil {
				t.Fatal(err)
			}
			*target = "/a/test/" + encoded
		}
	}

	tests := []srvt.Test{

		// Sequential tests first //
		// TODO make them all independent

		&srvt.T{
			Name:    "No segment",
			Request: srvt.Request{UserId: &userId},
			Checker: srvt.CheckStatus{http.StatusBadRequest},
		},
		&srvt.T{
			Name: "Wrong salt",
			Update: func(t *testing.T) {
				createPoll(&segment1, &target1, db.ElectorateLogged)(t)
				segment := salted.Segment{Id: segment1.Id, Salt: 9999}
				encoded, err := segment.Encode()
				if err != nil {
					t.Fatal(err)
				}
				target1wrong = "/a/test/" + encoded
			},
			Request: srvt.Request{Target: &target1wrong, UserId: &userId},
			Checker: srvt.CheckStatus{http.StatusNotFound},
		},
		&srvt.T{
			Name:    "Ok Hidden Registered",
			Request: srvt.Request{Target: &target1, UserId: &userId},
			Checker: srvt.CheckJSON{
				Body: &partialPollAnswer{
					Title:       "Test",
					Admin:       " Test ",
					Ballot:      BallotTypeUninominal,
					Information: InformationTypeNoneYet,
				},
				Partial: true,
			},
		},
		&srvt.T{
			Name: "Ok next round",
			Update: func(t *testing.T) {
				_, err := db.DB.Exec(qParticipate, segment1.Id, userId)
				if err != nil {
					t.Fatal(err)
				}
				env.NextRound(segment1.Id)
				env.Must(t)
			},
			Request: srvt.Request{Target: &target1, UserId: &userId},
			Checker: srvt.CheckJSON{
				Body: &partialPollAnswer{
					Title:        "Test",
					Admin:        " Test ",
					CurrentRound: 1,
					Ballot:       BallotTypeUninominal,
					Information:  InformationTypeCounts,
				},
				Partial: true,
			},
		},
		&srvt.T{
			Name: "Ok closed",
			Update: func(t *testing.T) {
				_, err := db.DB.Exec(qClosePoll, segment1.Id)
				if err != nil {
					t.Fatal(err)
				}
			},
			Request: srvt.Request{Target: &target1, UserId: &userId},
			Checker: srvt.CheckJSON{
				Body: &partialPollAnswer{
					Title:        "Test",
					Admin:        " Test ",
					CurrentRound: 1,
					Ballot:       BallotTypeClosed,
					Information:  InformationTypeCounts,
				},
				Partial: true,
			},
		},

		// Independent tests //

		&missingPollTest{},

		&pollTest{
			Name:       "No user access to public",
			Electorate: db.ElectorateAll,
			UserType:   pollTestUserTypeNone,
			Checker:    pollHandlerCheckerFactory,
		},
		&pollTest{
			Name:       "No user access to public hidden",
			Electorate: db.ElectorateAll,
			Hidden:     true,
			UserType:   pollTestUserTypeNone,
			Checker:    pollHandlerCheckerFactory,
		},
		&pollTest{
			Name:       "Unlogged access to public",
			Electorate: db.ElectorateAll,
			UserType:   pollTestUserTypeUnlogged,
			Checker:    pollHandlerCheckerFactory,
		},
		&pollTest{
			Name:       "Unlogged access to public hidden",
			Electorate: db.ElectorateAll,
			Hidden:     true,
			UserType:   pollTestUserTypeUnlogged,
			Checker:    pollHandlerCheckerFactory,
		},

		&pollTest{
			Name:       "No user no access to registered",
			Electorate: db.ElectorateLogged,
			UserType:   pollTestUserTypeNone,
			Checker:    srvt.CheckStatus{http.StatusNotFound},
		},
		&pollTest{
			Name:       "No user no access to hidden registered",
			Electorate: db.ElectorateLogged,
			Hidden:     true,
			UserType:   pollTestUserTypeNone,
			Checker:    srvt.CheckStatus{http.StatusNotFound},
		},
		&pollTest{
			Name:       "Unlogged no access to registered",
			Electorate: db.ElectorateLogged,
			UserType:   pollTestUserTypeUnlogged,
			Checker:    srvt.CheckStatus{http.StatusNotFound},
		},
		&pollTest{
			Name:       "Unlogged no access to hidden registered",
			Electorate: db.ElectorateLogged,
			Hidden:     true,
			UserType:   pollTestUserTypeUnlogged,
			Checker:    srvt.CheckStatus{http.StatusNotFound},
		},

		&pollTest{
			Name:       "No user access any round",
			Electorate: db.ElectorateAll,
			Round:      2,
			UserType:   pollTestUserTypeNone,
			Checker:    pollHandlerCheckerFactory,
		},
		&pollTest{
			Name:       "Logged access any round",
			Electorate: db.ElectorateAll,
			Round:      2,
			UserType:   pollTestUserTypeLogged,
			Checker:    pollHandlerCheckerFactory,
		},
		&pollTest{
			Name:       "Logged no access any round",
			Electorate: db.ElectorateLogged,
			Round:      2,
			UserType:   pollTestUserTypeLogged,
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
			Checker:    pollHandlerCheckerFactory,
		},
	}
	srvt.RunFunc(t, tests, PollHandler)
}

func pollHandlerCheckerFactory(param PollTestCheckerFactoryParam) srvt.Checker {
	body := &partialPollAnswer{
		Title:        param.PollTitle,
		Admin:        param.AdminName,
		CurrentRound: param.Round,
		Ballot:       BallotTypeUninominal,
		Information:  InformationTypeCounts,
	}
	if param.Round == 0 {
		body.Information = InformationTypeNoneYet
	}
	return srvt.CheckJSON{
		Body:    body,
		Partial: true,
	}
}

//
// missingPollTest
//

type missingPollTest struct {
	WithUser         // If WithUser.RequestFct is nil, uses RFGetSession.
	srvt.WithChecker // If WithChecker.Checker is nil, check StatusNotFound.

	pollSegment salted.Segment
}

func (self missingPollTest) GetName() string {
	return "Missing poll"
}

func (self *missingPollTest) Prepare(t *testing.T) *ioc.Locator {
	if self.WithUser.RequestFct == nil {
		self.WithUser.RequestFct = RFGetSession
	}
	self.WithUser.Prepare(t)

	var pollEnv dbt.Env
	self.pollSegment.Id = pollEnv.CreatePoll("Todel", self.User.Id, db.ElectorateAll)
	self.pollSegment.Salt = dbt.PollSalt
	pollEnv.Must(t)
	pollEnv.Close()

	if self.Checker == nil {
		self.Checker = srvt.CheckStatus{http.StatusNotFound}
	}
	return self.WithChecker.Prepare(t)
}

func (self missingPollTest) GetRequest(t *testing.T) *srvt.Request {
	req := self.WithUser.GetRequest(t)
	segment, err := self.pollSegment.Encode()
	mustt(t, err)
	segment = "/a/test/" + segment
	req.Target = &segment
	return req
}
