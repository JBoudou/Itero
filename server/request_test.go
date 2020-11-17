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
				original: httptest.NewRequest("GET", "/foo/bar/baz/last", nil),
			},
			expect: Request{
				SessionError: nil, // No session error when there is no cookie
				FullPath: []string{ "foo", "bar", "baz", "last" },
				RemainingPath: []string{ "baz", "last" },
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
	// TODO: Refactor this test.
	//			 Add failing ones (in particular when there is no sessionId or a wrong one).
	user := User{Name: "foo", Id: 42}
	mock := httptest.NewRecorder()

	response := Response{Writer: mock}
	response.SendLoginAccepted(context.Background(), user, &Request{original: &http.Request{}})

	var sessionID string
	var buff bytes.Buffer
	if _, err := buff.ReadFrom(mock.Result().Body); err != nil {
		t.Fatalf("Unable to read response body: %s", err)
	}
	if err := json.Unmarshal(buff.Bytes(), &sessionID); err != nil {
		t.Fatalf("Unable to convert response body: %s", err)
	}

	query := url.Values{}
	query.Add(queryKeySessionId, sessionID)
	reqURL := "/foo?" + query.Encode()

	originalRequest := httptest.NewRequest("GET", reqURL, nil)
	for _, cookie := range mock.Result().Cookies() {
		originalRequest.AddCookie(cookie)
	}

	got := newRequest("/foo", originalRequest)
	expect := Request{
		User: &user,
		FullPath: []string{ "foo" },
		RemainingPath: []string{},
		original: originalRequest,
	}
	if !reflect.DeepEqual(got, expect) {
		t.Fatalf("Got %v. Expect %v", got, expect)
	}
}

func TestRequest_UnmarshalJSONBody(t *testing.T) {
	type myStruct struct {
		Int     int
		String  string
		Float   float64
		Slice   []int
	}
	expected := myStruct{Int: 42, String: "foo", Float: 3.14, Slice: []int{ 1, 2, 3, 5, 8, 13 }}

	body, err := json.Marshal(expected)
	if err != nil {
		t.Fatal(err)
	}
	req := Request{ original: httptest.NewRequest("GET", "/foo", bytes.NewBuffer(body)) }

	var got myStruct
	err = req.UnmarshalJSONBody(&got)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("Got %v. Expect %v.", got, expected)
	}
}
