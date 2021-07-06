// Itero - Online iterative vote application
// Copyright (C) 2021 Joseph Boudou
// 
// This program is free software: you can redistribute it and/or modify it under
// the terms of the GNU General Public License as published by the Free Software
// Foundation, either version 3 of the License, or (at your option) any later
// version.
// 
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS
// FOR A PARTICULAR PURPOSE. See the GNU General Public License for more
// details.
// 
// You should have received a copy of the GNU General Public License along with
// this program. If not, see <https://www.gnu.org/licenses/>.

package handlers

import (
	"context"
	"net/http"

	"github.com/JBoudou/Itero/main/services"
	"github.com/JBoudou/Itero/mid/db"
	"github.com/JBoudou/Itero/mid/server"
	"github.com/JBoudou/Itero/pkg/events"
)

type forgotHandler struct {
	evtManager events.Manager
}

// ForgotHandler handles requests to change a user password when the user has forgotten the current
// password.
func ForgotHandler(evtManager events.Manager) forgotHandler {
	return forgotHandler{evtManager}
}

func (self forgotHandler) Handle(ctx context.Context, response server.Response, request *server.Request) {
	if err := request.CheckPOST(ctx); err != nil {
		response.SendError(ctx, err)
		return
	}
	
	var forgotQuery struct {
		User string
	}
	if err := request.UnmarshalJSONBody(&forgotQuery); err != nil {
		err = server.WrapError(http.StatusBadRequest, "Wrong request", err)
		response.SendError(ctx, err)
		return
	}

	userInfo, err := getUserInfo(ctx, forgotQuery.User)
	must(err)

	const qCheck =`
	  SELECT 1 FROM Confirmations WHERE User = ? AND Type = ? AND Expires > CURRENT_TIMESTAMP`
	rows, err := db.DB.QueryContext(ctx, qCheck, userInfo.Id, db.ConfirmationTypePasswd)
	must(err)
	defer rows.Close()
	if rows.Next() {
		panic(server.NewHttpError(http.StatusConflict,
			"Already sent", "A forgotten password request is still active"))
	}
	
	self.evtManager.Send(services.ForgotEvent{User: userInfo.Id})
	response.SendJSON(ctx, "Ok")
}
