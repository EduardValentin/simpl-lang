package simpl

import (
	"strconv"
	"strings"
	"testing"
)

// These tests cover the recursion-depth limit (maxNestingDepth, see
// limits.go) that bounds parser, checker, and interpreter recursion so
// pathologically nested source cannot exhaust the Go call stack and crash
// the host process with a fatal, unrecoverable "stack overflow" runtime
// error.
//
// Boundary note: for an expression built purely from one repeated nesting
// construct (parens, chained unary, or single-element nested arrays)
// wrapped in a single `write` statement, the checker is empirically the
// first stage to trip the shared depth limit, at two levels below the raw
// construct count. That is expected, not a bug: checkStmt spends one unit
// of depth entering the enclosing WriteStmt, and checkExpr spends one more
// unit reaching the innermost literal, on top of one unit per nested
// construct. So maxNestingDepth-2 repeated constructs is the last value
// that stays within budget end-to-end, and maxNestingDepth-1 is the first
// value that exceeds it. These exact offsets were verified empirically
// (by bisecting the pass/fail transition for each construct) rather than
// assumed from a naive 1:1 reading of maxNestingDepth.
const (
	nestingWithinBudget = maxNestingDepth - 2 // last value that still passes
	nestingOverBudget   = maxNestingDepth - 1 // first value that trips the limit
)

// wellOverLimit is comfortably larger than maxNestingDepth (enough to
// prove the guard fires deep inside real recursion) but nowhere near the
// several-hundred-thousand-level range that was empirically needed to
// crash the pre-fix interpreter with a megabyte of source. Using a value
// in that range here would make the regression suite slow without adding
// coverage.
const wellOverLimit = 10 * maxNestingDepth

func assertSingleMaxDepthDiagnostic(t *testing.T, diags []Diagnostic) {
	t.Helper()
	if len(diags) != 1 {
		t.Fatalf("expected exactly 1 diagnostic, got %d: %+v", len(diags), diags)
	}
	if diags[0].Code != maxDepthCode {
		t.Fatalf("expected code %s, got %s (%s)", maxDepthCode, diags[0].Code, diags[0].Message)
	}
	if diags[0].Category != maxDepthCategory {
		t.Fatalf("expected category %s, got %s", maxDepthCategory, diags[0].Category)
	}
}

// --- Regression tests: one per required construct ---------------------

func TestDepthLimit_NestedParensDoesNotCrashAndReturnsDiagnostic(t *testing.T) {
	src := "write " + strings.Repeat("(", wellOverLimit) + "1" + strings.Repeat(")", wellOverLimit)

	result := Run(src, "", RunOptions{})

	if result.Stdout != "" {
		t.Fatalf("expected no stdout for rejected program, got %q", result.Stdout)
	}
	assertSingleMaxDepthDiagnostic(t, result.Diagnostics)
}

func TestDepthLimit_ChainedUnaryDoesNotCrashAndReturnsDiagnostic(t *testing.T) {
	src := "write " + strings.Repeat("-", wellOverLimit) + "1"

	result := Run(src, "", RunOptions{})

	if result.Stdout != "" {
		t.Fatalf("expected no stdout for rejected program, got %q", result.Stdout)
	}
	assertSingleMaxDepthDiagnostic(t, result.Diagnostics)
}

func TestDepthLimit_NestedArrayLiteralDoesNotCrashAndReturnsDiagnostic(t *testing.T) {
	src := "write " + strings.Repeat("[", wellOverLimit) + "1" + strings.Repeat("]", wellOverLimit)

	result := Run(src, "", RunOptions{})

	if result.Stdout != "" {
		t.Fatalf("expected no stdout for rejected program, got %q", result.Stdout)
	}
	assertSingleMaxDepthDiagnostic(t, result.Diagnostics)
}

// --- Boundary tests ------------------------------------------------------

func TestDepthLimit_BoundaryAtLimitPasses(t *testing.T) {
	cases := map[string]string{
		"parens": "write " + strings.Repeat("(", nestingWithinBudget) + "1" + strings.Repeat(")", nestingWithinBudget),
		"unary":  "write " + strings.Repeat("-", nestingWithinBudget) + "1",
		"arrays": "write " + strings.Repeat("[", nestingWithinBudget) + "1" + strings.Repeat("]", nestingWithinBudget),
	}
	for name, src := range cases {
		t.Run(name, func(t *testing.T) {
			result := Run(src, "", RunOptions{})
			if len(result.Diagnostics) != 0 {
				t.Fatalf("expected no diagnostics at the depth budget, got %+v", result.Diagnostics)
			}
			if result.Stdout == "" {
				t.Fatalf("expected program to actually execute and write output")
			}
		})
	}
}

