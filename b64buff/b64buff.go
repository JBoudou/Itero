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

// Package b64buff implements a bit buffer that can be represented as URL string.
//
// The bit buffer allows to read and write any number of bits.
// The encoding is similar to Base64 but different. It has no padding. Currently it cannot be
// changed.
package b64buff

import (
	"bytes"
	"crypto/rand"
	"errors"
	"io"
)

// Buffer is a bit buffer.
type Buffer struct {
	buff bytes.Buffer

	// The readSize lower bits of readMore contains bits to read.
	readMore byte
	readSize uint8

	// The writeSize upper bits (6 based) of writeMore contains bits to write.
	writeMore byte
	writeSize uint8
}

const encoding = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ-_"

var decoding [256]byte

var (
	WrongNbBits   = errors.New("Impossible number of bits")
	NotEnoughData = errors.New("No enough data to read")
	NotAligned    = errors.New("Buffer not aligned")
	EncodingError = errors.New("Error in encoding")
)

func init() {
	for i := 0; i < len(decoding); i++ {
		decoding[i] = 0xFF
	}
	for i := 0; i < len(encoding); i++ {
		decoding[encoding[i]] = byte(i)
	}
}

// Random create a Buffer containing at least nbBits random bits.
func NewRandom(nbBits uint32) (ret *Buffer, err error) {
	ret = &Buffer{}

	nbBytes := (nbBits + 5) / 6
	buff := make([]byte, nbBytes)
	if _, err := rand.Reader.Read(buff); err != nil {
		return ret, err
	}

	for i := range buff {
		buff[i] &= 0x3F
	}

	ret.buff.Write(buff)
	return
}

// Len returns the number of bits in the buffer.
func (self *Buffer) Len() uint32 {
	return uint32(self.readSize) + uint32(self.writeSize) + (uint32(self.buff.Len()) * 6)
}

// WriteUInt32 writes the lower nbBits of data to the buffer.
func (self *Buffer) WriteUInt32(data uint32, nbBits uint8) error {
	if nbBits == 0 {
		return nil
	}
	if nbBits > 32 {
		return WrongNbBits
	}

	if self.writeSize+nbBits < 6 {
		mask := uint32(0x3F) >> (6 - nbBits)
		self.writeSize += nbBits
		self.writeMore |= byte(data&mask) << (6 - self.writeSize)
		return nil
	}

	if self.writeSize > 0 {
		diff := 6 - self.writeSize
		nbBits -= diff
		mask := (uint32(0x3F) >> diff) << nbBits
		self.writeMore |= byte((data & mask) >> nbBits)
		if err := self.buff.WriteByte(self.writeMore); err != nil {
			return err
		}
		self.writeMore = 0
		self.writeSize = 0
	}

	for nbBits >= 6 {
		nbBits -= 6
		mask := uint32(0x3F) << nbBits
		val := byte((data & mask) >> nbBits)
		if err := self.buff.WriteByte(val); err != nil {
			return err
		}
	}

	if nbBits > 0 {
		diff := 6 - nbBits
		mask := uint32(0x3F) >> diff
		self.writeMore = byte(data&mask) << diff
		self.writeSize = nbBits
	}
	return nil
}

// ReadUInt32 reads nbBits of the buffer into the lower bits of ret.
func (self *Buffer) ReadUInt32(nbBits uint8) (ret uint32, err error) {
	if nbBits == 0 {
		return
	}
	if nbBits > 32 {
		return 0, WrongNbBits
	}
	if uint32(nbBits) > self.Len() {
		return 0, NotEnoughData
	}

	if nbBits <= self.readSize {
		self.readSize -= nbBits
		mask := (byte(0x3F) >> (6 - nbBits)) << self.readSize
		ret = uint32(self.readMore&mask) >> self.readSize
		self.readMore &^= mask
		return
	}

	if self.readSize > 0 {
		nbBits -= self.readSize
		ret |= uint32(self.readMore) << nbBits
		self.readMore = 0
		self.readSize = 0
	}

	for nbBits >= 6 {
		nbBits -= 6
		var val uint8
		val, err = self.buff.ReadByte()
		if err != nil {
			return
		}
		ret |= uint32(val) << nbBits
	}

	if nbBits > 0 {
		if self.buff.Len() > 0 {
			self.readMore, err = self.buff.ReadByte()
			if err != nil {
				return
			}
			self.readSize = 6 - nbBits
			mask := (uint8(0x3F) >> self.readSize) << self.readSize
			ret |= uint32(self.readMore&mask) >> self.readSize
			self.readMore &^= mask
		} else {
			diff := 6 - nbBits
			mask := uint8(0x3F) << diff
			ret |= uint32(self.writeMore&mask) >> diff
			self.writeSize -= nbBits
			self.writeMore = (self.writeMore &^ mask) << nbBits
		}
	}
	return
}

// WriteB64 writes the whole string into the buffer.
//
// The string must have been correctly encoded by a call to ReadAllB64.
//
// The call fails if the write is not aligned. The write is aligned if the sum of written bits
// into the buffer can be divided by 6.
func (self *Buffer) WriteB64(str string) (err error) {
	if self.writeSize != 0 {
		return NotAligned
	}

	for _, c := range str {
		val := decoding[c]
		if val > 0x3F {
			return EncodingError
		}
		err = self.buff.WriteByte(val)
		if err != nil {
			return
		}
	}
	return
}

// ReadAllB64 reads the whole buffer to the string.
//
// The call fails if the read is not aligned. The read is aligned if the sum of read bits
// from the buffer can be divided by 6.
func (self *Buffer) ReadAllB64() (ret string, err error) {
	if self.readSize != 0 {
		return "", NotAligned
	}

	from := self.buff.Bytes()
	to := make([]byte, len(from))
	for i, val := range from {
		to[i] = encoding[val]
	}
	self.buff.Reset()
	return string(to), nil
}

type b64Reader struct {
	buffer *Buffer
}

func (self b64Reader) Read(p []byte) (n int, err error) {
	if self.buffer.readSize != 0 {
		return 0, NotAligned
	}

	n, err = self.buffer.buff.Read(p)
	for i := 0; i < n; i++ {
		p[i] = encoding[p[i]]
	}
	return
}

// B64Reader return a encoded reader on the Buffer.
//
// Calling Read on the returned value fail if the read is not aligned.
// The read is aligned if the sum of read bits from the buffer can be divided by 6.
func (self *Buffer) B64Reader() io.Reader {
	return b64Reader{self}
}

// AlignRead forces the read to be aligned, by discarding surnumerous bits.
//
// The discarded bits are the less significant bits that would have been returned by a call
// to ReadUInt32. Those bits are returned by the method.
func (self *Buffer) AlignRead() (ret byte) {
	ret = self.readMore
	self.readSize = 0
	self.readMore = 0
	return
}

// RandomString produces a random readable string of the given length.
// The returned string is a valid B64 encoded random value.
func RandomString(length uint32) (ret string, err error) {
	var rnd *Buffer
	if rnd, err = NewRandom(length * 6); err != nil {
		return "", err
	}
	bits := rnd.B64Reader()

	buff := make([]byte, length)
	if _, err := bits.Read(buff); err != nil {
		return "", err
	}
	return string(buff), nil
}
