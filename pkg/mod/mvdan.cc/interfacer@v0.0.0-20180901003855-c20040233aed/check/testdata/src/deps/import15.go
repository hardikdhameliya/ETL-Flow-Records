//go:build !go1.6
// +build !go1.6

package single

type Closer interface {
	Close()
}

type ReadCloser interface {
	Closer
	Read()
}
