package simpl

// maxNestingDepth bounds recursive-descent depth shared across the parser,
// checker, and interpreter.
//
// None of the three stages otherwise limits how deeply nested source can
// be: Go call-stack depth grows with the nesting of parentheses, unary/size
// prefix chains, array literals, indexing, array types, and blocks. On
// arm64/Go 1.25.4, a single expression built from roughly 1MB of chained
// unary operators (one nesting level per byte) reliably triggers a fatal
// "stack overflow" runtime error. That error is a fatal Go runtime error,
// not a panic: recover() cannot intercept it, and it terminates the entire
// process (every goroutine) immediately. The exact byte threshold is
// architecture- and compiler-dependent, but the nesting depth involved is
// on the order of several hundred thousand levels.
//
// Real Simpl programs are small teaching exercises; they rarely combine
// more than a handful of nesting levels across blocks and expressions.
// 1000 leaves at least 25-50x headroom above any plausible legitimate (or
// legitimately generated) program, while staying multiple orders of
// magnitude below the nesting depth empirically needed to overflow the
// stack.
//
// Each stage counts one unit of depth per genuine nesting level in its own
// representation (one paren, one prefix operator, one array/index bracket,
// one array-type level, or one block) rather than per internal helper-call
// frame, so the same source construct is measured consistently across the
// parser, checker, and interpreter.
const maxNestingDepth = 1000

// maxDepthMessage and maxDepthHint are shared across the parser, checker,
// and interpreter so the diagnostic reads the same regardless of which
// stage detects the excess nesting.
const (
	maxDepthCode     = "LIMIT_MAX_DEPTH_EXCEEDED"
	maxDepthCategory = "limit"
	maxDepthMessage  = "Source is nested too deeply to process."
	maxDepthHint     = "Reduce parenthesis, unary-operator, array, index, type, or block nesting."
)

// Diagnostic accumulation guarantee once the depth limit fires:
//
// Elsewhere, the checker's contract is to keep walking and report every
// independent error it finds (see AGENTS.md's diagnostics contract), and
// the parser/interpreter follow the same spirit within their own stages.
// The depth-limit guard breaks that contract on purpose, once it fires:
// each stage's depthReported/halted flag makes every sibling and
// subsequent node a fast no-op rather than walking it for its own,
// unrelated errors. That is deliberate — the alternative (suppressing
// only within the subtree that actually exceeded the limit, so unrelated
// siblings keep accumulating diagnostics normally) is a materially larger
// change than the problem is worth, given how the limit can only be
// reached by nesting a program more than maxNestingDepth levels deep.
//
// So the actual guarantee is narrower than elsewhere in this codebase:
// once a LIMIT_MAX_DEPTH_EXCEEDED diagnostic has been reported for a
// program, no further diagnostics are guaranteed to be reported for that
// same program, even for errors that are structurally unrelated to the
// excess nesting (for example, an undeclared identifier in a sibling
// `write` value next to a 1000-level-deep parenthesized expression).
// Do not assume full accumulation still holds past that point.
