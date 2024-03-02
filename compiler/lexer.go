package compiler

import (
	"fmt"
	"io"
	"os"
	"strings"
)

type LexicalTokenKind uint8

func (k LexicalTokenKind) String() string {
	return fmt.Sprintf("LexicalTokenKind(%d)", k)
}

const (
	// program name, command, argument, variable
	LexicalTokenIdentifier LexicalTokenKind = iota
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

type LexicalToken struct {
	Content string
	// Index is the position of the token in the text.
	// This is not 100% accurate because it doesn't take quotation into account.
	// It's good enough for error messages.
	// Index is always the position of the first character of the token or within the token (in case of quotation).
	Index int
	Kind  LexicalTokenKind
}

type lexicalQuotation uint8

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
	name := strings.Builder{}
	quotation := lexicalQuotationNone

	for i := 0; i < texLen; i++ {
		c := text[i]
		switch c {

		case '\n':
			if quotation != lexicalQuotationNone {
				return nil, newLexicalError(i, text, "quotation not closed at the end of the line")
			}
			tokens = append(tokens, LexicalToken{Kind: LexicalStop, Index: i})

		case '\r':
			if quotation != lexicalQuotationNone {
				name.WriteByte(c)
			}
			continue

		case ' ', '\t', '\v', '\f', 20:
			if quotation == lexicalQuotationNone {
				if name.Len() != 0 {
					n := name.String()
					tokens = append(tokens, LexicalToken{Kind: LexicalTokenIdentifier, Content: strings.TrimSpace(n), Index: i - len(n)})
					name.Reset()
				}
			} else {
				name.WriteByte(c)
			}

		case '$':
			if i+1 < texLen && text[i+1] == '(' {
				// subshell
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
				iop, r := subshellIoProvider(iop)
				defer iop.Close()
				if err := Execute(subshell.String(), iop); err != nil {
					return nil, newLexicalError(i, text, fmt.Sprintf("failed to execute subshell: %v", err))
				}
				subst := new(strings.Builder)
				if _, err := io.Copy(subst, r); err != nil {
					return nil, newLexicalError(i, text, fmt.Sprintf("failed to read subshell output: %v", err))
				}
				name.WriteString(strings.TrimSpace(subst.String()))
			} else {
				// variable
				startIndex := i - name.Len()
				varName := strings.Builder{}
				for i += 1; i < texLen; i++ {
					c := text[i]
					if c == ' ' || c == ';' || c == '\t' || c == '\v' || c == '\f' || c == '\n' || c == '\r' || c == '.' || c == ',' || c == '/' || c == '>' || c == '<' || c == '&' || c == '|' {
						i--
						break
					}
					varName.WriteByte(c)
				}
				name.WriteString(os.Getenv(varName.String()))
				if quotation == lexicalQuotationNone {
					token := LexicalToken{Kind: LexicalTokenIdentifier, Content: name.String(), Index: startIndex}
					tokens = append(tokens, token)
					name.Reset()
				}
			}

		case '"':
			switch quotation {
			case lexicalQuotationNone:
				quotation = lexicalQuotationDouble
			case lexicalQuotationDouble:
				quotation = lexicalQuotationNone
			case lexicalQuotationSingle:
				name.WriteByte(c)
			default:
				return nil, newLexicalError(i, text, fmt.Sprintf("invalid quotation state: %d", quotation))
			}

		case '\'':
			switch quotation {
			case lexicalQuotationNone:
				quotation = lexicalQuotationSingle
			case lexicalQuotationSingle:
				quotation = lexicalQuotationNone
			case lexicalQuotationDouble:
				name.WriteByte(c)
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
					i++
					c = text[i]
					switch c {
					case 'a':
						name.WriteString("\a")
					case 'b':
						name.WriteString("\b")
					case '$':
						name.WriteString("$")
					case 'n', '\n':
						name.WriteString("\n")
					case 'r', '\r':
						name.WriteString("\r")
					case 't':
						name.WriteString("\t")
					case 'v':
						name.WriteString("\v")
					case 'f':
						name.WriteString("\f")
					case '\\':
						name.WriteString("\\")
					case '"':
						name.WriteString("\"")
					case '\'':
						name.WriteString("'")
					case '0':
						name.WriteString("\x00")
					case ';':
						name.WriteString(";")
					case '&':
						name.WriteString("&")
					case '|':
						name.WriteString("|")
					case '>':
						name.WriteString(">")
					case '<':
						name.WriteString("<")
					default:
						name.WriteString(fmt.Sprintf("\\%c", c))
					}
				} else {
					return nil, newLexicalError(i, text, "escape character at the end of the text")
				}
			}

		case ';':
			if quotation == lexicalQuotationNone {
				if name.Len() != 0 {
					n := name.String()
					tokens = append(tokens, LexicalToken{Kind: LexicalTokenIdentifier, Content: strings.TrimSpace(n), Index: i - len(n)})
					name.Reset()
				}
				tokens = append(tokens, LexicalToken{Kind: LexicalStop, Index: i})
			} else {
				name.WriteByte(c)
			}

		case '&':
			if quotation == lexicalQuotationNone {
				if name.Len() != 0 {
					n := name.String()
					tokens = append(tokens, LexicalToken{Kind: LexicalTokenIdentifier, Content: strings.TrimSpace(n), Index: i - len(n)})
					name.Reset()
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
				name.WriteByte(c)
			}

		case '|':
			if quotation == lexicalQuotationNone {
				if name.Len() != 0 {
					n := name.String()
					tokens = append(tokens, LexicalToken{Kind: LexicalTokenIdentifier, Content: strings.TrimSpace(n), Index: i - len(n)})
					name.Reset()
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
				name.WriteByte(c)
			}

		case '>':
			if quotation == lexicalQuotationNone {
				if name.Len() != 0 {
					n := name.String()
					tokens = append(tokens, LexicalToken{Kind: LexicalTokenIdentifier, Content: strings.TrimSpace(n), Index: i - len(n)})
					name.Reset()
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
			name.WriteByte(c)

		case '<':
			if quotation == lexicalQuotationNone {
				if name.Len() != 0 {
					n := name.String()
					tokens = append(tokens, LexicalToken{Kind: LexicalTokenIdentifier, Content: strings.TrimSpace(n), Index: i - len(n)})
					name.Reset()
				}
				if i+1 < texLen && text[i+1] == '<' {
					// <<
					t := LexicalToken{Kind: LexicalHereDocument, Index: i}
					// get the here document name
					var docName string
					for i += 2; i < texLen; i++ {
						c := text[i]
						if c == '\n' {
							break
						}
						docName += string(c)
					}
					docName = strings.TrimSpace(docName)
					// get the here document content
					docContent, err := func() (string, error) {
						var docContent strings.Builder
						for i += 1; i < texLen; i++ {
							c := text[i]
							if c == '\n' {
								// check if full line content is equal to the here document name
								var lineContentBuilder strings.Builder
								for j := i + 1; j < texLen; j++ {
									c := text[j]
									if c == '\n' {
										break
									}
									lineContentBuilder.WriteByte(c)
								}
								lineContent := strings.TrimSpace(lineContentBuilder.String())
								if lineContent == docName {
									i += len(docName)
									return docContent.String(), nil
								}
							}
							docContent.WriteByte(c)
						}
						return "", newLexicalError(i, text, "here document not closed")
					}()

					if err != nil {
						return nil, err
					}

					t.Content = strings.TrimSpace(docContent)
					tokens = append(tokens, t)
					break
				} else {
					// <
					tokens = append(tokens, LexicalToken{Kind: LexicalRedirectStdin, Index: i})
					break
				}
			}
			name.WriteByte(c)

		case '1':
			if name.Len() != 0 {
				name.WriteByte(c)
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
			name.WriteByte(c)

		case '2':
			if name.Len() != 0 {
				name.WriteByte(c)
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
			name.WriteByte(c)

		default:
			name.WriteByte(c)
		}
	}
	if quotation != lexicalQuotationNone {
		return nil, newLexicalError(len(text)-1, text, "quotation not closed")
	}
	if name.Len() != 0 {
		n := name.String()
		tokens = append(tokens, LexicalToken{Kind: LexicalTokenIdentifier, Content: strings.TrimSpace(n), Index: texLen - len(n)})
	}
	// trim trailing LexicalStop tokens
	for tokens[len(tokens)-1].Kind == LexicalStop {
		tokens = tokens[:len(tokens)-1]
	}
	return tokens, nil
}
