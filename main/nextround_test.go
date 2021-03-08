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
	"strconv"
	"testing"

	"github.com/JBoudou/Itero/db"
	dbt "github.com/JBoudou/Itero/db/dbtest"
	"github.com/JBoudou/Itero/events"
	"github.com/JBoudou/Itero/events/eventstest"
)

func TestNextRound_fullCheck(t *testing.T) {
	const (
		nbParticipants = 3

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
		expired    bool    // whether RoundDeadline >= CURRENT_TIMESTAMP
		pubInvited bool    // whether Publicity is Invited
		threshold  float64 // RoundThreshold
		nbVoter    int     // number of Participant with LastRound = Poll.CurrentRound
		expectNext bool
	}{
		{
			name: "Default",
		},
		{
			name:       "Expired",
			round:      1,
			expired:    true,
			expectNext: true,
		},
		{
			name:       "Round zero - 2 participants",
			expired:    true,
			nbVoter:    2,
			expectNext: false,
		},
		{
			name:       "Round zero - 3 participants",
			expired:    true,
			nbVoter:    3,
			expectNext: true,
		},
		{
			name:       "Threshold zero",
			round:      1,
			threshold:  0,
			nbVoter:    1,
			expectNext: true,
		},
		{
			name:       "Threshold one",
			round:      1,
			threshold:  1,
			nbVoter:    2,
			expectNext: true,
		},
		{
			name:       "Invited",
			pubInvited: true,
			threshold:  1,
			nbVoter:    2,
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
				for i := 0; i < tt.nbVoter; i++ {
					_, err = stmt.Exec(tt.round, user[i], pollId)
					mustt(t, err)
				}
				err = stmt.Close()
			}
			mustt(t, err)

			originalManager := events.DefaultManager
			incremented := false
			events.DefaultManager = &eventstest.ManagerMock{
				T: t,
				Send_: func(evt events.Event) error {
					if nextEvent, ok := evt.(NextRoundEvent); ok && nextEvent.Poll == pollId {
						incremented = true
					}
					return nil
				},
			}

			mustt(t, newNextRound().fullCheck())

			events.DefaultManager = originalManager

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

func TestNextRound_checkOne(t *testing.T) {
	const (
		nbParticipants = 3

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
		expired    bool    // whether RoundDeadline >= CURRENT_TIMESTAMP
		pubInvited bool    // whether Publicity is Invited
		threshold  float64 // RoundThreshold
		nbVoter    int     // number of Participant with LastRound = Poll.CurrentRound
		expectNext bool
	}{
		{
			name: "Default",
		},
		{
			name:       "Expired",
			round:      1,
			expired:    true,
			expectNext: true,
		},
		{
			name:       "Round zero - 2 participants",
			expired:    true,
			nbVoter:    2,
			expectNext: false,
		},
		{
			name:       "Round zero - 3 participants",
			expired:    true,
			nbVoter:    3,
			expectNext: true,
		},
		{
			name:       "Threshold zero",
			round:      1,
			threshold:  0,
			nbVoter:    1,
			expectNext: true,
		},
		{
			name:       "Threshold one",
			round:      1,
			threshold:  1,
			nbVoter:    2,
			expectNext: true,
		},
		{
			name:       "Invited",
			pubInvited: true,
			threshold:  1,
			nbVoter:    2,
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
				for i := 0; i < tt.nbVoter; i++ {
					_, err = stmt.Exec(tt.round, user[i], pollId)
					mustt(t, err)
				}
				err = stmt.Close()
			}
			mustt(t, err)

			originalManager := events.DefaultManager
			incremented := false
			events.DefaultManager = &eventstest.ManagerMock{
				T: t,
				Send_: func(evt events.Event) error {
					if nextEvent, ok := evt.(NextRoundEvent); ok && nextEvent.Poll == pollId {
						incremented = true
					}
					return nil
				},
			}

			mustt(t, newNextRound().checkOne(pollId))

			events.DefaultManager = originalManager

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
