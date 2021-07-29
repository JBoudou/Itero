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
	"strings"
	"testing"

	"github.com/JBoudou/Itero/main/services"
	"github.com/JBoudou/Itero/mid/db"
	dbt "github.com/JBoudou/Itero/mid/db/dbtest"
	"github.com/JBoudou/Itero/mid/server"
	srvt "github.com/JBoudou/Itero/mid/server/servertest"
	"github.com/JBoudou/Itero/pkg/events"
	"github.com/JBoudou/Itero/pkg/ioc"
)

// signupHandlerTest

type signupHandlerTest struct {
	srvt.T
	WithEvent

	uid    uint32
	called int
}

func (self *signupHandlerTest) Prepare(t *testing.T, loc *ioc.Locator) *ioc.Locator {
	t.Parallel()
	return srvt.ChainPrepare(t, loc, &self.T, &self.WithEvent)
}

func (self *signupHandlerTest) ChangeResponse(t *testing.T, in server.Response) server.Response {
	return srvt.ResponseSpy{
		T:       t,
		Backend: in,
		LoginFct: func(_ *testing.T, _ context.Context, user server.User, _ *server.Request, _ interface{}) {
			self.uid = user.Id
			self.called += 1
		},
	}
}

func (self *signupHandlerTest) Check(t *testing.T, response *http.Response, request *server.Request) {
	expect := 1
	if self.Checker != nil {
		self.Checker.Check(t, response, request)
		expect = 0
	} else {
		if response.StatusCode != http.StatusOK {
			t.Errorf("Wrong status. Got %d. Expect %d.", response.StatusCode, http.StatusOK)
		}
	}

	if self.called != expect {
		t.Errorf("response.SendLoginAccepted called %d times. Expect %d.", self.called, expect)
	}
	got := self.CountRecorderEvents(func(evt events.Event) bool {
		_, ok := evt.(services.CreateUserEvent)
		return ok
	})
	if got != expect {
		t.Errorf("%d CreateUserEvent sent. Expect %d.", got, expect)
	}
}

func (self *signupHandlerTest) Close() {
	if self.called > 0 {
		const qDelete = `DELETE FROM Users WHERE Id = ?`
		db.DB.Exec(qDelete, self.uid)
	}
	self.T.Close()
}

// SignupHandlerTest with user

type signupHandlerTestWithUser struct {
	signupHandlerTest
	dbt.WithDB

	RequestFct func(t *testing.T, basename string) *srvt.Request
}

func (self *signupHandlerTestWithUser) Prepare(t *testing.T, loc *ioc.Locator) *ioc.Locator {
	uid := self.DB.CreateUserWith(t.Name())

	// We remove the spaces for error "Name has space" to not be raised
	const qNoSpace = `UPDATE Users SET Name = TRIM(Name) WHERE Id = ?`
	self.DB.QuietExec(qNoSpace, uid)
	self.DB.Must(t)

	return self.signupHandlerTest.Prepare(t, loc)
}

func (self *signupHandlerTestWithUser) GetRequest(t *testing.T) *srvt.Request {
	return self.RequestFct(t, t.Name())
}

func (self *signupHandlerTestWithUser) Close() {
	self.signupHandlerTest.Close()
	self.WithDB.Close()
}

type signupHandlerTestWithUser_ struct {
	Name       string
	RequestFct func(t *testing.T, basename string) *srvt.Request
	Checker    srvt.Checker
}

func SignupHandlerTest(c signupHandlerTestWithUser_) *signupHandlerTestWithUser {
	return &signupHandlerTestWithUser{
		signupHandlerTest: signupHandlerTest{
			T: srvt.T{
				Name:    c.Name,
				Checker: c.Checker,
			},
		},
		RequestFct: c.RequestFct,
	}
}


// Tests

