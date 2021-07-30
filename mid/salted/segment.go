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

// Package salted handles salted identifiers.
//
// A salted identifier is a pair consisting of an identifier and a random salt.
// Salted identifiers are used to prevent malicious agents to guess valid identifiers too easily.
package salted

import (
	"net/http"

	"github.com/JBoudou/Itero/mid/server"
	"github.com/JBoudou/Itero/pkg/b64buff"
)

// SaltLength is the number of bits of the salt that are really used. The more significant bits are
// ignored.
const SaltLength = 22

// Segment represents a segment encoding an id and a salt.
// This class is used to encode and decode such segments.
// The encoded form is a short string that can be used as an URI segment.
type Segment struct {
	Id   uint32
	Salt uint32
}

// New creates a Segment with the given id and a random salt.
func New(id uint32) (ret Segment, err error) {
	ret.Id = id
	ret.Salt, err = b64buff.RandomUInt32(SaltLength)
	return
}

// FromRequest creates a Segment from the last segment of the URL of a request.
func FromRequest(request *server.Request) (segment Segment, err error) {
	remainingLength := len(request.RemainingPath)
	if remainingLength == 0 {
		err = server.NewHttpError(http.StatusBadRequest, "No salted segment", "No salted segment")
		return
	}
	return Decode(request.RemainingPath[remainingLength-1])
}

// Decode creates a Segment from its URI representation.
func Decode(str string) (ret Segment, err error) {
	buff := b64buff.Buffer{}
	err = buff.WriteB64(str)
	if err == nil {
		ret.Salt, err = buff.ReadUInt32(SaltLength)
	}
	if err == nil {
		ret.Id, err = buff.ReadUInt32(32)
	}
	return
}

// Encode returns the URI representation of the Segment.
func (self Segment) Encode() (str string, err error) {
	buff := b64buff.Buffer{}
	err = buff.WriteUInt32(self.Salt, SaltLength)
	if err == nil {
		err = buff.WriteUInt32(self.Id, 32)
	}
	if err == nil {
		str, err = buff.ReadAllB64()
	}
	return
}
