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
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/JBoudou/Itero/alarm"
	"github.com/JBoudou/Itero/db"
	"github.com/JBoudou/Itero/events"
)

// InjectAlarmInService is the injector of an alarm into services.
var InjectAlarmInService = injectAlarmInService

func injectAlarmInService() alarm.Alarm {
	return alarm.New(16, alarm.DiscardLaterEvent, alarm.DiscardDuplicates)
}

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
	CheckAll() IdAndDateIterator

	// CheckOne returns the time at which the operation must be done on the object with the given id.
	// If no operation has to be done on that object, CheckOne must return zero time.Time.
	CheckOne(id uint32) time.Time

	// CheckInterval returns the maximal duration between two full check of the object to proceed.
	CheckInterval() time.Duration

	Logger() LevelLogger
}

// EventReceiver is the interface implemented by services willing to react to some events.
type EventReceiver interface {
	FilterEvent(events.Event) bool
	ReceiveEvent(events.Event, ServiceRunnerControl)
}

type IdAndDateIterator interface {

	// Next goes to the next entry if it can, returning false otherwise.
	// Returning true guarantees that a call to IdAndDate will succeed.
	// Next must be called before once before the first call to IdAndDate.
	Next() bool

	IdAndDate() (uint32, time.Time)
	Err() error
	Close() error
}

// LevelLogger is a temporary interface before the new logger facility is implemented.
type LevelLogger interface {
	Logf(format string, v ...interface{})
	Warnf(format string, v ...interface{})
	Errorf(format string, v ...interface{})
}

// EasyLogger is a temporary implementation of LevelLogger.
type EasyLogger struct{}

func (self EasyLogger) Logf(format string, v ...interface{}) {
	log.Printf(format, v...)
}

func (self EasyLogger) Warnf(format string, v ...interface{}) {
	log.Printf("Warn: "+format, v...)
}

func (self EasyLogger) Errorf(format string, v ...interface{}) {
	log.Printf("Err: "+format, v...)
}

type ServiceRunnerControl interface {

	// Schedule asks the runner to schedule the object with the given id for being processed.
	Schedule(id uint32)

	// StopService asks the runner to stop the service as soon as possible.
	StopService()
}

type StopServiceFunc func()

// RunService runs a service in the background.
//
// All methods of the service are called from the same goroutine, wich is different from the
// goroutine RunService was run from.
// If the service implements the EventReceiver interface, the runner installs an AsyncForwarder on
// events.DefaultManager and calls EventReceiver.ReceiveEvent for each received event.
// The returned function must be called to stop the service and free the resources associated with
// the runner.
func RunService(service Service) StopServiceFunc {
	runner := &serviceRunner{service: service}

	if eventReceiver, ok := service.(EventReceiver); ok {
		evtChan := make(chan events.Event, 64)
		events.AddReceiver(events.AsyncForwarder{
			Filter: eventReceiver.FilterEvent,
			Chan:   evtChan,
		})
		go runner.runWithEvents(evtChan, eventReceiver)

	} else {
		go runner.run()

	}
	return runner.StopService
}

// Implementation //

// In this first implementation of the new service framework, when the date of a task is already
// over but ProcessOne returned NothingToDoYet, then the task is scheduled after rescheduleDelay.
const rescheduleDelay = time.Second

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
	self.alarm = InjectAlarmInService()
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

	for it.Next() {
		id, date := it.IdAndDate()

		if date.Before(time.Now()) {
			if self.processWithDate(id, date) {
				return
			}
		} else {
			if self.schedule(id, date) {
				return
			}
		}
	}

	self.scheduleFullCheck()
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
	self.alarm.Send <- alarm.Event{Time: self.lastFullCheck.Add(self.service.CheckInterval())}
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
		return false
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

// OLD CODE //

const minWaitDuration = time.Second

// pollService is the base classes of some services like nextRound and closePoll.
type pollService struct {
	lastCheck time.Time
	adjust    time.Duration
	step      uint8

	// Logger to use for warning. Will be replaced by a leveled logger in near future.
	warn *log.Logger

	serviceName string
	makeEvent   func(pollId uint32) events.Event
}

