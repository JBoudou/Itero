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

package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/JBoudou/Itero/b64buff"
	"github.com/JBoudou/Itero/server/logger"

	gs "github.com/gorilla/sessions"
)

type Response interface {
	// SendJSON sends a JSON as response.
	// On success statuc code is http.StatusOK.
	SendJSON(ctx context.Context, data interface{})

	// SendError sends an error as response.
	// If the error is an HttpError, its code and msg are used in the HTPP response.
	// Also log the error.
	SendError(context.Context, error)
	
	// SendLoginAccepted create new credential for the user and send it as response.
	SendLoginAccepted(context.Context, User, *Request)
}

type response struct {
	writer http.ResponseWriter
}

func (self response) SendJSON(ctx context.Context, data interface{}) {
	if err := ctx.Err(); err != nil {
		self.SendError(ctx, err)
		return
	}
	buff, err := json.Marshal(data)
	if err != nil {
		self.SendError(ctx, err)
		return
	}
	self.writer.Header().Add("content-type", "application/JSON")
	if _, err = self.writer.Write(buff); err != nil {
		logger.Printf(ctx, "Write error: %v", err)
	}
}

func (self response) SendError(ctx context.Context, err error) {
	send := func(statusCode int, msg string) {
		http.Error(self.writer, msg, statusCode)
		logger.Printf(ctx, "%d %s: %s", statusCode, msg, err)
	}

	var pError HttpError
	if errors.As(err, &pError) {
		send(pError.Code, pError.Msg)
	} else if errors.Is(err, context.Canceled) {
		send(http.StatusInternalServerError, "Canceled")
	} else if errors.Is(err, context.DeadlineExceeded) {
		send(http.StatusGatewayTimeout, "Timed out")
	} else {
		send(http.StatusInternalServerError, "Internal error")
	}
}

func (self response) SendLoginAccepted(ctx context.Context, user User, req *Request) {
	if err := ctx.Err(); err != nil {
		self.SendError(ctx, err)
		return
	}

	sessionId, err := MakeSessionId()
	if err != nil {
		self.SendError(ctx, err)
	}
	session := NewSession(sessionStore, sessionStore.Options, sessionId, user)
	if err = session.Save(req.original, self.writer); err != nil {
		logger.Printf(ctx, "Error saving session: %v", err)
	}

	self.SendJSON(ctx, sessionId)
}

// MakeSessionId create a new session id.
//
// This is a low level function, made available for tests.
func MakeSessionId() (string, error) {
	return b64buff.RandomString(4)
}

// NewSession creates a new session for the given user.
//
// This is a low level function, made available for tests. Use SendLoginAccepted instead.
func NewSession(st gs.Store, opts *gs.Options, sessionId string, user User) (session *gs.Session) {
	session = gs.NewSession(st, sessionName)
	sessionOptions := *opts
	session.Options = &sessionOptions
	session.IsNew = true

	session.Values[sessionKeySessionId] = sessionId
	session.Values[sessionKeyUserName] = user.Name
	session.Values[sessionKeyUserId] = user.Id
	session.Values[sessionKeyDeadline] = time.Now().Unix() + sessionMaxAge + sessionGraceTime

	return
}
