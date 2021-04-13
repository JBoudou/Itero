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
	"github.com/JBoudou/Itero/server"
)

func DeleteHandler(ctx context.Context, response server.Response, request *server.Request) {
	const (
		ImpossibleStatus  = http.StatusLocked
		ImpossibleMessage = "Not deletable"
	)

	pollInfo, err := checkPollAccess(ctx, request)
	must(err)
	if pollInfo.CurrentRound > 0 {
		panic(server.NewHttpError(ImpossibleStatus, ImpossibleMessage, "First round terminated"))
	}

	const qDelete = `
	  DELETE FROM Polls
		 WHERE Id = ? AND Admin = ?
		   AND ( State = 'Waiting' OR
			       ( State = 'Active' AND CurrentRound = 0 AND
						   ADDTIME(CurrentRoundStart, MaxRoundDuration) > CURRENT_TIMESTAMP ) )`

	result, err := db.DB.ExecContext(ctx, qDelete, pollInfo.Id, request.User.Id)
	must(err)
	affected, err := result.RowsAffected()
	must(err)
	if affected == 0 {
		panic(server.NewHttpError(ImpossibleStatus, ImpossibleMessage, "The query affects no row"))
	}

	response.SendJSON(ctx, "Ok")
}
