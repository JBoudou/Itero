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

package main

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"hash"
	"log"
	"net/http"

	"github.com/JBoudou/Itero/db"
	"github.com/JBoudou/Itero/server"

	"golang.org/x/crypto/blake2b"
)

func passwdHash() (hash.Hash, error) {
	return blake2b.New256(nil)
}

func LoginHandler(ctx context.Context, response server.Response, request *server.Request) {
	var loginInfo struct {
		User string
		Passwd string
	}
	if err := request.UnmarshalJSONBody(&loginInfo); err != nil {
		// TODO: better login
		log.Print(err)
		err = server.NewHttpError(http.StatusBadRequest, "Wrong request", "Unable to read loginInfo")
		response.SendError(err)
		return
	}

	var id uint32
	var passwd []byte
	row := db.DB.QueryRowContext(ctx, "SELECT Id, Passwd FROM Users WHERE Name = ?", loginInfo.User)
	if err := row.Scan(&id, &passwd); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = server.NewHttpError(http.StatusForbidden, "Unauthorized", "User not found")
		}
		response.SendError(err)
		return
	}

	hashFct, err := passwdHash()
	if err != nil {
		response.SendError(err)
		return
	}
	hashFct.Write([]byte(loginInfo.Passwd))
	hashPwd := hashFct.Sum(nil)
	if !bytes.Equal(hashPwd, passwd) {
		response.SendError(server.NewHttpError(http.StatusForbidden, "Unauthorized", "Wrong password"))
		return
	}

	response.SendLoginAccepted(ctx, server.User{Name: loginInfo.User, Id: id}, request)
	return
}
