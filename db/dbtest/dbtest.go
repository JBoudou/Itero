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

// Package dbtest allows to temporarily add test data to the database.
package dbtest

import (
	"database/sql"

	"github.com/JBoudou/Itero/db"
)

// Env provides methods to add temporary test data. It collects functions to remove these data.
// These functions are called by Close, hence a call to Close must be defered for each Env object.
//
// Env's methods never return errors. Instead, the Error field is updated and checked by all
// methods (except Close).
type Env struct {
	toClose []func()
	Error   error
}

// Defer adds a function to be called by Close. As other Env's method, nothing is done if Error is
// not nil.
func (self *Env) Defer(fct func()) {
	if self.Error == nil {
		self.toClose = append(self.toClose, fct)
	}
}

// Close calls all collected functions, in reverse order, even if Error is not nil.
func (self *Env) Close() {
	for i := len(self.toClose) - 1; i >= 0; i-- {
		self.toClose[i]()
	}
}

// CreateUser adds a user to the database. The user has name ' Test ' (mind the spaces), email
// address 'test@example.test', and password 'XYZ'. It is deleted by Close.
func (self *Env) CreateUser() (userId uint32) {
	if self.Error != nil {
		return
	}
	const query = `INSERT INTO Users(Name, Email, Passwd)
	   VALUES(' Test ', 'test@example.test',
	   X'2e43477a2da06cb4aba764381086cbc9323945eb1bffb232f221e374af44f803')`
	var result sql.Result
	result, self.Error = db.DB.Exec(query)
	userId = self.extractId(result)

	self.Defer(func() {
		db.DB.Exec(`DELETE FROM Users WHERE Id = ?`, userId)
	})
	return
}

// CreatePoll adds a poll to the database. The poll has Salt 42, MaxNbRounds 3, and 2 alternatives
// 'No' and 'Yes' (in that order). The poll is deleted by Close.
func (self *Env) CreatePoll(title string, admin uint32, publicity uint8) (pollId uint32) {
	const (
		qCreatePoll = `
			INSERT INTO Polls(Title, Admin, Salt, NbChoices, Publicity, MaxNbRounds)
			VALUE (?, ?, 42, 2, ?, 3)`
		qCreateAlternatives = `
			INSERT INTO Alternatives(Poll, Id, Name) VALUES (?, 0, 'No'), (?, 1, 'Yes')`
		qRemovePoll = `
			DELETE FROM Polls WHERE Id = ?`
	)

	if self.Error != nil {
		return
	}

	var tx *sql.Tx
	tx, self.Error = db.DB.Begin()
	result := self.execTx(tx, qCreatePoll, title, admin, publicity)
	pollId = self.extractId(result)
	self.execTx(tx, qCreateAlternatives, pollId, pollId)
	self.closeTx(tx)

	self.Defer(func() {
		db.DB.Exec(qRemovePoll, pollId)
	})
	return
}

func (self *Env) execTx(tx *sql.Tx, query string, args... interface{}) (ret sql.Result) {
	if self.Error != nil {
		return
	}
	ret, self.Error = tx.Exec(query, args...)
	return
}

func (self *Env) closeTx(tx *sql.Tx) {
	if self.Error != nil {
		tx.Rollback()
	} else {
		self.Error = tx.Commit()
	}
}

func (self *Env) extractId(result sql.Result) uint32 {
	if self.Error != nil {
		return 0
	}
	var tmp int64
	tmp, self.Error = result.LastInsertId()
	return uint32(tmp)
}
