package httpfunc

import (
	"errors"
	"io"
)

var ErrMaxLimit = errors.New("max limit")

var _ io.Reader = (*MaxBytesReader)(nil)

// NewMaxBytesReader returns a new MaxBytesReader with the given reader and limit.
func NewMaxBytesReader(r io.Reader, n int64) *MaxBytesReader {
	return &MaxBytesReader{R: r, N: n}
}

type MaxBytesReader struct {
	R io.Reader
	N int64
}

func (l *MaxBytesReader) Read(p []byte) (n int, err error) {
	if l.N <= 0 {
		return 0, ErrMaxLimit
	}

	if int64(len(p)) > l.N {
		p = p[0:l.N]
	}

	n, err = l.R.Read(p)
	l.N -= int64(n)
	return
}
