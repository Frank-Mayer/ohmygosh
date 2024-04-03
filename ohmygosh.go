package ohmygosh

import (
	"github.com/Frank-Mayer/ohmygosh/compiler"
	"github.com/Frank-Mayer/ohmygosh/runtime"
)

func Execute(text string) error {
	io := runtime.DefaultIoProvider()
	defer io.Close()
	return compiler.Execute(text, io)
}
