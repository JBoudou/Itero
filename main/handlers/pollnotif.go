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

package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/JBoudou/Itero/main/services"
	"github.com/JBoudou/Itero/mid/db"
	"github.com/JBoudou/Itero/mid/salted"
	"github.com/JBoudou/Itero/mid/server"
)

//
// PollNotifHandler
//

type PollNotifQuery struct {
	LastUpdate time.Time
}

type PollNotifAnswerEntry struct {
	Timestamp time.Time
	Segment   string
	Title     string
	Round     uint8
	Action    uint8
}

type pollNotifHandler struct {
	notifChannel services.PollNotifChannel
}

func PollNotifHandler(notifChannel services.PollNotifChannel) *pollNotifHandler {
	return &pollNotifHandler{
		notifChannel: notifChannel,
	}
}

func (self *pollNotifHandler) Handle(ctx context.Context, response server.Response, request *server.Request) {
	if request.User == nil {
		if request.SessionError != nil {
			must(request.SessionError)
		} else {
			panic(server.UnauthorizedHttpError("Unlogged user"))
		}
	}
	must(request.CheckPOST(ctx))

	var query PollNotifQuery
	err := request.UnmarshalJSONBody(&query)
	if err != nil {
		panic(server.WrapError(http.StatusBadRequest, "Bad request", err))
	}

	baseList := <-self.notifChannel
	fmt.Printf("Handler received %v.\n", baseList)
	if len(baseList) == 0 {
		response.SendJSON(ctx, make([]PollNotifAnswerEntry, 0))
		return
	}

	const qCheck = `
	  SELECT Title, Salt
		  FROM Polls
		 WHERE Id = ?
		   AND (Admin = %[1]d OR Id IN ( SELECT Poll FROM Participants WHERE User = %[1]d ))`
	stmt, err := db.DB.PrepareContext(ctx, fmt.Sprintf(qCheck, request.User.Id))
	must(err)
	defer stmt.Close()

	answer := make([]PollNotifAnswerEntry, 0, len(baseList)/2)
	for _, notif := range baseList {
		if notif.Timestamp.Before(query.LastUpdate) {
			continue
		}

		entry := PollNotifAnswerEntry{
			Timestamp: notif.Timestamp,
			Round:     notif.Round,
			Action:    notif.Action,
		}

		if notif.Participants != nil {
			if member, ok := notif.Participants[request.User.Id]; !ok || !member {
				continue
			}
			entry.Title = notif.Title

		} else {
			rows, err := stmt.QueryContext(ctx, notif.Id)
			must(err)
			if !rows.Next() {
				continue
			}
			segment := salted.Segment{Id: notif.Id}
			err = rows.Scan(&entry.Title, &segment.Salt)
			rows.Close()
			must(err)

			entry.Segment, err = segment.Encode()
			must(err)
		}

		answer = append(answer, entry)
	}

	response.SendJSON(ctx, answer)
}
