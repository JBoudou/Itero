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

package db

import (
	"context"
	"testing"
	"time"
)

func TestCreateConfirmation(t *testing.T) {
	t.Parallel()

	const (
		qInsertUser       = `INSERT INTO Users (Name, Email, Passwd) VALUE (?,?,?)`
		qDeleteUser       = `DELETE FROM Users WHERE Id = ?`
		qFindConfirmation = `SELECT Type, Expires, Id, Salt FROM Confirmations WHERE User = ?`
	)

	// User
	result, err := DB.Exec(qInsertUser, t.Name(), t.Name()+"@example.com", "123456")
	mustt(t, err)
	uid, err := IdFromResult(result)
	defer func() { DB.Exec(qDeleteUser, uid) }()
	mustt(t, err)

	// Check there is initially no confirmation
	rows, err := DB.Query(qFindConfirmation, uid)
	defer rows.Close()
	mustt(t, err)
	if rows.Next() {
		t.Fatalf("There is a confirmation before none has been created")
	}

	// Create a confirmation
	const (
		type_    = ConfirmationTypeVerify
		duration = time.Hour
	)
	now := time.Now()
	segment, err := CreateConfirmation(context.Background(), uid, type_, duration)
	mustt(t, err)

	// Check it
	var gotType string
	var gotExpires time.Time
	var gotId, gotSalt uint32
	rows, err = DB.Query(qFindConfirmation, uid)
	mustt(t, err)
	if !rows.Next() {
		t.Fatalf("No confirmation found.")
	}
	mustt(t, rows.Scan(&gotType, &gotExpires, &gotId, &gotSalt))
	if gotType != type_ {
		t.Errorf("Wrong type. Got %s. Expect %s.", gotType, type_)
	}
	if gotId != segment.Id {
		t.Errorf("Wrong id. Got %d. Expect %d.", gotId, segment.Id)
	}
	if gotSalt != segment.Salt {
		t.Errorf("Wrong salt. Got %d. Expect %d.", gotSalt, segment.Salt)
	}
	gotDuration := gotExpires.Sub(now)
	diffDuration := gotDuration - duration
	if diffDuration > time.Second || diffDuration < -1*time.Second {
		t.Errorf("Wrong duration. Got %v. Expect %v.", gotDuration, duration)
	}
}
