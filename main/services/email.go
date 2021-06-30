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
	"github.com/JBoudou/Itero/mid/root"
	"github.com/JBoudou/Itero/mid/server"
	"github.com/JBoudou/Itero/mid/service"
	"github.com/JBoudou/Itero/pkg/config"
	"github.com/JBoudou/Itero/pkg/emailsender"
	"github.com/JBoudou/Itero/pkg/events"
	"github.com/JBoudou/Itero/pkg/slog"
)

// TmplBaseDir is the directory to find email templates into.
const TmplBaseDir = "email"

// EmailService constructs the email service.
// This service listen to events and send emails to users.
func EmailService(sender emailsender.Sender, log slog.StackedLeveled) emailService {
	return emailService{
		sender: sender,
		log:    log.With("Email"),
	}
}

//
// Implementation
//

var emailConfig struct {
	Sender string
}

func init() {
	// IoC
	root.IoC.Bind(func() (emailsender.Sender, error) {
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
	sender emailsender.Sender
	log    slog.Leveled
}

func (self emailService) ProcessOne(id uint32) error {
	const qDelete = `DELETE FROM Confirmations WHERE Id = ? AND Expires < CURRENT_TIMESTAMP`
	return service.SQLProcessOne(qDelete, id)
}

func (self emailService) CheckAll() service.Iterator {
	const qList = `SELECT Id, Expires FROM Confirmations`
	return service.SQLCheckAll(qList)
}

func (self emailService) CheckOne(id uint32) (ret time.Time) {
	const qCheck = `SELECT Expires FROM Confirmations WHERE Id = ?`
	rows, err := db.DB.Query(qCheck, id)
	defer rows.Close()
	if err == nil && rows.Next() {
		err = rows.Scan(&ret)
	}
	if err != nil {
		self.log.Errorf("Error in CheckOne: %v", err)
	}
	return
}

func (self emailService) Interval() time.Duration {
	return 24 * time.Hour
}

func (self emailService) Logger() slog.Leveled {
	return self.log
}

func (self emailService) FilterEvent(evt events.Event) bool {
	switch evt.(type) {
	case CreateUserEvent, ReverifyEvent, ForgotEvent:
		return true
	}
	return false
}

func (self emailService) ReceiveEvent(evt events.Event, ctrl service.RunnerControler) {
	switch converted := evt.(type) {
	case CreateUserEvent:
		self.confirmationEmail(converted.User, ctrl, "greeting.txt", db.ConfirmationTypeVerify, 48*time.Hour)
	case ReverifyEvent:
		self.confirmationEmail(converted.User, ctrl, "reverify.txt", db.ConfirmationTypeVerify, 48*time.Hour)
	case ForgotEvent:
		self.confirmationEmail(converted.User, ctrl, "forgot.txt", db.ConfirmationTypePasswd, 3*time.Hour)
	}
}

func (self emailService) confirmationEmail(userId uint32, ctrl service.RunnerControler,
	tmplFile string, confirmType db.ConfirmationType, confirmDuration time.Duration) {
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
	tmpl, err := template.ParseFiles(filepath.Join(config.BaseDir, TmplBaseDir, "en", tmplFile))
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
	rows.Close()

	// Create the confirmation
	segment, err := db.CreateConfirmation(context.Background(), userId, confirmType, confirmDuration)
	if err != nil {
		self.log.Errorf("Error creating confirmation %v.", err)
		return
	}
	ctrl.Schedule(segment.Id)
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
