package runtime

//go:generate go run flag-lang/internal/gobindgen

import (
	"fmt"
	"io"
	"time"
)

// This file provides the reflection-free conversion helpers used by the
// statically generated Go-function adapters in gobind_gen.go. Each helper
// mirrors the behavior of the reflection bridge (valueToGoArg /
// normalizeGoCallResult in go_bridge.go) for one concrete Go type, so that
// namespaced calls such as (str/trim s) or (csv/read-csv path) dispatch
// through a direct, statically bound call instead of a runtime name lookup
// plus reflection.
//
// The reflection bridge (GoFunction / lookupGoFunction) is retained for the
// explicit dynamic escape hatch (go-fn / go-fn-args); it is no longer on the
// path of ordinary namespaced calls.

// goArgArityExact panics unless the adapter received exactly n arguments.
func goArgArityExact(name string, args []Value, n int) {
	if len(args) != n {
		panic(fmt.Sprintf("%s expects exactly %d arguments, got %d", name, n, len(args)))
	}
}

// goArgArityAtLeast panics unless the adapter received at least n arguments
// (used for variadic Go functions, whose fixed parameters number n).
func goArgArityAtLeast(name string, args []Value, n int) {
	if len(args) < n {
		panic(fmt.Sprintf("%s expects at least %d arguments, got %d", name, n, len(args)))
	}
}

func goArgString(name string, idx int, v Value) string {
	switch v.tag {
	case TagSymbol:
		return v.SymbolObject().Name
	case TagString:
		return v.StringValue()
	}
	if raw, ok := ValueToAny(v).(string); ok {
		return raw
	}
	panic(fmt.Sprintf("%s argument %d: cannot convert %s to string", name, idx+1, ValueToString(v)))
}

func goArgInt64(name string, idx int, v Value) int64 {
	if v.tag != TagLong {
		panic(fmt.Sprintf("%s argument %d: cannot convert %s to int", name, idx+1, ValueToString(v)))
	}
	return v.Long()
}

func goArgTime(name string, idx int, v Value) time.Time {
	if v.tag != TagDate {
		panic(fmt.Sprintf("%s argument %d: cannot convert %s to time", name, idx+1, ValueToString(v)))
	}
	return v.DateTime()
}

func goArgReader(name string, idx int, v Value) io.Reader {
	if v.tag == TagFile {
		f := v.FileObject()
		if f == nil || f.File == nil || f.closed {
			panic(fmt.Sprintf("%s argument %d: expected open file", name, idx+1))
		}
		// Prefer the buffered view if line-seq already advanced the file.
		if f.lineReader != nil {
			return f.lineReader
		}
		return f.File
	}
	if raw, ok := ValueToAny(v).(io.Reader); ok {
		return raw
	}
	panic(fmt.Sprintf("%s argument %d: cannot convert %s to io.Reader", name, idx+1, ValueToString(v)))
}

func goArgMapAnyAny(name string, idx int, v Value) map[any]any {
	if v.tag == TagNil {
		return nil
	}
	if v.tag != TagMap {
		panic(fmt.Sprintf("%s argument %d: cannot convert %s to map", name, idx+1, ValueToString(v)))
	}
	return nativeMap(v.MapEntries())
}

// goArgAny converts a Value to the native Go representation used for empty
// interface parameters, matching nativeAny used by the reflection bridge.
func goArgAny(v Value) any {
	return nativeAny(v)
}

func goRetString(v string) Value { return NewString(v) }
func goRetBool(v bool) Value     { return NewBool(v) }
func goRetValue(v Value) Value   { return v }
func goRetAny(v any) Value       { return anyToValue(v) }

func goRetStrings(values []string) Value {
	items := make([]Value, 0, len(values))
	for _, s := range values {
		items = append(items, NewString(s))
	}
	return NewArray(items...)
}

func goRetStringMatrix(rows [][]string) Value {
	items := make([]Value, 0, len(rows))
	for _, row := range rows {
		items = append(items, goRetStrings(row))
	}
	return NewArray(items...)
}

func goRetStringErr(name, v string, err error) Value {
	if err != nil {
		panic(fmt.Sprintf("%s failed: %s", name, err.Error()))
	}
	return NewString(v)
}
