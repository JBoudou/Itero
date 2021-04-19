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
	"errors"
	"log"

	"github.com/JBoudou/Itero/server"
	"github.com/JBoudou/Itero/service"
)

func main() {
	// Services
	service.Run(StartPollService)
	service.Run(NextRoundService)
	service.Run(ClosePollService)

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

	log.Println("Server starting")
	if err := server.Start(); err != nil {
		log.Fatal(err)
	}
	log.Println("Server terminated")
}

// must ensures that err is nil. If it's not, the error is sent by panic, after being wrapped in a
// server.HttpError if it's not already one.
func must(err error) {
	if err != nil {
		var httpError server.HttpError
		if !errors.As(err, &httpError) {
			err = server.InternalHttpError(err)
		}
		panic(err)
	}
}
