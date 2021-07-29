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

	dbt "github.com/JBoudou/Itero/mid/db/dbtest"
	"github.com/JBoudou/Itero/mid/server"
	srvt "github.com/JBoudou/Itero/mid/server/servertest"
	"github.com/JBoudou/Itero/pkg/ioc"
)

type loginTest struct {
	Name    string
	Body    func(env *dbt.Env, t *testing.T) string
	Checker srvt.Checker

	dbEnv dbt.Env
}

func (self *loginTest) GetName() string {
	return self.Name
}

func (self *loginTest) Prepare(t *testing.T, loc *ioc.Locator) *ioc.Locator {
	t.Parallel()
	self.dbEnv.CreateUserWith(t.Name())
	if checker, ok := self.Checker.(interface{ Before(*testing.T) }); ok {
		checker.Before(t)
	}
	return loc
}

func (self *loginTest) GetRequest(t *testing.T) *srvt.Request {
	return &srvt.Request{
		Method: "POST",
		Body:   self.Body(&self.dbEnv, t),
	}
}

func (self *loginTest) Check(t *testing.T, response *http.Response, request *server.Request) {
	self.Checker.Check(t, response, request)
}

func (self *loginTest) Close() {
	self.dbEnv.Close()
}

func TestLoginHandler(t *testing.T) {
	precheck(t)
	t.Parallel()

	env := new(dbt.Env)
	defer env.Close()

	env.CreateUser()
	if env.Error != nil {
		t.Fatalf("Env failed: %s", env.Error)
	}

	tests := []srvt.Test{
		&loginTest{
			Name: "no body",
			Body: func(env *dbt.Env, t *testing.T) string {
				return ""
			},
			Checker: srvt.CheckStatus{http.StatusBadRequest},
		},
		&loginTest{
			Name: "empty user",
			Body: func(env *dbt.Env, t *testing.T) string {
				return `{"Passwd":"` + dbt.UserPasswd + `"}`
			},
			Checker: srvt.CheckStatus{http.StatusForbidden},
		},
		&loginTest{
			Name: "empty passwd",
			Body: func(env *dbt.Env, t *testing.T) string {
				return `{"User":"` + dbt.UserNameWith(t.Name()) + `"}`
			},
			Checker: srvt.CheckStatus{http.StatusForbidden},
		},
		&loginTest{
			Name: "success user",
			Body: func(env *dbt.Env, t *testing.T) string {
				return `{"User":"` + dbt.UserNameWith(t.Name()) + `","Passwd":"` + dbt.UserPasswd + `"}`
			},
			Checker: srvt.CheckStatus{http.StatusOK},
		},
		&loginTest{
			Name: "success email",
			Body: func(env *dbt.Env, t *testing.T) string {
				return `{"User":"` + dbt.UserEmailWith(t.Name()) + `","Passwd":"` + dbt.UserPasswd + `"}`
			},
			Checker: srvt.CheckStatus{http.StatusOK},
		},
	}
	srvt.RunFunc(t, tests, LoginHandler)
}
