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
)

// Error with corresponding HTTP code.
type HttpError struct {
	// HTTP code to send to reflect this error.
	Code   int
	msg    string
	detail string
}

// Create a new HttpError.
// The code is to be sent as the HTTP code of the response. It should be a constant from the
// net/http package. The message (msg) is to be sent as body of the HTTP response. This is the
// public description of the error. The detail is the private description of the error, to be
// displayed in the logs.
func NewHttpError(code int, msg string, detail string) HttpError {
	return HttpError{code, msg, detail}
}

func (self HttpError) Error() string {
	return self.detail
}

type Response struct {
	Writer http.ResponseWriter
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
	if _, err = self.Writer.Write(buff); err != nil {
		log.Printf("Write error: %v", err)
	}
}

// SendError sends an error as response.
// If err is an HttpError, its code and msg are used in the HTPP response.
// Also log the error.
func (self Response) SendError(err error) {
	var pError HttpError
	if errors.As(err, &pError) {
		http.Error(self.Writer, pError.msg, pError.Code)
		log.Printf("%d %s: %s", pError.Code, pError.msg, pError.detail)
	} else if errors.Is(err, context.Canceled) {
		http.Error(self.Writer, "Canceled", http.StatusInternalServerError)
		log.Printf("%d %s: %s", http.StatusInternalServerError, "Context canceled", err.Error())
	} else if errors.Is(err, context.DeadlineExceeded) {
		http.Error(self.Writer, "Timed out", http.StatusGatewayTimeout)
		log.Printf("%d %s: %s", http.StatusGatewayTimeout, "Context timed out", err.Error())
	} else {
		http.Error(self.Writer, err.Error(), http.StatusInternalServerError)
		log.Printf("%d %v", http.StatusInternalServerError, err)
	}
}

// SendLoginAccepted create new credential for the user and send it as response.
func (self Response) SendLoginAccepted(ctx context.Context, user User, req *Request) {
	if err := ctx.Err(); err != nil {
		self.SendError(err)
		return
	}

	session, err := sessionStore.New(req.original, sessionName)
	if err != nil {
		self.SendError(err)
		return
	}

	sessionId, err := randomShortId()
	if err != nil {
		self.SendError(err)
	}

	session.Values[sessionKeySessionId] = sessionId
	session.Values[sessionKeyUserName] = user.Name
	session.Values[sessionKeyUserId] = user.Id
	session.Values[sessionKeyDeadline] = time.Now().Unix() + sessionMaxAge + sessionGraceTime

	if err = session.Save(req.original, self.Writer); err != nil {
		log.Printf("Error saving session: %v", err)
	}

	self.SendJSON(ctx, sessionId)
}

func randomShortId() (string, error) {
	b := make([]byte, 3)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
