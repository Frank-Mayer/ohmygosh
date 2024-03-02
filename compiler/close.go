package compiler

import (
	"sync"
)

type (
	closableE interface {
		Close() error
	}
	closable interface {
		Close()
	}

	closer struct {
		mutex     sync.Mutex
		closable  []*closable
		closableE []*closableE
	}
)

func (c *closer) Close() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	wg := sync.WaitGroup{}
	wg.Add(len(c.closable) + len(c.closableE))
	for _, v := range c.closable {
		go func(v *closable) {
			(*v).Close()
			wg.Done()
		}(v)
	}
	for _, v := range c.closableE {
		go func(v *closableE) {
			_ = (*v).Close()
			wg.Done()
		}(v)
	}
	wg.Wait()
}

func (c *closer) Add(v closable) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.closable = append(c.closable, &v)
}

func (c *closer) AddE(v closableE) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.closableE = append(c.closableE, &v)
}

func NewCloser() *closer {
	return &closer{
		closable:  make([]*closable, 0),
		closableE: make([]*closableE, 0),
		mutex:     sync.Mutex{},
	}
}
