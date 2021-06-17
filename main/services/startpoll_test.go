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

package services

import (
	"errors"
	"testing"
	"time"

	"github.com/JBoudou/Itero/mid/db"
	dbt "github.com/JBoudou/Itero/mid/db/dbtest"
	"github.com/JBoudou/Itero/mid/service"
	"github.com/JBoudou/Itero/pkg/events"
	"github.com/JBoudou/Itero/pkg/events/eventstest"
	"github.com/JBoudou/Itero/pkg/ioc"
)

type startPollTestInstance struct {
	name       string
	state      string
	expired    bool // Whether Poll.Start is before now.
	expectNext bool // Whether the poll must be started.
	expectList bool // Whether the poll will have to be started in the future.
}

func metaTestStartPoll(t *testing.T, checker func(*testing.T, *startPollTestInstance, uint32)) {
	t.Parallel()
	const qSetPoll = `UPDATE Polls SET State = ?, Start = ? WHERE Id = ?`

	tests := []startPollTestInstance{
		{
			name:       "Start",
			state:      "Waiting",
			expired:    true,
			expectNext: true,
			expectList: true,
		},
		{
			name:       "Wait",
			state:      "Waiting",
			expired:    false,
			expectNext: false,
			expectList: true,
		},
		{
			name:       "Already active",
			state:      "Active",
			expired:    true,
			expectNext: false,
			expectList: false,
		},
		{
			name:       "Already terminated",
			state:      "Terminated",
			expired:    true,
			expectNext: false,
			expectList: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			env := new(dbt.Env)
			defer env.Close()
			admin := env.CreateUserWith(t.Name())
			poll := env.CreatePoll(tt.name, admin, db.ElectorateAll)
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

			checker(t, &tt, poll)
		})
	}
}

// ProcessOne //

func startPoll_processOne_checker(t *testing.T, tt *startPollTestInstance, poll uint32) {
	const qPollState = `SELECT State FROM Polls WHERE Id = ?`

	locator := ioc.Root.Sub()
	called := false
	locator.Bind(func() events.Manager {
		return &eventstest.ManagerMock{
			T: t,
			Send_: func(evt events.Event) error {
				if startEvent, ok := evt.(StartPollEvent); ok && startEvent.Poll == poll {
					called = true
				}
				return nil
			},
		}
	})

	var svc service.Service
	mustt(t, locator.Inject(StartPollService, &svc))
	err := svc.ProcessOne(poll)

	nothingToDoYet := false
	if errors.Is(err, service.NothingToDoYet) {
		nothingToDoYet = true
		err = nil
	}
	mustt(t, err)

	if tt.expectNext == nothingToDoYet {
		if tt.expectNext {
			t.Errorf("Expect start. Got NothingToDoYet.")
		} else {
			t.Errorf("Expect nothing to be done, but nil returned.")
		}
	}

	expectState := tt.state
	if tt.expectNext {
		expectState = "Active"
	}
	var gotState string
	row := db.DB.QueryRow(qPollState, poll)
	mustt(t, row.Scan(&gotState))
	if gotState != expectState {
		t.Errorf("Wrong state. Got %s. Expect %s.", gotState, expectState)
	}

	if called != tt.expectNext {
		if tt.expectNext {
			t.Errorf("StartPollEvent not received.")
		} else {
			t.Errorf("StartPollEvent received.")
		}
	}
}

func TestStartPollService_ProcessOne(t *testing.T) {
	metaTestStartPoll(t, startPoll_processOne_checker)
}

// CheckAll //

func startPoll_CheckAll_checker(t *testing.T, tt *startPollTestInstance, poll uint32) {
	var svc service.Service
	mustt(t, ioc.Root.Inject(StartPollService, &svc))

	iterator := svc.CheckAll()
	defer iterator.Close()

	listed := idDateIteratorHasId(t, iterator, poll)
	if listed != tt.expectList {
		if tt.expectList {
			t.Errorf("Poll not listed when it should.")
		} else {
			t.Errorf("Poll listed when it shouldn't.")
		}
	}
}

func TestStartPollService_CheckAll(t *testing.T) {
	metaTestStartPoll(t, startPoll_CheckAll_checker)
}

// CheckOne //

func startPoll_CheckOne_checker(t *testing.T, tt *startPollTestInstance, poll uint32) {
	var svc service.Service
	mustt(t, ioc.Root.Inject(StartPollService, &svc))

	got := svc.CheckOne(poll)

	isZero := got.IsZero()
	if isZero == tt.expectList {
		if tt.expectList {
			t.Errorf("Expect a date. Got Zero")
		} else {
			t.Errorf("Expect zero. Got %v.", got)
		}
	}
}

func TestStartPollService_CheckOne(t *testing.T) {
	metaTestStartPoll(t, startPoll_CheckOne_checker)
}

// events //

func TestStartPollService_Events(t *testing.T) {
	tests := []checkEventScheduleTest {
		{
			name: "CreatePollEvent",
			event: CreatePollEvent{42},
			schedule: []uint32{42},
		},
		{
			name: "StartPollEvent",
			event: StartPollEvent{42},
		},
	}
	checkEventSchedule(t, tests, StartPollService)
}
