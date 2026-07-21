package simpl

import (
	"context"
	"fmt"
	"os"
	"time"
)

const (
	defaultMaxSteps = int64(1_000_000)
	defaultTimeout  = 2 * time.Second
)

// RunOptions configures runtime safety limits.
type RunOptions struct {
	Context  context.Context
	MaxSteps int64
	// MaxValueAllocationBytes limits cumulative semantic value allocations, not the whole Go heap.
	// It excludes source loading and compilation, stdin tokenization, source-shaped index traversal
	// slices, scope maps and stacks, diagnostics, allocator overhead, and trusted host callbacks.
	// Index traversal storage is bounded by parsed source shape; Course hosts must bound source and
	// host input independently. Zero is unlimited; negative values are invalid.
	MaxValueAllocationBytes int64
	MaxOutputBytes          int
	Timeout                 time.Duration

	// OnStdoutChunk is called whenever write emits text during execution.
	OnStdoutChunk func(chunk string, stepsUsed int64)
	// OnDiagnostic is called whenever runtime/limit diagnostics are emitted.
	OnDiagnostic func(d Diagnostic)
}

// RunResult is the execution output and any diagnostics.
type RunResult struct {
	Stdout      string       `json:"stdout"`
	Diagnostics []Diagnostic `json:"diagnostics"`
	StepsUsed   int64        `json:"steps_used"`
	TimedOut    bool         `json:"timed_out"`
}

// Validate lexes, parses and type-checks source without executing it.
func Validate(source string) []Diagnostic {
	_, diags := compileSource(source)
	return diags
}

// Run lexes/parses/type-checks and executes source.
func Run(source string, stdin string, opts RunOptions) RunResult {
	opts = normalizeOptions(opts)
	if opts.Context != nil && opts.Context.Err() != nil {
		return RunResult{}
	}
	if diagnostic := runOptionsDiagnostic(opts); diagnostic != nil {
		if opts.OnDiagnostic != nil {
			opts.OnDiagnostic(*diagnostic)
		}
		return RunResult{Diagnostics: []Diagnostic{*diagnostic}}
	}

	prog, diags := compileSource(source)
	if opts.Context != nil && opts.Context.Err() != nil {
		return RunResult{}
	}
	if len(diags) > 0 {
		return RunResult{Diagnostics: diags}
	}

	interp := newInterpreter(stdin, opts)
	return interp.run(prog)
}

// RunFile reads a source file and executes it.
func RunFile(path string, stdin string, opts RunOptions) RunResult {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return RunResult{Diagnostics: []Diagnostic{
			newDiagnostic(
				"RUNTIME_FILE_READ",
				"runtime",
				fmt.Sprintf("Cannot read source file: %v", err),
				Position{Line: 1, Column: 1},
				"Check file path and permissions.",
			),
		}}
	}
	return Run(string(bytes), stdin, opts)
}

func normalizeOptions(opts RunOptions) RunOptions {
	if opts.MaxSteps <= 0 {
		opts.MaxSteps = defaultMaxSteps
	}
	if opts.Timeout <= 0 {
		opts.Timeout = defaultTimeout
	}
	return opts
}

func runOptionsDiagnostic(opts RunOptions) *Diagnostic {
	if opts.MaxValueAllocationBytes < 0 {
		diagnostic := newDiagnostic(
			"LIMIT_INVALID_VALUE_ALLOCATION_LIMIT",
			"limit",
			"Value allocation byte limit cannot be negative.",
			Position{Line: 1, Column: 1},
			"Use zero for unlimited value allocations or a positive byte limit.",
		)
		return &diagnostic
	}
	if opts.MaxOutputBytes < 0 {
		diagnostic := newDiagnostic(
			"LIMIT_INVALID_OUTPUT_LIMIT",
			"limit",
			"Output byte limit cannot be negative.",
			Position{Line: 1, Column: 1},
			"Use zero for unlimited output or a positive byte limit.",
		)
		return &diagnostic
	}
	return nil
}

func compileSource(source string) (*Program, []Diagnostic) {
	tokens, ldiags := lexSource(source)
	if len(ldiags) > 0 {
		return nil, ldiags
	}

	prog, pdiags := parseProgram(tokens)
	if len(pdiags) > 0 {
		return nil, pdiags
	}

	tdiags := checkProgram(prog)
	if len(tdiags) > 0 {
		return nil, tdiags
	}

	return prog, nil
}
