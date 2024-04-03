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
				err := command.Execute()
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
			err := command.Execute()
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

var builtinCommands map[string]func(*Command) error

func init() {
	builtinCommands = map[string]func(*Command) error{
		"cd":     (*Command).execute_cd,
		"exit":   (*Command).execute_exit,
		"echo":   (*Command).execute_echo,
		"cat":    (*Command).execute_cat,
		"export": (*Command).execute_export,
		"unset":  (*Command).execute_unset,
		"whoami": (*Command).execute_whoami,
		"pwd":    (*Command).execute_pwd,
		"which":  (*Command).execute_which,
		"type":   (*Command).execute_type,
		"sudo":   (*Command).execute_sudo,
		"yes":    (*Command).execute_yes,
		"true":   (*Command).execute_true,
		"false":  (*Command).execute_false,
		"sleep":  (*Command).execute_sleep,
	}
}

func (c *Command) Execute() error {
	var err error

	if fn, builtin := builtinCommands[strings.ToLower(c.Executable)]; builtin {
		err = fn(c)
	} else {
		err = c.execute_default()
	}

	if err != nil {
		// failed
		if c.Or != nil {
			return c.Or.Execute()
		} else {
			return err
		}
	} else {
		// succeeded
		if c.And != nil {
			return c.And.Execute()
		}
	}

	return nil
}
