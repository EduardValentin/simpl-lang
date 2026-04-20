package simpl

import (
	"strings"
	"testing"
	"time"
)

func mustRun(t *testing.T, source, stdin string, opts RunOptions) RunResult {
	t.Helper()
	res := Run(source, stdin, opts)
	if len(res.Diagnostics) > 0 {
		t.Fatalf("unexpected diagnostics: %+v", res.Diagnostics)
	}
	return res
}

func TestRuntimeReadmeDeclarationsRun(t *testing.T) {
	source := `
var v1 int
var v2 int = 2
var v3 string = "Hello"
var v4 float = 2.14
var v5 array[string] = ["a", "b", "c"]
var v6 array[array[int]] = [[1,2,3], [3,4,5]]
write v1, "|", v2, "|", v3, "|", v4, "|", v5[2], "|", v6[1][2]
`
	res := mustRun(t, source, "", RunOptions{})
	if res.Stdout != "0|2|Hello|2.14|c|5" {
		t.Fatalf("unexpected stdout: %q", res.Stdout)
	}
}

func TestRuntimeIfElseIfElse(t *testing.T) {
	source := `
var v int = 3
if v == 2 {
  write "two"
} else if v < 3 {
  write "lt"
} else {
  write "other"
}
`
	res := mustRun(t, source, "", RunOptions{})
	if res.Stdout != "other" {
		t.Fatalf("unexpected stdout: %q", res.Stdout)
	}
}

func TestRuntimeForExclusiveBounds(t *testing.T) {
	source := `for i from 0 until 5 step 2 { write i, " " }`
	res := mustRun(t, source, "", RunOptions{})
	if res.Stdout != "0 2 4 " {
		t.Fatalf("unexpected stdout: %q", res.Stdout)
	}
}

func TestRuntimeForNegativeStepExclusiveBounds(t *testing.T) {
	source := `for i from 5 until 0 step -2 { write i, " " }`
	res := mustRun(t, source, "", RunOptions{})
	if res.Stdout != "5 3 1 " {
		t.Fatalf("unexpected stdout: %q", res.Stdout)
	}
}

func TestRuntimeWhileLoop(t *testing.T) {
	source := `
var i int = 0
while i <= 5 {
  write i, " "
  i = i + 1
}
`
	res := mustRun(t, source, "", RunOptions{})
	if res.Stdout != "0 1 2 3 4 5 " {
		t.Fatalf("unexpected stdout: %q", res.Stdout)
	}
}

func TestRuntimeArrayIndexReadWrite(t *testing.T) {
	source := `
var a array[int] = [1,2,3]
a[1] = 9
write a[0], ",", a[1], ",", a[2]
`
	res := mustRun(t, source, "", RunOptions{})
	if res.Stdout != "1,9,3" {
		t.Fatalf("unexpected stdout: %q", res.Stdout)
	}
}

func TestRuntimeNestedArrayIndexWrite(t *testing.T) {
	source := `
var a array[array[int]] = [[1,2], [3,4]]
a[1][0] = 7
write a[1][0]
`
	res := mustRun(t, source, "", RunOptions{})
	if res.Stdout != "7" {
		t.Fatalf("unexpected stdout: %q", res.Stdout)
	}
}

func TestRuntimeStringSequenceOperations(t *testing.T) {
	source := `
var s string = "a\u0103c"
write size s, "|", s[1], "|"
s[1] = "x"
pop s
write s
`
	res := mustRun(t, source, "", RunOptions{})
	if res.Stdout != "3|ă|ax" {
		t.Fatalf("unexpected stdout: %q", res.Stdout)
	}
}

func TestRuntimePopNestedStringInArray(t *testing.T) {
	source := `
var words array[string] = ["abc", "xy"]
pop words[0]
write words[0], "|", words[1]
`
	res := mustRun(t, source, "", RunOptions{})
	if res.Stdout != "ab|xy" {
		t.Fatalf("unexpected stdout: %q", res.Stdout)
	}
}

func TestRuntimeStringIndexAssignmentRequiresSingleRune(t *testing.T) {
	source := `
var s string = "abc"
s[1] = "xy"
`
	res := Run(source, "", RunOptions{})
	if len(res.Diagnostics) == 0 {
		t.Fatal("expected diagnostics")
	}
	if res.Diagnostics[0].Code != "RUNTIME_TYPE" {
		t.Fatalf("unexpected code: %s", res.Diagnostics[0].Code)
	}
}

func TestRuntimePopEmptySequence(t *testing.T) {
	stringRes := Run("var s string = \"\"\npop s", "", RunOptions{})
	if len(stringRes.Diagnostics) == 0 {
		t.Fatal("expected string diagnostics")
	}
	if stringRes.Diagnostics[0].Code != "RUNTIME_POP_EMPTY" {
		t.Fatalf("unexpected string code: %s", stringRes.Diagnostics[0].Code)
	}

	arrayRes := Run("var a array[int] = []\npop a", "", RunOptions{})
	if len(arrayRes.Diagnostics) == 0 {
		t.Fatal("expected array diagnostics")
	}
	if arrayRes.Diagnostics[0].Code != "RUNTIME_POP_EMPTY" {
		t.Fatalf("unexpected array code: %s", arrayRes.Diagnostics[0].Code)
	}
}

func TestRuntimeReadTokenBased(t *testing.T) {
	source := `
var x int
var y string
read x
read y
write x, "-", y
`
	res := mustRun(t, source, "42 hello", RunOptions{})
	if res.Stdout != "42-hello" {
		t.Fatalf("unexpected stdout: %q", res.Stdout)
	}
}

func TestRuntimeReadParseError(t *testing.T) {
	source := `var x int
read x`
	res := Run(source, "abc", RunOptions{})
	if len(res.Diagnostics) == 0 {
		t.Fatal("expected diagnostics")
	}
	if res.Diagnostics[0].Code != "RUNTIME_READ_PARSE" {
		t.Fatalf("unexpected code: %s", res.Diagnostics[0].Code)
	}
}

func TestRuntimeWriteNoAutoNewline(t *testing.T) {
	source := `write "abc"`
	res := mustRun(t, source, "", RunOptions{})
	if strings.Contains(res.Stdout, "\n") {
		t.Fatalf("stdout should not contain newline: %q", res.Stdout)
	}
}

func TestRuntimeDivisionByZero(t *testing.T) {
	source := `
var a int = 1
var b int = 0
write a / b
`
	res := Run(source, "", RunOptions{})
	if len(res.Diagnostics) == 0 {
		t.Fatal("expected diagnostics")
	}
	if res.Diagnostics[0].Code != "RUNTIME_DIV_ZERO" {
		t.Fatalf("unexpected code: %s", res.Diagnostics[0].Code)
	}
}

func TestRuntimeStepLimitExceeded(t *testing.T) {
	source := `
var i int = 0
while true {
  i = i + 1
}
`
	res := Run(source, "", RunOptions{MaxSteps: 100, Timeout: 2 * time.Second})
	if len(res.Diagnostics) == 0 {
		t.Fatal("expected diagnostics")
	}
	if res.Diagnostics[0].Code != "LIMIT_STEPS_EXCEEDED" {
		t.Fatalf("unexpected code: %s", res.Diagnostics[0].Code)
	}
}

func TestRuntimeMalformedProgramNoPanic(t *testing.T) {
	diags := Validate("var a int =")
	if len(diags) == 0 {
		t.Fatal("expected diagnostics")
	}
	if diags[0].Category != "parser" {
		t.Fatalf("unexpected category: %s", diags[0].Category)
	}
}
