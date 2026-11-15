package runtime

import (
	"fmt"
	"reflect"
)

var (
	exMessageKey = NewKeyword("message")
	exDataKey    = NewKeyword("data")
	exCauseKey   = NewKeyword("cause")
)

// Throw panics v from native Go so compiled and Yaegi-interpreted try/catch
// both recover the original FLAG value.
func Throw(v Value) {
	panic(v)
}

// PanicValue converts a recovered panic into a FLAG value.
// Thrown FLAG values are returned as-is; other panics become strings.
func PanicValue(r any) Value {
	if r == nil {
		return NilValue
	}
	switch v := r.(type) {
	case Value:
		return v
	case *Value:
		if v != nil {
			return *v
		}
		return NilValue
	case reflect.Value:
		if v.IsValid() && v.CanInterface() {
			return PanicValue(v.Interface())
		}
	}
	return NewString(fmt.Sprint(r))
}

// IsExInfo reports whether v is an ExceptionInfo map from ex-info.
func IsExInfo(v Value) bool {
	if v.tag != TagMap {
		return false
	}
	return Contains(v, exMessageKey) && Contains(v, exDataKey)
}

// CatchMatches implements Clojure-style catch type tests.
// ExceptionInfo matches only ex-info maps; Exception, Throwable, and :default match any panic.
func CatchMatches(class string, thrown Value) bool {
	switch class {
	case "ExceptionInfo":
		return IsExInfo(thrown)
	case "Exception", "Throwable", ":default":
		return true
	default:
		return false
	}
}

// ExMessage returns the message of an exception, a thrown string, or nil.
func ExMessage(v Value) Value {
	return v.impl.exMessage(v)
}

// ExData returns the data map of an ExceptionInfo value, or nil.
func ExData(v Value) Value {
	if v.tag != TagMap {
		return NilValue
	}
	if data, ok := mapLookup(v, exDataKey); ok {
		return data
	}
	return NilValue
}

// ExCause returns the cause of an ExceptionInfo value, or nil.
func ExCause(v Value) Value {
	if v.tag != TagMap {
		return NilValue
	}
	if cause, ok := mapLookup(v, exCauseKey); ok {
		return cause
	}
	return NilValue
}

func mapLookup(v Value, key Value) (Value, bool) {
	if v.tag != TagMap {
		return NilValue, false
	}
	return v.mapPointer().items.Get(key)
}
