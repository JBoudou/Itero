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
	"time"

	"github.com/JBoudou/Itero/mid/db"
	"github.com/JBoudou/Itero/mid/salted"
	"github.com/JBoudou/Itero/mid/server"
)

/** PollInfo **/

// PollInfo holds information on a poll as returned by checkPollAccess.
type PollInfo struct {
	Id           uint32
	NbChoices    uint8
	Active       bool
	CurrentRound uint8
	Public       bool

	Logged      bool
	Participate bool
}

func noPollError(reason string) server.HttpError {
	return server.NewHttpError(http.StatusNotFound, "No poll", reason)
}

// checkPollAccess ensure that the user can access the poll.
//
// It checks that the request has a session and a valid poll segment. It also check that the user
// participates in the poll. If she doesn't, poll.Participate is set to false.
func checkPollAccess(ctx context.Context, request *server.Request) (poll PollInfo, err error) {
	// Check user
	if request.SessionError != nil {
		err = server.WrapUnauthorizedError(request.SessionError)
		return
	}
	poll.Logged = request.User != nil && request.User.Logged

	// Check segment
	var segment salted.Segment
	segment, err = salted.FromRequest(request)
	if err != nil {
		return
	}
	poll.Id = segment.Id

	// Check poll
	var salt uint32
	var electorate db.Electorate
	const qPoll = `SELECT Salt, Electorate, NbChoices, State = 'Active', CurrentRound FROM Polls WHERE Id = ?`
	rows, err := db.DB.QueryContext(ctx, qPoll, poll.Id)
	defer rows.Close()
	if err != nil {
		return
	}
	if !rows.Next() {
		err = noPollError("Id not found")
		return
	}
	err = rows.Scan(&salt, &electorate, &poll.NbChoices, &poll.Active, &poll.CurrentRound)
	if err != nil {
		return
	}
	rows.Close()
	if salt != segment.Salt {
		err = noPollError("Wrong salt")
		return
	}
	poll.Public = electorate == db.ElectorateAll
	if !poll.Logged && !poll.Public {
		err = noPollError("Non-public poll")
		return
	}

	const qUserVerified = `SELECT 1 FROM Users WHERE Id = ? AND Verified`
	if electorate == db.ElectorateVerified {
		var rows *sql.Rows
		rows, err = db.DB.QueryContext(ctx, qUserVerified, request.User.Id)
		defer rows.Close()
		if err != nil {
			return
		}
		if !rows.Next() {
			err = noPollError("Not verified user")
			return
		}
	}

	// Check participant
	if request.User != nil {
		const qParticipate = `SELECT 1 FROM Participants WHERE Poll = ? AND User = ?`
		var rows *sql.Rows
		rows, err = db.DB.QueryContext(ctx, qParticipate, poll.Id, request.User.Id)
		poll.Participate = rows.Next()
		rows.Close()
		if err != nil {
			return
		}
	}
	if !poll.Public && !poll.Participate && poll.CurrentRound > 0 {
		err = noPollError("Cannot join a poll after the first round")
		return
	}

	return
}

func (pollInfo PollInfo) BallotType() BallotType {
	// TODO really compute the value
	if !pollInfo.Active {
		return BallotTypeClosed
	}
	return BallotTypeUninominal
}

func (pollInfo PollInfo) InformationType() InformationType {
	// TODO really compute the value
	if pollInfo.CurrentRound == 0 {
		return InformationTypeNoneYet
	}
	return InformationTypeCounts
}

/** PollHandler **/

type BallotType uint8

const (
	BallotTypeClosed BallotType = iota
	BallotTypeUninominal
)

type InformationType uint8

const (
	InformationTypeNoneYet InformationType = iota
	InformationTypeCounts
)

type PollAnswer struct {
	Title            string
	Description      string
	Admin            string
	CreationTime     time.Time
	CurrentRound     uint8
	Active           bool
	State            string
	CarryForward     bool
	Start            time.Time
	RoundDeadline    time.Time
	PollDeadline     time.Time
	MaxRoundDuration uint64 // in milliseconds
	MinNbRounds      uint8
	MaxNbRounds      uint8
	Ballot           BallotType
	Information      InformationType
}

func PollHandler(ctx context.Context, response server.Response, request *server.Request) {
	pollInfo, err := checkPollAccess(ctx, request)
	must(err)

	answer := PollAnswer{
		Ballot:       pollInfo.BallotType(),
		Information:  pollInfo.InformationType(),
		CurrentRound: pollInfo.CurrentRound,
		Active:       pollInfo.Active,
	}

	// Additional informations for display
	const qSelect = `
	  SELECT p.Title, p.Description, u.Name, p.Created, p.State, p.ReportVote, p.Start,
	         RoundDeadline(p.CurrentRoundStart, p.MaxRoundDuration, p.Deadline, p.CurrentRound, p.MinNbRounds),
	         p.Deadline, TIME_TO_SEC(p.MaxRoundDuration) * 1000, p.MinNbRounds, p.MaxNbRounds
	    FROM Polls AS p, Users AS u
	   WHERE p.Id = ?
	     AND p.Admin = u.Id`
	row := db.DB.QueryRowContext(ctx, qSelect, pollInfo.Id)
	var desc sql.NullString
	var start, roundDeadline, pollDeadline sql.NullTime
	var maxRoundDuration sql.NullInt64
	must(row.Scan(
		&answer.Title, &desc, &answer.Admin, &answer.CreationTime, &answer.State, &answer.CarryForward,
		&start, &roundDeadline, &pollDeadline, &maxRoundDuration,
		&answer.MinNbRounds, &answer.MaxNbRounds))
	if desc.Valid {
		answer.Description = desc.String
	}
	if start.Valid {
		answer.Start = start.Time
	}
	if roundDeadline.Valid {
		answer.RoundDeadline = roundDeadline.Time
	}
	if pollDeadline.Valid {
		answer.PollDeadline = pollDeadline.Time
	}
	if maxRoundDuration.Valid && maxRoundDuration.Int64 > 0 {
		answer.MaxRoundDuration = uint64(maxRoundDuration.Int64)
	}

	response.SendJSON(ctx, answer)
}
