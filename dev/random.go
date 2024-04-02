package dev

import (
	"io"
)

// /dev/random is a special file in Unix-like operating systems that serves as a blocking pseudorandom number generator.
func NewRandomReader() io.ReadWriteCloser {
	return &urandomReader{}
}
