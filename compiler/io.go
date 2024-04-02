package compiler

import (
	"io"
	"os"
	"strings"
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

func TestIoProvider(stdin string) (*ioProvider, *strings.Builder, *strings.Builder) {
	outSB := &strings.Builder{}
	outW := WrapWriteFakeCloser(outSB)
	errSB := &strings.Builder{}
	errW := WrapWriteFakeCloser(errSB)
	var inR io.Reader
	if stdin == "" {
		inR = os.Stdin
	} else {
		inR = strings.NewReader(stdin)
	}
	return &ioProvider{
		DefaultOut: outW,
		DefaultErr: errW,
		DefaultIn:  inR,
		Closer:     NewCloser(),
	}, outSB, errSB
}

func subshellIoProvider(parent *ioProvider) (*ioProvider, *strings.Builder) {
	sb := &strings.Builder{}
	w := WrapWriteFakeCloser(sb)
	return &ioProvider{
		DefaultOut: w,
		DefaultErr: parent.DefaultErr,
		DefaultIn:  parent.DefaultIn,
		Closer:     NewCloser(),
	}, sb
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
	return nil
}
