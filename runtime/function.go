package runtime

import "unsafe"

type FunctionObject struct {
	Fn func(args ...Value) Value
}

func NewFunction(fn func(args ...Value) Value) Value {
	if fn == nil {
		panic("NewFunction expects non-nil function")
	}
	return Value{p: unsafe.Pointer(&FunctionObject{Fn: fn}), tag: TagFunction}
}

func (v Value) FunctionObject() *FunctionObject {
	if v.tag != TagFunction {
		panic("FunctionObject called on non-function Value")
	}
	if v.p == nil {
		panic("function Value does not contain function pointer")
	}
	return (*FunctionObject)(v.p)
}

func Call(fn Value, args ...Value) Value {
	switch fn.tag {
	case TagFunction:
		return fn.FunctionObject().Fn(args...)
	case TagMap, TagDate, TagRecord:
		if len(args) != 1 && len(args) != 2 {
			panic("map invocation expects key and optional default")
		}
		if len(args) == 1 {
			return Get(fn, args[0])
		}
		return Get(fn, args[0], args[1])
	case TagSet:
		if len(args) != 1 && len(args) != 2 {
			panic("set invocation expects value and optional default")
		}
		if Contains(fn, args[0]) {
			return args[0]
		}
		if len(args) == 2 {
			return args[1]
		}
		return NilValue()
	case TagSymbol:
		symbol := fn.SymbolObject()
		if !symbol.IsKeyword {
			break
		}
		if len(args) != 1 && len(args) != 2 {
			panic("keyword invocation expects map and optional default")
		}
		key := NewKeyword(symbol.Name)
		if len(args) == 1 {
			return Get(args[0], key)
		}
		return Get(args[0], key, args[1])
	default:
		// fall through
	}
	panic("call expects function, map, set, or keyword Value")
}
