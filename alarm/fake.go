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
	"container/heap"
	"sync"
)

type FakeAlarmController interface {
	// Tick asks the fake alarm to send its most recent event.
	Tick()

	Close()

	QueueLength() int
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

// fakeAlarm is both an implementation of FakeAlarmController and an alarm runner.
type fakeAlarm struct {
	control chan struct{}

	locker *sync.Mutex
	queue  evtHeap
}

func newFakeAlarm() *fakeAlarm {
	queue := evtHeap(make(evtHeap, 0, 2))
	heap.Init(&queue)
	return &fakeAlarm{
		control: make(chan struct{}),
		locker: new(sync.Mutex),
		queue: queue,
	}
}

func (self fakeAlarm) Tick() {
	self.control <- struct{}{}
}

func (self fakeAlarm) Close() {
	close(self.control)
}

func (self fakeAlarm) QueueLength() int {
	self.locker.Lock()
	defer self.locker.Unlock()
	return len(self.queue)
}

func (self *fakeAlarm) run(rcv <-chan Event, send chan<- Event) {
	closed := false

mainLoop:
	for true {
		select {

		case evt, ok := <-rcv:
			if !ok {
				if closed {
					break mainLoop
				} else {
					self.Close()
					continue
				}
			}
			if closed {
				continue
			}

			self.locker.Lock()
			heap.Push(&self.queue, evt)
			self.locker.Unlock()

		case _, ok := <-self.control:
			if closed {
				continue
			}
			if !ok {
				closed = true
				close(send)
				continue
			}

			self.locker.Lock()
			if len(self.queue) == 0 {
				self.locker.Unlock()
				continue
			}
			evt := heap.Pop(&self.queue).(Event)
			evt.Remaining = len(self.queue)
			self.locker.Unlock()

			send <- evt
		}
	}
}

func NewFakeAlarm() (Alarm, FakeAlarmController) {
	in := make(chan Event, 1)
	out := make(chan Event)
	alarm := Alarm{Send: in, Receive: out}
	runner := newFakeAlarm()
	go runner.run(in, out)
	return alarm, runner
}
