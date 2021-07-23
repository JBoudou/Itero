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
	"net/http"
	"strings"

	"github.com/JBoudou/Itero/mid/db"
	"github.com/JBoudou/Itero/mid/salted"
	"github.com/JBoudou/Itero/mid/server"
)

func ShortURLHandler(ctx context.Context, response server.Response, request *server.Request) {
	key := strings.Join(request.RemainingPath, "/")

	const qFind = `SELECT Id, Salt FROM Polls WHERE ShortURL = ?`
	rows, err := db.DB.QueryContext(ctx, qFind, key)
	must(err)
	defer rows.Close()
	if !rows.Next() {
		panic(server.NewHttpError(http.StatusNotFound, "Not found", "Unknown ShortURL"))
	}
	var segment salted.Segment
	must(rows.Scan(&segment.Id, &segment.Salt))

	encoded, err := segment.Encode()
	must(err)
	response.SendRedirect(ctx, request, "/r/poll/" + encoded)
}
