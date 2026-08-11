package simpl

import (
	"context"
	"reflect"
	"strings"
	"testing"
	"time"
	"unicode/utf8"
	"unsafe"
)

func TestRunOptionsExposesMaxValueAllocationBytes(t *testing.T) {
	field, ok := reflect.TypeOf(RunOptions{}).FieldByName("MaxValueAllocationBytes")
	if !ok {
		t.Fatal("expected RunOptions.MaxValueAllocationBytes")
	}
	if field.Type.Kind() != reflect.Int64 {
		t.Fatalf("expected int64 value allocation limit, got %s", field.Type)
	}
}

func runWithValueAllocationEvents(source string, budget int64) (RunResult, []string, []Diagnostic) {
	chunks := make([]string, 0, 1)
	diagnostics := make([]Diagnostic, 0, 1)
	result := Run(source, "", RunOptions{
		MaxValueAllocationBytes: budget,
		OnStdoutChunk: func(chunk string, _ int64) {
			chunks = append(chunks, chunk)
		},
		OnDiagnostic: func(diagnostic Diagnostic) {
			diagnostics = append(diagnostics, diagnostic)
		},
	})
	return result, chunks, diagnostics
}

func requireValueAllocationSuccess(t *testing.T, result RunResult, chunks []string, diagnostics []Diagnostic, expectedOutput string) {
	t.Helper()
	if result.Stdout != expectedOutput || strings.Join(chunks, "") != expectedOutput {
		t.Fatalf("expected stdout %q, got stdout=%q chunks=%#v", expectedOutput, result.Stdout, chunks)
	}
	if len(result.Diagnostics) != 0 || len(diagnostics) != 0 {
		t.Fatalf("expected no diagnostics, got result=%+v callbacks=%+v", result.Diagnostics, diagnostics)
	}
}

func requireValueAllocationLimit(t *testing.T, result RunResult, chunks []string, diagnostics []Diagnostic, line int) {
	t.Helper()
	if result.Stdout != "" || len(chunks) != 0 {
		t.Fatalf("expected no stdout callbacks, got stdout=%q chunks=%#v", result.Stdout, chunks)
	}
	if len(result.Diagnostics) != 1 || result.Diagnostics[0].Code != "LIMIT_VALUE_ALLOCATION_EXCEEDED" || result.Diagnostics[0].Category != "limit" {
		t.Fatalf("expected value allocation limit diagnostic, got %+v", result.Diagnostics)
	}
	if result.Diagnostics[0].Line != line {
		t.Fatalf("expected diagnostic on line %d, got %+v", line, result.Diagnostics[0])
	}
	if len(diagnostics) != 1 || diagnostics[0] != result.Diagnostics[0] {
		t.Fatalf("expected one matching diagnostic callback, got %+v", diagnostics)
	}
}

func TestRunOptionsOnStdoutChunk(t *testing.T) {
	source := `write "A", "B", 3`
	chunks := make([]string, 0, 3)
	steps := make([]int64, 0, 3)

	res := Run(source, "", RunOptions{
		OnStdoutChunk: func(chunk string, stepsUsed int64) {
			chunks = append(chunks, chunk)
			steps = append(steps, stepsUsed)
		},
	})

	if len(res.Diagnostics) > 0 {
		t.Fatalf("unexpected diagnostics: %+v", res.Diagnostics)
	}
	if res.Stdout != "AB3" {
		t.Fatalf("unexpected stdout: %q", res.Stdout)
	}
	if len(chunks) != 3 {
		t.Fatalf("expected 3 chunks, got %d", len(chunks))
	}
	if chunks[0] != "A" || chunks[1] != "B" || chunks[2] != "3" {
		t.Fatalf("unexpected chunks: %#v", chunks)
	}
	for i := range steps {
		if steps[i] <= 0 {
			t.Fatalf("expected positive steps in callback, got %d", steps[i])
		}
	}
}

