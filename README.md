# Oh My Gosh! :scream:

## Install

```bash
go get -u github.com/Frank-Mayer/ohmygosh
```

## Usage

```go
package main

import (
	"fmt"
	"os"

	"github.com/Frank-Mayer/ohmygosh"
)

func main() {
	if err := ohmygosh.Execute("echo $PWD"); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
```
