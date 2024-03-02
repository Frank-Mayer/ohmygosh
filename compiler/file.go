package compiler

import (
	"errors"
	"fmt"
	"io"
	"os"
)

func newFileWriter(c *closer, p string) (io.Writer, error) {
	f, err := os.OpenFile(p, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("could not open file %q", p), err)
	}
	c.AddE(f)
	if err := f.Truncate(0); err != nil {
		return nil, errors.Join(fmt.Errorf("could not truncate file %q", p), err)
	}
	if _, err := f.Seek(0, 0); err != nil {
		return nil, errors.Join(fmt.Errorf("could not seek file %q", p), err)
	}
	return f, nil
}

func newFileAppendWriter(c *closer, p string) (io.Writer, error) {
	f, err := os.OpenFile(p, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("could not open file %q", p), err)
	}
	c.AddE(f)
	return f, nil
}

func newFileReader(c *closer, p string) (io.Reader, error) {
	f, err := os.Open(p)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("could not open file %q", p), err)
	}
	c.AddE(f)
	return f, nil
}
