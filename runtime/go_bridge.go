package runtime

import (
	"fmt"
	"math"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"
	"unicode"
)

var (
	goFnRegistryMu sync.RWMutex
	goFnRegistry   = map[string]reflect.Value{}
	errorType      = reflect.TypeOf((*error)(nil)).Elem()
	valueType      = reflect.TypeOf(Value{})
)

func RegisterGoFunction(name string, fn any) {
	if strings.TrimSpace(name) == "" {
		panic("register-go-function expects non-empty name")
	}
	value := reflect.ValueOf(fn)
	if !value.IsValid() || value.Kind() != reflect.Func {
		panic("register-go-function expects a function")
	}
	goFnRegistryMu.Lock()
	goFnRegistry[name] = value
	goFnRegistryMu.Unlock()
}

// RegisteredGoFunctionValues returns a snapshot of the current Go-function
// registry. It exists for the build-time binding generator (internal/gobindgen),
// which reflects over the registrations to emit static, reflection-free
// adapters; it is not used on any runtime call path.
func RegisteredGoFunctionValues() map[string]reflect.Value {
	goFnRegistryMu.RLock()
	defer goFnRegistryMu.RUnlock()
	out := make(map[string]reflect.Value, len(goFnRegistry))
	for name, fn := range goFnRegistry {
		out[name] = fn
	}
	return out
}

func RegisterGoSymbols(symbols map[string]map[string]reflect.Value) {
	for packagePath, packageSymbols := range symbols {
		RegisterGoPackageFunctions(packagePath, packageSymbols)
	}
}

func RegisterGoPackageFunctions(packagePath string, symbols map[string]reflect.Value) {
	packagePath = strings.TrimSpace(packagePath)
	if packagePath == "" {
		return
	}
	packageName, importPath := deriveGoPackageNames(packagePath)
	for symbolName, symbolValue := range symbols {
		if !symbolValue.IsValid() || symbolValue.Kind() != reflect.Func {
			continue
		}
		if packageName != "" {
			RegisterGoFunction(packageName+"."+symbolName, symbolValue.Interface())
		}
		if importPath != "" && importPath != packageName {
			RegisterGoFunction(importPath+"."+symbolName, symbolValue.Interface())
		}
		RegisterGoFunction(packagePath+"."+symbolName, symbolValue.Interface())
	}
}

func deriveGoPackageNames(packagePath string) (string, string) {
	lastSlash := strings.LastIndex(packagePath, "/")
	if lastSlash < 0 {
		return packagePath, packagePath
	}
	pkg := packagePath[lastSlash+1:]
	importPath := packagePath
	if suffix := "/" + pkg; strings.HasSuffix(packagePath, suffix) {
		importPath = strings.TrimSuffix(packagePath, suffix)
	}
	return pkg, importPath
}

func GoFunction(name string) Value {
	fn := lookupGoFunction(name)
	return NewFunction(func(args ...Value) Value {
		return callGoFunction(name, fn, args...)
	})
}

func GoFunctionArgs(name string) Value {
	fn := lookupGoFunction(name)
	fnType := fn.Type()
	params := make([]Value, 0, fnType.NumIn())
	for i := 0; i < fnType.NumIn(); i++ {
		paramType := fnType.In(i)
		if fnType.IsVariadic() && i == fnType.NumIn()-1 {
			params = append(params, NewSymbol(paramType.Elem().String()+"..."))
			continue
		}
		params = append(params, NewSymbol(paramType.String()))
	}
	returns := make([]Value, 0, fnType.NumOut())
	for i := 0; i < fnType.NumOut(); i++ {
		returns = append(returns, NewSymbol(fnType.Out(i).String()))
	}

	return NewMap(
		NewKeyword("name"), NewSymbol(name),
		NewKeyword("variadic"), NewBool(fnType.IsVariadic()),
		NewKeyword("params"), NewArray(params...),
		NewKeyword("returns"), NewArray(returns...),
	)
}

