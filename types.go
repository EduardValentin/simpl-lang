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
