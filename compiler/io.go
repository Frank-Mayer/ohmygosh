package compiler

import (
	"io"
	"os"
)

type ioProvider struct {
	DefaultOut io.Writer
	DefaultErr io.Writer
	DefaultIn  io.Reader
	Closer     *closer
}

func DefaultIoProvider() *ioProvider {
	return &ioProvider{
		DefaultOut: os.Stdout,
		DefaultErr: os.Stderr,
		DefaultIn:  os.Stdin,
		Closer:     NewCloser(),
	}
}

func TestIoProvider() *ioProvider {
	return &ioProvider{
		DefaultOut: os.Stdout,
		DefaultErr: os.Stderr,
		DefaultIn:  os.Stdin,
		Closer:     NewCloser(),
	}
}

func subshellIoProvider(parent *ioProvider) (*ioProvider, io.Reader) {
	w, r := newPipe()
	return &ioProvider{
		DefaultOut: w,
		DefaultErr: parent.DefaultErr,
		DefaultIn:  parent.DefaultIn,
		Closer:     NewCloser(),
	}, r
}

func (i *ioProvider) Close() {
	i.Closer.Close()
}
