package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	simpl "github.com/EduardValentin/simpl-lang"
)

func main() {
	if len(os.Args) < 3 {
		printUsage()
		os.Exit(2)
	}

	switch os.Args[1] {
	case "check":
		runCheck(os.Args[2:])
	case "run":
		runRun(os.Args[2:])
	default:
		printUsage()
		os.Exit(2)
	}
}

func runCheck(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "check requires a source file")
		os.Exit(2)
	}
	file := args[0]

	fs := flag.NewFlagSet("check", flag.ContinueOnError)
	jsonOut := fs.Bool("json", false, "Output diagnostics as JSON")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args[1:]); err != nil {
		os.Exit(2)
	}
	if fs.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "check accepts flags only after the source file")
		os.Exit(2)
	}

	bytes, err := os.ReadFile(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot read source file: %v\n", err)
		os.Exit(1)
	}

	diags := simpl.Validate(string(bytes))
	if *jsonOut {
		payload := map[string]any{
			"ok":          len(diags) == 0,
			"diagnostics": diags,
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(payload)
	} else if len(diags) == 0 {
		fmt.Println("OK")
	} else {
		printDiagnostics(diags)
	}

	if len(diags) > 0 {
		os.Exit(1)
	}
}

func runRun(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "run requires a source file")
		os.Exit(2)
	}
	sourceFile := args[0]

	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	stdinFile := fs.String("stdin", "", "Path to file used as stdin token source")
	maxSteps := fs.Int64("max-steps", 0, "Execution step limit")
	timeoutMs := fs.Int64("timeout-ms", 0, "Execution timeout in milliseconds")
	jsonOut := fs.Bool("json", false, "Output result as JSON")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args[1:]); err != nil {
		os.Exit(2)
	}
	if fs.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "run accepts flags only after the source file")
		os.Exit(2)
	}

	input := ""
	if *stdinFile != "" {
		bytes, err := os.ReadFile(*stdinFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Cannot read stdin file: %v\n", err)
			os.Exit(1)
		}
		input = string(bytes)
	}

	opts := simpl.RunOptions{MaxSteps: *maxSteps}
	if *timeoutMs > 0 {
		opts.Timeout = time.Duration(*timeoutMs) * time.Millisecond
	}

	result := simpl.RunFile(sourceFile, input, opts)
	if *jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(result)
	} else {
		if result.Stdout != "" {
			fmt.Print(result.Stdout)
		}
		if len(result.Diagnostics) > 0 {
			printDiagnostics(result.Diagnostics)
		}
	}

	if len(result.Diagnostics) > 0 {
		os.Exit(1)
	}
}

func printDiagnostics(diags []simpl.Diagnostic) {
	for _, d := range diags {
		if d.Hint != "" {
			fmt.Fprintf(os.Stderr, "%d:%d %s %s (hint: %s)\n", d.Line, d.Column, d.Code, d.Message, d.Hint)
			continue
		}
		fmt.Fprintf(os.Stderr, "%d:%d %s %s\n", d.Line, d.Column, d.Code, d.Message)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintln(os.Stderr, "  simpl check <file> [--json]")
	fmt.Fprintln(os.Stderr, "  simpl run <file> [--stdin <file>] [--max-steps N] [--timeout-ms N] [--json]")
}
