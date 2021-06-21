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

type ResponseSpy struct {
	Backend     server.Response
	T           *testing.T
	JsonFct     func(*testing.T, context.Context, interface{})
	ErrorFct    func(*testing.T, context.Context, error)
	LoginFct    func(*testing.T, context.Context, server.User, *server.Request, interface{})
	UnloggedFct func(*testing.T, context.Context, server.User, *server.Request) error
}

func (self ResponseSpy) SendJSON(ctx context.Context, data interface{}) {
	self.T.Helper()
	if self.JsonFct != nil {
		self.JsonFct(self.T, ctx, data)
	}
	self.Backend.SendJSON(ctx, data)
}

func (self ResponseSpy) SendError(ctx context.Context, err error) {
	self.T.Helper()
	if self.ErrorFct != nil {
		self.ErrorFct(self.T, ctx, err)
	}
	self.Backend.SendError(ctx, err)
}

func (self ResponseSpy) SendLoginAccepted(ctx context.Context, user server.User,
	request *server.Request, profile interface{}) {

	self.T.Helper()
	if self.LoginFct != nil {
		self.LoginFct(self.T, ctx, user, request, profile)
	}
	self.Backend.SendLoginAccepted(ctx, user, request, profile)
}

func (self ResponseSpy) SendUnloggedId(ctx context.Context, user server.User,
	request *server.Request) error {

	self.T.Helper()
	if self.UnloggedFct != nil {
		return self.UnloggedFct(self.T, ctx, user, request)
	}
	self.Backend.SendUnloggedId(ctx, user, request)
}
