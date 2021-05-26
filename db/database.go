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
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"

	"github.com/JBoudou/Itero/config"
)

// Database pool for the application.
var DB *sql.DB

// Whether the database is usable. May be false if there is no configuration for the package.
var Ok bool

// Constants for Polls fields.
var (
	PollTypeAcceptanceSet uint8

	PollPublicityPublic           uint8
	PollPublicityPublicRegistered uint8
	PollPublicityHidden           uint8
	PollPublicityHiddenRegistered uint8
	PollPublicityInvited          uint8

	PollRulePlurality uint8

	RoundTypeFreelyAsynchronous uint8
)

const dsnOptions = `parseTime=true&` +
	`sql_mode=%27TRADITIONAL%2CNO_ENGINE_SUBSTITUTION%2CONLY_FULL_GROUP_BY%27`

type myConfig struct {
	DSN string

	MaxLifetime  string
	MaxIdleTime  string
	MaxIdleConns int
	MaxOpenConns int
}

func init() {
	// Read conf
	cfg := myConfig{
		MaxIdleConns: 2,
		MaxIdleTime:  "2m",
	}
	if err := config.Value("database", &cfg); err != nil {
		log.Print(err)
		log.Println("WARNING: Package db not usable because there is no configuration for it.")
		Ok = false
		return
	}
	Ok = true

	// Add DSN options
	cfg.DSN = AddURLQuery(cfg.DSN, dsnOptions)

	// Open DB
	var err error
	DB, err = sql.Open("mysql", cfg.DSN)
	mustm(err, "Error initializing database:")
	mustm(DB.Ping(), "Error connecting the the database:")

	// configure DB
	if dur, err := time.ParseDuration(cfg.MaxLifetime); err == nil {
		DB.SetConnMaxLifetime(dur)
	}
	if dur, err := time.ParseDuration(cfg.MaxIdleTime); err == nil {
		DB.SetConnMaxIdleTime(dur)
	}
	DB.SetMaxIdleConns(cfg.MaxIdleConns)
	DB.SetMaxOpenConns(cfg.MaxOpenConns)

	// Fill variables
	fillVars("PollType", map[string]*uint8{"Acceptance Set": &PollTypeAcceptanceSet})
	fillVars("PollPublicity", map[string]*uint8{
		"Public":            &PollPublicityPublic,
		"Public Registered": &PollPublicityPublicRegistered,
		"Hidden":            &PollPublicityHidden,
		"Hidden Registered": &PollPublicityHiddenRegistered,
		"Invited":           &PollPublicityInvited})
	fillVars("PollRule", map[string]*uint8{"Plurality": &PollRulePlurality})
	fillVars("RoundType", map[string]*uint8{"Freely Asynchronous": &RoundTypeFreelyAsynchronous})
}

// AddURLQuery adds a query string to an url string.
// TODO: Move this function somewhere else.
func AddURLQuery(url, query string) string {
	sep := "?"
	if strings.Contains(url, sep) {
		sep = "&"
	}
	return url + sep + query
}

func fillVars(table string, assoc map[string]*uint8) {
	rows, err := DB.Query("SELECT Id, Label FROM " + table)
	mustm(err, "Query on "+table+":")

	for rows.Next() {
		var id uint8
		var label string
		mustm(rows.Scan(&id, &label), "Parsing error:")

		ptr, ok := assoc[label]
		if !ok {
			log.Fatalf("Unknown label \"%s\" in table %s", label, table)
		}
		*ptr = id
		delete(assoc, label)
	}

	if len(assoc) != 0 {
		joined := ""
		for label := range assoc {
			joined += " \"" + label + "\""
		}
		log.Fatalf("Labels not found in %s:%s", table, joined)
	}
}

// MillisecondsToTime convert a time in millisecond into a time understandable by the database.
func MillisecondsToTime(milli uint64) string {
	return fmt.Sprintf("%d:%02d:%02d.%03d000",
		milli/(60*60*1000),
		(milli/(60*1000))%60,
		(milli/1000)%60,
		milli%1000,
	)
}

// RepeatDeadlocked repeats a transaction as long as it produce mySQL deadlocks.
// MySQL deadlocks are detected by a panic of a mysql.MySQLError with Number 1213.
func RepeatDeadlocked(ctx context.Context, opts *sql.TxOptions, fct func(tx *sql.Tx)) {
	must := func (err error) { if err != nil { panic(err) } }
	repeat := true
	for repeat {
		func() {
			tx, err := DB.BeginTx(ctx, opts)
			must(err)
			commited := false
			
			defer func() {
				if !commited {
					tx.Rollback()
				}
				exc := recover()
				if exc == nil {
					repeat = false
					return
				}
				err, ok := exc.(error)
				var mySqlError *mysql.MySQLError
				if !ok || !errors.As(err, &mySqlError) || mySqlError.Number != 1213 {
					panic(exc)
				}
				log.Println("SQL error 1213. Restarting transaction.")
			}()

			fct(tx)

			must(tx.Commit())
			commited = true
		}()
	}
}

func mustm(err error, msg string) {
	if err != nil {
		log.Fatal(msg, " ", err)
	}
}
