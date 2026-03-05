# AGENTS.md

## Repository Purpose
`simpl-lang` implements **Simpl**, an interpreted teaching language designed for beginner algorithm/data-structure exercises.

Canonical Go module path: `github.com/EduardValentin/simpl-lang`.

Primary goals:
- Keep syntax readable and minimal.
- Keep runtime behavior deterministic and safe for online code-submission platforms.
- Provide friendly, structured diagnostics.

## Current Architecture
Pipeline (library package `simpl`):
1. **Lexing**: `lexer.go`
2. **Parsing (AST)**: `parser.go`, `ast.go`
3. **Type checking**: `checker.go`, `types.go`
4. **Runtime interpretation**: `interpreter.go`, `value.go`
5. **Public API**: `main.go` (`Validate`, `Run`, `RunFile`)

CLI:
- `cmd/simpl/main.go`
- Commands:
  - `simpl check <file> [--json]`
  - `simpl run <file> [--stdin <file>] [--max-steps N] [--timeout-ms N] [--json]`

## Language Semantics (V1)
- Arrays are **0-indexed**.
- `for` syntax is: `for i from A until B step S { ... }`
  - `S > 0`: iterate while `i < B`
  - `S < 0`: iterate while `i > B`
  - `S == 0`: error
- `read` is token-based (`stdin` split on whitespace).
- `write` does **not** append newline automatically.
- Typing is strict (no implicit numeric coercion).

## Diagnostics Contract
Every failure should be returned as structured diagnostics:
- `code`
- `category` (`lexer`, `parser`, `type`, `runtime`, `limit`)
- `message`
- `line`, `column`
- optional `hint`

Do not introduce panic-based user-facing failures for source/program errors.

## Runtime Safety Defaults
Defined in `main.go`:
- Timeout: `2s`
- Max steps: `1_000_000`

These defaults are overrideable via `RunOptions` or CLI flags.

## Runtime Streaming Hooks
`RunOptions` also supports optional callbacks used by host platforms:
- `OnStdoutChunk(chunk string, stepsUsed int64)`
- `OnDiagnostic(d Diagnostic)`

Use these for live console/event streaming integrations (for example SSE/WebSocket in `course-platform`).

## Tests
Main tests live at repository root:
- `lexer_v1_test.go`
- `parser_v1_test.go`
- `checker_v1_test.go`
- `runtime_v1_test.go`
- `api_cli_v1_test.go`

Run all tests:
```bash
go test ./...
```

CI should run `go test ./...` on every push/PR.

## Agent Workflow Expectations
When changing syntax/semantics:
1. Update parser/runtime/checker together.
2. Update tests to match the behavior.
3. Update user docs:
   - `README.md`
   - `docs/simpl-v1-spec.md`
   - `docs/errors.md` (if error model changes)
4. Validate via:
   - `go test ./...`
   - CLI smoke checks (`check` + `run`).

## Out of Scope for V1
- Functions/procedures
- Modules/imports
- Objects/maps
- Implicit casts
- IO beyond `read`/`write`
