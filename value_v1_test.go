package simpl

import "testing"

func TestGrowthHelpersUseGeometricCapacity(t *testing.T) {
	values := make([]Value, 0, 1)
	values = growValueSlice(values, 2)
	if cap(values) != 4 {
		t.Fatalf("unexpected value capacity after first growth: %d", cap(values))
	}
	values = append(values, Value{Type: Type{Kind: TypeInt}, Int: 1}, Value{Type: Type{Kind: TypeInt}, Int: 2})
	values = growValueSlice(values, 3)
	if cap(values) != 8 {
		t.Fatalf("unexpected value capacity after second growth: %d", cap(values))
	}

	runes := make([]rune, 0, 1)
	runes = growRuneSlice(runes, 2)
	if cap(runes) != 4 {
		t.Fatalf("unexpected rune capacity after first growth: %d", cap(runes))
	}
	runes = append(runes, 'a', 'b')
	runes = growRuneSlice(runes, 3)
	if cap(runes) != 8 {
		t.Fatalf("unexpected rune capacity after second growth: %d", cap(runes))
	}
}

func TestCloneValuePreservesSequenceCapacity(t *testing.T) {
	elemType := Type{Kind: TypeInt}
	arrayType := Type{Kind: TypeArray, Elem: &elemType}

	array := make([]Value, 1, 8)
	array[0] = Value{Type: Type{Kind: TypeInt}, Int: 7}
	clonedArray := cloneValue(Value{Type: arrayType, Array: array})
	if cap(clonedArray.Array) != 8 {
		t.Fatalf("expected cloned array capacity 8, got %d", cap(clonedArray.Array))
	}

	runes := make([]rune, 1, 8)
	runes[0] = 'a'
	clonedString := cloneValue(newStringValueFromRunes(Type{Kind: TypeString}, runes))
	if cap(clonedString.Runes) != 8 {
		t.Fatalf("expected cloned string capacity 8, got %d", cap(clonedString.Runes))
	}
}
