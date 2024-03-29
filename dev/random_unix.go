//go:build !windows

package dev

import (
	"io"
	"os"
)

// /dev/random is a special file in Unix-like operating systems that serves as a blocking pseudorandom number generator.
func NewRandomReader() io.Reader {
	f, err := os.Open("/dev/random")
	if err != nil {
		panic(err)
	}
	return f
}
