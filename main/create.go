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

package main

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/JBoudou/Itero/b64buff"
	"github.com/JBoudou/Itero/db"
	"github.com/JBoudou/Itero/events"
	"github.com/JBoudou/Itero/server"
)

const (
	PollUserTypeSimple = iota
)

type SimpleAlternative struct {
	Name string
	Cost float64
}

type CreateQuery struct {
	UserType         uint8
	Title            string
	Description      string
	Hidden           bool
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
		UserType:         PollUserTypeSimple,
		ReportVote:       true,
		MinNbRounds:      2,
		MaxNbRounds:      10,
		Deadline:         time.Now().Add(7 * 24 * time.Hour),
		MaxRoundDuration: 24 * 3600 * 1000,
		RoundThreshold:   1.,
	}
}

type CreatePollEvent struct {
	Poll uint32
}

func CreateHandler(ctx context.Context, response server.Response, request *server.Request) {
	if request.User == nil {
		if request.SessionError != nil {
			must(request.SessionError)
		} else {
			panic(server.UnauthorizedHttpError("Unlogged user"))
		}
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

	var err error
	pollSegment := PollSegment{}
	pollSegment.Salt, err = b64buff.RandomUInt32(saltNbBits)
	must(err)
	publicity := db.PollPublicityPublicRegistered
	if query.Hidden {
		publicity = db.PollPublicityHiddenRegistered
	}

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
			INSERT INTO Polls (Title, Description, Admin, State, Start, Salt, Publicity, NbChoices,
												 ReportVote, MinNbRounds, MaxNbRounds, Deadline, MaxRoundDuration,
												 RoundThreshold)
				  	 VALUE (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
		qAlternative = `INSERT INTO Alternatives (Poll, Id, Name) VALUE (?, ?, ?)`
	)

	result, err := tx.ExecContext(ctx, qPoll,
		query.Title,
		query.Description,
		request.User.Id,
		state,
		start,
		pollSegment.Salt,
		publicity,
		len(query.Alternatives),
		query.ReportVote,
		query.MinNbRounds,
		query.MaxNbRounds,
		query.Deadline,
		db.MillisecondsToTime(query.MaxRoundDuration),
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
	events.Send(CreatePollEvent{pollSegment.Id})
}
