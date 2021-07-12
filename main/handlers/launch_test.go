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

	"github.com/JBoudou/Itero/main/services"
	"github.com/JBoudou/Itero/mid/db"
	"github.com/JBoudou/Itero/mid/server"
	srvt "github.com/JBoudou/Itero/mid/server/servertest"
	"github.com/JBoudou/Itero/pkg/events"
)

func launchHandlerCheckerFactory(param PollTestCheckerFactoryParam) srvt.Checker {
	return srvt.CheckerFun(func(t *testing.T, response *http.Response, request *server.Request) {
		if response.StatusCode != http.StatusOK {
			t.Errorf("Wrong status. Got %d. Expect %d.", response.StatusCode, http.StatusNotFound)
		}

		const qCheck = `SELECT State FROM Polls WHERE Id = ?`
		rows, err := db.DB.Query(qCheck, param.PollId)
		mustt(t, err)
		defer rows.Close()
		if !rows.Next() {
			t.Fatalf("Poll not found")
		}
		var state db.State
		mustt(t, rows.Scan(&state))
		
		if state != db.StateActive {
			t.Errorf("Wrong poll state. Got %v. Expect %v.", state, db.StateActive)
		}
	})
}

func launchHandlerEventPredicate(param PollTestCheckerFactoryParam, evt events.Event) bool {
	startEvt, ok := evt.(services.StartPollEvent)
	return ok && startEvt.Poll == param.PollId
}

func TestLaunchHandler(t *testing.T) {
	tests := []srvt.Test{
		&wrongPollTest{
			Kind: wrongPollTestKindNoPoll,
		},
		&wrongPollTest{
			Kind: wrongPollTestKindWrongSalt,
		},
		&pollTest{
			Name: "Not logged",
			Electorate: db.ElectorateAll,
			Waiting: true,
			UserType: pollTestUserTypeNone,
			Checker: srvt.CheckStatus{http.StatusForbidden},
		},
		&pollTest{
			Name: "Not admin",
			Electorate: db.ElectorateAll,
			Waiting: true,
			UserType: pollTestUserTypeLogged,
			Checker: srvt.CheckStatus{http.StatusForbidden},
		},
		&pollTest{
			Name: "Not waiting",
			Electorate: db.ElectorateAll,
			Waiting: false,
			UserType: pollTestUserTypeAdmin,
			Checker: srvt.CheckStatus{http.StatusBadRequest},
		},
		&pollTest{
			Name: "Success",
			Electorate: db.ElectorateAll,
			Waiting: true,
			UserType: pollTestUserTypeAdmin,
			Checker: launchHandlerCheckerFactory,
			EventPredicate: launchHandlerEventPredicate,
			EventCount: 1,
		},
	}
	srvt.Run(t, tests, LaunchHandler)
}
