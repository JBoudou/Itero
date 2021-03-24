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

// TODO: Rename this file into something like sqlservice.go.

package main

import (
	"database/sql"
	"time"

	"github.com/JBoudou/Itero/db"
	"github.com/JBoudou/Itero/events"
)

type RowsIdDateIterator struct {
	rows *sql.Rows
	err  error
	id   uint32
	date time.Time
}

func NewRowsIdDateIterator(rows *sql.Rows) *RowsIdDateIterator {
	return &RowsIdDateIterator{rows: rows, err: rows.Err()}
}

func (self *RowsIdDateIterator) Next() bool {
	if !self.rows.Next() {
		self.err = self.rows.Err()
		return false
	}
	self.err = self.rows.Scan(&self.id, &self.date)
	return self.err == nil
}

func (self *RowsIdDateIterator) IdAndDate() (uint32, time.Time) {
	return self.id, self.date
}

func (self *RowsIdDateIterator) Err() error {
	return self.err
}

func (self *RowsIdDateIterator) Close() error {
	return self.rows.Close()
}


type ErrorIdDateIterator struct {
	err error
}

func (self ErrorIdDateIterator) Next() bool {
	return false
}

func (self ErrorIdDateIterator) IdAndDate() (uint32, time.Time) {
	return 0, time.Time{}
}

func (self ErrorIdDateIterator) Err() error {
	return self.err
}

func (self ErrorIdDateIterator) Close() error {
	return nil
}

func SqlServiceProcessOne(query string, id uint32, evt events.Event) error {
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

	return events.Send(evt)
}

func SqlServiceCheckAll(query string) IdAndDateIterator {
	rows, err := db.DB.Query(query)
	if err != nil {
		return ErrorIdDateIterator{err}
	} else {
		return NewRowsIdDateIterator(rows)
	}
}
