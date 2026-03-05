package simpl

import (
	"strconv"
	"unicode"
)

type scanner struct {
	source    []rune
	start     int
	current   int
	line      int
	column    int
	startLine int
	startCol  int
	tokens    []Token
	diags     []Diagnostic
}

func lexSource(source string) ([]Token, []Diagnostic) {
	s := &scanner{
		source: []rune(source),
		line:   1,
		column: 1,
		tokens: make([]Token, 0, 64),
		diags:  make([]Diagnostic, 0, 8),
	}

	for !s.isAtEnd() {
		s.start = s.current
		s.startLine = s.line
		s.startCol = s.column
		s.scanToken()
	}

	s.tokens = append(s.tokens, Token{
		Type:   TokenEOF,
		Lexeme: "",
		Pos: Position{
			Line:   s.line,
			Column: s.column,
		},
	})

	return s.tokens, s.diags
}

func (s *scanner) isAtEnd() bool {
	return s.current >= len(s.source)
}

func (s *scanner) advance() rune {
	r := s.source[s.current]
	s.current++
	if r == '\n' {
		s.line++
		s.column = 1
	} else {
		s.column++
	}
	return r
}

func (s *scanner) peek() rune {
	if s.isAtEnd() {
		return 0
	}
	return s.source[s.current]
}

func (s *scanner) peekNext() rune {
	if s.current+1 >= len(s.source) {
		return 0
	}
	return s.source[s.current+1]
}

func (s *scanner) match(expected rune) bool {
	if s.isAtEnd() || s.source[s.current] != expected {
		return false
	}
	s.advance()
	return true
}

func (s *scanner) addToken(tt TokenType, literal any) {
	lexeme := string(s.source[s.start:s.current])
	s.tokens = append(s.tokens, Token{
		Type:    tt,
		Lexeme:  lexeme,
		Literal: literal,
		Pos: Position{
			Line:   s.startLine,
			Column: s.startCol,
		},
	})
}

func (s *scanner) addDiag(code, message, hint string, line, col int) {
	s.diags = append(s.diags, newDiagnostic(code, "lexer", message, Position{Line: line, Column: col}, hint))
}

func (s *scanner) scanToken() {
	c := s.advance()
	switch c {
	case ' ', '\r', '\t', '\n':
		return
	case '{':
		s.addToken(TokenLBrace, nil)
	case '}':
		s.addToken(TokenRBrace, nil)
	case '[':
		s.addToken(TokenLBracket, nil)
	case ']':
		s.addToken(TokenRBracket, nil)
	case '(':
		s.addToken(TokenLParen, nil)
	case ')':
		s.addToken(TokenRParen, nil)
	case ',':
		s.addToken(TokenComma, nil)
	case '+':
		s.addToken(TokenPlus, nil)
	case '-':
		s.addToken(TokenMinus, nil)
	case '*':
		s.addToken(TokenStar, nil)
	case '%':
		s.addToken(TokenPercent, nil)
	case '/':
		if s.match('/') {
			for s.peek() != '\n' && !s.isAtEnd() {
				s.advance()
			}
			return
		}
		s.addToken(TokenSlash, nil)
	case '!':
		if s.match('=') {
			s.addToken(TokenNotEqual, nil)
		} else {
			s.addToken(TokenBang, nil)
		}
	case '=':
		if s.match('=') {
			s.addToken(TokenEqual, nil)
		} else {
			s.addToken(TokenAssign, nil)
		}
	case '>':
		if s.match('=') {
			s.addToken(TokenGreaterEq, nil)
		} else {
			s.addToken(TokenGreater, nil)
		}
	case '<':
		if s.match('=') {
			s.addToken(TokenLessEq, nil)
		} else {
			s.addToken(TokenLess, nil)
		}
	case '"':
		s.scanString()
	default:
		if unicode.IsDigit(c) {
			s.scanNumber()
			return
		}
		if isIdentifierStart(c) {
			s.scanIdentifier()
			return
		}
		s.addDiag("LEX_UNKNOWN_CHAR", "Unknown character in source.", "Remove or replace unsupported character.", s.startLine, s.startCol)
	}
}

func (s *scanner) scanString() {
	startLine := s.startLine
	startCol := s.startCol
	for !s.isAtEnd() && s.peek() != '"' {
		if s.peek() == '\n' {
			s.addDiag("LEX_UNTERMINATED_STRING", "String literal must end on the same line.", "Add a closing quote before the line ends.", startLine, startCol)
			return
		}
		s.advance()
	}

	if s.isAtEnd() {
		s.addDiag("LEX_UNTERMINATED_STRING", "String literal is not terminated.", "Add a closing quote.", startLine, startCol)
		return
	}

	// Consume closing quote.
	s.advance()
	raw := string(s.source[s.start:s.current])
	parsed, err := strconv.Unquote(raw)
	if err != nil {
		s.addDiag("LEX_UNTERMINATED_STRING", "Invalid string literal.", "Use valid quotes and escapes.", startLine, startCol)
		return
	}
	s.addToken(TokenString, parsed)
}

func (s *scanner) scanNumber() {
	for unicode.IsDigit(s.peek()) {
		s.advance()
	}

	isFloat := false
	if s.peek() == '.' && unicode.IsDigit(s.peekNext()) {
		isFloat = true
		s.advance()
		for unicode.IsDigit(s.peek()) {
			s.advance()
		}
	}

	lexeme := string(s.source[s.start:s.current])
	if isFloat {
		v, err := strconv.ParseFloat(lexeme, 64)
		if err != nil {
			s.addDiag("LEX_INVALID_FLOAT", "Invalid float literal.", "Check the float format.", s.startLine, s.startCol)
			return
		}
		s.addToken(TokenFloat, v)
		return
	}

	v, err := strconv.ParseInt(lexeme, 10, 64)
	if err != nil {
		s.addDiag("LEX_INVALID_INT", "Invalid integer literal.", "Check the integer format.", s.startLine, s.startCol)
		return
	}
	s.addToken(TokenInt, v)
}

func (s *scanner) scanIdentifier() {
	for isIdentifierPart(s.peek()) {
		s.advance()
	}

	lexeme := string(s.source[s.start:s.current])
	tok, ok := keywordTokens[lexeme]
	if ok {
		if tok == TokenBool {
			s.addToken(TokenBool, lexeme == "true")
			return
		}
		s.addToken(tok, nil)
		return
	}
	s.addToken(TokenIdentifier, lexeme)
}

func isIdentifierStart(r rune) bool {
	return r == '_' || unicode.IsLetter(r)
}

func isIdentifierPart(r rune) bool {
	return isIdentifierStart(r) || unicode.IsDigit(r)
}
