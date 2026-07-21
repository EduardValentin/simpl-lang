package simpl

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode/utf8"
	"unsafe"
)

const (
	runeSlotBytes  = int64(unsafe.Sizeof(rune(0)))
	typeSlotBytes  = int64(unsafe.Sizeof(Type{}))
	valueSlotBytes = int64(unsafe.Sizeof(Value{}))
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
		return Value{Type: t}
	case TypeArray:
		return Value{Type: t}
	default:
		return Value{Type: invalidType()}
	}
}

func (i *interpreter) cloneValue(value Value, pos Position) (Value, bool) {
	if !i.reserveValueAllocation(0, true, pos) {
		return Value{}, false
	}

	cloned := value
	switch value.Type.Kind {
	case TypeString:
		runes, ok := i.valueRunes(value, pos)
		if !ok {
			return Value{}, false
		}
		stringBytes, valid := encodedRunesBytes(runes)
		allocationBytes, valid := checkedStringAllocationBytes(cap(runes), stringBytes, valid)
		if !i.reserveValueAllocation(allocationBytes, valid, pos) {
			return Value{}, false
		}
		cloned.Runes = make([]rune, len(runes), cap(runes))
		copy(cloned.Runes, runes)
		cloned.String = string(cloned.Runes)
	case TypeArray:
		array, ok := i.makeRuntimeValueSlice(len(value.Array), cap(value.Array), pos)
		if !ok {
			return Value{}, false
		}
		for index := range value.Array {
			array[index], ok = i.cloneValue(value.Array[index], pos)
			if !ok {
				return Value{}, false
			}
		}
		cloned.Array = array
	}
	return cloned, true
}

func (i *interpreter) newRuntimeStringValue(valueType Type, value string, pos Position) (Value, bool) {
	runeCount, stringBytes, valid := stringStorageSize(value)
	allocationBytes, valid := checkedStringAllocationBytes(runeCount, stringBytes, valid)
	if !i.reserveValueAllocation(allocationBytes, valid, pos) {
		return Value{}, false
	}

	runes := make([]rune, 0, runeCount)
	for _, current := range value {
		runes = append(runes, current)
	}
	return Value{Type: valueType, String: string(runes), Runes: runes}, true
}

func (i *interpreter) newRuntimeSingleRuneString(valueType Type, value rune, pos Position) (Value, bool) {
	stringBytes := utf8.RuneLen(value)
	if stringBytes < 0 {
		stringBytes = utf8.RuneLen(utf8.RuneError)
	}
	allocationBytes, valid := checkedStringAllocationBytes(1, int64(stringBytes), true)
	if !i.reserveValueAllocation(allocationBytes, valid, pos) {
		return Value{}, false
	}

	runes := make([]rune, 1)
	runes[0] = value
	return Value{Type: valueType, String: string(runes), Runes: runes}, true
}

func (i *interpreter) concatenateRuntimeStrings(left, right Value, pos Position) (Value, bool) {
	leftRunes, ok := i.valueRunes(left, pos)
	if !ok {
		return Value{}, false
	}
	rightRunes, ok := i.valueRunes(right, pos)
	if !ok {
		return Value{}, false
	}
	runeCount, valid := checkedSequenceLength(len(leftRunes), len(rightRunes))
	leftBytes, leftValid := encodedRunesBytes(leftRunes)
	rightBytes, rightValid := encodedRunesBytes(rightRunes)
	stringBytes, bytesValid := checkedAllocationByteSum(leftBytes, rightBytes)
	valid = valid && leftValid && rightValid && bytesValid
	allocationBytes, valid := checkedStringAllocationBytes(runeCount, stringBytes, valid)
	if !i.reserveValueAllocation(allocationBytes, valid, pos) {
		return Value{}, false
	}

	combined := make([]rune, 0, runeCount)
	combined = append(combined, leftRunes...)
	combined = append(combined, rightRunes...)
	return Value{Type: Type{Kind: TypeString}, String: string(combined), Runes: combined}, true
}

func (i *interpreter) refreshRuntimeString(value *Value, pos Position) bool {
	stringBytes, valid := encodedRunesBytes(value.Runes)
	if !i.reserveValueAllocation(stringBytes, valid, pos) {
		return false
	}
	value.String = string(value.Runes)
	return true
}

func (i *interpreter) valueRunes(value Value, pos Position) ([]rune, bool) {
	if value.Runes != nil || value.String == "" {
		return value.Runes, true
	}

	runeCount, _, valid := stringStorageSize(value.String)
	runeBytes, valid := checkedAllocationBytes(runeCount, runeSlotBytes)
	if !i.reserveValueAllocation(runeBytes, valid, pos) {
		return nil, false
	}
	runes := make([]rune, 0, runeCount)
	for _, current := range value.String {
		runes = append(runes, current)
	}
	return runes, true
}

func (i *interpreter) makeRuntimeValueSlice(length, capacity int, pos Position) ([]Value, bool) {
	valid := length >= 0 && capacity >= length
	allocationBytes, bytesValid := checkedAllocationBytes(capacity, valueSlotBytes)
	if !i.reserveValueAllocation(allocationBytes, valid && bytesValid, pos) {
		return nil, false
	}
	return make([]Value, length, capacity), true
}

func (i *interpreter) copyRuntimeType(value Type, pos Position) (*Type, bool) {
	if !i.reserveValueAllocation(typeSlotBytes, true, pos) {
		return nil, false
	}
	copied := value
	return &copied, true
}

