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
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
)

func TestNewRequest(t *testing.T) {
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
			if !reflect.DeepEqual(got, tt.expect) {
				t.Errorf("Got %v. Expect %v", got, tt.expect)
			}
		})
	}
}

// TestLoginThenNewRequest tests that Response.SendLoginAccepted works well with newRequest.
// This is a "black box test": we don't care what the functions precisely do.
func TestLoginThenNewRequest(t *testing.T) {
	const singlePath = "foo"
	const fullPath = "/" + singlePath
	user := User{Name: "John", Id: 42}

	makeCorrectURL := func(t *testing.T, result *http.Response) string {
		var sessionID string
		var buff bytes.Buffer
		if _, err := buff.ReadFrom(result.Body); err != nil {
			t.Fatalf("Unable to read response body: %s", err)
		}
		if err := json.Unmarshal(buff.Bytes(), &sessionID); err != nil {
			t.Fatalf("Unable to convert response body: %s", err)
		}

		query := url.Values{}
		query.Add(queryKeySessionId, sessionID)
		return fullPath + "?" + query.Encode()
	}

	makeForgedURL := func(value string) func(t *testing.T, result *http.Response) string {
		return func(t *testing.T, result *http.Response) string {
			query := url.Values{}
			query.Add(queryKeySessionId, value)
			return fullPath + "?" + query.Encode()
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
		name    string
		makeUrl func(t *testing.T, result *http.Response) string
		checker func(t *testing.T, got *Request, original *http.Request)
	}{
		{
			name:    "Success",
			makeUrl: makeCorrectURL,
			checker: checkSuccess,
		},
		{
			name:    "No session",
			makeUrl: func(t *testing.T, result *http.Response) string { return fullPath },
			checker: checkFail,
		},
		{
			name:    "Wrong session",
			makeUrl: makeForgedURL(";;;;"),
			checker: checkFail,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := httptest.NewRecorder()

			response := Response{writer: mock}
			response.SendLoginAccepted(context.Background(), user, &Request{original: &http.Request{}})
			result := mock.Result()

			reqURL := tt.makeUrl(t, result)
			originalRequest := httptest.NewRequest("GET", reqURL, nil)
			for _, cookie := range result.Cookies() {
				originalRequest.AddCookie(cookie)
			}

			got := newRequest(fullPath, originalRequest)
			tt.checker(t, &got, originalRequest)
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
