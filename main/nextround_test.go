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
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/JBoudou/Itero/db"
	dbt "github.com/JBoudou/Itero/db/dbtest"
	"github.com/JBoudou/Itero/events"
	"github.com/JBoudou/Itero/events/eventstest"
)

type nextRoundTestInstance struct {
	name         string
	round        uint8   // CurrentRound (MaxNbRounds = 3)
	minNbRounds  uint8   // applied only if >2
	nowFact      float32 // if  >0 set Now      = CurrentRoundStart + nowFact      * MaxRoundDuration
	deadlineFact float32 // if !=0 set Deadline = CurrentRoundStart + deadlineFact * MaxRoundDuration
	pubInvited   bool    // whether Publicity is Invited
	threshold    float64 // RoundThreshold
	nbVoter      int     // number of Participant with LastRound = Poll.CurrentRound
	expectNext   bool
	expectList   bool // whether it must be listed by CheckAll
	expectCheck  int  // kind of response from CheckOne (see testCheckOneResult*)
}

const (
	testCheckOneResultPast = iota
	testCheckOneResultFuture
	testCheckOneResultNever
)

func metaTestNextRound(t *testing.T, checker func(*testing.T, *nextRoundTestInstance, uint32)) {

	const (
		nbParticipants = 3

		qParticipate = `INSERT INTO Participants(Poll, User, Round) VALUE (?,?,?)`
		qUpdatePoll  = `UPDATE Polls SET CurrentRound = ?, RoundThreshold = ? WHERE Id = ?`
		qSetMin      = `UPDATE Polls SET MinNbRounds = ? WHERE Id = ?`
		qSetNow      = `
		  UPDATE Polls
		     SET CurrentRoundStart = SUBTIME(CURRENT_TIMESTAMP(), ? * MaxRoundDuration)
		   WHERE Id = ?`
		qSetDeadline = `
		  UPDATE Polls
			   SET Deadline = ADDTIME(CurrentRoundStart, ? * MaxRoundDuration)
			 WHERE Id = ?`
		qSetInvited = `UPDATE Polls SET Publicity = ? WHERE Id = ?`
	)

	// Tests are independent.
	tests := []nextRoundTestInstance{
		{
			name:       "Default",
			expectList: true,
			expectCheck: testCheckOneResultFuture,
		},
		{
			name:       "Round zero - 2 participants",
			nowFact:    1.,
			nbVoter:    2,
			expectNext: false,
			expectList: true,
			expectCheck: testCheckOneResultNever,
		},
		{
			name:       "Round zero - 3 participants",
			nowFact:    1.,
			nbVoter:    3,
			expectNext: true,
			expectList: true,
			expectCheck: testCheckOneResultPast,
		},
		{
			name:       "Expired",
			round:      1,
			nowFact:    1.,
			expectNext: true,
			expectList: true,
			expectCheck: testCheckOneResultPast,
		},
		{
			name:       "Threshold zero",
			round:      1,
			threshold:  0,
			nbVoter:    1,
			expectNext: true,
			expectList: true,
			expectCheck: testCheckOneResultPast,
		},
		{
			name:       "Threshold one",
			round:      1,
			threshold:  1,
			nbVoter:    3,
			expectNext: true,
			expectList: true,
			expectCheck: testCheckOneResultPast,
		},
		{
			name:         "Last round time",
			round:        1,
			threshold:    1,
			nowFact:      1.,
			deadlineFact: 1.5,
			nbVoter:      1,
			expectNext:   false,
			expectList:   true,
			expectCheck: testCheckOneResultFuture,
		},
		{
			name:         "Last round vote",
			round:        1,
			threshold:    1,
			nowFact:      0.75,
			deadlineFact: 1.25,
			nbVoter:      3,
			expectNext:   false,
			expectList:   true,
			expectCheck: testCheckOneResultFuture,
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
			expectList:   true,
			expectCheck: testCheckOneResultPast,
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
			expectList:   true,
			expectCheck: testCheckOneResultPast,
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

			if tt.round > 0 {
				stmt, err = db.DB.Prepare(qParticipate)
				mustt(t, err)
				for _, id := range user {
					_, err = stmt.Exec(pollId, id, 0)
					mustt(t, err)
				}
				mustt(t, stmt.Close())
			}

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
				stmt, err = db.DB.Prepare(qParticipate)
				mustt(t, err)
				for i := 0; i < tt.nbVoter; i++ {
					_, err = stmt.Exec(pollId, user[i], tt.round)
					mustt(t, err)
				}
				err = stmt.Close()
			}
			mustt(t, err)

			checker(t, &tt, pollId)
		})
	}
}

// ProcessOne //

func nextRound_processOne_checker(t *testing.T, tt *nextRoundTestInstance, pollId uint32) {
	const qGetRound = `SELECT CurrentRound FROM Polls WHERE Id = ?`

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

	err := NextRoundService.ProcessOne(pollId)

	events.DefaultManager = originalManager

	nothingToDoYet := false
	if errors.Is(err, NothingToDoYet) {
		nothingToDoYet = true
		err = nil
	}
	mustt(t, err)

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
	if incremented && nothingToDoYet {
		t.Errorf("Event received but NothingToDoYet returned")
	}
	if tt.expectNext && nothingToDoYet {
		t.Errorf("NothingToDoYet returned but expected to change round.")
	}
	if gotRound != expectRound {
		t.Errorf("Wrong round. Got %d. Expect %d.", gotRound, expectRound)
	}
}

func TestNextRoundService_ProcessOne(t *testing.T) {
	metaTestNextRound(t, nextRound_processOne_checker)
}

// CheckAll //

func idDateIteratorHasId(t *testing.T, iterator IdAndDateIterator, id uint32) bool {
	for iterator.Next() {
		got, _ := iterator.IdAndDate()
		if got == id {
			return true
		}
	}
	if err := iterator.Err(); err != nil {
		t.Errorf("Iterator error: %v.", err)
	}
	return false
}

func nextRound_checkAll_checker(t *testing.T, tt *nextRoundTestInstance, pollId uint32) {
	iterator := NextRoundService.CheckAll()

	listed := idDateIteratorHasId(t, iterator, pollId)
	if listed != tt.expectList {
		if tt.expectList {
			t.Errorf("Poll not listed when it should.")
		} else {
			t.Errorf("Poll listed when it shouldn't.")
		}
	}
}

func TestNextRoundService_CheckAll(t *testing.T) {
	metaTestNextRound(t, nextRound_checkAll_checker)
}

// CheckOne //

func nextRound_checkOne_checker(t *testing.T, tt *nextRoundTestInstance, pollId uint32) {
	got := NextRoundService.CheckOne(pollId)
	diff := time.Until(got)
	isZero := got.IsZero()

	switch tt.expectCheck {

	case testCheckOneResultPast:
		if isZero {
			t.Errorf("Got zero time. Expect time in the past.")
			break
		}
		if diff > 2 * time.Millisecond {
			t.Errorf("Got time in the future. Expect time in the past.")
		}

	case testCheckOneResultFuture:
		if isZero {
			t.Errorf("Got zero time. Expect time in the future.")
			break
		}
		if diff < 2 * time.Millisecond {
			t.Errorf("Got time in the past. Expect time in the future.")
		}

	case testCheckOneResultNever:
		if !isZero{
			t.Errorf("Got %v. Expect zero time.", got)
		}

	}
}

func TestNextRoundService_CheckOne(t *testing.T) {
	metaTestNextRound(t, nextRound_checkOne_checker)
}
