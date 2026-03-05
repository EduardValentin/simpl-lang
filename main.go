package simpl

import (
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
	MaxSteps int64
	Timeout  time.Duration
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
	prog, diags := compileSource(source)
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
