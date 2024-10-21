package lexer

import (
	"testing"
)

func TestInstantiation(t *testing.T) {
	lexer := Lexer{start: 0, current: 2, source: "Test source"}

	if lexer.start != 0 || lexer.current != 2 || lexer.source != "Test source" {
		t.Fatal("Wrong parameters after instantiation")
	}
}

func TestInit(t *testing.T) {
	lexer := Init(`I'm your father`)
	if lexer.start != 0 || lexer.current != 0 || lexer.source != "I'm your father" {
		t.Fatal("Wrong parameters after Init")
	}
}

func TestScan(t *testing.T) {
	lexer := Init(`var a int`)
	lexer.Scan()
}

func MatchesVarDeclaration(t *testing.T) {
	lexer := Init(`var a int`)
	lexer.Scan()

	if lexer.lexemes[0].token != VAR || lexer.lexemes[1].token != IDENTIFIER || lexer.lexemes[2].token != INT_TYPE {
		t.Fatal("Lexemes not properly matched")
	}
}

func TestMatchesVarDecWithIntAssignment(t *testing.T) {
	lexer := Init(`var v2 int = 2`)
	lexer.Scan()

	expected := []Token{VAR, IDENTIFIER, INT_TYPE, ASIGNMENT, INT}
	if !assertTokensMatch(lexer, expected) {
		t.Fatalf("Lexemes not properly matched. Given: %v.\n Expected: %v", lexer.lexemes, expected)
	}

	assertValue(t, lexer.lexemes[4].value, int64(2))
}

func TestMatchesVarDecWithStringAssignment(t *testing.T) {
	lexer := Init(`var v3 string = "Hello"`)
	lexer.Scan()

	expected := []Token{VAR, IDENTIFIER, STR_TYPE, ASIGNMENT, STR}
	if !assertTokensMatch(lexer, expected) {
		t.Fatalf("Lexemes not properly matched. Given: %v.\n Expected: %v", lexer.lexemes, expected)
	}

	assertValue(t, lexer.lexemes[4].value, `"Hello"`)
}

func TestMatchesVarDecWithBoolAssignment(t *testing.T) {
	lexer := Init(`var v3 bool = true`)
	lexer.Scan()

	expected := []Token{VAR, IDENTIFIER, BOOL_TYPE, ASIGNMENT, BOOL}
	if !assertTokensMatch(lexer, expected) {
		t.Fatalf("Lexemes not properly matched. Given: %v.\n Expected: %v", lexer.lexemes, expected)
	}

	assertValue(t, lexer.lexemes[4].value, true)
}

func TestMatchesVarDecWithFalseBoolAssignment(t *testing.T) {
	lexer := Init(`var v3 bool = false`)
	lexer.Scan()

	expected := []Token{VAR, IDENTIFIER, BOOL_TYPE, ASIGNMENT, BOOL}
	if !assertTokensMatch(lexer, expected) {
		t.Fatalf("Lexemes not properly matched. Given: %v.\n Expected: %v", lexer.lexemes, expected)
	}

	assertValue(t, lexer.lexemes[4].value, false)
}

func TestMatchesVarDecWithIntMulAssignment(t *testing.T) {
	lexer := Init(`var v3 int= 123 *1`)
	lexer.Scan()

	expected := []Token{VAR, IDENTIFIER, INT_TYPE, ASIGNMENT, INT, MUL, INT}
	if !assertTokensMatch(lexer, expected) {
		t.Fatalf("Lexemes not properly matched. Given: %v.\n Expected: %v", lexer.lexemes, expected)
	}

	assertValue(t, lexer.lexemes[4].value, int64(123))
	assertValue(t, lexer.lexemes[6].value, int64(1))
}

func TestMatchesVarDecWithIntModAssignment(t *testing.T) {
	lexer := Init(`var v3 int= 123 %1`)
	lexer.Scan()

	expected := []Token{VAR, IDENTIFIER, INT_TYPE, ASIGNMENT, INT, MOD, INT}
	if !assertTokensMatch(lexer, expected) {
		t.Fatalf("Lexemes not properly matched. Given: %v.\n Expected: %v", lexer.lexemes, expected)
	}

	assertValue(t, lexer.lexemes[4].value, int64(123))
	assertValue(t, lexer.lexemes[6].value, int64(1))
}

func TestMatchesVarDecWithIntDivAssignment(t *testing.T) {
	lexer := Init(`var v3 int= 123 /1`)
	lexer.Scan()

	expected := []Token{VAR, IDENTIFIER, INT_TYPE, ASIGNMENT, INT, DIV, INT}
	if !assertTokensMatch(lexer, expected) {
		t.Fatalf("Lexemes not properly matched. Given: %v.\n Expected: %v", lexer.lexemes, expected)
	}

	assertValue(t, lexer.lexemes[4].value, int64(123))
	assertValue(t, lexer.lexemes[6].value, int64(1))
}

