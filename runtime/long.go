package runtime

import "strconv"

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
