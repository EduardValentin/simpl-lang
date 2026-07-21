package simpl

import (
	"math"
	"testing"
)

func TestGrowthHelpersUseGeometricCapacity(t *testing.T) {
	interp := &interpreter{}
	values := make([]Value, 0, 1)
	values, ok := interp.growRuntimeValueSlice(values, 2, Position{})
	if !ok {
		t.Fatal("expected value growth to succeed")
	}
	if cap(values) != 4 {
		t.Fatalf("unexpected value capacity after first growth: %d", cap(values))
	}
	values = append(values, Value{Type: Type{Kind: TypeInt}, Int: 1}, Value{Type: Type{Kind: TypeInt}, Int: 2})
	values, ok = interp.growRuntimeValueSlice(values, 3, Position{})
	if !ok {
		t.Fatal("expected second value growth to succeed")
	}
	if cap(values) != 8 {
		t.Fatalf("unexpected value capacity after second growth: %d", cap(values))
	}

	runes := make([]rune, 0, 1)
	runes, ok = interp.growRuntimeRuneSlice(runes, 2, Position{})
	if !ok {
		t.Fatal("expected rune growth to succeed")
	}
	if cap(runes) != 4 {
		t.Fatalf("unexpected rune capacity after first growth: %d", cap(runes))
	}
	runes = append(runes, 'a', 'b')
	runes, ok = interp.growRuntimeRuneSlice(runes, 3, Position{})
	if !ok {
		t.Fatal("expected second rune growth to succeed")
	}
	if cap(runes) != 8 {
		t.Fatalf("unexpected rune capacity after second growth: %d", cap(runes))
	}
}

func TestCloneValuePreservesSequenceCapacity(t *testing.T) {
	interp := &interpreter{}
	elemType := Type{Kind: TypeInt}
	arrayType := Type{Kind: TypeArray, Elem: &elemType}

	array := make([]Value, 1, 8)
	array[0] = Value{Type: Type{Kind: TypeInt}, Int: 7}
	clonedArray, ok := interp.cloneValue(Value{Type: arrayType, Array: array}, Position{})
	if !ok {
		t.Fatal("expected array clone to succeed")
	}
	if cap(clonedArray.Array) != 8 {
		t.Fatalf("expected cloned array capacity 8, got %d", cap(clonedArray.Array))
	}

	runes := make([]rune, 1, 8)
	runes[0] = 'a'
	clonedString, ok := interp.cloneValue(Value{Type: Type{Kind: TypeString}, String: "a", Runes: runes}, Position{})
	if !ok {
		t.Fatal("expected string clone to succeed")
	}
	if cap(clonedString.Runes) != 8 {
		t.Fatalf("expected cloned string capacity 8, got %d", cap(clonedString.Runes))
	}
}

func TestValuesEqualStopsDuringNestedTraversal(t *testing.T) {
	elementType := Type{Kind: TypeInt}
	arrayType := Type{Kind: TypeArray, Elem: &elementType}
	values := make([]Value, 1_000)
	for index := range values {
		values[index] = Value{Type: elementType, Int: int64(index)}
	}
	checks := 0

	equal, canceled := valuesEqualWithCancel(
		Value{Type: arrayType, Array: values},
		Value{Type: arrayType, Array: values},
		func() bool {
			checks++
			return checks == 100
		},
	)

	if equal || !canceled {
		t.Fatalf("expected canceled equality traversal, got equal=%t canceled=%t", equal, canceled)
	}
}

func TestValueAllocationSizingRejectsIntegerOverflow(t *testing.T) {
	bytes, ok := checkedAllocationBytes(math.MaxInt, 0)
	if !ok || bytes != 0 {
		t.Fatalf("expected zero-width allocation to use zero bytes, got bytes=%d ok=%t", bytes, ok)
	}
	if _, ok := checkedSequenceLength(math.MaxInt, 1); ok {
		t.Fatal("expected sequence length overflow to be rejected")
	}
	if _, ok := nextGrowthCapacity(math.MaxInt/2+1, math.MaxInt); ok {
		t.Fatal("expected growth capacity overflow to be rejected")
	}
	if _, ok := checkedAllocationBytes(2, math.MaxInt64); ok {
		t.Fatal("expected allocation byte overflow to be rejected")
	}
}

func TestRuntimeGrowthOverflowFailsClosedWithLimitDiagnostic(t *testing.T) {
	interp := &interpreter{}
	position := Position{Line: 7, Column: 9}

	if _, ok := interp.growRuntimeValueSlice(nil, math.MaxInt, position); ok {
		t.Fatal("expected runtime growth overflow to fail")
	}
	if len(interp.diags) != 1 || interp.diags[0].Code != "LIMIT_VALUE_ALLOCATION_EXCEEDED" {
		t.Fatalf("expected value allocation limit diagnostic, got %+v", interp.diags)
	}
	if interp.diags[0].Category != "limit" || interp.diags[0].Line != position.Line || interp.diags[0].Column != position.Column {
		t.Fatalf("expected structured limit position, got %+v", interp.diags[0])
	}
}
