package runtime

import (
	"fmt"
	"reflect"
	"unsafe"
)

type recordBox struct {
	value any
}

func NewRecord(v any) Value {
	if v == nil {
		panic("NewRecord expects a non-nil Go struct")
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			panic("NewRecord expects a non-nil Go struct")
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		panic("NewRecord expects a Go struct")
	}
	box := &recordBox{value: v}
	return Value{p: unsafe.Pointer(box), tag: TagRecord}
}

func (v Value) recordBox() *recordBox {
	if v.tag != TagRecord {
		panic("record Value expected")
	}
	if v.p == nil {
		panic("record Value does not contain a struct")
	}
	return (*recordBox)(v.p)
}

func GoStruct[T any](v Value) T {
	rec, ok := v.recordBox().value.(T)
	if !ok {
		var zero T
		panic(fmt.Sprintf("record is %T, not %T", v.recordBox().value, zero))
	}
	return rec
}

func recordFieldValue(coll Value, key Value) (Value, bool) {
	name, ok := recordKeyName(key)
	if !ok {
		return NilValue(), false
	}
	rv := reflect.ValueOf(coll.recordBox().value)
	if rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return NilValue(), false
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return NilValue(), false
	}
	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		if recordFieldFlagName(field) != name {
			continue
		}
		return goValueToRecordField(rv.Field(i)), true
	}
	return NilValue(), false
}

func recordKeyName(key Value) (string, bool) {
	if key.tag != TagSymbol {
		return "", false
	}
	sym := key.SymbolObject()
	if !sym.IsKeyword {
		return "", false
	}
	return sym.Name, true
}

func recordFieldFlagName(field reflect.StructField) string {
	if tag := field.Tag.Get("flag"); tag != "" {
		return tag
	}
	return field.Name
}

func goValueToRecordField(field reflect.Value) Value {
	if !field.IsValid() {
		return NilValue()
	}
	switch field.Kind() {
	case reflect.String:
		return NewString(field.String())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return NewLong(field.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return NewLong(int64(field.Uint()))
	case reflect.Float32, reflect.Float64:
		return NewDouble(field.Float())
	case reflect.Bool:
		return NewBool(field.Bool())
	default:
		if field.Type() == valueType {
			return field.Interface().(Value)
		}
		if field.CanInterface() {
			return NewRecord(field.Interface())
		}
		return NilValue()
	}
}
