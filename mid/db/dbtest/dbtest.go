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
	"log"
	"testing"

	"github.com/JBoudou/Itero/mid/db"
	"github.com/JBoudou/Itero/mid/root"
)

const (
	UserPasswd = "XYZ"

	PollSalt = 42
	PollMaxNbRounds = 4

	// ImpossibleUserName is a user name which is guaranteed to not exist.
	ImpossibleUserName = "  "
)

var UserPasswdHash []byte

func init() {
	hashFct, err := root.PasswdHash()
	if err != nil {
		panic(err)
	}
	hashFct.Write([]byte(UserPasswd))
	UserPasswdHash = hashFct.Sum(nil)
}	

// Env provides methods to add temporary test data. It collects functions to remove these data.
// These functions are called by Close, hence a call to Close must be defered for each Env object.
//
// Env's methods never return errors. Instead, the Error field is updated and checked by all
// methods (except Close).
type Env struct {
	toClose []func()
	Error   error
}

// Defer adds a function to be called by Close. As other Env's method, nothing is added if Error is
// not nil.
func (self *Env) Defer(fct func()) {
	if self.Error == nil {
		self.toClose = append(self.toClose, fct)
	}
}

func (self *Env) logExec(query string, args ...interface{}) {
	_, err := db.DB.Exec(query, args...)
	if err != nil {
		log.Printf("dbtest.Close error: %v", err)
	}
}

// Close calls all collected functions, in reverse order, even if Error is not nil.
func (self *Env) Close() {
	for i := len(self.toClose) - 1; i >= 0; i-- {
		self.toClose[i]()
	}
}

// Must makes the test fail is some error have happened.
func (self *Env) Must(t *testing.T) {
	if self.Error != nil {
		t.Helper()
		t.Fatal(self.Error)
	}
}

// QuietExec executes the query without returning anything.
// Like most other Env's methods, the query is not executed if an error previously occured.
func (self *Env) QuietExec(query string, args ...interface{}) {
	if self.Error != nil {
		return
	}
	_, self.Error = db.DB.Exec(query, args...)
}

// CreateUser adds a user to the database. The user has name ' Test ' (mind the spaces), email
// address 'testTest@example.test', and password UserPasswd. It is deleted by Close.
func (self *Env) CreateUser() uint32 {
	return self.CreateUserWith("Test")
}

// CreateUser adds a user to the database. The user has name as returned by UserNameWith, email
// address as returned by UserEmailWith, and password is UserPasswd. It is deleted by Close.
func (self *Env) CreateUserWith(salt string) (userId uint32) {
	if self.Error != nil {
		return
	}

	const query = `
	   INSERT INTO Users(Name, Email, Passwd)
	   VALUES(?, ?, ?)`
	var result sql.Result
	result, self.Error = db.DB.Exec(query, UserNameWith(salt), UserEmailWith(salt), UserPasswdHash)
	userId = self.extractId(result)

	self.Defer(func() {
		self.logExec(`DELETE FROM Participants WHERE User = ?`, userId)
		self.logExec(`DELETE FROM Users WHERE Id = ?`, userId)
		self.logExec(`ALTER TABLE Users AUTO_INCREMENT = 1`)
	})
	return
}

// UserNameWith returns the name of the user created by CreateUserWith with the same salt.
func UserNameWith(salt string) string {
	const maxNameLen = 62
	if len(salt) > maxNameLen {
		salt = salt[len(salt)-maxNameLen:]
	}
	return " " + salt + " "
}

// UserNameWith returns the email address of the user created by CreateUserWith with the same salt.
func UserEmailWith(salt string) string {
	const (
		prefix     = "test"
		suffix     = "@example.test"
		maxNameLen = 128 - (len(prefix) + len(suffix))
	)
	if len(salt) > maxNameLen {
		salt = salt[len(salt)-maxNameLen:]
	}
	return prefix + salt + suffix
}

// CreatePoll adds a poll to the database. The poll has Salt PollSalt, MaxNbRounds PollMaxNbRounds, and 2 alternatives
// 'No' and 'Yes' (in that order). The poll is deleted by Close.
func (self *Env) CreatePoll(title string, admin uint32, electorate db.Electorate) uint32 {
	return self.CreatePollWith(title, admin, electorate, []string{"No", "Yes"})
}

