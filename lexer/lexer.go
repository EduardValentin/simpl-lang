package lexer

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Token int

const (
	UNKNOWN Token = iota
	VAR
	CONST
	IDENTIFIER
	ASIGNMENT
	EQ
	IF_STMT
	FOR_LOOP
	WHILE_LOOP
	FROM
	TO
	STEP
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
	ADD
	SUB
	MUL
	DIV
	MOD
	GT
	LT
	GTEQ
	LTEQ
	NEQ
	NOT
)

var TOKENS_STR = map[Token]string{
	INT_TYPE:   "int",
	FLOAT_TYPE: "float",
	STR_TYPE:   "string",
	BOOL_TYPE:  "bool",
}

var KEYWORDS = map[string]Token{
	"var":    VAR,
	"const":  CONST,
	"int":    INT_TYPE,
	"float":  FLOAT_TYPE,
	"bool":   BOOL_TYPE,
	"array":  ARRAY_TYPE,
	"string": STR_TYPE,
}

func intParser(item string) (any, error) {
	return strconv.ParseInt(item, 10, 64)
}

func floatParser(item string) (any, error) {
	return strconv.ParseFloat(item, 64)
}

func boolParser(item string) (any, error) {
	return strconv.ParseBool(item)
}

func stringParser(item string) (any, error) {
	runes := []rune(item)
	if runes[0] != '"' || runes[len(runes)-1] != '"' {
		return nil, errors.New("Array not properly formatted")
	}

	return strings.Trim(item, "\""), nil
}

var literalParsersByType = map[Token]func(string) (any, error){
	INT_TYPE:   intParser,
	FLOAT_TYPE: floatParser,
	STR_TYPE:   stringParser,
	BOOL_TYPE:  boolParser,
}

type Lexeme struct {
	token Token
	value any
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

func (l *Lexer) currentChar() rune {
	return rune(l.source[l.current])
}

func (l *Lexer) matchUntillFromStartPos(delimiters string, startPos int) (string, rune, error) {
	var delimiterFound rune

	for l.current < len(l.source) {
		currentChar := l.source[l.current]
		if strings.Contains(delimiters, string(currentChar)) {
			delimiterFound = rune(currentChar)
			break
		}
		l.advance()
	}

	if l.current >= len(l.source) {
		return "", delimiterFound, errors.New("Delimiter not found")
	}

	l.advance()
	substr := l.source[startPos:l.current]

	return substr, delimiterFound, nil
}

func (l *Lexer) matchUntil(delimiters string) (string, rune, error) {
	return l.matchUntillFromStartPos(delimiters, l.start)
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
			items := l.arrayLiteral()
			l.lexemes = append(l.lexemes, Lexeme{token: ARRAY, value: items, line: l.line})
		case '{':
			l.lexemes = append(l.lexemes, Lexeme{token: SQUIRLY_OPEN, line: l.line})
		case '}':
			l.lexemes = append(l.lexemes, Lexeme{token: SQUIRLY_CLOSE, line: l.line})
		case '=':
			l.lexemes = append(l.lexemes, Lexeme{token: ASIGNMENT, line: l.line})
		case '+':
			l.lexemes = append(l.lexemes, Lexeme{token: ADD, line: l.line})
		case '-':
			l.lexemes = append(l.lexemes, Lexeme{token: SUB, line: l.line})
		case '*':
			l.lexemes = append(l.lexemes, Lexeme{token: MUL, line: l.line})
		case '/':
			l.lexemes = append(l.lexemes, Lexeme{token: DIV, line: l.line})
		case '%':
			l.lexemes = append(l.lexemes, Lexeme{token: MOD, line: l.line})
		case '"':
			value, _, err := l.matchUntil("\"")
			if err != nil {
				panic("String not properly ended")
			}
			l.lexemes = append(l.lexemes, Lexeme{token: STR, value: value, line: l.line})
		case 'i':
			if l.peek("f") {
				l.lexemes = append(l.lexemes, Lexeme{token: IF_STMT, line: l.line})
				l.current += 1
				continue
			}
			fallthrough
		case 'f':
			if l.peek("or") {
				l.lexemes = append(l.lexemes, Lexeme{token: FOR_LOOP, line: l.line})
				l.current += 2
				continue
			} else if l.peek("rom") {
				l.lexemes = append(l.lexemes, Lexeme{token: FROM, line: l.line})
				l.current += 3
				continue
			}
			fallthrough
		case 't':
			if l.peek("o") {
				l.lexemes = append(l.lexemes, Lexeme{token: TO, line: l.line})
				l.current += 1
				continue
			}
			fallthrough
		case 's':
			if l.peek("tep") {
				l.lexemes = append(l.lexemes, Lexeme{token: STEP, line: l.line})
				l.current += 3
				continue
			}
			fallthrough
		case 'w':
			if l.peek("hile") {
				l.lexemes = append(l.lexemes, Lexeme{token: WHILE_LOOP, line: l.line})
				l.current += 4
				continue
			}
			fallthrough
		default:
			if isAlphaNum(rune(l.source[l.start])) {
				for l.current < len(l.source) && isAlphaNum(rune(l.source[l.current])) {
					l.advance()
				}
			}

			substr := l.source[l.start:l.current]

			if v, err := strconv.ParseInt(substr, 10, 64); err == nil {
				l.lexemes = append(l.lexemes, Lexeme{token: INT, value: v, line: l.line})
			} else if v, err := strconv.ParseFloat(substr, 64); err == nil {
				l.lexemes = append(l.lexemes, Lexeme{token: FLOAT, value: v, line: l.line})
			} else if v, err := strconv.ParseBool(substr); err == nil {
				l.lexemes = append(l.lexemes, Lexeme{token: BOOL, value: v, line: l.line})
			} else if t, isKeyword := KEYWORDS[substr]; isKeyword == true {
				l.lexemes = append(l.lexemes, Lexeme{token: t, value: TOKENS_STR[t], line: l.line})
				if t == ARRAY_TYPE {
					l.array()
				}
			} else if isAlphaNumStr(substr) {
				l.lexemes = append(l.lexemes, Lexeme{token: IDENTIFIER, value: substr, line: l.line})
			} else {
				panic("Unknown character parsed: " + string(l.source[l.current]))
			}
		}
	}
}

