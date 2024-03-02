package compiler

import "fmt"

type CompilerErrorKind uint8

func (k CompilerErrorKind) String() string {
	switch k {
	case CompilerErrorLexical:
		return "lexical"
	case CompilerErrorParser:
		return "parser"
	default:
		panic(fmt.Sprintf("unknown CompilerErrorKind: %d", k))
	}
}

const (
	CompilerErrorLexical CompilerErrorKind = iota
	CompilerErrorParser
)

type CompilerError struct {
	Line    int
	Column  int
	Comment string
	Kind    CompilerErrorKind
}

func (e CompilerError) Error() string {
	return fmt.Sprintf("%s error at line %d, column %d: %s", e.Kind.String(), e.Line, e.Column, e.Comment)
}

func (e *CompilerError) Wrap(kind CompilerErrorKind, comment string) *CompilerError {
	e.Comment = fmt.Sprintf("%s\n%s error: %s", comment, e.Kind.String(), e.Comment)
	e.Kind = kind
	return e
}

func newLexicalError(index int, text string, comment string) CompilerError {
	line, column := 1, 1
	for i, c := range text {
		if i == index {
			break
		}
		if c == '\n' {
			line++
			column = 1
		} else {
			column++
		}
	}
	return CompilerError{line, column, comment, CompilerErrorLexical}
}

func newParserError(index int, text string, comment string) CompilerError {
	line, column := 1, 1
	for i, c := range text {
		if i == index {
			break
		}
		if c == '\n' {
			line++
			column = 1
		} else {
			column++
		}
	}
	return CompilerError{line, column, comment, CompilerErrorParser}
}