func TestDepthLimit_BoundaryOnePastLimitFails(t *testing.T) {
	cases := map[string]string{
		"parens": "write " + strings.Repeat("(", nestingOverBudget) + "1" + strings.Repeat(")", nestingOverBudget),
		"unary":  "write " + strings.Repeat("-", nestingOverBudget) + "1",
		"arrays": "write " + strings.Repeat("[", nestingOverBudget) + "1" + strings.Repeat("]", nestingOverBudget),
	}
	for name, src := range cases {
		t.Run(name, func(t *testing.T) {
			result := Run(src, "", RunOptions{})
			assertSingleMaxDepthDiagnostic(t, result.Diagnostics)
		})
	}
}

// --- False-positive tests: the limit must not reject legitimate code ----

func TestDepthLimit_DoesNotRejectRealisticNestedProgram(t *testing.T) {
	// A small, realistic teaching exercise: a handful of nested control
	// flow blocks, a nested array, and normal expression nesting. Real
	// Simpl programs look like this, nowhere near maxNestingDepth.
	src := `
var matrix array[array[int]] = [[1, 2, 3], [4, 5, 6], [7, 8, 9]]
var total int = 0

for i from 0 until size matrix step 1 {
    for j from 0 until size matrix[i] step 1 {
        if matrix[i][j] % 2 == 0 {
            if matrix[i][j] > 0 {
                total = total + matrix[i][j]
            }
        } else {
            total = total + (matrix[i][j] * 2)
        }
    }
}

write total
`
	result := Run(src, "", RunOptions{})
	if len(result.Diagnostics) != 0 {
		t.Fatalf("expected no diagnostics for a realistic nested program, got %+v", result.Diagnostics)
	}
	if result.Stdout != "70" {
		t.Fatalf("expected stdout %q, got %q", "70", result.Stdout)
	}
}

func TestDepthLimit_DoesNotRejectManySequentialShallowStatements(t *testing.T) {
	// Depth accounting must be per-recursion-depth, not a monotonic
	// counter: a long program made of many shallow, sequential
	// statements/expressions must not falsely trip the limit just
	// because it has a lot of them. This guards against a depth
	// counter that is incremented on entry but never decremented on
	// exit (the exact bug called out in the implementation plan).
	const width = 5 * maxNestingDepth

	var b strings.Builder
	for i := 0; i < width; i++ {
		b.WriteString("var v")
		b.WriteString(strconv.Itoa(i))
		b.WriteString(" int = (1 + 2) * 3\n")
	}
	b.WriteString("write ")
	for i := 0; i < width; i++ {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString("v")
		b.WriteString(strconv.Itoa(i))
	}

	result := Run(b.String(), "", RunOptions{})
	if len(result.Diagnostics) != 0 {
		t.Fatalf("expected no diagnostics for %d sequential shallow statements, got %d diagnostics (first: %+v)", width, len(result.Diagnostics), result.Diagnostics[0])
	}
	if len(result.Stdout) == 0 {
		t.Fatalf("expected non-empty stdout")
	}
}

// --- Defense in depth: checker and interpreter must not rely solely on
// the parser having already rejected deep input. These tests build an AST
// directly (bypassing the parser entirely) so they exercise checkProgram
// and the interpreter's own independent guards. ------------------------

// deeplyNestedUnaryExpr builds `- - - ... - 1` as an AST, n levels deep,
// without going through the parser.
func deeplyNestedUnaryExpr(n int) Expr {
	var expr Expr = &LiteralExpr{Pos: Position{Line: 1, Column: 1}, Value: int64(1)}
	for i := 0; i < n; i++ {
		expr = &UnaryExpr{Pos: Position{Line: 1, Column: 1}, Operator: TokenMinus, Right: expr}
	}
	return expr
}

