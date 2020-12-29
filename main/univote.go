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
	"net/http"

	"github.com/JBoudou/Itero/db"
	"github.com/JBoudou/Itero/events"
	"github.com/JBoudou/Itero/server"
	"github.com/JBoudou/Itero/server/logger"
)

type VoteEvent struct {
	Poll uint32
}

type UninomialVoteQuery struct {
	Blank       bool `json:",omitempty"`
	Alternative uint8
}

func UninomialVoteHandler(ctx context.Context, response server.Response, request *server.Request) {
	// Verifications
	if err := request.CheckPOST(ctx); err != nil {
		response.SendError(ctx, err)
		return
	}
	pollInfo, err := checkPollAccess(ctx, request)
	must(err)
	if pollInfo.BallotType() != BallotTypeUninomial {
		err = server.NewHttpError(http.StatusBadRequest, "Wrong poll", "Poll is not uninomial")
		response.SendError(ctx, err)
		return
	}

	// Get query
	var voteQuery UninomialVoteQuery
	if err := request.UnmarshalJSONBody(&voteQuery); err != nil {
		logger.Print(ctx, err)
		err = server.NewHttpError(http.StatusBadRequest, "Wrong request",
			"Unable to read UninomialVoteQuery")
		response.SendError(ctx, err)
		return
	}

	const (
		qDeleteBallot      = `DELETE FROM Ballots WHERE User = ? AND Poll = ? AND Round = ?`
		qInsertBallot      = `INSERT INTO Ballots (User, Poll, Alternative, Round) VALUE (?, ?, ?, ?)`
		qInsertParticipant = `INSERT INTO Participants (User, Poll, LastRound) VALUE (?, ?, ?)`
		qUpdateParticipant = `UPDATE Participants SET LastRound = ? WHERE User = ? AND Poll = ?`
	)

	tx, err := db.DB.BeginTx(ctx, nil)
	must(err)
	commited := false
	defer func() {
		if !commited {
			tx.Rollback()
		}
	}()

	if pollInfo.Participate {
		_, err = tx.ExecContext(ctx, qDeleteBallot, request.User.Id, pollInfo.Id, pollInfo.CurrentRound)
	} else {
		_, err = tx.ExecContext(ctx, qInsertParticipant, request.User.Id, pollInfo.Id,
			pollInfo.CurrentRound)
	}
	must(err)
	if !voteQuery.Blank {
		_, err = tx.ExecContext(ctx, qInsertBallot, request.User.Id, pollInfo.Id, voteQuery.Alternative,
			pollInfo.CurrentRound)
		must(err)
	}
	if pollInfo.Participate {
		_, err = tx.ExecContext(ctx, qUpdateParticipant, pollInfo.CurrentRound, request.User.Id,
			pollInfo.Id)
		must(err)
	}

	must(tx.Commit())
	commited = true
	response.SendJSON(ctx, "Ok")
	events.Send(VoteEvent{pollInfo.Id})
}