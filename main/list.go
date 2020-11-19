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

	"github.com/JBoudou/Itero/b64buff"
	"github.com/JBoudou/Itero/db"
	"github.com/JBoudou/Itero/server"
)

// TODO Move PollSegment somewhere else.

type PollSegment struct {
	Id   uint32
	Salt uint32
}

func PollSegmentDecode(str string) (ret PollSegment, err error) {
	buff := b64buff.Buffer{}
	err = buff.WriteB64(str)
	if err == nil {
		ret.Salt, err = buff.ReadUInt32(23)
	}
	if err == nil {
		ret.Id, err = buff.ReadUInt32(31)
	}
	return
}

func (self PollSegment) Encode() (str string, err error) {
	buff := b64buff.Buffer{}
	err = buff.WriteUInt32(self.Salt, 23)
	if err == nil {
		err = buff.WriteUInt32(self.Id, 31)
	}
	if err == nil {
		str, err = buff.ReadAllB64()
	}
	return
}

// end PollSegment

func ListHandler(ctx context.Context, response server.Response, request *server.Request) {
	type entry struct {
		Title        string `json:"t"`
		Segment      string `json:"s"`
		CurrentRound uint8  `json:"c"`
		MaxRound     uint8  `json:"m"`
	}
	reply := struct {
		Ok   bool
		Data []entry `json:"d"`
	}{}

	if request.User == nil {
		// TODO change that
		reply.Ok = false
		response.SendJSON(ctx, reply)
		return
	}

	const query =
		`SELECT p.Id, p.Salt, p.Title, p.CurrentRound, p.MaxNbRounds
		   FROM Polls AS p
		  WHERE p.Active AND p.CurrentRound = 0 AND p.Publicity <= ?
		 UNION DISTINCT
		 SELECT p.Id, p.Salt, p.Title, p.CurrentRound, p.MaxNbRounds
		   FROM Polls AS p, Participants AS a
		  WHERE p.Active
		    AND p.Id = a.Poll
		    AND a.User = ?`
	rows, err := db.DB.QueryContext(ctx, query, db.PollPublicityPublicRegistered, request.User.Id)
	if err != nil {
		response.SendError(err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var entry entry
		var segment PollSegment

		err = rows.Scan(&segment.Id, &segment.Salt, &entry.Title, &entry.CurrentRound, &entry.MaxRound)
		if err != nil {
			response.SendError(err)
			return
		}

		entry.Segment, err = segment.Encode()
		if err != nil {
			response.SendError(err)
			return
		}

		reply.Data = append(reply.Data, entry)
	}

	response.SendJSON(ctx, reply)
	return
}
