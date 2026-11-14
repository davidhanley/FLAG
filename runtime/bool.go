package runtime

import "strconv"

type boolImpl struct{ defaultImpl }

func (boolImpl) hash(v Value) string {
	if v.Bool() {
		return "B:true"
	}
	return "B:false"
}
func (boolImpl) typ(Value) Value { return typeBool }
func (boolImpl) String(v Value) string {
	return strconv.FormatBool(v.Bool())
}
