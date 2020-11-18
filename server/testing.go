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

// The tests for this file are in handler_test.go.

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestHandler execute the given handler for the given request.
//
// The returned http.Response is what would have been received by a client if the handler
// was registered for the given pattern and the given http.Request was directed to that handler.
// Notice that there is not check that the path of the request corresponds to the pattern.
//
// This function is meant to be used to test Handlers.
func TestHandler(pattern string, handler Handler, request *http.Request) *http.Response {
	mock := httptest.NewRecorder()
	handlerWrapper{pattern: pattern, handler: handler}.ServeHTTP(mock, request)
	return mock.Result()
}

// TestHandlerFunc is a thin wrapper around TestHandler. See TestHandler for details.
func TestHandlerFunc(pattern string, handler HandleFunction, request *http.Request) *http.Response {
	return TestHandler(pattern, HandlerFunc(handler), request)
}

// TestChecker is a signature for functions that check the result of a request on a Handler.
type TestChecker = func(t *testing.T, response *http.Response, req *http.Request)

// TestCheckerJSON returns a TestChecker to check responses whose body is a JSON object.
//
// The returned function checks that the statuc code and the body are as expected.
func TestCheckerJSON(expectCode int, expectBody interface{}) TestChecker {
	return func(t *testing.T, response *http.Response, req *http.Request) {
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
		body = make([]byte, buff.Len())
		if _, err := buff.Read(body); err != nil {
			t.Fatalf("Error reading body: %s", err)
		}

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

// TestCheckerJSONString returns a TestChecker to check status code.
//
// The returned function checks that the statuc code is as expected. The body is not checked.
func TestCheckerStatus(expectCode int) TestChecker {
	return func(t *testing.T, response *http.Response, req *http.Request) {
		if response.StatusCode != expectCode {
			t.Errorf("Wrong status code. Got %d. Expect %d", response.StatusCode, expectCode)
		}
	}
}

// TestCheckerJSONString returns a TestChecker to check responses whose body is a JSON object
// representing a string.
//
// The returned function checks that the statuc code and the encoded string are as expected.
func TestCheckerJSONString(expectCode int, expectBody string) TestChecker {
	return func(t *testing.T, response *http.Response, req *http.Request) {
		if response.StatusCode != expectCode {
			t.Errorf("Wrong status code. Got %d. Expect %d", response.StatusCode, expectCode)
		}

		var body string
		var buff bytes.Buffer
		if _, err := buff.ReadFrom(response.Body); err != nil {
			t.Fatalf("Error reading body: %s", err)
		}
		if err := json.Unmarshal(buff.Bytes(), &body); err != nil {
			t.Fatalf("Error reading body: %s", err)
		}
		if body != expectBody {
			t.Errorf("Wrong body. Got %s. Expect %s", body, expectBody)
		}
	}
}

// TestCheckerJSONString returns a TestCheckerto check responses whose body is a JSON object
// representing a string.
//
// The returned function checks that the statuc code and the encoded string are as expected.
func TestCheckerRawString(expectCode int, expectBody string) TestChecker {
	return func(t *testing.T, response *http.Response, req *http.Request) {
		if response.StatusCode != expectCode {
			t.Errorf("Wrong status code. Got %d. Expect %d", response.StatusCode, expectCode)
		}

		var buff bytes.Buffer
		if _, err := buff.ReadFrom(response.Body); err != nil {
			t.Fatalf("Error reading body: %s", err)
		}
		if body := strings.TrimSpace(string(buff.Bytes())); body != expectBody {
			t.Errorf("Wrong body. Got %s. Expect %s", body, expectBody)
		}
	}
}
