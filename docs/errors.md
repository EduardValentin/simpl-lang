# Simpl Diagnostics Reference

## Structure
Every diagnostic includes:
- `code`
- `category`
- `message`
- `line`
- `column`
- optional `hint`

## Lexer Errors
- `LEX_UNKNOWN_CHAR`: Unsupported character in source.
- `LEX_UNTERMINATED_STRING`: String literal missing closing quote or invalid format.
- `LEX_INVALID_INT`: Invalid integer literal.
- `LEX_INVALID_FLOAT`: Invalid float literal.

## Parser Errors
- `PARSE_UNEXPECTED_TOKEN`: Token not valid in the current grammar position.
- `PARSE_EXPECTED_TOKEN`: Required token missing (e.g., `]`, `}`, `=`).

## Type Errors
- `TYPE_MISMATCH`: Expression or assignment type mismatch.
- `TYPE_UNDECLARED_IDENTIFIER`: Variable used before declaration.
- `TYPE_CONST_REASSIGN`: Attempt to assign/read into a constant.
- `TYPE_INVALID_INDEX`: Invalid index usage (non-sequence target or non-int index).
- `TYPE_INVALID_SIZE`: Invalid `size` usage on a non-array/non-string expression.
- `TYPE_INVALID_PUSH`: Invalid `push` usage on a non-array/non-string target.
- `TYPE_INVALID_POP`: Invalid `pop` usage on a non-array/non-string target.
- `TYPE_REDECLARED_IDENTIFIER`: Duplicate declaration in same scope.

## Runtime Errors
- `RUNTIME_DIV_ZERO`: Division or modulo by zero.
- `RUNTIME_INDEX_OOB`: Array/string index out of bounds.
- `RUNTIME_POP_EMPTY`: Attempt to pop from an empty array or string.
- `RUNTIME_READ_EOF`: Not enough input tokens for `read`.
- `RUNTIME_READ_PARSE`: Input token cannot be parsed to target type.
- `RUNTIME_TYPE`: Runtime type safety violation.
- `RUNTIME_UNDECLARED`: Runtime access to undeclared variable.
- `RUNTIME_CONST_REASSIGN`: Runtime assignment into constant target.
- `RUNTIME_INVALID_STEP`: For-loop step evaluated to zero.
- `RUNTIME_FILE_READ`: Source file could not be read.

## Limit Errors
- `LIMIT_STEPS_EXCEEDED`: Instruction step budget exhausted.
- `LIMIT_TIMEOUT`: Wall-clock timeout exceeded.
- `LIMIT_MAX_DEPTH_EXCEEDED`: Source nested too deeply (parentheses, unary/size prefix chains, array literals, indexing, array types, or blocks). Raised by whichever of the parser, checker, or interpreter first walks past the shared nesting-depth limit.

  **Accumulation guarantee is narrower here than elsewhere.** Everywhere
  else, the checker in particular keeps walking after an error and reports
  every independent issue it finds. Once `LIMIT_MAX_DEPTH_EXCEEDED` has
  been reported for a program, that guarantee no longer holds: no further
  diagnostics are guaranteed to be reported for that same program, even
  for errors structurally unrelated to the excess nesting (for example, an
  undeclared identifier next to a 1000-level-deep parenthesized
  expression in the same `write` statement). Treat
  `LIMIT_MAX_DEPTH_EXCEEDED` as reporting only that one problem, not as a
  complete list of everything wrong with the program.
- `LIMIT_VALUE_ALLOCATION_EXCEEDED`: Runtime value allocation budget exhausted or allocation sizing overflowed.
- `LIMIT_INVALID_VALUE_ALLOCATION_LIMIT`: Configured runtime value allocation byte budget is negative.
- `LIMIT_OUTPUT_EXCEEDED`: Configured stdout byte budget exhausted.
- `LIMIT_INVALID_OUTPUT_LIMIT`: Configured stdout byte budget is negative.
