package runtime

import (
	"fmt"
	"math/big"
	"strconv"
)

type ratioImpl struct{ defaultImpl }

func (ratioImpl) hash(v Value) string {
	if v.Ratio().IsInt() {
		num := v.Ratio().Num()
		if num.IsInt64() {
			return "L:" + strconv.FormatInt(num.Int64(), 10)
		}
		return "I:" + num.String()
	}
	return "R:" + v.Ratio().RatString()
}
func (ratioImpl) typ(Value) Value { return typeRatio }
func (ratioImpl) String(v Value) string {
	return v.Ratio().RatString()
}
func (ratioImpl) toAny(v Value) any    { return v.Ratio() }
func (ratioImpl) toNative(v Value) any { return v.Ratio() }
func (ratioImpl) toJSON(v Value) (any, error) {
	f, exact := v.Ratio().Float64()
	if !exact {
		return nil, fmt.Errorf("ratio %s cannot be represented as float64", v.Ratio().RatString())
	}
	return f, nil
}
func (ratioImpl) isNumeric(Value) bool { return true }
func (ratioImpl) add(a, b Value) Value {
	return b.impl.addRatio(b, a.Ratio())
}
func (ratioImpl) sub(a, b Value) Value {
	return b.impl.subRatio(b, a.Ratio())
}
func (ratioImpl) mul(a, b Value) Value {
	return b.impl.mulRatio(b, a.Ratio())
}
func (ratioImpl) div(a, b Value) Value {
	return b.impl.divRatio(b, a.Ratio())
}
func (ratioImpl) quot(a, b Value) Value {
	return b.impl.quotRatio(b, a.Ratio())
}
func (ratioImpl) cmpNum(a, b Value) int {
	return b.impl.cmpRatio(b, a.Ratio())
}
func (ratioImpl) abs(v Value) Value {
	rat := copyRat(v.Ratio())
	if rat.Sign() < 0 {
		rat.Neg(rat)
	}
	return NewRatioFromRat(rat)
}
func (ratioImpl) toDouble(v Value) Value {
	return NewDouble(ratToFloat64(v.Ratio()))
}
func (ratioImpl) numerator(v Value) Value {
	return newIntegerValueFromBigInt(new(big.Int).Set(v.Ratio().Num()))
}
func (ratioImpl) denominator(v Value) Value {
	return newIntegerValueFromBigInt(new(big.Int).Set(v.Ratio().Denom()))
}
func (ratioImpl) toFloat64(v Value) float64 {
	return ratToFloat64(v.Ratio())
}
func (ratioImpl) toRat(v Value) *big.Rat {
	return copyRat(v.Ratio())
}
func (ratioImpl) addLong(v Value, n int64) Value {
	return addRats(longRat(n), v.Ratio())
}
func (ratioImpl) addDouble(v Value, n float64) Value {
	return NewDouble(n + ratToFloat64(v.Ratio()))
}
func (ratioImpl) addRatio(v Value, n *big.Rat) Value {
	return addRats(n, v.Ratio())
}
func (ratioImpl) addBigInt(v Value, n *big.Int) Value {
	return addRats(intRat(n), v.Ratio())
}
func (ratioImpl) subLong(v Value, n int64) Value {
	return subRats(longRat(n), v.Ratio())
}
func (ratioImpl) subDouble(v Value, n float64) Value {
	return NewDouble(n - ratToFloat64(v.Ratio()))
}
func (ratioImpl) subRatio(v Value, n *big.Rat) Value {
	return subRats(n, v.Ratio())
}
func (ratioImpl) subBigInt(v Value, n *big.Int) Value {
	return subRats(intRat(n), v.Ratio())
}
func (ratioImpl) mulLong(v Value, n int64) Value {
	return mulRats(longRat(n), v.Ratio())
}
func (ratioImpl) mulDouble(v Value, n float64) Value {
	return NewDouble(n * ratToFloat64(v.Ratio()))
}
func (ratioImpl) mulRatio(v Value, n *big.Rat) Value {
	return mulRats(n, v.Ratio())
}
func (ratioImpl) mulBigInt(v Value, n *big.Int) Value {
	return mulRats(intRat(n), v.Ratio())
}
func (ratioImpl) divLong(v Value, n int64) Value {
	return divRats(longRat(n), v.Ratio())
}
func (ratioImpl) divDouble(v Value, n float64) Value {
	return NewDouble(n / ratToFloat64(v.Ratio()))
}
func (ratioImpl) divRatio(v Value, n *big.Rat) Value {
	return divRats(n, v.Ratio())
}
func (ratioImpl) divBigInt(v Value, n *big.Int) Value {
	return divRats(intRat(n), v.Ratio())
}
func (ratioImpl) quotLong(v Value, n int64) Value {
	return quotRats(longRat(n), v.Ratio())
}
func (ratioImpl) quotDouble(v Value, n float64) Value {
	return quotFloat(n, ratToFloat64(v.Ratio()))
}
func (ratioImpl) quotRatio(v Value, n *big.Rat) Value {
	return quotRats(n, v.Ratio())
}
func (ratioImpl) quotBigInt(v Value, n *big.Int) Value {
	return quotRats(intRat(n), v.Ratio())
}
func (ratioImpl) cmpLong(v Value, n int64) int {
	return longRat(n).Cmp(v.Ratio())
}
func (ratioImpl) cmpDouble(v Value, n float64) int {
	return compareFloat64(n, ratToFloat64(v.Ratio()))
}
func (ratioImpl) cmpRatio(v Value, n *big.Rat) int {
	return n.Cmp(v.Ratio())
}
func (ratioImpl) cmpBigInt(v Value, n *big.Int) int {
	return intRat(n).Cmp(v.Ratio())
}
