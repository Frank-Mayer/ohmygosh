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

func (c *Command) Execute() error {
	var err error

	switch strings.ToLower(c.Executable) {
	case "cd":
		err = c.execute_cd()
	case "exit":
		err = c.execute_exit()
	case "echo":
		err = c.execute_echo()
	case "cat":
		err = c.execute_cat()
	case "export":
		err = c.execute_export()
	case "unset":
		err = c.execute_unset()
	case "whoami":
		err = c.execute_whoami()
	case "pwd":
		err = c.execute_pwd()
	case "which":
		err = c.execute_which()
	case "sudo":
		err = c.execute_sudo()
	case "yes":
		err = c.execute_yes()
	case "true":
		err = c.execute_true()
	case "false":
		err = c.execute_false()
	default:
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
