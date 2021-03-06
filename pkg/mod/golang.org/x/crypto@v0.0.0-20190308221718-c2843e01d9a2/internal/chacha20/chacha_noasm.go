// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build (!arm64 && !s390x) || (arm64 && !go1.11) || gccgo || appengine
// +build !arm64,!s390x arm64,!go1.11 gccgo appengine

package chacha20

const (
	bufSize = 64
	haveAsm = false
)

func (*Cipher) xorKeyStreamAsm(dst, src []byte) {
	panic("not implemented")
}
