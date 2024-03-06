package compiler

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"strconv"
	"strings"
)

func (c *Command) execute_cd() error {
	switch len(c.Arguments) {
	case 0:
		home, err := os.UserHomeDir()
		if err != nil {
			return errors.Join(errors.New("cd: failed to get home directory"), err)
		}
		return os.Chdir(home)
	case 1:
		return os.Chdir(c.Arguments[0])
	default:
		fmt.Fprintln(**c.Stderr, "cd: too many arguments")
		return errors.New("cd: too many arguments")
	}
}

func (c *Command) execute_exit() error {
	switch len(c.Arguments) {
	case 0:
		os.Exit(0)
	case 1:
		code, err := strconv.Atoi(c.Arguments[0])
		if err != nil {
			fmt.Fprintln(**c.Stderr, "exit: ", err)
			return errors.Join(fmt.Errorf("exit: failed to parse argument %q as an integer", c.Arguments[0]), err)
		}
		os.Exit(code)
	default:
		fmt.Fprintln(**c.Stderr, "exit: too many arguments")
		return errors.New("exit: too many arguments")
	}
	return nil
}

func (c *Command) execute_echo() error {
	fmt.Fprintln(**c.Stdout, c.Arguments)
	return nil
}

func (c *Command) execute_export() error {
	if len(c.Arguments) > 0 {
		for _, arg := range c.Arguments {
			pair := os.ExpandEnv(arg)
			if pair != "" {
				kv := strings.Split(pair, "=")
				if len(kv) >= 2 {
					os.Setenv(kv[0], strings.Join(kv[1:], "="))
				} else {
					// does this variable exist?
					if _, ok := os.LookupEnv(kv[0]); !ok {
						// if not, set it to an empty string
						os.Setenv(kv[0], "")
					}
				}
			}
		}
	} else {
		for _, env := range os.Environ() {
			fmt.Fprintf(**c.Stdout, "declare -x %s\n", env)
		}
	}
	return nil
}

func (c *Command) execute_unset() error {
	if len(c.Arguments) > 0 {
		for _, arg := range c.Arguments {
			os.Unsetenv(arg)
		}
	}
	return nil
}

func (c *Command) execute_whoami() error {
	if u, err := user.Current(); err == nil {
		fmt.Fprintln(**c.Stdout, u.Username)
	} else {
		fmt.Fprintln(**c.Stderr, "whoami: ", err)
		return errors.Join(errors.New("whoami: failed to get current user"), err)
	}
	return nil
}

func (c *Command) execute_pwd() error {
	if wd, err := os.Getwd(); err == nil {
		fmt.Fprintln(**c.Stdout, wd)
	} else {
		fmt.Fprintln(**c.Stderr, "pwd: ", err)
		return errors.Join(errors.New("pwd: failed to get working directory"), err)
	}
	return nil
}
