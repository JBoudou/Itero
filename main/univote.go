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
	"context"
	"database/sql"
	"net/http"

	"github.com/JBoudou/Itero/db"
	"github.com/JBoudou/Itero/events"
	"github.com/JBoudou/Itero/server"
	"github.com/JBoudou/Itero/server/logger"
)

type VoteEvent struct {
	Poll uint32
}

type UninominalVoteQuery struct {
	Blank       bool `json:",omitempty"`
	Alternative uint8
}

func UninominalVoteHandler(ctx context.Context, response server.Response, request *server.Request) {
	// Verifications
	if err := request.CheckPOST(ctx); err != nil {
		response.SendError(ctx, err)
		return
	}
	pollInfo, err := checkPollAccess(ctx, request)
	must(err)
	if pollInfo.BallotType() != BallotTypeUninominal {
		err = server.NewHttpError(http.StatusBadRequest, "Wrong poll", "Poll is not uninominal")
		response.SendError(ctx, err)
		return
	}

	// Get query
	var voteQuery UninominalVoteQuery
	if err := request.UnmarshalJSONBody(&voteQuery); err != nil {
		logger.Print(ctx, err)
		err = server.NewHttpError(http.StatusBadRequest, "Wrong request",
			"Unable to read UninominalVoteQuery")
		response.SendError(ctx, err)
		return
	}

	const (
		qDeleteBallot      = `DELETE FROM Ballots WHERE User = ? AND Poll = ? AND Round = ?`
		qInsertBallot      = `INSERT INTO Ballots (User, Poll, Alternative, Round) VALUE (?, ?, ?, ?)`
		qLastRound         = `SELECT 1 FROM Participants WHERE User = ? AND Poll = ? AND Round = ?`
		qInsertParticipant = `INSERT INTO Participants (User, Poll, Round) VALUE (?, ?, ?)`
	)

	tx, err := db.DB.BeginTx(ctx, nil)
	must(err)
	commited := false
	defer func() {
		if !commited {
			tx.Rollback()
		}
	}()

	var result sql.Result
	result, err = tx.ExecContext(ctx, qDeleteBallot, request.User.Id, pollInfo.Id, pollInfo.CurrentRound)
	must(err)

	if affected, err := result.RowsAffected(); err == nil && affected == 0 {
		var rows *sql.Rows
		rows, err = tx.QueryContext(ctx, qLastRound, request.User.Id, pollInfo.Id, pollInfo.CurrentRound)
		must(err)
		if !rows.Next() {
			_, err = tx.ExecContext(ctx, qInsertParticipant, request.User.Id, pollInfo.Id,
				pollInfo.CurrentRound)
			must(err);
		}
		must(rows.Close())
	}

	if !voteQuery.Blank {
		_, err = tx.ExecContext(ctx, qInsertBallot, request.User.Id, pollInfo.Id, voteQuery.Alternative,
			pollInfo.CurrentRound)
		must(err)
	}

	must(tx.Commit())
	commited = true
	response.SendJSON(ctx, "Ok")
	events.Send(VoteEvent{pollInfo.Id})
}
