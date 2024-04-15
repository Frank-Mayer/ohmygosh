package compiler_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/tsukinoko-kun/ohmygosh/compiler"
	"github.com/tsukinoko-kun/ohmygosh/runtime"
)

func TestLexicalAnalysis(t *testing.T) {
	os.Setenv("TEST", "test_value")
	cases := []struct {
		in   string
		want []compiler.LexicalToken
	}{
		{
			"echo $TEST",
			[]compiler.LexicalToken{
				{Kind: compiler.LexicalIdentifier, Content: "echo", Index: 0},
				{Kind: compiler.LexicalIdentifier, Content: "test_value", Index: 5},
			},
		},
		{
			"echo $TEST/abc",
			[]compiler.LexicalToken{
				{Kind: compiler.LexicalIdentifier, Content: "echo", Index: 0},
				{Kind: compiler.LexicalIdentifier, Content: "test_value/abc", Index: 5},
			},
		},
		{
			"echo $TEST foo bar",
			[]compiler.LexicalToken{
				{Kind: compiler.LexicalIdentifier, Content: "echo", Index: 0},
				{Kind: compiler.LexicalIdentifier, Content: "test_value", Index: 5},
				{Kind: compiler.LexicalIdentifier, Content: "foo", Index: 11},
				{Kind: compiler.LexicalIdentifier, Content: "bar", Index: 15},
			},
		},
		{
			"echo $TEST;echo \"Hello World\"",
			[]compiler.LexicalToken{
				{Kind: compiler.LexicalIdentifier, Content: "echo", Index: 0},
				{Kind: compiler.LexicalIdentifier, Content: "test_value", Index: 5},
				{Kind: compiler.LexicalStop, Index: 10},
				{Kind: compiler.LexicalIdentifier, Content: "echo", Index: 11},
				{Kind: compiler.LexicalIdentifier, Content: "Hello World", Index: 16},
			},
		},
		{
			"command1 | command2",
			[]compiler.LexicalToken{
				{Kind: compiler.LexicalIdentifier, Content: "command1", Index: 0},
				{Kind: compiler.LexicalPipeStdout, Index: 9},
				{Kind: compiler.LexicalIdentifier, Content: "command2", Index: 11},
			},
		},
		{
			"command1 & command2",
			[]compiler.LexicalToken{
				{Kind: compiler.LexicalIdentifier, Content: "command1", Index: 0},
				{Kind: compiler.LexicalBackground, Index: 9},
				{Kind: compiler.LexicalIdentifier, Content: "command2", Index: 11},
			},
		},
		{
			"command1 > file",
			[]compiler.LexicalToken{
				{Kind: compiler.LexicalIdentifier, Content: "command1", Index: 0},
				{Kind: compiler.LexicalFileStdout, Index: 9},
				{Kind: compiler.LexicalIdentifier, Content: "file", Index: 11},
			},
		},
		{
			"command1 >> file",
			[]compiler.LexicalToken{
				{Kind: compiler.LexicalIdentifier, Content: "command1", Index: 0},
				{Kind: compiler.LexicalFileAppendStdout, Index: 9},
				{Kind: compiler.LexicalIdentifier, Content: "file", Index: 12},
			},
		},
		{
			"command1 2> file",
			[]compiler.LexicalToken{
				{Kind: compiler.LexicalIdentifier, Content: "command1", Index: 0},
				{Kind: compiler.LexicalFileStderr, Index: 9},
				{Kind: compiler.LexicalIdentifier, Content: "file", Index: 12},
			},
		},
		{
			"command1 2>> file",
			[]compiler.LexicalToken{
				{Kind: compiler.LexicalIdentifier, Content: "command1", Index: 0},
				{Kind: compiler.LexicalFileAppendStderr, Index: 9},
				{Kind: compiler.LexicalIdentifier, Content: "file", Index: 13},
			},
		},
		{
			"command1 &> file",
			[]compiler.LexicalToken{
				{Kind: compiler.LexicalIdentifier, Content: "command1", Index: 0},
				{Kind: compiler.LexicalFileStdoutAndStderr, Index: 9},
				{Kind: compiler.LexicalIdentifier, Content: "file", Index: 12},
			},
		},
		{
			"command1 < file",
			[]compiler.LexicalToken{
				{Kind: compiler.LexicalIdentifier, Content: "command1", Index: 0},
				{Kind: compiler.LexicalRedirectStdin, Index: 9},
				{Kind: compiler.LexicalIdentifier, Content: "file", Index: 11},
			},
		},
		{
			"command1 | command2 2>&1 | command3",
			[]compiler.LexicalToken{
				{Kind: compiler.LexicalIdentifier, Content: "command1", Index: 0},
				{Kind: compiler.LexicalPipeStdout, Index: 9},
				{Kind: compiler.LexicalIdentifier, Content: "command2", Index: 11},
				{Kind: compiler.LexicalStderrToStdout, Index: 20},
				{Kind: compiler.LexicalPipeStdout, Index: 25},
				{Kind: compiler.LexicalIdentifier, Content: "command3", Index: 27},
			},
		},
		{
			"cat<<x\nfoo\nbar\nx",
			[]compiler.LexicalToken{
				{Kind: compiler.LexicalIdentifier, Content: "cat", Index: 0},
				{Kind: compiler.LexicalHereDocument, Content: "foo\nbar\n", Index: 3},
			},
		},
		{
			"cat<<x\r\nfoo\r\nbar\r\nx\r\n",
			[]compiler.LexicalToken{
				{Kind: compiler.LexicalIdentifier, Content: "cat", Index: 0},
				{Kind: compiler.LexicalHereDocument, Content: "foo\nbar\n", Index: 3},
			},
		},
		{
			"cat << x\nfoo\nbar\nx\nrm -rf /",
			[]compiler.LexicalToken{
				{Kind: compiler.LexicalIdentifier, Content: "cat", Index: 0},
				{Kind: compiler.LexicalHereDocument, Content: "foo\nbar\n", Index: 4},
				{Kind: compiler.LexicalStop, Index: 18},
				{Kind: compiler.LexicalIdentifier, Content: "rm", Index: 19},
				{Kind: compiler.LexicalIdentifier, Content: "-rf", Index: 22},
				{Kind: compiler.LexicalIdentifier, Content: "/", Index: 26},
			},
		},
		{
			"cat << x\nfoo\nbar\nx\nrm -rf /",
			[]compiler.LexicalToken{
				{Kind: compiler.LexicalIdentifier, Content: "cat", Index: 0},
				{Kind: compiler.LexicalHereDocument, Content: "foo\nbar\n", Index: 4},
				{Kind: compiler.LexicalStop, Index: 18},
				{Kind: compiler.LexicalIdentifier, Content: "rm", Index: 19},
				{Kind: compiler.LexicalIdentifier, Content: "-rf", Index: 22},
				{Kind: compiler.LexicalIdentifier, Content: "/", Index: 26},
			},
		},
		{
			"echo \"Hello World\"",
			[]compiler.LexicalToken{
				{Kind: compiler.LexicalIdentifier, Content: "echo", Index: 0},
				{Kind: compiler.LexicalIdentifier, Content: "Hello World", Index: 5},
			},
		},
		{
			"echo 'Hello World'",
			[]compiler.LexicalToken{
				{Kind: compiler.LexicalIdentifier, Content: "echo", Index: 0},
				{Kind: compiler.LexicalIdentifier, Content: "Hello World", Index: 5},
			},
		},
		{
			"echo \"Hello World\";echo 'Hello World'",
			[]compiler.LexicalToken{
				{Kind: compiler.LexicalIdentifier, Content: "echo", Index: 0},
				{Kind: compiler.LexicalIdentifier, Content: "Hello World", Index: 5},
				{Kind: compiler.LexicalStop, Index: 18},
				{Kind: compiler.LexicalIdentifier, Content: "echo", Index: 19},
				{Kind: compiler.LexicalIdentifier, Content: "Hello World", Index: 24},
			},
		},
		{
			"echo a\\tb",
			[]compiler.LexicalToken{
				{Kind: compiler.LexicalIdentifier, Content: "echo", Index: 0},
				{Kind: compiler.LexicalIdentifier, Content: "atb", Index: 5},
			},
		},
		{
			"echo \"a\\tb\"",
			[]compiler.LexicalToken{
				{Kind: compiler.LexicalIdentifier, Content: "echo", Index: 0},
				{Kind: compiler.LexicalIdentifier, Content: "a\tb", Index: 5},
			},
		},
		{
			"echo \"\\g\\t\\g\"",
			[]compiler.LexicalToken{
				{Kind: compiler.LexicalIdentifier, Content: "echo", Index: 0},
				{Kind: compiler.LexicalIdentifier, Content: "\\g\t\\g", Index: 5},
			},
		},
		{
			"echo $(echo 1)",
			[]compiler.LexicalToken{
				{Kind: compiler.LexicalIdentifier, Content: "echo", Index: 0},
				{Kind: compiler.LexicalIdentifier, Content: "1", Index: 5},
			},
		},
		{
			"echo \">$(echo 1)<\"",
			[]compiler.LexicalToken{
				{Kind: compiler.LexicalIdentifier, Content: "echo", Index: 0},
				{Kind: compiler.LexicalIdentifier, Content: ">1<", Index: 5},
			},
		},
		{
			"echo \">$(echo $TEST)<\"",
			[]compiler.LexicalToken{
				{Kind: compiler.LexicalIdentifier, Content: "echo", Index: 0},
				{Kind: compiler.LexicalIdentifier, Content: ">test_value<", Index: 5},
			},
		},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("%d %q", i, c.in), func(t *testing.T) {
			iop, _, _ := runtime.TestIoProvider("")
			defer iop.Close()
			got, err := compiler.LexicalAnalysis(c.in, iop)
			if err != nil {
				t.Errorf("returned error: %v", err)
				return
			}
			if got == nil {
				t.Errorf("returned nil")
				return
			}
			if len(got) != len(c.want) {
				t.Errorf("got %d tokens, want %d: %v", len(got), len(c.want), got)
			}
			for i, wantToken := range c.want {
				if len(got) <= i {
					t.Errorf("not enough tokens returned: got %d, want %d", len(got), len(c.want))
					return
				}
				gotToken := got[i]
				if wantToken.Kind != gotToken.Kind {
					t.Errorf("wrong token kind at index %d: got %d, want %d", i, gotToken.Kind, wantToken.Kind)
					return
				}
				if !compareString(wantToken.Content, gotToken.Content) {
					t.Errorf("wrong token content at index %d: got %q, want %q", i, gotToken.Content, wantToken.Content)
					return
				}
				if wantToken.Index != gotToken.Index {
					t.Errorf("wrong token index at index %d: got %d, want %d", i, gotToken.Index, wantToken.Index)
					return
				}
			}
		})
	}
}

func compareString(a string, b string) bool {
	ba := []byte(a)
	bb := []byte(b)
	if len(ba) != len(bb) {
		return false
	}
	for i := range ba {
		if ba[i] != bb[i] {
			return false
		}
	}
	return true
}