func lookupGoFunction(name string) reflect.Value {
	name = strings.TrimSpace(name)
	if name == "" {
		panic("go function name cannot be empty")
	}
	goFnRegistryMu.RLock()
	fn, ok := goFnRegistry[name]
	goFnRegistryMu.RUnlock()
	if !ok {
		panic("unknown go function: " + name)
	}
	return fn
}

func callGoFunction(name string, fn reflect.Value, args ...Value) Value {
	fnType := fn.Type()
	minArgs := fnType.NumIn()
	if fnType.IsVariadic() {
		minArgs--
	}
	if len(args) < minArgs {
		panic(fmt.Sprintf("%s expects at least %d arguments, got %d", name, minArgs, len(args)))
	}
	if !fnType.IsVariadic() && len(args) != fnType.NumIn() {
		panic(fmt.Sprintf("%s expects exactly %d arguments, got %d", name, fnType.NumIn(), len(args)))
	}

	in := make([]reflect.Value, 0, fnType.NumIn())
	if fnType.IsVariadic() {
		fixedCount := fnType.NumIn() - 1
		for i := 0; i < fixedCount; i++ {
			converted, err := valueToGoArg(args[i], fnType.In(i))
			if err != nil {
				panic(fmt.Sprintf("%s argument %d: %v", name, i+1, err))
			}
			in = append(in, converted)
		}

		variadicType := fnType.In(fnType.NumIn() - 1).Elem()
		variadicLen := len(args) - fixedCount
		if variadicLen < 0 {
			variadicLen = 0
		}
		for i := 0; i < variadicLen; i++ {
			converted, err := valueToGoArg(args[fixedCount+i], variadicType)
			if err != nil {
				panic(fmt.Sprintf("%s argument %d: %v", name, fixedCount+i+1, err))
			}
			in = append(in, converted)
		}
		out := fn.Call(in)
		return normalizeGoCallResult(name, out)
	}

	for i, arg := range args {
		converted, err := valueToGoArg(arg, fnType.In(i))
		if err != nil {
			panic(fmt.Sprintf("%s argument %d: %v", name, i+1, err))
		}
		in = append(in, converted)
	}

	out := fn.Call(in)
	return normalizeGoCallResult(name, out)
}

func normalizeGoCallResult(name string, out []reflect.Value) Value {
	if len(out) == 0 {
		return NilValue()
	}
	if len(out) > 0 && out[len(out)-1].Type().Implements(errorType) {
		last := out[len(out)-1]
		if !last.IsNil() {
			panic(fmt.Sprintf("%s failed: %s", name, last.Interface().(error).Error()))
		}
		out = out[:len(out)-1]
		if len(out) == 0 {
			return NilValue()
		}
	}
	if len(out) == 1 {
		return anyToValue(out[0].Interface())
	}
	items := make([]Value, 0, len(out))
	for _, item := range out {
		items = append(items, anyToValue(item.Interface()))
	}
	return NewArray(items...)
}

