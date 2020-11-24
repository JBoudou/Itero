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

// Package server provides classes and functions for the HTTP server side of the middleware.
//
// In particular, the package handles client sessions by producing credentials for logged user and
// by verifying these credentials for each request.
//
// It is a wrapper around net/http.
package server

import (
	"log"
	"net/http"
	"os"

	"github.com/JBoudou/Itero/config"

	gs "github.com/gorilla/sessions"
)

const (
	// Max age of the cookies in seconds, and to compute the deadline.
	sessionMaxAge = 7200

	// Additional delay accorded after deadline is reached.
	sessionGraceTime = 20

	sessionName = "s"

	sessionKeySessionId = "sid"
	sessionKeyUserName  = "usr"
	sessionKeyUserId    = "uid"
	sessionKeyDeadline  = "dl"

	queryKeySessionId = "s"

	wwwroot = "app/dist/app"
)

var (
	cfg           myConfig
	sessionStore  *gs.CookieStore
)

// Whether the package is usable. May be false if there is no configuration for the package.
var Ok bool

// SessionOptions reflects the configured options for sessions.
// Modifying it has no effect on the sessions generated by the package.
var SessionOptions gs.Options

type myConfig struct {
	Address     string
	CertFile    string
	KeyFile     string
	SessionKeys [][]byte
}

func init() {
	// Configuration
	cfg.Address = ":8080"
	if err := config.Value("server", &cfg); err != nil {
		log.Print(err)
		log.Println("WARNING: Package server not usable because there is no configuration for it.")
		Ok = false
		return
	}
	Ok = true

	// Session
	sessionStore = gs.NewCookieStore(cfg.SessionKeys...)
	sessionStore.MaxAge(sessionMaxAge)
	SessionOptions = *sessionStore.Options
}

// User represents a logged user.
type User struct {
	Name string
	Id   uint32
}

type oneFile struct {
	path string
}

func (self *oneFile) Open(name string) (http.File, error) {
	log.Printf("Redirect for: %s", name)
	return os.Open(self.path)
}

// Start the server.
// Parameters are taken from the configuration.
func Start() (err error) {
	redirect := oneFile{wwwroot + "/index.html"}
	http.Handle("/r/", http.FileServer(&redirect))
	http.Handle("/", http.FileServer(http.Dir(wwwroot)))
	http.Handle("/s/", http.StripPrefix("/s/", http.FileServer(http.Dir("static"))))

	if cfg.CertFile == "" && cfg.KeyFile == "" {
		log.Println("WARNING: The server will be launched in HTTP mode, which is INSECURE.")
		log.Println("WARNING: Set server.CertFile and server.KeyFile in the configuration.")
		err = http.ListenAndServe(cfg.Address, nil)
	} else {
		err = http.ListenAndServeTLS(cfg.Address, cfg.CertFile, cfg.KeyFile, nil)
	}
	return
}

// SessionKeys retrieves the session keys for test purpose.
//
// This is a low level function, made available for tests.
func SessionKeys() [][]byte {
	return cfg.SessionKeys
}