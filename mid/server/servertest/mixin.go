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

package servertest

import (
	"net/http"
	"testing"

	"github.com/JBoudou/Itero/mid/server"
	"github.com/JBoudou/Itero/pkg/ioc"
)

// WithName //

type WithName struct {
	Name string
}

func (self WithName) GetName() string {
	return self.Name
}

// WithChecker //

type WithChecker struct {
	Checker Checker
}

func (self WithChecker) Prepare(t *testing.T) *ioc.Locator {
	if checker, ok := self.Checker.(interface{ Before(t *testing.T) }); ok {
		checker.Before(t)
	}
	return ioc.Root
}

func (self WithChecker) Check(t *testing.T, response *http.Response, request *server.Request) {
	if self.Checker == nil {
		t.Fatalf("No Checker")
	}
	self.Checker.Check(t, response, request)
}

// WithRequestFct //

type RequestFct = func (uid *uint32) *Request

type WithRequestFct struct {
	RequestFct RequestFct
	Uid uint32
}

func (self WithRequestFct) GetRequest(t *testing.T) *Request {
	return self.RequestFct(&self.Uid)
}

func RFGetNoSession(uid *uint32) *Request {
	return &Request{}
}

func RFPostNoSession(uid *uint32) *Request {
	return &Request{
		Method: "POST",
	}
}

func RFGetLogged(uid *uint32) *Request {
	return &Request{
		UserId: uid,
	}
}

func RFPostLogged(uid *uint32) *Request {
	return &Request{
		Method: "POST",
		UserId: uid,
	}
}
