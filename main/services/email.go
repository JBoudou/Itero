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
	"path/filepath"
	"text/template"
	"time"

	"github.com/JBoudou/Itero/mid/db"
	"github.com/JBoudou/Itero/mid/server"
	"github.com/JBoudou/Itero/mid/service"
	"github.com/JBoudou/Itero/pkg/config"
	"github.com/JBoudou/Itero/pkg/emailsender"
	"github.com/JBoudou/Itero/pkg/events"
	"github.com/JBoudou/Itero/pkg/ioc"
)

// TmplBaseDir is the directory to find email templates into.
const TmplBaseDir = "email"

// StartEmailService starts the email service.
// This service listen to events and send emails to users.
func StartEmailService(evtManager events.Manager,
	sender emailsender.Sender, logger service.LevelLogger) error {

	evtChan := make(chan events.Event, 128)
	err := evtManager.AddReceiver(events.AsyncForwarder{
		Filter: func(evt events.Event) bool {
			switch evt.(type) {
			case CreateUserEvent:
				return true
			}
			return false
		},
		Chan: evtChan,
	})
	if err != nil {
		return err
	}

	go emailService{
		sender:  sender,
		evtChan: evtChan,
		log:     logger,
	}.run()
	return nil
}

//
// Implementation
//

var emailConfig struct {
	Sender string
}

func init() {
	// IoC
	ioc.Root.Set(func() (emailsender.Sender, error) {
		options := emailsender.BatchSenderOptions{
			MinBatchLen: 2,
			MaxDelay:    "1m",
			SMTP:        "localhost:25",
		}

		err := config.Value("emails", &options)
		if err != nil {
			return nil, err
		}

		return emailsender.StartBatchSender(options)
	})

	// Config
	config.Value("emails", &emailConfig)
}

type emailService struct {
	sender  emailsender.Sender
	evtChan <-chan events.Event
	log     service.LevelLogger
}

func (self emailService) run() {
	for evt := range self.evtChan {
		self.handleEvent(evt)
	}
}

func (self emailService) handleEvent(evt events.Event) {
	switch converted := evt.(type) {
	case CreateUserEvent:
		self.createUser(converted.User)
	}
}

func (self emailService) createUser(userId uint32) {
	var data struct {
		Sender       string
		Name         string
		Address      string
		BaseURL      string
		Confirmation string
	}
	data.Sender = emailConfig.Sender
	data.BaseURL = server.BaseURL()

	// Find the template
	var tmpl *template.Template
	tmpl, err := template.ParseFiles(filepath.Join(config.BaseDir, TmplBaseDir, "en", "greeting.txt"))
	if err != nil {
		self.log.Errorf("Error retrieving template: %v", err)
		return
	}

	// Retrieve user data
	const qSelect = `
	  SELECT Name, Email FROM Users WHERE Id = ? AND Name IS NOT NULL AND Email IS NOT NULL`
	rows, err := db.DB.Query(qSelect, userId)
	defer rows.Close()
	if err != nil {
		self.log.Errorf("Error retrieving user %d: %v", userId, err)
		return
	}
	if !rows.Next() {
		self.log.Errorf("User %d not found", userId)
		return
	}
	err = rows.Scan(&data.Name, &data.Address)
	if err != nil {
		self.log.Errorf("Error retrieving user %d: %v", userId, err)
		return
	}

	// Create the confirmation
	segment, err := db.CreateConfirmation(context.Background(), userId, db.ConfirmationTypeVerify, 48 * time.Hour)
	if err != nil {
		self.log.Errorf("Error creating confirmation %v.", err)
		return
	}
	data.Confirmation, err = segment.Encode()
	if err != nil {
		self.log.Errorf("Error encoding confirmation %v.", err)
		return
	}

	// Send the email
	err = self.sender.Send(emailsender.Email{
		To:   []string{data.Address},
		Tmpl: tmpl,
		Data: data,
	})
	if err != nil {
		self.log.Errorf("Error sending email: %v", err)
		return
	}
}
