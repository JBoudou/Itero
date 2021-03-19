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
	"reflect"
	"testing"
	"time"

	"github.com/JBoudou/Itero/alarm"
	"github.com/JBoudou/Itero/events"
)

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
	state    uint32
	checkCnt int
	t        *testing.T
	closeCh  chan struct{}
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
	return &testRunServiceIterator{service: self, pos: self.state, start: time.Now()}
}

func (self *testRunServiceService) CheckOne(id uint32) time.Time {
	ret := time.Now()
	if id == 3 && self.state == 3 {
		self.nextState()
		return time.Time{}
	}
	if id == 2 && self.state == 2 && self.checkCnt < 2 {
		self.checkCnt += 1
		return ret.Add(testRunServiceTaskDelay)
	}
	if id > self.state {
		return ret.Add(time.Duration(id-self.state) * testRunServiceTaskDelay)
	}
	return ret
}

func (self *testRunServiceService) CheckInterval() time.Duration {
	return time.Duration(testRunServiceNbTasks+2) * testRunServiceTaskDelay
}

func (self *testRunServiceService) Logger() LevelLogger {
	return EasyLogger{}
}

type testRunServiceIterator struct {
	service *testRunServiceService
	pos     uint32
	err     error
	start   time.Time
}

func (self *testRunServiceIterator) Next() bool {
	if self.err != nil {
		return false
	}
	self.pos += 1
	return self.pos <= testRunServiceNbTasks
}

func (self *testRunServiceIterator) IdAndDate() (uint32, time.Time) {
	return self.pos - 1, self.start.Add(time.Duration(self.pos) * testRunServiceTaskDelay)
}

func (self *testRunServiceIterator) Err() error {
	return self.err
}

func (self *testRunServiceIterator) Close() error {
	return self.err
}

func TestRunService_noEvents(t *testing.T) {
	fakeAlarm, alarmCtrl := alarm.NewFakeAlarm()
	oldAlarmInjector := func() alarm.Alarm { return fakeAlarm }
	oldAlarmInjector, InjectAlarmInService = InjectAlarmInService, oldAlarmInjector

	service := &testRunServiceService{t: t, closeCh: make(chan struct{})}

	defer func() {
		InjectAlarmInService = oldAlarmInjector
		alarmCtrl.Close()
	}()

	stopFunc := RunService(service)
	defer stopFunc()

mainLoop:
	for true {
		select {
		case <-service.closeCh:
			time.Sleep(2 * time.Millisecond)
			break mainLoop

		default:
			alarmCtrl.Tick()
		}
	}

	if !t.Failed() && alarmCtrl.QueueLength() == 0 {
		t.Errorf("Alarm queue is empty.")
	}
}

// OLD CODE //

func TestPollService_updateLastCheck(t *testing.T) {
	pollService := &pollService{}
	pollService.updateLastCheck()
	now := time.Now()
	if pollService.adjust < 0 {
		pollService.adjust = -pollService.adjust
	}
	if pollService.lastCheck.Before(now.Add(-2 * (pollService.adjust + time.Millisecond))) {
		t.Errorf("lastCheck too early")
	}
	if pollService.lastCheck.After(now.Add(2 * (pollService.adjust + time.Millisecond))) {
		t.Errorf("lastCheck too late")
	}
}

func TestPollService_nextAlarm_helper(t *testing.T) {
	lastCheck := time.Now()
	const defaultWait = time.Hour

	tests := []struct {
		name   string
		query  string
		expect alarm.Event
	}{
		{
			name:   "Normal",
			query:  `SELECT 27, CAST(? AS DATETIME), CURRENT_TIMESTAMP()`,
			expect: alarm.Event{Time: lastCheck, Data: uint32(27)},
		},
		{
			name:   "Default",
			query:  `SELECT Id, ?, CURRENT_TIMESTAMP() FROM Polls WHERE Id = 0 AND Id != 0`,
			expect: alarm.Event{Time: time.Now().Add(defaultWait)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := newPollService("Test", func(uint32) events.Event { return nil })
			service.lastCheck = lastCheck
			service.adjust = 0

			got := service.nextAlarm_helper(tt.query, defaultWait)

			diff := tt.expect.Time.Sub(got.Time)
			if diff < 0 {
				diff = -diff
			}

			if diff > time.Minute {
				t.Errorf("Wrong time. Got %v. Expect %v. Diff %v.", got.Time, tt.expect.Time, diff)
			}
			if !reflect.DeepEqual(got.Data, tt.expect.Data) {
				t.Errorf("Wrong data. Got %v. Expect %v.", got.Data, tt.expect.Data)
			}
		})
	}
}
