package compiler_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/tsukinoko-kun/ohmygosh/compiler"
	"github.com/tsukinoko-kun/ohmygosh/runtime"
)

func TestParse(t *testing.T) {
	cases := []struct {
		text string
		in   []compiler.LexicalToken
		out  []runtime.Command
		fn   func(a *runtime.Command, b *runtime.Command) error
	}{

		{
			"echo Hello World",
			[]compiler.LexicalToken{
				{Kind: compiler.LexicalIdentifier, Content: "echo", Index: 0},
				{Kind: compiler.LexicalIdentifier, Content: "Hello World", Index: 5},
				{Kind: compiler.LexicalStop, Index: 16},
			},
			[]runtime.Command{
				{
					Executable: "echo",
					Arguments:  []string{"Hello World"},
					Background: false,
				},
			},
			nil,
		},

		{
			"command1 arg1 arg2 || command2",
			[]compiler.LexicalToken{
				{Kind: compiler.LexicalIdentifier, Content: "command1", Index: 0},
				{Kind: compiler.LexicalIdentifier, Content: "arg1", Index: 8},
				{Kind: compiler.LexicalIdentifier, Content: "arg2", Index: 13},
				{Kind: compiler.LexicalOr, Index: 18},
				{Kind: compiler.LexicalIdentifier, Content: "command2", Index: 21},
			},
			[]runtime.Command{
				{
					Executable: "command1",
					Arguments:  []string{"arg1", "arg2"},
					Background: false,
					Or: &runtime.Command{
						Executable: "command2",
						Arguments:  []string{},
						Background: false,
					},
				},
			},
			nil,
		},

		{
			"command1 arg1 arg2 && command2",
			[]compiler.LexicalToken{
				{Kind: compiler.LexicalIdentifier, Content: "command1", Index: 0},
				{Kind: compiler.LexicalIdentifier, Content: "arg1", Index: 8},
				{Kind: compiler.LexicalIdentifier, Content: "arg2", Index: 13},
				{Kind: compiler.LexicalAnd, Index: 18},
				{Kind: compiler.LexicalIdentifier, Content: "command2", Index: 21},
			},
			[]runtime.Command{
				{
					Executable: "command1",
					Arguments:  []string{"arg1", "arg2"},
					Background: false,
					And: &runtime.Command{
						Executable: "command2",
						Arguments:  []string{},
						Background: false,
					},
				},
			},
			nil,
		},

		{
			"command1 arg1 arg2 2>&1 | command2",
			[]compiler.LexicalToken{
				{Kind: compiler.LexicalIdentifier, Content: "command1", Index: 0},
				{Kind: compiler.LexicalIdentifier, Content: "arg1", Index: 8},
				{Kind: compiler.LexicalIdentifier, Content: "arg2", Index: 13},
				{Kind: compiler.LexicalStderrToStdout, Index: 18},
				{Kind: compiler.LexicalPipeStdout, Index: 23},
				{Kind: compiler.LexicalIdentifier, Content: "command2", Index: 25},
			},
			[]runtime.Command{
				{
					Executable: "command1",
					Arguments:  []string{"arg1", "arg2"},
					Background: true,
				},
				{
					Executable: "command2",
					Arguments:  []string{},
					Background: false,
				},
			},
			func(a *runtime.Command, b *runtime.Command) error {
				if a.Executable != "command1" {
					return nil
				}
				if *a.Stderr != *a.Stdout {
					return errors.New("stderr and stdout should be the same for command1")
				}
				return nil
			},
		},

		{
			"command1 arg1 arg2 | command2",
			[]compiler.LexicalToken{
				{Kind: compiler.LexicalIdentifier, Content: "command1", Index: 0},
				{Kind: compiler.LexicalIdentifier, Content: "arg1", Index: 8},
				{Kind: compiler.LexicalIdentifier, Content: "arg2", Index: 13},
				{Kind: compiler.LexicalPipeStdout, Index: 18},
				{Kind: compiler.LexicalIdentifier, Content: "command2", Index: 21},
			},
			[]runtime.Command{
				{
					Executable: "command1",
					Arguments:  []string{"arg1", "arg2"},
					Background: true,
				},
				{
					Executable: "command2",
					Arguments:  []string{},
					Background: false,
				},
			},
			func(a *runtime.Command, b *runtime.Command) error {
				if *a.Stderr == *a.Stdout {
					return errors.New("stderr and stdout should not be the same for command1")
				}
				return nil
			},
		},

		{
			"meep||echo ok&&echo meep",
			[]compiler.LexicalToken{
				{Kind: compiler.LexicalIdentifier, Content: "meep", Index: 0},
				{Kind: compiler.LexicalOr, Index: 4},
				{Kind: compiler.LexicalIdentifier, Content: "echo", Index: 6},
				{Kind: compiler.LexicalIdentifier, Content: "ok", Index: 11},
				{Kind: compiler.LexicalAnd, Index: 14},
				{Kind: compiler.LexicalIdentifier, Content: "echo", Index: 16},
				{Kind: compiler.LexicalIdentifier, Content: "meep", Index: 21},
			},
			[]runtime.Command{
				{
					Executable: "meep",
					Arguments:  []string{},
					Background: false,
					Or: &runtime.Command{
						Executable: "echo",
						Arguments:  []string{"ok"},
						Background: false,
						And: &runtime.Command{
							Executable: "echo",
							Arguments:  []string{"meep"},
							Background: false,
						},
					},
				},
			},
			nil,
		},

		{
			"echo 1 && echo 2 || echo 3",
			[]compiler.LexicalToken{
				{Kind: compiler.LexicalIdentifier, Content: "echo", Index: 0},
				{Kind: compiler.LexicalIdentifier, Content: "1", Index: 5},
				{Kind: compiler.LexicalAnd, Index: 7},
				{Kind: compiler.LexicalIdentifier, Content: "echo", Index: 10},
				{Kind: compiler.LexicalIdentifier, Content: "2", Index: 15},
				{Kind: compiler.LexicalOr, Index: 17},
				{Kind: compiler.LexicalIdentifier, Content: "echo", Index: 20},
				{Kind: compiler.LexicalIdentifier, Content: "3", Index: 25},
			},
			[]runtime.Command{
				{
					Executable: "echo",
					Arguments:  []string{"1"},
					Background: false,
					And: &runtime.Command{
						Executable: "echo",
						Arguments:  []string{"2"},
						Background: false,
						Or: &runtime.Command{
							Executable: "echo",
							Arguments:  []string{"3"},
							Background: false,
						},
					},
				},
			},
			nil,
		},
		{
			"pnpm exec astro dev --port 8002 > /dev/null &",
			[]compiler.LexicalToken{
				{Kind: compiler.LexicalIdentifier, Content: "pnpm", Index: 0},
				{Kind: compiler.LexicalIdentifier, Content: "exec", Index: 5},
				{Kind: compiler.LexicalIdentifier, Content: "astro", Index: 10},
				{Kind: compiler.LexicalIdentifier, Content: "dev", Index: 16},
				{Kind: compiler.LexicalIdentifier, Content: "--port", Index: 20},
				{Kind: compiler.LexicalIdentifier, Content: "8002", Index: 27},
				{Kind: compiler.LexicalFileStdout, Index: 32},
				{Kind: compiler.LexicalIdentifier, Content: "/dev/null", Index: 34},
				{Kind: compiler.LexicalBackground, Index: 44},
			},
			[]runtime.Command{
				{
					Executable: "pnpm",
					Arguments:  []string{"exec", "astro", "dev", "--port", "8002"},
					Background: true,
				},
			},
			nil,
		},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("Case %d %q", i, c.text), func(t *testing.T) {
			iop, _, _ := runtime.TestIoProvider("")
			defer iop.Close()
			got, err := compiler.Parse(c.text, c.in, iop)
			if err != nil {
				t.Errorf("error: %s", err)
			}
			if len(got) != len(c.out) {
				t.Errorf("got %d commands, want %d", len(got), len(c.out))
			}
			for i, want := range c.out {
				if len(got) <= i {
					t.Errorf("tried to access command %d, but there are only %d commands (expected a %q)", i, len(got), want.Executable)
					break
				}
				if err := cmdEq(got[i], &want, 0); err != nil {
					t.Errorf("command %d: %s", i, err)
				}
				if c.fn != nil {
					if err := c.fn(got[i], &want); err != nil {
						t.Errorf("command %d: %s", i, err)
					}
				}
			}
		})
	}
}

