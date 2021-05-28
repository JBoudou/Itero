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

package emailsender

import (
	"time"
)

type BatchSenderOptions struct {
	MinBatchLen int
	MaxDelay    string // string representation of a duration
	Sender      string // email address
	SMTP        string // host:port
}

func NewBatchSender(options BatchSenderOptions) (Sender, error) {
	// TODO
	return nil, nil
}

//
// expiringStore
//

type expiringStore struct {
	first *expiringNode
	last  *expiringNode
	wait  time.Duration
	size  int
}

type expiringNode struct {
	email Email
	due   time.Time
	next  *expiringNode
}

func newExpiringStore(maxWait time.Duration) *expiringStore {
	return &expiringStore{wait: maxWait}
}

func (self *expiringStore) Push(email Email) {
	node := &expiringNode{
		email: email,
		due:   time.Now().Add(self.wait),
	}
	if self.last == nil {
		self.first = node
		self.last = node
		self.size = 1
	} else {
		self.last.next = node
		self.last = node
		self.size += 1
	}
}

func (self *expiringStore) Pop() (email Email) {
	email = self.first.email
	self.first = self.first.next
	if self.first == nil {
		self.last = nil
		self.size = 0
	} else {
		self.size -= 1
	}
	return
}

func (self *expiringStore) Empty() bool {
	return self.first == nil
}

func (self *expiringStore) Len() int {
	return self.size
}

func (self *expiringStore) Overdue() bool {
	return self.first != nil && time.Now().After(self.first.due)
}

//
// batchSender
//

type batchSender struct {
	emailChan chan Email
	ticker    *time.Ticker
	store     *expiringStore
	back      Sender
	minLen    int
	open      bool
	closed    bool
}

// newBatchSender set a new batchSender, except field back
func newBatchSender(maxWait time.Duration, minLen int) *batchSender {
	return &batchSender{
		emailChan: make(chan Email, minLen * 2),
		ticker: time.NewTicker(maxWait / 10),
		store: newExpiringStore(maxWait),
		minLen: minLen,
	}
}

func (self *batchSender) Send(email Email) error {
	if len(email.To) < 1 || email.Tmpl == nil {
		return WrongEmailValue
	}
	self.emailChan <- email
	return nil
}

func (self *batchSender) Close() error {
	close(self.emailChan)
	return nil
}

func (self *batchSender) run() {
	for !self.closed {
		if self.open {
			self.nonblocking()
		} else {
			self.blocking()
		}
	}
}

func (self *batchSender) blocking() {
	select {
	case email, ok := <-self.emailChan:
		self.handleEmail(email, ok)

	case _ = <-self.ticker.C:
		self.open = self.store.Overdue()
	}
}

func (self *batchSender) nonblocking() {
	self.back.Send(self.store.Pop())

	select {
	case email, ok := <-self.emailChan:
		self.handleEmail(email, ok)

	default:
		if self.store.Empty() {
			if q, ok := self.back.(interface{ Quit() error }); ok {
				q.Quit()
			}
			self.open = false
		}
	}
}

func (self *batchSender) handleEmail(email Email, ok bool) {
	if !ok {
		self.ticker.Stop()
		self.back.Close()
		self.closed = true
		return
	}

	self.store.Push(email)
	if self.store.Len() >= self.minLen {
		self.open = true
	}
}
