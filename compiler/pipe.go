package compiler

import (
	"io"
	"sync/atomic"
)

type pipe struct {
	buffer chan []byte
	open   atomic.Bool
}

func NewPipe() (io.WriteCloser, io.ReadCloser) {
	p := &pipe{
		buffer: make(chan []byte, 1),
		open:   atomic.Bool{},
	}
	p.open.Store(true)
	return p, p
}

func (p *pipe) Write(b []byte) (int, error) {
	if !p.open.Load() {
		return 0, io.ErrClosedPipe
	}
	p.buffer <- b
	return len(b), nil
}

func (p *pipe) Read(b []byte) (int, error) {
	if !p.open.Load() {
		return 0, io.EOF
	}
	data, ok := <-p.buffer
	if !ok {
		return 0, io.EOF
	}
	n := copy(b, data)
	return n, nil
}

func (p *pipe) Close() error {
	if !p.open.Swap(false) {
		return nil
	}
	close(p.buffer)
	return nil
}
