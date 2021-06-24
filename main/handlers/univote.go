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

package handlers

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/JBoudou/Itero/mid/db"
	"github.com/JBoudou/Itero/pkg/events"
	"github.com/JBoudou/Itero/mid/server"
	"github.com/JBoudou/Itero/main/services"
)

type UninominalVoteQuery struct {
	Blank       bool `json:",omitempty"`
	Alternative uint8
	Round       uint8
}

type uninominalVoteHandler struct {
	evtManager events.Manager
}

func UninominalVoteHandler(evtManager events.Manager) uninominalVoteHandler {
	return uninominalVoteHandler{evtManager: evtManager}
}

func (self uninominalVoteHandler) Handle(ctx context.Context, response server.Response, request *server.Request) {

	// Verifications
	if err := request.CheckPOST(ctx); err != nil {
		response.SendError(ctx, err)
		return
	}
	pollInfo, err := checkPollAccess(ctx, request)
	must(err)
	if !pollInfo.Active {
		err = server.NewHttpError(http.StatusLocked, "Inactive poll", "Poll is currently not active")
		response.SendError(ctx, err)
		return
	}
	if pollInfo.BallotType() != BallotTypeUninominal {
		err = server.NewHttpError(http.StatusBadRequest, "Wrong poll", "Poll is not uninominal")
		response.SendError(ctx, err)
		return
	}

	// Unlogged
	sendUnloggedCookie := request.User == nil
	if sendUnloggedCookie {
		var user server.User
		user, err = UnloggedFromAddr(ctx, request.RemoteAddr())
		must(err)
		request.User = &user
	}

	// Get query
	var voteQuery UninominalVoteQuery
	if err := request.UnmarshalJSONBody(&voteQuery); err != nil {
		err = server.WrapError(http.StatusBadRequest, "Wrong request", err)
		response.SendError(ctx, err)
		return
	}

	// Check round before DB operations.
	// We should check after the DB operations, but it is more difficul and the difference is
	// insignificant.
	if voteQuery.Round != pollInfo.CurrentRound {
		if voteQuery.Round+1 == pollInfo.CurrentRound {
			err = server.NewHttpError(http.StatusLocked, "Next round",
				"Round may have changed while the user voted")
		} else {
			err = server.NewHttpError(http.StatusBadRequest, "Wrong round",
				"Round is neither current nor previous")
		}
		response.SendError(ctx, err)
		return
	}

	const (
		qDeleteBallot      = `DELETE FROM Ballots WHERE User = ? AND Poll = ? AND Round = ?`
		qInsertBallot      = `INSERT INTO Ballots (User, Poll, Alternative, Round) VALUE (?, ?, ?, ?)`
		qLastRound         = `SELECT 1 FROM Participants WHERE User = ? AND Poll = ? AND Round = ?`
		qInsertParticipant = `INSERT INTO Participants (User, Poll, Round) VALUE (?, ?, ?)`
	)

	db.RepeatDeadlocked(ctx, nil, func(tx *sql.Tx) {
		var result sql.Result
		result, err = tx.ExecContext(ctx, qDeleteBallot, request.User.Id, pollInfo.Id, pollInfo.CurrentRound)
		must(err)

		// Insert a row in Participants if needed
		if affected, err := result.RowsAffected(); err == nil && affected == 0 {
			var rows *sql.Rows
			rows, err = tx.QueryContext(ctx, qLastRound, request.User.Id, pollInfo.Id, pollInfo.CurrentRound)
			must(err)
			if !rows.Next() {
				_, err = tx.ExecContext(ctx, qInsertParticipant, request.User.Id, pollInfo.Id,
					pollInfo.CurrentRound)
				must(err)
			}
			must(rows.Close())
		}

		// Add the ballot
		if !voteQuery.Blank {
			_, err = tx.ExecContext(ctx, qInsertBallot, request.User.Id, pollInfo.Id, voteQuery.Alternative,
				pollInfo.CurrentRound)
			must(err)
		}
	})

	if sendUnloggedCookie {
		response.SendUnloggedId(ctx, *request.User, request)
	}
	response.SendJSON(ctx, "Ok")
	self.evtManager.Send(services.VoteEvent{pollInfo.Id})
}
