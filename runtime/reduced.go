package runtime

import (
	"unsafe"
)

type reducedImpl struct {
	defaultImpl
	val Value
}

// Reduced boxes x so reduce can stop early and unwrap it.
func Reduced(x Value) Value {
	return Value{impl: &reducedImpl{val: x}}
}

// IsReduced is true for a value produced by Reduced.
func IsReduced(v Value) bool {
	return isType[*reducedImpl](v)
}

// Unreduced unwraps a reduced box; other values are returned unchanged.
func Unreduced(v Value) Value {
	if r, ok := v.impl.(*reducedImpl); ok {
		return r.val
	}
	return v
}

// EnsureReduced returns v if it is already reduced, otherwise boxes it.
func EnsureReduced(v Value) Value {
	if IsReduced(v) {
		return v
	}
	return Reduced(v)
}

func (reducedImpl) hashImpl(v Value) uint64 {
	return hashPointer(hashSeedReduced, uintptr(unsafe.Pointer(v.impl.(*reducedImpl))))
}

func (reducedImpl) typ(Value) Value { return typeReduced }

func (reducedImpl) String(Value) string {
	return "#<reduced>"
}
