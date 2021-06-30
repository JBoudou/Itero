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
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JBoudou/Itero/mid/db"
	"github.com/JBoudou/Itero/mid/db/dbtest"
	"github.com/JBoudou/Itero/mid/root"
	"github.com/JBoudou/Itero/mid/service"
	"github.com/JBoudou/Itero/pkg/emailsender"
	estest "github.com/JBoudou/Itero/pkg/emailsender/emailsendertest"
	"github.com/JBoudou/Itero/pkg/events"
	evtest "github.com/JBoudou/Itero/pkg/events/eventstest"
)

func TestEmailService_CreateUserEvent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		event func(uid uint32) events.Event
	}{
		{
			name:  "CreateUserEvent",
			event: func(uid uint32) events.Event { return CreateUserEvent{User: uid} },
		},
		{
			name:  "ReverifyEvent",
			event: func(uid uint32) events.Event { return ReverifyEvent{User: uid} },
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var err error
			var rcv events.Receiver
			emailSent := false
			emailChan := make(chan bool)

			dbenv := dbtest.Env{}
			defer dbenv.Close()
			uid := dbenv.CreateUserWith(t.Name())
			dbenv.Must(t)

			locator := root.IoC.Sub()

			err = locator.Bind(func() events.Manager {
				return &evtest.ManagerMock{
					T: t,
					AddReceiver_: func(r events.Receiver) error {
						rcv = r
						rcv.Receive(tt.event(uid))
						return nil
					},
				}
			})
			mustt(t, err)

			err = locator.Bind(func() emailsender.Sender {
				return estest.SenderMock{
					T: t,
					Send_: func(emailsender.Email) error {
						emailChan <- true
						return nil
					},
				}
			})
			mustt(t, err)

			var stop service.StopFunction
			mustt(t, locator.Inject(EmailService, service.Run, &stop))
			defer stop()

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

			const qSelect = `SELECT 1 FROM Confirmations WHERE User = ? AND Type = 'verify'`
			rows, err := db.DB.Query(qSelect, uid)
			defer rows.Close()
			mustt(t, err)
			if !rows.Next() {
				t.Errorf("No confirmation created.")
			}
		})
	}
}

type emailTestInstance struct {
	name     string
	type_    string        // Type of the confirmation
	duration time.Duration // From now., for the Expires field of the confirmation
	expected bool          // Whether the confirmation should be processed/listed
}

func metaTestEmail(t *testing.T, checker func(*testing.T, *emailTestInstance, uint32)) {
	tests := []emailTestInstance{
		{
			name:     "Past",
			type_:    db.ConfirmationTypeVerify,
			duration: -1 * time.Hour,
			expected: true,
		},
		{
			name:     "Future",
			type_:    db.ConfirmationTypeVerify,
			duration: 1 * time.Hour,
			expected: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var env dbtest.Env
			defer env.Close()
			userId := env.CreateUserWith(t.Name())
			env.Must(t)

			segment, err := db.CreateConfirmation(context.Background(), userId, tt.type_, tt.duration)
			mustt(t, err)

			checker(t, &tt, segment.Id)
		})
	}
}

// ProcessOne //

func email_ProcessOne_checker(t *testing.T, tt *emailTestInstance, id uint32) {
	var svc service.Service
	mustt(t, root.IoC.Inject(EmailService, &svc))

	err := svc.ProcessOne(id)

	nothingToDoYet := false
	if errors.Is(err, service.NothingToDoYet) {
		nothingToDoYet = true
		err = nil
	}
	mustt(t, err)

	deleted := false
	const qCheck = `SELECT count(*) = 0 FROM Confirmations WHERE Id = ?`
	row := db.DB.QueryRow(qCheck, id)
	mustt(t, row.Scan(&deleted))

	if nothingToDoYet && tt.expected {
		t.Errorf("Nothing done while it should")
	}
	if !nothingToDoYet && !tt.expected {
		t.Errorf("Did something while it shouldn't")
	}
	if deleted && !tt.expected {
		t.Errorf("Confirmation deleted while it should not")
	}
	if !deleted && tt.expected {
		t.Errorf("Confirmation not deleted while it should")
	}
}

func TestEmailService_ProcessOne(t *testing.T) {
	metaTestEmail(t, email_ProcessOne_checker)
}

// CheckAll //

func email_CheckAll_checker(t *testing.T, tt *emailTestInstance, id uint32) {
	var svc service.Service
	mustt(t, root.IoC.Inject(EmailService, &svc))

	iterator := svc.CheckAll()
	defer iterator.Close()

	listed := idDateIteratorHasId(t, iterator, id)
	if !listed {
		t.Errorf("Confirmation not listed while it should")
	}
}

func TestEmailService_CheckAll(t *testing.T) {
	metaTestEmail(t, email_CheckAll_checker)
}

// CheckOne //

func email_CheckOne_checker(t *testing.T, tt *emailTestInstance, id uint32) {
	var svc service.Service
	mustt(t, root.IoC.Inject(EmailService, &svc))

	expires := svc.CheckOne(id)

	diff := tt.duration - expires.Sub(time.Now())
	if diff < -1*time.Second || diff > time.Second {
		t.Errorf("Wrong expires. Got %v. Expect %v.", expires, time.Now().Add(tt.duration))
	}
}

func TestEmailService_CheckOne(t *testing.T) {
	metaTestEmail(t, email_CheckOne_checker)
}
