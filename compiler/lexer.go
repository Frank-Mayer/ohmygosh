package compiler

import (
	"fmt"
	"os"
	"strings"
)

func (k LexicalTokenKind) String() string {
	return fmt.Sprintf("LexicalTokenKind(%d)", k)
}

const (
	// program name, command, argument, variable
	LexicalIdentifier LexicalTokenKind = iota
	// ;
	LexicalStop
	// &
	LexicalBackground
	// |
	LexicalPipeStdout
	// >
	LexicalFileStdout
	// >>
	LexicalFileAppendStdout
	// 2>
	LexicalFileStderr
	// 2>>
	LexicalFileAppendStderr
	// &>
	LexicalFileStdoutAndStderr
	// &>>
	LexicalFileAppendStdoutAndStderr
	// 2>&1
	LexicalStderrToStdout
	// 1>&2
	LexicalStdoutToStderr
	// <
	LexicalRedirectStdin
	// <<
	LexicalHereDocument
	// &&
	LexicalAnd
	// ||
	LexicalOr
)

type (
	LexicalToken struct {
		Content string
		// Index is the position of the token in the text.
		// This is not 100% accurate because it doesn't take quotation into account.
		// It's good enough for error messages.
		// Index is always the position of the first character of the token or within the token (in case of quotation).
		Index int
		Kind  LexicalTokenKind
	}

	LexicalTokenBuilder struct {
		Content strings.Builder
		Index   int
		Kind    LexicalTokenKind
	}

	lexicalQuotation uint8
	LexicalTokenKind uint8
)

func (t LexicalToken) String() string {
	return fmt.Sprintf("LexicalToken{Content: %q, Index: %d, Kind: %s}", t.Content, t.Index, t.Kind.String())
}

func newLexicalTokenBuilder() LexicalTokenBuilder {
	b := LexicalTokenBuilder{}
	b.Reset()
	return b
}

func (b *LexicalTokenBuilder) SetIndex(i int) {
	b.Index = i
}

func (b *LexicalTokenBuilder) SetIndexIfEmpty(i int) {
	if b.Index == -1 {
		b.Index = i
	}
}

func (b *LexicalTokenBuilder) WriteChar(c byte, i int) {
	b.Content.WriteByte(c)
	if b.Index == -1 {
		b.Index = i
	}
}

func (b *LexicalTokenBuilder) WriteString(s string, i int) {
	b.Content.WriteString(s)
	if b.Index == -1 {
		b.Index = i
	}
}
func (b *LexicalTokenBuilder) WriteRune(r rune, i int) {
	b.Content.WriteRune(r)
	if b.Index == -1 {
		b.Index = i
	}
}

func (b *LexicalTokenBuilder) Build() LexicalToken {
	t := LexicalToken{
		Content: b.Content.String(),
		Index:   b.Index,
		Kind:    b.Kind,
	}
	b.Reset()
	return t
}

func (t *LexicalTokenBuilder) Reset() {
	t.Content.Reset()
	t.Kind = LexicalIdentifier
	t.Index = -1
}

func (t *LexicalTokenBuilder) SetKind(kind LexicalTokenKind) {
	t.Kind = kind
}

func (t *LexicalTokenBuilder) IsPresent() bool {
	return t.Index != -1
}

const (
	lexicalQuotationNone lexicalQuotation = iota
	lexicalQuotationSingle
	lexicalQuotationDouble
)

