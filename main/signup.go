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
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/go-sql-driver/mysql"

	"github.com/JBoudou/Itero/db"
	"github.com/JBoudou/Itero/server"
	"github.com/JBoudou/Itero/server/logger"
)

func SignupHandler(ctx context.Context, response server.Response, request *server.Request) {
	if err := request.CheckPOST(ctx); err != nil {
		response.SendError(ctx, err)
		return
	}

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
	if strings.ContainsRune(signupQuery.Name, '@') {
		err := server.NewHttpError(http.StatusBadRequest, "Name has at sign",
			"User contains the at sign rune")
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

	ok, err := regexp.MatchString("^[^\\s@]+@[^\\s.]+\\.\\S\\S+$", signupQuery.Email)
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

	const qInsert = `INSERT INTO Users (Name, Email, Passwd) VALUE (?, ?, ?)`

	result, err := db.DB.ExecContext(ctx, qInsert, signupQuery.Name, signupQuery.Email, hashPwd)
	if err != nil {
		sqlError, ok := err.(*mysql.MySQLError)
		if ok && sqlError.Number == 1062 {
			err = server.NewHttpError(http.StatusBadRequest, "Already exists",
				"The Name or Password already exists")
		}
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
