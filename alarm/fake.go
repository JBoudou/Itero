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

import "container/heap"

type FakeAlarmController interface {
	// Tick asks the fake alarm to send its most recent event.
	Tick()
	
	Close()
}

type fakeAlarm struct {
	// true is tick, false is close
	control chan<- bool
}

func (self fakeAlarm) Tick() {
	self.control <- true
}

func (self fakeAlarm) Close() {
	self.control <- false
}

type evtHeap []Event

func (self evtHeap) Len() int {
	return len(self)
}

func (self evtHeap) Less(i, j int) bool {
	return self[i].Time.Before(self[j].Time)
}

func (self evtHeap) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}

func (self *evtHeap) Push(evt interface{}) {
	*self = append(*self, evt.(Event))
}

func (self *evtHeap) Pop() (ret interface{}) {
	old := *self
	n := len(old) - 1
	ret = (*self)[n]
	*self = old[0:n]
	return
}


func (self *fakeAlarm) run(rcv <-chan Event, ctrl <-chan bool, send chan<- Event) {
	queue := evtHeap(make(evtHeap, 0, 2))
	heap.Init(&queue)
	closed := false

mainLoop:
	for true {
		select {

		case evt, ok := <-rcv:
			if !ok {
				if closed {
					break mainLoop
				} else {
					self.control <- false
					continue
				}
			} 
			if closed {
				continue
			}
			heap.Push(&queue, evt)

		case cmd := <- ctrl:
			if closed {
				continue
			}
			if !cmd {
				closed = true
				close(self.control)
				close(send)
			}
			if len(queue) == 0 {
				continue
			}
			send <- heap.Pop(&queue).(Event)
		}
	}
}

func NewFakeAlarm() (Alarm, FakeAlarmController) {
	in := make(chan Event, 1)
	out := make(chan Event)
	bch := make(chan bool)
	alarm := Alarm{Send: in, Receive: out}
	runner := fakeAlarm{control: bch}
	go runner.run(in, bch, out)
	return alarm, runner
}