// LexicalAnalysis performs lexical analysis on the given text and returns a slice of tokens.
// If an error gets returned it will be of type CompilerError.
func LexicalAnalysis(text string, iop *ioProvider) ([]LexicalToken, error) {
	texLen := len(text)
	tokens := make([]LexicalToken, 0)
	quotation := lexicalQuotationNone
	tb := newLexicalTokenBuilder()

	for i := 0; i < texLen; i++ {
		switch c := text[i]; c {

		case '\n':
			if quotation != lexicalQuotationNone {
				return nil, newLexicalError(i, text, "quotation not closed at the end of the line")
			}
			if tb.IsPresent() {
				tokens = append(tokens, tb.Build())
			}
			tokens = append(tokens, LexicalToken{Kind: LexicalStop, Index: i})

		case '\r':
			if quotation != lexicalQuotationNone {
				tb.WriteChar(c, i)
			}
			continue

		case ' ', '\t', '\v', '\f', 20:
			if quotation == lexicalQuotationNone {
				if tb.IsPresent() {
					tokens = append(tokens, tb.Build())
				}
			} else {
				tb.WriteChar(c, i)
			}

		case '$':
			if i+1 < texLen && text[i+1] == '(' {
				// subshell
				tb.SetIndexIfEmpty(i)
				i += 2
				var subshell strings.Builder
				braceCount := 1
				for i < texLen {
					c := text[i]
					if c == '(' {
						braceCount++
					} else if c == ')' {
						braceCount--
						if braceCount == 0 {
							break
						}
					}
					subshell.WriteByte(c)
					i++
				}
				iop, sb := subshellIoProvider(iop)
				defer iop.Close()
				if err := Execute(subshell.String(), iop); err != nil {
					return nil, newLexicalError(i, text, fmt.Sprintf("failed to execute subshell: %v", err))
				}
				tb.WriteString(strings.TrimSpace(sb.String()), i)
			} else {
				// variable
				tb.SetIndexIfEmpty(i)
				varName := strings.Builder{}
				for i += 1; i < texLen; i++ {
					c := text[i]
					if c == ' ' || c == ';' || c == '\t' || c == '\v' || c == '\f' || c == '\n' || c == '\r' || c == '.' || c == ',' || c == '/' || c == '>' || c == '<' || c == '&' || c == '|' {
						i--
						break
					}
					varName.WriteByte(c)
				}
				tb.WriteString(os.Getenv(varName.String()), i)
			}

		case '"':
			tb.SetIndexIfEmpty(i)
			switch quotation {
			case lexicalQuotationNone:
				quotation = lexicalQuotationDouble
			case lexicalQuotationDouble:
				quotation = lexicalQuotationNone
			case lexicalQuotationSingle:
				tb.WriteChar(c, i)
			default:
				return nil, newLexicalError(i, text, fmt.Sprintf("invalid quotation state: %d", quotation))
			}

		case '\'':
			tb.SetIndexIfEmpty(i)
			switch quotation {
			case lexicalQuotationNone:
				quotation = lexicalQuotationSingle
			case lexicalQuotationSingle:
				quotation = lexicalQuotationNone
			case lexicalQuotationDouble:
				tb.WriteChar(c, i)
			default:
				return nil, newLexicalError(i, text, fmt.Sprintf("invalid quotation state: %d", quotation))
			}

		case '\\':
			if quotation == lexicalQuotationNone {
				if i == texLen-1 {
					return nil, newLexicalError(i, text, "iscape character at the end of the text")
				}
			} else {
				if i+1 < texLen {
					switch c = text[i+1]; c {
					case 'a':
						tb.WriteString("\a", i)
					case 'b':
						tb.WriteString("\b", i)
					case '$':
						tb.WriteString("$", i)
					case 'n', '\n':
						tb.WriteString("\n", i)
					case 'r', '\r':
						tb.WriteString("\r", i)
					case 't':
						tb.WriteString("\t", i)
					case 'v':
						tb.WriteString("\v", i)
					case 'f':
						tb.WriteString("\f", i)
					case '\\':
						tb.WriteString("\\", i)
					case '"':
						tb.WriteString("\"", i)
					case '\'':
						tb.WriteString("'", i)
					case '0':
						tb.WriteString("\x00", i)
					case ';':
						tb.WriteString(";", i)
					case '&':
						tb.WriteString("&", i)
					case '|':
						tb.WriteString("|", i)
					case '>':
						tb.WriteString(">", i)
					case '<':
						tb.WriteString("<", i)
					default:
						tb.WriteString(fmt.Sprintf("\\%c", c), i)
					}
					i++
				} else {
					return nil, newLexicalError(i, text, "escape character at the end of the text")
				}
			}

		case ';':
			if quotation == lexicalQuotationNone {
				if tb.IsPresent() {
					tokens = append(tokens, tb.Build())
				}
				tokens = append(tokens, LexicalToken{Kind: LexicalStop, Index: i})
			} else {
				tb.WriteChar(c, i)
			}

		case '&':
			if quotation == lexicalQuotationNone {
				if tb.IsPresent() {
					tokens = append(tokens, tb.Build())
				}
				if i+1 < texLen && text[i+1] == '&' {
					// &&
					tokens = append(tokens, LexicalToken{Kind: LexicalAnd, Index: i})
					i++
					break
				} else if i+1 < texLen && text[i+1] == '>' {
					if i+2 < texLen && text[i+2] == '>' {
						// &>>
						tokens = append(tokens, LexicalToken{Kind: LexicalFileAppendStdoutAndStderr, Index: i})
						i += 2
						break
					} else {
						// &>
						tokens = append(tokens, LexicalToken{Kind: LexicalFileStdoutAndStderr, Index: i})
						i++
						break
					}
				} else {
					// &
					tokens = append(tokens, LexicalToken{Kind: LexicalBackground, Index: i})
					break
				}
			} else {
				tb.WriteChar(c, i)
			}

		case '|':
			if quotation == lexicalQuotationNone {
				if tb.IsPresent() {
					tokens = append(tokens, tb.Build())
				}
				if i+1 < texLen && text[i+1] == '|' {
					tokens = append(tokens, LexicalToken{Kind: LexicalOr, Index: i})
					i++
					break
				} else {
					// |
					tokens = append(tokens, LexicalToken{Kind: LexicalPipeStdout, Index: i})
					break
				}
			} else {
				tb.WriteChar(c, i)
			}

		case '>':
			if quotation == lexicalQuotationNone {
				if tb.IsPresent() {
					tokens = append(tokens, tb.Build())
				}
				if i+1 < texLen && text[i+1] == '>' {
					// >>
					tokens = append(tokens, LexicalToken{Kind: LexicalFileAppendStdout, Index: i})
					i++
					break
				} else {
					// >
					tokens = append(tokens, LexicalToken{Kind: LexicalFileStdout, Index: i})
					break
				}
			}
			tb.WriteChar(c, i)

		case '<':
			if quotation == lexicalQuotationNone {
				if tb.IsPresent() {
					tokens = append(tokens, tb.Build())
				}
				if i+1 < texLen && text[i+1] == '<' {
					// <<
					tb.SetIndexIfEmpty(i)
					tb.SetKind(LexicalHereDocument)
					// get the here document name
					docNameB := strings.Builder{}
					for i += 2; i < texLen; i++ {
						c := text[i]
						if c == '\n' {
							break
						}
						docNameB.WriteByte(c)
					}
					docName := strings.TrimSpace(docNameB.String())
					// get the here document content
					err := func() error {
						var lineContentBuilder strings.Builder
						for i += 1; i < texLen; i++ {
							switch c := text[i]; c {
							case '\n':
								lineContent := lineContentBuilder.String()
								if strings.TrimSpace(lineContent) == docName {
									i--
									return nil
								} else {
									lineContentBuilder.Reset()
									tb.WriteString(lineContent, i)
									tb.WriteRune('\n', i)
								}
							case '\r':
								continue
							default:
								lineContentBuilder.WriteByte(c)
							}
						}
						lineContent := lineContentBuilder.String()
						if strings.TrimSpace(lineContent) == docName {
							return nil
						}
						return newLexicalError(tb.Index, text, "here document not closed")
					}()

					if err != nil {
						return nil, err
					}

					t := tb.Build()
					t.Content = strings.Trim(t.Content, "\n") + "\n"
					tokens = append(tokens, t)
					break
				} else {
					// <
					tokens = append(tokens, LexicalToken{Kind: LexicalRedirectStdin, Index: i})
					break
				}
			}
			tb.WriteChar(c, i)

		case '1':
			if tb.IsPresent() {
				tb.WriteChar(c, i)
				break
			}
			if i+1 < texLen && text[i+1] == '>' {
				if i+2 < texLen && text[i+2] == '&' {
					if i+3 < texLen && text[i+3] == '2' {
						// 1>&2
						tokens = append(tokens, LexicalToken{Kind: LexicalStdoutToStderr, Index: i})
						i += 3
						break
					}
				}
			}
			tb.WriteChar(c, i)

		case '2':
			if tb.IsPresent() {
				tb.WriteChar(c, i)
				break
			}
			if i+1 < texLen && text[i+1] == '>' {
				if i+2 < texLen && text[i+2] == '&' {
					if i+3 < texLen && text[i+3] == '1' {
						// 2>&1
						tokens = append(tokens, LexicalToken{Kind: LexicalStderrToStdout, Index: i})
						i += 3
						break
					}
				} else if i+2 < texLen && text[i+2] == '>' {
					// 2>>
					tokens = append(tokens, LexicalToken{Kind: LexicalFileAppendStderr, Index: i})
					i += 2
					break
				} else {
					// 2>
					tokens = append(tokens, LexicalToken{Kind: LexicalFileStderr, Index: i})
					i++
					break
				}
			}
			tb.WriteChar(c, i)

		default:
			tb.WriteChar(c, i)
		}
	}
	if quotation != lexicalQuotationNone {
		return nil, newLexicalError(len(text)-1, text, "quotation not closed")
	}
	if tb.IsPresent() {
		tokens = append(tokens, tb.Build())
	}
	// trim trailing LexicalStop tokens
	for len(tokens) > 0 && tokens[len(tokens)-1].Kind == LexicalStop {
		tokens = tokens[:len(tokens)-1]
	}
	return tokens, nil
}
