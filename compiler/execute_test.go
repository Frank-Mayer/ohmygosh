package compiler_test

import (
	"fmt"
	"testing"

	"github.com/Frank-Mayer/ohmygosh/compiler"
	"github.com/Frank-Mayer/ohmygosh/runtime"
)

func TestExecute(t *testing.T) {
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
			iop, stdout, stderr := runtime.TestIoProvider(c.stdin)
			defer iop.Close()
			if err := compiler.Execute(c.in, iop); err != nil {
				t.Error(err)
				return
			}
			stdoutStr := stdout.String()
			if stdoutStr != c.stdout {
				t.Errorf("stdout: %q, expected: %q", stdoutStr, c.stdout)
			}
			stderrStr := stderr.String()
			if stderrStr != c.stderr {
				t.Errorf("stderr: %q, expected: %q", stderrStr, c.stderr)
			}
		})
	}
}
