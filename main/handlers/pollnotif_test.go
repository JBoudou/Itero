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

package handlers

import (
	"encoding/json"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/JBoudou/Itero/main/services"
	"github.com/JBoudou/Itero/mid/db"
	dbt "github.com/JBoudou/Itero/mid/db/dbtest"
	"github.com/JBoudou/Itero/mid/salted"
	"github.com/JBoudou/Itero/mid/server"
	srvt "github.com/JBoudou/Itero/mid/server/servertest"
	"github.com/JBoudou/Itero/pkg/events"
	"github.com/JBoudou/Itero/pkg/ioc"
)

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

	dbEnv      dbt.Env
	evtManager events.Manager
	admnId     uint32
	partId     uint32
	pollId     uint32
}

func (self *pollNotifHandlerTest) GetName() string {
	return self.name
}

func (self *pollNotifHandlerTest) Prepare(t *testing.T) *ioc.Locator {
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

	ret := ioc.Root.Sub()
	// TODO: replace the two following lines with the commented line below.
	ret.Set(func() events.Manager { return events.NewAsyncManager(events.DefaultManagerChannelSize) })
	ret.Get(&self.evtManager)
	//ret.Refresh(&self.evtManager)
	ret.Set(func(evtManager events.Manager) (services.PollNotifChannel, error) {
		return services.RunPollNotif(time.Second, evtManager)
	})

	return ret
}

func (self *pollNotifHandlerTest) GetRequest(t *testing.T) *srvt.Request {
	// Send the events now that the handler is created
	for _, fct := range self.events {
		mustt(t, self.evtManager.Send(fct(self.pollId)))
		time.Sleep(time.Millisecond)
	}

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

	segment, err := salted.Segment{Id: self.pollId, Salt: 42}.Encode()
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
	if self.evtManager != nil {
		self.evtManager.Close()
	}
}

func TestPollNotifHandler(t *testing.T) {
	create := func(id uint32) events.Event { return services.StartPollEvent{Poll: id} }
	next := func(id uint32) events.Event { return services.NextRoundEvent{Poll: id} }
	term := func(id uint32) events.Event { return services.ClosePollEvent{Poll: id} }

	tests := []srvt.Test{
		&pollNotifHandlerTest{
			name:     "Admin Create",
			events:   []func(uint32) events.Event{create},
			userKind: pollNotifHandlerTestUserAdmin,
			expect:   []PollNotifAnswerEntry{{Action: services.PollNotifStart}},
		},
		&pollNotifHandlerTest{
			name:     "Admin Next",
			events:   []func(uint32) events.Event{next},
			userKind: pollNotifHandlerTestUserAdmin,
			expect:   []PollNotifAnswerEntry{{Action: services.PollNotifNext}},
		},
		&pollNotifHandlerTest{
			name:     "Admin Term",
			events:   []func(uint32) events.Event{term},
			userKind: pollNotifHandlerTestUserAdmin,
			expect:   []PollNotifAnswerEntry{{Action: services.PollNotifTerm}},
		},
		&pollNotifHandlerTest{
			name:     "Admin All",
			events:   []func(uint32) events.Event{create, next, term},
			userKind: pollNotifHandlerTestUserAdmin,
			expect: []PollNotifAnswerEntry{
				{Action: services.PollNotifStart},
				{Action: services.PollNotifNext},
				{Action: services.PollNotifTerm},
			},
		},
		&pollNotifHandlerTest{
			name:     "Part Next",
			events:   []func(uint32) events.Event{next},
			userKind: pollNotifHandlerTestUserPart,
			expect:   []PollNotifAnswerEntry{{Action: services.PollNotifNext}},
		},
		&pollNotifHandlerTest{
			name:     "Part Term",
			events:   []func(uint32) events.Event{term},
			userKind: pollNotifHandlerTestUserPart,
			expect:   []PollNotifAnswerEntry{{Action: services.PollNotifTerm}},
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

	srvt.Run(t, tests, PollNotifHandler)
}
