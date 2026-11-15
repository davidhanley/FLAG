package runtime

import (
	"math/big"
	"strconv"
)

type longImpl struct{ defaultImpl }

func (longImpl) hash(v Value) string {
	return "L:" + strconv.FormatInt(v.Long(), 10)
}
func (longImpl) typ(Value) Value { return typeInt }
func (longImpl) String(v Value) string {
	return strconv.FormatInt(v.Long(), 10)
}
func (longImpl) toAny(v Value) any           { return v.Long() }
func (longImpl) toNative(v Value) any        { return v.Long() }
func (longImpl) toJSON(v Value) (any, error) { return v.Long(), nil }
func (longImpl) isNumeric(Value) bool        { return true }
func (longImpl) add(a, b Value) Value        { return b.impl.addLong(b, a.Long()) }
func (longImpl) sub(a, b Value) Value        { return b.impl.subLong(b, a.Long()) }
func (longImpl) mul(a, b Value) Value        { return b.impl.mulLong(b, a.Long()) }
func (longImpl) div(a, b Value) Value        { return b.impl.divLong(b, a.Long()) }
func (longImpl) quot(a, b Value) Value       { return b.impl.quotLong(b, a.Long()) }
func (longImpl) cmpNum(a, b Value) int       { return b.impl.cmpLong(b, a.Long()) }
func (longImpl) abs(v Value) Value {
	n := v.Long()
	if n >= 0 {
		return v
	}
	if n == minInt64 {
		return NewBigIntFromBigInt(new(big.Int).Abs(big.NewInt(n)))
	}
	return NewLong(-n)
}
func (longImpl) toDouble(v Value) Value          { return NewDouble(float64(v.Long())) }
func (longImpl) numerator(v Value) Value         { return v }
func (longImpl) denominator(Value) Value         { return NewLong(1) }
func (longImpl) toFloat64(v Value) float64       { return float64(v.Long()) }
func (longImpl) toRat(v Value) *big.Rat          { return longRat(v.Long()) }
func (longImpl) toBigInt(v Value) *big.Int       { return big.NewInt(v.Long()) }
func (longImpl) bitLong(v Value, _ string) int64 { return v.Long() }
func (longImpl) addLong(v Value, n int64) Value  { return addLongs(n, v.Long()) }
func (longImpl) addDouble(v Value, n float64) Value {
	return NewDouble(n + float64(v.Long()))
}
func (longImpl) addRatio(v Value, n *big.Rat) Value {
	return addRats(n, longRat(v.Long()))
}
func (longImpl) addBigInt(v Value, n *big.Int) Value {
	return addInts(n, big.NewInt(v.Long()))
}
func (longImpl) subLong(v Value, n int64) Value { return subLongs(n, v.Long()) }
func (longImpl) subDouble(v Value, n float64) Value {
	return NewDouble(n - float64(v.Long()))
}
func (longImpl) subRatio(v Value, n *big.Rat) Value {
	return subRats(n, longRat(v.Long()))
}
func (longImpl) subBigInt(v Value, n *big.Int) Value {
	return subInts(n, big.NewInt(v.Long()))
}
func (longImpl) mulLong(v Value, n int64) Value { return mulLongs(n, v.Long()) }
func (longImpl) mulDouble(v Value, n float64) Value {
	return NewDouble(n * float64(v.Long()))
}
func (longImpl) mulRatio(v Value, n *big.Rat) Value {
	return mulRats(n, longRat(v.Long()))
}
func (longImpl) mulBigInt(v Value, n *big.Int) Value {
	return mulInts(n, big.NewInt(v.Long()))
}
func (longImpl) divLong(v Value, n int64) Value {
	return divRats(longRat(n), longRat(v.Long()))
}
func (longImpl) divDouble(v Value, n float64) Value {
	return NewDouble(n / float64(v.Long()))
}
func (longImpl) divRatio(v Value, n *big.Rat) Value {
	return divRats(n, longRat(v.Long()))
}
func (longImpl) divBigInt(v Value, n *big.Int) Value {
	return divRats(intRat(n), longRat(v.Long()))
}
func (longImpl) quotLong(v Value, n int64) Value {
	return quotRats(longRat(n), longRat(v.Long()))
}
func (longImpl) quotDouble(v Value, n float64) Value {
	return quotFloat(n, float64(v.Long()))
}
func (longImpl) quotRatio(v Value, n *big.Rat) Value {
	return quotRats(n, longRat(v.Long()))
}
func (longImpl) quotBigInt(v Value, n *big.Int) Value {
	return quotRats(intRat(n), longRat(v.Long()))
}
func (longImpl) cmpLong(v Value, n int64) int {
	return compareInt64(n, v.Long())
}
func (longImpl) cmpDouble(v Value, n float64) int {
	return compareFloat64(n, float64(v.Long()))
}
func (longImpl) cmpRatio(v Value, n *big.Rat) int {
	return n.Cmp(longRat(v.Long()))
}
func (longImpl) cmpBigInt(v Value, n *big.Int) int {
	return -compareLongAndBigInt(v.Long(), n)
}
