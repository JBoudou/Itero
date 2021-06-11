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
	"net/http"

	"github.com/JBoudou/Itero/main/services"
	"github.com/JBoudou/Itero/mid/db"
	"github.com/JBoudou/Itero/mid/server"
	"github.com/JBoudou/Itero/pkg/events"
)

type reverifyHandler struct {
	evtManager events.Manager
}

func ReverifyHandler(evtManager events.Manager) reverifyHandler {
	return reverifyHandler{evtManager}
}

func (self reverifyHandler) Handle(ctx context.Context, response server.Response, request *server.Request) {
	if request.User == nil || !request.User.Logged {
		panic(server.UnauthorizedHttpError("No session"))
	}

	const qCheck = `
	  SELECT 1 FROM Confirmations
		WHERE User = ? AND Type = ? AND Expires > CURRENT_TIMESTAMP`
	rows, err := db.DB.QueryContext(ctx, qCheck, request.User.Id, db.ConfirmationTypeVerify)
	must(err)
	defer rows.Close()
	if rows.Next() {
		panic(server.NewHttpError(http.StatusConflict,
			"Already sent", "A verify confirmation is still active"))
	}

	self.evtManager.Send(services.ReverifyEvent{User: request.User.Id})
	response.SendJSON(ctx, "Ok")
}
