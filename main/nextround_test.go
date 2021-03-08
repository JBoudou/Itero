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

func metaTestNextRound(t *testing.T, run func(pollId uint32) error) {
	const (
		nbParticipants = 3

		qAddParticipants = `INSERT INTO Participants(Poll, User) VALUE (?,?)`
		qUpdatePoll      = `UPDATE Polls SET CurrentRound = ?, RoundThreshold = ? WHERE Id = ?`
		qSetMin          = `UPDATE Polls SET MinNbRounds = ? WHERE Id = ?`
		qSetNow          = `
		  UPDATE Polls
		     SET CurrentRoundStart = SUBTIME(CURRENT_TIMESTAMP(), ? * MaxRoundDuration)
		   WHERE Id = ?`
		qSetDeadline = `
		  UPDATE Polls
			   SET Deadline = ADDTIME(CurrentRoundStart, ? * MaxRoundDuration)
			 WHERE Id = ?`
		qSetInvited = `UPDATE Polls SET Publicity = ? WHERE Id = ?`
		qAddVoters  = `UPDATE Participants SET LastRound = ? WHERE User = ? AND Poll = ?`
		qGetRound   = `SELECT CurrentRound FROM Polls WHERE Id = ?`
	)

	// Tests are independent.
	tests := []struct {
		name         string
		round        uint8   // CurrentRound (MaxNbRounds = 3)
		minNbRounds  uint8   // applied only if >2
		nowFact      float32 // if  >0 set Now      = CurrentRoundStart + nowFact      * MaxRoundDuration
		deadlineFact float32 // if !=0 set Deadline = CurrentRoundStart + deadlineFact * MaxRoundDuration
		pubInvited   bool    // whether Publicity is Invited
		threshold    float64 // RoundThreshold
		nbVoter      int     // number of Participant with LastRound = Poll.CurrentRound
		expectNext   bool
	}{
		{
			name: "Default",
		},
		{
			name:       "Round zero - 2 participants",
			nowFact:    1.,
			nbVoter:    2,
			expectNext: false,
		},
		{
			name:       "Round zero - 3 participants",
			nowFact:    1.,
			nbVoter:    3,
			expectNext: true,
		},
		{
			name:       "Expired",
			round:      1,
			nowFact:    1.,
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
			nbVoter:    3,
			expectNext: true,
		},
		{
			name:         "Last round time",
			round:        1,
			threshold:    1,
			nowFact:      1.,
			deadlineFact: 1.5,
			nbVoter:      1,
			expectNext:   false,
		},
		{
			name:         "Last round vote",
			round:        1,
			threshold:    1,
			nowFact:      0.75,
			deadlineFact: 1.25,
			nbVoter:      3,
			expectNext:   false,
		},
		{
			name:         "Missing rounds time",
			round:        1,
			minNbRounds:  3,
			threshold:    1,
			nowFact:      1.,
			deadlineFact: 1.5,
			nbVoter:      1,
			expectNext:   true,
		},
		{
			name:         "Missing rounds vote",
			round:        1,
			minNbRounds:  3,
			threshold:    1,
			nowFact:      0.75,
			deadlineFact: 1.25,
			nbVoter:      3,
			expectNext:   true,
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
			if err == nil && tt.minNbRounds > 2 {
				_, err = db.DB.Exec(qSetMin, tt.minNbRounds, pollId)
			}
			if err == nil && tt.nowFact > 0 {
				_, err = db.DB.Exec(qSetNow, tt.nowFact, pollId)
			}
			if err == nil && tt.deadlineFact != 0 {
				_, err = db.DB.Exec(qSetDeadline, tt.deadlineFact, pollId)
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

			mustt(t, run(pollId))

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

func TestNextRound_fullCheck(t *testing.T) {
	metaTestNextRound(t, func(pollId uint32) error {
		return newNextRound().fullCheck()
	})
}

func TestNextRound_checkOne(t *testing.T) {
	metaTestNextRound(t, func(pollId uint32) error {
		return newNextRound().checkOne(pollId)
	})
}
