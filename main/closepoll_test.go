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
	"fmt"
	"testing"

	"github.com/JBoudou/Itero/db"
	dbt "github.com/JBoudou/Itero/db/dbtest"
	"github.com/JBoudou/Itero/events"
	"github.com/JBoudou/Itero/events/eventstest"
	"github.com/JBoudou/Itero/service"
)

type closePollTestInstance struct {
	round        int  // CurrentRound
	expiredPoll  bool // whether Deadline >= CURRENT_TIMESTAMP()
	expectClosed bool
}

func metaTestClosePoll(t *testing.T, checker func(*testing.T, *closePollTestInstance, uint32)) {
	const (
		qSetMinMax      = `UPDATE Polls SET MinNbRounds = 1, MaxNbRounds = 2 WHERE Id = ?`
		qSetRound       = `UPDATE Polls SET CurrentRound = ? WHERE Id = ?`
		qSetExpiredPoll = `UPDATE Polls SET Deadline = CURRENT_TIMESTAMP() WHERE Id = ?`
	)

	// Tests are independent.
	// A poll is created with MinNbRound = 1 and MaxNbRound = 2.
	tests := []closePollTestInstance{
		{
			round:        0,
			expiredPoll:  false,
			expectClosed: false,
		},
		{
			round:        0,
			expiredPoll:  true,
			expectClosed: false,
		},
		{
			round:        1,
			expiredPoll:  false,
			expectClosed: false,
		},
		{
			round:        1,
			expiredPoll:  true,
			expectClosed: true,
		},
		{
			round:        2,
			expiredPoll:  false,
			expectClosed: true,
		},
		{
			round:        2,
			expiredPoll:  true,
			expectClosed: true,
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprint(tt), func(t *testing.T) {
			env := new(dbt.Env)
			defer env.Close()
			userId := env.CreateUser()
			pollId := env.CreatePoll("TestRoundCheckAllPolls_Close", userId, db.PollPublicityPublic)
			env.Must(t)

			var err error
			_, err = db.DB.Exec(qSetMinMax, pollId)
			mustt(t, err)
			_, err = db.DB.Exec(qSetRound, tt.round, pollId)
			if err == nil && tt.expiredPoll {
				_, err = db.DB.Exec(qSetExpiredPoll, pollId)
			}
			mustt(t, err)

			checker(t, &tt, pollId)
		})
	}
}

// ProcessOne //

func closePoll_processOne_checker(t *testing.T, tt *closePollTestInstance, pollId uint32) {
	const qIsActive = `SELECT State = 'Active' FROM Polls WHERE Id = ?`

	originalManager := events.DefaultManager
	closed := false
	events.DefaultManager = &eventstest.ManagerMock{
		T: t,
		Send_: func(evt events.Event) error {
			if closeEvent, ok := evt.(ClosePollEvent); ok && closeEvent.Poll == pollId {
				closed = true
			}
			return nil
		},
	}

	err := ClosePollService.ProcessOne(pollId)

	events.DefaultManager = originalManager

	nothingToDoYet := false
	if errors.Is(err, service.NothingToDoYet) {
		nothingToDoYet = true
		err = nil
	}
	mustt(t, err)

	if tt.expectClosed == nothingToDoYet {
		if tt.expectClosed {
			t.Errorf("Expect close. Got NothingToDoYet.")
		} else {
			t.Errorf("Expect nothing to be done, but nil returned.")
		}
	}

	var active bool
	row := db.DB.QueryRow(qIsActive, pollId)
	mustt(t, row.Scan(&active))

	if closed != tt.expectClosed {
		if tt.expectClosed {
			t.Errorf("PollClosedEvent not received.")
		} else {
			t.Errorf("PollClosedEvent received.")
		}
	}
	if active == tt.expectClosed {
		if tt.expectClosed {
			t.Errorf("Poll still active.")
		} else {
			t.Errorf("Poll inactive.")
		}
	}
}

func TestClosePollService_ProcessOne(t *testing.T) {
	metaTestClosePoll(t, closePoll_processOne_checker)
}

// CheckAll //

func closePoll_CheckAll_checker(t *testing.T, tt *closePollTestInstance, poll uint32) {
	iterator := ClosePollService.CheckAll()

	listed := idDateIteratorHasId(t, iterator, poll)
	if listed != tt.expectClosed {
		if tt.expectClosed {
			t.Errorf("Poll not listed when it should.")
		} else {
			t.Errorf("Poll listed when it shouldn't.")
		}
	}
}

func TestClosePollService_CheckAll(t *testing.T) {
	metaTestClosePoll(t, closePoll_CheckAll_checker)
}

// CheckOne //

func closePoll_CheckOne_checker(t *testing.T, tt *closePollTestInstance, poll uint32) {
	got := ClosePollService.CheckOne(poll)

	isZero := got.IsZero()
	if isZero == tt.expectClosed {
		if tt.expectClosed {
			t.Errorf("Expect a date. Got Zero")
		} else {
			t.Errorf("Expect zero. Got %v.", got)
		}
	}
}

func TestClosePollService_CheckOne(t *testing.T) {
	metaTestClosePoll(t, closePoll_CheckOne_checker)
}