func TestRunOptionsOnDiagnosticRuntimeAndLimit(t *testing.T) {
	t.Run("runtime diagnostic", func(t *testing.T) {
		diags := make([]Diagnostic, 0, 1)
		res := Run(`var a int = 1
var b int = 0
write a / b`, "", RunOptions{
			OnDiagnostic: func(d Diagnostic) {
				diags = append(diags, d)
			},
		})

		if len(res.Diagnostics) == 0 {
			t.Fatal("expected diagnostics")
		}
		if len(diags) == 0 {
			t.Fatal("expected callback diagnostics")
		}
		if diags[0].Code != "RUNTIME_DIV_ZERO" {
			t.Fatalf("unexpected callback diagnostic code: %s", diags[0].Code)
		}
	})

	t.Run("limit diagnostic", func(t *testing.T) {
		diags := make([]Diagnostic, 0, 1)
		res := Run(`var i int = 0
while true { i = i + 1 }`, "", RunOptions{
			MaxSteps: 10,
			Timeout:  2 * time.Second,
			OnDiagnostic: func(d Diagnostic) {
				diags = append(diags, d)
			},
		})

		if len(res.Diagnostics) == 0 {
			t.Fatal("expected diagnostics")
		}
		if len(diags) == 0 {
			t.Fatal("expected callback diagnostics")
		}
		if diags[0].Code != "LIMIT_STEPS_EXCEEDED" {
			t.Fatalf("unexpected callback diagnostic code: %s", diags[0].Code)
		}
	})
}

func TestRunOptionsMaxOutputBytes(t *testing.T) {
	chunks := make([]string, 0, 2)
	diagnostics := make([]Diagnostic, 0, 1)

	result := Run(`write "abc", "éé"`, "", RunOptions{
		MaxOutputBytes: 5,
		OnStdoutChunk: func(chunk string, _ int64) {
			chunks = append(chunks, chunk)
		},
		OnDiagnostic: func(diagnostic Diagnostic) {
			diagnostics = append(diagnostics, diagnostic)
		},
	})

	if result.Stdout != "abcé" {
		t.Fatalf("unexpected bounded stdout: %q", result.Stdout)
	}
	if len(result.Stdout) != 5 {
		t.Fatalf("expected 5 output bytes, got %d", len(result.Stdout))
	}
	if !utf8.ValidString(result.Stdout) {
		t.Fatalf("expected valid UTF-8 output: %q", result.Stdout)
	}
	if strings.Join(chunks, "") != result.Stdout {
		t.Fatalf("expected bounded callback chunks, got %#v", chunks)
	}
	if len(result.Diagnostics) != 1 || result.Diagnostics[0].Code != "LIMIT_OUTPUT_EXCEEDED" {
		t.Fatalf("expected output limit diagnostic, got %+v", result.Diagnostics)
	}
	if len(diagnostics) != 1 || diagnostics[0].Code != "LIMIT_OUTPUT_EXCEEDED" {
		t.Fatalf("expected output limit callback, got %+v", diagnostics)
	}
}

func TestRunOptionsMaxOutputBytesAllowsExactBoundary(t *testing.T) {
	result := Run(`write "abc", "é"`, "", RunOptions{MaxOutputBytes: 5})

	if result.Stdout != "abcé" {
		t.Fatalf("unexpected stdout: %q", result.Stdout)
	}
	if len(result.Diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %+v", result.Diagnostics)
	}
}

func TestRunOptionsMaxOutputBytesRejectsNegativeLimit(t *testing.T) {
	result := Run(`write "abc"`, "", RunOptions{MaxOutputBytes: -1})

	if result.Stdout != "" {
		t.Fatalf("expected no stdout, got %q", result.Stdout)
	}
	if len(result.Diagnostics) != 1 || result.Diagnostics[0].Code != "LIMIT_INVALID_OUTPUT_LIMIT" {
		t.Fatalf("expected invalid output limit diagnostic, got %+v", result.Diagnostics)
	}
}

func TestRunOptionsZeroMaxOutputBytesRemainsUnlimited(t *testing.T) {
	result := Run(`write "abc"`, "", RunOptions{MaxOutputBytes: 0})

	if result.Stdout != "abc" || len(result.Diagnostics) != 0 {
		t.Fatalf("expected unlimited output, got %+v", result)
	}
}

func TestRunOptionsMaxValueAllocationBytesAllowsExactBoundary(t *testing.T) {
	result := Run(`write "a"`, "", RunOptions{MaxValueAllocationBytes: 5})

	if result.Stdout != "a" || len(result.Diagnostics) != 0 {
		t.Fatalf("expected exact value allocation boundary to succeed, got %+v", result)
	}
}

