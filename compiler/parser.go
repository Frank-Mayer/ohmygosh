package compiler

import (
	"io"
	"strings"

	"github.com/tsukinoko-kun/ohmygosh/iohelper"
	"github.com/tsukinoko-kun/ohmygosh/runtime"
)

func Parse(text string, tokens []LexicalToken, iop *runtime.IoProvider) ([]*runtime.Command, error) {
	commands := make([]*runtime.Command, 0)
	command := runtime.NewCommand(iop)
	chainMode := false
	done := func() {
		if !chainMode {
			commands = append(commands, command)
		}
		command = runtime.NewCommand(iop)
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
				w, r = iohelper.NewPipe()
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
					if w, err := iohelper.NewFileWriter(iop.Closer, targetToken.Content); err != nil {
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
					if w, err := iohelper.NewFileAppendWriter(iop.Closer, targetToken.Content); err != nil {
						return nil, newParserError(targetToken.Index, text, err.Error())
					} else {
						wc := iohelper.WrapWriteFakeCloser(w)
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
					if w, err := iohelper.NewFileWriter(iop.Closer, targetToken.Content); err != nil {
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
					if w, err := iohelper.NewFileAppendWriter(iop.Closer, targetToken.Content); err != nil {
						return nil, newParserError(targetToken.Index, text, err.Error())
					} else {
						wc := iohelper.WrapWriteFakeCloser(w)
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
					if w, err := iohelper.NewFileWriter(iop.Closer, targetToken.Content); err != nil {
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
					if w, err := iohelper.NewFileAppendWriter(iop.Closer, targetToken.Content); err != nil {
						return nil, newParserError(targetToken.Index, text, err.Error())
					} else {
						wc := iohelper.WrapWriteFakeCloser(w)
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
					if r, err := iohelper.NewFileReader(iop.Closer, targetToken.Content); err != nil {
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
					command.And = runtime.NewCommand(iop)
					command = command.And
				} else {
					// this is the first command in the chain
					chainMode = true
					command.And = runtime.NewCommand(iop)
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
					command.Or = runtime.NewCommand(iop)
					command = command.Or
				} else {
					// this is the first command in the chain
					chainMode = true
					command.Or = runtime.NewCommand(iop)
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
