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
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/JBoudou/Itero/server/logger"
	"github.com/gorilla/securecookie"
)

func precheck(t *testing.T) {
	if !Ok {
		t.Log("Impossible to test package server: there is no configuration.")
		t.Log("Add a configuration file in server/ (may be a link to the main configuration file).")
		t.SkipNow()
	}
}

func canceledContext() (ret context.Context) {
	ret, cancelFct := context.WithCancel(context.Background())
	cancelFct()
	return
}

func TestResponse_SendJSON(t *testing.T) {
	type args struct {
		ctx  context.Context
		data interface{}
	}

	checkSuccess := func(t *testing.T, mock *httptest.ResponseRecorder, args *args) {
		// Check status code
		result := mock.Result()
		if result.StatusCode < 200 || result.StatusCode >= 300 {
			t.Errorf("Wrong StatusCode %d", result.StatusCode)
		}

		// Check content-type
		header := result.Header
		if !strings.Contains(header.Get("content-type"), "application/JSON") {
			t.Errorf("Content-Type doesn't contain application/JSON")
		}

		// Read the body
		var buff bytes.Buffer
		if _, err := buff.ReadFrom(result.Body); err != nil {
			t.Fatalf("Error reading body: %s", err.Error())
		}
		body := buff.Bytes()

		// Compare the body
		if !json.Valid(body) {
			t.Fatal("Invalid JSON in body")
		}
		buff.Reset()
		if err := json.Compact(&buff, body); err != nil {
			t.Fatal(err)
		}
		got := buff.Bytes()
		// We assume json.Marshal produce a compact representation
		expected, err := json.Marshal(args.data)
		if err != nil {
			t.Fatalf("Marshal error: %s", err.Error())
		}
		if !bytes.Equal(got, expected) {
			t.Errorf("Wrong body. Got %s. Expect %s", got, expected)
		}
	}
	checkFail := func(t *testing.T, mock *httptest.ResponseRecorder, args *args) {
		result := mock.Result()
		if result.StatusCode < 400 {
			t.Errorf("Wrong StatusCode %d", result.StatusCode)
		}
	}

	tests := []struct {
		name  string
		args  args
		check func(t *testing.T, mock *httptest.ResponseRecorder, args *args)
	}{
		{
			name:  "String",
			args:  args{ctx: context.Background(), data: "foobar"},
			check: checkSuccess,
		},
		{
			name:  "Number",
			args:  args{ctx: context.Background(), data: 42},
			check: checkSuccess,
		},
		{
			name:  "Slice",
			args:  args{ctx: context.Background(), data: []string{"a", "b", "c", "d"}},
			check: checkSuccess,
		},
		{
			name: "Struct",
			args: args{ctx: context.Background(), data: struct {
				foo string
				bar int
			}{"foobar", 42}},
			check: checkSuccess,
		},
		{
			name:  "Canceled",
			args:  args{ctx: canceledContext(), data: 0},
			check: checkFail,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := httptest.NewRecorder()
			self := response{
				writer: mock,
			}
			self.SendJSON(tt.args.ctx, tt.args.data)
			tt.check(t, mock, &tt.args)
		})
	}
}

func TestResponse_SendError(t *testing.T) {
	ctx := logger.New(context.Background())

	tests := []struct {
		name         string
		err          error
		expectedCode int // set to zero to disable
	}{
		{
			name:         "403",
			err:          NewHttpError(http.StatusForbidden, "Forbidden", "Test"),
			expectedCode: http.StatusForbidden,
		},
		{
			name:         "explicit 500",
			err:          NewHttpError(http.StatusInternalServerError, "Server error", "Test"),
			expectedCode: http.StatusInternalServerError,
		},
		{
			name:         "implicit 500",
			err:          errors.New("Internal error"),
			expectedCode: http.StatusInternalServerError,
		},
		{
			name: "context canceled",
			err:  context.Canceled,
		},
		{
			name: "context deadline exceeded",
			err:  context.DeadlineExceeded,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := httptest.NewRecorder()
			self := response{
				writer: mock,
			}
			self.SendError(ctx, tt.err)

			got := mock.Result().StatusCode
			if got < 400 {
				t.Errorf("Wrong status code %d", got)
			}
			if tt.expectedCode != 0 && got != tt.expectedCode {
				t.Errorf("Status code. Got %d. Expect %d", got, tt.expectedCode)
			}
		})
	}
}

