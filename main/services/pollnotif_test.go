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

package services

import (
	"testing"
	"time"

	"github.com/JBoudou/Itero/pkg/events"
)

// TestPollNotifList_Global tests all PollNotifList's methods but in only one scenario.
func TestPollNotifList_Global(t *testing.T) {
	const delay = 210 * time.Millisecond

	tests := []struct {
		read  [2]uint32 // {min, max+1}
		write [2]uint32 // {min, max+1}
		sleep int       // milliseconds
	}{
		{
			read:  [2]uint32{0, 0},
			write: [2]uint32{1, 4},
			sleep: 180,
		},
		{
			read:  [2]uint32{1, 4},
			write: [2]uint32{4, 6},
			sleep: 60,
		},
		{
			read:  [2]uint32{4, 6},
			write: [2]uint32{6, 20},
			sleep: 30,
		},
		{
			read:  [2]uint32{4, 20},
			write: [2]uint32{0, 0},
			sleep: 210,
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
	t.Parallel()

	evtManager := events.NewAsyncManager(0)
	defer evtManager.Close()
	notifChannel, err := RunPollNotif(10*time.Millisecond, evtManager)
	mustt(t, err)

	elements := []struct {
		event  events.Event
		id     uint32
		round  uint8
		action PollNotifAction
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
		evtManager.Send(elt.event)
	}
	time.Sleep(5 * time.Millisecond)

	sl := <-notifChannel
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
