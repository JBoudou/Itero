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

package events

import (
	"sync"
	"testing"
)

type muxTestReceiver struct {
	id  int
	ch  chan<- [2]int
	t   *testing.T
	grp *sync.WaitGroup
}

func (self *muxTestReceiver) Receive(evt Event) {
	i, ok := evt.(int)
	if !ok {
		self.t.Logf("Unknown event %v.", evt)
		return
	}
	self.ch <- [2]int{self.id, i}
}

func (self *muxTestReceiver) Close() {
	self.grp.Done()
}

type muxTestChecker struct {
	missing  map[[2]int]bool
	unwanted [][2]int
	grp *sync.WaitGroup
}

func (self *muxTestChecker) Listen(ch <-chan [2]int) {
	for result := range ch {
		if _, found := self.missing[result]; found {
			delete(self.missing, result)
		} else {
			self.unwanted = append(self.unwanted, result)
		}
	}
	self.grp.Done()
}

func TestMux(t *testing.T) {
	const chanSize = 128

	tests := []struct {
		name      string
		receivers []int
		events    []int
		expect    [][2]int
	}{
		{
			name:   "No receiver",
			events: []int{27},
		},
		{
			name:      "One receiver",
			receivers: []int{1},
			events:    []int{2, 3},
			expect:    [][2]int{{1, 2}, {1, 3}},
		},
		{
			name:      "Two receiver",
			receivers: []int{1, 2},
			events:    []int{2, 3},
			expect:    [][2]int{{1, 2}, {1, 3}, {2, 2}, {2, 3}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var grp sync.WaitGroup
			grp.Add(len(tt.receivers))

			checker := &muxTestChecker{missing: make(map[[2]int]bool, len(tt.expect)), grp: &grp}
			for _, result := range tt.expect {
				checker.missing[result] = true
			}
			resultCh := make(chan [2]int, 0)
			go checker.Listen(resultCh)

			manager := NewAsyncManager(chanSize)
			for _, id := range tt.receivers {
				manager.AddReceiver(&muxTestReceiver{id: id, ch: resultCh, t: t, grp: &grp})
			}

			for _, event := range tt.events {
				manager.Send(event)
			}
			manager.Close()
			grp.Wait()
			grp.Add(1)
			close(resultCh)
			grp.Wait()

			for result := range checker.missing {
				t.Errorf("Missing %v.", result)
			}
			for result := range checker.unwanted {
				t.Errorf("Unexpected %v.", result)
			}
		})
	}
}
