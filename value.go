package simpl

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
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

type valueFormatter struct {
	builder      strings.Builder
	remaining    int
	bounded      bool
	truncated    bool
	canceled     bool
	shouldCancel func() bool
}

func valueToStringWithinBytes(value Value, maxBytes int, shouldCancel func() bool) (string, bool, bool) {
	return formatValue(value, maxBytes, true, shouldCancel)
}

func formatValue(value Value, maxBytes int, bounded bool, shouldCancel func() bool) (string, bool, bool) {
	formatter := valueFormatter{
		remaining:    maxBytes,
		bounded:      bounded,
		shouldCancel: shouldCancel,
	}
	formatter.writeValue(value)
	return formatter.builder.String(), formatter.truncated, formatter.canceled
}

func (f *valueFormatter) writeValue(value Value) {
	if f.done() {
		return
	}

	switch value.Type.Kind {
	case TypeInt:
		f.writeString(strconv.FormatInt(value.Int, 10))
	case TypeFloat:
		f.writeString(strconv.FormatFloat(value.Float, 'f', -1, 64))
	case TypeBool:
		if value.Bool {
			f.writeString("true")
			return
		}
		f.writeString("false")
	case TypeString:
		f.writeString(value.String)
	case TypeArray:
		f.writeString("[")
		for index := range value.Array {
			if f.done() {
				return
			}
			if index > 0 {
				f.writeString(", ")
			}
			f.writeValue(value.Array[index])
		}
		f.writeString("]")
	default:
		f.writeString("<invalid>")
	}
}

func (f *valueFormatter) writeString(value string) {
	if f.done() || value == "" {
		return
	}
	if !f.bounded {
		f.builder.WriteString(value)
		return
	}
	if len(value) <= f.remaining {
		f.builder.WriteString(value)
		f.remaining -= len(value)
		return
	}

	prefix := value[:f.remaining]
	for !utf8.ValidString(prefix) {
		prefix = prefix[:len(prefix)-1]
	}
	f.builder.WriteString(prefix)
	f.remaining -= len(prefix)
	f.truncated = true
}

func (f *valueFormatter) done() bool {
	if f.truncated || f.canceled {
		return true
	}
	if f.shouldCancel != nil && f.shouldCancel() {
		f.canceled = true
		return true
	}
	return false
}

func valuesEqualWithCancel(a, b Value, shouldCancel func() bool) (bool, bool) {
	if shouldCancel != nil && shouldCancel() {
		return false, true
	}
	if !a.Type.Equals(b.Type) {
		return false, false
	}
	switch a.Type.Kind {
	case TypeInt:
		return a.Int == b.Int, false
	case TypeFloat:
		return a.Float == b.Float, false
	case TypeBool:
		return a.Bool == b.Bool, false
	case TypeString:
		return a.String == b.String, false
	case TypeArray:
		if len(a.Array) != len(b.Array) {
			return false, false
		}
		for i := range a.Array {
			equal, canceled := valuesEqualWithCancel(a.Array[i], b.Array[i], shouldCancel)
			if canceled || !equal {
				return false, canceled
			}
		}
		return true, false
	default:
		return false, false
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
