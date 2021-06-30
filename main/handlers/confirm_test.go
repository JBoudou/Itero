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
	"encoding/json"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/JBoudou/Itero/mid/db"
	dbt "github.com/JBoudou/Itero/mid/db/dbtest"
	"github.com/JBoudou/Itero/mid/root"
	"github.com/JBoudou/Itero/mid/salted"
	"github.com/JBoudou/Itero/mid/server"
	srvt "github.com/JBoudou/Itero/mid/server/servertest"
	"github.com/JBoudou/Itero/pkg/ioc"
)

type confirmTest struct {
	Name   string
	Create []confirmTestEntry

	// Fct is called to make the request. If it is nil, the first segment is used for the request.
	Fct func(*testing.T, []salted.Segment) salted.Segment

	// If nil, check the returned type and the deletion of the confirmation
	Checker srvt.Checker

	dbEnv     dbt.Env
	uid       uint32
	segments  []salted.Segment
	requested salted.Segment
}

type confirmTestEntry struct {
	typ db.ConfirmationType
	dur time.Duration
}

func (self *confirmTest) GetName() string {
	return self.Name
}

func (self *confirmTest) Prepare(t *testing.T) *ioc.Locator {
	t.Parallel()

	const qDelete = `DELETE FROM Confirmations WHERE Id = ?`
	self.segments = make([]salted.Segment, len(self.Create))
	var err error
	ctx := context.Background()
	for i, entry := range self.Create {
		self.uid = self.dbEnv.CreateUserWith(t.Name() + strconv.Itoa(i))
		self.dbEnv.Must(t)
		self.segments[i], err = db.CreateConfirmation(ctx, self.uid, entry.typ, entry.dur)
		mustt(t, err)
		self.dbEnv.Defer(func() { db.DB.Exec(qDelete, self.segments[i].Id) })
	}

	if self.Fct == nil {
		self.Fct = func(_ *testing.T, lst []salted.Segment) salted.Segment { return lst[0] }
	}

	if checker, ok := self.Checker.(interface{ Before(*testing.T) }); checker != nil && ok {
		checker.Before(t)
	}
	return root.IoC
}

func (self *confirmTest) GetRequest(t *testing.T) *srvt.Request {
	self.requested = self.Fct(t, self.segments)
	segment, err := self.requested.Encode()
	mustt(t, err)
	target := "/a/test/" + segment
	return &srvt.Request{
		Target: &target,
	}
}

func (self *confirmTest) Check(t *testing.T, response *http.Response, request *server.Request) {
	if self.Checker != nil {
		self.Checker.Check(t, response, request)
		return
	}

	if response.StatusCode != http.StatusOK {
		t.Errorf("Wrong status. Got %d. Expect %d.", response.StatusCode, http.StatusOK)
	}

	var answer ConfirmAnswer
	mustt(t, json.NewDecoder(response.Body).Decode(&answer))
	expectType := self.Create[self.requestedIndex()].typ
	if answer.Type != expectType {
		t.Errorf("Wrong type. Got %s. Expect %s.", answer.Type, expectType)
	}

	const qFind = `SELECT 1 FROM Confirmations WHERE Id = ?`
	rows, err := db.DB.Query(qFind, self.requested.Id)
	defer rows.Close()
	mustt(t, err)
	if rows.Next() {
		t.Errorf("Confirmation with id %d has not been deleted", self.requested.Id)
	}

	switch expectType {
	case db.ConfirmationTypeVerify:
		const qCheckVerify = `SELECT Verified FROM Users WHERE Id = ?`
		var verified bool
		mustt(t, db.DB.QueryRow(qCheckVerify, self.uid).Scan(&verified))
		if !verified {
			t.Errorf("User %d not verified.", self.uid)
		}
	}
}

func (self *confirmTest) Close() {
	self.dbEnv.Close()
}

func (self *confirmTest) requestedIndex() (ret int) {
	for ret = 0; ret < len(self.segments); ret++ {
		if self.segments[ret].Id == self.requested.Id {
			return
		}
	}
	return
}

func TestConfirmHandler(t *testing.T) {
	t.Parallel()

	tests := []srvt.Test{

		&confirmTest{
			Name: "Id not found",
			Create: []confirmTestEntry{
				{typ: db.ConfirmationTypeVerify, dur: time.Minute},
				{typ: db.ConfirmationTypeVerify, dur: time.Minute},
			},
			Fct: func(t *testing.T, segments []salted.Segment) salted.Segment {
				const qDelete = `DELETE FROM Confirmations WHERE Id = ?`
				_, err := db.DB.Exec(qDelete, segments[0].Id)
				mustt(t, err)
				return segments[0]
			},
			Checker: srvt.CheckStatus{Code: http.StatusNotFound},
		},

		&confirmTest{
			Name: "Wrong salt",
			Create: []confirmTestEntry{
				{typ: db.ConfirmationTypeVerify, dur: time.Minute},
			},
			Fct: func(t *testing.T, segments []salted.Segment) salted.Segment {
				ret := segments[0]
				ret.Salt += 42
				return ret
			},
			Checker: srvt.CheckStatus{Code: http.StatusNotFound},
		},

		&confirmTest{
			Name: "Expired",
			Create: []confirmTestEntry{
				{typ: db.ConfirmationTypeVerify, dur: 0},
			},
			Fct: func(t *testing.T, segments []salted.Segment) salted.Segment {
				time.Sleep(10 * time.Millisecond)
				return segments[0]
			},
			Checker: srvt.CheckStatus{Code: http.StatusNotFound},
		},

		&confirmTest{
			Name: "Expired",
			Create: []confirmTestEntry{
				{typ: db.ConfirmationTypeVerify, dur: time.Minute},
			},
			Fct: func(t *testing.T, segments []salted.Segment) salted.Segment {
				return segments[0]
			},
		},
	}
	srvt.Run(t, tests, ConfirmHandler)
}
