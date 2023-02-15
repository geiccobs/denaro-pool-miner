// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package sha256 implements SHA256 hash algorithm as defined
// in FIPS 180-4.
package sha256

import (
	"encoding/binary"
	"errors"
)

// The size of a SHA256 checksum in bytes.
const Size = 32

// The blocksize of SHA256 in bytes.
const BlockSize = 64

const (
	chunk = 64
	init0 = 0x6A09E667
	init1 = 0xBB67AE85
	init2 = 0x3C6EF372
	init3 = 0xA54FF53A
	init4 = 0x510E527F
	init5 = 0x9B05688C
	init6 = 0x1F83D9AB
	init7 = 0x5BE0CD19
)

// Digest represents the partial evaluation of a checksum.
type Digest struct {
	h   [8]uint32
	x   [chunk]byte
	nx  int
	len uint64
}

const (
	magic224      = "sha\x02"
	magic256      = "sha\x03"
	marshaledSize = len(magic256) + 8*4 + chunk + 8
)

func (d *Digest) MarshalBinary() ([]byte, error) {
	b := make([]byte, 0, marshaledSize)
	b = append(b, magic256...)

	b = binary.BigEndian.AppendUint32(b, d.h[0])
	b = binary.BigEndian.AppendUint32(b, d.h[1])
	b = binary.BigEndian.AppendUint32(b, d.h[2])
	b = binary.BigEndian.AppendUint32(b, d.h[3])
	b = binary.BigEndian.AppendUint32(b, d.h[4])
	b = binary.BigEndian.AppendUint32(b, d.h[5])
	b = binary.BigEndian.AppendUint32(b, d.h[6])
	b = binary.BigEndian.AppendUint32(b, d.h[7])
	b = append(b, d.x[:d.nx]...)
	b = b[:len(b)+len(d.x)-d.nx] // already zero
	b = binary.BigEndian.AppendUint64(b, d.len)
	return b, nil
}

func (d *Digest) UnmarshalBinary(b []byte) error {
	if len(b) < len(magic224) || string(b[:len(magic256)]) != magic256 {
		return errors.New("crypto/sha256: invalid hash state identifier")
	}
	if len(b) != marshaledSize {
		return errors.New("crypto/sha256: invalid hash state size")
	}
	b = b[len(magic224):]
	b, d.h[0] = consumeUint32(b)
	b, d.h[1] = consumeUint32(b)
	b, d.h[2] = consumeUint32(b)
	b, d.h[3] = consumeUint32(b)
	b, d.h[4] = consumeUint32(b)
	b, d.h[5] = consumeUint32(b)
	b, d.h[6] = consumeUint32(b)
	b, d.h[7] = consumeUint32(b)
	b = b[copy(d.x[:], b):]
	b, d.len = consumeUint64(b)
	d.nx = int(d.len % chunk)
	return nil
}

func consumeUint64(b []byte) ([]byte, uint64) {
	_ = b[7]
	x := uint64(b[7]) | uint64(b[6])<<8 | uint64(b[5])<<16 | uint64(b[4])<<24 |
		uint64(b[3])<<32 | uint64(b[2])<<40 | uint64(b[1])<<48 | uint64(b[0])<<56
	return b[8:], x
}

func consumeUint32(b []byte) ([]byte, uint32) {
	_ = b[3]
	x := uint32(b[3]) | uint32(b[2])<<8 | uint32(b[1])<<16 | uint32(b[0])<<24
	return b[4:], x
}

func (d *Digest) Reset() {
	d.h[0] = init0
	d.h[1] = init1
	d.h[2] = init2
	d.h[3] = init3
	d.h[4] = init4
	d.h[5] = init5
	d.h[6] = init6
	d.h[7] = init7
	d.nx = 0
	d.len = 0
}

// New returns a new hash.Hash computing the SHA256 checksum. The Hash
// also implements encoding.BinaryMarshaler and
// encoding.BinaryUnmarshaler to marshal and unmarshal the internal
// state of the hash.
func New() Digest {
	d := new(Digest)
	d.Reset()
	return *d
}

func (d *Digest) Size() int { return Size }

func (d *Digest) BlockSize() int { return BlockSize }

func (d *Digest) Write(p []byte) (nn int, err error) {
	nn = len(p)
	d.len += uint64(nn)
	if d.nx > 0 {
		n := copy(d.x[d.nx:], p)
		d.nx += n
		if d.nx == chunk {
			block(d, d.x[:])
			d.nx = 0
		}
		p = p[n:]
	}
	if len(p) >= chunk {
		n := len(p) &^ (chunk - 1)
		block(d, p[:n])
		p = p[n:]
	}
	if len(p) > 0 {
		d.nx = copy(d.x[:], p)
	}
	return
}

func (d *Digest) Sum(in []byte) []byte {
	// Make a copy of d so that caller can keep writing and summing.
	hash := (*d).checkSum()
	return append(in, hash[:]...)
}

func (d *Digest) checkSum() [Size]byte {
	length := d.len
	// Padding. Add a 1 bit and 0 bits until 56 bytes mod 64.
	var tmp [64 + 8]byte // padding + length buffer
	tmp[0] = 0x80
	var t uint64
	if length%64 < 56 {
		t = 56 - length%64
	} else {
		t = 64 + 56 - length%64
	}

	// Length in bits.
	length <<= 3
	padlen := tmp[:t+8]
	binary.BigEndian.PutUint64(padlen[t+0:], length)
	if _, err := d.Write(padlen); err != nil {
		panic(err)
	}

	if d.nx != 0 {
		panic("d.nx != 0")
	}

	var digest [Size]byte

	binary.BigEndian.PutUint32(digest[0:], d.h[0])
	binary.BigEndian.PutUint32(digest[4:], d.h[1])
	binary.BigEndian.PutUint32(digest[8:], d.h[2])
	binary.BigEndian.PutUint32(digest[12:], d.h[3])
	binary.BigEndian.PutUint32(digest[16:], d.h[4])
	binary.BigEndian.PutUint32(digest[20:], d.h[5])
	binary.BigEndian.PutUint32(digest[24:], d.h[6])

	binary.BigEndian.PutUint32(digest[28:], d.h[7])

	return digest
}

// Sum256 returns the SHA256 checksum of the data.
func Sum256(data []byte) [Size]byte {
	var d Digest
	d.Reset()
	if _, err := d.Write(data); err != nil {
		panic(err)
	}
	return d.checkSum()
}
