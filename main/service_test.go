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
