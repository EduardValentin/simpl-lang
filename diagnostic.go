package simpl

import "fmt"

// Diagnostic is a machine-readable issue emitted by the compiler/runtime pipeline.
type Diagnostic struct {
	Code     string `json:"code"`
	Category string `json:"category"`
	Message  string `json:"message"`
	Line     int    `json:"line"`
	Column   int    `json:"column"`
	Hint     string `json:"hint,omitempty"`
}

func newDiagnostic(code, category, message string, pos Position, hint string) Diagnostic {
	line := pos.Line
	col := pos.Column
	if line <= 0 {
		line = 1
	}
	if col <= 0 {
		col = 1
	}
	return Diagnostic{
		Code:     code,
		Category: category,
		Message:  message,
		Line:     line,
		Column:   col,
		Hint:     hint,
	}
}

func formatDiagnostics(diags []Diagnostic) string {
	if len(diags) == 0 {
		return ""
	}
	out := ""
	for i, d := range diags {
		if i > 0 {
			out += "\n"
		}
		if d.Hint != "" {
			out += fmt.Sprintf("%d:%d %s %s (hint: %s)", d.Line, d.Column, d.Code, d.Message, d.Hint)
			continue
		}
		out += fmt.Sprintf("%d:%d %s %s", d.Line, d.Column, d.Code, d.Message)
	}
	return out
}