func cmdEq(got *runtime.Command, expected *runtime.Command, recursion int) error {
	if recursion > 10 {
		return fmt.Errorf("recursion limit reached")
	}
	if got.Executable != expected.Executable {
		return fmt.Errorf("executable: got: %q, want: %q", got.Executable, expected.Executable)
	}
	if len(got.Arguments) != len(expected.Arguments) {
		return fmt.Errorf("arguments: got: %s, want: %s", strArrToStr(got.Arguments), strArrToStr(expected.Arguments))
	}
	for i, arg := range got.Arguments {
		if arg != expected.Arguments[i] {
			return fmt.Errorf("arguments: got: %s, want: %s", strArrToStr(got.Arguments), strArrToStr(expected.Arguments))
		}
	}
	if got.Background != expected.Background {
		return fmt.Errorf("background: got: %t, want: %t", got.Background, expected.Background)
	}
	if expected.Or == nil {
		if got.Or != nil {
			return fmt.Errorf("or: should be nil, got: %v", got.Or)
		}
	} else {
		if got.Or == nil {
			return fmt.Errorf("or: should not be nil, got: nil")
		} else {
			if err := cmdEq(got.Or, expected.Or, recursion+1); err != nil {
				return errors.Join(errors.New("or not equal"), err)
			}
		}
	}
	if expected.And == nil {
		if got.And != nil {
			return fmt.Errorf("and: thould be nil, got: %v", got.And)
		}
	} else {
		if got.And == nil {
			return fmt.Errorf("and: should not be nil, got: nil")
		} else {
			if err := cmdEq(got.And, expected.And, recursion+1); err != nil {
				return errors.Join(errors.New("and not equal"), err)
			}
		}
	}

	return nil
}

func strArrToStr(arr []string) string {
	str := strings.Builder{}
	str.WriteByte('[')
	for i, s := range arr {
		str.WriteString(fmt.Sprintf("%q", s))
		if i < len(arr)-1 {
			str.WriteByte(',')
		}
	}
	s := str.String()
	return s
}
