package iohelper

import (
	"io"
)

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
