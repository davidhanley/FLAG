package runtime

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/bytedance/sonic"
)

func ToJSON(v Value) string {
	jsonVal, err := valueToJSONValue(v)
	if err != nil {
		panic("to-json: " + err.Error())
	}
	jsonStr, err := sonic.MarshalString(jsonVal)
	if err != nil {
		panic("to-json: " + err.Error())
	}
	return jsonStr
}

func FromJSON(jsonStr string) Value {
	var obj any
	err := sonic.UnmarshalString(jsonStr, &obj)
	if err != nil {
		panic("from-json: " + err.Error())
	}
	return jsonToValue(obj)
}

type jsonReadOptions struct {
	defaultKeywordKeys bool
	keyFn              Value
	hasKeyFn           bool
	valueFn            Value
	hasValueFn         bool
}

func JSONReadStr(source Value, opts ...Value) Value {
	options := parseJSONReadOptions("json/read-str", opts)
	return readJSONValue("json/read-str", source, options)
}

func JSONRead(source Value, opts ...Value) Value {
	options := parseJSONReadOptions("json/read", opts)
	return readJSONValue("json/read", source, options)
}

func parseJSONReadOptions(name string, opts []Value) jsonReadOptions {
	if len(opts)%2 != 0 {
		panic(name + " expects keyword/value option pairs")
	}

	options := jsonReadOptions{}
	for i := 0; i < len(opts); i += 2 {
		optionName := jsonOptionName(name, opts[i])
		switch optionName {
		case "key-fn":
			options.keyFn = opts[i+1]
			options.hasKeyFn = true
		case "value-fn":
			options.valueFn = opts[i+1]
			options.hasValueFn = true
		default:
			panic(fmt.Sprintf("%s does not support option %q", name, optionName))
		}
	}
	return options
}

func jsonOptionName(name string, option Value) string {
	if s, ok := option.impl.asIdent(option); ok {
		return s
	}
	panic(name + " option keys must be strings, symbols, or keywords")
}

func readJSONValue(name string, source Value, options jsonReadOptions) Value {
	switch TypeOf(source) {
	case typeString:
		decoder := json.NewDecoder(strings.NewReader(source.StringValue()))
		decoder.UseNumber()
		obj, ok := decodeJSONObject(name, decoder, false)
		if !ok {
			return NilValue
		}
		return jsonToValueWithOptions(obj, options)
	case typeFile:
		file := source.FileObject()
		file.mu.Lock()
		defer file.mu.Unlock()
		reader := file.bufferedReaderLocked()
		decoder := json.NewDecoder(reader)
		decoder.UseNumber()
		obj, ok := decodeJSONObject(name, decoder, true)
		if !ok {
			return NilValue
		}
		restoreJSONBufferedInput(file, decoder, reader)
		return jsonToValueWithOptions(obj, options)
	default:
		panic(name + " expects a string or file")
	}
}

func decodeJSONObject(name string, decoder *json.Decoder, allowEOF bool) (any, bool) {
	var obj any
	if err := decoder.Decode(&obj); err != nil {
		if err == io.EOF && allowEOF {
			return nil, false
		}
		panic(name + " failed: " + err.Error())
	}
	return obj, true
}

func restoreJSONBufferedInput(file *FileObject, decoder *json.Decoder, reader *bufio.Reader) {
	buffered, err := io.ReadAll(decoder.Buffered())
	if err != nil {
		panic("json/read failed: " + err.Error())
	}
	if len(buffered) == 0 {
		file.lineReader = reader
		return
	}
	file.lineReader = bufio.NewReader(io.MultiReader(bytes.NewReader(buffered), reader))
}

func valueToJSONValue(v Value) (any, error) {
	return v.impl.toJSON(v)
}

func jsonToValue(obj any) Value {
	return jsonToValueWithOptions(obj, jsonReadOptions{defaultKeywordKeys: true})
}

func jsonToValueWithOptions(obj any, options jsonReadOptions) Value {
	switch val := obj.(type) {
	case nil:
		return NilValue
	case bool:
		return NewBool(val)
	case int64:
		return NewLong(val)
	case json.Number:
		if integer, err := val.Int64(); err == nil {
			return NewLong(integer)
		}
		floatVal, err := val.Float64()
		if err != nil {
			panic("unsupported JSON number: " + val.String())
		}
		return NewDouble(floatVal)
	case float64:
		if val == float64(int64(val)) {
			return NewLong(int64(val))
		}
		return NewDouble(val)
	case string:
		return NewString(val)
	case []any:
		items := make([]Value, 0, len(val))
		for _, item := range val {
			items = append(items, jsonToValueWithOptions(item, options))
		}
		return NewArray(items...)
	case map[string]any:
		entries := make([]Value, 0, len(val)*2)
		for key, value := range val {
			valueKey := jsonObjectKeyValue(key, options)
			valueItem := jsonToValueWithOptions(value, options)
			if options.hasValueFn {
				valueItem = Call(options.valueFn, valueKey, valueItem)
			}
			entries = append(entries, valueKey, valueItem)
		}
		return NewMap(entries...)
	default:
		panic(fmt.Sprintf("unsupported JSON type: %T", val))
	}
}

func jsonObjectKeyValue(key string, options jsonReadOptions) Value {
	keyValue := NewString(key)
	if options.hasKeyFn {
		return Call(options.keyFn, keyValue)
	}
	if options.defaultKeywordKeys {
		return NewKeyword(key)
	}
	return keyValue
}

func mapToJSONValue(v Value) (map[string]any, error) {
	entries := v.MapEntries()
	obj := make(map[string]any, len(entries))
	for _, entry := range entries {
		key := entry.Key.String()
		if s, ok := entry.Key.impl.asIdent(entry.Key); ok {
			key = s
		}
		val, err := valueToJSONValue(entry.Value)
		if err != nil {
			return nil, err
		}
		obj[key] = val
	}
	return obj, nil
}

func arrayToJSONValue(items []Value) ([]any, error) {
	arr := make([]any, len(items))
	for i, item := range items {
		jsonVal, err := valueToJSONValue(item)
		if err != nil {
			return nil, err
		}
		arr[i] = jsonVal
	}
	return arr, nil
}
