package runtime

import (
	"fmt"
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
