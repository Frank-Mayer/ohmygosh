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
		fmt.Fprintf(**c.Stderr, "%s: failed to execute command: %s\n", filepath.Base(cmd.Path), err)
		return errors.Join(fmt.Errorf("failed to execute command %q", c.String()), err)
	}
	return nil
}

func (c *Command) execute_sudo() error {
	sudoPath, err := exec.LookPath("sudo")
	if err != nil {
		fmt.Fprintf(**c.Stderr, "sudo: %s\n", err)
		return errors.Join(fmt.Errorf("failed to execute command %q", c.String()), err)
	}
	cmd := &exec.Cmd{
		Stdin:  **c.Stdin,
		Stdout: **c.Stdout,
		Stderr: **c.Stderr,
		Path:   sudoPath,
		Args:   append([]string{"sudo", c.Executable}, c.Arguments...),
	}

	err = cmd.Run()
	if err != nil {
		fmt.Fprintf(**c.Stderr, "sudo: failed to execute command: %s\n", err)
		return errors.Join(fmt.Errorf("failed to execute command %q", c.String()), err)
	}
	return nil
}
