package compiler

import (
	"fmt"
	"io"
)

func newPipe() (io.WriteCloser, io.ReadCloser) {
	fmt.Println("newPipe")
	r, w := io.Pipe()
	return w, r
}
