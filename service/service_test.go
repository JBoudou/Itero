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

package service

import (
	"testing"
	"time"

	"github.com/JBoudou/Itero/alarm"
	"github.com/JBoudou/Itero/events"
)

type serviceToTest interface {
	Service
	closeChannel() <-chan struct{}
}

func testRunService(t *testing.T, service serviceToTest, idle func()) {
	var alarmCtrl alarm.FakeAlarmController
	oldAlarmInjector := func(chanSize int, opts ...alarm.Option) (ret alarm.Alarm) {
		ret, alarmCtrl = alarm.NewFakeAlarm(chanSize, opts...)
		return
	}
	oldAlarmInjector, InjectAlarmInService = InjectAlarmInService, oldAlarmInjector
	ticker := time.NewTicker(time.Second / 5)

	defer func() {
		ticker.Stop()
		InjectAlarmInService = oldAlarmInjector
		alarmCtrl.Close()
	}()

	stopFunc := Run(service)
	defer stopFunc()

mainLoop:
	for true {
		select {
		case <-service.closeChannel():
			time.Sleep(2 * time.Millisecond)
			break mainLoop

		case <-ticker.C:
			idle()
			alarmCtrl.Tick()
		}
	}

	if !t.Failed() && alarmCtrl.QueueLength() == 0 {
		t.Errorf("Alarm queue is empty.")
	}
}

// Base test without any event //

const (
	testRunServiceNbTasks   = 4
	testRunServiceTaskDelay = time.Minute
)

// testRunServiceTaskDelay is a testing service.
// It has testRunServiceNbTasks tasks:
//  0 -> has to be done right now
//  1 -> has to be done after
//  2 -> must be delayed twice before been done
//  3 -> does not have to be done once the previous ones have been done
type testRunServiceService struct {
	t        *testing.T
	start time.Time
	state    uint32
	checkCnt int
	closeCh  chan struct{}
}

func newTestRunServiceService(t *testing.T) *testRunServiceService {
	return &testRunServiceService{
		t: t,
		start: time.Now(),
		closeCh: make(chan struct{}),
	}
}

func (self *testRunServiceService) nextState() {
	self.state += 1
	if self.state >= testRunServiceNbTasks {
		close(self.closeCh)
	}
}

func (self *testRunServiceService) ProcessOne(id uint32) error {
	if self.state != id {
		self.t.Errorf("Proceeding %d instead of %d", id, self.state)
		if id > self.state {
			self.state = id
		}
	}

	if id == 2 && self.checkCnt < 2 {
		self.checkCnt += 1
		return NothingToDoYet
	}
	if id == 3 {
		return NothingToDoYet
	}

	self.nextState()
	return nil
}

func (self *testRunServiceService) CheckAll() IdAndDateIterator {
	return &testRunServiceIterator{
		service: self,
		pos:     self.state,
		end:     testRunServiceNbTasks,
		start:   self.start,
	}
}

func (self *testRunServiceService) CheckOne(id uint32) time.Time {
	if id < self.state {
		return time.Time{}
	}
	ret := self.start.Add(time.Duration(id) * testRunServiceTaskDelay)
	if id == 3 && self.state == 3 {
		self.nextState()
		return time.Time{}
	}
	if id == 2 && self.state == 2 && self.checkCnt < 2 {
		self.checkCnt += 1
		return ret.Add(time.Duration(time.Duration(self.checkCnt) * time.Second))
	}
	return ret
}

func (self *testRunServiceService) CheckInterval() time.Duration {
	return time.Duration(testRunServiceNbTasks+2) * testRunServiceTaskDelay
}

func (self *testRunServiceService) Logger() LevelLogger {
	return EasyLogger{}
}

func (self *testRunServiceService) closeChannel() <-chan struct{} {
	return self.closeCh
}

type testRunServiceIterator struct {
	service *testRunServiceService
	pos     uint32
	end     uint32
	err     error
	start   time.Time
}

func (self *testRunServiceIterator) Next() bool {
	if self.err != nil {
		return false
	}
	self.pos += 1
	return self.pos <= self.end
}

func (self *testRunServiceIterator) IdAndDate() (uint32, time.Time) {
	return self.pos - 1, self.start.Add(time.Duration(self.pos - 1) * testRunServiceTaskDelay)
}

func (self *testRunServiceIterator) Err() error {
	return self.err
}

func (self *testRunServiceIterator) Close() error {
	return self.err
}

func TestRunService_noEvents(t *testing.T) {
	testRunService(t, newTestRunServiceService(t), func(){})
}

// Test with events, but no event received //

type testRunServiceServiceDumb struct {
	*testRunServiceService
}

func (self *testRunServiceServiceDumb) FilterEvent(events.Event) bool {
	return false
}

func (self *testRunServiceServiceDumb) ReceiveEvent(events.Event, ServiceRunnerControl) {
}

func TestRunService_dumbEvents(t *testing.T) {
	testRunService(t, &testRunServiceServiceDumb{
		testRunServiceService: newTestRunServiceService(t),
	}, func() {})
}

// Test with real events //

type testRunServiceEvent uint32

type testRunServiceServiceEvent struct {
	*testRunServiceService
}

func (self *testRunServiceServiceEvent) FilterEvent(evt events.Event) bool {
	_, ok := evt.(testRunServiceEvent)
	return ok
}

func (self *testRunServiceServiceEvent) ReceiveEvent(evt events.Event, ctrl ServiceRunnerControl) {
	evtId, ok := evt.(testRunServiceEvent)
	if !ok {
		return
	}
	ctrl.Schedule(uint32(evtId))
}

func (self *testRunServiceServiceEvent) CheckAll() IdAndDateIterator {
	return &testRunServiceIterator{
		service: self.testRunServiceService,
		pos:     self.state,
		end:     2,
		start:   time.Now().Add(-1 * time.Second),
	}
}

type testRunServiceEventSender struct {
	wait uint32
	pos uint32
	end uint32
}

func (self *testRunServiceEventSender) send() {
	if self.wait > 0 {
		self.wait -= 1
		return
	}
	if self.pos < self.end {
		events.Send(testRunServiceEvent(self.pos))
		self.pos += 1
		return
	}
}

func TestRunService_events(t *testing.T) {
	testRunService(t, &testRunServiceServiceEvent{
		testRunServiceService: newTestRunServiceService(t),
	},
	(&testRunServiceEventSender{wait: 1, pos: 2, end: testRunServiceNbTasks}).send)
}
