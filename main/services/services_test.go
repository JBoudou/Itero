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
	"reflect"
	"testing"

	"github.com/JBoudou/Itero/mid/root"
	"github.com/JBoudou/Itero/mid/service"
	"github.com/JBoudou/Itero/pkg/events"
)

func mustt(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

//
// checkEventSchedule
//

type mockRunnerController struct {
	schedule []uint32
	stop     bool
}

func (self *mockRunnerController) Schedule(id uint32) {
	self.schedule = append(self.schedule, id)
}

func (self *mockRunnerController) StopService() {
	self.stop = true
}

type checkEventScheduleTest struct {
	name     string
	event    events.Event
	schedule []uint32
}

func checkEventSchedule(t *testing.T, tests []checkEventScheduleTest, factory interface{}) {
	t.Parallel()
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var svc service.EventReceiver
			mustt(t, root.IoC.Inject(factory, &svc))

			controler := &mockRunnerController{}
			if svc.FilterEvent(tt.event) {
				svc.ReceiveEvent(tt.event, controler)
			}

			if !reflect.DeepEqual(controler.schedule, tt.schedule) {
				t.Errorf("Wrong calls to Schedule. Got %v. Expect %v.", controler.schedule, tt.schedule)
			}
		})
	}
}
