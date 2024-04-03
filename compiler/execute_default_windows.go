//go:build windows

package compiler

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func execute_default(c *Command, _ *ioProvider) error {
	cmd := &exec.Cmd{
		Stdin:  **c.Stdin,
		Stdout: **c.Stdout,
		Stderr: **c.Stderr,
	}

	if exe, err := exec.LookPath(c.Executable); err == nil {
		cmd.Path = exe
		cmd.Args = append([]string{exe}, c.Arguments...)
	} else if exe, err := filepath.Abs(c.Executable); err == nil && exists(exe) {
		cmd.Path = exe
		cmd.Args = append([]string{exe}, c.Arguments...)
	} else if pwsh, err := exec.LookPath("pwsh"); err == nil {
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

	err := cmd.Run()
	if err != nil {
		_, _ = fmt.Fprintf(**c.Stderr, "%s: failed to execute command: %s\n", filepath.Base(cmd.Path), err)
		return errors.Join(fmt.Errorf("failed to execute command %q", c.String()), err)
	}
	return nil
}

func execute_sudo(c *Command, _ *ioProvider) error {
	sudoPath, err := exec.LookPath("sudo")
	if err != nil {
		_, _ = fmt.Fprintf(**c.Stderr, "sudo: %s\n", err)
		_, _ = fmt.Fprintln(**c.Stderr, "install sudo for Windows from https://github.com/microsoft/sudo")
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
		_, _ = fmt.Fprintf(**c.Stderr, "sudo: failed to execute command: %s\n", err)
		return errors.Join(fmt.Errorf("failed to execute command %q", c.String()), err)
	}
	return nil
}

func exists(path string) bool {
	_, err := exec.LookPath(path)
	return err == nil
}

var pathExt = strings.Split(os.Getenv("PATHEXT"), string(os.PathListSeparator))

func isExecutable(path string) (string, bool) {
	if fi, err := os.Stat(path); err == nil && !fi.IsDir() {
		return path, true
	}
	for _, ext := range pathExt {
		extendedPath := path + strings.TrimSpace(ext)
		if fi, err := os.Stat(extendedPath); err == nil && !fi.IsDir() {
			return extendedPath, true
		}
	}
	return "", false
}
