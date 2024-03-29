package compiler

import (
	"errors"
	"fmt"
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

	for i, command := range commands {
		if command.Background {
			go func() {
				err := command.Execute()
				iop.Close()
				if err != nil {
					err = errors.Join(fmt.Errorf("failed to execute command %d: %q", i, command.String()), err)
					fmt.Fprintln(iop.DefaultErr, err)
				}
			}()
		} else {
			err := command.Execute()
			iop.Close()
			if err != nil {
				return errors.Join(fmt.Errorf("failed to execute command %d: %q", i, command.String()), err)
			}
		}
	}

	return nil
}

func (c *Command) Execute() error {
	var err error

	switch c.Executable {
	case "cd":
		err = c.execute_cd()
	case "exit":
		err = c.execute_exit()
	case "echo":
		err = c.execute_echo()
	case "export":
		err = c.execute_export()
	case "unset":
		err = c.execute_unset()
	case "whoami":
		err = c.execute_whoami()
	case "pwd":
		err = c.execute_pwd()
	case "sudo":
		err = c.execute_sudo()
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
