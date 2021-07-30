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
	"github.com/JBoudou/Itero/mid/unlogged"
	"github.com/JBoudou/Itero/pkg/events"
	"github.com/JBoudou/Itero/pkg/ioc"
)

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
	dbt.WithDB
	WithEvent

	Name         string        // Required.
	Electorate   db.Electorate // Required.
	Hidden       bool
	Alternatives []string
	Round        uint8
	Waiting      bool
	Participate  []pollTestParticipate // No need to add an entry for each vote.
	Vote         []pollTestVote

	UserType pollTestUserType // Required.
	Verified bool             // Whether the user doing the request is verified.

	Request        srvt.Request // Just a squeleton that will be completed by the test.
	Checker        interface{}  // Required. Either an srvt.Checker or a PollTestCheckerFactory.
	EventPredicate func(PollTestCheckerFactoryParam, events.Event) bool
	EventCount     int

	pollId uint32
	userId []uint32
}

// PollTestCheckerFactoryParam contains the parameter to construct a default Checker.
// We use a struct type instead of a list of arguments to ease the addition of parameters.
type PollTestCheckerFactoryParam struct {
	PollTitle string
	PollId    uint32
	AdminName string
	UserId    uint32
	Round     uint8
}

type pollTestCheckerFactory = func(param PollTestCheckerFactoryParam) srvt.Checker

var pollHandlerTestUnloggedHash uint32 = 42

func (self *pollTest) GetName() string {
	return self.Name
}

func stats(ctx string) {
	stats := db.DB.Stats()
	fmt.Printf("%s -- open %d, use %d, idle %d, goroutines %d\n", ctx,
		stats.OpenConnections, stats.InUse, stats.Idle, runtime.NumGoroutine())
}

func (self *pollTest) Prepare(t *testing.T, loc *ioc.Locator) *ioc.Locator {
	t.Parallel()

	self.userId = make([]uint32, 2)
	self.userId[0] = self.DB.CreateUserWith(t.Name())
	if len(self.Alternatives) < 2 {
		self.pollId = self.DB.CreatePoll("Title", self.userId[0], self.Electorate)
	} else {
		self.pollId = self.DB.CreatePollWith("Title", self.userId[0], self.Electorate,
			self.Alternatives)
	}

	// Hidden
	const qHidden = `UPDATE Polls SET Hidden = TRUE WHERE Id = ?`
	if self.Hidden {
		self.DB.Must(t)
		_, err := db.DB.Exec(qHidden, self.pollId)
		mustt(t, err)
	}

	// Users
	switch self.UserType {
	case pollTestUserTypeAdmin:
		self.userId[1] = self.userId[0]

	case pollTestUserTypeLogged:
		self.userId[1] = self.DB.CreateUserWith(t.Name() + "Logged")

	case pollTestUserTypeUnlogged:
		self.DB.Must(t)
		user, err := unlogged.FromHash(context.Background(), 42)
		mustt(t, err)
		self.userId[1] = user.Id

	case pollTestUserTypeNone:
		if self.Request.RemoteAddr != nil {
			self.DB.Must(t)
			user, err := unlogged.FromAddr(context.Background(), *self.Request.RemoteAddr)
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
		self.DB.QuietExec(qVerified, self.userId[1])
	}

	// Round
	const qRound = `UPDATE Polls SET CurrentRound = ? WHERE Id = ?`
	if self.Round > 0 {
		self.DB.QuietExec(qRound, self.Round, self.pollId)
	}

	// Waiting
	const qWaiting = `
	  UPDATE Polls
		   SET State = 'Waiting', Start = ADDTIME(CURRENT_TIMESTAMP, '1:00:00')
	   WHERE Id = ?`
	if self.Waiting {
		self.DB.QuietExec(qWaiting, self.pollId)
	}

	// Participate
	const qParticipate = `INSERT INTO Participants (User, Poll, Round) VALUE (?,?,?)`
	for _, participate := range self.Participate {
		for len(self.userId) <= int(participate.User) {
			self.userId = append(self.userId, self.DB.CreateUserWith(t.Name()+strconv.Itoa(len(self.userId))))
		}
		self.DB.QuietExec(qParticipate, self.userId[participate.User], self.pollId, participate.Round)
	}

	// Vote
	const qVote = `INSERT INTO Ballots (User, Poll, Alternative, Round) VALUE (?,?,?,?)`
	for _, vote := range self.Vote {
		for len(self.userId) <= int(vote.User) {
			self.userId = append(self.userId, self.DB.CreateUserWith(t.Name()+strconv.Itoa(len(self.userId))))
		}
		// We don't care about errors when adding to Participants
		db.DB.Exec(qParticipate, self.userId[vote.User], self.pollId, vote.Round)
		self.DB.QuietExec(qVote, self.userId[vote.User], self.pollId, vote.Alt, vote.Round)
	}

	// Checker
	if factory, ok := self.Checker.(pollTestCheckerFactory); ok {
		self.Checker = factory(self.makeParam(t))
	}

	self.DB.Must(t)
	if checker, ok := self.Checker.(interface{ Before(*testing.T) }); ok {
		checker.Before(t)
	}
	return self.WithEvent.Prepare(t, loc)
}

func (self *pollTest) GetRequest(t *testing.T) *srvt.Request {
	segment, err := salted.Segment{Id: self.pollId, Salt: dbt.PollSalt}.Encode()
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
		self.Request.Hash = &pollHandlerTestUnloggedHash
	}

	return &self.Request
}

