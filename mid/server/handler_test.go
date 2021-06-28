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

package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/JBoudou/Itero/mid/root"
	. "github.com/JBoudou/Itero/mid/server"
	srvt "github.com/JBoudou/Itero/mid/server/servertest"
	"github.com/JBoudou/Itero/pkg/slog"
)

func precheck(t *testing.T) {
	if !Ok {
		t.Log("Impossible to test package server_test: there is no configuration.")
		t.Log("Add a configuration file in server/ (may be a link to the main configuration file).")
		t.SkipNow()
	}
}

// checkerJSONString returns a srvt.Checker to check responses whose body is a JSON object
// representing a string.
//
// The returned function checks that the statuc code and the encoded string are as expected.
func checkerJSONString(expectCode int, expectBody string) srvt.Checker {
	return srvt.CheckerFun(func(t *testing.T, response *http.Response, req *Request) {
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
	})
}

type handlerArgs struct {
	pattern string
	fct     func(ctx context.Context, resp Response, req *Request)
}
type handlerTestsStruct struct {
	name    string
	args    handlerArgs
	req     srvt.Request
	checker srvt.Checker
}

func TestHandleFunc(t *testing.T) {
	precheck(t)

	var (
		tWrite  = "/t/write"
		tEcho   = "/t/echo"
		tError  = "/t/error"
		tPanic  = "/t/panic"
		tStruct = "/t/struct"
	)

	handlerTests := []handlerTestsStruct{
		{
			name: "write",
			args: handlerArgs{
				pattern: tWrite,
				fct: func(ctx context.Context, resp Response, req *Request) {
					msg := "bar"
					resp.SendJSON(ctx, msg)
				},
			},
			req: srvt.Request{
				Method: "GET",
				Target: &tWrite,
			},
			checker: checkerJSONString(http.StatusOK, "bar"),
		},
		{
			name: "echo",
			args: handlerArgs{
				pattern: tEcho,
				fct: func(ctx context.Context, resp Response, req *Request) {
					var msg string
					if err := req.UnmarshalJSONBody(&msg); err != nil {
						resp.SendError(ctx, err)
						return
					}
					resp.SendJSON(ctx, msg)
				},
			},
			req: srvt.Request{
				Method: "POST",
				Target: &tEcho,
				Body:   `"Hello"`,
			},
			checker: checkerJSONString(http.StatusOK, "Hello"),
		},
		{
			name: "error",
			args: handlerArgs{
				pattern: tError,
				fct: func(ctx context.Context, resp Response, req *Request) {
					resp.SendError(ctx, NewHttpError(http.StatusPaymentRequired, "Flublu", "Test"))
				},
			},
			req: srvt.Request{
				Method: "GET",
				Target: &tError,
			},
			checker: srvt.CheckError{http.StatusPaymentRequired, "Flublu"},
		},
		{
			name: "panic",
			args: handlerArgs{
				pattern: tPanic,
				fct: func(ctx context.Context, resp Response, req *Request) {
					panic(NewHttpError(http.StatusPaymentRequired, "Barbaz", "Test"))
				},
			},
			req: srvt.Request{
				Method: "GET",
				Target: &tPanic,
			},
			checker: srvt.CheckError{http.StatusPaymentRequired, "Barbaz"},
		},
		{
			name: "struct",
			args: handlerArgs{
				pattern: tStruct,
				fct: func(ctx context.Context, resp Response, req *Request) {
					resp.SendJSON(ctx, struct {
						A int
						B string
					}{A: 42, B: "Foobar"})
				},
			},
			req: srvt.Request{
				Method: "GET",
				Target: &tStruct,
			},
			checker: srvt.CheckJSON{Body: struct {
				A int
				B string
			}{A: 42, B: "Foobar"}},
		},
	}

	root.IoC.Bind(func() slog.Printer { return log.New(os.Stderr, "", log.LstdFlags) })

	for _, tt := range handlerTests {
		t.Run(tt.name, func(t *testing.T) {
			HandleFunc(tt.args.pattern, tt.args.fct)

			wr := httptest.NewRecorder()
			req, err := tt.req.Make(t)
			if err != nil {
				t.Fatal(err)
			}
			http.DefaultServeMux.ServeHTTP(wr, req)

			// TODO: replace the dummy request with a better one.
			tt.checker.Check(t, wr.Result(), &Request{})
		})
	}
}
