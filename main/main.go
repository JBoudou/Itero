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

package main

import (
	"reflect"

	. "github.com/JBoudou/Itero/main/handlers"
	. "github.com/JBoudou/Itero/main/services"
	"github.com/JBoudou/Itero/mid/root"
	"github.com/JBoudou/Itero/mid/server"
	"github.com/JBoudou/Itero/mid/service"
	"github.com/JBoudou/Itero/pkg/slog"
)

func StartService(serviceFactory interface{}) {
	err := root.IoC.Inject(serviceFactory, service.Run)
	if err != nil {
		panic(err)
	}
}

func StartHandler(url string, fct interface{}, interceptors ...server.Interceptor) {
	if reflect.TypeOf(fct).AssignableTo(reflect.TypeOf(server.HandleFunc).In(1)) {
		var handlerFunc server.HandlerFunc
		reflect.ValueOf(&handlerFunc).Elem().Set(reflect.ValueOf(fct))
		server.HandleFunc(url, handlerFunc, interceptors...)

	} else {
		var handler server.Handler
		err := root.IoC.Inject(fct, &handler)
		if err != nil {
			panic(err)
		}
		server.Handle(url, handler, interceptors...)
	}
}

func main() {
	// Services
	StartService(StartPollService)
	StartService(NextRoundService)
	StartService(ClosePollService)
	StartService(EmailService)

	// Handlers
	StartHandler("/a/login", LoginHandler)
	StartHandler("/a/signup", SignupHandler)
	StartHandler("/a/refresh", RefreshHandler)
	StartHandler("/a/list", ListHandler, server.Compress)
	StartHandler("/a/poll/", PollHandler)
	StartHandler("/a/ballot/uninominal/", UninominalBallotHandler, server.Compress)
	StartHandler("/a/vote/uninominal/", UninominalVoteHandler)
	StartHandler("/a/info/count/", CountInfoHandler, server.Compress)
	StartHandler("/a/create", CreateHandler)
	StartHandler("/a/delete/", DeleteHandler)
	StartHandler("/a/pollnotif", PollNotifHandler, server.Compress)
	StartHandler("/a/config", ConfigHandler)
	StartHandler("/a/confirm/", ConfirmHandler)
	StartHandler("/a/reverify", ReverifyHandler)
	StartHandler("/a/forgot", ForgotHandler)
	StartHandler("/a/passwd/", PasswdHandler)

	var logger slog.Leveled
	root.IoC.Inject(&logger)
	logger.Log("Server starting")
	if err := server.Start(); err != nil {
		logger.Error(err)
	}
	logger.Log("Server terminated")
}
