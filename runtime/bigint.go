package runtime

type bigIntImpl struct{ defaultImpl }

func (bigIntImpl) hash(v Value) string {
	return "I:" + v.BigInt().String()
}
func (bigIntImpl) typ(Value) Value { return typeBigInt }
func (bigIntImpl) String(v Value) string {
	return v.BigInt().String()
}
func (bigIntImpl) toAny(v Value) any    { return v.BigInt() }
func (bigIntImpl) toNative(v Value) any { return v.BigInt() }
func (bigIntImpl) toJSON(v Value) (any, error) {
	return v.BigInt().String(), nil
}
