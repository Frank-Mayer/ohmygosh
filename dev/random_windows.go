//go:build windows

package dev

import (
	"crypto/rand"
	"io"
)

// /dev/random is a special file in Unix-like operating systems that serves as a blocking pseudorandom number generator.
func NewRandomReader() io.Reader {
	return rand.Reader
}