func TestRunOptionsMaxValueAllocationBytesRejectsPastBoundary(t *testing.T) {
	result := Run(`write "a"`, "", RunOptions{MaxValueAllocationBytes: 4})

	if result.Stdout != "" {
		t.Fatalf("expected no stdout, got %q", result.Stdout)
	}
	if len(result.Diagnostics) != 1 || result.Diagnostics[0].Code != "LIMIT_VALUE_ALLOCATION_EXCEEDED" {
		t.Fatalf("expected value allocation limit diagnostic, got %+v", result.Diagnostics)
	}
}

func TestRunOptionsMaxValueAllocationBytesStopsStringDoubling(t *testing.T) {
	result := Run(`var value string = "a"
for i from 0 until 12 step 1 {
    value = value + value
}
write value`, "", RunOptions{
		MaxSteps:                1 << 62,
		MaxValueAllocationBytes: 256,
		Timeout:                 time.Minute,
	})

	if result.Stdout != "" {
		t.Fatalf("expected no stdout, got %d bytes", len(result.Stdout))
	}
	if len(result.Diagnostics) != 1 || result.Diagnostics[0].Code != "LIMIT_VALUE_ALLOCATION_EXCEEDED" {
		t.Fatalf("expected value allocation limit diagnostic, got %+v", result.Diagnostics)
	}
}

func TestRunOptionsMaxValueAllocationBytesStopsArrayGrowth(t *testing.T) {
	valueSlotBytes := int64(unsafe.Sizeof(Value{}))
	typeSlotBytes := int64(unsafe.Sizeof(Type{}))
	result := Run(`var values array[int] = [0]
push values, 1
write "unreachable"`, "", RunOptions{MaxValueAllocationBytes: typeSlotBytes + 4*valueSlotBytes})

	if result.Stdout != "" {
		t.Fatalf("expected no stdout, got %q", result.Stdout)
	}
	if len(result.Diagnostics) != 1 || result.Diagnostics[0].Code != "LIMIT_VALUE_ALLOCATION_EXCEEDED" {
		t.Fatalf("expected value allocation limit diagnostic, got %+v", result.Diagnostics)
	}
}

func TestRunOptionsMaxValueAllocationBytesStopsEmptyArrayAllocation(t *testing.T) {
	typeSlotBytes := int64(unsafe.Sizeof(Type{}))
	result := Run(`var values array[int] = []
write "unreachable"`, "", RunOptions{MaxValueAllocationBytes: typeSlotBytes - 1})

	if result.Stdout != "" {
		t.Fatalf("expected no stdout, got %q", result.Stdout)
	}
	if len(result.Diagnostics) != 1 || result.Diagnostics[0].Code != "LIMIT_VALUE_ALLOCATION_EXCEEDED" {
		t.Fatalf("expected value allocation limit diagnostic, got %+v", result.Diagnostics)
	}
	if result.Diagnostics[0].Line != 1 {
		t.Fatalf("expected empty array allocation to fail on line 1, got %+v", result.Diagnostics[0])
	}
}

func TestRunOptionsMaxValueAllocationBytesStopsNestedClone(t *testing.T) {
	valueSlotBytes := int64(unsafe.Sizeof(Value{}))
	typeSlotBytes := int64(unsafe.Sizeof(Type{}))
	result := Run(`var nested array[array[int]] = [[1]]
var copied array[array[int]] = nested
write copied`, "", RunOptions{MaxValueAllocationBytes: 2*typeSlotBytes + 6*valueSlotBytes})

	if result.Stdout != "" {
		t.Fatalf("expected no stdout, got %q", result.Stdout)
	}
	if len(result.Diagnostics) != 1 || result.Diagnostics[0].Code != "LIMIT_VALUE_ALLOCATION_EXCEEDED" {
		t.Fatalf("expected value allocation limit diagnostic, got %+v", result.Diagnostics)
	}
}

func TestRunOptionsMaxValueAllocationBytesRejectsNegativeLimit(t *testing.T) {
	diagnostics := make([]Diagnostic, 0, 1)
	result := Run(`write "unreachable"`, "", RunOptions{
		MaxValueAllocationBytes: -1,
		OnDiagnostic: func(diagnostic Diagnostic) {
			diagnostics = append(diagnostics, diagnostic)
		},
	})

	if result.Stdout != "" {
		t.Fatalf("expected no stdout, got %q", result.Stdout)
	}
	if len(result.Diagnostics) != 1 || result.Diagnostics[0].Code != "LIMIT_INVALID_VALUE_ALLOCATION_LIMIT" {
		t.Fatalf("expected invalid value allocation limit diagnostic, got %+v", result.Diagnostics)
	}
	if len(diagnostics) != 1 || diagnostics[0].Code != "LIMIT_INVALID_VALUE_ALLOCATION_LIMIT" {
		t.Fatalf("expected invalid limit callback, got %+v", diagnostics)
	}
}

