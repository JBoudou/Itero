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
	"fmt"
	"testing"

	"github.com/JBoudou/Itero/db"
	dbt "github.com/JBoudou/Itero/db/dbtest"
	"github.com/JBoudou/Itero/events"
)

// TODO: We should have that in an eventstest package.
type eventsReceiverMock struct {
	t       *testing.T
	receive func(events.Event)
	_close  func()
}

func (self eventsReceiverMock) Receive(evt events.Event) {
	if self.receive == nil {
		self.t.Fatal("Receive unexpectedly called.")
	}
	self.receive(evt)
}

func (self eventsReceiverMock) Close() {
	if self._close == nil {
		self.t.Fatal("Close unexpectedly called.")
	}
	self._close()
}

func TestRoundCheckAllPolls_Close(t *testing.T) {
	const (
		qSetMinMax       = `UPDATE Polls SET MinNbRounds = 1, MaxNbRounds = 2 WHERE Id = ?`
		qSetRound        = `UPDATE Polls SET CurrentRound = ? WHERE Id = ?`
		qSetExpiredPoll  = `UPDATE Polls SET Deadline = CURRENT_TIMESTAMP() WHERE Id = ?`
		qSetExpiredRound = `
		  UPDATE Polls
		     SET CurrentRoundStart = SUBTIME(CURRENT_TIMESTAMP(), MaxRoundDuration)
		   WHERE Id = ?`
		qIsActive = `SELECT Active FROM Polls WHERE Id = ?`
	)

	// Tests are independent.
	// A poll is created with MinNbRound = 1 and MaxNbRound = 2.
	tests := []struct {
		round        int  // CurrentRound
		expiredPoll  bool // whether Deadline >= CURRENT_TIMESTAMP()
		expiredRound bool // whether ADDTIME(CurrentRoundStart, MaxRoundDuration) >= CURRENT_TIMESTAMP
		expectClosed bool
	}{
		// This test is exhaustive.
		{
			round:        0,
			expiredPoll:  false,
			expiredRound: false,
			expectClosed: false,
		},
		{
			round:        0,
			expiredPoll:  false,
			expiredRound: true,
			expectClosed: false,
		},
		{
			round:        0,
			expiredPoll:  true,
			expiredRound: false,
			expectClosed: false,
		},
		{
			round:        0,
			expiredPoll:  true,
			expiredRound: true,
			expectClosed: true,
		},
		{
			round:        1,
			expiredPoll:  false,
			expiredRound: false,
			expectClosed: false,
		},
		{
			round:        1,
			expiredPoll:  false,
			expiredRound: true,
			expectClosed: true,
		},
		{
			round:        1,
			expiredPoll:  true,
			expiredRound: false,
			expectClosed: true,
		},
		{
			round:        1,
			expiredPoll:  true,
			expiredRound: true,
			expectClosed: true,
		},
		{
			round:        2,
			expiredPoll:  false,
			expiredRound: false,
			expectClosed: true,
		},
		{
			round:        2,
			expiredPoll:  false,
			expiredRound: true,
			expectClosed: true,
		},
		{
			round:        2,
			expiredPoll:  true,
			expiredRound: false,
			expectClosed: true,
		},
		{
			round:        2,
			expiredPoll:  true,
			expiredRound: true,
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
			if err == nil && tt.expiredRound {
				_, err = db.DB.Exec(qSetExpiredRound, pollId)
			}
			mustt(t, err)

			originalManager := events.DefaultManager
			events.DefaultManager = events.NewAsyncManager(4)
			closed := false
			synchro := make(chan bool)
			events.AddReceiver(eventsReceiverMock{
				t: t,
				receive: func(evt events.Event) {
					if closeEvent, ok := evt.(ClosePollEvent); ok && closeEvent.Poll == pollId {
						closed = true
					}
				},
				_close: func() {
					synchro <- true
				},
			})

			mustt(t, roundCheckAllPolls())

			fakeManager := events.DefaultManager
			events.DefaultManager = originalManager
			fakeManager.Close()
			<-synchro

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
		})
	}
}
