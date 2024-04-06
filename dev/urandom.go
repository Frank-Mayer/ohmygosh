package dev

import (
	"crypto/rand"
	"io"
)

type urandomReader struct{}

// /dev/random is a special file in Unix-like operating systems that serves as a blocking pseudorandom number generator.
func NewUrandomReader() io.ReadWriteCloser {
	return &urandomReader{}
}

func (_ *urandomReader) Read(p []byte) (n int, err error) {
	return rand.Read(p)
}

func (_ *urandomReader) Write(p []byte) (n int, err error) {
	return io.Discard.Write(p)
}

func (_ *urandomReader) Close() error {
	return nil
}
