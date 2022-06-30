//go:build !go1.11
// +build !go1.11

package lsp

func init() {
	goVersion111 = false
}