func TestRunOptionsZeroMaxValueAllocationBytesRemainsUnlimited(t *testing.T) {
	result := Run(`var value string = "a"
value = value + value
write value`, "", RunOptions{MaxValueAllocationBytes: 0})

	if result.Stdout != "aa" || len(result.Diagnostics) != 0 {
		t.Fatalf("expected unlimited value allocations, got %+v", result)
	}
}

func TestRunOptionsMaxValueAllocationBytesChargesQuotedInputDecoding(t *testing.T) {
	source := `var value string
read value
write value`

	boundary := Run(source, `"\u0061"`, RunOptions{MaxValueAllocationBytes: 25})
	if boundary.Stdout != "a" || len(boundary.Diagnostics) != 0 {
		t.Fatalf("expected quoted input allocation boundary to succeed, got %+v", boundary)
	}

	exceeded := Run(source, `"\u0061"`, RunOptions{MaxValueAllocationBytes: 24})
	if exceeded.Stdout != "" {
		t.Fatalf("expected no stdout, got %q", exceeded.Stdout)
	}
	if len(exceeded.Diagnostics) != 1 || exceeded.Diagnostics[0].Code != "LIMIT_VALUE_ALLOCATION_EXCEEDED" {
		t.Fatalf("expected value allocation limit diagnostic, got %+v", exceeded.Diagnostics)
	}
}

func TestRunOptionsMaxValueAllocationBytesBoundsStringPush(t *testing.T) {
	source := `var value string = "a"
push value, "b"
write value`
	runeBytes := int64(unsafe.Sizeof(rune(0)))
	singleRuneValueBytes := runeBytes + 1
	pushStagingBytes := int64(unsafe.Sizeof(Value{}))
	growthBackingBytes := 4 * runeBytes
	replacementStringBytes := int64(2)
	exactBudget := 5*singleRuneValueBytes + pushStagingBytes + growthBackingBytes + replacementStringBytes

	t.Run("exact boundary", func(t *testing.T) {
		result, chunks, diagnostics := runWithValueAllocationEvents(source, exactBudget)
		requireValueAllocationSuccess(t, result, chunks, diagnostics, "ab")
	})

	t.Run("growth backing exhausted", func(t *testing.T) {
		budget := 5*singleRuneValueBytes + pushStagingBytes + growthBackingBytes - 1
		result, chunks, diagnostics := runWithValueAllocationEvents(source, budget)
		requireValueAllocationLimit(t, result, chunks, diagnostics, 2)
	})

	t.Run("replacement string exhausted", func(t *testing.T) {
		result, chunks, diagnostics := runWithValueAllocationEvents(source, exactBudget-1)
		requireValueAllocationLimit(t, result, chunks, diagnostics, 2)
	})
}

func TestRunOptionsMaxValueAllocationBytesBoundsStringPop(t *testing.T) {
	source := `var value string = "ab"
pop value
write value`
	runeBytes := int64(unsafe.Sizeof(rune(0)))
	twoRuneValueBytes := 2*runeBytes + 2
	replacementStringBytes := int64(1)
	exactBudget := 3*twoRuneValueBytes + replacementStringBytes

	t.Run("exact boundary", func(t *testing.T) {
		result, chunks, diagnostics := runWithValueAllocationEvents(source, exactBudget)
		requireValueAllocationSuccess(t, result, chunks, diagnostics, "a")
	})

	t.Run("replacement string exhausted", func(t *testing.T) {
		result, chunks, diagnostics := runWithValueAllocationEvents(source, exactBudget-1)
		requireValueAllocationLimit(t, result, chunks, diagnostics, 2)
	})
}

