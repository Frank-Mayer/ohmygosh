package compiler_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/Frank-Mayer/ohmygosh/compiler"
)

func TestLexicalAnalysis(t *testing.T) {
	t.Parallel()
	os.Setenv("TEST", "test_value")
	cases := []struct {
		in   string
		want []compiler.LexicalToken
	}{
		{
			"echo $TEST",
			[]compiler.LexicalToken{
				{Kind: compiler.LexicalTokenIdentifier, Content: "echo", Index: 0},
				{Kind: compiler.LexicalTokenIdentifier, Content: "test_value", Index: 5},
			},
		},
		{
			"echo $TEST;echo \"Hello World\"",
			[]compiler.LexicalToken{
				{Kind: compiler.LexicalTokenIdentifier, Content: "echo", Index: 0},
				{Kind: compiler.LexicalTokenIdentifier, Content: "test_value", Index: 5},
				{Kind: compiler.LexicalStop, Index: 10},
				{Kind: compiler.LexicalTokenIdentifier, Content: "echo", Index: 11},
				{Kind: compiler.LexicalTokenIdentifier, Content: "Hello World", Index: 18}, // because of the quotes; not correct but good enough for now
			},
		},
		{
			"command1 | command2",
			[]compiler.LexicalToken{
				{Kind: compiler.LexicalTokenIdentifier, Content: "command1", Index: 0},
				{Kind: compiler.LexicalPipeStdout, Index: 9},
				{Kind: compiler.LexicalTokenIdentifier, Content: "command2", Index: 11},
			},
		},
		{
			"command1 & command2",
			[]compiler.LexicalToken{
				{Kind: compiler.LexicalTokenIdentifier, Content: "command1", Index: 0},
				{Kind: compiler.LexicalBackground, Index: 9},
				{Kind: compiler.LexicalTokenIdentifier, Content: "command2", Index: 11},
			},
		},
		{
			"command1 > file",
			[]compiler.LexicalToken{
				{Kind: compiler.LexicalTokenIdentifier, Content: "command1", Index: 0},
				{Kind: compiler.LexicalFileStdout, Index: 9},
				{Kind: compiler.LexicalTokenIdentifier, Content: "file", Index: 11},
			},
		},
		{
			"command1 >> file",
			[]compiler.LexicalToken{
				{Kind: compiler.LexicalTokenIdentifier, Content: "command1", Index: 0},
				{Kind: compiler.LexicalFileAppendStdout, Index: 9},
				{Kind: compiler.LexicalTokenIdentifier, Content: "file", Index: 12},
			},
		},
		{
			"command1 2> file",
			[]compiler.LexicalToken{
				{Kind: compiler.LexicalTokenIdentifier, Content: "command1", Index: 0},
				{Kind: compiler.LexicalFileStderr, Index: 9},
				{Kind: compiler.LexicalTokenIdentifier, Content: "file", Index: 12},
			},
		},
		{
			"command1 2>> file",
			[]compiler.LexicalToken{
				{Kind: compiler.LexicalTokenIdentifier, Content: "command1", Index: 0},
				{Kind: compiler.LexicalFileAppendStderr, Index: 9},
				{Kind: compiler.LexicalTokenIdentifier, Content: "file", Index: 13},
			},
		},
		{
			"command1 &> file",
			[]compiler.LexicalToken{
				{Kind: compiler.LexicalTokenIdentifier, Content: "command1", Index: 0},
				{Kind: compiler.LexicalFileStdoutAndStderr, Index: 9},
				{Kind: compiler.LexicalTokenIdentifier, Content: "file", Index: 12},
			},
		},
		{
			"command1 < file",
			[]compiler.LexicalToken{
				{Kind: compiler.LexicalTokenIdentifier, Content: "command1", Index: 0},
				{Kind: compiler.LexicalRedirectStdin, Index: 9},
				{Kind: compiler.LexicalTokenIdentifier, Content: "file", Index: 11},
			},
		},
		{
			"command1 | command2 2>&1 | command3",
			[]compiler.LexicalToken{
				{Kind: compiler.LexicalTokenIdentifier, Content: "command1", Index: 0},
				{Kind: compiler.LexicalPipeStdout, Index: 9},
				{Kind: compiler.LexicalTokenIdentifier, Content: "command2", Index: 11},
				{Kind: compiler.LexicalStderrToStdout, Index: 20},
				{Kind: compiler.LexicalPipeStdout, Index: 25},
				{Kind: compiler.LexicalTokenIdentifier, Content: "command3", Index: 27},
			},
		},
		{
			"cat<<x\nfoo\nbar\nx",
			[]compiler.LexicalToken{
				{Kind: compiler.LexicalTokenIdentifier, Content: "cat", Index: 0},
				{Kind: compiler.LexicalHereDocument, Content: "foo\nbar", Index: 3},
			},
		},
		{
			"cat<<x\r\nfoo\r\nbar\r\nx\r\n",
			[]compiler.LexicalToken{
				{Kind: compiler.LexicalTokenIdentifier, Content: "cat", Index: 0},
				{Kind: compiler.LexicalHereDocument, Content: "foo\r\nbar", Index: 3},
			},
		},
		{
			"cat << x\nfoo\nbar\nx\nrm -rf /",
			[]compiler.LexicalToken{
				{Kind: compiler.LexicalTokenIdentifier, Content: "cat", Index: 0},
				{Kind: compiler.LexicalHereDocument, Content: "foo\nbar", Index: 4},
				{Kind: compiler.LexicalStop, Index: 18},
				{Kind: compiler.LexicalTokenIdentifier, Content: "rm", Index: 19},
				{Kind: compiler.LexicalTokenIdentifier, Content: "-rf", Index: 22},
				{Kind: compiler.LexicalTokenIdentifier, Content: "/", Index: 26},
			},
		},
		{
			"cat << x\nfoo\nbar\nx\nrm -rf /",
			[]compiler.LexicalToken{
				{Kind: compiler.LexicalTokenIdentifier, Content: "cat", Index: 0},
				{Kind: compiler.LexicalHereDocument, Content: "foo\nbar", Index: 4},
				{Kind: compiler.LexicalStop, Index: 18},
				{Kind: compiler.LexicalTokenIdentifier, Content: "rm", Index: 19},
				{Kind: compiler.LexicalTokenIdentifier, Content: "-rf", Index: 22},
				{Kind: compiler.LexicalTokenIdentifier, Content: "/", Index: 26},
			},
		},
		{
			"echo \"Hello World\"",
			[]compiler.LexicalToken{
				{Kind: compiler.LexicalTokenIdentifier, Content: "echo", Index: 0},
				{Kind: compiler.LexicalTokenIdentifier, Content: "Hello World", Index: 7}, // because of the quotes; not correct but good enough for now
			},
		},
		{
			"echo 'Hello World'",
			[]compiler.LexicalToken{
				{Kind: compiler.LexicalTokenIdentifier, Content: "echo", Index: 0},
				{Kind: compiler.LexicalTokenIdentifier, Content: "Hello World", Index: 7}, // because of the quotes; not correct but good enough for now
			},
		},
		{
			"echo \"Hello World\";echo 'Hello World'",
			[]compiler.LexicalToken{
				{Kind: compiler.LexicalTokenIdentifier, Content: "echo", Index: 0},
				{Kind: compiler.LexicalTokenIdentifier, Content: "Hello World", Index: 7}, // because of the quotes; not correct but good enough for now
				{Kind: compiler.LexicalStop, Index: 18},
				{Kind: compiler.LexicalTokenIdentifier, Content: "echo", Index: 19},
				{Kind: compiler.LexicalTokenIdentifier, Content: "Hello World", Index: 26}, // because of the quotes; not correct but good enough for now
			},
		},
		{
			"echo a\\tb",
			[]compiler.LexicalToken{
				{Kind: compiler.LexicalTokenIdentifier, Content: "echo", Index: 0},
				{Kind: compiler.LexicalTokenIdentifier, Content: "atb", Index: 6}, // because of the quotes; not correct but good enough for now
			},
		},
		{
			"echo \"a\\tb\"",
			[]compiler.LexicalToken{
				{Kind: compiler.LexicalTokenIdentifier, Content: "echo", Index: 0},
				{Kind: compiler.LexicalTokenIdentifier, Content: "a\tb", Index: 8}, // because of the quotes; not correct but good enough for now
			},
		},
		{
			"echo \"\\g\\t\\g\"",
			[]compiler.LexicalToken{
				{Kind: compiler.LexicalTokenIdentifier, Content: "echo", Index: 0},
				{Kind: compiler.LexicalTokenIdentifier, Content: "\\g\t\\g", Index: 8}, // because of the quotes; not correct but good enough for now
			},
		},
		{
			"echo $(echo 1)",
			[]compiler.LexicalToken{
				{Kind: compiler.LexicalTokenIdentifier, Content: "echo", Index: 0},
				{Kind: compiler.LexicalTokenIdentifier, Content: "1", Index: 13}, // this index is wrong because of the $(...) substitution
			},
		},
		{
			"echo \">$(echo 1)<\"",
			[]compiler.LexicalToken{
				{Kind: compiler.LexicalTokenIdentifier, Content: "echo", Index: 0},
				{Kind: compiler.LexicalTokenIdentifier, Content: ">1<", Index: 15}, // this index is wrong because of the $(...) substitution
			},
		},
		{
			"echo \">$(echo $TEST)<\"",
			[]compiler.LexicalToken{
				{Kind: compiler.LexicalTokenIdentifier, Content: "echo", Index: 0},
				{Kind: compiler.LexicalTokenIdentifier, Content: ">test_value<", Index: 10}, // this index is wrong because of the $(...) substitution
			},
		},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("%d %q", i, c.in), func(t *testing.T) {
			iop := compiler.TestIoProvider()
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
