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
	get(v Value, key Value, missing Value) Value
	contains(v Value, key Value) bool
	find(v Value, key Value) Value
	keys(v Value) Value
	vals(v Value) Value
	first(v Value) Value
	rest(v Value) Value
	next(v Value) Value
	last(v Value) Value
	conj(v Value, entries ...Value) Value
	count(v Value) int
	isEmpty(v Value) bool
	notEmpty(v Value) Value
	seq(v Value) Value
	nth(v Value, i int, missing Value, hasMissing bool) Value
	slowNth(v Value, i int, missing Value, hasMissing bool) Value
	drop(v Value, n int) Value
	cons(v Value, item Value) Value
	seqQ(v Value) bool
	seqCursor(v Value) seqCursor
	eq(v Value, other Value) bool
	isTruthy(v Value) bool
	isNil(v Value) bool
	compare(v Value, other Value) Value
	toAny(v Value) any
	toNative(v Value) any
	asString(v Value) string
	asIdent(v Value) (string, bool)
	toJSON(v Value) (any, error)
	name(v Value) string
	exMessage(v Value) Value
	goSeq(v Value) []Value
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

type nilImpl struct{ defaultImpl }

func (nilImpl) hash(Value) string { return "N:" }
func (nilImpl) typ(Value) Value   { return typeNil }
func (nilImpl) String(Value) string {
	return ""
}
func (nilImpl) dissoc(Value, ...Value) Value {
	return NilValue
}
func (nilImpl) get(_ Value, _ Value, missing Value) Value {
	return missing
}
func (nilImpl) contains(Value, Value) bool {
	return false
}
func (nilImpl) find(Value, Value) Value {
	return NilValue
}
func (nilImpl) keys(Value) Value   { return NilValue }
func (nilImpl) vals(Value) Value   { return NilValue }
func (nilImpl) first(Value) Value  { return NilValue }
func (nilImpl) next(Value) Value   { return NilValue }
func (nilImpl) count(Value) int    { return 0 }
func (nilImpl) isEmpty(Value) bool { return true }
func (nilImpl) cons(_ Value, item Value) Value {
	return NewList(item)
}
func (nilImpl) nth(_ Value, _ int, missing Value, hasMissing bool) Value {
	if hasMissing {
		return missing
	}
	panic("nth index out of range")
}
func (nilImpl) slowNth(v Value, i int, missing Value, hasMissing bool) Value {
	return slowNthWalk(v, i, missing, hasMissing)
}
func (nilImpl) seqCursor(Value) seqCursor { return emptySeqCursor() }
func (nilImpl) eq(_ Value, other Value) bool {
	return other.impl.isNil(other)
}
func (nilImpl) isTruthy(Value) bool { return false }
func (nilImpl) isNil(Value) bool    { return true }
func (nilImpl) compare(_ Value, other Value) Value {
	if other.impl.isNil(other) {
		return NewLong(0)
	}
	return NewLong(-1)
}
func (nilImpl) toAny(Value) any           { return nil }
func (nilImpl) toNative(Value) any        { return nil }
func (nilImpl) toJSON(Value) (any, error) { return nil, nil }
func (nilImpl) exMessage(Value) Value     { return NilValue }
