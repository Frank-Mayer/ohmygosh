package compiler

import (
	"fmt"
	"io"
	"strings"
)

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

func Parse(text string, tokens []LexicalToken, iop *ioProvider) ([]*Command, error) {
	commands := make([]*Command, 0)
	command := newCommand(iop)
	chainMode := false
	done := func() {
		if !chainMode {
			commands = append(commands, command)
		}
		command = newCommand(iop)
	}
	for i := 0; i < len(tokens); i++ {
		switch token := tokens[i]; token.Kind {

		case LexicalIdentifier:
			if command.Executable == "" {
				command.Executable = token.Content
			} else {
				command.Arguments = append(command.Arguments, token.Content)
			}

		case LexicalStop:
			done()
			chainMode = false

		case LexicalBackground:
			command.Background = true
			done()

		case LexicalPipeStdout:
			if i+1 < len(tokens) {
				var w io.WriteCloser
				var r io.Reader
				w, r = NewPipe()
				command.Background = true
				*command.Stdout = &w
				done()
				*command.Stdin = &r
			} else {
				return nil, newParserError(token.Index, text, "unexpected end of input after pipe")
			}

		case LexicalFileStdout:
			if i+1 < len(tokens) {
				targetToken := tokens[i+1]
				if targetToken.Kind == LexicalIdentifier {
					if w, err := newFileWriter(iop.Closer, targetToken.Content); err != nil {
						return nil, newParserError(targetToken.Index, text, err.Error())
					} else {
						*command.Stdout = &w
						i++
						done()
					}
				} else {
					return nil, newParserError(token.Index, text, "expected identifier after redirect")
				}
			} else {
				return nil, newParserError(token.Index, text, "unexpected end of input after redirect")
			}

		case LexicalFileAppendStdout:
			if i+1 < len(tokens) {
				targetToken := tokens[i+1]
				if targetToken.Kind == LexicalIdentifier {
					if w, err := newFileAppendWriter(iop.Closer, targetToken.Content); err != nil {
						return nil, newParserError(targetToken.Index, text, err.Error())
					} else {
						wc := WrapWriteFakeCloser(w)
						*command.Stdout = &wc
						i++
						done()
					}
				} else {
					return nil, newParserError(token.Index, text, "expected identifier after redirect")
				}
			} else {
				return nil, newParserError(token.Index, text, "unexpected end of input after redirect")
			}

		case LexicalFileStderr:
			if i+1 < len(tokens) {
				targetToken := tokens[i+1]
				if targetToken.Kind == LexicalIdentifier {
					if w, err := newFileWriter(iop.Closer, targetToken.Content); err != nil {
						return nil, newParserError(targetToken.Index, text, err.Error())
					} else {
						*command.Stderr = &w
						i++
						done()
					}
				} else {
					return nil, newParserError(token.Index, text, "expected identifier after redirect")
				}
			} else {
				return nil, newParserError(token.Index, text, "unexpected end of input after redirect")
			}

		case LexicalFileAppendStderr:
			if i+1 < len(tokens) {
				targetToken := tokens[i+1]
				if targetToken.Kind == LexicalIdentifier {
					if w, err := newFileAppendWriter(iop.Closer, targetToken.Content); err != nil {
						return nil, newParserError(targetToken.Index, text, err.Error())
					} else {
						wc := WrapWriteFakeCloser(w)
						*command.Stderr = &wc
						i++
						done()
					}
				} else {
					return nil, newParserError(token.Index, text, "expected identifier after redirect")
				}
			} else {
				return nil, newParserError(token.Index, text, "unexpected end of input after redirect")
			}

		case LexicalFileStdoutAndStderr:
			if i+1 < len(tokens) {
				targetToken := tokens[i+1]
				if targetToken.Kind == LexicalIdentifier {
					if w, err := newFileWriter(iop.Closer, targetToken.Content); err != nil {
						return nil, newParserError(targetToken.Index, text, err.Error())
					} else {
						*command.Stdout = &w
						*command.Stderr = &w
						i++
						done()
					}
				} else {
					return nil, newParserError(token.Index, text, "expected identifier after redirect")
				}
			} else {
				return nil, newParserError(token.Index, text, "unexpected end of input after redirect")
			}

		case LexicalFileAppendStdoutAndStderr:
			if i+1 < len(tokens) {
				targetToken := tokens[i+1]
				if targetToken.Kind == LexicalIdentifier {
					if w, err := newFileAppendWriter(iop.Closer, targetToken.Content); err != nil {
						return nil, newParserError(targetToken.Index, text, err.Error())
					} else {
						wc := WrapWriteFakeCloser(w)
						*command.Stdout = &wc
						*command.Stderr = &wc
						i++
						done()
					}
				} else {
					return nil, newParserError(token.Index, text, "expected identifier after redirect")
				}
			} else {
				return nil, newParserError(token.Index, text, "unexpected end of input after redirect")
			}

		case LexicalStderrToStdout:
			command.Stderr = command.Stdout

		case LexicalStdoutToStderr:
			command.Stdout = command.Stderr

		case LexicalRedirectStdin:
			if i+1 < len(tokens) {
				targetToken := tokens[i+1]
				if targetToken.Kind == LexicalIdentifier {
					if r, err := newFileReader(iop.Closer, targetToken.Content); err != nil {
						return nil, newParserError(targetToken.Index, text, err.Error())
					} else {
						*command.Stdin = &r
						i++
					}
				} else {
					return nil, newParserError(token.Index, text, "expected identifier after redirect")
				}
			} else {
				return nil, newParserError(token.Index, text, "unexpected end of input after redirect")
			}

		case LexicalHereDocument:
			var r io.Reader = strings.NewReader(token.Content)
			*command.Stdin = &r
			done()

		case LexicalAnd:
			if i+1 < len(tokens) {
				if chainMode {
					// this is NOT the first command in the chain
					command.And = newCommand(iop)
					command = command.And
				} else {
					// this is the first command in the chain
					chainMode = true
					command.And = newCommand(iop)
					commands = append(commands, command)
					command = command.And
				}
			} else {
				return nil, newParserError(token.Index, text, "unexpected end of input after and")
			}

		case LexicalOr:
			if i+1 < len(tokens) {
				if chainMode {
					// this is NOT the first command in the chain
					command.Or = newCommand(iop)
					command = command.Or
				} else {
					// this is the first command in the chain
					chainMode = true
					command.Or = newCommand(iop)
					commands = append(commands, command)
					command = command.Or
				}
			} else {
				return nil, newParserError(token.Index, text, "unexpected end of input after or")
			}
		}
	}
	if !chainMode && command.Executable != "" {
		commands = append(commands, command)
	}
	return commands, nil
}

func newCommand(iop *ioProvider) *Command {
	stdout := &iop.DefaultOut
	stderr := &iop.DefaultErr
	stdin := &iop.DefaultIn
	return &Command{Stdout: &stdout, Stderr: &stderr, Stdin: &stdin}
}
