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

// Package db provides access to the database backend of the application.
//
// In particular, the package reads the parameters of the database given in the configuration file
// and provides the pool of database connections to the application.
package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"

	"github.com/JBoudou/Itero/mid/root"
	"github.com/JBoudou/Itero/pkg/config"
	"github.com/JBoudou/Itero/pkg/slog"
)

// DB is the pool of database connections for the application.
var DB *sql.DB

// Ok indicates whether the database is usable. May be false if there is no configuration for the
// package.
var Ok bool

var (
	PollTypeAcceptanceSet uint8

	PollRulePlurality uint8

	RoundTypeFreelyAsynchronous uint8
)

// Electorate is the enum type for the field Electorate of table Polls.
type Electorate string

const (
	ElectorateAll      Electorate = "All"
	ElectorateLogged   Electorate = "Logged"
	ElectorateVerified Electorate = "Verified"
)

var (
	NotFound = errors.New("Not found")
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
	var logger slog.Leveled
	if err := root.IoC.Inject(&logger); err != nil {
		panic(err)
	}

	// Read conf
	cfg := myConfig{
		MaxIdleConns: 2,
		MaxIdleTime:  "2m",
	}
	if err := config.Value("database", &cfg); err != nil {
		logger.Error(err)
		logger.Error("Package db not usable because there is no configuration for it.")
		Ok = false
		return
	}
	Ok = true

	// Add DSN options
	cfg.DSN = AddURLQuery(cfg.DSN, dsnOptions)

	// Open DB
	var err error
	DB, err = sql.Open("mysql", cfg.DSN)
	mustm(logger, err, "Error initializing database:")
	mustm(logger, DB.Ping(), "Error connecting the the database:")

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
	fillVars(logger, "PollType", map[string]*uint8{"Acceptance Set": &PollTypeAcceptanceSet})
	fillVars(logger, "PollRule", map[string]*uint8{"Plurality": &PollRulePlurality})
	fillVars(logger, "RoundType", map[string]*uint8{"Freely Asynchronous": &RoundTypeFreelyAsynchronous})
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

func fillVars(logger slog.Leveled, table string, assoc map[string]*uint8) {
	rows, err := DB.Query("SELECT Id, Label FROM " + table)
	mustm(logger, err, "Query on "+table+":")

	for rows.Next() {
		var id uint8
		var label string
		mustm(logger, rows.Scan(&id, &label), "Parsing error:")

		ptr, ok := assoc[label]
		if !ok {
			logger.Errorf("Unknown label \"%s\" in table %s", label, table)
			panic(nil)
		}
		*ptr = id
		delete(assoc, label)
	}

	if len(assoc) != 0 {
		joined := ""
		for label := range assoc {
			joined += " \"" + label + "\""
		}
		logger.Errorf("Labels not found in %s:%s", table, joined)
		panic(nil)
	}
}

// DurationToTime convert a duration into a time understandable by the database.
func DurationToTime(duration time.Duration) string {
	milli := duration.Milliseconds()
	prefix := ""
	if milli < 0 {
		milli = -milli
		prefix = "-"
	}
	return prefix + fmt.Sprintf("%d:%02d:%02d.%03d000",
		milli/(60*60*1000),
		(milli/(60*1000))%60,
		(milli/1000)%60,
		milli%1000,
	)
}

// IdFromResult extracts an identifier from a query result of type sql.Result.
// Such identifiers are used as primary keys for Polls, Users and other tables.
func IdFromResult(result sql.Result) (uint32, error) {
	id, err := result.LastInsertId()
	return uint32(id), err
}

// RepeatDeadlocked repeats a transaction as long as it produce mySQL deadlocks.
// MySQL deadlocks are detected by a panic of a mysql.MySQLError with Number 1213.
func RepeatDeadlocked(logger slog.Logger, ctx context.Context, opts *sql.TxOptions, fct func(tx *sql.Tx)) {
	must := func(err error) {
		if err != nil {
			panic(err)
		}
	}
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
				logger.Log("SQL error 1213. Restarting transaction.")
			}()

			fct(tx)

			must(tx.Commit())
			commited = true
		}()
	}
}

func mustm(logger slog.Leveled, err error, msg string) {
	if err != nil {
		logger.Error(msg)
		panic(err)
	}
}
