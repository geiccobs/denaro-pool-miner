// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sha256

var k = _K

//go:noescape
func sha256block(h []uint32, p []byte, k []uint32)

func block(dig *Digest, p []byte) {
	h := dig.h[:]
	sha256block(h, p, k)
}
