package ohmygosh

import (
	"github.com/Frank-Mayer/ohmygosh/compiler"
)

func Execute(text string) error {
	io := compiler.DefaultIoProvider()
	defer io.Close()
	return compiler.Execute(text, io)
}
