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
	"net/http"
	"testing"

	dbt "github.com/JBoudou/Itero/mid/db/dbtest"
	srvt "github.com/JBoudou/Itero/mid/server/servertest"
	"github.com/JBoudou/Itero/pkg/ioc"
)

type createUserTest struct {
	srvt.WithName
	srvt.WithChecker
	dbt.WithDB

	Request func(uid *uint32) *srvt.Request

	uid uint32
}

func (self *createUserTest) Prepare(t *testing.T) *ioc.Locator {
	self.uid = self.DB.CreateUserWith(t.Name())
	self.DB.Must(t)

	return self.WithChecker.Prepare(t)
}

func (self *createUserTest) GetRequest(t *testing.T) *srvt.Request {
	return self.Request(&self.uid)
}

type createUserTest_ struct {
	Name    string
	Request func(uid *uint32) *srvt.Request
	Checker srvt.Checker
}

func CreateUserTest(c createUserTest_) *createUserTest {
	return &createUserTest{
		WithName:    srvt.WithName{c.Name},
		WithChecker: srvt.WithChecker{c.Checker},
		Request:     c.Request,
	}
}

func TestRefreshHandler(t *testing.T) {
	tests := []srvt.Test{
		CreateUserTest(createUserTest_{
			Name: "No user",
			Request: func(_ *uint32) *srvt.Request {
				return &srvt.Request{
					Method: "POST",
				}
			},
			Checker: srvt.CheckStatus{http.StatusForbidden},
		}),
		CreateUserTest(createUserTest_{
			Name: "GET",
			Request: func(uid *uint32) *srvt.Request {
				return &srvt.Request{
					UserId: uid,
				}
			},
			Checker: srvt.CheckStatus{http.StatusForbidden},
		}),
		CreateUserTest(createUserTest_{
			Name: "Success",
			Request: func(uid *uint32) *srvt.Request {
				return &srvt.Request{
					UserId: uid,
					Method: "POST",
				}
			},
			Checker: srvt.CheckStatus{http.StatusOK},
		}),
	}

	srvt.RunFunc(t, tests, RefreshHandler)
}