func TestMatchesVarDecWithIntSubAssignment(t *testing.T) {
	lexer := Init(`var v3 int= 123 -1`)
	lexer.Scan()

	expected := []Token{VAR, IDENTIFIER, INT_TYPE, ASIGNMENT, INT, SUB, INT}
	if !assertTokensMatch(lexer, expected) {
		t.Fatalf("Lexemes not properly matched. Given: %v.\n Expected: %v", lexer.lexemes, expected)
	}

	assertValue(t, lexer.lexemes[4].value, int64(123))
	assertValue(t, lexer.lexemes[6].value, int64(1))
}

func TestMatchesVarDecWithFloatAdditionAssignment(t *testing.T) {
	lexer := Init(`var v3 float = 2.156 + 1.0`)
	lexer.Scan()

	expected := []Token{VAR, IDENTIFIER, FLOAT_TYPE, ASIGNMENT, FLOAT, ADD, FLOAT}
	if !assertTokensMatch(lexer, expected) {
		t.Fatalf("Lexemes not properly matched. Given: %v.\n Expected: %v", lexer.lexemes, expected)
	}

	assertValue(t, lexer.lexemes[4].value, 2.156)
	assertValue(t, lexer.lexemes[6].value, 1.0)
}

func TestMatchesVarDecWithFloatAssignment(t *testing.T) {
	lexer := Init(`var v3 float = 2.156`)
	lexer.Scan()

	expected := []Token{VAR, IDENTIFIER, FLOAT_TYPE, ASIGNMENT, FLOAT}
	if !assertTokensMatch(lexer, expected) {
		t.Fatalf("Lexemes not properly matched. Given: %v.\n Expected: %v", lexer.lexemes, expected)
	}

	assertValue(t, lexer.lexemes[4].value, 2.156)
}

func TestMatchesVarDecWithArrayOfBoolAssignment(t *testing.T) {
	lexer := Init(`var v3 array[bool] = [true, false]`)
	lexer.Scan()

	expected := []Token{VAR, IDENTIFIER, ARRAY_TYPE, BRACKET_OPEN, BOOL_TYPE, BRACKET_CLOSE, ASIGNMENT, ARRAY}
	if !assertTokensMatch(lexer, expected) {
		t.Fatalf("Lexemes not properly matched. Given: %v.\n Expected: %v", lexer.lexemes, expected)
	}

	parsedItems := lexer.lexemes[7].value.([]any)
	expectedItems := []bool{true, false}

	if parsedItems[0].(bool) != expectedItems[0] || parsedItems[1].(bool) != expectedItems[1] {
		t.Fatalf("Parsed array literal has wrong values. Given: %v.\n Expected: %v", parsedItems, expectedItems)
	}
}

func TestMatchesVarDecWithArrayOfFloatAssignment(t *testing.T) {
	lexer := Init(`var v3 array[float] = [1.20, 20.0]`)
	lexer.Scan()

	expected := []Token{VAR, IDENTIFIER, ARRAY_TYPE, BRACKET_OPEN, FLOAT_TYPE, BRACKET_CLOSE, ASIGNMENT, ARRAY}
	if !assertTokensMatch(lexer, expected) {
		t.Fatalf("Lexemes not properly matched. Given: %v.\n Expected: %v", lexer.lexemes, expected)
	}

	parsedItems := lexer.lexemes[7].value.([]any)
	expectedItems := []float64{1.20, 20.0}

	if parsedItems[0].(float64) != expectedItems[0] || parsedItems[1].(float64) != expectedItems[1] {
		t.Fatalf("Parsed array literal has wrong values. Given: %v.\n Expected: %v", parsedItems, expectedItems)
	}
}

func TestMatchesVarDecWithArrayOfIntAssignmentOnNewInstruction(t *testing.T) {
	lexer := Init(`
		var v3 array[int]
		v3 = [1, 20]
	`)

	lexer.Scan()

	expected := []Token{VAR, IDENTIFIER, ARRAY_TYPE, BRACKET_OPEN, INT_TYPE, BRACKET_CLOSE, IDENTIFIER, ASIGNMENT, ARRAY}

	if !assertTokensMatch(lexer, expected) {
		t.Fatalf("Lexemes not properly matched. Given: %v.\n Expected: %v", lexer.lexemes, expected)
	}

	parsedItems := lexer.lexemes[8].value.([]any)
	expectedItems := []int64{1, 20}

	if parsedItems[0].(int64) != expectedItems[0] || parsedItems[1].(int64) != expectedItems[1] {
		t.Fatalf("Parsed array literal has wrong values. Given: %v.\n Expected: %v", parsedItems, expectedItems)
	}
}

