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

	"github.com/JBoudou/Itero/mid/db/dbtest"
	"github.com/JBoudou/Itero/pkg/emailsender"
	estest "github.com/JBoudou/Itero/pkg/emailsender/emailsendertest"
	"github.com/JBoudou/Itero/pkg/events"
	evtest "github.com/JBoudou/Itero/pkg/events/eventstest"
	"github.com/JBoudou/Itero/pkg/ioc"
)

func TestEmailService(t *testing.T) {
	var err error
	var rcv events.Receiver
	emailSent := false
	emailChan := make(chan bool)

	dbenv := dbtest.Env{}
	defer dbenv.Close()
	uid := dbenv.CreateUserWith(t.Name())
	dbenv.Must(t)
	
	locator := ioc.Root.Sub()

	err = locator.Set(func () events.Manager {
		return &evtest.ManagerMock{
			T: t,
			AddReceiver_: func(r events.Receiver) error {
				rcv = r
				rcv.Receive(CreateUserEvent{User: uid})
				return nil
			},
		}
	})
	mustt(t, err)

	err = locator.Set(func () emailsender.Sender {
		return estest.SenderMock{
			T: t,
			Send_: func (emailsender.Email) error {
				emailChan <- true
				return nil
			},
		}
	})
	mustt(t, err)

	err = locator.Get(StartEmailService)
	mustt(t, err)

testLoop:
	for i := 0; i < 20; i++ {
		select {
		case emailSent = <-emailChan:
			break testLoop

		default:
			time.Sleep(10 * time.Millisecond)
		}
	}

	if rcv == nil {
		t.Errorf("Receiver not registered")
	}
	if !emailSent {
		t.Errorf("No email sent")
	}
}
