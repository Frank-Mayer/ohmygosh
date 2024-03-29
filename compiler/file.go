package compiler

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/Frank-Mayer/ohmygosh/dev"
)

func newFileWriter(c *closer, p string) (io.Writer, error) {
	p = filepath.Clean(p)
	switch p {
	case "/dev/null":
		return io.Discard, nil
	case "/dev/stdout":
		return os.Stdout, nil
	case "/dev/stderr":
		return os.Stderr, nil
	case "/dev/stdin":
		return os.Stdin, nil
	}
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
	p = filepath.Clean(p)
	switch p {
	case "/dev/null":
		return io.Discard, nil
	case "/dev/stdout":
		return os.Stdout, nil
	case "/dev/stderr":
		return os.Stderr, nil
	case "/dev/stdin":
		return os.Stdin, nil
	}
	f, err := os.OpenFile(p, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("could not open file %q", p), err)
	}
	c.AddE(f)
	return f, nil
}

func newFileReader(c *closer, p string) (io.Reader, error) {
	p = filepath.Clean(p)
	switch p {
	case "/dev/stdout":
		return os.Stdout, nil
	case "/dev/stderr":
		return os.Stderr, nil
	case "/dev/stdin":
		return os.Stdin, nil
	case "/dev/zero":
		return dev.NewZeroReader(), nil
	case "/dev/random":
		return dev.NewRandomReader(), nil
	case "/dev/urandom":
		return dev.NewUrandomReader(), nil
	}
	f, err := os.Open(p)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("could not open file %q", p), err)
	}
	c.AddE(f)
	return f, nil
}
