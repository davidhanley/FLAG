package runtime

import (
	"fmt"
	"strconv"
	"unicode/utf8"
	"unsafe"
)

type StringObject struct {
	Value string
}

func NewString(value string) Value {
	return makeValue(TagString, 0, unsafe.Pointer(&StringObject{Value: value}))
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

type stringImpl struct{ defaultImpl }

func (stringImpl) hash(v Value) string {
	return "Q:" + v.StringValue()
}
func (stringImpl) typ(Value) Value { return typeString }
func (stringImpl) String(v Value) string {
	return strconv.Quote(v.StringValue())
}
func (stringImpl) get(v Value, key Value, missing Value) Value {
	if i, ok := associativeIndex(key); ok {
		runes := []rune(v.StringValue())
		if i < len(runes) {
			return NewString(string(runes[i]))
		}
	}
	return missing
}
func (stringImpl) first(v Value) Value {
	value := v.StringValue()
	if value == "" {
		return NilValue
	}
	_, size := utf8.DecodeRuneInString(value)
	return NewString(value[:size])
}
func (stringImpl) rest(v Value) Value {
	value := v.StringValue()
	if value == "" {
		return NewArray()
	}
	_, size := utf8.DecodeRuneInString(value)
	return Vec(NewString(value[size:]))
}
func (stringImpl) next(v Value) Value {
	rest := Rest(v)
	return nextIfEmpty(rest, rest.ArrayLen() == 0)
}

func Subs(s Value, start Value, end ...Value) Value {
	if s.tag != TagString {
		panic("subs expects a string")
	}
	if len(end) > 1 {
		panic("subs expects 2 or 3 arguments")
	}
	runes := []rune(s.StringValue())
	n := len(runes)
	i := nonNegativeCount("subs", start)
	j := n
	if len(end) == 1 {
		j = nonNegativeCount("subs", end[0])
	}
	if i > n || j > n || i > j {
		panic("subs index out of range")
	}
	return NewString(string(runes[i:j]))
}
