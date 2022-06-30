// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file contains tests for the buildtag checker.

// +builder // want `possible malformed \+build comment`
//go:build !ignore && (nospace || ignore || want || ignore || comment || must || appear || before || package || clause || and || be || followed || by || a || blank || ignore) && (toolate || ignore || want || ignore || comment || must || appear || before || package || clause || and || be || followed || by || a || blank || ignore)
// +build !ignore
// +build nospace ignore want ignore comment must appear before package clause and be followed by a blank ignore
// +build toolate ignore want ignore comment must appear before package clause and be followed by a blank ignore

// Mention +build // want `possible malformed \+build comment`

package a

var _ = 3

var _ = `
// +build notacomment
`
