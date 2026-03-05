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
		return Value{Type: t, String: ""}
	case TypeArray:
		return Value{Type: t, Array: make([]Value, 0)}
	default:
		return Value{Type: invalidType()}
	}
}

func cloneValue(v Value) Value {
	out := v
	if v.Type.Kind == TypeArray {
		out.Array = make([]Value, len(v.Array))
		for i := range v.Array {
			out.Array[i] = cloneValue(v.Array[i])
		}
	}
	return out
}

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
				return Value{Type: t, String: parsed}, nil
			}
		}
		return Value{Type: t, String: token}, nil
	default:
		return Value{}, fmt.Errorf("read only supports scalar variables")
	}
}