func valueToGoArg(value Value, targetType reflect.Type) (reflect.Value, error) {
	if targetType == valueType {
		return reflect.ValueOf(value), nil
	}
	if isEmptyInterface(targetType) {
		native := nativeAny(value)
		if native == nil {
			return reflect.Zero(targetType), nil
		}
		return reflect.ValueOf(native), nil
	}
	if (value.tag == TagSymbol || value.tag == TagString) && targetType.Kind() == reflect.String {
		s := ""
		if value.tag == TagSymbol {
			s = value.SymbolObject().Name
		} else {
			s = value.StringValue()
		}
		rv := reflect.ValueOf(s)
		if rv.Type().AssignableTo(targetType) {
			return rv, nil
		}
		return rv.Convert(targetType), nil
	}

	raw := ValueToAny(value)
	if raw == nil {
		if canBeNilType(targetType) {
			return reflect.Zero(targetType), nil
		}
		return reflect.Value{}, fmt.Errorf("cannot assign nil to %s", targetType)
	}

	rv := reflect.ValueOf(raw)
	if rv.Type().AssignableTo(targetType) {
		return rv, nil
	}
	if rv.Type().ConvertibleTo(targetType) {
		return rv.Convert(targetType), nil
	}

	if value.tag == TagArray || value.tag == TagList {
		if targetType.Kind() == reflect.Slice {
			return convertSeqToSlice(value, targetType)
		}
	}
	if value.tag == TagMap {
		switch targetType.Kind() {
		case reflect.Map:
			return convertMapToMap(value, targetType)
		case reflect.Struct:
			return convertMapToStruct(value, targetType)
		case reflect.Pointer:
			if targetType.Elem().Kind() == reflect.Struct {
				structValue, err := convertMapToStruct(value, targetType.Elem())
				if err != nil {
					return reflect.Value{}, err
				}
				ptr := reflect.New(targetType.Elem())
				ptr.Elem().Set(structValue)
				return ptr, nil
			}
		}
	}

	return reflect.Value{}, fmt.Errorf("cannot convert %T to %s", raw, targetType)
}

func isEmptyInterface(t reflect.Type) bool {
	return t.Kind() == reflect.Interface && t.NumMethod() == 0
}

func nativeAny(value Value) any {
	switch value.tag {
	case TagNil:
		return nil
	case TagLong:
		return value.Long()
	case TagDouble:
		return value.Double()
	case TagRatio:
		return value.Ratio()
	case TagBigInt:
		return value.BigInt()
	case TagBool:
		return value.Bool()
	case TagString:
		return value.StringValue()
	case TagDate:
		return value.DateTime()
	case TagFile:
		return value.FileObject().File
	case TagSymbol:
		symbol := value.SymbolObject()
		if symbol.IsKeyword {
			return ":" + symbol.Name
		}
		return symbol.Name
	case TagList:
		return nativeSeq(value.ListValues())
	case TagArray:
		return nativeSeq(value.ArrayValues())
	case TagSet:
		return nativeSeq(value.SetValues())
	case TagLazyList:
		return nativeLazySeq(value)
	case TagMap:
		return nativeMap(value.MapEntries())
	case TagFunction:
		return value
	default:
		return value
	}
}

func nativeSeq(values []Value) []any {
	out := make([]any, 0, len(values))
	for _, item := range values {
		out = append(out, nativeAny(item))
	}
	return out
}

func nativeLazySeq(value Value) []any {
	out := make([]any, 0)
	current := value
	for {
		first := First(current)
		if first.tag == TagNil {
			return out
		}
		out = append(out, nativeAny(first))
		current = Rest(current)
		if current.tag == TagNil {
			return out
		}
	}
}

func nativeMap(entries []MapEntry) map[any]any {
	out := make(map[any]any, len(entries))
	for _, entry := range entries {
		out[nativeAny(entry.Key)] = nativeAny(entry.Value)
	}
	return out
}

func canBeNilType(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Interface, reflect.Pointer, reflect.Map, reflect.Slice, reflect.Func, reflect.Chan:
		return true
	default:
		return false
	}
}

func convertSeqToSlice(value Value, targetType reflect.Type) (reflect.Value, error) {
	var items []Value
	if value.tag == TagArray {
		items = value.ArrayValues()
	} else {
		items = value.ListValues()
	}
	out := reflect.MakeSlice(targetType, len(items), len(items))
	for i, item := range items {
		converted, err := valueToGoArg(item, targetType.Elem())
		if err != nil {
			return reflect.Value{}, fmt.Errorf("index %d: %w", i, err)
		}
		out.Index(i).Set(converted)
	}
	return out, nil
}