func (l *Lexer) peek(match string) bool {
	for i, v := range match {
		if rune(l.source[l.current+i]) != v {
			return false
		}
	}
	return true
}

func isAlphaNumStr(str string) bool {
	for _, chr := range str {
		if !isAlphaNum(chr) {
			return false
		}
	}
	return true
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
	for l.current < len(l.source) && rune(l.source[l.current]) != ']' && rune(l.source[l.current]) != '[' {
		l.advance()
	}

	if l.current >= len(l.source) {
		panic("Syntax error: Array not properly ended at line: " + fmt.Sprint(l.line))
	}

	substr := l.source[idx:l.current]

	if substr == "int" || substr == "float" || substr == "string" || substr == "bool" {
		l.lexemes = append(l.lexemes, Lexeme{token: KEYWORDS[substr], value: TOKENS_STR[KEYWORDS[substr]], line: l.line})
	} else if substr == "array" {
		l.array()
	}

	l.lexemes = append(l.lexemes, Lexeme{token: BRACKET_CLOSE, line: l.line})
	l.advance()
}

func (l *Lexer) arrayLiteral() []any {
	items := make([]any, 0, 5)

	var literalEnded bool = false

	for literalEnded == false {
		item, delimiterFound, err := l.matchUntillFromStartPos(",]", l.current)
		item = strings.Trim(item, " ,]")

		if err != nil {
			panic("Error: Array literal not properly formatted at line: " + fmt.Sprint(l.line))
		}

		items = append(items, item)

		if delimiterFound == ']' {
			literalEnded = true
		}
	}

	if !literalEnded {
		panic("Error: Array literal not properly ended")
	}

	var arrayType Token
	lastLexemeIdx := len(l.lexemes) - 1

	if l.lexemes[lastLexemeIdx].token == ASIGNMENT && l.lexemes[lastLexemeIdx-1].token == BRACKET_CLOSE {
		arrayType = l.lexemes[lastLexemeIdx-2].token
	} else if l.lexemes[lastLexemeIdx].token == ASIGNMENT && l.lexemes[lastLexemeIdx-1].token == IDENTIFIER {
		identifier := l.lexemes[lastLexemeIdx-1].value

		_, idx, err := l.findTypeForIdentifier(identifier.(string))
		if err != nil {
			panic(err)
		}
		arrayType = l.lexemes[idx+2].token
	}

	for i, item := range items {
		parsed, err := literalParsersByType[arrayType](item.(string))
		if err == nil {
			items[i] = parsed
		} else {
			panic(err)
		}
	}

	return items
}

func isAlphaNum(chr rune) bool {
	return chr == '_' || chr == '.' || (chr >= 'a' && chr <= 'z') || (chr >= '0' && chr <= '9')
}

func (l *Lexer) findTypeForIdentifier(identifier string) (Token, int, error) {
	for i, val := range l.lexemes {
		if val.token == IDENTIFIER && val.value == identifier {
			return l.lexemes[i+1].token, i + 1, nil
		}
	}
	return UNKNOWN, -1, errors.New("Could not find type for identifier")
}
