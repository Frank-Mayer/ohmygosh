package compiler

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"
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
		path := c.Arguments[0]
		if strings.HasPrefix(path, "~") {
			if home, err := os.UserHomeDir(); err == nil {
				path = filepath.Join(home, path[1:])
			} else {
				return errors.Join(errors.New("cd: failed to get home directory"), err)
			}
		}
		return os.Chdir(path)
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
	fmt.Fprintln(**c.Stdout, strings.Join(c.Arguments, " "))
	return nil
}

func (c *Command) execute_cat() error {
	if len(c.Arguments) > 0 {
		// for each argument, open the file and copy its contents to stdout
		for _, arg := range c.Arguments {
			if f, err := os.Open(arg); err == nil {
				if _, err := io.Copy(**c.Stdout, f); err != nil {
					fmt.Fprintln(**c.Stderr, "cat: ", err)
					return errors.Join(fmt.Errorf("cat: failed to copy file %q", arg), err)
				}
				f.Close()
			} else {
				fmt.Fprintln(**c.Stderr, "cat: ", err)
				return errors.Join(fmt.Errorf("cat: failed to open file %q", arg), err)
			}
		}
	} else {
		r := io.TeeReader(**c.Stdin, **c.Stdout)
		// read from r until EOF
		for {
			buf := make([]byte, 1024)
			_, err := r.Read(buf)
			if err != nil {
				if err == io.EOF {
					break
				}
				fmt.Fprintln(**c.Stderr, "cat: ", err)
				return errors.Join(errors.New("cat: failed to read from stdin"), err)
			}
		}
	}
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

func (c *Command) execute_which() error {
	if len(c.Arguments) > 0 {
		fs := flag.NewFlagSet("which", flag.ContinueOnError)
		fs.SetOutput(**c.Stderr)
		all := fs.Bool("a", false, "List all instances of executables found (instead of just the first one).")
		silent := fs.Bool("s", false, "Do not print anything, only return an exit status.")
		if err := fs.Parse(c.Arguments); err != nil {
			return errors.Join(fmt.Errorf("which: failed to parse arguments %q", c.Arguments), err)
		}

		pathList := strings.Split(os.Getenv("PATH"), string(os.PathListSeparator))
		for _, arg := range fs.Args() {
			found := false
			for _, path := range pathList {
				exe := filepath.Join(path, arg)
				if exe, is := isExecutable(exe); is {
					found = true
					if !*silent {
						fmt.Fprintln(**c.Stdout, exe)
					}
					if !*all {
						break
					}
				}
			}
			if !found {
				return fmt.Errorf("which: failed to find executable %q", arg)
			}
		}
	} else {
		fmt.Fprintln(**c.Stderr, "which: missing arguments")
		return errors.New("which: missing arguments")
	}
	return nil
}

func (c *Command) execute_yes() error {
	if len(c.Arguments) == 0 {
		for {
			fmt.Fprintln(**c.Stdout, "y")
			<-time.After(200 * time.Millisecond)
		}
	} else {
		for {
			fmt.Fprintln(**c.Stdout, strings.Join(c.Arguments, " "))
			<-time.After(200 * time.Millisecond)
		}
	}
}

func (c *Command) execute_true() error {
	return nil
}

func (c *Command) execute_false() error {
	return errors.New("false")
}