func (self *pollTest) Check(t *testing.T, response *http.Response, request *server.Request) {
	if checker, ok := self.Checker.(srvt.Checker); ok {
		checker.Check(t, response, request)
	} else {
		t.Errorf("Checker is not an srvt.Checker")
	}

	if self.EventPredicate != nil {
		param := self.makeParam(t)
		gotEvents := self.WithEvent.CountRecorderEvents(func(evt events.Event) bool {
			return self.EventPredicate(param, evt)
		})
		if gotEvents != self.EventCount {
			t.Errorf("Got %d events. Expect %d.", gotEvents, self.EventCount)
		}
	}
}

func (self *pollTest) makeParam(t *testing.T) PollTestCheckerFactoryParam {
	return PollTestCheckerFactoryParam{
		PollTitle: "Title",
		PollId:    self.pollId,
		AdminName: dbt.UserNameWith(t.Name()),
		UserId:    self.userId[1],
		Round:     self.Round,
	}
}

//
// No poll test
//

type wrongPollTest struct {
	dbt.WithDB
	WithEvent

	Kind wrongPollTestKind

	uid    uint32
	pollId uint32
}

type wrongPollTestKind uint8

const (
	wrongPollTestKindNoPoll wrongPollTestKind = iota
	wrongPollTestKindWrongSalt
)

func (self wrongPollTest) GetName() string {
	switch self.Kind {
	case wrongPollTestKindNoPoll:
		return "No poll"
	case wrongPollTestKindWrongSalt:
		return "Wrong salt"
	}
	return "Invalid test"
}

func (self *wrongPollTest) Prepare(t *testing.T, loc *ioc.Locator) *ioc.Locator {
	t.Parallel()

	self.uid = self.DB.CreateUserWith(t.Name())
	self.pollId = self.DB.CreatePoll("Test", self.uid, db.ElectorateAll)

	const qDelete = `DELETE FROM Polls WHERE Id = ?`
	if self.Kind == wrongPollTestKindNoPoll {
		self.DB.QuietExec(qDelete, self.pollId)
	}

	self.DB.Must(t)
	return self.WithEvent.Prepare(t, loc)
}

func (self wrongPollTest) GetRequest(t *testing.T) *srvt.Request {
	var salt uint32 = dbt.PollSalt
	if self.Kind == wrongPollTestKindWrongSalt {
		salt += 1
	}

	segment, err := salted.Segment{Id: self.pollId, Salt: salt}.Encode()
	if err != nil {
		t.Errorf("Error encoding poll segment: %v.", err)
	}
	target := "/a/test/" + segment

	return &srvt.Request{
		Target: &target,
		UserId: &self.uid,
	}
}

func (self wrongPollTest) Check(t *testing.T, response *http.Response, request *server.Request) {
	if response.StatusCode != http.StatusNotFound {
		t.Errorf("Wrong status. Got %d. Expect %d.", response.StatusCode, http.StatusNotFound)
	}
	if len(self.RecordedEvents) != 0 {
		t.Errorf("Events sent: %v", self.RecordedEvents)
	}
}
