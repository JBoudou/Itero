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
	dbt "github.com/JBoudou/Itero/mid/db/dbtest"
	"github.com/JBoudou/Itero/mid/server"
	srvt "github.com/JBoudou/Itero/mid/server/servertest"
	"github.com/JBoudou/Itero/pkg/events"
	"github.com/JBoudou/Itero/pkg/ioc"
)

type forgotTest_ struct {
	Name     string
	Previous bool                    // Whether there already is a confirmation for the user.
	ExpDiff  time.Duration           // Difference from now for Expires (positive value are in the future)
	UserFct  func(*testing.T) string // Produces the User field of the query
	Checker  srvt.Checker            // nil to check for success.
}

func ForgotTest(c forgotTest_) *forgotTest {
	return &forgotTest{
		WithName: srvt.WithName{Name: c.Name},
		Previous: c.Previous,
		ExpDiff: c.ExpDiff,
		UserFct: c.UserFct,
		Checker: c.Checker,
	}
}

type forgotTest struct {
	srvt.WithName
	dbt.WithDB
	WithEvent

	Previous bool
	ExpDiff  time.Duration
	UserFct  func(*testing.T) string
	Checker  srvt.Checker

	uid uint32
}

func (self *forgotTest) Prepare(t *testing.T) *ioc.Locator {
	t.Parallel()

	self.uid = self.DB.CreateUserWith(t.Name())
	self.DB.Must(t)

	if self.Previous {
		_, err := db.CreateConfirmation(context.Background(),
			self.uid, db.ConfirmationTypePasswd, self.ExpDiff)
		mustt(t, err)
	}

	return self.WithEvent.Prepare(t)
}

func (self forgotTest) GetRequest(t *testing.T) *srvt.Request {
	return &srvt.Request{
		Method: "POST",
		Body: `{"User":"` + self.UserFct(t) + `"}`,
	}
}

func (self forgotTest) Check(t *testing.T, response *http.Response, request *server.Request) {
	if self.Checker != nil {
		self.Checker.Check(t, response, request)
		return
	}

	if response.StatusCode != http.StatusOK {
		t.Errorf("Wrong status code. Got %d. Expect %d.", response.StatusCode, http.StatusOK)
	}
	countEvents := self.CountRecorderEvents(func(evt events.Event) bool {
		converted, ok := evt.(services.ForgotEvent)
		return ok && converted.User == self.uid
	})
	if countEvents != 1 {
		t.Errorf("Wrong number of events sent. Got %d. Expect %d.", countEvents, 1)
	}
}

func TestForgotHandler(t *testing.T) {
	tests := []srvt.Test{
		ForgotTest(forgotTest_{
			Name: "Login",
			UserFct: func(t *testing.T) string {return dbt.UserNameWith(t.Name())},
		}),
		ForgotTest(forgotTest_{
			Name: "Email",
			UserFct: func(t *testing.T) string {return dbt.UserEmailWith(t.Name())},
		}),
		ForgotTest(forgotTest_{
			Name: "Unknown user",
			UserFct: func(t *testing.T) string {return dbt.ImpossibleUserName},
			Checker: srvt.CheckError{Code: http.StatusForbidden, Body: server.UnauthorizedHttpErrorMsg},
		}),
		ForgotTest(forgotTest_{
			Name: "Valid confirmation",
			UserFct: func(t *testing.T) string {return dbt.UserNameWith(t.Name())},
			Previous: true,
			ExpDiff: time.Minute,
			Checker: srvt.CheckError{Code: http.StatusConflict, Body: "Already sent"},
		}),
		ForgotTest(forgotTest_{
			Name: "Invalid confirmation",
			UserFct: func(t *testing.T) string {return dbt.UserNameWith(t.Name())},
			Previous: true,
			ExpDiff: -1 * time.Minute,
		}),
	}
	srvt.Run(t, tests, ForgotHandler)
}
