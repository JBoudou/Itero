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
	"github.com/JBoudou/Itero/mid/salted"
	"github.com/JBoudou/Itero/mid/server"
	"github.com/JBoudou/Itero/pkg/events"
)

func DeleteHandler(ctx context.Context, response server.Response, request *server.Request) {
	const (
		ImpossibleStatus  = http.StatusLocked
		ImpossibleMessage = "Not deletable"
	)

	if request.User == nil {
		panic(server.UnauthorizedHttpError("No session"))
	}

	segment, err := salted.FromRequest(request)
	must(err)
	event := services.DeletePollEvent{Poll: segment.Id, Participants: make(map[uint32]bool, 2)}

	const (
		qTitle = `
		  SELECT Title FROM Polls
		   WHERE Id = ? AND Admin = ?
		     AND ( State = 'Waiting' OR
		  	       ( State = 'Active' AND CurrentRound = 0 AND
		  				   ADDTIME(CurrentRoundStart, MaxRoundDuration) < CURRENT_TIMESTAMP ) )`
		qParticipants = `SELECT DISTINCT User FROM Participants WHERE Poll = ?`
		qDelete       = `DELETE FROM Polls WHERE Id = ?`
	)

	rows, err := db.DB.QueryContext(ctx, qTitle, segment.Id, request.User.Id)
	defer rows.Close()
	must(err)
	if !rows.Next() {
		panic(server.NewHttpError(ImpossibleStatus, ImpossibleMessage, ""))
	}
	must(rows.Scan(&event.Title))
	if rows.Next() {
		panic(server.NewHttpError(http.StatusInternalServerError, server.InternalHttpErrorMsg,
			"Two polls witht the same Id"))
	}

	rows, err = db.DB.QueryContext(ctx, qParticipants, segment.Id)
	defer rows.Close()
	must(err)
	for rows.Next() {
		var uid uint32
		must(rows.Scan(&uid))
		event.Participants[uid] = true
	}
	event.Participants[request.User.Id] = true

	result, err := db.DB.ExecContext(ctx, qDelete, segment.Id)
	must(err)
	affected, err := result.RowsAffected()
	must(err)
	if affected == 0 {
		panic(server.NewHttpError(ImpossibleStatus, ImpossibleMessage, "The query affects no row"))
	}

	events.Send(event)
	response.SendJSON(ctx, "Ok")
}
