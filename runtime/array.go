package runtime

import (
	"strings"
	"unsafe"
)

type arrayData struct {
	items []Value
}

func NewArray(values ...Value) Value {
	items := make([]Value, len(values))
	copy(items, values)
	return newArrayValue(items, len(values))
}

func (v Value) ArrayLen() int {
	if v.tag != TagArray {
		panic("ArrayLen called on non-array Value")
	}
	return int(v.Long())
}

func (v Value) ArrayValues() []Value {
	if v.tag != TagArray {
		panic("ArrayValues called on non-array Value")
	}
	items := v.arrayItems()
	out := make([]Value, v.ArrayLen())
	copy(out, items[:v.ArrayLen()])
	return out
}

func ArrayGet(arrayValue Value, index int) Value {
	if arrayValue.tag != TagArray {
		panic("ArrayGet expects array Value")
	}
	if index < 0 || index >= arrayValue.ArrayLen() {
		panic("array index out of range")
	}
	items := arrayValue.arrayItems()
	return items[index]
}

func ArrayAssoc(arrayValue Value, index int, item Value) Value {
	if arrayValue.tag != TagArray {
		panic("ArrayAssoc expects array Value")
	}
	length := arrayValue.ArrayLen()
	if index < 0 || index >= length {
		panic("array index out of range")
	}
	items := arrayValue.arrayItems()
	next := make([]Value, len(items))
	copy(next, items)
	next[index] = item
	return newArrayValue(next, length)
}

func ArrayAppend(arrayValue Value, item Value) Value {
	if arrayValue.tag != TagArray {
		panic("ArrayAppend expects array Value")
	}
	length := arrayValue.ArrayLen()
	items := arrayValue.arrayItems()

	if length < len(items) {
		items[length] = item
		return newArrayValueWithData(arrayValue.arrayPointer(), length+1)
	}

	nextCapacity := grownCapacity(length)
	next := make([]Value, nextCapacity)
	copy(next, items[:length])
	next[length] = item
	return newArrayValue(next, length+1)
}

func ArrayRest(arrayValue Value) Value {
	if arrayValue.tag != TagArray {
		panic("ArrayRest expects array Value")
	}
	length := arrayValue.ArrayLen()
	items := arrayValue.arrayItems()
	if length == 0 {
		return newArrayValue(items[:0], 0)
	}
	return newArrayValue(items[1:length], length-1)
}

func arrayValueToAny(v Value) []any {
	return valuesToAny(v.ArrayValues())
}

func valuesToAny(values []Value) []any {
	out := make([]any, 0, len(values))
	for _, value := range values {
		out = append(out, ValueToAny(value))
	}
	return out
}

func newArrayValue(items []Value, length int) Value {
	return newArrayValueWithData(&arrayData{items: items}, length)
}

func newArrayValueWithData(data *arrayData, length int) Value {
	if length < 0 || length > len(data.items) {
		panic("invalid array length")
	}
	return makeValue(TagArray, packLong(int64(length)), unsafe.Pointer(data))
}

func grownCapacity(length int) int {
	grown := length + length/2
	if grown < 2 {
		grown = 2
	}
	if grown < length+1 {
		grown = length + 1
	}
	return grown
}

func (v Value) arrayPointer() *arrayData {
	if v.tag != TagArray {
		panic("arrayPointer called on non-array Value")
	}
	if v.p == nil {
		panic("array Value does not contain array pointer")
	}
	return (*arrayData)(v.p)
}

func (v Value) arrayItems() []Value {
	return v.arrayPointer().items
}

type arrayImpl struct{ defaultImpl }

func (arrayImpl) hash(v Value) string {
	values := v.ArrayValues()
	parts := make([]string, 0, len(values))
	for _, value := range values {
		parts = append(parts, valueIdentity(value))
	}
	return "A:[" + strings.Join(parts, ",") + "]"
}
func (arrayImpl) typ(Value) Value { return typeArray }
func (arrayImpl) String(v Value) string {
	values := v.ArrayValues()
	var out strings.Builder
	out.WriteByte('[')
	for i, value := range values {
		if i > 0 {
			out.WriteByte(' ')
		}
		out.WriteString(value.String())
	}
	out.WriteByte(']')
	return out.String()
}
func (arrayImpl) assoc(v Value, entries ...Value) Value {
	return assocIndexed(v, entries, func(coll Value) int { return coll.ArrayLen() }, ArrayAssoc, ArrayAppend)
}
func (arrayImpl) get(v Value, key Value, missing Value) Value {
	if i, ok := associativeIndex(key); ok && i < v.ArrayLen() {
		return ArrayGet(v, i)
	}
	return missing
}
