// Itero - Online iterative vote application
// Copyright (C) 2020 Joseph Boudou
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
	"bytes"
	"context"
	"database/sql"
	"errors"
	"hash"
	"net/http"
	"strings"

	"golang.org/x/crypto/blake2b"

	"github.com/JBoudou/Itero/mid/db"
	"github.com/JBoudou/Itero/mid/server"
)

type ProfileInfo struct {
	Verified bool
}

func passwdHash() (hash.Hash, error) {
	return blake2b.New256(nil)
}

func LoginHandler(ctx context.Context, response server.Response, request *server.Request) {
	if err := request.CheckPOST(ctx); err != nil {
		response.SendError(ctx, err)
		return
	}

	var loginQuery struct {
		User   string
		Passwd string
	}
	if err := request.UnmarshalJSONBody(&loginQuery); err != nil {
		err = server.WrapError(http.StatusBadRequest, "Wrong request", err)
		response.SendError(ctx, err)
		return
	}

	const (
		qName  = `SELECT Id, Passwd, Verified FROM Users WHERE Name = ?`
		qEmail = `SELECT Id, Passwd, Verified FROM Users WHERE Email = ?`
	)
	query := qName
	if strings.ContainsRune(loginQuery.User, '@') {
		query = qEmail
	}

	var id uint32
	var passwd []byte
	var profileInfo ProfileInfo
	row := db.DB.QueryRowContext(ctx, query, loginQuery.User)
	if err := row.Scan(&id, &passwd, &profileInfo.Verified); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = server.UnauthorizedHttpError("User not found")
		}
		response.SendError(ctx, err)
		return
	}

	hashFct, err := passwdHash()
	if err != nil {
		response.SendError(ctx, err)
		return
	}
	hashFct.Write([]byte(loginQuery.Passwd))
	hashPwd := hashFct.Sum(nil)
	if !bytes.Equal(hashPwd, passwd) {
		response.SendError(ctx, server.UnauthorizedHttpError("Wrong password"))
		return
	}

	response.SendLoginAccepted(ctx, server.User{Name: loginQuery.User, Id: id, Logged: true}, request, profileInfo)
	return
}
