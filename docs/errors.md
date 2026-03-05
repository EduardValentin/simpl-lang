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
- `TYPE_INVALID_INDEX`: Invalid index usage (non-array target or non-int index).
- `TYPE_REDECLARED_IDENTIFIER`: Duplicate declaration in same scope.

## Runtime Errors
- `RUNTIME_DIV_ZERO`: Division or modulo by zero.
- `RUNTIME_INDEX_OOB`: Array index out of bounds.
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
