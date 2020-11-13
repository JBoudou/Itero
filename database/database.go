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

package database

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"

	"github.com/JBoudou/Itero/config"
)

var DB *sql.DB

type myConfig struct {
	DSN string
}

func init() {
	var cfg myConfig
	err := config.Value("database", &cfg)
	if err != nil {
		log.Fatalf("Error loading database configuration: %s", err.Error())
	}

	DB, err = sql.Open("mysql", cfg.DSN)
	if err != nil {
		log.Fatalf("Error initializing database: %s", err.Error())
	}
}
