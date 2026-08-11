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
