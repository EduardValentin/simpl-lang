package simpl

import "testing"

func TestCheckerRejectsMixedNumericTypes(t *testing.T) {
	diags := Validate("var a int = 1 + 2.0")
	if len(diags) == 0 {
		t.Fatal("expected diagnostics")
	}
	if diags[0].Code != "TYPE_MISMATCH" {
		t.Fatalf("unexpected code: %s", diags[0].Code)
	}
}

func TestCheckerRejectsConstReassign(t *testing.T) {
	source := `
const a int = 1
a = 2
`
	diags := Validate(source)
	if len(diags) == 0 {
		t.Fatal("expected diagnostics")
	}
	if diags[0].Code != "TYPE_CONST_REASSIGN" {
		t.Fatalf("unexpected code: %s", diags[0].Code)
	}
}

func TestCheckerRejectsUndeclaredIdentifier(t *testing.T) {
	diags := Validate("write missing")
	if len(diags) == 0 {
		t.Fatal("expected diagnostics")
	}
	if diags[0].Code != "TYPE_UNDECLARED_IDENTIFIER" {
		t.Fatalf("unexpected code: %s", diags[0].Code)
	}
}
