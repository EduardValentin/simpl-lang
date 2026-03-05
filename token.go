package simpl

type Position struct {
	Line   int
	Column int
}

type TokenType int

const (
	TokenEOF TokenType = iota
	TokenIdentifier
	TokenInt
	TokenFloat
	TokenString
	TokenBool

	// Keywords
	TokenVar
	TokenConst
	TokenRead
	TokenWrite
	TokenIf
	TokenElse
	TokenFor
	TokenFrom
	TokenUntil
	TokenStep
	TokenWhile
	TokenTypeInt
	TokenTypeFloat
	TokenTypeBool
	TokenTypeString
	TokenTypeArray

	// Punctuation
	TokenLBrace
	TokenRBrace
	TokenLBracket
	TokenRBracket
	TokenLParen
	TokenRParen
	TokenComma

	// Operators
	TokenAssign   // =
	TokenEqual    // ==
	TokenNotEqual // !=
	TokenGreater  // >
	TokenGreaterEq
	TokenLess
	TokenLessEq
	TokenPlus
	TokenMinus
	TokenStar
	TokenSlash
	TokenPercent
	TokenBang
)

func (t TokenType) String() string {
	switch t {
	case TokenEOF:
		return "EOF"
	case TokenIdentifier:
		return "IDENTIFIER"
	case TokenInt:
		return "INT"
	case TokenFloat:
		return "FLOAT"
	case TokenString:
		return "STRING"
	case TokenBool:
		return "BOOL"
	case TokenVar:
		return "var"
	case TokenConst:
		return "const"
	case TokenRead:
		return "read"
	case TokenWrite:
		return "write"
	case TokenIf:
		return "if"
	case TokenElse:
		return "else"
	case TokenFor:
		return "for"
	case TokenFrom:
		return "from"
	case TokenUntil:
		return "until"
	case TokenStep:
		return "step"
	case TokenWhile:
		return "while"
	case TokenTypeInt:
		return "int"
	case TokenTypeFloat:
		return "float"
	case TokenTypeBool:
		return "bool"
	case TokenTypeString:
		return "string"
	case TokenTypeArray:
		return "array"
	case TokenLBrace:
		return "{"
	case TokenRBrace:
		return "}"
	case TokenLBracket:
		return "["
	case TokenRBracket:
		return "]"
	case TokenLParen:
		return "("
	case TokenRParen:
		return ")"
	case TokenComma:
		return ","
	case TokenAssign:
		return "="
	case TokenEqual:
		return "=="
	case TokenNotEqual:
		return "!="
	case TokenGreater:
		return ">"
	case TokenGreaterEq:
		return ">="
	case TokenLess:
		return "<"
	case TokenLessEq:
		return "<="
	case TokenPlus:
		return "+"
	case TokenMinus:
		return "-"
	case TokenStar:
		return "*"
	case TokenSlash:
		return "/"
	case TokenPercent:
		return "%"
	case TokenBang:
		return "!"
	default:
		return "UNKNOWN"
	}
}

type Token struct {
	Type    TokenType
	Lexeme  string
	Literal any
	Pos     Position
}

var keywordTokens = map[string]TokenType{
	"var":    TokenVar,
	"const":  TokenConst,
	"read":   TokenRead,
	"write":  TokenWrite,
	"if":     TokenIf,
	"else":   TokenElse,
	"for":    TokenFor,
	"from":   TokenFrom,
	"until":  TokenUntil,
	"step":   TokenStep,
	"while":  TokenWhile,
	"int":    TokenTypeInt,
	"float":  TokenTypeFloat,
	"bool":   TokenTypeBool,
	"string": TokenTypeString,
	"array":  TokenTypeArray,
	"true":   TokenBool,
	"false":  TokenBool,
}
