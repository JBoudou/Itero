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
	"time"

	"github.com/JBoudou/Itero/b64buff"
	"github.com/JBoudou/Itero/db"
	"github.com/JBoudou/Itero/server"
)

/** PollSegment **/

const saltNbBits = 22

type PollSegment struct {
	Id   uint32
	Salt uint32
}

func PollSegmentDecode(str string) (ret PollSegment, err error) {
	buff := b64buff.Buffer{}
	err = buff.WriteB64(str)
	if err == nil {
		ret.Salt, err = buff.ReadUInt32(saltNbBits)
	}
	if err == nil {
		ret.Id, err = buff.ReadUInt32(32)
	}
	return
}

func (self PollSegment) Encode() (str string, err error) {
	buff := b64buff.Buffer{}
	err = buff.WriteUInt32(self.Salt, saltNbBits)
	if err == nil {
		err = buff.WriteUInt32(self.Id, 32)
	}
	if err == nil {
		str, err = buff.ReadAllB64()
	}
	return
}

/** PollInfo **/

type PollInfo struct {
	Id           uint32
	NbChoices    uint8
	Active       bool
	CurrentRound uint8
	Participate  bool
}

func noPollError(reason string) server.HttpError {
	return server.NewHttpError(http.StatusNotFound, "No poll", reason)
}

// pollSegmentFromRequest retrieves the poll id from the last segment of the URL.
func pollSegmentFromRequest(request *server.Request) (segment PollSegment, err error) {
	remainingLength := len(request.RemainingPath)
	if remainingLength == 0 {
		err = server.NewHttpError(http.StatusBadRequest, "No poll segment", "No poll segment")
		return
	}
  return PollSegmentDecode(request.RemainingPath[remainingLength-1])
}

// checkPollAccess ensure that the user can access the poll.
//
// It checks that the request has a session and a valid poll segment. It also check that the user
// participates in the poll. If she doesn't, poll.Participate is set to false.
func checkPollAccess(ctx context.Context, request *server.Request) (poll PollInfo, err error) {
	// Check user
	// TODO allow unregistered poll
	if request.User == nil {
		if request.SessionError != nil {
			err = server.WrapUnauthorizedError(request.SessionError)
		} else {
			err = server.UnauthorizedHttpError("Unlogged user")
		}
		return
	}

	// Check segment
	var segment PollSegment
	segment, err = pollSegmentFromRequest(request)
	if err != nil {
		return
	}
	poll.Id = segment.Id

	// Check poll
	var salt uint32
	var publicity uint8
	const qPoll = `SELECT Salt, Publicity, NbChoices, State = 'Active', CurrentRound FROM Polls WHERE Id = ?`
	row := db.DB.QueryRowContext(ctx, qPoll, poll.Id)
	err = row.Scan(&salt, &publicity, &poll.NbChoices, &poll.Active, &poll.CurrentRound)
	if err != nil {
		return
	}
	if salt != segment.Salt {
		err = noPollError("Wrong salt")
		return
	}

	// Check participant
	const qParticipate = `SELECT 1 FROM Participants WHERE Poll = ? AND User = ?`
	rows, err := db.DB.QueryContext(ctx, qParticipate, poll.Id, request.User.Id)
	poll.Participate = err != nil || rows.Next()
	rows.Close()
	if poll.Participate {
		return
	}
	if publicity >= db.PollPublicityInvited {
		err = noPollError("Private poll")
		return
	}
	if poll.CurrentRound > 0 {
		err = noPollError("Cannot join a poll after the first round")
		return
	}

	return
}

func (pollInfo PollInfo) BallotType() uint8 {
	// TODO really compute the value
	if !pollInfo.Active {
		return BallotTypeClosed
	}
	return BallotTypeUninominal
}

func (pollInfo PollInfo) InformationType() uint8 {
	// TODO really compute the value
	if pollInfo.CurrentRound == 0 {
		return InformationTypeNoneYet
	}
	return InformationTypeCounts
}

/** PollHandler **/

const (
	BallotTypeClosed = iota
	BallotTypeUninominal
)

const (
	InformationTypeNoneYet = iota
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
	Ballot           uint8
	Information      uint8
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
