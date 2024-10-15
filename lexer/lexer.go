package lexer

import (
	"fmt"
	"strconv"
)

type Token int

const (
	VAR Token = iota
	CONST
	IDENTIFIER
	ASIGNMENT
	EQ

	IF_STMT
	FOR_STMT

	STR
	INT
	BOOL
	FLOAT
	ARRAY

	STR_TYPE
	INT_TYPE
	BOOL_TYPE
	FLOAT_TYPE
	ARRAY_TYPE

	PARAN_OPEN
	PARAN_CLOSE
	SQUIRLY_OPEN
	SQUIRLY_CLOSE
	BRACKET_OPEN
	BRACKET_CLOSE
)

var KEYWORDS = map[string]Token{
	"var":    VAR,
	"const":  CONST,
	"int":    INT_TYPE,
	"float":  FLOAT_TYPE,
	"bool":   BOOL_TYPE,
	"array":  ARRAY_TYPE,
	"string": STR_TYPE,
	"if":     IF_STMT,
	"for":    FOR_STMT,
}

type Lexeme struct {
	token Token
	value string
	line  int
}

type Lexer struct {
	start   int
	current int
	source  string
	lexemes []Lexeme
	line    int
}

func Init(source string) *Lexer {

	return &Lexer{
		start:   0,
		current: 0,
		source:  source,
		lexemes: make([]Lexeme, 0),
	}
}

func (l *Lexer) peek() rune {
	return rune(l.source[l.current])
}

func (l *Lexer) matchUntil(delimiter rune) string {
	idx := l.current - 1
	for idx < len(l.source) && rune(l.source[idx]) != delimiter {
		idx++
	}

	if idx >= len(l.source) {
		panic("Syntax error: closing delimiter not found at line: " + fmt.Sprint(l.line))
	}

	idx++
	substr := l.source[l.current:idx]
	l.current = idx

	return substr
}

func (l *Lexer) Scan() {
	l.line = 0
	for l.current < len(l.source) {
		l.start = l.current
		switch l.advance() {
		case '\n':
			l.line++
		case '\r', '\t', ' ':
		case '[':
			l.matchUntil(']')
			l.lexemes = append(l.lexemes, Lexeme{token: ARRAY, line: l.line})
		case '=':
			l.lexemes = append(l.lexemes, Lexeme{token: ASIGNMENT, line: l.line})
		case '"':
			l.matchUntil('"')
			l.lexemes = append(l.lexemes, Lexeme{token: STR, line: l.line})
		default:

			idx := l.current - 1
			for idx < len(l.source) && isAlphaNum(rune(l.source[idx])) {
				idx++
			}

			substr := l.source[l.start:idx]
			l.current = idx

			if isInteger(substr) {
				l.lexemes = append(l.lexemes, Lexeme{token: INT, line: l.line})
			} else if isFloat(substr) {
				l.lexemes = append(l.lexemes, Lexeme{token: FLOAT, line: l.line})
			} else if isBool(substr) {
				l.lexemes = append(l.lexemes, Lexeme{token: BOOL, line: l.line})
			} else if t, isKeyword := KEYWORDS[substr]; isKeyword == true {
				l.lexemes = append(l.lexemes, Lexeme{token: t, line: l.line})
				if t == ARRAY_TYPE {
					l.array()
				}
			} else {
				l.lexemes = append(l.lexemes, Lexeme{token: IDENTIFIER, line: l.line})
			}
		}
	}
}

func (l *Lexer) advance() rune {
	aux := l.current
	l.current++

	return rune(l.source[aux])
}

func (l *Lexer) array() {
	if l.advance() != '[' {
		panic("Syntax error: Array not properly declared at line: " + fmt.Sprint(l.line))
	}

	l.lexemes = append(l.lexemes, Lexeme{token: BRACKET_OPEN, line: l.line})

	idx := l.current
	for idx < len(l.source) && rune(l.source[idx]) != ']' {
		idx++
	}
	substr := l.source[l.current:idx]
	if substr == "int" || substr == "float" || substr == "string" || substr == "bool" {
		l.lexemes = append(l.lexemes, Lexeme{token: KEYWORDS[substr], line: l.line})
	}

	l.current = idx
	if l.advance() != ']' {
		panic("Syntax error: Array not properly declared at line: " + fmt.Sprint(l.line))
	}
	l.lexemes = append(l.lexemes, Lexeme{token: BRACKET_CLOSE, line: l.line})

}

func isInteger(str string) bool {
	_, err := strconv.ParseInt(str, 10, 64)
	return err == nil
}

func isFloat(str string) bool {
	_, err := strconv.ParseFloat(str, 64)
	return err == nil
}

func isBool(str string) bool {
	_, err := strconv.ParseBool(str)
	return err == nil
}

func isAlphaNum(chr rune) bool {
	return chr == '_' || chr == '.' || (chr >= 'a' && chr <= 'z') || (chr >= '0' && chr <= '9')
}
