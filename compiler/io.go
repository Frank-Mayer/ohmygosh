package compiler

import (
	"fmt"
	"io"
	"os"
)

type ioProvider struct {
	DefaultOut io.WriteCloser
	DefaultErr io.WriteCloser
	DefaultIn  io.Reader
	Closer     *closer
}

func DefaultIoProvider() *ioProvider {
	return &ioProvider{
		DefaultOut: WrapWriteFakeCloser(os.Stdout),
		DefaultErr: WrapWriteFakeCloser(os.Stderr),
		DefaultIn:  os.Stdin,
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

func WrapWriteFakeCloser(w io.Writer) io.WriteCloser {
	wc := &wrappedWriterFakeCloser{w}
	return wc
}

type wrappedWriterFakeCloser struct {
	w io.Writer
}

func (ww *wrappedWriterFakeCloser) Write(p []byte) (n int, err error) {
	return ww.w.Write(p)
}

func (ww *wrappedWriterFakeCloser) Close() error {
	return nil
}

func WrapReadFakeCloser(r io.Reader) io.ReadCloser {
	wr := &wrappedReaderFakeCloser{r}
	return wr
}

type wrappedReaderFakeCloser struct {
	r io.Reader
}

func (wr *wrappedReaderFakeCloser) Read(p []byte) (n int, err error) {
	return wr.r.Read(p)
}

func (wr *wrappedReaderFakeCloser) Close() error {
	fmt.Println("wrappedReader.Close")
	return nil
}
