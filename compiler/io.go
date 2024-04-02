package compiler

import (
	"io"
	"os"
)

type ioProvider struct {
	DefaultOut io.WriteCloser
	DefaultErr io.WriteCloser
	DefaultIn  io.ReadCloser
	Closer     *closer
}

func DefaultIoProvider() *ioProvider {
	return &ioProvider{
		DefaultOut: WrapWriteCloser(os.Stdout),
		DefaultErr: WrapWriteCloser(os.Stderr),
		DefaultIn:  WrapReadFakeCloser(os.Stdin),
		Closer:     NewCloser(),
	}
}

func TestIoProvider() (*ioProvider, io.ReadCloser, io.ReadCloser, io.WriteCloser) {
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

func subshellIoProvider(parent *ioProvider) (*ioProvider, io.ReadCloser) {
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

func WrapWriteCloser(w io.Writer) io.WriteCloser {
	wc := &wrappedWriter{w}
	return wc
}

type wrappedWriter struct {
	w io.Writer
}

func (ww *wrappedWriter) Write(p []byte) (n int, err error) {
	return ww.w.Write(p)
}

func (ww *wrappedWriter) Close() error {
	return nil
}

func WrapReadFakeCloser(r io.Reader) io.ReadCloser {
	wr := &wrappedReader{r}
	return wr
}

type wrappedReader struct {
	r io.Reader
}

func (wr *wrappedReader) Read(p []byte) (n int, err error) {
	return wr.r.Read(p)
}

func (_ *wrappedReader) Close() error {
	return nil
}
