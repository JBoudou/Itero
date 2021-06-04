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
	"errors"
	"time"

	"github.com/JBoudou/Itero/pkg/alarm"
	"github.com/JBoudou/Itero/pkg/events"
	"github.com/JBoudou/Itero/pkg/ioc"
)

// AlarmInjector is the injector of an alarm into services.
var AlarmInjector = alarm.New

var (
	NothingToDoYet = errors.New("Nothing to do yet")
)

// Service is the interface for sercives ran by RunService.
// Such a service must perform some operations on objects identified by an uint32 value.
// Those operations must be performed at some time.
type Service interface {

	// ProcessOne performs the operation on the object with the given id.
	// If no operation has to be done on that object yet, ProcessOne must return NothingToDoYet.
	ProcessOne(id uint32) error

	// CheckAll returns a list of all objects on which the operation will have to be done.
	// The list must be sorted in ascending order on the date.
	// In case of error, Next() called on the returned iterator must return false and Error() must
	// return the error.
	CheckAll() Iterator

	// CheckOne returns the time at which the operation must be done on the object with the given id.
	// If no operation has to be done on that object, CheckOne must return zero time.Time.
	CheckOne(id uint32) time.Time

	// Interval returns the maximal duration between two full check of the object to proceed.
	Interval() time.Duration

	Logger() LevelLogger
}

// EventReceiver is the interface implemented by services willing to react to some events.
type EventReceiver interface {
	FilterEvent(events.Event) bool
	ReceiveEvent(events.Event, RunnerControler)
}

// Iterator iterates on a list of Id and Date representing tasks for a service.
type Iterator interface {

	// Next goes to the next entry if it can, returning false otherwise.
	// Returning true guarantees that a call to IdAndDate will succeed.
	// Next must be called before once before the first call to IdAndDate.
	Next() bool

	IdAndDate() (uint32, time.Time)
	Err() error
	Close() error
}

// RunnerControler allows to control the service runner from the service.
// It should be used only from EventReceiver.ReceiveEvent().
type RunnerControler interface {

	// Schedule asks the runner to schedule the object with the given id for being processed.
	Schedule(id uint32)

	// StopService asks the runner to stop the service as soon as possible.
	StopService()
}

// StopFunction must be called to cleanly stop a service.
type StopFunction func()

// RunService runs a service in the background.
//
// All methods of the service are called from the same goroutine, wich is different from the
// goroutine RunService was run from.
// If the service implements the EventReceiver interface, the runner installs an AsyncForwarder on
// events.DefaultManager and calls EventReceiver.ReceiveEvent for each received event.
// The returned function must be called to stop the service and free the resources associated with
// the runner.
func Run(service interface{}, loc *ioc.Locator) StopFunction {
	runner := &serviceRunner{}
	err := loc.Inject(service, &runner.service)
	if err != nil {
		panic(err)
	}

	if eventReceiver, ok := service.(EventReceiver); ok {
		evtChan := make(chan events.Event, 64)
		var evtManager events.Manager
		loc.Get(&evtManager)
		evtManager.AddReceiver(events.AsyncForwarder{
			Filter: eventReceiver.FilterEvent,
			Chan:   evtChan,
		})
		go runner.runWithEvents(evtChan, eventReceiver)

	} else {
		go runner.run()

	}
	return runner.StopService
}

//
// Implementation //
//

const (
	// In this first implementation of the new service framework, when the date of a task is already
	// over but ProcessOne returned NothingToDoYet, then the task is scheduled after rescheduleDelay.
	rescheduleDelay = time.Second

	// Maximal number of ids that are considered at each full check.
	maxHandledIds = 1024
)

type runner interface {
	run()
}

type serviceRunner struct {
	service       Service
	alarm         alarm.Alarm
	lastFullCheck time.Time
	stopped       chan struct{}
}

func (self *serviceRunner) run() {
	self.init()
mainLoop:
	for true {
		select {

		case evt, ok := <-self.alarm.Receive:
			if !ok {
				break mainLoop
			}
			self.handleEvent(evt)

		case <-self.stopped:
			break mainLoop

		}
	}
}

func (self *serviceRunner) runWithEvents(evtCh <-chan events.Event, receiver EventReceiver) {
	self.init()
mainLoop:
	for true {
		select {

		case evt, ok := <-self.alarm.Receive:
			if !ok {
				break mainLoop
			}
			self.handleEvent(evt)

		case evt, ok := <-evtCh:
			if !ok {
				break mainLoop
			}
			receiver.ReceiveEvent(evt, self)

		case <-self.stopped:
			break mainLoop

		}
	}
}

func (self *serviceRunner) init() {
	self.alarm = AlarmInjector(maxHandledIds, alarm.DiscardLateDuplicates)
	self.stopped = make(chan struct{})
	self.fullCheck()
}

func (self *serviceRunner) StopService() {
	close(self.stopped)
}

func (self *serviceRunner) fullCheck() {
	self.lastFullCheck = time.Now()
	it := self.service.CheckAll()
	defer it.Close()

	scheduled := 0
	for it.Next() && scheduled < maxHandledIds {
		id, date := it.IdAndDate()

		if date.Before(time.Now()) {
			if self.processWithDate(id, date) {
				scheduled += 1
			}
		} else {
			if self.schedule(id, date) {
				return
			}
		}
	}

	if scheduled == 0 {
		self.scheduleFullCheck()
	}
}

func (self *serviceRunner) schedule(id uint32, date time.Time) bool {
	if date.IsZero() {
		date = self.service.CheckOne(id)
		if date.IsZero() {
			self.service.Logger().Logf("Nothing to do for %d", id)
			return false
		}
	}

	minFuture := time.Now().Add(rescheduleDelay)
	if date.Before(minFuture) {
		date = minFuture
	}
	self.alarm.Send <- alarm.Event{Time: date, Data: id}
	self.service.Logger().Logf("Next action %v for %d", date, id)
	return true
}

func (self *serviceRunner) scheduleFullCheck() {
	date := self.lastFullCheck.Add(self.service.Interval())
	self.alarm.Send <- alarm.Event{Time: date}
	self.service.Logger().Logf("Next full check at %v", date)
}

func (self *serviceRunner) handleEvent(evt alarm.Event) {
	if evt.Data == nil {
		self.fullCheck()
	} else {
		sent := self.processNoDate(evt.Data.(uint32))
		if !sent && evt.Remaining == 0 {
			self.scheduleFullCheck()
		}
	}
}

func (self *serviceRunner) processNoDate(id uint32) bool {
	return self.processWithDate(id, time.Time{})
}

func (self *serviceRunner) processWithDate(id uint32, date time.Time) bool {
	err := self.service.ProcessOne(id)
	if err == nil {
		self.service.Logger().Logf("Done for %d", id)
		return self.schedule(id, time.Time{})
	}
	if errors.Is(err, NothingToDoYet) {
		return self.schedule(id, date)
	}
	self.service.Logger().Errorf("Error processing %d: %v", id, err)
	return false
}

func (self *serviceRunner) Schedule(id uint32) {
	self.schedule(id, time.Time{})
}
