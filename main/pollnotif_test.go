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
	"encoding/json"
	"net/http"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/JBoudou/Itero/db"
	dbt "github.com/JBoudou/Itero/db/dbtest"
	"github.com/JBoudou/Itero/events"
	"github.com/JBoudou/Itero/server"
	srvt "github.com/JBoudou/Itero/server/servertest"
)

//
// Implementation details tests
//

// TestPollNotifList_Global tests all PollNotifList's methods but in only one scenario.
func TestPollNotifList_Global(t *testing.T) {
	const delay = 70 * time.Millisecond

	tests := []struct {
		read  [2]uint32 // {min, max+1}
		write [2]uint32 // {min, max+1}
		sleep int       // milliseconds
	}{
		{
			read:  [2]uint32{0, 0},
			write: [2]uint32{1, 4},
			sleep: 60,
		},
		{
			read:  [2]uint32{1, 4},
			write: [2]uint32{4, 6},
			sleep: 20,
		},
		{
			read:  [2]uint32{4, 6},
			write: [2]uint32{6, 20},
			sleep: 10,
		},
		{
			read:  [2]uint32{4, 20},
			write: [2]uint32{0, 0},
			sleep: 70,
		},
		{
			read:  [2]uint32{0, 0},
			write: [2]uint32{20, 21},
			sleep: 1,
		},
		{
			read:  [2]uint32{20, 21},
			write: [2]uint32{21, 22},
			sleep: 1,
		},
	}

	list := NewPollNotifList(delay)
	for step, tt := range tests {
		ts := time.Now()

		// Read
		sl := list.Slice()
		sllen := len(sl)
		if sllen != int(tt.read[1]-tt.read[0]) {
			t.Errorf("Step %d. Read. Got len %d. Expect %d.", step, sllen, tt.read[1]-tt.read[0])
		}
		for i, got := range sl {
			if expect := tt.read[0] + uint32(i); got.Id != expect {
				t.Errorf("Step %d. Read. Got slice[%d] = %d. Expect %d.", step, i, got.Id, expect)
			}
		}

		// Copy
		slcopy := list.Copy().Slice()
		cplen := len(slcopy)
		if sllen != cplen {
			t.Errorf("Step %d. Copy. Got len %d. Expect %d.", step, cplen, sllen)
			if cplen > sllen {
				cplen = sllen
			}
		}
		for i := 0; i < cplen; i++ {
			if slcopy[i].Id != sl[i].Id {
				t.Errorf("Step %d. Copy. Got copy[%d] = %d. Expect %d.", step, i, slcopy[i].Id, sl[i].Id)
			}
		}

		// Write
		for id := tt.write[0]; id < tt.write[1]; id++ {
			list.Add(&PollNotification{Timestamp: ts, Id: id})
		}

		// Sleep
		time.Sleep(time.Duration(tt.sleep) * time.Millisecond)
	}
}

func TestPollNotif(t *testing.T) {
	mustt(t, RunPollNotif(10*time.Millisecond))

	elements := []struct {
		event  events.Event
		id     uint32
		round  uint8
		action uint8
	}{
		{
			event:  StartPollEvent{Poll: 1},
			id:     1,
			action: PollNotifStart,
		},
		{
			event:  NextRoundEvent{Poll: 2, Round: 4},
			id:     2,
			round:  4,
			action: PollNotifNext,
		},
		{
			event:  ClosePollEvent{Poll: 3},
			id:     3,
			action: PollNotifTerm,
		},
	}
	for _, elt := range elements {
		events.Send(elt.event)
	}
	time.Sleep(5 * time.Millisecond)

	sl := <-PollNotifChannel
	sllen := len(sl)
	if sllen != len(elements) {
		t.Errorf("Wrong slice len. Got %d. Expect %d.", sllen, len(elements))
		if sllen > len(elements) {
			sllen = len(elements)
		}
	}
	for i := 0; i < sllen; i++ {
		if sl[i].Id != elements[i].id {
			t.Errorf("Wrong notif id at index %d. Got %d. Expect %d.", i, sl[i].Id, elements[i].id)
		}
		if sl[i].Round != elements[i].round {
			t.Errorf("Wrong notif round at index %d. Got %d. Expect %d.", i, sl[i].Round, elements[i].round)
		}
		if sl[i].Action != elements[i].action {
			t.Errorf("Wrong notif action at index %d. Got %d. Expect %d.", i, sl[i].Action, elements[i].action)
		}
	}
}

//
// Handler test
//

const (
	pollNotifHandlerTestUserAdmin = iota
	pollNotifHandlerTestUserPart
	pollNotifHandlerTestUserAlien
)

type pollNotifHandlerTest struct {
	name      string
	events    []func(uint32) events.Event
	userKind  uint8
	lastDelay time.Duration // Compute LastUpdate in query as Now - lastDelay. Default to one second.
	expect    []PollNotifAnswerEntry

	dbEnv  dbt.Env
	admnId uint32
	partId uint32
	pollId uint32
}

func (self *pollNotifHandlerTest) GetName() string {
	return self.name
}

