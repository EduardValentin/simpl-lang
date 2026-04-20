package simpl

import "testing"

func TestLexerTokensAndComments(t *testing.T) {
	source := `
// comment
var a int = 1
if a == 1 {
  write "ok"
}
`
	tokens, diags := lexSource(source)
	if len(diags) != 0 {
		t.Fatalf("unexpected diagnostics: %+v", diags)
	}

	want := []TokenType{
		TokenVar, TokenIdentifier, TokenTypeInt, TokenAssign, TokenInt,
		TokenIf, TokenIdentifier, TokenEqual, TokenInt, TokenLBrace,
		TokenWrite, TokenString, TokenRBrace,
		TokenEOF,
	}
	if len(tokens) != len(want) {
		t.Fatalf("token count mismatch: got %d want %d", len(tokens), len(want))
	}
	for i := range want {
		if tokens[i].Type != want[i] {
			t.Fatalf("token[%d] mismatch: got %s want %s", i, tokens[i].Type.String(), want[i].String())
		}
	}
}

func TestLexerUnterminatedString(t *testing.T) {
	_, diags := lexSource(`write "nope`)
	if len(diags) == 0 {
		t.Fatal("expected diagnostics")
	}
	if diags[0].Code != "LEX_UNTERMINATED_STRING" {
		t.Fatalf("unexpected code: %s", diags[0].Code)
	}
}

func TestLexerUnknownCharacter(t *testing.T) {
	_, diags := lexSource("var a int = 1 @ 2")
	if len(diags) == 0 {
		t.Fatal("expected diagnostics")
	}
	if diags[0].Code != "LEX_UNKNOWN_CHAR" {
		t.Fatalf("unexpected code: %s", diags[0].Code)
	}
}

func TestLexerRecognizesSequenceKeywords(t *testing.T) {
	tokens, diags := lexSource("write size name\npush name, \"a\"\npop name")
	if len(diags) != 0 {
		t.Fatalf("unexpected diagnostics: %+v", diags)
	}

	want := []TokenType{
		TokenWrite, TokenSize, TokenIdentifier,
		TokenPush, TokenIdentifier, TokenComma, TokenString,
		TokenPop, TokenIdentifier,
		TokenEOF,
	}
	if len(tokens) != len(want) {
		t.Fatalf("token count mismatch: got %d want %d", len(tokens), len(want))
	}
	for i := range want {
		if tokens[i].Type != want[i] {
			t.Fatalf("token[%d] mismatch: got %s want %s", i, tokens[i].Type.String(), want[i].String())
		}
	}
}
