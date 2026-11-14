package runtime

import (
	"math"
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
