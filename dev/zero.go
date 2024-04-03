package dev

import (
	"io"
)

type zeroReader struct{}

// /dev/zero is a special file in Unix-like operating systems that provides as many null characters (ASCII NUL, 0x00) as are read from it.
func NewZeroReader() io.ReadWriteCloser {
	return &zeroReader{}
}

func (z *zeroReader) Read(p []byte) (n int, err error) {
	for i := range p {
		p[i] = byte(0)
	}
	return len(p), nil
}

func (_ *zeroReader) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (_ *zeroReader) Close() error {
	return nil
}
