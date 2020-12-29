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
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/JBoudou/Itero/server/logger"

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
	body     []byte
}

// newRequest is the only constructor for Request.
func newRequest(basePattern string, original *http.Request) (req *Request) {
	req = &Request{original: original}

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

// CheckPOST ensures that the request is particularly safe.
//
// This method returns nil only if the method is POST and
// there is an Origin header with the correct host.
func (self *Request) CheckPOST(ctx context.Context) error {
	if self.original.Method != "POST" {
		return NewHttpError(http.StatusForbidden, "Unauthorized", "Not a POST")
	}

	origin := self.original.Header.Values("Origin")
	if origin == nil || len(origin) != 1 {
		logger.Print(ctx, "No Origin: header.")
		origin = self.original.Header.Values("Referer")
		if origin == nil || len(origin) != 1 {
			return NewHttpError(http.StatusForbidden, "Missing Origin", "No Origin nor Referer header")
		}
	}
	originUrl, err := url.Parse(origin[0])
	if err != nil {
		return err
	}
	baseUrl, err := url.Parse(BaseURL())
	if err != nil {
		return err
	}
	if originUrl.Scheme != baseUrl.Scheme ||
		originUrl.Hostname() != baseUrl.Hostname() ||
		URLPortWithDefault(originUrl) != URLPortWithDefault(baseUrl) {
		logger.Printf(ctx, "Wrong origin. Got %s. Expect %s.", originUrl, baseUrl)
		return NewHttpError(http.StatusForbidden, "Unauthorized", "Wrong origin")
	}

	return nil
}

// URLPortWithDefault returns the port part of url.Host, without the leading colon.
// If url does not have a port, a default value is guessed from url.Scheme.
func URLPortWithDefault(url *url.URL) (port string) {
	port = url.Port()
	if port == "" {
		switch url.Scheme {
		case "http":
			return "80"
		case "https":
			return "443"
		}
	}
	return
}

// UnmarshalJSONBody retrieves the body of the request as a JSON object.
// Successive calls to this method on the same object store identical objects.
// See json.Unmarshal for details of the unmarshalling process.
func (self *Request) UnmarshalJSONBody(dst interface{}) (err error) {
	if self.body == nil {
		if self.body, err = ioutil.ReadAll(self.original.Body); err != nil {
			return err
		}
	}
	return json.Unmarshal(self.body, &dst)
}

// AddSessionIdToRequest adds a session id to an http.Request.
// This function is meant to be used by HTTP clients and tests.
func AddSessionIdToRequest(req *http.Request, sessionId string) {
	req.Header.Set(sessionHeader, sessionId)
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
	queryId := self.original.Header.Get(sessionHeader)
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
	if queryId != sessionId {
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
	logger.Push(self.original.Context(), sessionId)
}

func splitPath(pathStr string) (pathSli []string) {
	clean := path.Clean(pathStr)
	pathSli = strings.Split(clean, "/")
	if clean[0] == '/' {
		pathSli = pathSli[1:]
	}
	return
}
