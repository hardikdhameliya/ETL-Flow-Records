// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !amd64 || gccgo || appengine
// +build !amd64 gccgo appengine

package poly1305

type mac struct{ macGeneric }

func newMAC(key *[32]byte) mac { return mac{newMACGeneric(key)} }
