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
	"bytes"
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/JBoudou/Itero/mid/db"
	dbt "github.com/JBoudou/Itero/mid/db/dbtest"
	"github.com/JBoudou/Itero/mid/salted"
	"github.com/JBoudou/Itero/mid/server"
	srvt "github.com/JBoudou/Itero/mid/server/servertest"
	"github.com/JBoudou/Itero/pkg/ioc"
)

type passwdTest_ struct {
	Name            string
	ConfirmType     db.ConfirmationType
	ConfirmDuration time.Duration
	Delete          bool // Whether the confirmation is deleted before the request.
	Password        string
	Checker         srvt.Checker // Must be nil for successful requests.
}

func PasswdTest(c passwdTest_) *passwdTest {
	return &passwdTest{
		WithName:        srvt.WithName{Name: c.Name},
		ConfirmType:     c.ConfirmType,
		ConfirmDuration: c.ConfirmDuration,
		Delete:          c.Delete,
		Password:        c.Password,
		Checker:         c.Checker,
	}
}

type passwdTest struct {
	srvt.WithName
	dbt.WithDB

	ConfirmType     db.ConfirmationType
	ConfirmDuration time.Duration
	Delete          bool
	Password        string
	Checker         srvt.Checker

	uid     uint32
	segment salted.Segment
}

func (self *passwdTest) Prepare(t *testing.T, loc *ioc.Locator) *ioc.Locator {
	t.Parallel()

	self.uid = self.DB.CreateUserWith(t.Name())
	self.DB.Must(t)

	var err error
	self.segment, err = db.CreateConfirmation(context.Background(),
		self.uid, self.ConfirmType, self.ConfirmDuration)
	mustt(t, err)

	const qDelete = `DELETE FROM Confirmations WHERE Id = ?`
	if self.Delete {
		_, err = db.DB.Exec(qDelete, self.segment.Id)
		mustt(t, err)
	}

	return loc
}

func (self passwdTest) GetRequest(t *testing.T) *srvt.Request {
	encoded, err := self.segment.Encode()
	mustt(t, err)
	target := "/a/test/" + encoded
	return &srvt.Request{
		Method: "POST",
		Target: &target,
		Body:   `{"Passwd":"` + self.Password + `"}`,
	}
}

func (self passwdTest) Check(t *testing.T, response *http.Response, request *server.Request) {
	if self.Checker != nil {
		self.Checker.Check(t, response, request)
	}

	const (
		qConfirm  = `SELECT 1 FROM Confirmations WHERE Id = ?`
		qPassword = `SELECT Passwd FROM Users WHERE Id = ?`
	)
	success := self.Checker == nil

	rows, err := db.DB.Query(qConfirm, self.segment.Id)
	mustt(t, err)
	defer rows.Close()
	gotDeleted := !rows.Next()
	expectDeleted := success || self.Delete
	if gotDeleted != expectDeleted {
		t.Errorf("Confirmation delete %t. Expect %t.", gotDeleted, expectDeleted)
	}
	rows.Close()

	rows, err = db.DB.Query(qPassword, self.uid)
	mustt(t, err)
	if !rows.Next() {
		t.Errorf("User not found")
	}
	var passwd []byte
	mustt(t, rows.Scan(&passwd))
	gotPasswd := !bytes.Equal(passwd, dbt.UserPasswdHash)
	if gotPasswd != success {
		t.Errorf("Password changed %t. Expect %t.", gotPasswd, success)
	}
}

func TestPasswdHandler(t *testing.T) {
	precheck(t)
	t.Parallel()

	tests := []srvt.Test{
		PasswdTest(passwdTest_{
			Name: "No Confirmation",
			ConfirmType: db.ConfirmationTypePasswd,
			ConfirmDuration: time.Minute,
			Delete: true,
			Password: "123456",
			Checker: srvt.CheckStatus{http.StatusNotFound},
		}),
		PasswdTest(passwdTest_{
			Name: "Expired",
			ConfirmType: db.ConfirmationTypePasswd,
			ConfirmDuration: -1 * time.Minute,
			Password: "123456",
			Checker: srvt.CheckStatus{http.StatusNotFound},
		}),
		PasswdTest(passwdTest_{
			Name: "Wrong type",
			ConfirmType: db.ConfirmationTypeVerify,
			ConfirmDuration: time.Minute,
			Password: "123456",
			Checker: srvt.CheckAnyErrorStatus,
		}),
		PasswdTest(passwdTest_{
			Name: "Too short",
			ConfirmType: db.ConfirmationTypePasswd,
			ConfirmDuration: time.Minute,
			Password: "12",
			Checker: srvt.CheckError{Code: http.StatusBadRequest, Body: "Passwd too short"},
		}),
		PasswdTest(passwdTest_{
			Name: "Success",
			ConfirmType: db.ConfirmationTypePasswd,
			ConfirmDuration: time.Minute,
			Password: "123456",
		}),
	}
	
	srvt.RunFunc(t, tests, PasswdHandler)
}
