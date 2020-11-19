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

package b64buff

import (
	"testing"
)

func TestUInt32UInt32(t *testing.T) {
	type entry struct {
		value  uint32
		nbBits uint8
	}
	tests := []struct {
		name  string
		write []entry
		read  []entry
	}{
		{
			name:  "zero",
			write: []entry{{value: 0x7, nbBits: 0}},
			read:  []entry{{value: 0x0, nbBits: 0}},
		},
		{
			name:  "write single 6",
			write: []entry{{value: 0x27, nbBits: 6}},
			read:  []entry{{value: 0x27, nbBits: 6}},
		},
		{
			name: "write aligned",
			write: []entry{
				{value: 0x22, nbBits: 6},
				{value: 0x345, nbBits: 12},
				{value: 0x26789, nbBits: 18},
			},
			read: []entry{
				{value: 0x88d166, nbBits: 24},
				{value: 0x789, nbBits: 12},
			},
		},
		{
			name: "write unfull filling",
			write: []entry{
				{value: 0x17, nbBits: 5},
				{value: 0x42, nbBits: 7},
			},
			read: []entry{
				{value: 0xbc2, nbBits: 12},
			},
		},
		{
			name: "write chained overflow read all",
			write: []entry{
				{value: 0x178, nbBits: 9},
				{value: 0x2A, nbBits: 7},
			},
			read: []entry{
				{value: 0xbc2a, nbBits: 16},
			},
		},
		{
			name: "write small finally aligned",
			write: []entry{
				{value: 0x2, nbBits: 2},
				{value: 0x2, nbBits: 3},
				{value: 0x1, nbBits: 1},
			},
			read: []entry{
				{value: 0x25, nbBits: 6},
			},
		},
		{
			name: "read small finally aligned",
			write: []entry{
				{value: 0x25, nbBits: 6},
			},
			read: []entry{
				{value: 0x2, nbBits: 2},
				{value: 0x2, nbBits: 3},
				{value: 0x1, nbBits: 1},
			},
		},
		{
			name: "read chained overflow",
			write: []entry{
				{value: 0xbc2a, nbBits: 16},
			},
			read: []entry{
				{value: 0x178, nbBits: 9},
				{value: 0x2A, nbBits: 7},
			},
		},
		{
			name: "read unaligned",
			write: []entry{
				{value: 0x13c2a, nbBits: 17},
			},
			read: []entry{
				{value: 0x4, nbBits: 3},
				{value: 0x1E15, nbBits: 13},
				{value: 0x0, nbBits: 1},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer := Buffer{}
			for _, write := range tt.write {
				if err := buffer.WriteUInt32(write.value, write.nbBits); err != nil {
					t.Fatalf("Writing error: %s", err)
				}
			}
			for _, read := range tt.read {
				result, err := buffer.ReadUInt32(read.nbBits)
				if err != nil {
					t.Fatalf("Read error: %s", err)
				}
				if result != read.value {
					t.Errorf("Got %x. Expect %x.", result, read.value)
				}
			}
		})
	}
}

func TestWriteUInt32(t *testing.T) {
	buffer := Buffer{}
	if err := buffer.WriteUInt32(0x27, 33); err != WrongNbBits {
		t.Errorf("Got %s. Expect %s", err, WrongNbBits)
	}
}

func TestUReadUInt32(t *testing.T) {
	type entry struct {
		value  uint32
		nbBits uint8
	}
	tests := []struct {
		name   string
		write  []entry
		nbBits uint8
		expect error
	}{
		{
			name: "Too big",
			write: []entry{
				{value: 0x01234567, nbBits: 32},
				{value: 0x76543210, nbBits: 32},
			},
			nbBits: 33,
			expect: WrongNbBits,
		},
		{
			name:   "Empty",
			write:  []entry{},
			nbBits: 3,
			expect: NotEnoughData,
		},
		{
			name:   "Small",
			write:  []entry{{value: 0x111, nbBits: 9}},
			nbBits: 10,
			expect: NotEnoughData,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer := Buffer{}
			for _, write := range tt.write {
				if err := buffer.WriteUInt32(write.value, write.nbBits); err != nil {
					t.Fatalf("Writing error: %s", err)
				}
			}
			if _, err := buffer.ReadUInt32(tt.nbBits); err != tt.expect {
				t.Errorf("Got %s. Expect %s", err, tt.expect)
			}
		})
	}
}

func TestThroughB64(t *testing.T) {
	type entry struct {
		value  uint32
		nbBits uint8
	}
	tests := []struct {
		name  string
		bits  []entry
	}{
		{
			name: "canonical",
			bits: []entry{
				{value: 0x412345, nbBits: 23},
				{value: 0x41234567, nbBits: 31},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer := Buffer{}
			for _, write := range tt.bits {
				if err := buffer.WriteUInt32(write.value, write.nbBits); err != nil {
					t.Fatalf("Writing error: %s", err)
				}
			}
			str, err := buffer.ReadAllB64()
			if err != nil {
				t.Fatalf("Error in ReadAllB64: %s", err)
			}
			if err := buffer.WriteB64(str); err != nil {
				t.Fatalf("Error in WriteB64: %s", err)
			}
			for _, read := range tt.bits {
				result, err := buffer.ReadUInt32(read.nbBits)
				if err != nil {
					t.Fatalf("Read error: %s", err)
				}
				if result != read.value {
					t.Errorf("Got %x. Expect %x.", result, read.value)
				}
			}
		})
	}
}
