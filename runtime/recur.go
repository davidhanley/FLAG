package runtime

import "unsafe"

type recurPayload struct {
	values []Value
}

// NewRecur creates an internal loop/recur marker consumed by compiler-emitted
// loop scaffolding.
func NewRecur(values ...Value) Value {
	copied := append([]Value(nil), values...)
	payload := &recurPayload{values: copied}
	return Value{p: unsafe.Pointer(payload), tag: TagRecur}
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
