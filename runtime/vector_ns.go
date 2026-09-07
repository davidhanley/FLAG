package runtime

import "fmt"

func VectorNSVector(values ...Value) Value {
	return NewVector(values...)
}

func VectorNSGet(coll Value, index Value, notFound ...Value) Value {
	if len(notFound) > 1 {
		panic("vector/get expects at most one not-found value")
	}
	i := nonNegativeCount("vector/get", index)
	missing := NilValue()
	if len(notFound) == 1 {
		missing = notFound[0]
	}
	if coll.tag == TagNil {
		return missing
	}
	if coll.tag != TagVector {
		panic("vector/get expects vector Value")
	}
	if i >= coll.VectorLen() {
		return missing
	}
	return VectorGet(coll, i)
}

func VectorNSSet(coll Value, index Value, item Value) Value {
	items := vectorNSArrayOrNil(coll)
	i := vectorNSIndex("vector/set", index, len(items), true)
	n := len(items)
	if i == n {
		n++
	}
	out := make([]Value, n)
	copy(out, items)
	out[i] = item
	return NewVector(out...)
}

func VectorNSAppend(coll Value, values ...Value) Value {
	items := vectorNSArrayOrNil(coll)
	out := make([]Value, 0, len(items)+len(values))
	out = append(out, items...)
	out = append(out, values...)
	return NewVector(out...)
}

func VectorNSPrepend(coll Value, values ...Value) Value {
	items := vectorNSArrayOrNil(coll)
	out := make([]Value, 0, len(items)+len(values))
	out = append(out, values...)
	out = append(out, items...)
	return NewVector(out...)
}

func VectorNSPop(coll Value) Value {
	items := vectorNSArrayOrNil(coll)
	if len(items) == 0 {
		panic("vector/pop expects non-empty vector")
	}
	out := make([]Value, len(items)-1)
	copy(out, items[:len(items)-1])
	return NewVector(out...)
}

func VectorNSInsert(coll Value, index Value, values ...Value) Value {
	items := vectorNSArrayOrNil(coll)
	i := vectorNSIndex("vector/insert", index, len(items), true)
	out := make([]Value, 0, len(items)+len(values))
	out = append(out, items[:i]...)
	out = append(out, values...)
	out = append(out, items[i:]...)
	return NewVector(out...)
}

func VectorNSRemove(coll Value, index Value) Value {
	items := vectorNSArrayOrNil(coll)
	i := vectorNSIndex("vector/remove", index, len(items), false)
	out := make([]Value, 0, len(items)-1)
	out = append(out, items[:i]...)
	out = append(out, items[i+1:]...)
	return NewVector(out...)
}

func vectorNSIndex(op string, index Value, length int, allowEnd bool) int {
	i := nonNegativeCount(op, index)
	max := length
	if allowEnd {
		max = length + 1
	}
	if i >= max {
		panic(fmt.Sprintf("%s index out of bounds: %d", op, i))
	}
	return i
}

func vectorNSArrayOrNil(coll Value) []Value {
	if coll.tag == TagNil {
		return nil
	}
	if coll.tag != TagVector {
		panic("vector/* expects vector Value")
	}
	return coll.VectorValues()
}
