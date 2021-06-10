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
	"reflect"
	"testing"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
)

func TestNewRequest(t *testing.T) {
	precheck(t)

	codecs := securecookie.CodecsFromPairs(SessionKeys()...)

	type cookie struct {
		name   string
		values map[interface{}]interface{}
	}
	type args struct {
		basePattern string
		method      string
		target      string
		cookies     []cookie
	}
	type expect struct {
		user          *User
		sessionError  error
		fullPath      []string
		remainingPath []string
	}
	tests := []struct {
		name   string
		args   args
		expect expect
	}{
		{
			name: "Path",
			args: args{
				basePattern: "/foo/bar",
				method:      "GET",
				target:      "/foo/bar/baz/last",
			},
			expect: expect{
				sessionError:  nil, // No session error when there is no cookie
				fullPath:      []string{"foo", "bar", "baz", "last"},
				remainingPath: []string{"baz", "last"},
			},
		},
		{
			name: "Full Unlogged",
			args: args{
				method: "GET",
				cookies: []cookie{{
					name: SessionUnlogged,
					values: map[interface{}]interface{}{
						sessionKeyUserId: uint32(42),
						sessionKeyHash:   uint32(27),
					},
				}},
			},
			expect: expect{
				user: &User{Id: 42, Hash: 27, Logged: false},
			},
		},
		{
			name: "Unlogged no Id",
			args: args{
				method: "GET",
				cookies: []cookie{{
					name: SessionUnlogged,
					values: map[interface{}]interface{}{
						sessionKeyHash:   uint32(27),
					},
				}},
			},
			expect: expect{
				user: nil,
			},
		},
		{
			name: "Unlogged no Hash",
			args: args{
				method: "GET",
				cookies: []cookie{{
					name: SessionUnlogged,
					values: map[interface{}]interface{}{
						sessionKeyUserId: uint32(42),
					},
				}},
			},
			expect: expect{
				user: nil,
			},
		},
		{
			name: "Unlogged wrong type",
			args: args{
				method: "GET",
				cookies: []cookie{{
					name: SessionUnlogged,
					values: map[interface{}]interface{}{
						sessionKeyUserId: "42",
						sessionKeyHash:   "27",
					},
				}},
			},
			expect: expect{
				user: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Fix tt
			if tt.args.target == "" {
				tt.args.target = "/"
			}
			if tt.expect.fullPath == nil {
				tt.expect.fullPath = []string{""}
			}
			if tt.expect.remainingPath == nil {
				tt.expect.remainingPath = []string{}
			}

			// Make http.Request
			request := httptest.NewRequest(tt.args.method, tt.args.target, nil)
			for _, cookie := range tt.args.cookies {
				encoded, err := securecookie.EncodeMulti(cookie.name, cookie.values, codecs...)
				mustt(t, err)
				request.AddCookie(sessions.NewCookie(cookie.name, encoded, &SessionOptions))
			}

			// Run test
			got := newRequest(tt.args.basePattern, request)
			if !reflect.DeepEqual(got.User, tt.expect.user) {
				t.Errorf("Wrong user. Got %v. Expect %v.", got.User, tt.expect.user)
			}
			if !errors.Is(got.SessionError, tt.expect.sessionError) {
				t.Errorf("Wrong session error. Got %v. Expect %v.", got.SessionError, tt.expect.sessionError)
			}
			if !reflect.DeepEqual(got.FullPath, tt.expect.fullPath) {
				t.Errorf("Wrong full path. Got %v. Expect %v.", got.FullPath, tt.expect.fullPath)
			}
			if !reflect.DeepEqual(got.RemainingPath, tt.expect.remainingPath) {
				t.Errorf("Wrong remaining path. Got %v. Expect %v.", got.RemainingPath, tt.expect.remainingPath)
			}
		})
	}
}

// TestLoginThenNewRequest tests that Response.SendLoginAccepted works well with NewRequest.
// This is a "black box test": we don't care what the functions precisely do.
func TestLoginThenNewRequest(t *testing.T) {
	precheck(t)

	const singlePath = "foo"
	const fullPath = "/" + singlePath
	user := User{Name: "John", Id: 42, Logged: true}

	addCorrectSession := func(t *testing.T, result *http.Response, request *http.Request) {
		var body SessionAnswer
		if err := json.NewDecoder(result.Body).Decode(&body); err != nil {
			t.Fatalf("Unable to read response body: %s", err)
		}

		AddSessionIdToRequest(request, body.SessionId)
	}

	addForgedSession := func(value string) func(t *testing.T, result *http.Response,
		request *http.Request) {
		return func(t *testing.T, result *http.Response, request *http.Request) {
			AddSessionIdToRequest(request, value)
		}
	}

	checkSuccess := func(t *testing.T, got *Request, original *http.Request) {
		if !reflect.DeepEqual(got.User, &user) {
			t.Errorf("Wrong User. Got %v. Expect %v.", got.User, user)
		}
		if !reflect.DeepEqual(got.FullPath, []string{singlePath}) {
			t.Errorf("Wrong FullPath. Got %v. Expect [%s].", got.FullPath, singlePath)
		}
		if !reflect.DeepEqual(got.RemainingPath, []string{}) {
			t.Errorf("Wrong RemainingPath. Got %v. Expect [].", got.RemainingPath)
		}
	}

	checkFail := func(t *testing.T, got *Request, original *http.Request) {
		if got.User != nil {
			t.Errorf("Got %v. Expect nil.", got.User)
		}
		if got.SessionError == nil {
			t.Errorf("Expect SessionError not to be nil.")
		}
	}

	tests := []struct {
		name       string
		addSession func(t *testing.T, result *http.Response, request *http.Request)
		checker    func(t *testing.T, got *Request, original *http.Request)
	}{
		{
			name:       "Success",
			addSession: addCorrectSession,
			checker:    checkSuccess,
		},
		{
			name:       "No session",
			addSession: func(t *testing.T, result *http.Response, request *http.Request) {},
			checker:    checkFail,
		},
		{
			name:       "Wrong session",
			addSession: addForgedSession(";;;;"),
			checker:    checkFail,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := httptest.NewRecorder()

			response := response{writer: mock}
			response.SendLoginAccepted(context.Background(), user, &Request{original: &http.Request{}}, 0)
			result := mock.Result()

			originalRequest := httptest.NewRequest("GET", fullPath, nil)
			for _, cookie := range result.Cookies() {
				originalRequest.AddCookie(cookie)
			}
			tt.addSession(t, result, originalRequest)

			got := newRequest(fullPath, originalRequest)
			tt.checker(t, got, originalRequest)
		})
	}
}

