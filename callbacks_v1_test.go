package simpl

import (
	"context"
	"strings"
	"testing"
	"time"
	"unicode/utf8"
)

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