func TestDepthLimit_CheckerRejectsDeepASTEvenWithoutParserInvolved(t *testing.T) {
	prog := &Program{
		Statements: []Stmt{
			&WriteStmt{
				Pos:    Position{Line: 1, Column: 1},
				Values: []Expr{deeplyNestedUnaryExpr(wellOverLimit)},
			},
		},
	}

	diags := checkProgram(prog)

	assertSingleMaxDepthDiagnostic(t, diags)
}

func TestDepthLimit_InterpreterRejectsDeepASTEvenWithoutParserInvolved(t *testing.T) {
	prog := &Program{
		Statements: []Stmt{
			&WriteStmt{
				Pos:    Position{Line: 1, Column: 1},
				Values: []Expr{deeplyNestedUnaryExpr(wellOverLimit)},
			},
		},
	}

	// Go through normalizeOptions the same way Run() does: an
	// unnormalized zero-value RunOptions sets the deadline to
	// time.Now(), which would trip the unrelated timeout limit before
	// the depth guard ever gets a chance to run.
	interp := newInterpreter("", normalizeOptions(RunOptions{}))
	result := interp.run(prog)

	if result.Stdout != "" {
		t.Fatalf("expected no stdout for rejected program, got %q", result.Stdout)
	}
	assertSingleMaxDepthDiagnostic(t, result.Diagnostics)
}

// deeplyNestedIndexExpr builds `name[0][0]...[0]` as an AST, n index
// brackets deep, without going through the parser.
func deeplyNestedIndexExpr(name string, n int) Expr {
	var expr Expr = &IdentifierExpr{Pos: Position{Line: 1, Column: 1}, Name: name}
	for i := 0; i < n; i++ {
		expr = &IndexExpr{
			Pos:        Position{Line: 1, Column: 1},
			Collection: expr,
			Index:      &LiteralExpr{Pos: Position{Line: 1, Column: 1}, Value: int64(0)},
		}
	}
	return expr
}

// deeplyNestedArrayValue builds a runtime Value nested `depth` levels deep
// (array[array[...int...]]), with a single int at the bottom. It is built
// with a plain for-loop, not recursion, so constructing it is safe
// regardless of depth — it exists purely so a test can seed a target that
// mutateValueAtIndexes will actually keep descending into (an array whose
// declared/runtime type is too shallow would instead hit an unrelated
// RUNTIME_TYPE mismatch a couple of levels in, before ever reaching the
// depth guard).
func deeplyNestedArrayValue(depth int) Value {
	v := Value{Type: Type{Kind: TypeInt}, Int: 0}
	for i := 0; i < depth; i++ {
		elemType := v.Type
		v = Value{Type: Type{Kind: TypeArray, Elem: &elemType}, Array: []Value{v}}
	}
	return v
}

// Regression test for F1: mutateValueAtIndexes (reached via an indexed
// assignment target like a[0][0][0]...) recurses once per index level and
// must carry its own depth guard rather than relying on the parser or
// checker having already rejected the source. This builds the assignment
// target AST directly (bypassing the parser and checker entirely, the
// same way the tests above do for the checker and the top-level
// interpreter guard) so it exercises mutateValueAtIndexes's own,
// independent protection.
func TestDepthLimit_InterpreterRejectsDeepIndexChainAssignmentEvenWithoutParserInvolved(t *testing.T) {
	interp := newInterpreter("", normalizeOptions(RunOptions{}))
	interp.pushScope()
	interp.scopes[0]["a"] = runtimeVar{value: deeplyNestedArrayValue(wellOverLimit), mutable: true}

	prog := &Program{
		Statements: []Stmt{
			&AssignStmt{
				Pos:    Position{Line: 1, Column: 1},
				Target: deeplyNestedIndexExpr("a", wellOverLimit),
				Value:  &LiteralExpr{Pos: Position{Line: 1, Column: 1}, Value: int64(7)},
			},
		},
	}

	result := interp.run(prog)

	assertSingleMaxDepthDiagnostic(t, result.Diagnostics)
}

func TestDepthLimit_NestedBlocksDoNotCrashAndReturnDiagnostic(t *testing.T) {
	// Blocks (if/while/for bodies) are a second, independent recursion
	// axis from expressions. A deeply nested if-chain must be rejected
	// the same way as deeply nested expressions.
	const n = 20 * maxNestingDepth
	src := strings.Repeat("if true {", n) + "write 1" + strings.Repeat("}", n)

	result := Run(src, "", RunOptions{})

	assertSingleMaxDepthDiagnostic(t, result.Diagnostics)
}