func convertMapToMap(value Value, targetType reflect.Type) (reflect.Value, error) {
	out := reflect.MakeMapWithSize(targetType, value.MapLen())
	for _, entry := range value.MapEntries() {
		key, err := valueToGoArg(entry.Key, targetType.Key())
		if err != nil {
			return reflect.Value{}, fmt.Errorf("map key %s: %w", ValueToString(entry.Key), err)
		}
		val, err := valueToGoArg(entry.Value, targetType.Elem())
		if err != nil {
			return reflect.Value{}, fmt.Errorf("map value for key %s: %w", ValueToString(entry.Key), err)
		}
		out.SetMapIndex(key, val)
	}
	return out, nil
}

func convertMapToStruct(value Value, targetType reflect.Type) (reflect.Value, error) {
	out := reflect.New(targetType).Elem()
	fieldByNormalizedName := make(map[string]int, targetType.NumField())
	for i := 0; i < targetType.NumField(); i++ {
		field := targetType.Field(i)
		if field.PkgPath != "" {
			continue
		}
		fieldByNormalizedName[normalizeInteropFieldName(field.Name)] = i
	}

	for _, entry := range value.MapEntries() {
		keyName, ok := mapKeyName(entry.Key)
		if !ok {
			continue
		}
		fieldIndex, ok := fieldByNormalizedName[normalizeInteropFieldName(keyName)]
		if !ok {
			continue
		}
		fieldType := targetType.Field(fieldIndex)
		fieldValue := out.Field(fieldIndex)
		if !fieldValue.CanSet() {
			continue
		}
		converted, err := valueToGoArg(entry.Value, fieldType.Type)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("field %s: %w", fieldType.Name, err)
		}
		fieldValue.Set(converted)
	}

	return out, nil
}

func mapKeyName(key Value) (string, bool) {
	switch key.tag {
	case TagSymbol:
		return key.SymbolObject().Name, true
	case TagString:
		return key.StringValue(), true
	default:
		return "", false
	}
}

func normalizeInteropFieldName(name string) string {
	var out strings.Builder
	for _, ch := range name {
		if unicode.IsLetter(ch) || unicode.IsDigit(ch) {
			out.WriteRune(unicode.ToLower(ch))
		}
	}
	return out.String()
}

func anyToValue(raw any) Value {
	switch v := raw.(type) {
	case nil:
		return NilValue()
	case Value:
		return v
	case bool:
		return NewBool(v)
	case int:
		return NewLong(int64(v))
	case int8:
		return NewLong(int64(v))
	case int16:
		return NewLong(int64(v))
	case int32:
		return NewLong(int64(v))
	case int64:
		return NewLong(v)
	case uint:
		return uintToValue(uint64(v))
	case uint8:
		return uintToValue(uint64(v))
	case uint16:
		return uintToValue(uint64(v))
	case uint32:
		return uintToValue(uint64(v))
	case uint64:
		return uintToValue(v)
	case float32:
		return NewDouble(float64(v))
	case float64:
		return NewDouble(v)
	case string:
		return NewString(v)
	case time.Time:
		return NewDate(v)
	case *os.File:
		return NewFile(v)
	case []any:
		items := make([]Value, 0, len(v))
		for _, item := range v {
			items = append(items, anyToValue(item))
		}
		return NewArray(items...)
	default:
		rv := reflect.ValueOf(raw)
		switch rv.Kind() {
		case reflect.Slice, reflect.Array:
			items := make([]Value, 0, rv.Len())
			for i := 0; i < rv.Len(); i++ {
				items = append(items, anyToValue(rv.Index(i).Interface()))
			}
			return NewArray(items...)
		case reflect.Map:
			entries := make([]Value, 0, rv.Len()*2)
			iter := rv.MapRange()
			for iter.Next() {
				entries = append(entries, anyToValue(iter.Key().Interface()))
				entries = append(entries, anyToValue(iter.Value().Interface()))
			}
			return NewMap(entries...)
		}
		return NewSymbol(fmt.Sprint(raw))
	}
}

func uintToValue(v uint64) Value {
	if v <= math.MaxInt64 {
		return NewLong(int64(v))
	}
	return NewSymbol(fmt.Sprintf("%d", v))
}
