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
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JBoudou/Itero/db"
	"github.com/JBoudou/Itero/server"
	srvt "github.com/JBoudou/Itero/server/servertest"
)

func TestSignupHandler_Error(t *testing.T) {
	precheck(t)

	tests := []srvt.Test{
		&srvt.T{
			Name:    "Bad request",
			Request: srvt.Request{Method: "POST"},
			Checker: srvt.CheckStatus{http.StatusBadRequest},
		},
		&srvt.T{
			Name: "Name too short",
			Request: srvt.Request{
				Method: "POST",
				Body:   `{"Name":"a","Email":"toto@example.com","Passwd":"tititi"}`,
			},
			Checker: srvt.CheckError{http.StatusBadRequest, "Name too short"},
		},
		&srvt.T{
			Name: "Name starting with a space",
			Request: srvt.Request{
				Method: "POST",
				Body:   `{"Name":" tototo","Email":"toto@example.com","Passwd":"tititi"}`,
			},
			Checker: srvt.CheckError{http.StatusBadRequest, "Name has spaces"},
		},
		&srvt.T{
			Name: "Name ending with a space",
			Request: srvt.Request{
				Method: "POST",
				Body:   `{"Name":"tototo ","Email":"toto@example.com","Passwd":"tititi"}`,
			},
			Checker: srvt.CheckError{http.StatusBadRequest, "Name has spaces"},
		},
		&srvt.T{
			Name: "Name containing @",
			Request: srvt.Request{
				Method: "POST",
				Body:   `{"Name":"toto@to","Email":"toto@example.com","Passwd":"tititi"}`,
			},
			Checker: srvt.CheckError{http.StatusBadRequest, "Name has at sign"},
		},
		&srvt.T{
			Name: "Password too short",
			Request: srvt.Request{
				Method: "POST",
				Body:   `{"Name":"tototo","Email":"toto@example.com","Passwd":"t"}`,
			},
			Checker: srvt.CheckError{http.StatusBadRequest, "Passwd too short"},
		},
		&srvt.T{
			Name: "Wrong email 1",
			Request: srvt.Request{
				Method: "POST",
				Body:   `{"Name":"tototo","Email":"toto.example.com","Passwd":"tititi"}`,
			},
			Checker: srvt.CheckError{http.StatusBadRequest, "Email invalid"},
		},
		&srvt.T{
			Name: "Wrong email 2",
			Request: srvt.Request{
				Method: "POST",
				Body:   `{"Name":"tototo","Email":"toto@examplecom","Passwd":"tititi"}`,
			},
			Checker: srvt.CheckError{http.StatusBadRequest, "Email invalid"},
		},
	}
	srvt.Run(t, tests, server.HandlerFunc(SignupHandler))
}

func TestSignupHandler_Success(t *testing.T) {
	precheck(t)

	const name = "toto_my_test_user_with_a_long_name"
	var called bool
	var userId uint32
	response := srvt.MockResponse{
		T: t,
		LoginFct: func(t *testing.T, ctx context.Context, user server.User, request *server.Request) {
			if user.Name != name {
				t.Errorf("Wrong name. Got %s. Expect %s.", user.Name, name)
			}
			userId = user.Id
			called = true
		},
	}

	const body = `{"Name":"` + name + `","Email":"` + name + `@example.com","Passwd":"tititi"}`
	tRequest := srvt.Request{Method: "POST", Body: body}
	hRequest, err := tRequest.Make()
	if err != nil {
		t.Fatal(err)
	}
	wrapper := server.NewHandlerWrapper("/a/test", server.HandlerFunc(SignupHandler))
	ctx, _, sRequest := wrapper.MakeParams(httptest.NewRecorder(), hRequest)
	wrapper.Exec(ctx, response, sRequest)

	if !called {
		t.Fatal("SendLoginAccepted not called")
	}

	tests := []srvt.Test{
		&srvt.T{
			Name: "Name already exists",
			Request: srvt.Request{
				Method: "POST",
				Body:   `{"Name":"` + name + `","Email":"another_long_dummy@example.com","Passwd":"tititi"}`,
			},
			Checker: srvt.CheckError{http.StatusBadRequest, "Already exists"},
		},
		&srvt.T{
			Name: "Name already exists",
			Request: srvt.Request{
				Method: "POST",
				Body:   `{"Name":"another_long_dummy","Email":"` + name + `@example.com","Passwd":"tititi"}`,
			},
			Checker: srvt.CheckError{http.StatusBadRequest, "Already exists"},
		},
	}
	srvt.Run(t, tests, server.HandlerFunc(SignupHandler))

	const qDelete = `DELETE FROM Users WHERE Id = ?`
	result, err := db.DB.Exec(qDelete, userId)
	if err != nil {
		t.Fatal(err)
	}
	nbRows, err := result.RowsAffected()
	if err != nil {
		t.Fatal(err)
	}
	if nbRows != 1 {
		t.Fatalf("Not deleting the user (%d rows instead of 1).", nbRows)
	}
}
