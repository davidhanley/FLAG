package runtime

import "unsafe"

func NewVector(values ...Value) Value {
	items := make([]Value, len(values))
	copy(items, values)
	return newVectorValue(items, len(values))
}

func (v Value) VectorLen() int {
	if v.tag != TagVector {
		panic("VectorLen called on non-vector Value")
	}
	return int(v.Long())
}

func (v Value) VectorValues() []Value {
	if v.tag != TagVector {
		panic("VectorValues called on non-vector Value")
	}
	items := v.vectorItems()
	out := make([]Value, v.VectorLen())
	copy(out, items[:v.VectorLen()])
	return out
}

func VectorGet(vectorValue Value, index int) Value {
	if vectorValue.tag != TagVector {
		panic("VectorGet expects vector Value")
	}
	if index < 0 || index >= vectorValue.VectorLen() {
		panic("vector index out of range")
	}
	return vectorValue.vectorItems()[index]
}

func VectorAssoc(vectorValue Value, index int, item Value) Value {
	if vectorValue.tag != TagVector {
		panic("VectorAssoc expects vector Value")
	}
	length := vectorValue.VectorLen()
	if index < 0 || index >= length {
		panic("vector index out of range")
	}
	items := vectorValue.vectorItems()
	next := make([]Value, len(items))
	copy(next, items)
	next[index] = item
	return newVectorValue(next, length)
}

func VectorRest(vectorValue Value) Value {
	if vectorValue.tag != TagVector {
		panic("VectorRest expects vector Value")
	}
	length := vectorValue.VectorLen()
	items := vectorValue.vectorItems()
	if length == 0 {
		return newVectorValue(items[:0], 0)
	}
	return newVectorValue(items[1:length], length-1)
}

func VectorAppend(vectorValue Value, item Value) Value {
	if vectorValue.tag != TagVector {
		panic("VectorAppend expects vector Value")
	}
	length := vectorValue.VectorLen()
	items := vectorValue.vectorItems()
	if length < len(items) {
		items[length] = item
		return newVectorValueWithData(vectorValue.vectorPointer(), length+1)
	}
	nextCapacity := grownCapacity(length)
	next := make([]Value, nextCapacity)
	copy(next, items[:length])
	next[length] = item
	return newVectorValue(next, length+1)
}

func newVectorValue(items []Value, length int) Value {
	return newVectorValueWithData(&arrayData{items: items}, length)
}

func newVectorValueWithData(data *arrayData, length int) Value {
	if length < 0 || length > len(data.items) {
		panic("invalid vector length")
	}
	out := Value{p: unsafe.Pointer(data), tag: TagVector}
	*(*int64)(unsafe.Pointer(&out.d)) = int64(length)
	return out
}

func (v Value) vectorPointer() *arrayData {
	if v.tag != TagVector {
		panic("vectorPointer called on non-vector Value")
	}
	if v.p == nil {
		panic("vector Value does not contain vector pointer")
	}
	return (*arrayData)(v.p)
}

func (v Value) vectorItems() []Value {
	return v.vectorPointer().items
}
