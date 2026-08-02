package runtime

import (
	"fmt"
	"unsafe"
)

type StringObject struct {
	Value string
}

func NewString(value string) Value {
	return Value{p: unsafe.Pointer(&StringObject{Value: value}), tag: TagString}
}

func (v Value) StringObject() *StringObject {
	if v.tag != TagString {
		panic("StringObject called on non-string Value")
	}
	if v.p == nil {
		panic("string Value does not contain string pointer")
	}
	return (*StringObject)(v.p)
}

func (v Value) StringValue() string {
	return v.StringObject().Value
}

func Format(format string, args ...Value) string {
	values := make([]any, 0, len(args))
	for _, arg := range args {
		values = append(values, ValueToAny(arg))
	}
	return fmt.Sprintf(format, values...)
}
