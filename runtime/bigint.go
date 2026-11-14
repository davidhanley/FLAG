package runtime

type bigIntImpl struct{ defaultImpl }

func (bigIntImpl) hash(v Value) string {
	return "I:" + v.BigInt().String()
}
func (bigIntImpl) typ(Value) Value { return typeBigInt }
func (bigIntImpl) String(v Value) string {
	return v.BigInt().String()
}
