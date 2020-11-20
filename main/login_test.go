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
	"net/http"
	"testing"

	"github.com/JBoudou/Itero/server"
	srvt "github.com/JBoudou/Itero/server/servertest"
	dbt  "github.com/JBoudou/Itero/db/dbtest"
)

func TestLoginHandler(t *testing.T) {
	env := new(dbt.Env)
	defer env.Close()

	env.CreateUser()
	if env.Error != nil {
		t.Fatalf("Env failed: %s", env.Error)
	}

	tests := []srvt.Test {
		{
			Name: "no body",
			Request: srvt.Request{
				Method: "GET",
			},
			Checker: srvt.CheckerStatus(http.StatusBadRequest),
		},
		{
			Name: "empty user",
			Request: srvt.Request{
				Method: "POST",
				Body:   `{"Passwd":"XYZ"}`,
			},
			Checker: srvt.CheckerStatus(http.StatusForbidden),
		},
		{
			Name: "empty passwd",
			Request: srvt.Request{
				Method: "POST",
				Body:   `{"User":" Test "}`,
			},
			Checker: srvt.CheckerStatus(http.StatusForbidden),
		},
		{
			Name: "success",
			Request: srvt.Request{
				Method: "POST",
				Body:   `{"User":" Test ","Passwd":"XYZ"}`,
			},
			Checker: srvt.CheckerStatus(http.StatusOK),
		},
	}
	srvt.Run(t, tests, server.HandlerFunc(LoginHandler))
}
