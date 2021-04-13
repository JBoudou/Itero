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
	"testing"
	"net/http"
	
	srvt "github.com/JBoudou/Itero/server/servertest"
)

func TestRefreshHandler(t *testing.T) {
	var userId uint32 = 27

	tests := []srvt.Test{
		&srvt.T{
			Name: "No user",
			Request: srvt.Request {
				Method: "POST",
			},
			Checker: srvt.CheckStatus{http.StatusForbidden},
		},
		&srvt.T{
			Name: "GET",
			Request: srvt.Request {
				UserId: &userId,
			},
			Checker: srvt.CheckStatus{http.StatusForbidden},
		},
		&srvt.T{
			Name: "Success",
			Request: srvt.Request {
				UserId: &userId,
				Method: "POST",
			},
			Checker: srvt.CheckStatus{http.StatusOK},
		},
	}

	srvt.RunFunc(t, tests, RefreshHandler)
}
