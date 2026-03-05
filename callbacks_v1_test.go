package simpl

import (
	"testing"
	"time"
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
