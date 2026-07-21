# Simpl Lang

Simpl is an interpreted teaching language designed for beginner algorithm and data-structure exercises.

Repository: `github.com/EduardValentin/simpl-lang`  
Go module path: `github.com/EduardValentin/simpl-lang`

## What this package provides
- Embeddable Go runtime API for validating and executing Simpl source code.
- Structured diagnostics for lexer/parser/type/runtime/limit errors.
- Runtime safety limits (time, steps, value allocations, and output).
- Optional runtime callbacks for live output/diagnostic streaming.
- CLI for local checks and execution.

## Language Scope (V1)

Note: `for` loops use `until` (exclusive end). The old `to` keyword is not supported.

### Types
- `int`
- `float`
- `bool`
- `string`
- `array[T]` (including nested arrays)

### Declarations
```simpl
var v1 int
var v2 int = 2
var v3 string = "Hello"
var v4 float = 2.14
var v5 array[string] = ["a", "b", "c"]
var v6 array[array[int]] = [[1,2,3], [3,4,5]]

const c1 int = 1
const c2 string = "Hi!"
```

### Input / Output
- `read name`: consumes next whitespace-delimited stdin token and parses by variable type.
- `write expr1, expr2, ...`: prints values exactly as concatenated text (no automatic newline).

### Control Flow
```simpl
if v1 == 2 {
    write "v1 is 2"
} else if v1 < 2 {
    write "v1 is less than 2"
} else {
    write "v1 is greater than 2"
}

var i int = 0
while i <= 5 {
    write i, " "
    i = i + 1
}

for i from 0 until 5 step 2 {
    write i, " "
}
```

`for i from A until B step S` semantics:
- `S > 0`: iterate while `i < B`
- `S < 0`: iterate while `i > B`
- `S == 0`: error

### Arrays
```simpl
var a array[int] = [1, 2, 3]
a[1] = 9
write a[1]

var matrix array[array[int]] = [[1, 2], [3, 4]]
matrix[1][0] = 7
write matrix[1][0]
```

### Sequence Primitives
```simpl
var name string = "abc"
write size name
write name[0]
push name, "d", "e"
name[1] = "x"
pop name

var values array[int] = [1, 2, 3]
push values, 4, 5
write size values
pop values
```

- `size expr` works on arrays and strings and returns `int`.
- `push target, value1, value2, ...` appends elements to a mutable array or string target.
- `pop target` removes the last element from a mutable array or string target.
- Strings are indexed by Unicode rune and `name[i]` has type `string`.
- String indexed assignment requires exactly one character on the right-hand side.
- `push` on strings is character-based: each pushed value must be exactly one character.
- `string + string` still works for chunk concatenation, e.g. `name = name + "abc"`.

### Comments
- Line comments only: `// ...`

### Type Rules
- Strict static typing.
- No implicit coercions (for example `int + float` is invalid).

## Runtime limits
Defaults (overridable via `RunOptions`):
- Timeout: `2s`
- Max steps: `1_000_000`
- Max value allocation bytes: unlimited when set to `0`; positive values enforce a cumulative runtime value allocation budget; negative values are invalid
- Max output bytes: unlimited when set to `0`; positive values enforce a byte limit; negative values are invalid

`MaxValueAllocationBytes` charges requested backing storage and array type metadata before runtime strings, arrays, sequence growth, concatenation, and recursive value clones allocate. The budget also conservatively charges quoted string input decoding, transient allocations, and full replacement capacity during growth, and is not refunded when values are released. It excludes source loading and compiler allocations, stdin tokenization, source-shaped index traversal slices, scope maps and stacks, diagnostics, allocator overhead, and trusted host callbacks. Index traversal storage is bounded by parsed source shape; Course hosts bound source and host input independently. This is a semantic value allocation budget, not an exact whole-Go-heap limit. Exhaustion or allocation-size overflow emits `LIMIT_VALUE_ALLOCATION_EXCEEDED`.

`RunOptions.Context` stops execution when its context is canceled. Cancellation returns stdout already emitted, suppresses later callbacks and limit diagnostics, and remains observable through `Context.Err()`. Callback panics are trusted host failures and propagate to the caller.

## Go API

```go
package example

import (
    "fmt"
    simpl "github.com/EduardValentin/simpl-lang"
)

func main() {
    source := `var x int
read x
write x`

    result := simpl.Run(source, "42", simpl.RunOptions{})
    if len(result.Diagnostics) > 0 {
        fmt.Println(result.Diagnostics[0].Message)
        return
    }

    fmt.Println(result.Stdout) // 42
}
```

### Streaming callbacks (for live console integration)

```go
package example

import (
    "context"
    "fmt"
    simpl "github.com/EduardValentin/simpl-lang"
)

func main() {
    source := `for i from 0 until 5 step 1 { write i, " " }`
    ctx, cancel := context.WithCancel(context.Background())

    result := simpl.Run(source, "", simpl.RunOptions{
        Context:                  ctx,
        MaxValueAllocationBytes: 8 * 1024 * 1024,
        MaxOutputBytes:          64 * 1024,
        OnStdoutChunk: func(chunk string, stepsUsed int64) {
            // Send chunk to SSE/WebSocket client
            fmt.Printf("chunk=%q steps=%d\n", chunk, stepsUsed)
            if chunk == "2" {
                cancel()
            }
        },
        OnDiagnostic: func(d simpl.Diagnostic) {
            // Forward runtime/limit diagnostic event
            fmt.Printf("diag=%s %s\n", d.Code, d.Message)
        },
    })

    fmt.Printf("stdout=%q canceled=%v\n", result.Stdout, ctx.Err())
}
```

## CLI

```bash
# Validate source only
go run ./cmd/simpl check path/to/file.simpl

# Execute source
go run ./cmd/simpl run path/to/file.simpl --stdin path/to/input.txt --max-value-allocation-bytes 8388608 --max-output-bytes 65536

# JSON mode
go run ./cmd/simpl check path/to/file.simpl --json
go run ./cmd/simpl run path/to/file.simpl --json
```

## Diagnostics model
Diagnostics include:
- `code`
- `category` (`lexer`, `parser`, `type`, `runtime`, `limit`)
- `message`
- `line`, `column`
- optional `hint`

Examples:
- `TYPE_MISMATCH`
- `TYPE_CONST_REASSIGN`
- `RUNTIME_DIV_ZERO`
- `RUNTIME_READ_PARSE`
- `LIMIT_STEPS_EXCEEDED`
- `LIMIT_TIMEOUT`
- `LIMIT_VALUE_ALLOCATION_EXCEEDED`

See:
- `docs/simpl-v1-spec.md`
- `docs/errors.md`

## Use from `course-platform`

### Production dependency
In `course-platform`:
```bash
go get github.com/EduardValentin/simpl-lang@v0.4.0
```

### Local development override (optional)
In `course-platform/go.mod`:
```go
replace github.com/EduardValentin/simpl-lang => /Users/trocaneduard/Documents/Personal/simpl-lang
```

## Publishing checklist
1. Ensure tests pass:
```bash
go test ./...
```
2. Commit and push `main`.
3. Create and push a semver tag:
```bash
git tag v0.4.0
git push origin v0.4.0
```
4. Update consuming app (`course-platform`) with `go get ...@v0.4.0`.

## Development
Run all tests:
```bash
go test ./...
```
