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

	"github.com/JBoudou/Itero/mid/db"
	"github.com/JBoudou/Itero/mid/salted"
	"github.com/JBoudou/Itero/mid/server"
)

// PasswdHandler changes the password of an existing user. The request must reference a valid
// confirmation of type passwd.
func PasswdHandler(ctx context.Context, response server.Response, request *server.Request) {
	const (
		qVerify = `
		  SELECT Salt, User FROM Confirmations
			 WHERE Id = ? AND Type = ? AND Expires > CURRENT_TIMESTAMP`
		qDelete = `DELETE FROM Confirmations WHERE Id = ? AND Type = ? AND Expires > CURRENT_TIMESTAMP`
		qUpdate = `UPDATE Users SET Passwd = ? WHERE Id = ?`
	)

	segment, err := salted.FromRequest(request)
	must(err)

	var passwdQuery struct {
		Passwd string
	}
	if err := request.UnmarshalJSONBody(&passwdQuery); err != nil {
		err = server.WrapError(http.StatusBadRequest, "Wrong request", err)
		response.SendError(ctx, err)
		return
	}

	// Verify
	var uid uint32
	var salt uint32
	rows, err := db.DB.QueryContext(ctx, qVerify, segment.Id, db.ConfirmationTypePasswd)
	must(err)
	defer rows.Close()
	if !rows.Next() {
		panic(server.NewHttpError(http.StatusNotFound, "Not found", "No such id"))
	}
	must(rows.Scan(&salt, &uid))
	if segment.Salt != salt {
		panic(server.NewHttpError(http.StatusNotFound, "Not found", "Wrong salt"))
	}
	
	// Update
	hashPwd, err := checkAndHashPasswd(passwdQuery.Passwd)
	must(err)
	result, err := db.DB.ExecContext(ctx, qUpdate, hashPwd, uid)
	must(err)
	if nb, err := result.RowsAffected(); err == nil && nb != 1 {
		err = server.NewHttpError(http.StatusInternalServerError,
			"User not found", "Unable to update user password")
		response.SendError(ctx, err)
		return
	}

	// Delete confirmation
	result, err = db.DB.ExecContext(ctx, qDelete, segment.Id, db.ConfirmationTypePasswd)
	must(err)

	return
}
