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
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"testing"

	"github.com/JBoudou/Itero/mid/db"
	dbt "github.com/JBoudou/Itero/mid/db/dbtest"
	"github.com/JBoudou/Itero/mid/salted"
	"github.com/JBoudou/Itero/mid/server"
	srvt "github.com/JBoudou/Itero/mid/server/servertest"
	"github.com/JBoudou/Itero/pkg/ioc"
)

type partialPollAnswer struct {
	Title        string
	Description  string
	Admin        string
	CurrentRound uint8
	Ballot       uint8
	Information  uint8
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

// What follows is generic and used for other handler.
// TODO move that in its own place.

type pollTestUserType uint8

const (
	pollTestUserTypeNone pollTestUserType = iota
	pollTestUserTypeAdmin
	pollTestUserTypeLogged
	pollTestUserTypeUnlogged
)

type pollTestParticipate struct {
	User  uint8 // 0 is admin, 1 is request user, n > 1 are additional users.
	Round uint8
}

type pollTestVote struct {
	User  uint8 // 0 is admin, 1 is request user, n > 1 are additional users.
	Round uint8
	Alt   uint8
}

type pollTest struct {
	Name         string        // Required.
	Sequential   bool          // If the test cannot be run in parallel.
	Electorate   db.Electorate // Required.
	Hidden       bool
	Alternatives []string
	Round        uint8
	Participate  []pollTestParticipate // No need to add an entry for each vote.
	Vote         []pollTestVote

	UserType pollTestUserType // Required.
	Verified bool             // Whether the user doing the request is verified.
	Request  srvt.Request     // Just a squeleton that will be completed by the test.
	Checker  interface{}      // Required. Either an srvt.Checker or a PollTestCheckerFactoryParam.

	dbEnv  dbt.Env
	pollId uint32
	userId []uint32
}

// PollTestCheckerFactoryParam contains the parameter to construct a default Checker.
// We use a struct type instead of a list of arguments to ease the addition of parameters.
// TODO consider moving that to a _test package.
type PollTestCheckerFactoryParam struct {
	PollTitle string
	PollId    uint32
	AdminName string
	UserId    uint32
	Round     uint8
}

type pollTestCheckerFactory = func(param PollTestCheckerFactoryParam) srvt.Checker

var pollHanlerTestUnloggedHash uint32 = 42

func (self *pollTest) GetName() string {
	return self.Name
}

func stats(ctx string) {
	stats := db.DB.Stats()
	fmt.Printf("%s -- open %d, use %d, idle %d, goroutines %d\n", ctx,
		stats.OpenConnections, stats.InUse, stats.Idle, runtime.NumGoroutine())
}

func (self *pollTest) Prepare(t *testing.T) *ioc.Locator {
	stats("Before Prepare")

	if !self.Sequential {
		t.Parallel()
	}

	self.userId = make([]uint32, 2)
	self.userId[0] = self.dbEnv.CreateUserWith(t.Name())
	if len(self.Alternatives) < 2 {
		self.pollId = self.dbEnv.CreatePoll("Title", self.userId[0], self.Electorate)
	} else {
		self.pollId = self.dbEnv.CreatePollWith("Title", self.userId[0], self.Electorate,
			self.Alternatives)
	}

	// Hidden
	const qHidden = `UPDATE Polls SET Hidden = TRUE WHERE Id = ?`
	if self.Hidden {
		self.dbEnv.Must(t)
		_, err := db.DB.Exec(qHidden, self.pollId)
		mustt(t, err)
	}

	// Users
	switch self.UserType {
	case pollTestUserTypeAdmin:
		self.userId[1] = self.userId[0]

	case pollTestUserTypeLogged:
		self.userId[1] = self.dbEnv.CreateUserWith(t.Name() + "Logged")

	case pollTestUserTypeUnlogged:
		self.dbEnv.Must(t)
		user, err := UnloggedFromHash(context.Background(), 42)
		mustt(t, err)
		self.userId[1] = user.Id

	case pollTestUserTypeNone:
		if self.Request.RemoteAddr != nil {
			self.dbEnv.Must(t)
			user, err := UnloggedFromAddr(context.Background(), *self.Request.RemoteAddr)
			mustt(t, err)
			self.userId[1] = user.Id
			break
		}
		fallthrough // There must be a break at the end of the previous if

	default:
		self.userId[1] = self.userId[0]
	}

	// Verified
	const qVerified = `UPDATE Users SET Verified = TRUE WHERE Id = ?`
	if self.Verified {
		self.dbEnv.QuietExec(qVerified, self.userId[1])
	}

	// Round
	const qRound = `UPDATE Polls SET CurrentRound = ? WHERE Id = ?`
	if self.Round > 0 {
		self.dbEnv.QuietExec(qRound, self.Round, self.pollId)
	}

	// Participate
	const qParticipate = `INSERT INTO Participants (User, Poll, Round) VALUE (?,?,?)`
	for _, participate := range self.Participate {
		for len(self.userId) <= int(participate.User) {
			self.userId = append(self.userId, self.dbEnv.CreateUserWith(t.Name()+strconv.Itoa(len(self.userId))))
		}
		self.dbEnv.QuietExec(qParticipate, self.userId[participate.User], self.pollId, participate.Round)
	}

	// Vote
	const qVote = `INSERT INTO Ballots (User, Poll, Alternative, Round) VALUE (?,?,?,?)`
	for _, vote := range self.Vote {
		for len(self.userId) <= int(vote.User) {
			self.userId = append(self.userId, self.dbEnv.CreateUserWith(t.Name()+strconv.Itoa(len(self.userId))))
		}
		// We don't care about errors when adding to Participants
		db.DB.Exec(qParticipate, self.userId[vote.User], self.pollId, vote.Round)
		self.dbEnv.QuietExec(qVote, self.userId[vote.User], self.pollId, vote.Alt, vote.Round)
	}

	// Checker
	if factory, ok := self.Checker.(pollTestCheckerFactory); ok {
		self.Checker = factory(PollTestCheckerFactoryParam{
			PollTitle: "Title",
			PollId:    self.pollId,
			AdminName: self.dbEnv.UserNameWith(t.Name()),
			UserId:    self.userId[1],
			Round:     self.Round,
		})
	}

	self.dbEnv.Must(t)
	if checker, ok := self.Checker.(interface{ Before(*testing.T) }); ok {
		checker.Before(t)
	}
	stats("After Prepare")
	return ioc.Root
}

func (self *pollTest) GetRequest(t *testing.T) *srvt.Request {
	segment, err := salted.Segment{Id: self.pollId, Salt: 42}.Encode()
	if err != nil {
		t.Errorf("Error encoding poll segment: %v.", err)
	}
	target := "/a/test/" + segment
	self.Request.Target = &target

	switch self.UserType {
	case pollTestUserTypeAdmin:
		self.Request.UserId = &self.userId[0]

	case pollTestUserTypeLogged:
		self.Request.UserId = &self.userId[1]

	case pollTestUserTypeUnlogged:
		self.Request.UserId = &self.userId[1]
		self.Request.Hash = &pollHanlerTestUnloggedHash
	}

	return &self.Request
}

func (self *pollTest) Check(t *testing.T, response *http.Response, request *server.Request) {
	if checker, ok := self.Checker.(srvt.Checker); ok {
		checker.Check(t, response, request)
	} else {
		t.Errorf("Checker is not an srvt.Checker")
	}
	stats("After Check")
}

func (self *pollTest) Close() {
	self.dbEnv.Close()
}
