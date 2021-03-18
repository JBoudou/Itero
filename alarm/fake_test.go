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

package alarm

import (
	"testing"
	"time"
)

func nonBlockingEventReceive(ch <-chan Event) *Event {
	select {
	case evt := <-ch:
		return &evt
	default:
		return nil
	}
}

func TestFakeAlarm(t *testing.T) {
	alarm, ctrl := NewFakeAlarm()

	alarm.Send <- Event{Time: time.Date(2000, time.January, 1, 1, 1, 1, 0, time.Local)}
	time.Sleep(10*time.Millisecond)
	if got := nonBlockingEventReceive(alarm.Receive); got != nil {
		t.Errorf("Expect no event to be received. Got %v.", *got)
	}

	ctrl.Tick()
	time.Sleep(10*time.Millisecond)
	if got := nonBlockingEventReceive(alarm.Receive); got != nil {
		t.Errorf("Expect event to be received. Got nil.")
	}
}
