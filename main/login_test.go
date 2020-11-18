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
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/JBoudou/Itero/db"
	"github.com/JBoudou/Itero/server"
)

func TestLoginHandler(t *testing.T) {
	result, err := db.DB.Exec(`INSERT INTO Users(Name, Email, Passwd)
	                           VALUES(' Test ', 'test@example.test',
	                           X'2e43477a2da06cb4aba764381086cbc9323945eb1bffb232f221e374af44f803')`)
	if err != nil {
		t.Fatalf("Insert failed: %s", err)
	}
	userId, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("Insert failed: %s", err)
	}
	defer func () {
		db.DB.Exec(`DELETE FROM Users WHERE Id = ?`, userId)
	} ()

	type req struct {
		method string
		target string
		body   string
	}
	tests := []struct {
		name    string
		req     req
		checker server.TestChecker
	}{
		{
			name: "no body",
			req: req{
				method: "GET",
			},
			checker: server.TestCheckerStatus(http.StatusBadRequest),
		},
		{
			name: "empty user",
			req: req{
				method: "POST",
				body:   `{"Passwd":"XYZ"}`,
			},
			checker: server.TestCheckerStatus(http.StatusForbidden),
		},
		{
			name: "empty passwd",
			req: req{
				method: "POST",
				body:   `{"User":" Test "}`,
			},
			checker: server.TestCheckerStatus(http.StatusForbidden),
		},
		{
			name: "success",
			req: req{
				method: "POST",
				body:   `{"User":" Test ","Passwd":"XYZ"}`,
			},
			checker: server.TestCheckerStatus(http.StatusOK),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target := tt.req.target
			if target == "" {
				target = "/a/test"
			}
			var body io.Reader
			if tt.req.body != "" {
				body = strings.NewReader(tt.req.body)
			}
			req := httptest.NewRequest(tt.req.method, target, body)
			response := server.TestHandlerFunc("/a/test", LoginHandler, req)
			tt.checker(t, response, req)
		})
	}
}
