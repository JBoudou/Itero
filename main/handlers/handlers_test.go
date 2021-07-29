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
	"testing"

	"github.com/JBoudou/Itero/mid/db"
	dbt "github.com/JBoudou/Itero/mid/db/dbtest"
	"github.com/JBoudou/Itero/mid/root"
	"github.com/JBoudou/Itero/mid/salted"
	"github.com/JBoudou/Itero/mid/server"
	srvt "github.com/JBoudou/Itero/mid/server/servertest"
	"github.com/JBoudou/Itero/pkg/events"
	"github.com/JBoudou/Itero/pkg/events/eventstest"
	"github.com/JBoudou/Itero/pkg/ioc"
)

func mustt(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func precheck(t *testing.T) {
	if !(root.Configured && db.Ok && server.Ok) {
		t.Log("Impossible to test the main package: some dependent packages are not ok.")
		t.Log("Check that there is a configuration file in main/. (or a link to the main configuation file).")
		t.SkipNow()
	}
}

func makePollRequest(t *testing.T, pollId uint32, userId *uint32) *srvt.Request {
	pollSegment := salted.Segment{Salt: dbt.PollSalt, Id: pollId}
	encoded, err := pollSegment.Encode()
	mustt(t, err)
	target := "/a/test/" + encoded
	return &srvt.Request{Target: &target, UserId: userId}
}

//
// WithUser //
//

var withUserFakeAddress string = "1.2.3.4"

type RequestFct = func(user *server.User) *srvt.Request

type WithUser struct {
	dbt.WithDB

	Unlogged   bool
	Verified   bool
	RequestFct RequestFct

	User server.User
}

func (self *WithUser) Prepare(t *testing.T, loc *ioc.Locator) *ioc.Locator {
	if self.Unlogged {
		var err error
		self.User, err = UnloggedFromAddr(context.Background(), withUserFakeAddress)
		mustt(t, err)

	} else {
		self.User = server.User{
			Id:     self.DB.CreateUserWith(t.Name()),
			Name:   dbt.UserNameWith(t.Name()),
			Logged: true,
		}
		self.DB.Must(t)

		const qVerified = `UPDATE Users SET Verified = TRUE WHERE Id = ?`
		if self.Verified {
			_, err := db.DB.Exec(qVerified, self.User.Id)
			mustt(t, err)
		}
	}

	return loc
}

func (self *WithUser) GetRequest(t *testing.T) *srvt.Request {
	return self.RequestFct(&self.User)
}

func RFGetNoSession(user *server.User) *srvt.Request {
	return &srvt.Request{}
}

func RFPostNoSession(body string) RequestFct {
	return func(user *server.User) *srvt.Request {
		return &srvt.Request{
			Method: "POST",
			Body:   body,
		}
	}
}

func rfSession(method string, body string, user *server.User) (req *srvt.Request) {
	req = &srvt.Request{
		Method:     method,
		RemoteAddr: &withUserFakeAddress,
		Body:       body,
		UserId:     &user.Id,
	}
	if !user.Logged {
		req.Hash = &user.Hash
	}
	return
}

func RFGetSession(user *server.User) *srvt.Request {
	return rfSession("GET", "", user)
}

func RFPostSession(body string) RequestFct {
	return func(user *server.User) *srvt.Request {
		return rfSession("POST", body, user)
	}
}

//
// WithEvent //
//

type WithEvent struct {
	RecordedEvents []events.Event
}

func (self *WithEvent) Prepare(t *testing.T, loc *ioc.Locator) *ioc.Locator {
	manager := &eventstest.ManagerMock{
		T: t,
		Send_: func(evt events.Event) error {
			self.RecordedEvents = append(self.RecordedEvents, evt)
			return nil
		},
	}

	loc = loc.Sub()
	mustt(t, loc.Bind(func() events.Manager { return manager }))
	return loc
}

func (self *WithEvent) CountRecorderEvents(predicate func(events.Event) bool) (ret int) {
	for _, recorded := range self.RecordedEvents {
		if predicate(recorded) {
			ret += 1
		}
	}
	return
}
