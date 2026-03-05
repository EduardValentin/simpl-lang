# Simpl Lang

Simpl is an interpreted teaching language designed for algorithm and data-structure exercises.

This repository provides:
- A Go library API (`Validate`, `Run`, `RunFile`) for platform integration.
- A CLI (`simpl`) for local validation and execution.

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
- `read name`: consumes the next whitespace-delimited stdin token and parses it to the variable type.
- `write expr1, expr2, ...`: prints values exactly as concatenated text, with no automatic newline.

```simpl
read v1
write "Value: ", v1
```

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

`for i from A until B step S` uses exclusive bounds:
- `S > 0`: loop runs while `i < B`
- `S < 0`: loop runs while `i > B`

### Arrays
```simpl
var a array[int] = [1, 2, 3]
a[1] = 9
write a[1]

var matrix array[array[int]] = [[1, 2], [3, 4]]
matrix[1][0] = 7
write matrix[1][0]
```

### Comments
- Line comments only: `// ...`

### Type Rules
- Strict static typing.
- No implicit coercions (e.g. `int + float` is invalid).

## Runtime Limits
Default limits (overridable via API/CLI):
- Timeout: `2s`
- Max steps: `1,000,000`

## Go API

```go
package example

import (
    "fmt"
    "github.com/EduardValentin/simpl"
)

func main() {
    source := `var x int\nread x\nwrite x`
    result := simpl.Run(source, "42", simpl.RunOptions{})

    if len(result.Diagnostics) > 0 {
        fmt.Println(result.Diagnostics[0].Message)
        return
    }

    fmt.Println(result.Stdout) // 42
}
```

## CLI

```bash
# Validate source only
go run ./cmd/simpl check path/to/file.simpl

# Execute source
go run ./cmd/simpl run path/to/file.simpl --stdin path/to/input.txt

# JSON mode
go run ./cmd/simpl check path/to/file.simpl --json
go run ./cmd/simpl run path/to/file.simpl --json
```

## Error Model
Diagnostics are structured with:
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

For full details, see:
- `docs/simpl-v1-spec.md`
- `docs/errors.md`
