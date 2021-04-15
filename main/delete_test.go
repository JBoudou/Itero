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

package main

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/JBoudou/Itero/db"
	dbt "github.com/JBoudou/Itero/db/dbtest"
	"github.com/JBoudou/Itero/server"
	srvt "github.com/JBoudou/Itero/server/servertest"
)

type deleteHandlerTest struct {
	name          string
	state         string
	wrongUser     bool
	round         uint8   // First round is 0
	roundTime     float32 // if not zero, set CurrentRoundStart = NOW - roundTime * MaxRoundDuration
	expectStatus  int
	expectMessage string

	dbEnv  dbt.Env
	userId uint32
	pollId uint32
}

func (self *deleteHandlerTest) GetName() string {
	return self.name
}

func (self *deleteHandlerTest) Prepare(t *testing.T) {
	t.Parallel()

	self.userId = self.dbEnv.CreateUserWith(self.name)
	self.pollId = self.dbEnv.CreatePoll("DeleteHandler", self.userId, db.PollPublicityPublicRegistered)

	const (
		qState = `UPDATE Polls SET State = ?, Start = CURRENT_TIMESTAMP WHERE Id = ?`
		qRound = `UPDATE Polls SET CurrentRound = ? WHERE Id = ?`
		qTime = `
		  UPDATE Polls
			   SET CurrentRoundStart = ADDTIME(CURRENT_TIMESTAMP, -1 * ? * MaxRoundDuration)
			 WHERE Id = ?`
	)

	if self.state != "" {
		self.dbEnv.QuietExec(qState, self.state, self.pollId)
	}
	if self.round != 0 {
		self.dbEnv.QuietExec(qRound, self.round, self.pollId)
	}
	if self.roundTime != 0. {
		self.dbEnv.QuietExec(qTime, self.roundTime, self.pollId)
	}
		
	self.dbEnv.Must(t)
}

func (self *deleteHandlerTest) GetRequest(t *testing.T) *srvt.Request {
	userId := &self.userId
	if self.wrongUser {
		userId = new(uint32)
		*userId = self.userId + 1;
	}
	return makePollRequest(t, self.pollId, userId)
}

func (self *deleteHandlerTest) Check(t *testing.T, response *http.Response, request *server.Request) {
	if self.expectStatus == 0 {
		self.expectStatus = http.StatusOK
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
}

func (self *deleteHandlerTest) Close() {
	self.dbEnv.Close()
}

func TestDeleteHandler(t *testing.T) {
	tests := []srvt.Test {
		&deleteHandlerTest{
			name: "Waiting",
			state: "Waiting",
		},
		&deleteHandlerTest{
			name: "Terminated",
			state: "Terminated",
			expectStatus: http.StatusLocked,
			expectMessage: "Not deletable",
		},
		&deleteHandlerTest{
			name: "Wrong admin",
			wrongUser: true,
			expectStatus: http.StatusLocked,
			expectMessage: "Not deletable",
		},
		&deleteHandlerTest{
			name: "Round 2",
			round: 1,
			roundTime: 1.5,
			expectStatus: http.StatusLocked,
			expectMessage: "Not deletable",
		},
		&deleteHandlerTest{
			name: "Round 1 not over",
			roundTime: 0.5,
			expectStatus: http.StatusLocked,
			expectMessage: "Not deletable",
		},
		&deleteHandlerTest{
			name: "Round 1 over",
			roundTime: 1.1,
		},
	}

	srvt.RunFunc(t, tests, DeleteHandler)
}
