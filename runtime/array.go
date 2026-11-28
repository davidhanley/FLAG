package runtime

import (
	"strings"
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
	if !isType[*arrayImpl](v) {
		panic("ArrayLen called on non-array Value")
	}
	return int(v.Long())
}

func (v Value) ArrayValues() []Value {
	if !isType[*arrayImpl](v) {
		panic("ArrayValues called on non-array Value")
	}
	items := v.arrayItems()
	out := make([]Value, v.ArrayLen())
	copy(out, items[:v.ArrayLen()])
	return out
}

func ArrayGet(arrayValue Value, index int) Value {
	if !isType[*arrayImpl](arrayValue) {
		panic("ArrayGet expects array Value")
	}
	if index < 0 || index >= arrayValue.ArrayLen() {
		panic("array index out of range")
	}
	items := arrayValue.arrayItems()
	return items[index]
}

func ArrayAssoc(arrayValue Value, index int, item Value) Value {
	if !isType[*arrayImpl](arrayValue) {
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
	if !isType[*arrayImpl](arrayValue) {
		panic("ArrayAppend expects array Value")
	}
	length := arrayValue.ArrayLen()
	items := arrayValue.arrayItems()
	next := make([]Value, length+1)
	copy(next, items[:length])
	next[length] = item
	return newArrayValue(next, length+1)
}

func ArrayRest(arrayValue Value) Value {
	if !isType[*arrayImpl](arrayValue) {
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
	return Value{d: packLong(int64(length)), impl: &arrayImpl{data: data}}
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
	return asImpl[*arrayImpl](v, "arrayPointer called on non-array Value").data
}

func (v Value) arrayItems() []Value {
	return v.arrayPointer().items
}

type arrayImpl struct {
	defaultImpl
	data *arrayData
}

func (arrayImpl) hashImpl(v Value) uint64 {
	return hashOrdered(hashSeedArray, v.arrayItems()[:v.ArrayLen()])
}
func (arrayImpl) eq(v Value, other Value) bool {
	if !isType[*arrayImpl](other) {
		return false
	}
	return valuesEq(v.arrayItems()[:v.ArrayLen()], other.arrayItems()[:other.ArrayLen()])
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
func (arrayImpl) first(v Value) Value {
	if v.ArrayLen() == 0 {
		return NilValue
	}
	return v.arrayItems()[0]
}
func (arrayImpl) rest(v Value) Value {
	return ArrayRest(v)
}
func (arrayImpl) next(v Value) Value {
	rest := ArrayRest(v)
	return nextIfEmpty(rest, rest.ArrayLen() == 0)
}
func (arrayImpl) last(v Value) Value {
	n := v.ArrayLen()
	if n == 0 {
		return NilValue
	}
	return v.arrayItems()[n-1]
}
func (arrayImpl) conj(v Value, entries ...Value) Value {
	out := v
	for _, item := range entries {
		out = ArrayAppend(out, item)
	}
	return out
}
func (arrayImpl) count(v Value) int      { return v.ArrayLen() }
func (arrayImpl) isEmpty(v Value) bool   { return v.ArrayLen() == 0 }
func (arrayImpl) notEmpty(v Value) Value { return notEmptyIf(v, v.ArrayLen() == 0) }
func (arrayImpl) nth(v Value, i int, missing Value, hasMissing bool) Value {
	return nthAt(v.ArrayLen(), i, missing, hasMissing, func(idx int) Value { return ArrayGet(v, idx) })
}
func (arrayImpl) slowNth(v Value, i int, missing Value, hasMissing bool) Value {
	return slowNthWalk(v, i, missing, hasMissing)
}
func (arrayImpl) drop(v Value, n int) Value {
	length := v.ArrayLen()
	if n >= length {
		return NewArray()
	}
	items := v.arrayItems()
	return newArrayValue(items[n:length], length-n)
}
func (arrayImpl) seqQ(Value) bool { return true }
func (arrayImpl) seqCursor(v Value) seqCursor {
	length := v.ArrayLen()
	return arraySeqCursor(v.arrayItems()[:length])
}
func (arrayImpl) toAny(v Value) any {
	return arrayValueToAny(v)
}
func (arrayImpl) toNative(v Value) any {
	return nativeSeq(v.ArrayValues())
}
func (arrayImpl) toJSON(v Value) (any, error) {
	return arrayToJSONValue(v.ArrayValues())
}
func (arrayImpl) goSeq(v Value) []Value {
	return v.ArrayValues()
}
