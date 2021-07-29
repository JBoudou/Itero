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
	"strings"
	"testing"

	"github.com/JBoudou/Itero/mid/db"
	dbt "github.com/JBoudou/Itero/mid/db/dbtest"
	"github.com/JBoudou/Itero/mid/salted"
	"github.com/JBoudou/Itero/mid/server"
	srvt "github.com/JBoudou/Itero/mid/server/servertest"
	"github.com/JBoudou/Itero/pkg/b64buff"
	"github.com/JBoudou/Itero/pkg/ioc"
)

type shortURLTest struct {
	dbt.WithDB

	Name    string
	NoPoll  bool
	Checker srvt.Checker

	pid      uint32
	shortURL string
}

func (self shortURLTest) GetName() string {
	return self.Name
}

func (self *shortURLTest) Prepare(t *testing.T, loc *ioc.Locator) *ioc.Locator {
	t.Parallel()

	var err error
	self.shortURL, err = b64buff.RandomString(15)
	mustt(t, err)

	if !self.NoPoll {
		uid := self.DB.CreateUserWith(t.Name())
		self.pid = self.DB.CreatePoll("Test", uid, db.ElectorateAll)

		const qSetURL = `UPDATE Polls SET ShortURL = ? WHERE Id = ?`
		self.DB.QuietExec(qSetURL, self.shortURL, self.pid)

		self.DB.Must(t)
	}

	return loc
}

func (self shortURLTest) GetRequest(t *testing.T) *srvt.Request {
	target := "/a/test/" + self.shortURL
	return &srvt.Request{
		Target: &target,
	}
}

func (self shortURLTest) Check(t *testing.T, response *http.Response, request *server.Request) {
	if self.Checker != nil {
		self.Checker.Check(t, response, request)
		return
	}

	if response.StatusCode != 308 && response.StatusCode != 301 {
		t.Errorf("Wrong status code. Got %d. Expect 308 or 301.", response.StatusCode)
	}

	allSegments := strings.Split(response.Header.Get("Location"), "/")
	segment, err := salted.Decode(allSegments[len(allSegments) - 1])
	if err != nil {
		t.Fatalf("Unreadable segment \"%s\": %v.", allSegments[len(allSegments) - 1], err)
	}
	if segment.Id != self.pid {
		t.Errorf("Wrong poll id. Got %d. Expect %d.", segment.Id, self.pid)
	}
	if segment.Salt != dbt.PollSalt {
		t.Errorf("Wrong poll salt. Got %d. Expect %d.", segment.Salt, dbt.PollSalt)
	}
}

func TestShortURLHanlder(t *testing.T) {
	t.Parallel()

	tests := []srvt.Test{
		&shortURLTest{
			Name: "Success",
		},
		&shortURLTest{
			Name: "Not found",
			NoPoll: true,
			Checker: srvt.CheckStatus{http.StatusNotFound},
		},
	}

	srvt.RunFunc(t, tests, ShortURLHandler)
}
