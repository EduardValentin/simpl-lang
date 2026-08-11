package simpl

import (
	"fmt"
	"strconv"
	"strings"
)

type Value struct {
	Type   Type
	Int    int64
	Float  float64
	Bool   bool
	String string
	Runes  []rune
	Array  []Value
}

func zeroValue(t Type) Value {
	switch t.Kind {
	case TypeInt:
		return Value{Type: t, Int: 0}
	case TypeFloat:
		return Value{Type: t, Float: 0}
	case TypeBool:
		return Value{Type: t, Bool: false}
	case TypeString:
		return newStringValueOfType(t, "")
	case TypeArray:
		return Value{Type: t, Array: make([]Value, 0)}
	default:
		return Value{Type: invalidType()}
	}
}

// cloneValue recurses through Array to deep-copy nested array values, with
// no depth guard of its own. That is safe today, but only structurally:
// a Value's array nesting can never be deeper than the Type it was built
// from (every array Value is produced from a TypeArray whose Elem chain
// mirrors it), and every Type reaching this function already passed
// through the parser's parseType/checker's checkExpr guards, which cap
// type/array-literal nesting at maxNestingDepth (see limits.go). So this
// recursion is bounded transitively, not by anything enforced here. If a
// future change ever lets a Value's array nesting diverge from its Type's
// nesting — e.g. a mutation path that appends without going through the
// checked/parsed type — that invariant would break silently and this
// recursion would no longer be bounded.
func cloneValue(v Value) Value {
	out := v
	switch v.Type.Kind {
	case TypeString:
		runes := stringRunes(v)
		out.Runes = make([]rune, len(runes), cap(runes))
		copy(out.Runes, runes)
		out.String = string(out.Runes)
	case TypeArray:
		out.Array = make([]Value, len(v.Array))
		if cap(v.Array) > len(v.Array) {
			out.Array = make([]Value, len(v.Array), cap(v.Array))
		}
		for i := range v.Array {
			out.Array[i] = cloneValue(v.Array[i])
		}
	}
	return out
}

// valueToString recurses through Array the same way cloneValue does, and
// relies on the same structural invariant (see cloneValue's doc comment):
// array-value nesting cannot exceed the already-capped type/array-literal
// nesting, so this has no depth guard of its own but is bounded
// transitively rather than by anything enforced here.
func valueToString(v Value) string {
	switch v.Type.Kind {
	case TypeInt:
		return strconv.FormatInt(v.Int, 10)
	case TypeFloat:
		return strconv.FormatFloat(v.Float, 'f', -1, 64)
	case TypeBool:
		if v.Bool {
			return "true"
		}
		return "false"
	case TypeString:
		return v.String
	case TypeArray:
		parts := make([]string, len(v.Array))
		for i := range v.Array {
			parts[i] = valueToString(v.Array[i])
		}
		return "[" + strings.Join(parts, ", ") + "]"
	default:
		return "<invalid>"
	}
}

// valuesEqual recurses through Array the same way cloneValue does, and
// relies on the same structural invariant (see cloneValue's doc comment):
// array-value nesting cannot exceed the already-capped type/array-literal
// nesting, so this has no depth guard of its own but is bounded
// transitively rather than by anything enforced here.
func valuesEqual(a, b Value) bool {
	if !a.Type.Equals(b.Type) {
		return false
	}
	switch a.Type.Kind {
	case TypeInt:
		return a.Int == b.Int
	case TypeFloat:
		return a.Float == b.Float
	case TypeBool:
		return a.Bool == b.Bool
	case TypeString:
		return a.String == b.String
	case TypeArray:
		if len(a.Array) != len(b.Array) {
			return false
		}
		for i := range a.Array {
			if !valuesEqual(a.Array[i], b.Array[i]) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func parseInputToken(token string, t Type) (Value, error) {
	switch t.Kind {
	case TypeInt:
		v, err := strconv.ParseInt(token, 10, 64)
		if err != nil {
			return Value{}, fmt.Errorf("expected int input")
		}
		return Value{Type: t, Int: v}, nil
	case TypeFloat:
		v, err := strconv.ParseFloat(token, 64)
		if err != nil {
			return Value{}, fmt.Errorf("expected float input")
		}
		return Value{Type: t, Float: v}, nil
	case TypeBool:
		v, err := strconv.ParseBool(token)
		if err != nil {
			return Value{}, fmt.Errorf("expected bool input (true/false)")
		}
		return Value{Type: t, Bool: v}, nil
	case TypeString:
		if len(token) >= 2 && token[0] == '"' && token[len(token)-1] == '"' {
			parsed, err := strconv.Unquote(token)
			if err == nil {
				return newStringValueOfType(t, parsed), nil
			}
		}
		return newStringValueOfType(t, token), nil
	default:
		return Value{}, fmt.Errorf("read only supports scalar variables")
	}
}

func newStringValue(s string) Value {
	return newStringValueOfType(Type{Kind: TypeString}, s)
}

func newStringValueOfType(t Type, s string) Value {
	return newStringValueFromRunes(t, []rune(s))
}

func newStringValueFromRunes(t Type, runes []rune) Value {
	return Value{Type: t, String: string(runes), Runes: runes}
}

func stringRunes(v Value) []rune {
	if v.Type.Kind != TypeString {
		return nil
	}
	if v.Runes != nil {
		return v.Runes
	}
	return []rune(v.String)
}

func nextGrowthCapacity(currentCap, needed int) int {
	newCap := currentCap
	if newCap < 4 {
		newCap = 4
	}
	for newCap < needed {
		newCap *= 2
	}
	return newCap
}

func growValueSlice(values []Value, extra int) []Value {
	needed := len(values) + extra
	if needed <= cap(values) {
		return values
	}
	grown := make([]Value, len(values), nextGrowthCapacity(cap(values), needed))
	copy(grown, values)
	return grown
}

func growRuneSlice(runes []rune, extra int) []rune {
	needed := len(runes) + extra
	if needed <= cap(runes) {
		return runes
	}
	grown := make([]rune, len(runes), nextGrowthCapacity(cap(runes), needed))
	copy(grown, runes)
	return grown
}
