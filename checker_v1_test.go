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

func TestCheckerAcceptsStringSequenceOperations(t *testing.T) {
	source := `
var s string = "abc"
write size s, s[0]
s[1] = "x"
pop s
`
	diags := Validate(source)
	if len(diags) != 0 {
		t.Fatalf("unexpected diagnostics: %+v", diags)
	}
}

func TestCheckerRejectsInvalidSequencePrimitives(t *testing.T) {
	source := `
var n int = 1
write size n
pop n
`
	diags := Validate(source)
	if len(diags) < 2 {
		t.Fatalf("expected at least 2 diagnostics, got %+v", diags)
	}
	if diags[0].Code != "TYPE_INVALID_SIZE" {
		t.Fatalf("unexpected first code: %s", diags[0].Code)
	}
	if diags[1].Code != "TYPE_INVALID_POP" {
		t.Fatalf("unexpected second code: %s", diags[1].Code)
	}
}

func TestCheckerRejectsConstStringMutation(t *testing.T) {
	source := `
const s string = "abc"
s[0] = "x"
`
	diags := Validate(source)
	if len(diags) == 0 {
		t.Fatal("expected diagnostics")
	}
	if diags[0].Code != "TYPE_CONST_REASSIGN" {
		t.Fatalf("unexpected code: %s", diags[0].Code)
	}
}
