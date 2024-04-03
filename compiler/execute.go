package compiler

import (
	"errors"
	"fmt"
	"strings"
	"sync"
)

func Execute(text string, iop *ioProvider) error {
	tokens, err := LexicalAnalysis(text, iop)
	if err != nil {
		return errors.Join(errors.New("failed to lexically analyze input"), err)
	}

	commands, err := Parse(text, tokens, iop)
	if err != nil {
		return errors.Join(errors.New("failed to parse input"), err)
	}

	wg := sync.WaitGroup{}

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
				return errors.Join(fmt.Errorf("failed to execute command %d: %q", i, command.String()), err)
			}
		}
	}

	wg.Wait()

	return nil
}

var builtinCommands map[string]func(*Command, *ioProvider) error

func init() {
	builtinCommands = map[string]func(*Command, *ioProvider) error{
		"cd":     execute_cd,
		"exit":   execute_exit,
		"echo":   execute_echo,
		"cat":    execute_cat,
		"export": execute_export,
		"unset":  execute_unset,
		"whoami": execute_whoami,
		"pwd":    execute_pwd,
		"which":  execute_which,
		"type":   execute_type,
		"sudo":   execute_sudo,
		"yes":    execute_yes,
		"true":   execute_true,
		"false":  execute_false,
		"sleep":  execute_sleep,
	}
}

func (c *Command) Execute(iop *ioProvider) error {
	var err error

	if fn, builtin := builtinCommands[strings.ToLower(c.Executable)]; builtin {
		err = fn(c, iop)
	} else {
		err = execute_default(c, iop)
	}

	if err != nil {
		// failed
		if c.Or != nil {
			return c.Or.Execute(iop)
		} else {
			return err
		}
	} else {
		// succeeded
		if c.And != nil {
			return c.And.Execute(iop)
		}
	}

	return nil
}
