package runtime

import "strconv"

type boolImpl struct{ defaultImpl }

func (boolImpl) hashImpl(v Value) uint64 {
	if v.Bool() {
		return mix(hashSeedBool, 1)
	}
	return mix(hashSeedBool, 0)
}
func (boolImpl) typ(Value) Value { return typeBool }
func (boolImpl) String(v Value) string {
	return strconv.FormatBool(v.Bool())
}
func (boolImpl) isTruthy(v Value) bool       { return v.Bool() }
func (boolImpl) toAny(v Value) any           { return v.Bool() }
func (boolImpl) toNative(v Value) any        { return v.Bool() }
func (boolImpl) toJSON(v Value) (any, error) { return v.Bool(), nil }
