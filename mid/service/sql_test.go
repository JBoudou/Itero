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

package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JBoudou/Itero/mid/db"
	"github.com/JBoudou/Itero/mid/db/dbtest"
)

func mustt(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func TestIteratorFromRows(t *testing.T) {
	tests := []struct {
		name  string
		table []time.Time
	}{
		{
			name: "Zero",
		},
		{
			name: "One",
			table: []time.Time{
				time.Date(2001, 1, 1, 1, 1, 1, 0, time.Local),
			},
		},
		{
			name: "Two",
			table: []time.Time{
				time.Date(2001, 2, 1, 1, 1, 1, 0, time.Local),
				time.Date(2002, 2, 2, 2, 2, 2, 0, time.Local),
			},
		},
	}

	const (
		qCreateTable = `
		  CREATE OR REPLACE TEMPORARY TABLE TestIteratorFromRows (
				id int unsigned,
				da timestamp
			)`
		qInsert = `INSERT INTO TestIteratorFromRows(id, da) VALUE (?,?)`
		qSelect = `SELECT id, da FROM TestIteratorFromRows ORDER BY id ASC`
	)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			conn, err := db.DB.Conn(ctx)
			mustt(t, err)
			defer conn.Close()

			_, err = conn.ExecContext(ctx, qCreateTable)
			mustt(t, err)
			for id, date := range tt.table {
				_, err = conn.ExecContext(ctx, qInsert, id, date)
				mustt(t, err)
			}

			rows, err := conn.QueryContext(ctx, qSelect)
			mustt(t, err)
			it := IteratorFromRows(rows)
			defer it.Close()

			for expect_id, expect_date := range tt.table {
				if !it.Next() {
					t.Fatalf("Prematurely ends at row %d", expect_id)
				}
				got_id, got_date := it.IdAndDate()
				if got_id != uint32(expect_id) {
					t.Errorf("Expect id %d. Got %d.", expect_id, got_id)
				}
				if got_date != expect_date {
					t.Errorf("Step %d. Expect date %v. Got %v.", expect_id, expect_date, got_date)
				}
			}
		})
	}
}

func TestSQLProcessOne(t *testing.T) {
	env := dbtest.Env{}
	defer env.Close()
	userId := env.CreateUserWith(t.Name())
	env.Must(t)

	tests := []struct {
		name           string
		query          string
		id             uint32
		nothingToDoYet bool
	}{
		{
			name:           "NothingToDoYet",
			query:          `UPDATE Users SET Name="mu" WHERE 0=?`,
			id:             42,
			nothingToDoYet: true,
		},
		{
			name:  "Send event",
			query: `UPDATE Users SET Name="mu" WHERE Id=?`,
			id:    userId,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			err := SQLProcessOne(tt.query, tt.id)

			got_nothingToDoYet := false
			if errors.Is(err, NothingToDoYet) {
				got_nothingToDoYet = true
			} else {
				mustt(t, err)
			}

			if got_nothingToDoYet != tt.nothingToDoYet {
				t.Errorf("Wrong nothingToDoYet. Expect %t. Got %t.", tt.nothingToDoYet, got_nothingToDoYet)
			}
		})
	}
}
