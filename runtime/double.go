package runtime

import (
	"math"
	"math/big"
	"strconv"
)

type doubleImpl struct{ defaultImpl }

func (doubleImpl) hash(v Value) string {
	if math.Trunc(v.Double()) == v.Double() {
		return "L:" + strconv.FormatInt(int64(v.Double()), 10)
	}
	return "D:" + strconv.FormatFloat(v.Double(), 'g', -1, 64)
}
func (doubleImpl) typ(Value) Value { return typeFloat }
func (doubleImpl) String(v Value) string {
	return strconv.FormatFloat(v.Double(), 'g', -1, 64)
}
func (doubleImpl) toAny(v Value) any           { return v.Double() }
func (doubleImpl) toNative(v Value) any        { return v.Double() }
func (doubleImpl) toJSON(v Value) (any, error) { return v.Double(), nil }
func (doubleImpl) isNumeric(Value) bool        { return true }
func (doubleImpl) add(a, b Value) Value        { return b.impl.addDouble(b, a.Double()) }
func (doubleImpl) sub(a, b Value) Value        { return b.impl.subDouble(b, a.Double()) }
func (doubleImpl) mul(a, b Value) Value        { return b.impl.mulDouble(b, a.Double()) }
func (doubleImpl) div(a, b Value) Value        { return b.impl.divDouble(b, a.Double()) }
func (doubleImpl) quot(a, b Value) Value       { return b.impl.quotDouble(b, a.Double()) }
func (doubleImpl) cmpNum(a, b Value) int       { return b.impl.cmpDouble(b, a.Double()) }
func (doubleImpl) abs(v Value) Value           { return NewDouble(math.Abs(v.Double())) }
func (doubleImpl) toDouble(v Value) Value      { return v }
func (doubleImpl) toFloat64(v Value) float64   { return v.Double() }
func (doubleImpl) addLong(v Value, n int64) Value {
	return NewDouble(float64(n) + v.Double())
}
func (doubleImpl) addDouble(v Value, n float64) Value {
	return NewDouble(n + v.Double())
}
func (doubleImpl) addRatio(v Value, n *big.Rat) Value {
	return NewDouble(ratToFloat64(n) + v.Double())
}
func (doubleImpl) addBigInt(v Value, n *big.Int) Value {
	return NewDouble(bigIntToFloat64(n) + v.Double())
}
func (doubleImpl) subLong(v Value, n int64) Value {
	return NewDouble(float64(n) - v.Double())
}
func (doubleImpl) subDouble(v Value, n float64) Value {
	return NewDouble(n - v.Double())
}
func (doubleImpl) subRatio(v Value, n *big.Rat) Value {
	return NewDouble(ratToFloat64(n) - v.Double())
}
func (doubleImpl) subBigInt(v Value, n *big.Int) Value {
	return NewDouble(bigIntToFloat64(n) - v.Double())
}
func (doubleImpl) mulLong(v Value, n int64) Value {
	return NewDouble(float64(n) * v.Double())
}
func (doubleImpl) mulDouble(v Value, n float64) Value {
	return NewDouble(n * v.Double())
}
func (doubleImpl) mulRatio(v Value, n *big.Rat) Value {
	return NewDouble(ratToFloat64(n) * v.Double())
}
func (doubleImpl) mulBigInt(v Value, n *big.Int) Value {
	return NewDouble(bigIntToFloat64(n) * v.Double())
}
func (doubleImpl) divLong(v Value, n int64) Value {
	return NewDouble(float64(n) / v.Double())
}
func (doubleImpl) divDouble(v Value, n float64) Value {
	return NewDouble(n / v.Double())
}
func (doubleImpl) divRatio(v Value, n *big.Rat) Value {
	return NewDouble(ratToFloat64(n) / v.Double())
}
func (doubleImpl) divBigInt(v Value, n *big.Int) Value {
	return NewDouble(bigIntToFloat64(n) / v.Double())
}
func (doubleImpl) quotLong(v Value, n int64) Value {
	return quotFloat(float64(n), v.Double())
}
func (doubleImpl) quotDouble(v Value, n float64) Value {
	return quotFloat(n, v.Double())
}
func (doubleImpl) quotRatio(v Value, n *big.Rat) Value {
	return quotFloat(ratToFloat64(n), v.Double())
}
func (doubleImpl) quotBigInt(v Value, n *big.Int) Value {
	return quotFloat(bigIntToFloat64(n), v.Double())
}
func (doubleImpl) cmpLong(v Value, n int64) int {
	return compareFloat64(float64(n), v.Double())
}
func (doubleImpl) cmpDouble(v Value, n float64) int {
	return compareFloat64(n, v.Double())
}
func (doubleImpl) cmpRatio(v Value, n *big.Rat) int {
	return compareFloat64(ratToFloat64(n), v.Double())
}
func (doubleImpl) cmpBigInt(v Value, n *big.Int) int {
	return compareFloat64(bigIntToFloat64(n), v.Double())
}
