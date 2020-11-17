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

package db

import (
	"strings"
	"testing"
)

func noe(err error, t *testing.T) {
	if err != nil {
		t.Helper()
		t.Error(err)
	}
}

func nof(err error, t *testing.T) {
	if err != nil {
		t.Helper()
		t.Fatal(err)
	}
}

func TestConnection(t *testing.T) {
	t.Run("ping", func(t *testing.T) {
		nof(DB.Ping(), t)
	})

	t.Run("sql_mode", func(t *testing.T) {

		// The boolean indicate whether the tag is necessary
		tagSet := map[string]bool{
			"ONLY_FULL_GROUP_BY":         true,
			"STRICT_TRANS_TABLES":        false,
			"STRICT_ALL_TABLES":          true,
			"NO_ZERO_IN_DATE":            true,
			"NO_ZERO_DATE":               true,
			"ERROR_FOR_DIVISION_BY_ZERO": true,
			"TRADITIONAL":                false,
			"NO_AUTO_CREATE_USER":        false,
			"NO_ENGINE_SUBSTITUTION":     true}

		row := DB.QueryRow("SELECT @@sql_mode")
		var result string
		nof(row.Scan(&result), t)

		for _, tag := range strings.Split(result, ",") {
			if len(tag) == 0 {
				continue
			}
			if _, ok := tagSet[tag]; !ok {
				t.Errorf("Unexpected tag %s", tag)
			}
			delete(tagSet, tag)
		}

		for tag, needed := range tagSet {
			if needed {
				t.Errorf("Missing tag %s", tag)
			}
		}
	})
}

func TestInsert(t *testing.T) {
	tx, err := DB.Begin()
	nof(err, t)
	defer func() { nof(tx.Rollback(), t) }()

	execLIR := func(cmd string, args ...interface{}) (ret int64) {
		t.Helper()
		result, err := tx.Exec(cmd, args...)
		nof(err, t)
		ret, err = result.LastInsertId()
		nof(err, t)
		return
	}

	exec := func(cmd string, args ...interface{}) {
		t.Helper()
		_, err := tx.Exec(cmd, args...)
		nof(err, t)
	}

	// MariaDB does not handle $1 placeholders...
	userId := execLIR("INSERT INTO Users (Email, Name, Passwd) VALUES ('jogo@matabio.net', 'joseph', 'XXXXXXXXXXXX')")
	pollId := execLIR("INSERT INTO Polls (Title, Admin, NbChoices) VALUES ('Test', ?, 2)", userId)
	exec("INSERT INTO Alternatives (Poll, Id, Name) VALUES (?, 0, 'Blue'), (?, 1, 'Yellow')", pollId, pollId)
	exec("INSERT INTO Participants (User, Poll) VALUES (?, ?)", userId, pollId)
	exec("INSERT INTO Ballots (User, Poll, Alternative, Round) VALUES (?, ?, 0, 0)", userId, pollId)
	exec("UPDATE Polls SET CurrentRound = 1 WHERE Id = ?", pollId)
	exec("INSERT INTO Ballots (User, Poll, Alternative, Round) VALUES (?, ?, 1, 1)", userId, pollId)
}

func TestVariables(t *testing.T) {
	allDifferent := func(vals ...uint8) {
		t.Helper()
		set := map[uint8]bool{}
		for _, val := range vals {
			if _, found := set[val]; found {
				t.Fatalf("Value %d duplicated", val)
			}
			set[val] = true
		}
	}

	allDifferent(PollPublicityPublic, PollPublicityPublicRegistered, PollPublicityHidden,
		PollPublicityHiddenRegistered, PollPublicityInvited)
}
