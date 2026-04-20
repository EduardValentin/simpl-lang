package simpl

import "testing"

func TestParserBuildsIfElseTree(t *testing.T) {
	source := `
var a int = 2
if a == 1 {
  write "one"
} else if a == 2 {
  write "two"
} else {
  write "other"
}
`
	tokens, ldiags := lexSource(source)
	if len(ldiags) != 0 {
		t.Fatalf("unexpected lexer diagnostics: %+v", ldiags)
	}
	prog, pdiags := parseProgram(tokens)
	if len(pdiags) != 0 {
		t.Fatalf("unexpected parser diagnostics: %+v", pdiags)
	}
	if len(prog.Statements) != 2 {
		t.Fatalf("expected 2 statements, got %d", len(prog.Statements))
	}
	ifStmt, ok := prog.Statements[1].(*IfStmt)
	if !ok {
		t.Fatalf("expected if statement, got %T", prog.Statements[1])
	}
	if len(ifStmt.ElseIfs) != 1 {
		t.Fatalf("expected 1 else-if, got %d", len(ifStmt.ElseIfs))
	}
	if ifStmt.ElsePart == nil {
		t.Fatal("expected else block")
	}
}

func TestParserAssignmentTargetIndex(t *testing.T) {
	source := `var a array[int] = [1,2,3]
a[1] = 9`
	tokens, ldiags := lexSource(source)
	if len(ldiags) != 0 {
		t.Fatalf("unexpected lexer diagnostics: %+v", ldiags)
	}
	prog, pdiags := parseProgram(tokens)
	if len(pdiags) != 0 {
		t.Fatalf("unexpected parser diagnostics: %+v", pdiags)
	}
	if len(prog.Statements) != 2 {
		t.Fatalf("expected 2 statements, got %d", len(prog.Statements))
	}
	assign, ok := prog.Statements[1].(*AssignStmt)
	if !ok {
		t.Fatalf("expected assignment statement, got %T", prog.Statements[1])
	}
	if _, ok := assign.Target.(*IndexExpr); !ok {
		t.Fatalf("expected index target, got %T", assign.Target)
	}
}

func TestParserBuildsSequencePrimitives(t *testing.T) {
	source := `var s string = "abc"
write size s
push s, "d", "e"
pop s
s[1] = "x"`
	tokens, ldiags := lexSource(source)
	if len(ldiags) != 0 {
		t.Fatalf("unexpected lexer diagnostics: %+v", ldiags)
	}
	prog, pdiags := parseProgram(tokens)
	if len(pdiags) != 0 {
		t.Fatalf("unexpected parser diagnostics: %+v", pdiags)
	}
	if len(prog.Statements) != 5 {
		t.Fatalf("expected 5 statements, got %d", len(prog.Statements))
	}

	writeStmt, ok := prog.Statements[1].(*WriteStmt)
	if !ok {
		t.Fatalf("expected write statement, got %T", prog.Statements[1])
	}
	if len(writeStmt.Values) != 1 {
		t.Fatalf("expected 1 write value, got %d", len(writeStmt.Values))
	}
	if _, ok := writeStmt.Values[0].(*SizeExpr); !ok {
		t.Fatalf("expected size expression, got %T", writeStmt.Values[0])
	}

	pushStmt, ok := prog.Statements[2].(*PushStmt)
	if !ok {
		t.Fatalf("expected push statement, got %T", prog.Statements[2])
	}
	if len(pushStmt.Values) != 2 {
		t.Fatalf("expected 2 push values, got %d", len(pushStmt.Values))
	}

	if _, ok := prog.Statements[3].(*PopStmt); !ok {
		t.Fatalf("expected pop statement, got %T", prog.Statements[3])
	}

	assignStmt, ok := prog.Statements[4].(*AssignStmt)
	if !ok {
		t.Fatalf("expected assignment statement, got %T", prog.Statements[4])
	}
	if _, ok := assignStmt.Target.(*IndexExpr); !ok {
		t.Fatalf("expected index target, got %T", assignStmt.Target)
	}
}

func TestParserRejectsPushWithoutValue(t *testing.T) {
	source := `var a array[int] = [1]
push a`
	tokens, ldiags := lexSource(source)
	if len(ldiags) != 0 {
		t.Fatalf("unexpected lexer diagnostics: %+v", ldiags)
	}
	_, pdiags := parseProgram(tokens)
	if len(pdiags) == 0 {
		t.Fatal("expected parser diagnostics")
	}
	if pdiags[0].Code != "PARSE_EXPECTED_TOKEN" {
		t.Fatalf("unexpected code: %s", pdiags[0].Code)
	}
}
