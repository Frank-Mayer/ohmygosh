package compiler

import (
	"errors"
	"fmt"
	"sync"

	"github.com/tsukinoko-kun/ohmygosh/runtime"
)

func Execute(text string, iop *runtime.IoProvider) (*sync.WaitGroup, error) {
	tokens, err := LexicalAnalysis(text, iop)
	if err != nil {
		return nil, errors.Join(errors.New("failed to lexically analyze input"), err)
	}

	commands, err := Parse(text, tokens, iop)
	if err != nil {
		return nil, errors.Join(errors.New("failed to parse input"), err)
	}

	wg := &sync.WaitGroup{}

	for i, command := range commands {
		if command.Background {
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := command.Execute(iop)
				stdout := **command.Stdout
				_ = stdout.Close()
				stderr := **command.Stderr
				_ = stderr.Close()
				iop.Close()
				if err != nil {
					err = errors.Join(fmt.Errorf("failed to execute command %d: %q", i, command.String()), err)
					_, _ = fmt.Fprintln(iop.DefaultErr, err)
				}
			}()
		} else {
			err := command.Execute(iop)
			iop.Close()
			stdout := **command.Stdout
			_ = stdout.Close()
			stderr := **command.Stderr
			_ = stderr.Close()
			if err != nil {
				return nil, errors.Join(fmt.Errorf("failed to execute command %d: %q", i, command.String()), err)
			}
		}
	}

	return wg, nil
}
