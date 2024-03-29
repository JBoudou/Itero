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

// Package servertest provides methods and types to test server.Handler implementations.
package servertest

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/JBoudou/Itero/mid/root"
	"github.com/JBoudou/Itero/mid/server"
	"github.com/JBoudou/Itero/pkg/ioc"
	"github.com/JBoudou/Itero/pkg/slog"
)

var (
	clientStore *ClientStore
)

func init() {
	clientStore = NewClientStore(server.SessionKeys()...)
}

// Request provides information to create an http.Request.
//
// Its zero value is valid and produces the request "GET /a/test". See Make() for details.
type Request struct {
	Method     string
	Target     *string
	RemoteAddr *string
	Body       string
	UserId     *uint32
	Hash       *uint32
}

// Make generates an http.Request.
//
// Default value for Method is "GET". Default value for Target is "/a/test".
// If RemoteAddr is not nil, the RemoteAddr field of the returned request is set to its value.
// If UserId is not nil and Hash is nil then a valid session for that user is added to the request.
// If UserId and Hash are both non-nil then an "unlogged cookie" is added to the request.
func (self *Request) Make(t *testing.T) (req *http.Request, err error) {
	var target string
	if self.Target == nil {
		target = "/a/test"
	} else {
		target = *self.Target
	}

	var body io.Reader
	if self.Body != "" {
		body = strings.NewReader(self.Body)
	}

	req = httptest.NewRequest(self.Method, target, body)

	if req.Method == "POST" {
		req.Header.Add("Origin", server.BaseURL())
	}

	if self.RemoteAddr != nil {
		req.RemoteAddr = *self.RemoteAddr
	}

	if self.UserId != nil && self.Hash == nil {
		var sessionId string
		sessionId, err = server.MakeSessionId()
		if err != nil {
			return
		}
		server.AddSessionIdToRequest(req, sessionId)
		user := server.User{Name: " Test ", Id: *self.UserId, Logged: true}
		sessionAnswer := server.SessionAnswer{SessionId: sessionId}
		session := server.NewSession(clientStore, &server.SessionOptions, &sessionAnswer, user)
		clientStore.Save(req, nil, session)
	}
	if self.UserId != nil && self.Hash != nil {
		user := server.User{Id: *self.UserId, Hash: *self.Hash, Logged: false}
		session := server.NewUnloggedUser(clientStore, &server.SessionOptions, user)
		clientStore.Save(req, nil, session)
	}

	ctx := slog.CtxSaveLogger(req.Context(), &slog.WithStack{Target: t})
	req = req.WithContext(ctx)

	return
}

// Test represents a test to be executed by Run().
// A simple implementation is given by T.
type Test interface {
	GetName() string

	// Prepare is called before the handler is created.
	Prepare(t *testing.T, loc *ioc.Locator) *ioc.Locator

	// GetRequest is called just after the handler is created.
	GetRequest(t *testing.T) *Request

	Checker
	Close()
}

// ChainPrepare sequentially call the Prepare method of its arguments.
// The methods are called in the same order as the arguments.
// The locator is updated by each call, then returned.
func ChainPrepare(t *testing.T, loc *ioc.Locator, chain ...interface {
	Prepare(*testing.T, *ioc.Locator) *ioc.Locator
}) *ioc.Locator {
	for _, pre := range chain {
		loc = pre.Prepare(t, loc)
	}
	return loc
}

// T is a simple implementation of Test.
type T struct {
	Name    string
	Request Request

	// Update is called before the test, if not nil.
	Update func(t *testing.T)

	Checker Checker
}

func (self *T) GetName() string {
	return self.Name
}

func (self *T) GetRequest(t *testing.T) *Request {
	return &self.Request
}

// Prepare runs before the handler is executed.
// If Update is not nil, it is called first.
// If Checker implements a method Before(*testing.T) then it is called next.
func (self *T) Prepare(t *testing.T, loc *ioc.Locator) *ioc.Locator {
	if self.Update != nil {
		self.Update(t)
	}
	if checker, ok := self.Checker.(interface{ Before(t *testing.T) }); ok {
		checker.Before(t)
	}
	return loc
}

func (self *T) Close() {
}

func (self *T) Check(t *testing.T, response *http.Response, request *server.Request) {
	self.Checker.Check(t, response, request)
}

// Run executes all the given tests on the given Handler.
//
// The tests are executed in the given order. For each test, the handler is created by calling
// handlerFactory using the ioc.Locator returned by Prepare. The handler is registered for the
// pattern "/a/test".
//
// If a test implement the method
//    ChangeResponse(*testing.T, server.Response) server.Response
// then this method is called with the usual server.Response objet, and the returned object is used
// by the handler. This method allows the test to spy the calls to server.Response methods.
//
// Each test is executed inside testing.T.Run, hence calling t.Fatal in the checker abort only the
// current test.
func Run(t *testing.T, tests []Test, handlerFactory interface{}) {
	t.Helper()
	for _, tt := range tests {
		tt := tt // Copy in case tests are run in parallel.
		t.Run(tt.GetName(), func(t *testing.T) {
			t.Helper()
			defer tt.Close()

			locator := tt.Prepare(t, root.IoC)
			var handler server.Handler
			err := locator.Inject(handlerFactory, &handler)
			if err != nil {
				t.Fatalf("Injection error: %v", err)
			}

			req, err := tt.GetRequest(t).Make(t)
			if err != nil {
				t.Fatalf("Error creating request: %s", err)
			}

			mock := httptest.NewRecorder()
			wrapper := server.NewHandlerWrapper("/a/test", handler)
			ctx, sResp, sReq := wrapper.MakeParams(mock, req)

			if withResponse, ok := tt.(interface {
				ChangeResponse(*testing.T, server.Response) server.Response
			}); ok {
				sResp = withResponse.ChangeResponse(t, sResp)
			}

			wrapper.Exec(ctx, sResp, sReq)

			tt.Check(t, mock.Result(), sReq)
		})
	}
}

// RunFunc is a convenient wrapper around Run for HandleFunction.
func RunFunc(t *testing.T, tests []Test, handler server.HandleFunction) {
	t.Helper()
	Run(t, tests,
		func() server.Handler { return server.HandlerFunc(handler) })
}

// FindCookie returns a cookie by name.
func FindCookie(response *http.Response, name string) (found *http.Cookie) {
	for _, cookie := range response.Cookies() {
		if cookie.Name == name {
			return cookie
		}
	}
	return
}
