package runtime

import "math/big"

type bigIntImpl struct {
	defaultImpl
	n *big.Int
}

func (bigIntImpl) hashImpl(v Value) uint64 {
	return hashInt(v.BigInt())
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
func (bigIntImpl) isNumeric(Value) bool { return true }
func (bigIntImpl) asIndex(v Value) (int, bool) {
	bi := v.BigInt()
	if bi.Sign() < 0 || bi.BitLen() > 62 {
		return 0, false
	}
	return int(bi.Int64()), true
}
func (bigIntImpl) asCount(v Value, name string) int {
	bi := v.BigInt()
	if bi.Sign() < 0 {
		panic(name + " expects non-negative count")
	}
	if bi.BitLen() > 62 {
		panic(name + " count is too large")
	}
	return int(bi.Int64())
}
func (bigIntImpl) add(a, b Value) Value {
	return b.impl.addBigInt(b, a.BigInt())
}
func (bigIntImpl) sub(a, b Value) Value {
	return b.impl.subBigInt(b, a.BigInt())
}
func (bigIntImpl) mul(a, b Value) Value {
	return b.impl.mulBigInt(b, a.BigInt())
}
func (bigIntImpl) div(a, b Value) Value {
	return b.impl.divBigInt(b, a.BigInt())
}
func (bigIntImpl) quot(a, b Value) Value {
	return b.impl.quotBigInt(b, a.BigInt())
}
func (bigIntImpl) cmpNum(a, b Value) int {
	return b.impl.cmpBigInt(b, a.BigInt())
}
func (bigIntImpl) abs(v Value) Value {
	return NewBigIntFromBigInt(new(big.Int).Abs(v.BigInt()))
}
func (bigIntImpl) toDouble(v Value) Value {
	return NewDouble(bigIntToFloat64(v.BigInt()))
}
func (bigIntImpl) numerator(v Value) Value { return v }
func (bigIntImpl) denominator(Value) Value { return NewLong(1) }
func (bigIntImpl) toFloat64(v Value) float64 {
	return bigIntToFloat64(v.BigInt())
}
func (bigIntImpl) toRat(v Value) *big.Rat {
	return intRat(v.BigInt())
}
func (bigIntImpl) toBigInt(v Value) *big.Int {
	return new(big.Int).Set(v.BigInt())
}
func (bigIntImpl) bitLong(v Value, name string) int64 {
	if v.BigInt().IsInt64() {
		return v.BigInt().Int64()
	}
	panic(name + " value out of range for long")
}
func (bigIntImpl) addLong(v Value, n int64) Value {
	return addInts(big.NewInt(n), v.BigInt())
}
func (bigIntImpl) addDouble(v Value, n float64) Value {
	return NewDouble(n + bigIntToFloat64(v.BigInt()))
}
func (bigIntImpl) addRatio(v Value, n *big.Rat) Value {
	return addRats(n, intRat(v.BigInt()))
}
func (bigIntImpl) addBigInt(v Value, n *big.Int) Value {
	return addInts(n, v.BigInt())
}
func (bigIntImpl) subLong(v Value, n int64) Value {
	return subInts(big.NewInt(n), v.BigInt())
}
func (bigIntImpl) subDouble(v Value, n float64) Value {
	return NewDouble(n - bigIntToFloat64(v.BigInt()))
}
func (bigIntImpl) subRatio(v Value, n *big.Rat) Value {
	return subRats(n, intRat(v.BigInt()))
}
func (bigIntImpl) subBigInt(v Value, n *big.Int) Value {
	return subInts(n, v.BigInt())
}
func (bigIntImpl) mulLong(v Value, n int64) Value {
	return mulInts(big.NewInt(n), v.BigInt())
}
func (bigIntImpl) mulDouble(v Value, n float64) Value {
	return NewDouble(n * bigIntToFloat64(v.BigInt()))
}
func (bigIntImpl) mulRatio(v Value, n *big.Rat) Value {
	return mulRats(n, intRat(v.BigInt()))
}
func (bigIntImpl) mulBigInt(v Value, n *big.Int) Value {
	return mulInts(n, v.BigInt())
}
func (bigIntImpl) divLong(v Value, n int64) Value {
	return divRats(longRat(n), intRat(v.BigInt()))
}
func (bigIntImpl) divDouble(v Value, n float64) Value {
	return NewDouble(n / bigIntToFloat64(v.BigInt()))
}
func (bigIntImpl) divRatio(v Value, n *big.Rat) Value {
	return divRats(n, intRat(v.BigInt()))
}
func (bigIntImpl) divBigInt(v Value, n *big.Int) Value {
	return divRats(intRat(n), intRat(v.BigInt()))
}
func (bigIntImpl) quotLong(v Value, n int64) Value {
	return quotRats(longRat(n), intRat(v.BigInt()))
}
func (bigIntImpl) quotDouble(v Value, n float64) Value {
	return quotFloat(n, bigIntToFloat64(v.BigInt()))
}
func (bigIntImpl) quotRatio(v Value, n *big.Rat) Value {
	return quotRats(n, intRat(v.BigInt()))
}
func (bigIntImpl) quotBigInt(v Value, n *big.Int) Value {
	return quotRats(intRat(n), intRat(v.BigInt()))
}
func (bigIntImpl) cmpLong(v Value, n int64) int {
	return compareLongAndBigInt(n, v.BigInt())
}
func (bigIntImpl) cmpDouble(v Value, n float64) int {
	return compareFloat64(n, bigIntToFloat64(v.BigInt()))
}
func (bigIntImpl) cmpRatio(v Value, n *big.Rat) int {
	return n.Cmp(intRat(v.BigInt()))
}
func (bigIntImpl) cmpBigInt(v Value, n *big.Int) int {
	return n.Cmp(v.BigInt())
}
