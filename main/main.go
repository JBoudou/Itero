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
	"log"

	"github.com/JBoudou/Itero/mid/server"
	"github.com/JBoudou/Itero/mid/service"
	. "github.com/JBoudou/Itero/main/handlers"
	. "github.com/JBoudou/Itero/main/services"
	"github.com/JBoudou/Itero/pkg/ioc"
	"github.com/JBoudou/Itero/pkg/events"
)

func init() {
	// Events
	// TODO: do not have events.DefaultManager
	ioc.Root.Set(func () events.Manager { return events.DefaultManager })
}

func StartHandler(url string, constructor interface{}, interceptors ...server.Interceptor) {
	var handler server.Handler
	err := ioc.Root.Inject(constructor, &handler)
	if err != nil {
		panic(err)
	}
	server.Handle(url, handler, interceptors...)
}

func main() {
	// Services
	service.Run(StartPollService)
	service.Run(NextRoundService)
	service.Run(ClosePollService)
	ioc.Root.Get(StartEmailService)

	// Misc
	RunPollNotif(PollNotifDelay)

	// Handlers
	server.HandleFunc("/a/login", LoginHandler)
	server.HandleFunc("/a/signup", SignupHandler)
	server.HandleFunc("/a/refresh", RefreshHandler)
	server.HandleFunc("/a/list", ListHandler, server.Compress)
	server.HandleFunc("/a/poll/", PollHandler)
	server.HandleFunc("/a/ballot/uninominal/", UninominalBallotHandler, server.Compress)
	server.HandleFunc("/a/vote/uninominal/", UninominalVoteHandler)
	server.HandleFunc("/a/info/count/", CountInfoHandler, server.Compress)
	server.HandleFunc("/a/create", CreateHandler)
	server.HandleFunc("/a/delete/", DeleteHandler)
	server.HandleFunc("/a/pollnotif", PollNotifHandler, server.Compress)
	server.HandleFunc("/a/config", ConfigHandler)
	StartHandler("/a/confirm/", ConfirmHandler)

	log.Println("Server starting")
	if err := server.Start(); err != nil {
		log.Fatal(err)
	}
	log.Println("Server terminated")
}
