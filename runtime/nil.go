package runtime

// NilValue is the FLAG nil. Use the var directly; do not wrap it in a call.
var NilValue = Value{impl: nilImpl{}}

type nilImpl struct{ defaultImpl }

func (nilImpl) hashImpl(Value) uint64 { return hashSeedNil }
func (nilImpl) typ(Value) Value       { return typeNil }
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
