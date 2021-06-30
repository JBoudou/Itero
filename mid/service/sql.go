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
	"database/sql"
	"time"

	"github.com/JBoudou/Itero/mid/db"
)

// IteratorFromRows constructs an Iterator from an *sql.Rows.
// Each rows must have exacly two cells: one that can be scanned as a uint32 and one that can be
// scanned as a time.Time.
func IteratorFromRows(rows *sql.Rows) Iterator {
	return &rowsIterator{rows: rows, err: rows.Err()}
}

// SQLProcessOne is a helper function to implement Service.ProcessOne.
// The given query is executed with id as parameter.
// If the query succeed but no row is affected, NothingToDoYet is returned.
func SQLProcessOne(query string, id uint32) error {
	result, err := db.DB.Exec(query, id)
	if err != nil {
		return err
	}
	
	var affected int64
	affected, err = result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return NothingToDoYet
	}
	return nil
}

// SQLCheckAll is a helper function to implement Service.CheckAll.
// It executes the query and return an iterator from the returned rows.
// The query must return a list of task, each task consisting in an id and a date.
// See IteratorFromRows for details.
func SQLCheckAll(query string) Iterator {
	rows, err := db.DB.Query(query)
	if err != nil {
		return errorIdDateIterator{err}
	} else {
		return IteratorFromRows(rows)
	}
}

//
// Implementation
//

type rowsIterator struct {
	rows *sql.Rows
	err  error
	id   uint32
	date time.Time
}

func (self *rowsIterator) Next() bool {
	if !self.rows.Next() {
		self.err = self.rows.Err()
		return false
	}
	self.err = self.rows.Scan(&self.id, &self.date)
	return self.err == nil
}

func (self *rowsIterator) IdAndDate() (uint32, time.Time) {
	return self.id, self.date
}

func (self *rowsIterator) Err() error {
	return self.err
}

func (self *rowsIterator) Close() error {
	return self.rows.Close()
}


type errorIdDateIterator struct {
	err error
}

func (self errorIdDateIterator) Next() bool {
	return false
}

func (self errorIdDateIterator) IdAndDate() (uint32, time.Time) {
	return 0, time.Time{}
}

func (self errorIdDateIterator) Err() error {
	return self.err
}

func (self errorIdDateIterator) Close() error {
	return nil
}
