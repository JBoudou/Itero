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
	"bytes"
	"context"
	"encoding/json"
	"errors"

	"github.com/JBoudou/Itero/db"
	"github.com/JBoudou/Itero/server"
)

type PollAlternative struct {
	Id   uint8
	Name string
	Cost float64
}

func allAlternatives(ctx context.Context, poll PollInfo, out *[]PollAlternative) {
	// TODO: Allocate the slice here and no copy
	const qSelect = `SELECT Id, Name, Cost FROM Alternatives WHERE Poll = ? ORDER BY Id ASC`
	rows, err := db.DB.QueryContext(ctx, qSelect, poll.Id)
	must(err)
	for rows.Next() {
		var entry PollAlternative
		must(rows.Scan(&entry.Id, &entry.Name, &entry.Cost))
		*out = append(*out, entry)
	}
}

type NullUInt8 struct {
	Value uint8
	Valid bool
}

func (self *NullUInt8) Set(value uint8) {
	self.Value, self.Valid = value, true
}

type UninomialBallotAnswer struct {
	Previous     NullUInt8
	Current      NullUInt8
	Alternatives []PollAlternative
}

func (self *UninomialBallotAnswer) MarshalJSON() (ret []byte, err error) {
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)

	_, err = buffer.WriteRune('{')
	if err != nil {
		return
	}

	encodeHeader := func(name string) {
		if err == nil {
			err = encoder.Encode(name)
		}
		if err == nil {
			_, err = buffer.WriteRune(':')
		}
	}
	encodeBallot := func(name string, ballot NullUInt8) {
		if err == nil && ballot.Valid {
			encodeHeader(name)
			if err == nil {
				err = encoder.Encode(ballot.Value)
			}
			if err == nil {
				_, err = buffer.WriteRune(',')
			}
		}
	}
	encodeBallot("Previous", self.Previous)
	encodeBallot("Current", self.Current)

	encodeHeader("Alternatives")
	if err != nil {
		return
	}
	err = encoder.Encode(self.Alternatives)
	if err != nil {
		return
	}

	_, err = buffer.WriteRune('}')
	return buffer.Bytes(), err
}

func UninomialBallotHandler(ctx context.Context, response server.Response, request *server.Request) {
	pollInfo, err := checkPollAccess(ctx, request)
	must(err)

	const qGetBallots = `
		SELECT Round, Alternative FROM Ballots
		WHERE User = ? AND Poll = ? AND Round IN (?, ?)
		ORDER BY Round`
	var answer UninomialBallotAnswer

	var previousRound uint8
	if pollInfo.CurrentRound > 0 {
		previousRound = pollInfo.CurrentRound - 1
	}
	rows, err := db.DB.QueryContext(ctx, qGetBallots,
		request.User.Id, pollInfo.Id, previousRound, pollInfo.CurrentRound)
	must(err)
	for rows.Next() {
		var round, alternative uint8
		must(rows.Scan(&round, &alternative))
		setBallot := func(field *NullUInt8) {
			if field.Valid {
				must(errors.New("Duplicated ballot"))
			}
			field.Set(alternative)
		}
		switch round {
		case pollInfo.CurrentRound:
			setBallot(&answer.Current)
			break
		case previousRound:
			setBallot(&answer.Previous)
			break
		default:
			must(errors.New("Impossible round"))
		}
	}

	allAlternatives(ctx, pollInfo, &answer.Alternatives)
	response.SendJSON(ctx, answer)
}
