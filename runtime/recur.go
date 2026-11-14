package runtime

import (
	"strconv"
	"unsafe"
)

type recurPayload struct {
	values []Value
}

// NewRecur creates an internal loop/recur marker consumed by compiler-emitted
// loop scaffolding.
func NewRecur(values ...Value) Value {
	copied := append([]Value(nil), values...)
	payload := &recurPayload{values: copied}
	return makeValue(TagRecur, 0, unsafe.Pointer(payload))
}

// UnwrapRecur returns loop replacement values when v is a recur marker.
func UnwrapRecur(v Value) ([]Value, bool) {
	if v.tag != TagRecur {
		return nil, false
	}
	if v.p == nil {
		return []Value{}, true
	}
	payload := (*recurPayload)(v.p)
	return append([]Value(nil), payload.values...), true
}

type recurImpl struct{ defaultImpl }

func (recurImpl) hash(v Value) string {
	return "U:" + strconv.FormatUint(uint64(uintptr(v.p)), 16)
}
func (recurImpl) typ(Value) Value { return typeRecur }
func (recurImpl) String(Value) string {
	return "#<recur>"
}