func TestRequest_CheckPOST(t *testing.T) {
	const target = "/a/test"

	originAddress := cfg.Address
	defer func() { cfg.Address = originAddress }()

	tests := []struct {
		name     string
		method   string
		headers  map[string]string
		address  string // internal server address, i.e., cfg.Address
		errorMsg string // empty for success
	}{
		{
			name:     "Get",
			method:   "GET",
			errorMsg: "Unauthorized",
		},
		{
			name:     "No additional header",
			method:   "POST",
			errorMsg: "Missing Origin",
		},
		{
			name:    "Wrong origin host",
			method:  "POST",
			address: "example.com",
			headers: map[string]string{
				"Origin": "https://impossible.dummy.address/",
			},
			errorMsg: "Unauthorized",
		},
		{
			name:    "Wrong origin scheme",
			method:  "POST",
			address: "example.com",
			headers: map[string]string{
				"Origin": "http://example.com/",
			},
			errorMsg: "Unauthorized",
		},
		{
			name:    "Wrong port",
			method:  "POST",
			address: "example.com:17",
			headers: map[string]string{
				"Origin": "https://example.com:18/",
			},
			errorMsg: "Unauthorized",
		},
		{
			name:    "Wrong port default 1",
			method:  "POST",
			address: "example.com",
			headers: map[string]string{
				"Origin": "https://example.com:18/",
			},
			errorMsg: "Unauthorized",
		},
		{
			name:    "Wrong port default 2",
			method:  "POST",
			address: "example.com:42",
			headers: map[string]string{
				"Origin": "https://example.com/",
			},
			errorMsg: "Unauthorized",
		},
		{
			name:    "Good port default 1",
			method:  "POST",
			address: "example.com",
			headers: map[string]string{
				"Origin": "https://example.com:443/",
			},
		},
		{
			name:    "Good port default 2",
			method:  "POST",
			address: "example.com:443",
			headers: map[string]string{
				"Origin": "https://example.com/",
			},
		},
		{
			name:    "Good port explicit",
			method:  "POST",
			address: "example.com:42",
			headers: map[string]string{
				"Origin": "https://example.com:42/",
			},
		},
		{
			name:    "Success Origin",
			method:  "POST",
			address: "example.com",
			headers: map[string]string{
				"Origin": "https://example.com/",
			},
		},
		{
			name:    "Success Referer",
			method:  "POST",
			address: "example.com",
			headers: map[string]string{
				"Referer": "https://example.com/something/else",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.method, target, nil)
			for name, value := range tt.headers {
				request.Header.Add(name, value)
			}
			self := newRequest(target, request)

			if tt.address != "" {
				cfg.Address = tt.address
			}

			err := self.CheckPOST(context.Background())

			if (err == nil) != (len(tt.errorMsg) == 0) {
				if err == nil {
					t.Errorf(`Expected error "%s", got nil.`, tt.errorMsg)
				} else {
					t.Errorf(`Unexpected error "%s".`, err)
				}
			} else if err != nil {
				var httpError HttpError
				ok := errors.As(err, &httpError)
				if !ok {
					t.Errorf("Unexpected error type for %v.", err)
				} else if httpError.Msg != tt.errorMsg {
					t.Errorf(`Wrong error message. Got "%s". Expect "%s".`, httpError.Msg, tt.errorMsg)
				}
			}
		})
	}
}

func TestRequest_UnmarshalJSONBody(t *testing.T) {
	type myStruct struct {
		Int    int
		String string
		Float  float64
		Slice  []int
	}
	expected := myStruct{Int: 42, String: "foo", Float: 3.14, Slice: []int{1, 2, 3, 5, 8, 13}}

	body, err := json.Marshal(expected)
	if err != nil {
		t.Fatal(err)
	}
	req := Request{original: httptest.NewRequest("GET", "/foo", bytes.NewBuffer(body))}

	var got myStruct
	err = req.UnmarshalJSONBody(&got)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("Got %v. Expect %v.", got, expected)
	}
}

func TestRequest_UnmarshalJSONBody_repeat(t *testing.T) {
	type myStruct struct {
		Int    int
		String string
		Float  float64
		Slice  []int
	}
	sBody := myStruct{Int: 42, String: "foo", Float: 3.14, Slice: []int{1, 2, 3, 5, 8, 13}}

	jBody, err := json.Marshal(sBody)
	if err != nil {
		t.Fatal(err)
	}
	req := Request{original: httptest.NewRequest("GET", "/foo", bytes.NewBuffer(jBody))}

	var got [2]myStruct
	for i := range got {
		err = req.UnmarshalJSONBody(&got[i])
		if err != nil {
			t.Error(err)
		}
	}
	if !reflect.DeepEqual(got[0], got[1]) {
		t.Errorf("Differing values: %v then %v.", got[0], got[1])
	}
}
