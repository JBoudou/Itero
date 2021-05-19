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

package servertest

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/JBoudou/Itero/server"
)

// Checker checks that a given response is as expected.
type Checker interface {
	// Check is called at the end of each test, after the handler has been called. This method must
	// check that the handler has done its job correctly.
	Check(t *testing.T, response *http.Response, request *server.Request)
}

/* ChecherFun */

type CheckerFun func(t *testing.T, response *http.Response, request *server.Request)

func (self CheckerFun) Check(t *testing.T, response *http.Response, request *server.Request) {
	self(t, response, request)
}

/* CheckJSON */

// CheckJSON checks that the response body is similar to the JSON marshaling of a given value.
type CheckJSON struct {
	Code    int // If zero, http.StatusOK is used instead.
	Body    interface{}
	Partial bool // If true, Body may lack some field of the response.
}

// Check implements Checker.
func (self CheckJSON) Check(t *testing.T, response *http.Response, request *server.Request) {
	t.Helper()

	// Code
	expectCode := self.Code
	if expectCode == 0 {
		expectCode = http.StatusOK
	}
	if response.StatusCode != expectCode {
		t.Errorf("Wrong status code. Got %d. Expect %d", response.StatusCode, expectCode)
	}

	// Body
	if self.Partial {
		self.partial(t, response)
	} else {
		self.full(t, response)
	}
}

func (self CheckJSON) full(t *testing.T, response *http.Response) {
	t.Helper()

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
	expectBodyEncoded, err := json.Marshal(self.Body)
	if err != nil {
		t.Fatalf("Error encoding body: %s", err)
	}

	if !bytes.Equal(body, expectBodyEncoded) {
		t.Errorf("Wrong body. Got %s. Expect %s", body, expectBodyEncoded)
	}
}

func (self CheckJSON) partial(t *testing.T, response *http.Response) {
	t.Helper()

	expectJSON, err := json.Marshal(self.Body)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { json.Unmarshal(expectJSON, self.Body) } ()

	bodyDecoder := json.NewDecoder(response.Body)
	if err = bodyDecoder.Decode(self.Body); err != nil {
		t.Fatal(err)
	}
	gotJSON, err := json.Marshal(self.Body)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(gotJSON, expectJSON) {
		t.Errorf("Got %s. Expect %s.", gotJSON, expectJSON)
	}
}


/* CheckError */

// CheckError checks the statusCode and error message of a response.
type CheckError struct {
	Code int
	Body string
}

// Check implements Checker.
func (self CheckError) Check(t *testing.T, response *http.Response, request *server.Request) {
	t.Helper()
	if response.StatusCode != self.Code {
		t.Errorf("Wrong status code. Got %d. Expect %d", response.StatusCode, self.Code)
	}

	var buff bytes.Buffer
	if _, err := buff.ReadFrom(response.Body); err != nil {
		t.Fatalf("Error reading body: %s", err)
	}
	body := strings.TrimSpace(string(buff.Bytes()))

	if body != self.Body {
		t.Errorf("Wrong error. Got %s. Expect %s.", body, self.Body)
	}
}

/* CheckStatus */

// CheckStatus checks only the statusCode of a response.
type CheckStatus struct {
	Code int
}

// Check implements Checker.
func (self CheckStatus) Check(t *testing.T, response *http.Response, request *server.Request) {
	t.Helper()
	if response.StatusCode != self.Code {
		t.Errorf("Wrong status code. Got %d. Expect %d", response.StatusCode, self.Code)
	}
}

/* CheckAnyErrorStatus */

var CheckAnyErrorStatus =
	CheckerFun(func(t *testing.T, response *http.Response, request *server.Request) {
		t.Helper()
		if response.StatusCode < 400 {
			t.Errorf("Wrong status code. Got %d. Expect at least 400.", response.StatusCode)
		}
	})

/* CheckCookieIsSet */

type CheckCookieIsSet struct {
	Name string
}

func (self CheckCookieIsSet) Check(t *testing.T, response *http.Response, request *server.Request) {
	if FindCookie(response, self.Name) == nil {
		t.Errorf("No cookie named %s.", self.Name)
	}
}