// CreatePoll adds a poll to the database. The poll has Salt 42, MaxNbRounds 4, and the alternatives
// given as arguments. All alternatives have Cost 1. The poll is deleted by Close.
func (self *Env) CreatePollWith(title string, admin uint32, electorate db.Electorate,
	alternatives []string) (pollId uint32) {

	const (
		qCreatePoll = `
			INSERT INTO Polls(Title, Admin, Salt, NbChoices, Electorate, MaxNbRounds)
			VALUE (?, ?, ?, ?, ?, ?)`
		qCreateAlternative = `
			INSERT INTO Alternatives(Poll, Id, Name) VALUE (?, ?, ?)`
		qRemovePoll = `
			DELETE FROM Polls WHERE Id = ?`
	)

	if self.Error != nil {
		return
	}

	var tx *sql.Tx
	tx, self.Error = db.DB.Begin()
	result := self.execTx(tx, qCreatePoll,
		title, admin, PollSalt, len(alternatives), electorate, PollMaxNbRounds)
	pollId = self.extractId(result)
	altStmt := self.prepareTx(tx, qCreateAlternative)
	for i, alt := range alternatives {
		self.execStmt(altStmt, pollId, i, alt)
	}
	self.closeTx(tx)

	self.Defer(func() {
		self.logExec(qRemovePoll, pollId)
		self.logExec(`ALTER TABLE Polls AUTO_INCREMENT = 1`)
	})
	return
}

// NextRound advances a poll to the next round.
func (self *Env) NextRound(pollId uint32) {
	const qNext = `UPDATE Polls SET CurrentRound = CurrentRound + 1 WHERE Id = ?`
	if self.Error == nil {
		_, self.Error = db.DB.Exec(qNext, pollId)
	}
	return
}

// Vote submits a ballot. If the user does not participate yet in the poll, it is added to the
// participant. No other check is done.
func (self *Env) Vote(pollId uint32, round uint8, userId uint32, alternative uint8) {
	const (
		qCheckParticipant = `SELECT 1 FROM Participants WHERE Poll = ? AND User = ? AND Round = ?`
		qAddParticipant   = `INSERT INTO Participants (Poll, User, Round) VALUE (?, ?, ?)`
		qVote             = `INSERT INTO Ballots (Poll, Round, User, Alternative) VALUE (?, ?, ?, ?)`
	)

	// Create transaction
	var tx *sql.Tx
	if self.Error == nil {
		tx, self.Error = db.DB.Begin()
	}
	if self.Error != nil {
		return
	}
	defer self.closeTx(tx)

	// Ensure the user participate in the poll
	var rows *sql.Rows
	rows, self.Error = tx.Query(qCheckParticipant, pollId, userId, round)
	if self.Error != nil {
		return
	}
	if !rows.Next() {
		_, self.Error = tx.Exec(qAddParticipant, pollId, userId, round)
	} else {
		self.Error = rows.Close()
	}

	// Vote
	self.execTx(tx, qVote, pollId, round, userId, alternative)
}

//
// WithDB
//

// WithDB is a servertest.Test mixin providing a field of type Env.
type WithDB struct {
	DB Env
}

// Close closes the Env field, launching all defered functions.
func (self *WithDB) Close() {
	self.DB.Close()
}

//
// implementation
//

func (self *Env) execTx(tx *sql.Tx, query string, args ...interface{}) (ret sql.Result) {
	if self.Error != nil {
		return
	}
	ret, self.Error = tx.Exec(query, args...)
	return
}

func (self *Env) prepareTx(tx *sql.Tx, query string) (ret *sql.Stmt) {
	if self.Error != nil {
		return
	}
	ret, self.Error = tx.Prepare(query)
	return
}

func (self *Env) closeTx(tx *sql.Tx) {
	if self.Error != nil {
		tx.Rollback()
	} else {
		self.Error = tx.Commit()
	}
}

func (self *Env) execStmt(stmt *sql.Stmt, args ...interface{}) (ret sql.Result) {
	if self.Error != nil {
		return
	}
	ret, self.Error = stmt.Exec(args...)
	return
}

func (self *Env) extractId(result sql.Result) uint32 {
	if self.Error != nil {
		return 0
	}
	var tmp int64
	tmp, self.Error = result.LastInsertId()
	return uint32(tmp)
}