func (i *interpreter) growRuntimeValueSlice(values []Value, extra int, pos Position) ([]Value, bool) {
	needed, valid := checkedSequenceLength(len(values), extra)
	if !valid {
		i.rejectValueAllocation(pos)
		return nil, false
	}
	if needed <= cap(values) {
		return values, true
	}

	capacity, valid := nextGrowthCapacity(cap(values), needed)
	if !valid {
		i.rejectValueAllocation(pos)
		return nil, false
	}
	grown, ok := i.makeRuntimeValueSlice(len(values), capacity, pos)
	if !ok {
		return nil, false
	}
	copy(grown, values)
	return grown, true
}

func (i *interpreter) growRuntimeRuneSlice(runes []rune, extra int, pos Position) ([]rune, bool) {
	needed, valid := checkedSequenceLength(len(runes), extra)
	if !valid {
		i.rejectValueAllocation(pos)
		return nil, false
	}
	if needed <= cap(runes) {
		return runes, true
	}

	capacity, valid := nextGrowthCapacity(cap(runes), needed)
	if !valid {
		i.rejectValueAllocation(pos)
		return nil, false
	}
	allocationBytes, valid := checkedAllocationBytes(capacity, runeSlotBytes)
	if !i.reserveValueAllocation(allocationBytes, valid, pos) {
		return nil, false
	}
	grown := make([]rune, len(runes), capacity)
	copy(grown, runes)
	return grown, true
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

func parseScalarInputToken(token string, t Type) (Value, error) {
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
	default:
		return Value{}, fmt.Errorf("read only supports scalar variables")
	}
}

func (i *interpreter) parseRuntimeInputToken(token string, valueType Type, pos Position) (Value, error, bool) {
	if valueType.Kind != TypeString {
		value, err := parseScalarInputToken(token, valueType)
		return value, err, true
	}

	value := token
	if len(token) >= 2 && token[0] == '"' && token[len(token)-1] == '"' {
		temporaryBytes, valid := quotedInputTemporaryAllocationBytes(len(token))
		if !i.reserveValueAllocation(temporaryBytes, valid, pos) {
			return Value{}, nil, false
		}
		parsed, err := strconv.Unquote(token)
		if err == nil {
			value = parsed
		}
	}

	runtimeValue, ok := i.newRuntimeStringValue(valueType, value, pos)
	return runtimeValue, nil, ok
}

func quotedInputTemporaryAllocationBytes(length int) (int64, bool) {
	bufferBytes, valid := checkedAllocationBytes(length, 3)
	if !valid {
		return 0, false
	}
	resultBytes, valid := checkedAllocationBytes(length, 1)
	if !valid {
		return 0, false
	}
	return checkedAllocationByteSum(bufferBytes/2, resultBytes)
}

func checkedSequenceLength(length, extra int) (int, bool) {
	if length < 0 || extra < 0 || length > math.MaxInt-extra {
		return 0, false
	}
	return length + extra, true
}

func checkedAllocationBytes(count int, elementBytes int64) (int64, bool) {
	if count < 0 || elementBytes < 0 {
		return 0, false
	}
	if elementBytes == 0 {
		return 0, true
	}
	if count > 0 && int64(count) > math.MaxInt64/elementBytes {
		return 0, false
	}
	return int64(count) * elementBytes, true
}

func checkedAllocationByteSum(left, right int64) (int64, bool) {
	if left < 0 || right < 0 || left > math.MaxInt64-right {
		return 0, false
	}
	return left + right, true
}

func checkedStringAllocationBytes(runeCapacity int, stringBytes int64, valid bool) (int64, bool) {
	if !valid {
		return 0, false
	}
	runeBytes, valid := checkedAllocationBytes(runeCapacity, runeSlotBytes)
	if !valid {
		return 0, false
	}
	return checkedAllocationByteSum(runeBytes, stringBytes)
}

func stringStorageSize(value string) (int, int64, bool) {
	runeCount := 0
	stringBytes := int64(0)
	for _, current := range value {
		if runeCount == math.MaxInt {
			return 0, 0, false
		}
		runeCount++
		currentBytes := utf8.RuneLen(current)
		if currentBytes < 0 {
			currentBytes = utf8.RuneLen(utf8.RuneError)
		}
		var valid bool
		stringBytes, valid = checkedAllocationByteSum(stringBytes, int64(currentBytes))
		if !valid {
			return 0, 0, false
		}
	}
	return runeCount, stringBytes, true
}

func encodedRunesBytes(runes []rune) (int64, bool) {
	stringBytes := int64(0)
	for _, current := range runes {
		currentBytes := utf8.RuneLen(current)
		if currentBytes < 0 {
			currentBytes = utf8.RuneLen(utf8.RuneError)
		}
		var valid bool
		stringBytes, valid = checkedAllocationByteSum(stringBytes, int64(currentBytes))
		if !valid {
			return 0, false
		}
	}
	return stringBytes, true
}

func nextGrowthCapacity(currentCap, needed int) (int, bool) {
	if currentCap < 0 || needed < 0 {
		return 0, false
	}
	newCap := currentCap
	if newCap < 4 {
		newCap = 4
	}
	for newCap < needed {
		if newCap > math.MaxInt/2 {
			return 0, false
		}
		newCap *= 2
	}
	return newCap, true
}
