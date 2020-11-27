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
	"net/http"
	"regexp"
	"unicode"
	"unicode/utf8"

	"github.com/JBoudou/Itero/db"
	"github.com/JBoudou/Itero/server"
	"github.com/JBoudou/Itero/server/logger"

	"golang.org/x/crypto/blake2b"
)

func passwdHash() (hash.Hash, error) {
	return blake2b.New256(nil)
}

func LoginHandler(ctx context.Context, response server.Response, request *server.Request) {
	var loginInfo struct {
		User   string
		Passwd string
	}
	if err := request.UnmarshalJSONBody(&loginInfo); err != nil {
		logger.Print(ctx, err)
		err = server.NewHttpError(http.StatusBadRequest, "Wrong request", "Unable to read loginInfo")
		response.SendError(ctx, err)
		return
	}

	var id uint32
	var passwd []byte
	row := db.DB.QueryRowContext(ctx, "SELECT Id, Passwd FROM Users WHERE Name = ?", loginInfo.User)
	if err := row.Scan(&id, &passwd); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = server.NewHttpError(http.StatusForbidden, "Unauthorized", "User not found")
		}
		response.SendError(ctx, err)
		return
	}

	hashFct, err := passwdHash()
	if err != nil {
		response.SendError(ctx, err)
		return
	}
	hashFct.Write([]byte(loginInfo.Passwd))
	hashPwd := hashFct.Sum(nil)
	if !bytes.Equal(hashPwd, passwd) {
		response.SendError(ctx, server.NewHttpError(http.StatusForbidden, "Unauthorized", "Wrong password"))
		return
	}

	response.SendLoginAccepted(ctx, server.User{Name: loginInfo.User, Id: id}, request)
	return
}

func SignupHandler(ctx context.Context, response server.Response, request *server.Request) {
	var signupQuery struct {
		Name   string
		Email  string
		Passwd string
	}
	if err := request.UnmarshalJSONBody(&signupQuery); err != nil {
		logger.Print(ctx, err)
		err = server.NewHttpError(http.StatusBadRequest, "Wrong request", "Unable to read SignupQuery")
		response.SendError(ctx, err)
		return
	}

	// Check query //

	if len(signupQuery.Name) < 5 {
		err := server.NewHttpError(http.StatusBadRequest, "Name too short", "User name too short")
		response.SendError(ctx, err)
		return
	}
	firstRune, _ := utf8.DecodeRuneInString(signupQuery.Name)
	lastRune, _ := utf8.DecodeLastRuneInString(signupQuery.Name)
	if unicode.IsSpace(firstRune) || unicode.IsSpace(lastRune) {
		err := server.NewHttpError(http.StatusBadRequest, "Name has spaces",
			"User starts or ends with space")
		response.SendError(ctx, err)
		return
	}

	if len(signupQuery.Passwd) < 5 {
		err := server.NewHttpError(http.StatusBadRequest, "Passwd too short", "Password too short")
		response.SendError(ctx, err)
		return
	}
	hashFct, err := passwdHash()
	if err != nil {
		response.SendError(ctx, err)
		return
	}
	hashFct.Write([]byte(signupQuery.Passwd))
	hashPwd := hashFct.Sum(nil)

	ok, err := regexp.MatchString("^[^\\s@]+@[^\\s.]+\\.\\S+$", signupQuery.Email)
	if err != nil {
		response.SendError(ctx, err)
		return
	}
	if !ok {
		err := server.NewHttpError(http.StatusBadRequest, "Email invalid", "Wrong email format")
		response.SendError(ctx, err)
		return
	}

	// Perform request //

	const qSetAutocommit = `SET autocommit=?`
	const qLock = `LOCK TABLE Users WRITE CONCURRENT`
	const qUnlock = `UNLOCK TABLES`
	const qSelect = `SELECT 1 FROM Users WHERE Name = ? OR Email = ?`
	const qInsert = `INSERT INTO Users (Name, Email, Passwd) VALUE (?, ?, ?)`

	conn, err := db.DB.Conn(ctx)
	if err != nil {
		response.SendError(ctx, err)
		return
	}
	defer conn.Close()
	exec := func(query string, args ...interface{}) (result sql.Result) {
		if err == nil {
			result, err = conn.ExecContext(ctx, query, args...)
		}
		return
	}
	exec(qSetAutocommit, 0)
	exec(qLock)
	if err != nil {
		response.SendError(ctx, err)
		return
	}
	defer func() {
		conn.ExecContext(ctx, qUnlock)
		conn.ExecContext(ctx, qSetAutocommit, 1)
	}()

	rows, err := conn.QueryContext(ctx, qSelect, signupQuery.Name, signupQuery.Email)
	if err != nil {
		response.SendError(ctx, err)
		return
	}
	if rows.Next() {
		rows.Close();
		err = server.NewHttpError(http.StatusBadRequest, "Already exists",
			"Name or Email already exists")
		response.SendError(ctx, err)
		return
	}

	result := exec(qInsert, signupQuery.Name, signupQuery.Email, hashPwd)
	if err != nil {
		response.SendError(ctx, err)
		return
	}
	rawId, err := result.LastInsertId()
	if err != nil {
		response.SendError(ctx, err)
		return
	}

	// Start session //

	response.SendLoginAccepted(ctx, server.User{Name: signupQuery.Name, Id: uint32(rawId)}, request)
	return
}