func TestMatchesVarDecWithArrayOfIntAssignment(t *testing.T) {
	lexer := Init(`var v3 array[int] = [1, 20]`)
	lexer.Scan()

	expected := []Token{VAR, IDENTIFIER, ARRAY_TYPE, BRACKET_OPEN, INT_TYPE, BRACKET_CLOSE, ASIGNMENT, ARRAY}
	if !assertTokensMatch(lexer, expected) {
		t.Fatalf("Lexemes not properly matched. Given: %v.\n Expected: %v", lexer.lexemes, expected)
	}

	parsedItems := lexer.lexemes[7].value.([]any)
	expectedItems := []int64{1, 20}

	if parsedItems[0].(int64) != expectedItems[0] || parsedItems[1].(int64) != expectedItems[1] {
		t.Fatalf("Parsed array literal has wrong values. Given: %v.\n Expected: %v", parsedItems, expectedItems)
	}
}

func TestMatchesVarDecWithArrayOfStringAssignment(t *testing.T) {
	lexer := Init(`var v3 array[string] = ["hello", "world"]`)
	lexer.Scan()

	expected := []Token{VAR, IDENTIFIER, ARRAY_TYPE, BRACKET_OPEN, STR_TYPE, BRACKET_CLOSE, ASIGNMENT, ARRAY}
	if !assertTokensMatch(lexer, expected) {
		t.Fatalf("Lexemes not properly matched. Given: %v.\n Expected: %v", lexer.lexemes, expected)
	}

	parsedItems := lexer.lexemes[7].value.([]any)
	expectedItems := []string{"hello", "world"}

	if parsedItems[0].(string) != expectedItems[0] || parsedItems[1].(string) != expectedItems[1] {
		t.Fatalf("Parsed array literal has wrong values. Given: %v.\n Expected: %v", parsedItems, expectedItems)
	}
}

func TestIfStatement(t *testing.T) {
	lexer := Init(`
		if true {
			var a int = 2
		}
	`)

	lexer.Scan()

	expected := []Token{IF_STMT, BOOL, SQUIRLY_OPEN, VAR, IDENTIFIER, INT_TYPE, ASIGNMENT, INT, SQUIRLY_CLOSE}
	if !assertTokensMatch(lexer, expected) {
		t.Fatalf("Lexemes not properly matched. Given: %v.\n Expected: %v", lexer.lexemes, expected)
	}
}

func TestWhileLoop(t *testing.T) {
	lexer := Init(`
		var a int = 2
		while true {
			a = a + 2
		}
	`)

	lexer.Scan()

	expected := []Token{VAR, IDENTIFIER, INT_TYPE, ASIGNMENT, INT, WHILE_LOOP, BOOL, SQUIRLY_OPEN, IDENTIFIER, ASIGNMENT, IDENTIFIER, ADD, INT, SQUIRLY_CLOSE}
	if !assertTokensMatch(lexer, expected) {
		t.Fatalf("Lexemes not properly matched. Given: %v.\n Expected: %v", lexer.lexemes, expected)
	}
}

func TestForLoop(t *testing.T) {
	lexer := Init(`
		var a int = 2
		for i from 0 to 10 step 1 {
			a = a + 2
		}
	`)

	lexer.Scan()

	expected := []Token{VAR, IDENTIFIER, INT_TYPE, ASIGNMENT, INT, FOR_LOOP, IDENTIFIER, FROM, INT, TO, INT, STEP, INT, SQUIRLY_OPEN, IDENTIFIER, ASIGNMENT, IDENTIFIER, ADD, INT, SQUIRLY_CLOSE}
	if !assertTokensMatch(lexer, expected) {
		t.Fatalf("Lexemes not properly matched. Given: %v.\n Expected: %v", lexer.lexemes, expected)
	}
}

func Test2DArray(t *testing.T) {
	lexer := Init(`
		var a array[array[int]]= [[1,2,3], [3,4,5]]
	`)

	lexer.Scan()

	expected := []Token{VAR, IDENTIFIER, ARRAY_TYPE, BRACKET_OPEN, ARRAY, BRACKET_OPEN, INT, BRACKET_CLOSE, BRACKET_CLOSE, ASIGNMENT, ARRAY}
	if !assertTokensMatch(lexer, expected) {
		t.Fatalf("Lexemes not properly matched. Given: %v.\n Expected: %v", lexer.lexemes, expected)
	}
}

func assertValue(t *testing.T, given any, expected any) {
	if given != expected {
		t.Fatalf("Value doesn't match. Given: %s.\n Expected: %s.", given, expected)
	}
}

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