func TestResponse_SendLoginAccepted(t *testing.T) {
	precheck(t)

	type args struct {
		ctx  context.Context
		user User
		req  *Request
	}

	checkSuccess := func(t *testing.T, mock *httptest.ResponseRecorder, args *args) {
		now := time.Now()

		// Check status code
		result := mock.Result()
		if result.StatusCode < 200 || result.StatusCode >= 300 {
			t.Errorf("Wrong StatusCode %d", result.StatusCode)
		}

		// Read the body
		var body SessionAnswer
		if err := json.NewDecoder(result.Body).Decode(&body); err != nil {
			t.Fatalf("Error reading body: %s", err)
		}

		// Read the cookie
		cookie := findCookie(result.Cookies(), sessionName)
		if cookie == nil {
			t.Fatalf("No cookie named %s", sessionName)
		}
		codecs := securecookie.CodecsFromPairs(cfg.SessionKeys...)
		var values map[interface{}]interface{}
		if err := securecookie.DecodeMulti(sessionName, cookie.Value, &values, codecs...); err != nil {
			t.Fatalf("Decode cookie: %s", err)
		}

		// Check expire and deadline
		expireDuration := cookie.Expires.Sub(now).Seconds()
		if limit := float64(sessionMaxAge - sessionGraceTime); expireDuration < limit {
			t.Errorf("Expire too early. Got %f. Expect %f", expireDuration, limit)
		}
		if limit := float64(sessionMaxAge + sessionGraceTime); expireDuration > limit {
			t.Errorf("Expire too late. Got %f. Expect %f", expireDuration, limit)
		}
		untyped, ok := values[sessionKeyDeadline]
		if !ok {
			t.Fatalf("Not found key %s in cookie", sessionKeyDeadline)
		}
		deadline, ok := untyped.(int64)
		if !ok {
			t.Fatalf("Wrong type for key %s in cookie", sessionKeyDeadline)
		}
		if limit := now.Unix() + sessionMaxAge; deadline < limit {
			t.Errorf("Expire too early. Got %d. Expect %d", deadline, limit)
		}
		if limit := now.Unix() + sessionMaxAge + (2 * sessionGraceTime); deadline > limit {
			t.Errorf("Expire too late. Got %d. Expect %d", deadline, limit)
		}
		if body.Expires.Sub(cookie.Expires) > sessionGraceTime*time.Second {
			t.Errorf("Body Expires too late. Got %v. Expect %v.", body.Expires, cookie.Expires)
		}

		// Check the other session values
		getString := func(key string) (value string) {
			untyped, ok := values[key]
			if !ok {
				t.Fatalf("Not found key %s in cookie", key)
			}
			value, ok = untyped.(string)
			if !ok {
				t.Fatalf("Wrong type for key %s in cookie", key)
			}
			return
		}
		if cookieSessionId := getString(sessionKeySessionId); cookieSessionId != body.SessionId {
			t.Errorf("Wrong session ID. Body %s. Cookie %s.", body.SessionId, cookieSessionId)
		}
		if userName := getString(sessionKeyUserName); userName != args.user.Name {
			t.Errorf("Wrong user name. Got %s. Expect %s.", userName, args.user.Name)
		}
		untyped, ok = values[sessionKeyUserId]
		if !ok {
			t.Fatalf("Not found key %s in cookie", sessionKeyUserId)
		}
		userId, ok := untyped.(uint32)
		if !ok {
			t.Fatalf("Wrong type for key %s in cookie", sessionKeyUserId)
		}
		if userId != args.user.Id {
			t.Errorf("Wrong user Id. Got %d. Expect %d", userId, args.user.Id)
		}
	}
	checkFail := func(t *testing.T, mock *httptest.ResponseRecorder, args *args) {
		result := mock.Result()
		if result.StatusCode < 400 {
			t.Errorf("Wrong StatusCode %d", result.StatusCode)
		}
	}

	tests := []struct {
		name  string
		args  args
		check func(t *testing.T, mock *httptest.ResponseRecorder, args *args)
	}{
		{
			name: "Success",
			args: args{
				ctx:  context.Background(),
				user: User{Name: "Foo", Id: 42},
				req:  &Request{original: &http.Request{}},
			},
			check: checkSuccess,
		},
		{
			name: "Failure",
			args: args{
				ctx:  canceledContext(),
				user: User{Name: "Foo", Id: 42},
				req:  &Request{original: &http.Request{}},
			},
			check: checkFail,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := httptest.NewRecorder()
			self := response{
				writer: mock,
			}
			self.SendLoginAccepted(tt.args.ctx, tt.args.user, tt.args.req)
			tt.check(t, mock, &tt.args)
		})
	}
}

func findCookie(cookies []*http.Cookie, name string) (found *http.Cookie) {
	for _, cookie := range cookies {
		if cookie.Name == name {
			return cookie
		}
	}
	return
}
