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

type ConfirmAnswer struct {
	Type db.ConfirmationType
}

type confirmHandler struct {
}

func ConfirmHandler() confirmHandler {
	return confirmHandler{}
}

func (self confirmHandler) Handle(ctx context.Context, response server.Response, request *server.Request) {
	const (
		qSelect = `SELECT Salt, Type, User FROM Confirmations WHERE Id = ? AND Expires > CURRENT_TIMESTAMP`
		qDelete = `DELETE FROM Confirmations WHERE Id = ?`
	)

	segment, err := salted.FromRequest(request)
	must(err)

	// Verify
	var answer ConfirmAnswer
	var salt, uid uint32
	rows, err := db.DB.QueryContext(ctx, qSelect, segment.Id)
	defer rows.Close()
	must(err)
	if !rows.Next() {
		panic(server.NewHttpError(http.StatusNotFound, "Not found", "No such id"))
	}
	must(rows.Scan(&salt, &answer.Type, &uid))
	if segment.Salt != salt {
		panic(server.NewHttpError(http.StatusNotFound, "Not found", "Wrong salt"))
	}

	// Execute
	var delConfirm bool
	switch answer.Type {
	case db.ConfirmationTypeVerify:
		delConfirm, err = self.verify(ctx, uid)
	case db.ConfirmationTypePasswd:
		delConfirm = false
	}
	must(err)

	// Delete
	if delConfirm{
		_, err = db.DB.ExecContext(ctx, qDelete, segment.Id)
		must(err)
	}

	response.SendJSON(ctx, answer)
}

func (self confirmHandler) verify(ctx context.Context, uid uint32) (bool, error) {
	const qUpdate = `UPDATE Users SET Verified = TRUE WHERE Id = ?`
	_, err := db.DB.ExecContext(ctx, qUpdate, uid)
	return true, err
}
