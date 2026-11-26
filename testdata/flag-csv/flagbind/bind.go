// Package flagbind is the FLAG FFI for github.com/davidhanley/CSV.
// FLAG copies these sources into generated builds as flagbuild/flagbind.
package flagbind

import (
	"fmt"
	"io"

	csvlib "github.com/davidhanley/CSV"
	flagrt "flag-lang/runtime"
)

var GoBind_csv_ReadCSV = flagrt.NewFunction(adaptCSVReadCSV)
var GoBind_csv_ReadCSVPath = flagrt.NewFunction(adaptCSVReadCSVPath)
var GoBind_csv_ReadCSVReader = flagrt.NewFunction(adaptCSVReadCSVReader)
var GoBind_csv_ReadCSVLines = flagrt.NewFunction(adaptCSVReadCSVLines)

func adaptCSVReadCSV(args ...flagrt.Value) flagrt.Value {
	source, opt := csvSourceAndOptions("csv/read-csv", args)
	rows, err := csvReadDispatch(source, opt)
	if err != nil {
		panic(err.Error())
	}
	return flagrt.GoRetStringMatrix(rows)
}

func adaptCSVReadCSVPath(args ...flagrt.Value) flagrt.Value {
	path, opt := csvStringAndOptions("csv/read-csv-path", args)
	rows, err := csvlib.ReadFile(path, opt)
	if err != nil {
		panic(err.Error())
	}
	return flagrt.GoRetStringMatrix(rows)
}

func adaptCSVReadCSVReader(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 && len(args) != 2 {
		panic(fmt.Sprintf("csv/read-csv-reader expects 1 or 2 arguments, got %d", len(args)))
	}
	r := flagrt.GoArgReader("csv/read-csv-reader", 0, args[0])
	opt := csvlib.DefaultOptions()
	if len(args) == 2 {
		opt = csvOptionsFromValue("csv/read-csv-reader", args[1])
	}
	rows, err := csvlib.ReadAll(r, opt)
	if err != nil {
		panic(err.Error())
	}
	return flagrt.GoRetStringMatrix(rows)
}

func adaptCSVReadCSVLines(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 && len(args) != 2 {
		panic(fmt.Sprintf("csv/read-csv-lines expects 1 or 2 arguments, got %d", len(args)))
	}
	lines := flagrt.SeqStrings("csv/read-csv-lines", args[0])
	opt := csvlib.DefaultOptions()
	if len(args) == 2 {
		opt = csvOptionsFromValue("csv/read-csv-lines", args[1])
	}
	rows, err := csvlib.ReadLines(lines, opt)
	if err != nil {
		panic(err.Error())
	}
	return flagrt.GoRetStringMatrix(rows)
}

func csvStringAndOptions(name string, args []flagrt.Value) (string, csvlib.Options) {
	if len(args) != 1 && len(args) != 2 {
		panic(fmt.Sprintf("%s expects 1 or 2 arguments, got %d", name, len(args)))
	}
	path := flagrt.GoArgString(name, 0, args[0])
	opt := csvlib.DefaultOptions()
	if len(args) == 2 {
		opt = csvOptionsFromValue(name, args[1])
	}
	return path, opt
}

func csvSourceAndOptions(name string, args []flagrt.Value) (flagrt.Value, csvlib.Options) {
	if len(args) != 1 && len(args) != 2 {
		panic(fmt.Sprintf("%s expects 1 or 2 arguments, got %d", name, len(args)))
	}
	opt := csvlib.DefaultOptions()
	if len(args) == 2 {
		opt = csvOptionsFromValue(name, args[1])
	}
	return args[0], opt
}

func csvReadDispatch(source flagrt.Value, opt csvlib.Options) ([][]string, error) {
	if _, ok := flagrt.AsIdent(source); ok {
		return csvlib.ReadFile(flagrt.GoArgString("csv/read-csv", 0, source), opt)
	}
	if flagrt.IsFile(source) {
		return csvlib.ReadAll(flagrt.GoArgReader("csv/read-csv", 0, source), opt)
	}
	if r, ok := flagrt.ValueToAny(source).(io.Reader); ok {
		return csvlib.ReadAll(r, opt)
	}
	if flagrt.IsSeq(source) {
		return csvlib.ReadLines(flagrt.SeqStrings("csv/read-csv", source), opt)
	}
	return nil, fmt.Errorf("csv/read-csv expects a path, reader, or line sequence; got %s", source.String())
}

func csvOptionsFromValue(name string, v flagrt.Value) csvlib.Options {
	opt := csvlib.DefaultOptions()
	if flagrt.IsNil(v) {
		return opt
	}
	if !flagrt.IsMapType(v) {
		panic(fmt.Sprintf("%s options: expected map, got %s", name, v.String()))
	}
	for _, e := range v.MapEntries() {
		key := flagrt.OptionKeyName(e.Key)
		switch key {
		case "fields-per-record", "fieldsPerRecord", "FieldsPerRecord":
			opt.FieldsPerRecord = int(flagrt.GoArgInt64(name, 0, e.Value))
		case "lazy-quotes", "lazyQuotes", "LazyQuotes":
			opt.LazyQuotes = flagrt.GoArgBool(name, 0, e.Value)
		case "trim-leading-space", "trimLeadingSpace", "TrimLeadingSpace":
			opt.TrimLeadingSpace = flagrt.GoArgBool(name, 0, e.Value)
		case "comma", "Comma":
			opt.Comma = flagrt.GoArgRune(name, 0, e.Value)
		case "comment", "Comment":
			opt.Comment = flagrt.GoArgRune(name, 0, e.Value)
		case "keep-blank-lines", "keepBlankLines", "KeepBlankLines":
			opt.KeepBlankLines = flagrt.GoArgBool(name, 0, e.Value)
		default:
			panic(fmt.Sprintf("%s options: unknown key %q", name, key))
		}
	}
	return opt
}
