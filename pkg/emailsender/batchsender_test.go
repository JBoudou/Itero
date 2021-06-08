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
	"reflect"
	"testing"
	"text/template"
	"time"
)

type recorderSender struct {
	records [][]string
}

func (self *recorderSender) Send(email Email) error {
	self.records = append(self.records, email.To)
	return nil
}

func (self *recorderSender) Close() error {
	return nil
}

func TestBatchSender(t *testing.T) {
	const (
		maxWait = 200 * time.Millisecond
		minLen  = 3
	)

	tests := []struct {
		name   string
		to     [][]string
		wait   time.Duration
		expect int // number of sent email
	}{
		{
			name:   "too fast",
			to:     [][]string{{"one"}, {"two"}},
			wait:   20 * time.Millisecond,
			expect: 0,
		},
		{
			name: "by number",
			to:     [][]string{{"one"}, {"two"},{"three"}},
			wait:   30 * time.Millisecond,
			expect: 3,
		},
		{
			name:   "by time",
			to:     [][]string{{"one"}, {"two"}},
			wait:   maxWait + (10 * time.Millisecond),
			expect: 2,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			sender := newBatchSender(maxWait, minLen)
			recorder := recorderSender{}
			sender.back = &recorder
			go sender.run()

			for _, to := range tt.to {
				sender.Send(Email{
					To:   to,
					Tmpl: &template.Template{},
				})
			}
			if tt.wait > 0 {
				time.Sleep(tt.wait)
			}
			sender.Close()

			recLen := len(recorder.records)
			if recLen != tt.expect {
				t.Errorf("Wrong record length. Got %d. Expect %d.", recLen, tt.expect)
				if tt.expect < recLen {
					recLen = tt.expect
				}
			}
			for i := 0; i < recLen; i++ {
				if !reflect.DeepEqual(recorder.records[i], tt.to[i]) {
					t.Errorf("Wrong email index %d. Got %v. Expect %v.", i, recorder.records[i], tt.to[i])
				}
			}
		})
	}
}