func TestRunOptionsMaxValueAllocationBytesBoundsStringIndexedReplacement(t *testing.T) {
	source := `var value string = "ab"
value[0] = "x"
write value`
	runeBytes := int64(unsafe.Sizeof(rune(0)))
	singleRuneValueBytes := runeBytes + 1
	twoRuneValueBytes := 2*runeBytes + 2
	replacementStringBytes := int64(2)
	exactBudget := 3*twoRuneValueBytes + 3*singleRuneValueBytes + replacementStringBytes

	t.Run("exact boundary", func(t *testing.T) {
		result, chunks, diagnostics := runWithValueAllocationEvents(source, exactBudget)
		requireValueAllocationSuccess(t, result, chunks, diagnostics, "xb")
	})

	t.Run("single rune temporary exhausted", func(t *testing.T) {
		budgetBeforeTemporary := 3*twoRuneValueBytes + singleRuneValueBytes
		result, chunks, diagnostics := runWithValueAllocationEvents(source, budgetBeforeTemporary+singleRuneValueBytes-1)
		requireValueAllocationLimit(t, result, chunks, diagnostics, 2)
	})

	t.Run("replacement string exhausted", func(t *testing.T) {
		result, chunks, diagnostics := runWithValueAllocationEvents(source, exactBudget-1)
		requireValueAllocationLimit(t, result, chunks, diagnostics, 2)
	})
}

func TestRunOptionsContextCancellationPrecedesValueAllocationLimit(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	stdoutCallbacks := 0
	diagnosticCallbacks := 0

	result := Run(`write "unreachable"`, "", RunOptions{
		Context:                 ctx,
		MaxValueAllocationBytes: -1,
		OnStdoutChunk: func(string, int64) {
			stdoutCallbacks++
		},
		OnDiagnostic: func(Diagnostic) {
			diagnosticCallbacks++
		},
	})

	if result.Stdout != "" || len(result.Diagnostics) != 0 {
		t.Fatalf("expected cancellation without diagnostics, got %+v", result)
	}
	if stdoutCallbacks != 0 || diagnosticCallbacks != 0 {
		t.Fatalf("expected no callbacks, got stdout=%d diagnostics=%d", stdoutCallbacks, diagnosticCallbacks)
	}
}

func TestRunOptionsValueAllocationLimitStopsLaterStdoutCallbacks(t *testing.T) {
	chunks := make([]string, 0, 1)
	diagnostics := make([]Diagnostic, 0, 1)

	result := Run(`write "a", "b", "c"`, "", RunOptions{
		MaxValueAllocationBytes: 5,
		OnStdoutChunk: func(chunk string, _ int64) {
			chunks = append(chunks, chunk)
		},
		OnDiagnostic: func(diagnostic Diagnostic) {
			diagnostics = append(diagnostics, diagnostic)
		},
	})

	if result.Stdout != "a" || strings.Join(chunks, "") != "a" {
		t.Fatalf("expected only pre-limit output, got stdout=%q chunks=%#v", result.Stdout, chunks)
	}
	if len(result.Diagnostics) != 1 || result.Diagnostics[0].Code != "LIMIT_VALUE_ALLOCATION_EXCEEDED" {
		t.Fatalf("expected value allocation limit diagnostic, got %+v", result.Diagnostics)
	}
	if len(diagnostics) != 1 || diagnostics[0].Code != "LIMIT_VALUE_ALLOCATION_EXCEEDED" {
		t.Fatalf("expected one limit callback, got %+v", diagnostics)
	}
}

func TestRunOptionsMaxOutputBytesBoundsArrayFormatting(t *testing.T) {
	result := Run(`var values array[int] = [0]
var i int = 1
while i < 1000 {
    push values, i
    i = i + 1
}
write values`, "", RunOptions{
		MaxOutputBytes: 7,
		MaxSteps:       10_000,
	})

	if result.Stdout != "[0, 1, " {
		t.Fatalf("unexpected bounded array output: %q with diagnostics %+v", result.Stdout, result.Diagnostics)
	}
	if len(result.Diagnostics) != 1 || result.Diagnostics[0].Code != "LIMIT_OUTPUT_EXCEEDED" {
		t.Fatalf("expected output limit diagnostic, got %+v", result.Diagnostics)
	}
}

func TestRunOptionsContextStopsSilentExecution(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	time.AfterFunc(5*time.Millisecond, cancel)
	startedAt := time.Now()

	result := Run(`while true {}`, "", RunOptions{
		Context:  ctx,
		MaxSteps: 1 << 62,
		Timeout:  time.Minute,
	})

	if ctx.Err() != context.Canceled {
		t.Fatalf("expected canceled context, got %v", ctx.Err())
	}
	if time.Since(startedAt) > time.Second {
		t.Fatal("expected context cancellation to stop silent execution promptly")
	}
	if result.Stdout != "" || len(result.Diagnostics) != 0 {
		t.Fatalf("expected cancellation without program diagnostics, got %+v", result)
	}
}

