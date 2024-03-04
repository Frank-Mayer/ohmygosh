package compiler_test

import (
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/Frank-Mayer/ohmygosh/compiler"
)

func TestExecute(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in     string
		stdout string
		stderr string
		stdin  string
	}{
		{
			`echo "hello world"`,
			"hello world\n",
			"",
			"",
		},
		{
			"echo hello world",
			"hello world\n",
			"",
			"",
		},
		{
			`echo "hello world"|cat`,
			"hello world\n",
			"",
			"",
		},
		{
			"cat",
			"hello world\n",
			"",
			"hello world\n",
		},
		{
			`cat <<xyz
hello
world
xyz`,
			"hello\nworld\n",
			"",
			"",
		},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("case %d %q", i, c.in), func(t *testing.T) {
			iop, stdout, stderr, stdin := compiler.TestIoProvider()
			defer iop.Close()
			if _, err := stdin.Write([]byte(c.stdin)); err != nil {
				t.Errorf("failed to write to stdin: %v", err)
				return
			}
			if err := compiler.Execute(c.in, iop); err != nil {
				t.Error(err)
				return
			}
			stdoutB := &strings.Builder{}
			if _, err := io.Copy(stdoutB, stdout); err != nil {
				t.Errorf("failed to read from stdout: %v", err)
				return
			}
			stderrB := &strings.Builder{}
			if _, err := io.Copy(stderrB, stderr); err != nil {
				t.Errorf("failed to read from stderr: %v", err)
				return
			}
			stdoutStr := stdoutB.String()
			if stdoutStr != c.stdout {
				t.Errorf("stdout: %q, expected: %q", stdoutStr, c.stdout)
			}
			stderrStr := stderrB.String()
			if stderrStr != c.stderr {
				t.Errorf("stderr: %q, expected: %q", stderrStr, c.stderr)
			}
		})
	}
}
