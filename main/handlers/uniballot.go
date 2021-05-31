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
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/JBoudou/Itero/mid/db"
	"github.com/JBoudou/Itero/mid/server"
)

type PollAlternative struct {
	Id   uint8
	Name string
	Cost float64
}

// allAlternatives retrieves all the alternatives for a poll and store them in out.
// Error in allAlternatives are wrapped in server.HttpError and sent by panic.
func allAlternatives(ctx context.Context, poll PollInfo, out *[]PollAlternative) {
	const qSelect = `SELECT Id, Name, Cost FROM Alternatives WHERE Poll = ? ORDER BY Id ASC`
	rows, err := db.DB.QueryContext(ctx, qSelect, poll.Id)
	must(err)
	*out = make([]PollAlternative, poll.NbChoices)
	for i := 0; rows.Next(); i++ {
		must(rows.Scan(&(*out)[i].Id, &(*out)[i].Name, &(*out)[i].Cost))
	}
}

type uninominalBallot struct {
	Value uint8
	State uint8
}

const (
	uninominalBallotStateUndefined = iota
	uninominalBallotStateBlank
	uninominalBallotStateValid
)

// Scan implements database/sql.Scanner.
func (self *uninominalBallot) Scan(src interface{}) error {
	switch v := src.(type) {
	case int64:
		self.Value = uint8(v)
		self.State = uninominalBallotStateValid
	case nil:
		self.State = uninominalBallotStateBlank
	case []byte:
		conv, err := strconv.Atoi(string(v))
		if err != nil {
			return err
		}
		self.Value = uint8(conv)
		self.State = uninominalBallotStateValid
	default:
		return errors.New(fmt.Sprintf("Type of %v incompatible with uninominalBallot", src))
	}
	return nil
}


// UninominalBallotAnswer represents the response sent by UninominalBallotHandler.
// The fields Previous and Current are not sent in the JSON representation if the user did not vote.
// If the user abstained, these fields are replaced with fields PreviousIsBlank or CurrentIsBlank
// with the boolean value true.
type UninominalBallotAnswer struct {
	Previous     uninominalBallot
	Current      uninominalBallot
	Alternatives []PollAlternative
}

// MarshalJSON implements json.Marshaler.
func (self *UninominalBallotAnswer) MarshalJSON() (ret []byte, err error) {
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
	encodeBallot := func(name string, ballot uninominalBallot) {
		switch ballot.State {
		case uninominalBallotStateUndefined:
			return
		case uninominalBallotStateBlank:
			encodeHeader(name + "IsBlank")
			if err == nil {
				err = encoder.Encode(true)
			}
		case uninominalBallotStateValid:
			encodeHeader(name)
			if err == nil {
				err = encoder.Encode(ballot.Value)
			}
		}

		if err == nil {
			_, err = buffer.WriteRune(',')
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

// UninominalBallotAnswer sends the previous ballot (if any), the current one (if any) and all
// the alternatives.
func UninominalBallotHandler(ctx context.Context, response server.Response, request *server.Request) {
	pollInfo, err := checkPollAccess(ctx, request)
	must(err)

	const qGetBallots = `
		SELECT p.Round, b.Alternative
		  FROM Participants AS p
			LEFT OUTER JOIN Ballots AS b ON (p.User, p.Poll, p.Round) = (b.User, b.Poll, b.Round)
		 WHERE p.User = ? AND p.Poll = ? AND p.Round IN (?, ?)
		 ORDER BY p.Round`
	var answer UninominalBallotAnswer

	if request.User == nil {
		var user server.User
		user, err = UnloggedFromAddr(ctx, request.RemoteAddr())
		must(err)
		request.User = &user
	}

	var previousRound uint8
	if pollInfo.CurrentRound > 0 {
		// Round is unsigned
		previousRound = pollInfo.CurrentRound - 1
	}
	rows, err := db.DB.QueryContext(ctx, qGetBallots,
		request.User.Id, pollInfo.Id, previousRound, pollInfo.CurrentRound)
	must(err)
	for rows.Next() {
		var round uint8
		var alternative uninominalBallot
		must(rows.Scan(&round, &alternative))
		setBallot := func(field *uninominalBallot) {
			if field.State != uninominalBallotStateUndefined {
				must(errors.New("Duplicated ballot"))
			}
			*field = alternative
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
	response.SendJSON(ctx, &answer)
}
