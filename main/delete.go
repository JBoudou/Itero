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

	"github.com/JBoudou/Itero/db"
	"github.com/JBoudou/Itero/events"
	"github.com/JBoudou/Itero/server"
)

type DeletePollEvent struct {
	Poll uint32
}

func DeleteHandler(ctx context.Context, response server.Response, request *server.Request) {
	const (
		ImpossibleStatus  = http.StatusLocked
		ImpossibleMessage = "Not deletable"
	)

	if request.User == nil {
		panic(server.UnauthorizedHttpError("No session"))
	}

	segment, err := pollSegmentFromRequest(request)
	must(err)

	const qDelete = `
	  DELETE FROM Polls
		 WHERE Id = ? AND Admin = ?
		   AND ( State = 'Waiting' OR
			       ( State = 'Active' AND CurrentRound = 0 AND
						   ADDTIME(CurrentRoundStart, MaxRoundDuration) < CURRENT_TIMESTAMP ) )`

	result, err := db.DB.ExecContext(ctx, qDelete, segment.Id, request.User.Id)
	must(err)
	affected, err := result.RowsAffected()
	must(err)
	if affected == 0 {
		panic(server.NewHttpError(ImpossibleStatus, ImpossibleMessage, "The query affects no row"))
	}

	events.Send(DeletePollEvent{segment.Id})
	response.SendJSON(ctx, "Ok")
}
