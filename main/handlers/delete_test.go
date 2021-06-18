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
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/JBoudou/Itero/main/services"
	"github.com/JBoudou/Itero/mid/db"
	dbt "github.com/JBoudou/Itero/mid/db/dbtest"
	"github.com/JBoudou/Itero/mid/salted"
	"github.com/JBoudou/Itero/mid/server"
	srvt "github.com/JBoudou/Itero/mid/server/servertest"
	"github.com/JBoudou/Itero/pkg/events"
	"github.com/JBoudou/Itero/pkg/ioc"
)

type deleteHandlerTest struct {
	dbt.WithDB
	WithEvent

	name          string
	state         string
	round         uint8   // First round is 0
	roundTime     float32 // if not zero, set CurrentRoundStart = NOW - roundTime * MaxRoundDuration
	expectStatus  int
	expectMessage string

	userId uint32
	pollId uint32
}

func (self *deleteHandlerTest) GetName() string {
	return self.name
}

func (self *deleteHandlerTest) Prepare(t *testing.T) *ioc.Locator {
	t.Parallel()

	self.userId = self.DB.CreateUserWith("DeleteHandle" + self.name)
	self.pollId = self.DB.CreatePoll("DeleteHandler", self.userId, db.ElectorateLogged)

	const (
		qState = `UPDATE Polls SET State = ?, Start = CURRENT_TIMESTAMP WHERE Id = ?`
		qRound = `UPDATE Polls SET CurrentRound = ? WHERE Id = ?`
		qTime  = `
		  UPDATE Polls
			   SET CurrentRoundStart = ADDTIME(CURRENT_TIMESTAMP, -1 * ? * MaxRoundDuration)
			 WHERE Id = ?`
	)

	if self.state != "" {
		self.DB.QuietExec(qState, self.state, self.pollId)
	}
	if self.round != 0 {
		self.DB.QuietExec(qRound, self.round, self.pollId)
	}
	if self.roundTime != 0. {
		self.DB.QuietExec(qTime, self.roundTime, self.pollId)
	}

	self.DB.Must(t)
	return self.WithEvent.Prepare(t)
}

func (self *deleteHandlerTest) GetRequest(t *testing.T) *srvt.Request {
	return makePollRequest(t, self.pollId, &self.userId)
}

func (self *deleteHandlerTest) Check(t *testing.T, response *http.Response, request *server.Request) {
	expectEvents := 0
	if self.expectStatus == 0 {
		self.expectStatus = http.StatusOK
		expectEvents = 1
	}
	if self.expectStatus != response.StatusCode {
		t.Errorf("Wrong status code. Got %d. Expect %d.", response.StatusCode, self.expectStatus)
	}

	if self.expectMessage != "" {
		builder := &strings.Builder{}
		_, err := io.Copy(builder, response.Body)
		mustt(t, err)
		message := strings.TrimSpace(builder.String())

		if self.expectMessage != message {
			t.Errorf("Wrong message. Got %s. Expect %s.", message, self.expectMessage)
		}
	}

	segment, err := salted.FromRequest(request)
	mustt(t, err)
	gotEvents := self.CountRecorderEvents(func(evt events.Event) bool {
		converted, ok := evt.(services.DeletePollEvent)
		return ok && converted.Poll == segment.Id
	})
	if gotEvents != expectEvents {
		t.Errorf("Got %d events. Expect %d.", gotEvents, expectEvents)
	}
}

type deleteHandlerTestWrongUser struct {
	deleteHandlerTest
}

func (self *deleteHandlerTestWrongUser) GetRequest(t *testing.T) *srvt.Request {
	userId := new(uint32)
	*userId = self.userId + 1
	return makePollRequest(t, self.pollId, userId)
}

type deleteHandlerTestNoUser struct {
	deleteHandlerTest
}

func (self *deleteHandlerTestNoUser) GetRequest(t *testing.T) *srvt.Request {
	return makePollRequest(t, self.pollId, nil)
}

type deleteHandlerTestWithVote struct {
	deleteHandlerTest
}

func (self *deleteHandlerTestWithVote) Prepare(t *testing.T) *ioc.Locator {
	ret := self.deleteHandlerTest.Prepare(t)
	for i := uint8(0); i <= self.round; i++ {
		self.DB.Vote(self.pollId, i, self.userId, 0)
	}
	self.DB.Must(t)
	return ret
}

func TestDeleteHandler(t *testing.T) {
	tests := []srvt.Test{
		&deleteHandlerTest{
			name:  "Waiting",
			state: "Waiting",
		},
		&deleteHandlerTest{
			name:          "Terminated",
			state:         "Terminated",
			expectStatus:  http.StatusLocked,
			expectMessage: "Not deletable",
		},
		&deleteHandlerTestNoUser{deleteHandlerTest{
			name:          "No user",
			expectStatus:  http.StatusForbidden,
			expectMessage: server.UnauthorizedHttpErrorMsg,
		}},
		&deleteHandlerTestWrongUser{deleteHandlerTest{
			name:          "Wrong admin",
			expectStatus:  http.StatusLocked,
			expectMessage: "Not deletable",
		}},
		&deleteHandlerTest{
			name:          "Round 2",
			round:         1,
			roundTime:     1.5,
			expectStatus:  http.StatusLocked,
			expectMessage: "Not deletable",
		},
		&deleteHandlerTest{
			name:          "Round 1 not over",
			roundTime:     0.5,
			expectStatus:  http.StatusLocked,
			expectMessage: "Not deletable",
		},
		&deleteHandlerTest{
			name:      "Round 1 over",
			roundTime: 1.1,
		},
		&deleteHandlerTestWithVote{deleteHandlerTest{
			name:      "Round 1 with vote",
			roundTime: 1.1,
		}},
	}

	srvt.Run(t, tests, DeleteHandler)
}
