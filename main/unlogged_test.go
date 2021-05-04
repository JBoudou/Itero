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

package main

import (
	"bytes"
	"context"
	"reflect"
	"testing"

	"github.com/JBoudou/Itero/db"
)

func TestHashAddr(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{name: "Simple", input: "192.168.26.0:7080"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var got uint32 = HashAddr(tt.input)
			if got&0xFF000000 != 0 {
				t.Errorf("Hash is not 24 bits long for %s.", tt.input)
			}
		})
	}
}

func TestLE24Bits(t *testing.T) {
	tests := []struct {
		name   string
		input  uint32
		expect []byte
	}{
		{name: "first byte", input: 1, expect: []byte{1, 0, 0}},
		{name: "second byte", input: 0x8000, expect: []byte{0, 0x80, 0}},
		{name: "third byte", input: 0x180000, expect: []byte{0, 0, 0x18}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := LE24Bits(tt.input)
			if bytes.Compare(got, tt.expect) != 0 {
				t.Errorf("Got %v. Expect %v", got, tt.expect)
			}
		})
	}
}

func TestUnloggedFromHash(t *testing.T) {
	tests := []struct {
		name string
		hash uint32
	}{
		{name: "Simple", hash: 42},
	}

	const (
		qDelete = `DELETE FROM Users WHERE Hash = ?`
		qSelect = `SELECT Id FROM Users WHERE Hash = ?`
	)

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			bin := LE24Bits(tt.hash)
			_, err := db.DB.Exec(qDelete, bin)
			mustt(t, err)

			got1, err := UnloggedFromHash(context.Background(), tt.hash)
			mustt(t, err)

			if got1.Hash != tt.hash {
				t.Errorf("Wrong hash. Got %X. Expect %X.", got1.Hash, tt.hash)
			}
			if got1.Logged {
				t.Errorf("Got logged true")
			}

			rows, err := db.DB.Query(qSelect, bin)
			mustt(t, err)
			if !rows.Next() {
				t.Fatalf("No pseudo-user created")
			}
			var expectId uint32
			mustt(t, rows.Scan(&expectId))
			if got1.Id != expectId {
				t.Errorf("Wrong Id. Got %d. Expect %d.", got1.Id, expectId)
			}

			got2, err := UnloggedFromHash(context.Background(), tt.hash)
			mustt(t, err)
			if !reflect.DeepEqual(got1, got2) {
				t.Errorf("Not the same pseudo-user on second call. Got %v. Expect %v.", got2, got1)
			}
		})
	}
}
