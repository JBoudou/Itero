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
	"context"
	"testing"

	"github.com/JBoudou/Itero/mid/server"
)

type MockResponse struct {
	T           *testing.T
	JsonFct     func(*testing.T, context.Context, interface{})
	ErrorFct    func(*testing.T, context.Context, error)
	LoginFct    func(*testing.T, context.Context, server.User, *server.Request, interface{})
	UnloggedFct func(*testing.T, context.Context, server.User, *server.Request) error
}

func (self MockResponse) SendJSON(ctx context.Context, data interface{}) {
	self.T.Helper()
	if self.JsonFct == nil {
		self.T.Errorf("SendJSON called with data %v", data)
		return
	}
	self.JsonFct(self.T, ctx, data)
}

func (self MockResponse) SendError(ctx context.Context, err error) {
	self.T.Helper()
	if self.ErrorFct == nil {
		self.T.Errorf("SendError called with error %s", err)
		return
	}
	self.ErrorFct(self.T, ctx, err)
}

func (self MockResponse) SendLoginAccepted(ctx context.Context, user server.User,
	request *server.Request, profile interface{}) {

	self.T.Helper()
	if self.LoginFct == nil {
		self.T.Errorf("SendLoginAccepted called with user %v", user)
		return
	}
	self.LoginFct(self.T, ctx, user, request, profile)
}

func (self MockResponse) SendUnloggedId(ctx context.Context, user server.User,
	request *server.Request) error {

	self.T.Helper()
	if self.UnloggedFct == nil {
		self.T.Errorf("SendUnloggedId called with user %v", user)
		return nil
	}
	return self.UnloggedFct(self.T, ctx, user, request)
}