// newPollService creates a new pollService.
func newPollService(serviceName string, makeEvent func(pollId uint32) events.Event) pollService {
	return pollService{
		adjust:      1 * time.Minute,
		serviceName: serviceName,
		warn:        log.New(os.Stderr, serviceName, log.LstdFlags|log.Lshortfile|log.Lmsgprefix),
		makeEvent:   makeEvent,
	}
}

// fullCheck_helper helps to implement full checks.
//
// Query qSelect selects all the ids (uint32) on which an update is needed. This query takes one
// parameter which is lastCheck.
// Query qUpdate is then called on each collected id. It takes the id as only parameter.
// If the query is successful, an event constructed by makeEvent is emitted and step is reset to 0.
func (self *pollService) fullCheck_helper(qSelect, qUpdate string) error {
	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}
	commited := false
	defer func() {
		if !commited {
			tx.Rollback()
		}
	}()

	closeSet := make(map[uint32]bool)
	if err := self.collectId(closeSet, tx, qSelect); err != nil {
		return err
	}
	if err := self.execUpdate(closeSet, tx, qUpdate); err != nil {
		return err
	}

	// Commit
	if err := tx.Commit(); err != nil {
		return err
	}
	commited = true

	// Send events
	for id := range closeSet {
		events.Send(self.makeEvent(id))
	}

	self.step = 0
	log.Print(self.serviceName, " fullCkeck terminated.")
	return nil
}

// collectId is an internal (private) method.
func (self pollService) collectId(set map[uint32]bool, tx *sql.Tx, query string) error {
	rows, err := tx.Query(query)
	if err != nil {
		return err
	}
	for rows.Next() {
		var key uint32
		if err := rows.Scan(&key); err != nil {
			return nil
		}
		set[key] = true
	}
	return nil
}

// execUpdate is an internal (private) method
func (self pollService) execUpdate(set map[uint32]bool, tx *sql.Tx, query string) error {
	stmt, err := tx.Prepare(query)
	if err != nil {
		return err
	}
	for id := range set {
		if _, err := stmt.Exec(id); err != nil {
			return err
		}
	}
	stmt.Close()
	return nil
}

// checkOne_helper helps to implement a checker for one identifier (usually a poll).
//
// The function is assumed to execute a query updating or not the poll with id pollId.
// If the query affects exactly one row, an event constructed by makeEvent is emitted,
// and step is increased.
func (self *pollService) checkOne_helper(pollId uint32, fct func() (sql.Result, error)) error {
	result, err := fct()
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected > 1 {
		return errors.New(fmt.Sprintf("More than one poll with Id %d. No event send.", pollId))
	}
	if affected == 1 {
		events.Send(self.makeEvent(pollId))
	}

	self.step++
	log.Printf("%s Check for poll %d terminated.", self.serviceName, pollId)
	return nil
}

// updateLastCheck update the lastCheck field.
// The new value is computed from the database.
func (self *pollService) updateLastCheck() {
	row := db.DB.QueryRow(`SELECT CURRENT_TIMESTAMP()`)
	if err := row.Scan(&self.lastCheck); err != nil {
		self.warn.Print(err)
	}
	self.adjust = (self.adjust - time.Since(self.lastCheck)) / 2
}

// nextAlarm_helper helps to implement a method providing events for an alarm.
//
// The query must take one parameter, which is set to lastCheck, and must return 3 values: the id
// (uint32) to inspect when the alarm will fire, the time the alarm must fire, and the current time
// for the database. This query must return zero or one row. If it returns zero row, the returned
// event has Time computed using defaultWait and Data nil.
func (self *pollService) nextAlarm_helper(qNext string, defaultWait time.Duration) (ret alarm.Event) {
	var pollId uint32
	var timestamp time.Time
	row := db.DB.QueryRow(qNext, self.lastCheck)
	if err := row.Scan(&pollId, &ret.Time, &timestamp); err == nil {
		self.adjust = (self.adjust - time.Since(timestamp)) / 2
		ret.Time = ret.Time.Add(self.adjust)
		if time.Until(ret.Time) < minWaitDuration {
			ret.Time = time.Now().Add(minWaitDuration)
		}
		ret.Data = pollId
	} else {
		if !errors.Is(err, sql.ErrNoRows) {
			self.warn.Print(err)
		}
		ret.Time = time.Now().Add(defaultWait)
	}

	log.Printf("%s Next alarm at %v.", self.serviceName, ret.Time)
	return
}
