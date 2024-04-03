package runtime

import (
	"fmt"
	"io"
	"strings"
)

func NewCommand(iop *IoProvider) *Command {
	stdout := &iop.DefaultOut
	stderr := &iop.DefaultErr
	stdin := &iop.DefaultIn
	return &Command{Stdout: &stdout, Stderr: &stderr, Stdin: &stdin}
}

type (
	Command struct {
		Executable string
		Arguments  []string
		Background bool
		Stdout     **io.WriteCloser
		Stderr     **io.WriteCloser
		Stdin      **io.Reader
		And        *Command
		Or         *Command
	}
)

func (c *Command) String() string {
	str := strings.Builder{}
	str.WriteString(c.Executable)
	for _, arg := range c.Arguments {
		str.WriteString(" ")
		str.WriteString(fmt.Sprintf("%q", arg))
	}
	if c.Or != nil {
		str.WriteString(" || ")
		str.WriteString(c.Or.String())
	}
	if c.And != nil {
		str.WriteString(" && ")
		str.WriteString(c.And.String())
	}
	return str.String()
}

func (c *Command) Execute(iop *IoProvider) error {
	var err error

	if fn, builtin := BuiltinCommands[strings.ToLower(c.Executable)]; builtin {
		err = fn(c, iop)
	} else {
		err = Execute_default(c, iop)
	}

	if err != nil {
		// failed
		if c.Or != nil {
			return c.Or.Execute(iop)
		} else {
			return err
		}
	} else {
		// succeeded
		if c.And != nil {
			return c.And.Execute(iop)
		}
	}

	return nil
}
