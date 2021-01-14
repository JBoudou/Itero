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
	"net/http"
	"time"

	"github.com/JBoudou/Itero/b64buff"
	"github.com/JBoudou/Itero/events"
	"github.com/JBoudou/Itero/db"
	"github.com/JBoudou/Itero/server"
)

const (
	PollUserTypeSimple = iota
)

type CreateQuery struct {
	UserType         uint8
	Title            string
	Description      string
	Alternatives     []PollAlternative
	MinNbRounds      uint8
	MaxNbRounds      uint8
	Deadline         time.Time
	MaxRoundDuration uint64 // milliseconds
	RoundThreshold   float64
}

func defaultCreateQuery() CreateQuery {
	return CreateQuery{
		UserType: PollUserTypeSimple,
		MinNbRounds: 2,
		MaxNbRounds: 10,
		Deadline: time.Now().Add(7 * 24 * time.Hour),
		MaxRoundDuration: 24 * 3600 * 1000,
		RoundThreshold: 1.,
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
			must(server.NewHttpError(http.StatusForbidden, "Unauthorized", "Unlogged user"))
		}
	}
	must(request.CheckPOST(ctx))

	query := defaultCreateQuery()
	must(request.UnmarshalJSONBody(&query))

	if len(query.Title) < 0 {
		must(server.NewHttpError(http.StatusBadRequest, "Wrong request", "Missing title"))
	}
	if len(query.Alternatives) < 2 {
		must(server.NewHttpError(http.StatusBadRequest, "Wrong request", "Too few alternatives"))
	}

	var err error
	pollSegment := PollSegment{}
	pollSegment.Salt, err = b64buff.RandomUInt32(saltNbBits)
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
			INSERT INTO Polls (Title, Description, Admin, Salt, NbChoices, MinNbRounds, MaxNbRounds,
		                     Deadline, MaxRoundDuration, RoundThreshold)
				  	 VALUE (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
		qAlternative = `INSERT INTO Alternatives (Poll, Id, Name) VALUE (?, ?, ?)`
	)

	result, err := tx.ExecContext(ctx, qPoll,
		query.Title,
		query.Description,
		request.User.Id,
		pollSegment.Salt,
		len(query.Alternatives),
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
	for _, alt := range query.Alternatives {
		_, err := tx.ExecContext(ctx, qAlternative, pollSegment.Id, alt.Id, alt.Name)
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
