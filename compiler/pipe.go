package compiler

import (
	"io"
	"sync"
)

type pipe struct {
	buffer []byte
	mutex  sync.Mutex
}

func newPipe() (io.Writer, io.Reader) {
	p := &pipe{buffer: make([]byte, 0)}
	return p, p
}

func (p *pipe) Write(b []byte) (n int, err error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.buffer = append(p.buffer, b...)
	return len(b), nil
}

func (p *pipe) Read(b []byte) (n int, err error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if len(p.buffer) == 0 {
		return 0, io.EOF
	}
	n = copy(b, p.buffer)
	p.buffer = p.buffer[n:]
	return n, nil
}
