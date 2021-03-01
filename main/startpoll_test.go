// Itero - Online iterative vote application
// Copyright (C) 2021 Joseph Boudou
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
	"testing"
	"time"

	"github.com/JBoudou/Itero/db"
	dbt "github.com/JBoudou/Itero/db/dbtest"
	"github.com/JBoudou/Itero/events"
	"github.com/JBoudou/Itero/events/eventstest"
)

func TestStartPoll_fullCheck(t *testing.T) {
	const (
		qSetPoll   = `UPDATE Polls SET State = ?, Start = ? WHERE Id = ?`
		qPollState = `SELECT State FROM Polls WHERE Id = ?`
	)

	tests := []struct {
		name    string
		state   string
		expired bool // Whether Poll.Start is before now.
		expect  bool // Whether the poll must be started.
	}{
		{
			name:    "Start",
			state:   "Waiting",
			expired: true,
			expect:  true,
		},
		{
			name:    "Wait",
			state:   "Waiting",
			expired: false,
			expect:  false,
		},
		{
			name:    "Already active",
			state:   "Active",
			expired: true,
			expect:  false,
		},
		{
			name:    "Already terminated",
			state:   "Terminated",
			expired: true,
			expect:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := new(dbt.Env)
			defer env.Close()
			admin := env.CreateUser()
			poll := env.CreatePoll(tt.name, admin, db.PollPublicityPublic)
			env.Must(t)
			var err error

			startDate := time.Now()
			if tt.expired {
				startDate = startDate.Add(-time.Minute)
			} else {
				startDate = startDate.Add(time.Minute)
			}
			_, err = db.DB.Exec(qSetPoll, tt.state, startDate, poll)
			mustt(t, err)

			// TODO: Implement a better mock to prevent copy/paste.
			originalManager := events.DefaultManager
			called := false
			events.DefaultManager = &eventstest.ManagerMock{
				T: t,
				Send_: func(evt events.Event) error {
					if startEvent, ok := evt.(StartPollEvent); ok && startEvent.Poll == poll {
						called = true
					}
					return nil
				},
			}

			mustt(t, newStartPoll().fullCheck())

			events.DefaultManager = originalManager

			expectState := tt.state
			if tt.expect {
				expectState = "Active"
			}
			var gotState string
			row := db.DB.QueryRow(qPollState, poll)
			mustt(t, row.Scan(&gotState))
			if gotState != expectState {
				t.Errorf("Wrong state. Got %s. Expect %s.", gotState, expectState)
			}

			if called != tt.expect {
				if tt.expect {
					t.Errorf("StartPollEvent not received.")
				} else {
					t.Errorf("StartPollEvent received.")
				}
			}
		})
	}
}

func TestStartPoll_checkOne(t *testing.T) {
	const (
		qSetPoll   = `UPDATE Polls SET State = ?, Start = ? WHERE Id = ?`
		qPollState = `SELECT State FROM Polls WHERE Id = ?`
	)

	tests := []struct {
		name    string
		state   string
		expired bool // Whether Poll.Start is before now.
		expect  bool // Whether the poll must be started.
	}{
		{
			name:    "Start",
			state:   "Waiting",
			expired: true,
			expect:  true,
		},
		{
			name:    "Wait",
			state:   "Waiting",
			expired: false,
			expect:  false,
		},
		{
			name:    "Already active",
			state:   "Active",
			expired: true,
			expect:  false,
		},
		{
			name:    "Already terminated",
			state:   "Terminated",
			expired: true,
			expect:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := new(dbt.Env)
			defer env.Close()
			admin := env.CreateUser()
			poll := env.CreatePoll(tt.name, admin, db.PollPublicityPublic)
			env.Must(t)
			var err error

			startDate := time.Now()
			if tt.expired {
				startDate = startDate.Add(-time.Minute)
			} else {
				startDate = startDate.Add(time.Minute)
			}
			_, err = db.DB.Exec(qSetPoll, tt.state, startDate, poll)
			mustt(t, err)

			// TODO: Implement a better mock to prevent copy/paste.
			originalManager := events.DefaultManager
			called := false
			events.DefaultManager = &eventstest.ManagerMock{
				T: t,
				Send_: func(evt events.Event) error {
					if startEvent, ok := evt.(StartPollEvent); ok && startEvent.Poll == poll {
						called = true
					}
					return nil
				},
			}

			mustt(t, newStartPoll().checkOne(poll))

			events.DefaultManager = originalManager

			expectState := tt.state
			if tt.expect {
				expectState = "Active"
			}
			var gotState string
			row := db.DB.QueryRow(qPollState, poll)
			mustt(t, row.Scan(&gotState))
			if gotState != expectState {
				t.Errorf("Wrong state. Got %s. Expect %s.", gotState, expectState)
			}

			if called != tt.expect {
				if tt.expect {
					t.Errorf("StartPollEvent not received.")
				} else {
					t.Errorf("StartPollEvent received.")
				}
			}
		})
	}
}
