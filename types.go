package simpl

import "fmt"

type TypeKind string

const (
	TypeInt     TypeKind = "int"
	TypeFloat   TypeKind = "float"
	TypeBool    TypeKind = "bool"
	TypeString  TypeKind = "string"
	TypeArray   TypeKind = "array"
	TypeInvalid TypeKind = "invalid"
)

type Type struct {
	Kind TypeKind
	Elem *Type
}

// Equals recurses through Elem for array types, with no depth guard of
// its own. That is safe today, but only structurally: every Type reaching
// this method was built by parseType or by the checker's array-literal
// inference, both of which cap nesting at maxNestingDepth (see
// limits.go) before a Type this deep could ever exist. This recursion is
// bounded transitively by that upstream cap, not by anything enforced
// here — if a future code path ever constructs a Type by some means that
// bypasses parseType/checkExpr, that invariant would break silently.
func (t Type) Equals(other Type) bool {
	if t.Kind != other.Kind {
		return false
	}
	if t.Kind != TypeArray {
		return true
	}
	if t.Elem == nil || other.Elem == nil {
		return t.Elem == nil && other.Elem == nil
	}
	return t.Elem.Equals(*other.Elem)
}

// String recurses through Elem the same way Equals does, and relies on
// the same structural invariant (see Equals's doc comment): Type nesting
// is already capped upstream by parseType/checkExpr, so this has no depth
// guard of its own but is bounded transitively rather than by anything
// enforced here.
func (t Type) String() string {
	switch t.Kind {
	case TypeArray:
		if t.Elem == nil {
			return "array[?]"
		}
		return fmt.Sprintf("array[%s]", t.Elem.String())
	case TypeInt, TypeFloat, TypeBool, TypeString:
		return string(t.Kind)
	default:
		return "invalid"
	}
}

func invalidType() Type {
	return Type{Kind: TypeInvalid}
}
