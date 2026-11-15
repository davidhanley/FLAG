package runtime

import (
	"unsafe"
)

type FunctionObject struct {
	Fn func(args ...Value) Value
}

type functionImpl struct {
	defaultImpl
	obj *FunctionObject
}

func NewFunction(fn func(args ...Value) Value) Value {
	if fn == nil {
		panic("NewFunction expects non-nil function")
	}
	return Value{impl: &functionImpl{obj: &FunctionObject{Fn: fn}}}
}

func (v Value) FunctionObject() *FunctionObject {
	return asImpl[*functionImpl](v, "FunctionObject called on non-function Value").obj
}

// Fn invokes a function Value. Compiled FLAG uses this for `def`/`GoBind`
// values that are already functions (`translate_name.Fn(x)`), because a
// `flagrt.Value` is a struct and cannot be called as `translate_name(x)`.
func (v Value) Fn(args ...Value) Value {
	return v.FunctionObject().Fn(args...)
}

func Call(fn Value, args ...Value) Value {
	return fn.impl.apply(fn, args...)
}

func lookupApply(coll Value, args []Value, kind string) Value {
	if len(args) != 1 && len(args) != 2 {
		panic(kind + " invocation expects key and optional default")
	}
	if len(args) == 1 {
		return Get(coll, args[0])
	}
	return Get(coll, args[0], args[1])
}

func (functionImpl) hashImpl(v Value) uint64 {
	return hashPointer(hashSeedFn, uintptr(unsafe.Pointer(v.FunctionObject())))
}
func (functionImpl) typ(Value) Value { return typeFn }
func (functionImpl) String(Value) string {
	return "#<fn>"
}
func (functionImpl) toJSON(Value) (any, error) {
	return "#<fn>", nil
}
func (functionImpl) apply(v Value, args ...Value) Value {
	return v.FunctionObject().Fn(args...)
}