func (self *pollNotifHandlerTest) Prepare(t *testing.T) {
	t.Parallel()

	self.admnId = self.dbEnv.CreateUserWith("PollNotifHandler" + self.name + "Admin")
	if self.userKind != pollNotifHandlerTestUserAdmin {
		self.partId = self.dbEnv.CreateUserWith("PollNotifHandler" + self.name + "Part")
	}
	self.pollId = self.dbEnv.CreatePoll("Title", self.admnId, db.PollPublicityPublicRegistered)

	const (
		qParticipate = `INSERT INTO Participants(User,Poll,Round) VALUE (?,?,0)`
	)

	if self.userKind == pollNotifHandlerTestUserPart {
		self.dbEnv.QuietExec(qParticipate, self.partId, self.pollId)
	}

	self.dbEnv.Must(t)

	for _, fct := range self.events {
		mustt(t, events.Send(fct(self.pollId)))
		time.Sleep(time.Millisecond)
	}
}

func (self *pollNotifHandlerTest) GetRequest(t *testing.T) *srvt.Request {
	userId := &self.partId
	if self.userKind == pollNotifHandlerTestUserAdmin {
		userId = &self.admnId
	}

	// Body
	if self.lastDelay == 0 {
		self.lastDelay = time.Second
	}
	last := time.Now().Add(-1 * self.lastDelay)
	body, err := json.Marshal(PollNotifQuery{LastUpdate: last})
	mustt(t, err)

	return &srvt.Request{Method: "POST", UserId: userId, Body: string(body)}
}

func (self *pollNotifHandlerTest) Check(t *testing.T, response *http.Response, request *server.Request) {
	if response.StatusCode != http.StatusOK {
		t.Errorf("Unexpected HTTP error %d.", response.StatusCode)
	}

	decoder := json.NewDecoder(response.Body)
	var got []PollNotifAnswerEntry
	mustt(t, decoder.Decode(&got))

	glen := len(got)
	if glen != len(self.expect) {
		t.Errorf("Wrong length. Got %d. Expect %d.", glen, len(self.expect))
		if glen > len(self.expect) {
			glen = len(self.expect)
		}
	}

	segment, err := PollSegment{Id: self.pollId, Salt: 42}.Encode()
	mustt(t, err)
	for i := 0; i < glen; i++ {

		// Fix things
		got[i].Timestamp = time.Time{}
		self.expect[i].Segment = segment
		if self.expect[i].Title == "" {
			self.expect[i].Title = "Title"
		}

		// Check
		if !reflect.DeepEqual(got[i], self.expect[i]) {
			t.Errorf("Wrong value index %d. Got %v. Expect %v.", i, got[i], self.expect[i])
		}
	}
}

func (self *pollNotifHandlerTest) Close() {
	self.dbEnv.Close()
}

var testPollNotifHandlerRunPollNotif sync.Once

func TestPollNotifHandler(t *testing.T) {
	testPollNotifHandlerRunPollNotif.Do(func() { RunPollNotif(time.Second) })

	create := func(id uint32) events.Event { return StartPollEvent{Poll: id} }
	next := func(id uint32) events.Event { return NextRoundEvent{Poll: id} }
	term := func(id uint32) events.Event { return ClosePollEvent{Poll: id} }

	tests := []srvt.Test{
		&pollNotifHandlerTest{
			name:     "Admin Create",
			events:   []func(uint32) events.Event{create},
			userKind: pollNotifHandlerTestUserAdmin,
			expect:   []PollNotifAnswerEntry{{Action: PollNotifStart}},
		},
		&pollNotifHandlerTest{
			name:     "Admin Next",
			events:   []func(uint32) events.Event{next},
			userKind: pollNotifHandlerTestUserAdmin,
			expect:   []PollNotifAnswerEntry{{Action: PollNotifNext}},
		},
		&pollNotifHandlerTest{
			name:     "Admin Term",
			events:   []func(uint32) events.Event{term},
			userKind: pollNotifHandlerTestUserAdmin,
			expect:   []PollNotifAnswerEntry{{Action: PollNotifTerm}},
		},
		&pollNotifHandlerTest{
			name:     "Admin All",
			events:   []func(uint32) events.Event{create, next, term},
			userKind: pollNotifHandlerTestUserAdmin,
			expect: []PollNotifAnswerEntry{
				{Action: PollNotifStart},
				{Action: PollNotifNext},
				{Action: PollNotifTerm},
			},
		},
		&pollNotifHandlerTest{
			name:     "Part Next",
			events:   []func(uint32) events.Event{next},
			userKind: pollNotifHandlerTestUserPart,
			expect:   []PollNotifAnswerEntry{{Action: PollNotifNext}},
		},
		&pollNotifHandlerTest{
			name:     "Part Term",
			events:   []func(uint32) events.Event{term},
			userKind: pollNotifHandlerTestUserPart,
			expect:   []PollNotifAnswerEntry{{Action: PollNotifTerm}},
		},
		&pollNotifHandlerTest{
			name:     "Alien Create",
			events:   []func(uint32) events.Event{create},
			userKind: pollNotifHandlerTestUserAlien,
			expect:   []PollNotifAnswerEntry{},
		},
		&pollNotifHandlerTest{
			name:     "Alien Next",
			events:   []func(uint32) events.Event{next},
			userKind: pollNotifHandlerTestUserAlien,
			expect:   []PollNotifAnswerEntry{},
		},
		&pollNotifHandlerTest{
			name:     "Alien Term",
			events:   []func(uint32) events.Event{term},
			userKind: pollNotifHandlerTestUserAlien,
			expect:   []PollNotifAnswerEntry{},
		},
	}

	srvt.RunFunc(t, tests, PollNotifHandler)
}
