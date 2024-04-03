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

func execute_cd(c *Command, _ *ioProvider) error {
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
		_, _ = fmt.Fprintln(**c.Stderr, "cd: too many arguments")
		return errors.New("cd: too many arguments")
	}
}

func execute_exit(c *Command, _ *ioProvider) error {
	switch len(c.Arguments) {
	case 0:
		os.Exit(0)
	case 1:
		code, err := strconv.Atoi(c.Arguments[0])
		if err != nil {
			_, _ = fmt.Fprintln(**c.Stderr, "exit: ", err)
			return errors.Join(fmt.Errorf("exit: failed to parse argument %q as an integer", c.Arguments[0]), err)
		}
		os.Exit(code)
	default:
		_, _ = fmt.Fprintln(**c.Stderr, "exit: too many arguments")
		return errors.New("exit: too many arguments")
	}
	return nil
}

func execute_echo(c *Command, _ *ioProvider) error {
	_, _ = fmt.Fprintln(**c.Stdout, strings.Join(c.Arguments, " "))
	return nil
}

func execute_cat(c *Command, iop *ioProvider) error {
	if len(c.Arguments) > 0 {
		// for each argument, open the file and copy its contents to stdout
		for _, arg := range c.Arguments {
			r, err := newFileReader(iop.Closer, arg)
			if err != nil {
				return errors.Join(fmt.Errorf("cat: failed to open file %q", arg), err)
			}
			_, err = io.Copy(**c.Stdout, r)
			if err != nil {
				return errors.Join(fmt.Errorf("cat: failed to read file %q", arg), err)
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
				_, _ = fmt.Fprintln(**c.Stderr, "cat: ", err)
				return errors.Join(errors.New("cat: failed to read from stdin"), err)
			}
		}
	}
	return nil
}

func execute_export(c *Command, _ *ioProvider) error {
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
			_, _ = fmt.Fprintf(**c.Stdout, "declare -x %s\n", env)
		}
	}
	return nil
}

func execute_unset(c *Command, _ *ioProvider) error {
	if len(c.Arguments) > 0 {
		for _, arg := range c.Arguments {
			os.Unsetenv(arg)
		}
	}
	return nil
}

func execute_whoami(c *Command, _ *ioProvider) error {
	if u, err := user.Current(); err == nil {
		_, _ = fmt.Fprintln(**c.Stdout, u.Username)
	} else {
		_, _ = fmt.Fprintln(**c.Stderr, "whoami: ", err)
		return errors.Join(errors.New("whoami: failed to get current user"), err)
	}
	return nil
}

func execute_pwd(c *Command, _ *ioProvider) error {
	if wd, err := os.Getwd(); err == nil {
		_, _ = fmt.Fprintln(**c.Stdout, wd)
	} else {
		_, _ = fmt.Fprintln(**c.Stderr, "pwd: ", err)
		return errors.Join(errors.New("pwd: failed to get working directory"), err)
	}
	return nil
}

func findExecutable(name string, all bool) []string {
	pathList := strings.Split(os.Getenv("PATH"), string(os.PathListSeparator))
	foundBinaries := []string{}

	for _, path := range pathList {
		exe := filepath.Join(path, name)
		if exe, is := isExecutable(exe); is {
			foundBinaries = append(foundBinaries, exe)
			if !all {
				break
			}
		}
	}

	return foundBinaries
}

func execute_type(c *Command, _ *ioProvider) error {
	if len(c.Arguments) != 0 {
		for _, arg := range c.Arguments {
			if _, ok := builtinCommands[arg]; ok {
				_, _ = fmt.Fprintf(**c.Stdout, "%s is a builtin\n", arg)
				continue
			}
			foundBinaries := findExecutable(arg, false)
			if len(foundBinaries) != 0 {
				_, _ = fmt.Fprintf(**c.Stdout, "%s is %s\n", arg, foundBinaries[0])
			} else {
				_, _ = fmt.Fprintf(**c.Stderr, "type: could not find %s\n", arg)
			}
		}
	} else {
		_, _ = fmt.Fprintln(**c.Stderr, "type: missing arguments")
		return errors.New("type: missing arguments")
	}
	return nil
}

func execute_which(c *Command, _ *ioProvider) error {
	if len(c.Arguments) != 0 {
		fs := flag.NewFlagSet("which", flag.ContinueOnError)
		fs.SetOutput(**c.Stderr)
		all := fs.Bool("a", false, "List all instances of executables found (instead of just the first one).")
		silent := fs.Bool("s", false, "Do not print anything, only return an exit status.")
		if err := fs.Parse(c.Arguments); err != nil {
			return errors.Join(fmt.Errorf("which: failed to parse arguments %q", c.Arguments), err)
		}

		for _, arg := range fs.Args() {
			foundBinaries := findExecutable(arg, *all)
			if len(foundBinaries) != 0 {
				if !*silent {
					for _, exe := range foundBinaries {
						_, _ = fmt.Fprintln(**c.Stdout, exe)
					}
				}
			} else {
				return fmt.Errorf("which: %s not found", arg)
			}
		}
	} else {
		_, _ = fmt.Fprintln(**c.Stderr, "which: missing arguments")
		return errors.New("which: missing arguments")
	}
	return nil
}

func execute_yes(c *Command, _ *ioProvider) error {
	if len(c.Arguments) == 0 {
		for {
			_, _ = fmt.Fprintln(**c.Stdout, "y")
			<-time.After(200 * time.Millisecond)
		}
	} else {
		for {
			_, _ = fmt.Fprintln(**c.Stdout, strings.Join(c.Arguments, " "))
			<-time.After(200 * time.Millisecond)
		}
	}
}

func execute_true(c *Command, _ *ioProvider) error {
	return nil
}

func execute_false(c *Command, _ *ioProvider) error {
	return errors.New("false")
}

func execute_sleep(c *Command, _ *ioProvider) error {
	if len(c.Arguments) == 0 {
		return errors.New("sleep: missing operand")
	}
	if len(c.Arguments) > 1 {
		return errors.New("sleep: too many arguments")
	}
	d, err := time.ParseDuration(c.Arguments[0])
	if err != nil {
		// try to parse the duration as a float
		if f, err := strconv.ParseFloat(c.Arguments[0], 64); err == nil {
			d = time.Duration(f * float64(time.Second))
		} else {
			return errors.Join(fmt.Errorf("sleep: failed to parse argument %q as a duration", c.Arguments[0]), err)
		}
	}
	<-time.After(d)
	return nil
}
