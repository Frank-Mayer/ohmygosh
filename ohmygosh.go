package ohmygosh

import (
	"sync"

	"github.com/tsukinoko-kun/ohmygosh/compiler"
	"github.com/tsukinoko-kun/ohmygosh/runtime"
)

func Execute(text string) (*sync.WaitGroup, error) {
	io := runtime.DefaultIoProvider()
	defer io.Close()
	return compiler.Execute(text, io)
}
