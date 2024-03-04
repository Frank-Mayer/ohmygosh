package compiler

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
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
		err = os.Chdir(c.Arguments[0])
	case "exit":
		if len(c.Arguments) > 0 {
			exitCode, err := strconv.Atoi(c.Arguments[0])
			if err != nil {
				return err
			}
			os.Exit(exitCode)
		} else {
			os.Exit(0)
		}
	case "echo":
		if c.Stdout != nil {
			_, err = (**c.Stdout).Write([]byte(strings.Join(c.Arguments, " ") + "\n"))
		}
	case "export":
		if len(c.Arguments) > 0 {
			for _, arg := range c.Arguments {
				parts := strings.SplitN(arg, "=", 2)
				if len(parts) >= 2 {
					os.Setenv(parts[0], strings.Join(parts[1:], "="))
				}
			}
		}
	case "unset":
		if len(c.Arguments) > 0 {
			for _, arg := range c.Arguments {
				os.Unsetenv(arg)
			}
		}
	case "whoami":
		if c.Stdout != nil {
			var u *user.User
			u, err = user.Current()
			if err != nil {
				break
			}
			_, err = (**c.Stdout).Write([]byte(u.Username + "\n"))
		}
	case "pwd":
		var pwd string
		pwd, err = os.Getwd()
		if err == nil {
			if c.Stdout != nil {
				_, err = (**c.Stdout).Write([]byte(pwd + "\n"))
			}
		}

	default:
		var exe string
		// look for the executable in the PATH
		exe, err = exec.LookPath(c.Executable)
		if err != nil {
			// get absolute path
			exe, err = filepath.Abs(c.Executable)
			if err != nil {
				exe = c.Executable
			}
		}
		cmd := &exec.Cmd{
			Path: exe,
			Args: append([]string{exe}, c.Arguments...),
		}
		cmd.Stdin = **c.Stdin
		cmd.Stdout = **c.Stdout
		cmd.Stderr = **c.Stderr
		err = cmd.Run()
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
