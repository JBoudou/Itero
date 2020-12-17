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
	"database/sql"
	"fmt"
	"strconv"
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
					if closeEvent, ok := evt.(PollClosedEvent); ok && closeEvent.Poll == pollId {
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

func TestRoundCheckAllPolls_Next(t *testing.T) {
	const (
		nbParticipants = 2

		qAddParticipants = `INSERT INTO Participants(Poll, User) VALUE (?,?)`
		qUpdatePoll      = `UPDATE Polls SET CurrentRound = ?, RoundThreshold = ? WHERE Id = ?`
		qSetExpiredRound = `
		  UPDATE Polls
		     SET CurrentRoundStart = SUBTIME(CURRENT_TIMESTAMP(), MaxRoundDuration)
		   WHERE Id = ?`
		qSetInvited = `UPDATE Polls SET Publicity = ? WHERE Id = ?`
		qAddVoters  = `UPDATE Participants SET LastRound = ? WHERE User = ? AND Poll = ?`
		qGetRound   = `SELECT CurrentRound FROM Polls WHERE Id = ?`
	)

	// Tests are independent.
	tests := []struct {
		name       string
		round      uint8   // CurrentRound
		expired    bool    // whether ADDTIME(CurrentRoundStart, MaxRoundDuration) >= CURRENT_TIMESTAMP
		pubInvited bool    // whether Publicity is Invited
		threshold  float64 // RoundThreshold
		nbVoter    int     // number of Participant with LastRound = Poll.CurrentRound
		expectNext bool
	}{
		{
			name: "Default",
		},
		{
			name: "Expired",
			expired: true,
			expectNext: true,
		},
		{
			name: "Threshold zero",
			round: 1,
			threshold: 0,
			nbVoter: 1,
			expectNext: true,
		},
		{
			name: "Threshold one",
			round: 1,
			threshold: 1,
			nbVoter: 2,
			expectNext: true,
		},
		{
			name: "Invited",
			pubInvited: true,
			threshold: 1,
			nbVoter: 2,
			expectNext: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := new(dbt.Env)
			defer env.Close()
			var user [nbParticipants]uint32
			for i := range user {
				user[i] = env.CreateUserWith(strconv.FormatInt(int64(i), 10))
			}
			pollId := env.CreatePoll("TestRoundCheckAllPolls_Next", user[0], db.PollPublicityPublic)
			env.Must(t)

			var err error
			var stmt *sql.Stmt

			stmt, err = db.DB.Prepare(qAddParticipants)
			mustt(t, err)
			for _, id := range user {
				_, err = stmt.Exec(pollId, id)
				mustt(t, err)
			}
			mustt(t, stmt.Close())

			_, err = db.DB.Exec(qUpdatePoll, tt.round, tt.threshold, pollId)
			if err == nil && tt.expired {
				_, err = db.DB.Exec(qSetExpiredRound, pollId)
			}
			if err == nil && tt.pubInvited {
				_, err = db.DB.Exec(qSetInvited, db.PollPublicityInvited, pollId)
			}
			if err == nil && tt.nbVoter > 0 {
				stmt, err = db.DB.Prepare(qAddVoters)
				mustt(t, err)
				for _, id := range user {
					_, err = stmt.Exec(tt.round, id, pollId)
					mustt(t, err)
				}
				err = stmt.Close()
			}
			mustt(t, err)

			originalManager := events.DefaultManager
			events.DefaultManager = events.NewAsyncManager(4)
			incremented := false
			synchro := make(chan bool)
			events.AddReceiver(eventsReceiverMock{
				t: t,
				receive: func(evt events.Event) {
					if nextEvent, ok := evt.(NextRoundEvent); ok && nextEvent.Poll == pollId {
						incremented = true
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

			var gotRound uint8
			row := db.DB.QueryRow(qGetRound, pollId)
			mustt(t, row.Scan(&gotRound))
			expectRound := tt.round
			if tt.expectNext {
				expectRound += 1
			}

			if incremented != tt.expectNext {
				if tt.expectNext {
					t.Errorf("NextRoundEvent not received.")
				} else {
					t.Errorf("NextRoundEvent received.")
				}
			}
			if gotRound != expectRound {
				t.Errorf("Wrong round. Got %d. Expect %d.", gotRound, expectRound)
			}
		})
	}
}
