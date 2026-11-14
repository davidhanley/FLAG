package runtime

import "unsafe"

// impl is the per-type vtable on Value. Methods take Value so immediates can
// still read d. typ is type-of; type is a Go keyword.
type impl interface {
	hash(v Value) string
	typ(v Value) Value
	String(v Value) string
	assoc(v Value, entries ...Value) Value
	dissoc(v Value, keys ...Value) Value
}

type Value struct {
	d    float64
	p    unsafe.Pointer
	tag  ValueTag
	impl impl
}

func packLong(n int64) float64 {
	var d float64
	*(*int64)(unsafe.Pointer(&d)) = n
	return d
}

func makeValue(tag ValueTag, d float64, p unsafe.Pointer) Value {
	return Value{d: d, p: p, tag: tag, impl: implFor(tag)}
}

func (v Value) String() string {
	return v.impl.String(v)
}

// NilValue is the FLAG nil. Use the var directly; do not wrap it in a call.
var NilValue = makeValue(TagNil, 0, nil)
