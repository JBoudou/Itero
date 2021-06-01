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

package salted

import (
	"testing"
)

func TestSegment(t *testing.T) {
	tests := []struct {
		name    string
		segment Segment
	}{
		{
			name:    "Simple",
			segment: Segment{Id: 0xF1234567, Salt: 0x312345},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := tt.segment.Encode()
			if err != nil {
				t.Fatalf("Encode error: %s", err)
			}
			got, err := Decode(encoded)
			if err != nil {
				t.Fatalf("Decode error: %s", err)
			}
			if got != tt.segment {
				t.Errorf("Got %v. Expect %v", got, tt.segment)
			}
		})
	}
}

func TestNew(t *testing.T) {
	const id = 42
	segment, err := New(id)
	if err != nil {
		t.Errorf("New returned %v.", err)
	}
	if segment.Id != id {
		t.Errorf("Wrong id. Got %d. Expect %d.", segment.Id, id)
	}
}
