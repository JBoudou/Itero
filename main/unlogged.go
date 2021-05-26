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
	"database/sql"
	"encoding/binary"
	"errors"
	"hash/fnv"
	"strconv"
	"strings"
	"unicode"

	"github.com/JBoudou/Itero/db"
	"github.com/JBoudou/Itero/server"

	"github.com/go-sql-driver/mysql"
)

// UnloggedFromHash constructs a server.User from a hash, creating the user if necessary.
func UnloggedFromHash(ctx context.Context, hash uint32) (user server.User, err error) {
	const (
		qSelect = `SELECT Id FROM Users WHERE Hash = ?`
		qInsert = `INSERT INTO Users (Hash) VALUE (?)`
	)

	if hash > 0xFFFFFF {
		return user, errors.New("Hash overflow")
	}
	bin := LE24Bits(hash)

	var rows *sql.Rows
	_, err = db.DB.ExecContext(ctx, qInsert, bin)
	if err != nil {
		var mySQLerror *mysql.MySQLError
		if !errors.As(err, &mySQLerror) || mySQLerror.Number != 1062 {
			return
		}
	}

	rows, err = db.DB.QueryContext(ctx, qSelect, bin)
	if err != nil {
		return
	}
	if !rows.Next() {
		return user, errors.New("Query error")
	}
	err = rows.Scan(&user.Id)
	if err != nil {
		return
	}
	rows.Close()

	user.Hash = hash
	user.Logged = false
	return
}

// LE24Bits converts a 24 bits unsigned integer to its Little-Endian representation.
func LE24Bits(input uint32) []byte {
	return []byte{byte(input), byte(input >> 8), byte(input >> 16)}
}

// HashAddr produces a 24 bits hash sum from address strings like http.Request.RemoteAddr.
// The sum is guaranteed to be the same from the same IP.
// The fonction is only pertinent for IPv4 addresses, but should never fail.
func HashAddr(addr string) uint32 {
	pos := strings.LastIndexByte(addr, ':')
	if pos >= 0 {
		addr = string(addr[:pos])
	}
	
	hash := fnv.New32()
	nonDigit := func(c rune) bool { return !unicode.IsNumber(c) }
	for _, str := range(strings.FieldsFunc(addr, nonDigit)) {
		val, err := strconv.Atoi(str)
		if err != nil {
			panic(err)
		}
		binary.Write(hash, binary.LittleEndian, val)
	}
	
	sum := hash.Sum32()
	return (sum >> 24) ^ (sum & 0xFFFFFF)
}

func UnloggedFromAddr(ctx context.Context, addr string) (server.User, error) {
	return UnloggedFromHash(ctx, HashAddr(addr))
}
