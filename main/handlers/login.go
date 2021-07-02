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
	"net/http"
	"strings"

	"github.com/JBoudou/Itero/mid/db"
	"github.com/JBoudou/Itero/mid/root"
	"github.com/JBoudou/Itero/mid/server"
)

type ProfileInfo struct {
	Verified bool
}

type userInfo struct {
	Id       uint32
	Passwd   []byte
	Verified bool
}

func getUserInfo(ctx context.Context, login string) (info userInfo, err error) {
	const (
		qName  = `SELECT Id, Passwd, Verified FROM Users WHERE Name = ?`
		qEmail = `SELECT Id, Passwd, Verified FROM Users WHERE Email = ?`
	)
	query := qName
	if strings.ContainsRune(login, '@') {
		query = qEmail
	}

	row := db.DB.QueryRowContext(ctx, query, login)
	err = row.Scan(&info.Id, &info.Passwd, &info.Verified)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		err = server.UnauthorizedHttpError("User not found")
	}
	return
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

	userInfo, err := getUserInfo(ctx, loginQuery.User)
	must(err)

	hashFct, err := root.PasswdHash()
	if err != nil {
		response.SendError(ctx, err)
		return
	}
	hashFct.Write([]byte(loginQuery.Passwd))
	hashPwd := hashFct.Sum(nil)
	if !bytes.Equal(hashPwd, userInfo.Passwd) {
		response.SendError(ctx, server.UnauthorizedHttpError("Wrong password"))
		return
	}

	response.SendLoginAccepted(ctx, server.User{Name: loginQuery.User, Id: userInfo.Id, Logged: true},
		request, ProfileInfo{Verified: userInfo.Verified})
	return
}
