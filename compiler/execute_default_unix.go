//go:build !windows

package compiler

import (
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
)

func (c *Command) execute_default() error {
	cmd := &exec.Cmd{
		Stdin:  **c.Stdin,
		Stdout: **c.Stdout,
		Stderr: **c.Stderr,
	}

	if exe, err := exec.LookPath(c.Executable); err == nil {
		cmd.Path = exe
		cmd.Args = append([]string{exe}, c.Arguments...)
	} else {
		if exe, err := filepath.Abs(c.Executable); err == nil {
			cmd.Path = exe
			cmd.Args = append([]string{exe}, c.Arguments...)
		} else {
			cmd.Path = c.Executable
			cmd.Args = append([]string{c.Executable}, c.Arguments...)
		}
	}

	err := cmd.Run()
	if err != nil {
		fmt.Fprintf(**c.Stderr, "failed to execute command: %s\n", c.String())
		fmt.Fprintln(**c.Stderr, err.Error())
		return errors.Join(fmt.Errorf("failed to execute command %q", c.String()), err)
	}
	return nil
}

func (c *Command) execute_sudo() error {
	cmd := &exec.Cmd{
		Stdin:  **c.Stdin,
		Stdout: **c.Stdout,
		Stderr: **c.Stderr,
		Path:   "sudo",
		Args:   append([]string{"sudo", c.Executable}, c.Arguments...),
	}

	err := cmd.Run()
	if err != nil {
		return errors.Join(fmt.Errorf("failed to execute command %q", c.String()), err)
	}
	return nil
}
