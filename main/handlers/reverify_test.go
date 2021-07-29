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
	"net/http"
	"testing"
	"time"

	"github.com/JBoudou/Itero/main/services"
	"github.com/JBoudou/Itero/mid/db"
	"github.com/JBoudou/Itero/mid/server"
	srvt "github.com/JBoudou/Itero/mid/server/servertest"
	"github.com/JBoudou/Itero/pkg/events"
	"github.com/JBoudou/Itero/pkg/ioc"
)

type reverifyTest_ struct {
	Name       string
	Unlogged   bool
	Verified   bool
	Previous   bool // Whether there already is a confirmation for the user.
	ExpDiff    time.Duration
	RequestFct RequestFct
	Checker    srvt.Checker
}

func ReverifyTest(c reverifyTest_) *reverifyTest {
	return &reverifyTest{
		WithName: srvt.WithName{Name: c.Name},
		WithUser: WithUser{
			Unlogged:   c.Unlogged,
			Verified:   c.Verified,
			RequestFct: c.RequestFct,
		},
		Previous: c.Previous,
		ExpDiff:  c.ExpDiff,
		Checker:  c.Checker,
	}
}

type reverifyTest struct {
	srvt.WithName
	WithUser
	WithEvent

	Previous bool
	ExpDiff  time.Duration // Difference from now for Expires (positive value are in the future)
	Checker  srvt.Checker
}

func (self *reverifyTest) Prepare(t *testing.T, loc *ioc.Locator) *ioc.Locator {
	t.Parallel()

	loc = srvt.ChainPrepare(t, loc, &self.WithUser, &self.WithEvent)

	if self.Previous {
		_, err := db.CreateConfirmation(context.Background(),
			self.User.Id, db.ConfirmationTypeVerify, self.ExpDiff)
		mustt(t, err)
	}

	if checker, ok := self.Checker.(interface{ Before(t *testing.T) }); ok {
		checker.Before(t)
	}

	return loc
}

func (self *reverifyTest) Check(t *testing.T, response *http.Response, request *server.Request) {
	var expectEvents int
	if self.Checker != nil {
		self.Checker.Check(t, response, request)
		expectEvents = 0
	} else {
		if response.StatusCode != http.StatusOK {
			t.Errorf("Wrong status code. Got %d. Expect %d.", response.StatusCode, http.StatusOK)
		}
		expectEvents = 1
	}

	gotEvents := self.CountRecorderEvents(func(evt events.Event) bool {
		converted, ok := evt.(services.ReverifyEvent)
		return ok && converted.User == self.User.Id
	})
	if gotEvents != expectEvents {
		t.Errorf("Received %d events, %d expected.", gotEvents, expectEvents)
	}
}

func TestReverifyHandler(t *testing.T) {
	t.Parallel()

	tests := []srvt.Test{

		ReverifyTest(reverifyTest_{
			Name:       "No session",
			RequestFct: RFGetNoSession,
			Checker:    srvt.CheckStatus{http.StatusForbidden},
		}),

		ReverifyTest(reverifyTest_{
			Name: "Unlogged",
			Unlogged: true,
			RequestFct: RFGetSession,
			Checker: srvt.CheckStatus{http.StatusForbidden},
		}),

		ReverifyTest(reverifyTest_{
			Name:       "No previous",
			RequestFct: RFGetSession,
		}),

		ReverifyTest(reverifyTest_{
			Name:       "Expired previous",
			Previous:   true,
			ExpDiff:    -1 * time.Second,
			RequestFct: RFGetSession,
		}),

		ReverifyTest(reverifyTest_{
			Name:       "Active previous",
			Previous:   true,
			ExpDiff:    time.Second,
			RequestFct: RFGetSession,
			Checker: srvt.CheckError{
				Code: http.StatusConflict,
				Body: "Already sent",
			},
		}),

		ReverifyTest(reverifyTest_{
			Name:       "Already verified",
			Verified:   true,
			RequestFct: RFGetSession,
			Checker: srvt.CheckError{
				Code: http.StatusBadRequest,
				Body: "Already verified",
			},
		}),
	}

	srvt.Run(t, tests, ReverifyHandler)
}
