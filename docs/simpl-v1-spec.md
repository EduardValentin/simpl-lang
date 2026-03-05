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
- Input: `read name`
- Output: `write Expr (, Expr)*`
- If / else-if / else
- While
- For: `for i from Expr until Expr step Expr { ... }`
- Block: `{ ... }`

## Expressions
- Literals: int, float, bool, string, array literal
- Identifier
- Indexing: `arr[idx]`
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

Limits are configurable through `RunOptions`.

## Comments
- Only line comments are supported: `// ...`
