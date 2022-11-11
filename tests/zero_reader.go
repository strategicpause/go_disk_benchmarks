package tests

import (
	"io"
)

type ZeroReader struct {
	size int64
}

func NewZeroReader(sizeInBytes int64) *ZeroReader {
	return &ZeroReader{
		size: sizeInBytes,
	}
}

func (r *ZeroReader) Read(b []byte) (int, error) {
	if r.size <= 0 {
		return 0, io.EOF
	}

	l := len(b)
	r.size -= int64(l)

	return l, nil
}
