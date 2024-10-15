package lexer

import (
	"testing"
)

//	func TestInstantiation(t *testing.T) {
//		lexer := Lexer{start: 0, current: 2, source: "Test source"}
//
//		if lexer.start != 0 || lexer.current != 2 || lexer.source != "Test source" {
//			t.Fatal("Wrong parameters after instantiation")
//		}
//	}
//
//	func TestInit(t *testing.T) {
//		lexer := Init("I'm your father")
//		if lexer.start != 0 || lexer.current != 0 || lexer.source != "I'm your father" {
//			t.Fatal("Wrong parameters after Init")
//		}
//	}
//
//	func TestScan(t *testing.T) {
//		lexer := Init("var a int")
//		lexer.Scan()
//	}
//
//	func MatchesVarDeclaration(t *testing.T) {
//		lexer := Init("var a int")
//		lexer.Scan()
//
//		if lexer.lexemes[0].token != VAR || lexer.lexemes[1].token != IDENTIFIER || lexer.lexemes[2].token != INT_TYPE {
//			t.Fatal("Lexemes not properly matched")
//		}
//	}
//
//	func TestMatchesVarDecWithIntAssignment(t *testing.T) {
//		lexer := Init("var v2 int = 2")
//		lexer.Scan()
//
//		expected := []Token{VAR, IDENTIFIER, INT_TYPE, ASIGNMENT, INT}
//		if !assertTokensMatch(lexer, expected) {
//			t.Fatalf("Lexemes not properly matched. Given: %v. Expected: %v", lexer.lexemes, expected)
//		}
//	}
func TestMatchesVarDecWithStringAssignment(t *testing.T) {
	lexer := Init("var v3 string = \"Hello\"")
	lexer.Scan()

	expected := []Token{VAR, IDENTIFIER, STR_TYPE, ASIGNMENT, STR}
	if !assertTokensMatch(lexer, expected) {
		t.Fatalf("Lexemes not properly matched. Given: %v. Expected: %v", lexer.lexemes, expected)
	}
}

//	func TestMatchesVarDecWithFloatAssignment(t *testing.T) {
//		lexer := Init("var v3 float = 2.156")
//		lexer.Scan()
//
//		expected := []Token{VAR, IDENTIFIER, FLOAT_TYPE, ASIGNMENT, FLOAT}
//		if !assertTokensMatch(lexer, expected) {
//			t.Fatalf("Lexemes not properly matched. Given: %v. Expected: %v", lexer.lexemes, expected)
//		}
//	}
//
//	func TestMatchesVarDecWithArrayOfStringAssignment(t *testing.T) {
//		lexer := Init(`var v3 array[string] = ["hello", "world"]`)
//		lexer.Scan()
//
//		expected := []Token{VAR, IDENTIFIER, ARRAY_TYPE, BRACKET_OPEN, STR_TYPE, BRACKET_CLOSE, ASIGNMENT, ARRAY}
//		if !assertTokensMatch(lexer, expected) {
//			t.Fatalf("Lexemes not properly matched. Given: %v. Expected: %v", lexer.lexemes, expected)
//		}
//	}
func assertTokensMatch(l *Lexer, expected []Token) bool {
	if len(l.lexemes) != len(expected) {
		return false
	}

	for i := 0; i < len(l.lexemes); i++ {
		if l.lexemes[i].token != expected[i] {
			return false
		}
	}

	return true
}
