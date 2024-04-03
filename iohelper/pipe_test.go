package iohelper_test

import (
	"fmt"
	"testing"

	"github.com/Frank-Mayer/ohmygosh/iohelper"
)

func TestPipe(t *testing.T) {
	t.Run("simle write and read", func(t *testing.T) {
		t.Parallel()

		w, r := iohelper.NewPipe()
		defer w.Close()
		defer r.Close()

		// write
		n, err := w.Write([]byte("hello"))
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if n != 5 {
			t.Errorf("unexpected number of bytes written: %v", n)
		}

		// read
		buf := make([]byte, 5)
		n, err = r.Read(buf)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if n != 5 {
			t.Errorf("unexpected number of bytes read: %v", n)
		}
		if string(buf) != "hello" {
			t.Errorf("unexpected data read: %v", string(buf))
		}
	})

	t.Run("io.copy", func(t *testing.T) {
		t.Parallel()

		w, r := iohelper.NewPipe()
		defer w.Close()
		defer r.Close()

		go func() {
			if _, err := w.Write([]byte("hello")); err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			w.Close()
		}()

		buf := make([]byte, 5)
		n, err := r.Read(buf)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if n != 5 {
			t.Errorf("unexpected number of bytes read: %v", n)
		}
		if string(buf) != "hello" {
			t.Errorf("unexpected data read: %v", string(buf))
		}
	})

	t.Run("fmt", func(t *testing.T) {
		t.Parallel()

		w, r := iohelper.NewPipe()
		defer w.Close()
		defer r.Close()

		go func() {
			_, _ = fmt.Fprint(w, "hello")
			w.Close()
		}()

		buf := make([]byte, 5)
		n, err := r.Read(buf)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if n != 5 {
			t.Errorf("unexpected number of bytes read: %v", n)
		}
		if string(buf) != "hello" {
			t.Errorf("unexpected data read: %v", string(buf))
		}
	})
}