func TestSignupHandler(t *testing.T) {
	precheck(t)
	t.Parallel()

	// We have no choice but to try to find an unlikely name
	const name = "toto_my_test_user_with_a_long_name"
	const body = `{"Name":"` + name + `","Email":"` + name + `@example.com","Passwd":"tititi"}`

	tests := []srvt.Test{
		&signupHandlerTest{T: srvt.T{
			Name:    "Bad request",
			Request: srvt.Request{Method: "POST"},
			Checker: srvt.CheckStatus{http.StatusBadRequest},
		}},
		&signupHandlerTest{T: srvt.T{
			Name: "Name too short",
			Request: srvt.Request{
				Method: "POST",
				Body:   `{"Name":"a","Email":"toto@example.com","Passwd":"tititi"}`,
			},
			Checker: srvt.CheckError{http.StatusBadRequest, "Name too short"},
		}},
		&signupHandlerTest{T: srvt.T{
			Name: "Name starting with a space",
			Request: srvt.Request{
				Method: "POST",
				Body:   `{"Name":" tototo","Email":"toto@example.com","Passwd":"tititi"}`,
			},
			Checker: srvt.CheckError{http.StatusBadRequest, "Name has spaces"},
		}},
		&signupHandlerTest{T: srvt.T{
			Name: "Name ending with a space",
			Request: srvt.Request{
				Method: "POST",
				Body:   `{"Name":"tototo ","Email":"toto@example.com","Passwd":"tititi"}`,
			},
			Checker: srvt.CheckError{http.StatusBadRequest, "Name has spaces"},
		}},
		&signupHandlerTest{T: srvt.T{
			Name: "Name containing @",
			Request: srvt.Request{
				Method: "POST",
				Body:   `{"Name":"toto@to","Email":"toto@example.com","Passwd":"tititi"}`,
			},
			Checker: srvt.CheckError{http.StatusBadRequest, "Name has at sign"},
		}},
		&signupHandlerTest{T: srvt.T{
			Name: "Password too short",
			Request: srvt.Request{
				Method: "POST",
				Body:   `{"Name":"tototo","Email":"toto@example.com","Passwd":"t"}`,
			},
			Checker: srvt.CheckError{http.StatusBadRequest, "Passwd too short"},
		}},
		&signupHandlerTest{T: srvt.T{
			Name: "Wrong email 1",
			Request: srvt.Request{
				Method: "POST",
				Body:   `{"Name":"tototo","Email":"toto.example.com","Passwd":"tititi"}`,
			},
			Checker: srvt.CheckError{http.StatusBadRequest, "Email invalid"},
		}},
		&signupHandlerTest{T: srvt.T{
			Name: "Wrong email 2",
			Request: srvt.Request{
				Method: "POST",
				Body:   `{"Name":"tototo","Email":"toto@examplecom","Passwd":"tititi"}`,
			},
			Checker: srvt.CheckError{http.StatusBadRequest, "Email invalid"},
		}},
		&signupHandlerTest{T: srvt.T{
			Name: "Success",
			Request: srvt.Request{
				Method: "POST",
				Body:   body,
			},
		}},
		SignupHandlerTest(signupHandlerTestWithUser_{
			Name: "Name already exists",
			RequestFct: func(t *testing.T, basename string) *srvt.Request {
				other := basename + "-req"
				return &srvt.Request{
					Method: "POST",
					Body: `{"Name":"` + strings.TrimSpace(dbt.UserNameWith(basename)) +
						`","Email":"` + dbt.UserEmailWith(other) + `","Passwd":"tititi"}`,
				}
			},
			Checker: srvt.CheckError{http.StatusConflict, "Already exists"},
		}),
		SignupHandlerTest(signupHandlerTestWithUser_{
			Name: "Email already exists",
			RequestFct: func(t *testing.T, basename string) *srvt.Request {
				other := basename + "-req"
				return &srvt.Request{
					Method: "POST",
					Body: `{"Name":"` + strings.TrimSpace(dbt.UserNameWith(other)) +
						`","Email":"` + dbt.UserEmailWith(basename) + `","Passwd":"tititi"}`,
				}
			},
			Checker: srvt.CheckError{http.StatusConflict, "Already exists"},
		}),
	}
	srvt.Run(t, tests, SignupHandler)
}
