//go:build appengine || gopherjs
// +build appengine gopherjs

package logrus

import (
	"io"
)

func checkIfTerminal(w io.Writer) bool {
	return true
}