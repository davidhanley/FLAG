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
	return Value{impl: &recurImpl{payload: payload}}
}

// UnwrapRecur returns loop replacement values when v is a recur marker.
func UnwrapRecur(v Value) ([]Value, bool) {
	r, ok := v.impl.(*recurImpl)
	if !ok {
		return nil, false
	}
	if r.payload == nil {
		return []Value{}, true
	}
	return append([]Value(nil), r.payload.values...), true
}

type recurImpl struct {
	defaultImpl
	payload *recurPayload
}

func (recurImpl) hash(v Value) string {
	return "U:" + strconv.FormatUint(uint64(uintptr(unsafe.Pointer(asImpl[*recurImpl](v, "recur hash")))), 16)
}
func (recurImpl) typ(Value) Value { return typeRecur }
func (recurImpl) String(Value) string {
	return "#<recur>"
}
