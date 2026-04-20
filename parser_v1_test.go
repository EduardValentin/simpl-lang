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
	if len(prog.Statements) != 4 {
		t.Fatalf("expected 4 statements, got %d", len(prog.Statements))
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

	if _, ok := prog.Statements[2].(*PopStmt); !ok {
		t.Fatalf("expected pop statement, got %T", prog.Statements[2])
	}

	assignStmt, ok := prog.Statements[3].(*AssignStmt)
	if !ok {
		t.Fatalf("expected assignment statement, got %T", prog.Statements[3])
	}
	if _, ok := assignStmt.Target.(*IndexExpr); !ok {
		t.Fatalf("expected index target, got %T", assignStmt.Target)
	}
}
