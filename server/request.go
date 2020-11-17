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

type Request struct {
	User					*User
	SessionError	error
	FullPath      []string
	RemainingPath []string

	original *http.Request
}

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

func (self *Request) UnmarshalJSONBody(dst interface{}) error {
		var buff bytes.Buffer
		if _, err := buff.ReadFrom(self.original.Body); err != nil {
			return err
		}
		return json.Unmarshal(buff.Bytes(), &dst)
}


func (self *Request) addSession(session *gs.Session) {
	registerError := func (detail string) {
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
