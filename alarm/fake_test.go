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

const otherRoutineWait = 2 * time.Millisecond

func nonBlockingEventReceive(ch <-chan Event) *Event {
	select {
	case evt := <-ch:
		return &evt
	default:
		return nil
	}
}

func isIntWithValue(a interface{}, b int) bool {
	asInt, ok := a.(int)
	return ok && (asInt == b)
}

func TestFakeAlarm(t *testing.T) {
	alarm, ctrl := NewFakeAlarm()
	defer ctrl.Close()

	t.Run("Tick sends event", func(t *testing.T) {
		alarm.Send <- Event{Time: time.Date(2000, time.January, 1, 1, 1, 1, 0, time.Local)}
		time.Sleep(otherRoutineWait)
		if got := nonBlockingEventReceive(alarm.Receive); got != nil {
			t.Errorf("Expect no event to be received. Got %v.", *got)
		}

		ctrl.Tick()
		time.Sleep(otherRoutineWait)
		if got := nonBlockingEventReceive(alarm.Receive); got == nil {
			t.Errorf("Expect event to be received. Got nil.")
		}
	})

	t.Run("Event are ordered", func(t *testing.T) {
		alarm.Send <- Event{
			Time: time.Date(2002, time.January, 1, 1, 1, 1, 0, time.Local),
			Data: int(1),
		}
		alarm.Send <- Event{
			Time: time.Date(2000, time.January, 1, 1, 1, 1, 0, time.Local),
			Data: int(0),
		}
		time.Sleep(otherRoutineWait)

		ctrl.Tick()
		time.Sleep(otherRoutineWait)
		if got := nonBlockingEventReceive(alarm.Receive); got != nil && !isIntWithValue(got.Data, 0) {
			t.Errorf("Wrong event received. Expect 0. Got %v.", got)
		}

		ctrl.Tick()
		time.Sleep(otherRoutineWait)
		if got := nonBlockingEventReceive(alarm.Receive); got != nil && !isIntWithValue(got.Data, 1) {
			t.Errorf("Wrong event received. Expect 1. Got %v.", got)
		}
	})

	t.Run("Remaining is set correctly", func(t *testing.T) {
		alarm.Send <- Event{ Time: time.Date(2000, time.January, 1, 1, 1, 1, 0, time.Local) }
		alarm.Send <- Event{ Time: time.Date(2002, time.January, 1, 1, 1, 1, 0, time.Local) }
		time.Sleep(otherRoutineWait)

		ctrl.Tick()
		time.Sleep(otherRoutineWait)
		if got := nonBlockingEventReceive(alarm.Receive); got != nil && got.Remaining != 1 {
			t.Errorf("Wrong Remaining. Expect 1. Got %v.", got)
		}

		ctrl.Tick()
		time.Sleep(otherRoutineWait)
		if got := nonBlockingEventReceive(alarm.Receive); got != nil && got.Remaining != 0 {
			t.Errorf("Wrong Remaining. Expect 0. Got %v.", got)
		}
	})
}