func TestRunOptionsContextRemainsCanceledAfterDiagnosticCallback(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	result := Run(`write 1 / 0`, "", RunOptions{
		Context: ctx,
		OnDiagnostic: func(Diagnostic) {
			cancel()
		},
	})

	if ctx.Err() != context.Canceled {
		t.Fatalf("expected canceled context, got %v", ctx.Err())
	}
	if len(result.Diagnostics) != 1 || result.Diagnostics[0].Code != "RUNTIME_DIV_ZERO" {
		t.Fatalf("expected runtime diagnostic, got %+v", result.Diagnostics)
	}
}

func TestRunOptionsContextCancellationFromStdoutStopsRemainingChunks(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	chunks := make([]string, 0, 2)

	result := Run(`write "first", "second"`, "", RunOptions{
		Context: ctx,
		OnStdoutChunk: func(chunk string, _ int64) {
			chunks = append(chunks, chunk)
			cancel()
		},
	})

	if strings.Join(chunks, "") != "first" {
		t.Fatalf("expected cancellation after the first chunk, got %#v", chunks)
	}
	if result.Stdout != "first" || len(result.Diagnostics) != 0 {
		t.Fatalf("expected partial output without diagnostics, got %+v", result)
	}
}

func TestRunOptionsContextCancellationTakesPrecedenceDuringOutputTruncation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	result := Run(`write "abc"`, "", RunOptions{
		Context:        ctx,
		MaxOutputBytes: 2,
		OnStdoutChunk: func(string, int64) {
			cancel()
		},
	})

	if result.Stdout != "ab" {
		t.Fatalf("expected bounded partial output, got %q", result.Stdout)
	}
	if len(result.Diagnostics) != 0 {
		t.Fatalf("expected cancellation to suppress the output diagnostic, got %+v", result.Diagnostics)
	}
}

func TestRunOptionsCallbackPanicsPropagate(t *testing.T) {
	panicValue := "callback failed"
	defer func() {
		if recovered := recover(); recovered != panicValue {
			t.Fatalf("expected callback panic %q, got %v", panicValue, recovered)
		}
	}()

	Run(`write "abc"`, "", RunOptions{
		OnStdoutChunk: func(string, int64) {
			panic(panicValue)
		},
	})
}

func TestRunOptionsDiagnosticCallbackPanicsPropagate(t *testing.T) {
	panicValue := "diagnostic callback failed"
	defer func() {
		if recovered := recover(); recovered != panicValue {
			t.Fatalf("expected callback panic %q, got %v", panicValue, recovered)
		}
	}()

	Run(`write 1 / 0`, "", RunOptions{
		OnDiagnostic: func(Diagnostic) {
			panic(panicValue)
		},
	})
}

func TestOutputExpressionReadsIdentifierWithoutCloning(t *testing.T) {
	elementType := Type{Kind: TypeInt}
	value := Value{
		Type:  Type{Kind: TypeArray, Elem: &elementType},
		Array: []Value{{Type: elementType, Int: 1}},
	}
	interpreter := &interpreter{
		scopes: []map[string]runtimeVar{{
			"values": {value: value},
		}},
	}

	result, ok := interpreter.evalOutputExpr(&IdentifierExpr{Name: "values"})

	if !ok {
		t.Fatal("expected output identifier evaluation to succeed")
	}
	if &result.Array[0] != &value.Array[0] {
		t.Fatal("expected output evaluation to share the read-only array backing")
	}

	largeArray := make([]Value, 100_000)
	for index := range largeArray {
		largeArray[index] = Value{Type: elementType, Int: int64(index)}
	}
	interpreter.scopes[0]["values"] = runtimeVar{value: Value{
		Type:  Type{Kind: TypeArray, Elem: &elementType},
		Array: largeArray,
	}}
	allocations := testing.AllocsPerRun(10, func() {
		result, ok = interpreter.evalOutputExpr(&IdentifierExpr{Name: "values"})
		if !ok || len(result.Array) != len(largeArray) {
			panic("output identifier evaluation failed")
		}
	})
	if allocations > 10 {
		t.Fatalf("expected bounded output evaluation allocations, got %.0f", allocations)
	}
}
