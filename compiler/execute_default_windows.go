//go:build windows

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
			if pwsh, err := exec.LookPath("pwsh"); err == nil {
				// use pwsh as a fallback (PowerShell Core)
				cmd.Path = pwsh
				cmd.Args = make([]string, 0, len(c.Arguments)+3)
				cmd.Args = append(cmd.Args, pwsh)
				cmd.Args = append(cmd.Args, "-Command")
				cmd.Args = append(cmd.Args, c.Executable)
				cmd.Args = append(cmd.Args, c.Arguments...)
			} else if powershell, err := exec.LookPath("powershell"); err == nil {
				// use powershell as a fallback
				cmd.Path = powershell
				cmd.Args = make([]string, 0, len(c.Arguments)+3)
				cmd.Args = append(cmd.Args, powershell)
				cmd.Args = append(cmd.Args, "-Command")
				cmd.Args = append(cmd.Args, c.Executable)
				cmd.Args = append(cmd.Args, c.Arguments...)
			} else if cmdExe, err := exec.LookPath("cmd"); err == nil {
				// use cmd.exe as a fallback
				cmd.Path = cmdExe
				cmd.Args = make([]string, 0, len(c.Arguments)+3)
				cmd.Args = append(cmd.Args, cmdExe)
				cmd.Args = append(cmd.Args, "/C")
				cmd.Args = append(cmd.Args, c.Executable)
				cmd.Args = append(cmd.Args, c.Arguments...)
			} else {
				return fmt.Errorf("failed to execute command %q", c.String())
			}
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
	return fmt.Errorf("sudo is not supported on Windows")
}
