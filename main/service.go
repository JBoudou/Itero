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
	"database/sql"
)

func collectUI32Id(set map[uint32]bool, tx *sql.Tx, query string, queryArgs ...interface{}) error {
	rows, err := tx.Query(query, queryArgs...)
	if err != nil {
		return err
	}
	for rows.Next() {
		var key uint32
		if err := rows.Scan(&key); err != nil {
			return nil
		}
		set[key] = true
	}
	return nil
}

func execOnUI32Id(set map[uint32]bool, tx *sql.Tx, query string) error {
	stmt, err := tx.Prepare(query)
	if err != nil {
		return err
	}
	for id := range set {
		if _, err := stmt.Exec(id); err != nil {
			return err
		}
	}
	stmt.Close()
	return nil
}
