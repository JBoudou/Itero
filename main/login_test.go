// Itero - Online iterative vote application
// Copyright (C) 2020 Joseph Boudou
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
	dbt "github.com/JBoudou/Itero/db/dbtest"
	"github.com/JBoudou/Itero/server"
	srvt "github.com/JBoudou/Itero/server/servertest"
)

func TestLoginHandler(t *testing.T) {
	precheck(t)

	env := new(dbt.Env)
	defer env.Close()

	env.CreateUser()
	if env.Error != nil {
		t.Fatalf("Env failed: %s", env.Error)
	}

	tests := []srvt.Test{
		{
			Name: "no body",
			Request: srvt.Request{
				Method: "POST",
			},
			Checker: srvt.CheckStatus{http.StatusBadRequest},
		},
		{
			Name: "empty user",
			Request: srvt.Request{
				Method: "POST",
				Body:   `{"Passwd":"XYZ"}`,
			},
			Checker: srvt.CheckStatus{http.StatusForbidden},
		},
		{
			Name: "empty passwd",
			Request: srvt.Request{
				Method: "POST",
				Body:   `{"User":" Test "}`,
			},
			Checker: srvt.CheckStatus{http.StatusForbidden},
		},
		{
			Name: "success",
			Request: srvt.Request{
				Method: "POST",
				Body:   `{"User":" Test ","Passwd":"XYZ"}`,
			},
			Checker: srvt.CheckStatus{http.StatusOK},
		},
	}
	srvt.Run(t, tests, server.HandlerFunc(LoginHandler))
}

func TestSignupHandler_Error(t *testing.T) {
	precheck(t)

	tests := []srvt.Test{
		{
			Name:    "Bad request",
			Request: srvt.Request{Method: "POST"},
			Checker: srvt.CheckStatus{http.StatusBadRequest},
		},
		{
			Name: "Name too short",
			Request: srvt.Request{
				Method: "POST",
				Body:   `{"Name":"a","Email":"toto@example.com","Passwd":"tititi"}`,
			},
			Checker: srvt.CheckError{http.StatusBadRequest, "Name too short"},
		},
		{
			Name: "Name starting with a space",
			Request: srvt.Request{
				Method: "POST",
				Body:   `{"Name":" tototo","Email":"toto@example.com","Passwd":"tititi"}`,
			},
			Checker: srvt.CheckError{http.StatusBadRequest, "Name has spaces"},
		},
		{
			Name: "Name ending with a space",
			Request: srvt.Request{
				Method: "POST",
				Body:   `{"Name":"tototo ","Email":"toto@example.com","Passwd":"tititi"}`,
			},
			Checker: srvt.CheckError{http.StatusBadRequest, "Name has spaces"},
		},
		{
			Name: "Password too short",
			Request: srvt.Request{
				Method: "POST",
				Body:   `{"Name":"tototo","Email":"toto@example.com","Passwd":"t"}`,
			},
			Checker: srvt.CheckError{http.StatusBadRequest, "Passwd too short"},
		},
		{
			Name: "Wrong email 1",
			Request: srvt.Request{
				Method: "POST",
				Body:   `{"Name":"tototo","Email":"toto.example.com","Passwd":"tititi"}`,
			},
			Checker: srvt.CheckError{http.StatusBadRequest, "Email invalid"},
		},
		{
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
		{
			Name: "Name already exists",
			Request: srvt.Request{
				Method: "POST",
				Body:   `{"Name":"` + name + `","Email":"another_long_dummy@example.com","Passwd":"tititi"}`,
			},
			Checker: srvt.CheckError{http.StatusBadRequest, "Already exists"},
		},
		{
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
