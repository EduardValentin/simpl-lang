# Simpl V1 Specification

## Program Model
- Single source file.
- Statement-oriented language with block scopes.
- Execution is interpreted after successful lexing, parsing, and type checking.

## Statements
- Variable declaration: `var name Type` or `var name Type = Expr`
- Constant declaration: `const name Type = Expr`
- Assignment: `name = Expr`
- Indexed assignment: `name[idx] = Expr` (supports nested indexes)
- Sequence push: `push name, Expr (, Expr)*`
- Sequence pop: `pop name` or `pop name[idx]`
- Input: `read name`
- Output: `write Expr (, Expr)*`
- If / else-if / else
- While
- For: `for i from Expr until Expr step Expr { ... }`
- Block: `{ ... }`

## Expressions
- Literals: int, float, bool, string, array literal
- Identifier
- Sequence size: `size expr`
- Indexing: `arr[idx]`, `str[idx]`
- Unary: `-expr`, `!expr`
- Binary:
  - Arithmetic: `+ - * / %`
  - Comparison: `> >= < <=`
  - Equality: `== !=`
- Grouping: `(expr)`

## Type System
- Primitive: `int`, `float`, `bool`, `string`
- Composite: `array[T]`
- Strict typing:
  - No implicit conversions.
  - Arithmetic uses matching numeric types (`int` with `int`, `float` with `float`).
  - `%` only for `int`.
  - Conditions must be `bool`.
  - Equality requires identical types.

## Array Rules
- Homogeneous element types.
- Zero-indexed.
- Index type must be `int`.
- Out-of-range indexes are runtime errors.

## String Sequence Rules
- Strings remain type `string`.
- Strings are zero-indexed by Unicode rune.
- `str[idx]` returns a `string` containing exactly one rune.
- `push str, expr1, expr2, ...` appends one rune per pushed string expression.
- Indexed string assignment requires the assigned value to contain exactly one rune.
- `size str` returns the rune count.
- `pop str` removes the last rune.
- `str = str + "chunk"` is also valid and appends a whole string chunk.

## Input/Output Rules
- `read` consumes next whitespace token and parses to target variable type.
- `read` supports scalar types (`int`, `float`, `bool`, `string`).
- `write` concatenates values exactly in argument order with no auto newline.

## For Loop Semantics
`for i from A until B step S`:
- `S` must be non-zero `int`.
- If `S > 0`, iterate while `i < B`.
- If `S < 0`, iterate while `i > B`.

## Runtime Limits
Default limits:
- Max steps: `1_000_000`
- Timeout: `2 seconds`
- Max value allocation bytes: unlimited at `0`; positive values enforce a cumulative runtime value allocation budget; negative values are invalid
- Max output bytes: unlimited at `0`; positive values enforce a byte limit; negative values are invalid

Limits and context cancellation are configurable through `RunOptions`. `MaxValueAllocationBytes` charges requested string and array backing storage plus runtime array type metadata before allocation, including sequence growth, concatenation, recursive clones, quoted string input decoding, and transient replacement storage. Charges are cumulative for the run and are not refunded. The budget excludes source loading and compiler allocations, stdin tokenization, source-shaped index traversal slices, scope maps and stacks, diagnostics, allocator overhead, and trusted host callbacks. Index traversal storage is bounded by parsed source shape; Course hosts bound source and host input independently. This is a semantic value allocation budget, not an exact whole-Go-heap limit. Allocation-size overflow fails with `LIMIT_VALUE_ALLOCATION_EXCEEDED` even when the configured budget is unlimited. Output limits are measured separately in bytes and never return partial UTF-8 code points.

## Nesting Depth Limit
The parser, checker, and interpreter each enforce a shared recursion-depth
limit (`1000` levels) on nested structure: parentheses, unary/`size`
prefix chains, array literals, indexing, array types, and blocks. Source
that nests deeper than the limit is rejected with a `LIMIT_MAX_DEPTH_EXCEEDED`
diagnostic instead of being parsed, type-checked, or executed. This is not
configurable through `RunOptions` — real Simpl programs never approach it,
and it exists purely to keep pathologically nested input from exhausting
the host process's call stack.

## Comments
- Only line comments are supported: `// ...`
