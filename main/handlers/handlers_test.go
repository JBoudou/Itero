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
	"testing"

	"github.com/JBoudou/Itero/mid/db"
	"github.com/JBoudou/Itero/mid/server"
	srvt "github.com/JBoudou/Itero/mid/server/servertest"
	"github.com/JBoudou/Itero/pkg/config"
)

func mustt(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func precheck(t *testing.T) {
	if !(config.Ok && db.Ok && server.Ok) {
		t.Log("Impossible to test the main package: some dependent packages are not ok.")
		t.Log("Check that there is a configuration file in main/. (or a link to the main configuation file).")
		t.SkipNow()
	}
}

func makePollRequest(t *testing.T, pollId uint32, userId *uint32) *srvt.Request {
	pollSegment := PollSegment{Salt: 42, Id: pollId}
	encoded, err := pollSegment.Encode()
	mustt(t, err)
	target := "/a/test/" + encoded
	return &srvt.Request{Target: &target, UserId: userId}
}
