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

package handlers

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/JBoudou/Itero/main/services"
	"github.com/JBoudou/Itero/mid/db"
	"github.com/JBoudou/Itero/mid/salted"
	"github.com/JBoudou/Itero/mid/server"
	"github.com/JBoudou/Itero/pkg/events"
)

type CreatePollElectorate int8

const (
	CreatePollElectorateAll CreatePollElectorate = iota - 1
	CreatePollElectorateLogged
	CreatePollElectorateVerified
)

func (self CreatePollElectorate) ToDB() db.Electorate {
	switch self {
	case CreatePollElectorateAll:
		return db.ElectorateAll
	case CreatePollElectorateVerified:
		return db.ElectorateVerified
	default:
		return db.ElectorateLogged
	}
}

type SimpleAlternative struct {
	Name string
	Cost float64
}

type CreateQuery struct {
	Title            string
	Description      string
	Hidden           bool
	Electorate       CreatePollElectorate
	Start            time.Time
	Alternatives     []SimpleAlternative
	ReportVote       bool
	MinNbRounds      uint8
	MaxNbRounds      uint8
	Deadline         time.Time
	MaxRoundDuration uint64 // milliseconds
	RoundThreshold   float64
}

func defaultCreateQuery() CreateQuery {
	return CreateQuery{
		ReportVote:       true,
		MinNbRounds:      2,
		MaxNbRounds:      10,
		Deadline:         time.Now().Add(7 * 24 * time.Hour),
		MaxRoundDuration: 24 * 3600 * 1000,
		RoundThreshold:   1.,
	}
}

type createHandler struct {
	evtManager events.Manager
}

func CreateHandler(evtManager events.Manager) createHandler {
	return createHandler{evtManager: evtManager}
}

func (self createHandler) Handle(ctx context.Context, response server.Response, request *server.Request) {
	if request.User == nil || !request.User.Logged {
		must(request.SessionError)
		panic(server.UnauthorizedHttpError("Unlogged user"))
	}
	must(request.CheckPOST(ctx))

	query := defaultCreateQuery()
	must(request.UnmarshalJSONBody(&query))

	if len(query.Title) < 0 {
		must(server.NewHttpError(http.StatusBadRequest, "Bad request", "Missing title"))
	}
	if len(query.Alternatives) < 2 {
		must(server.NewHttpError(http.StatusBadRequest, "Bad request", "Too few alternatives"))
	}

	var start sql.NullTime
	var state string
	if query.Start.After(time.Now()) {
		start.Time = query.Start
		start.Valid = true
		state = "Waiting"
	} else {
		if !query.Start.IsZero() {
			must(server.NewHttpError(http.StatusBadRequest, "Bad request", "Start must be after now"))
		}
		state = "Active"
	}

	electorate := query.Electorate.ToDB()
	const qVerified = `SELECT 1 FROM Users WHERE Id = ? AND Verified`
	if electorate == db.ElectorateVerified {
		rows, err := db.DB.QueryContext(ctx, qVerified, request.User.Id)
		must(err)
		defer rows.Close()
		if !rows.Next() {
			panic(server.NewHttpError(http.StatusBadRequest, "Not verified", "The user is not verified"))
		}
	}

	pollSegment, err := salted.New(0)
	must(err)

	tx, err := db.DB.BeginTx(ctx, nil)
	must(err)
	commited := false
	defer func() {
		if !commited {
			tx.Rollback()
		}
	}()

	const (
		qPoll = `
			INSERT INTO Polls (Title, Description, Admin, State, Start, Salt, Electorate, Hidden,
			                   NbChoices, ReportVote, MinNbRounds, MaxNbRounds, Deadline, MaxRoundDuration,
												 RoundThreshold)
				  	 VALUE (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
		qAlternative = `INSERT INTO Alternatives (Poll, Id, Name) VALUE (?, ?, ?)`
	)

	result, err := tx.ExecContext(ctx, qPoll,
		query.Title,
		query.Description,
		request.User.Id,
		state,
		start,
		pollSegment.Salt,
		electorate,
		query.Hidden,
		len(query.Alternatives),
		query.ReportVote,
		query.MinNbRounds,
		query.MaxNbRounds,
		query.Deadline,
		db.DurationToTime(time.Duration(query.MaxRoundDuration)*time.Millisecond),
		query.RoundThreshold,
	)
	must(err)
	tmp, err := result.LastInsertId()
	must(err)
	pollSegment.Id = uint32(tmp)
	for id, alt := range query.Alternatives {
		_, err = tx.ExecContext(ctx, qAlternative, pollSegment.Id, id, alt.Name)
		must(err)
	}

	err = tx.Commit()
	commited = true
	must(err)

	segment, err := pollSegment.Encode()
	must(err)
	response.SendJSON(ctx, segment)
	self.evtManager.Send(services.CreatePollEvent{pollSegment.Id})
}
