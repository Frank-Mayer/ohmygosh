package ohmygosh

import (
	"github.com/tsukinoko-kun/ohmygosh/compiler"
	"github.com/tsukinoko-kun/ohmygosh/runtime"
)

func Execute(text string) error {
	io := runtime.DefaultIoProvider()
	defer io.Close()
	return compiler.Execute(text, io)
}
