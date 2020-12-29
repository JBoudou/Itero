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
)

func TestNewRequest(t *testing.T) {
	precheck(t)

	type args struct {
		basePattern string
		original    *http.Request
	}
	tests := []struct {
		name   string
		args   args
		expect Request
	}{
		{
			name: "Basic",
			args: args{
				basePattern: "/foo/bar",
				original:    httptest.NewRequest("GET", "/foo/bar/baz/last", nil),
			},
			expect: Request{
				SessionError:  nil, // No session error when there is no cookie
				FullPath:      []string{"foo", "bar", "baz", "last"},
				RemainingPath: []string{"baz", "last"},
				// original is set to args.original during the execution
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newRequest(tt.args.basePattern, tt.args.original)
			tt.expect.original = tt.args.original
			if !reflect.DeepEqual(*got, tt.expect) {
				t.Errorf("Got %v. Expect %v", got, tt.expect)
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
	user := User{Name: "John", Id: 42}

	addCorrectSession := func(t *testing.T, result *http.Response, request *http.Request) {
		var sessionID string
		var buff bytes.Buffer
		if _, err := buff.ReadFrom(result.Body); err != nil {
			t.Fatalf("Unable to read response body: %s", err)
		}
		if err := json.Unmarshal(buff.Bytes(), &sessionID); err != nil {
			t.Fatalf("Unable to convert response body: %s", err)
		}

		AddSessionIdToRequest(request, sessionID)
	}

	addForgedSession := func(value string) func(t *testing.T, result *http.Response,
		request *http.Request) {
		return func(t *testing.T, result *http.Response, request *http.Request) {
			AddSessionIdToRequest(request, value)
		}
	}

	checkSuccess := func(t *testing.T, got *Request, original *http.Request) {
		expect := Request{
			User:          &user,
			FullPath:      []string{singlePath},
			RemainingPath: []string{},
			original:      original,
		}
		if !reflect.DeepEqual(*got, expect) {
			t.Errorf("Got %v. Expect %v.", got, expect)
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
			response.SendLoginAccepted(context.Background(), user, &Request{original: &http.Request{}})
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
			name:   "Wrong origin host",
			method: "POST",
			address: "example.com",
			headers: map[string]string{
				"Origin": "https://impossible.dummy.address/",
			},
			errorMsg: "Unauthorized",
		},
		{
			name:   "Wrong origin scheme",
			method: "POST",
			address: "example.com",
			headers: map[string]string{
				"Origin": "http://example.com/",
			},
			errorMsg: "Unauthorized",
		},
		{
			name:   "Wrong port",
			method: "POST",
			address: "example.com:17",
			headers: map[string]string{
				"Origin": "https://example.com:18/",
			},
			errorMsg: "Unauthorized",
		},
		{
			name:   "Wrong port default 1",
			method: "POST",
			address: "example.com",
			headers: map[string]string{
				"Origin": "https://example.com:18/",
			},
			errorMsg: "Unauthorized",
		},
		{
			name:   "Wrong port default 2",
			method: "POST",
			address: "example.com:42",
			headers: map[string]string{
				"Origin": "https://example.com/",
			},
			errorMsg: "Unauthorized",
		},
		{
			name:   "Good port default 1",
			method: "POST",
			address: "example.com",
			headers: map[string]string{
				"Origin": "https://example.com:443/",
			},
		},
		{
			name:   "Good port default 2",
			method: "POST",
			address: "example.com:443",
			headers: map[string]string{
				"Origin": "https://example.com/",
			},
		},
		{
			name:   "Good port explicit",
			method: "POST",
			address: "example.com:42",
			headers: map[string]string{
				"Origin": "https://example.com:42/",
			},
		},
		{
			name:   "Success Origin",
			method: "POST",
			address: "example.com",
			headers: map[string]string{
				"Origin": "https://example.com/",
			},
		},
		{
			name:   "Success Referer",
			method: "POST",
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
				} else if httpError.msg != tt.errorMsg {
					t.Errorf(`Wrong error message. Got "%s". Expect "%s".`, httpError.msg, tt.errorMsg)
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
