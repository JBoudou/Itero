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
	"database/sql"
	"net/http"

	"github.com/JBoudou/Itero/main/services"
	"github.com/JBoudou/Itero/mid/db"
	"github.com/JBoudou/Itero/mid/salted"
	"github.com/JBoudou/Itero/mid/server"
	"github.com/JBoudou/Itero/pkg/events"
	"github.com/JBoudou/Itero/pkg/slog"
)

type launchHandler struct {
	evtManager events.Manager
}

// LaunchHandler forces the start of a waiting poll.
func LaunchHandler(evtManager events.Manager) launchHandler {
	return launchHandler{evtManager: evtManager}
}

func (self launchHandler) Handle(ctx context.Context, response server.Response, request *server.Request) {
	if request.User == nil || !request.User.Logged {
		response.SendError(ctx, server.UnauthorizedHttpError("Not logged"))
		return
	}

	segment, err := salted.FromRequest(request)
	must(err)

	const (
		qCheck  = `SELECT Salt, Admin, State FROM Polls WHERE Id = ? LOCK IN SHARE MODE`
		qUpdate = `
		  UPDATE Polls SET State = 'Active', CurrentRoundStart = CURRENT_TIMESTAMP
			WHERE Id = ? AND Admin = ? AND State = 'Waiting'`
	)

	db.RepeatDeadlocked(slog.CtxLoadLogger(ctx), ctx, nil, func(tx *sql.Tx) {
		rows, err := tx.QueryContext(ctx, qCheck, segment.Id)
		must(err)
		defer rows.Close()
		if !rows.Next() {
			panic(noPollError("No such Id"))
		}

		var salt, admin uint32
		var state db.State
		must(rows.Scan(&salt, &admin, &state))
		rows.Close()
		if salt != segment.Salt {
			panic(noPollError("Wrong salt"))
		}
		if admin != request.User.Id {
			panic(server.UnauthorizedHttpError("Not admin"))
		}
		if state != db.StateWaiting {
			panic(server.NewHttpError(http.StatusBadRequest, "Not waiting", "Not waiting"))
		}

		result, err := tx.ExecContext(ctx, qUpdate, segment.Id, request.User.Id)
		must(err)
		affected, err := result.RowsAffected()
		must(err)
		if affected != 1 {
			panic(server.NewHttpError(http.StatusInternalServerError, "Not started",
				"The request did not change one row"))
		}
	})

	self.evtManager.Send(services.StartPollEvent{Poll: segment.Id})
}
