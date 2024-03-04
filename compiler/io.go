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

func TestIoProvider() (*ioProvider, io.Reader, io.Reader, io.Writer) {
	outW, outR := newPipe()
	errW, errR := newPipe()
	inW, inR := newPipe()
	return &ioProvider{
		DefaultOut: outW,
		DefaultErr: errW,
		DefaultIn:  inR,
		Closer:     NewCloser(),
	}, outR, errR, inW
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
