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

	"github.com/JBoudou/Itero/mid/db"
	"github.com/JBoudou/Itero/mid/server"
)

// RefreshHandler creates a fresh new session from an older active session.
func RefreshHandler(ctx context.Context, response server.Response, request *server.Request) {
	if request.User == nil {
		if request.SessionError != nil {
			must(server.WrapUnauthorizedError(request.SessionError))
		} else {
			must(server.UnauthorizedHttpError("Cannot refresh without a session"))
		}
	}
	if err := request.CheckPOST(ctx); err != nil {
		must(server.WrapUnauthorizedError(err))
	}

	const qProfile = `SELECT Verified FROM Users WHERE Id = ?`
	var profileInfo ProfileInfo
	must(db.DB.QueryRowContext(ctx, qProfile, request.User.Id).Scan(&profileInfo.Verified))

	response.SendLoginAccepted(ctx, *request.User, request, profileInfo)
}
