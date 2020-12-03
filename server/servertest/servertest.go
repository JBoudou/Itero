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
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/JBoudou/Itero/server"
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
	Method string
	Target *string
	Body   string
	UserId *uint32
}

// Make generates an http.Request.
//
// Default value for Method is "GET". Default value for Target is "/a/test". If UserId is not nil
// then a valid session for that user is added to the request.
func (self *Request) Make() (req *http.Request, err error) {
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
	
	if self.UserId != nil {
		var sessionId string
		sessionId, err = server.MakeSessionId()
		if err != nil {
			return
		}
		server.AddSessionIdToRequest(req, sessionId)
		user := server.User{Name: " Test ", Id: *self.UserId}
		session := server.NewSession(clientStore, &server.SessionOptions, sessionId, user)
		clientStore.Save(req, nil, session)
	}
	
	return
}

// Checker is a signature for functions that check the result of a request on a Handler.
type Checker = func(t *testing.T, response *http.Response, req *http.Request)

// Test represents a test to be executed by Run().
//
// Update is called before the test, if not nil.
type Test struct {
	Name    string
	Update  func(t *testing.T)
	Request Request
	Checker Checker
}

// Run executes all the given tests on the given Handler.
//
// The same handler is used for all tests. The tests are executed in the given order.
// The requests received by the handler are as if the handler had been registered for the pattern
// "/a/test".
//
// Each test is executed inside testing.T.Run, hence calling t.Fatal in the checker abort only the
// current test.
func Run(t *testing.T, tests []Test, handler server.Handler) {
	t.Helper()
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Helper()
			if tt.Update != nil {
				tt.Update(t)
			}
			req, err := tt.Request.Make()
			if err != nil {
				t.Fatalf("Error creating request: %s", err)
			}
			mock := httptest.NewRecorder()
			server.NewHandlerWrapper("/a/test", handler).ServeHTTP(mock, req)
			tt.Checker(t, mock.Result(), req)
		})
	}
}

// RunFunc is a convenient wrapper around Run for HandleFunction.
func RunFunc(t *testing.T, tests []Test, handler server.HandleFunction) {
	t.Helper()
	Run(t, tests, server.HandlerFunc(handler))
}

// CheckerJSON returns a Checker to check responses whose body is a JSON object.
//
// The returned function checks that the statuc code and the body are as expected.
func CheckerJSON(expectCode int, expectBody interface{}) Checker {
	return func(t *testing.T, response *http.Response, req *http.Request) {
		t.Helper()
		if response.StatusCode != expectCode {
			t.Errorf("Wrong status code. Got %d. Expect %d", response.StatusCode, expectCode)
		}

		var body []byte
		var buff bytes.Buffer
		if _, err := buff.ReadFrom(response.Body); err != nil {
			t.Fatalf("Error reading body: %s", err)
		}
		body = make([]byte, buff.Len())
		if _, err := buff.Read(body); err != nil {
			t.Fatalf("Error reading body: %s", err)
		}
		if err := json.Compact(&buff, body); err != nil {
			t.Fatalf("Error reading body: %s", err)
		}
		body = buff.Bytes()

		// We assume json.Marshal produces compact JSON.
		expectBodyEncoded, err := json.Marshal(expectBody)
		if err != nil {
			t.Fatalf("Error encoding body: %s", err)
		}

		if !bytes.Equal(body, expectBodyEncoded) {
			t.Errorf("Wrong body. Got %s. Expect %s", body, expectBodyEncoded)
		}
	}
}

func CheckerError(expectCode int, expectBody string) Checker {
	return func(t *testing.T, response *http.Response, req *http.Request) {
		t.Helper()
		if response.StatusCode != expectCode {
			t.Errorf("Wrong status code. Got %d. Expect %d", response.StatusCode, expectCode)
		}

		var buff bytes.Buffer
		if _, err := buff.ReadFrom(response.Body); err != nil {
			t.Fatalf("Error reading body: %s", err)
		}
		body := strings.TrimSpace(string(buff.Bytes()))

		if body != expectBody {
			t.Errorf("Wrong error. Got %s. Expect %s.", body, expectBody)
		}
	}
}

// CheckerJSONString returns a Checker to check status code.
//
// The returned function checks that the statuc code is as expected. The body is not checked.
func CheckerStatus(expectCode int) Checker {
	return func(t *testing.T, response *http.Response, req *http.Request) {
		t.Helper()
		if response.StatusCode != expectCode {
			t.Errorf("Wrong status code. Got %d. Expect %d", response.StatusCode, expectCode)
		}
	}
}
