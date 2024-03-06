package main

import (
	"bufio"
	"os"

	"github.com/Frank-Mayer/ohmygosh"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	for {
		if wd, err := os.Getwd(); err == nil {
			print(wd + " ")
		}
		print("$ ")
		text, _ := reader.ReadString('\n')
		_ = ohmygosh.Execute(text)
	}
}
