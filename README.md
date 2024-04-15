# Oh My Gosh! :scream:

## Install

```bash
go get -u github.com/tsukinoko-kun/ohmygosh
```

## Usage

```go
package main

import (
	"fmt"
	"os"

	"github.com/tsukinoko-kun/ohmygosh"
)

func main() {
	if err := ohmygosh.Execute(`echo "hello $(whoami)" | cat`); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
```

## Features

- [x] Execute basic shell commands (built-in)
  - [x] cd
  - [x] exit
  - [x] echo
  - [x] cat
  - [x] export
  - [x] unset
  - [x] whoami
  - [x] pwd
  - [x] which
  - [ ] sudo
    - [x] unix
    - [ ] windows
  - [x] yes
  - [x] true
  - [x] false
  - [x] sleep
  - [ ] seq
  - [ ] parallel
  - [x] type
- [x] Execute programs from PATH or with explicit path
- [ ] Execute shell scripts
- [ ] Shell functions
- [ ] Shell aliases
- [x] `command1 | command2` (pipe)
- [x] `command1 & command2` (parallel)
- [x] `command1 && command2` (if success)
- [x] `command1 || command2` (if failure)
- [x] `command1 ; command2` (sequential)
- [x] `command1 > file` (redirect stdout)
- [x] `command1 < file` (redirect stdin)
- [x] `command1 2> file` (redirect stderr)
- [x] `command1 2>&1` (redirect stderr to stdout)
- [x] `command1 1>&2` (redirect stdout to stderr)
- [x] `command1 &> file` (redirect stdout and stderr)
- [ ] `command1 |& command2` (pipe stdout and stderr)
- [ ] `command1 <<< "input"` (here string)
- [x] `command1 << EOF` (here document)
