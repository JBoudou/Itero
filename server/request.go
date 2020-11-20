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
	"bytes"
	"encoding/json"
	"net/http"
	"path"
	"strings"
	"time"

	gs "github.com/gorilla/sessions"
)

// A request represents an HTTP request to be handled by the server.
type Request struct {
	// User is the session user.
	// It is nil if no user is successfully logged in the current session.
	User *User

	// SessionError is the error raised when checking the session informations send by the client.
	// It is nil either if the client did not send any session information (in which case User is nil
	// too) or if the session has been successfully checked (in which case Use is not nil).
	SessionError error

	// FullPath contains all path elements of the request made by the client.
	FullPath []string

	// RemainingPath contains the path elements after the pattern corresponding to the current
	// Handler.
	RemainingPath []string

	original *http.Request
}

// newRequest the only constructor for Request.
func newRequest(basePattern string, original *http.Request) (req Request) {
	req.original = original

	var session *gs.Session
	session, req.SessionError = sessionStore.Get(original, sessionName)
	if req.SessionError == nil && !session.IsNew {
		req.addSession(session)
	}

	req.FullPath = splitPath(req.original.URL.Path)
	basePath := splitPath(basePattern)
	req.RemainingPath = req.FullPath[len(basePath):]
	return
}

// UnmarshalJSONBody retrieves the body of the request as a JSON object.
// See json.Unmarshal for details of the unmarshalling process.
func (self *Request) UnmarshalJSONBody(dst interface{}) error {
	var buff bytes.Buffer
	if _, err := buff.ReadFrom(self.original.Body); err != nil {
		return err
	}
	return json.Unmarshal(buff.Bytes(), &dst)
}

// AddSessionIdToPath adds a session id to a URL path.
// This function is meant to be used by HTTP clients and tests.
func AddSessionIdToPath(path *string, sessionId string) {
	var inter string
	if strings.Contains(*path, "?") {
		inter = "&"
	} else {
		inter = "?"
	}
	*path = *path + inter + queryKeySessionId + "=" + sessionId
}

/* What follows are private methods and functions */

func (self *Request) addSession(session *gs.Session) {
	registerError := func(detail string) {
		self.SessionError = NewHttpError(http.StatusForbidden, "Unauthorized", detail)
	}
	// TODO use functions to retrieve a typed value from a session

	// Check deadline
	unconverted, ok := session.Values[sessionKeyDeadline]
	if !ok {
		registerError("no deadline in the session cookie")
		return
	}
	deadline, ok := unconverted.(int64)
	if !ok {
		registerError("the deadline in the session cookie has wrong type")
		return
	}

	if time.Now().Unix() > deadline {
		registerError("the session cookie has expired")
		return
	}

	// Check session id
	queryId, ok := self.original.URL.Query()[queryKeySessionId]
	if !ok {
		registerError("no session id in query")
		return
	}
	unconverted, ok = session.Values[sessionKeySessionId]
	if !ok {
		registerError("no session id in the cookie")
		return
	}
	sessionId, ok := unconverted.(string)
	if !ok {
		registerError("wrong type for the session id in the cookie")
		return
	}
	if queryId[0] != sessionId {
		registerError("wrong session id")
		return
	}

	// Set self.User
	unconverted, ok = session.Values[sessionKeyUserName]
	if !ok {
		registerError("no user name in the cookie")
		return
	}
	userName, ok := unconverted.(string)
	if !ok {
		registerError("wrong type for the user name in the cookie")
		return
	}
	unconverted, ok = session.Values[sessionKeyUserId]
	if !ok {
		registerError("no user id in the cookie")
		return
	}
	userId, ok := unconverted.(uint32)
	if !ok {
		registerError("wrong type for the user id in the cookie")
		return
	}
	self.User = &User{Name: userName, Id: userId}
	self.SessionError = nil
}

func splitPath(pathStr string) (pathSli []string) {
	clean := path.Clean(pathStr)
	pathSli = strings.Split(clean, "/")
	if clean[0] == '/' {
		pathSli = pathSli[1:]
	}
	return
}
