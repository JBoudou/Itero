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
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	gs "github.com/gorilla/sessions"
)

type Response struct {
	writer http.ResponseWriter
}

// SendJSON sends a JSON as response.
// On success statuc code is http.StatusOK.
func (self Response) SendJSON(ctx context.Context, data interface{}) {
	if err := ctx.Err(); err != nil {
		self.SendError(err)
		return
	}
	buff, err := json.Marshal(data)
	if err != nil {
		self.SendError(err)
		return
	}
	if _, err = self.writer.Write(buff); err != nil {
		log.Printf("Write error: %v", err)
	}
}

// SendError sends an error as response.
// If err is an HttpError, its code and msg are used in the HTPP response.
// Also log the error.
func (self Response) SendError(err error) {
	send := func(statusCode int, msg string) {
		http.Error(self.writer, msg, statusCode)
		log.Printf("%d %s: %s", statusCode, msg, err)
	}

	var pError HttpError
	if errors.As(err, &pError) {
		send(pError.Code, pError.msg)
	} else if errors.Is(err, context.Canceled) {
		send(http.StatusInternalServerError, "Canceled")
	} else if errors.Is(err, context.DeadlineExceeded) {
		send(http.StatusGatewayTimeout, "Timed out")
	} else {
		send(http.StatusInternalServerError, "Internal error")
	}
}

// SendLoginAccepted create new credential for the user and send it as response.
func (self Response) SendLoginAccepted(ctx context.Context, user User, req *Request) {
	if err := ctx.Err(); err != nil {
		self.SendError(err)
		return
	}

	sessionId, err := MakeSessionId()
	if err != nil {
		self.SendError(err)
	}
	session := NewSession(sessionStore, sessionStore.Options, sessionId, user)
	if err = session.Save(req.original, self.writer); err != nil {
		log.Printf("Error saving session: %v", err)
	}

	self.SendJSON(ctx, sessionId)
}

// MakeSessionId create a new session id.
//
// This is a low level function, made available for tests.
func MakeSessionId() (string, error) {
	b := make([]byte, 3)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
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
