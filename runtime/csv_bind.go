package runtime

import (
	"fmt"
	"io"

	csvlib "flag-lang/libraries/csv"
)

// Hand-written CSV adapters: unbox FLAG Values (and option maps), call the
// pure libraries/csv package, box results. See docs/go-libraries.md.

// GoBind_csv_ReadCSV dispatches (csv/read-csv source) or (csv/read-csv source opts).
// source may be a path string/symbol, an io.Reader, or a sequence of line strings.
var GoBind_csv_ReadCSV = NewFunction(adaptCSVReadCSV)

// GoBind_csv_ReadCSVPath dispatches (csv/read-csv-path path) or (... path opts).
var GoBind_csv_ReadCSVPath = NewFunction(adaptCSVReadCSVPath)

// GoBind_csv_ReadCSVReader dispatches (csv/read-csv-reader rdr) or (... rdr opts).
var GoBind_csv_ReadCSVReader = NewFunction(adaptCSVReadCSVReader)

// GoBind_csv_ReadCSVLines dispatches (csv/read-csv-lines lines) or (... lines opts).
var GoBind_csv_ReadCSVLines = NewFunction(adaptCSVReadCSVLines)

func adaptCSVReadCSV(args ...Value) Value {
	source, opt := csvSourceAndOptions("csv/read-csv", args)
	rows, err := csvReadDispatch(source, opt)
	if err != nil {
		panic(err.Error())
	}
	return goRetStringMatrix(rows)
}

func adaptCSVReadCSVPath(args ...Value) Value {
	path, opt := csvStringAndOptions("csv/read-csv-path", args)
	rows, err := csvlib.ReadFile(path, opt)
	if err != nil {
		panic(err.Error())
	}
	return goRetStringMatrix(rows)
}

func adaptCSVReadCSVReader(args ...Value) Value {
	if len(args) != 1 && len(args) != 2 {
		panic(fmt.Sprintf("csv/read-csv-reader expects 1 or 2 arguments, got %d", len(args)))
	}
	r := goArgReader("csv/read-csv-reader", 0, args[0])
	opt := csvlib.DefaultOptions()
	if len(args) == 2 {
		opt = csvOptionsFromValue("csv/read-csv-reader", args[1])
	}
	rows, err := csvlib.ReadAll(r, opt)
	if err != nil {
		panic(err.Error())
	}
	return goRetStringMatrix(rows)
}

func adaptCSVReadCSVLines(args ...Value) Value {
	if len(args) != 1 && len(args) != 2 {
		panic(fmt.Sprintf("csv/read-csv-lines expects 1 or 2 arguments, got %d", len(args)))
	}
	lines := csvLinesFromValue("csv/read-csv-lines", args[0])
	opt := csvlib.DefaultOptions()
	if len(args) == 2 {
		opt = csvOptionsFromValue("csv/read-csv-lines", args[1])
	}
	rows, err := csvlib.ReadLines(lines, opt)
	if err != nil {
		panic(err.Error())
	}
	return goRetStringMatrix(rows)
}

func csvStringAndOptions(name string, args []Value) (string, csvlib.Options) {
	if len(args) != 1 && len(args) != 2 {
		panic(fmt.Sprintf("%s expects 1 or 2 arguments, got %d", name, len(args)))
	}
	path := goArgString(name, 0, args[0])
	opt := csvlib.DefaultOptions()
	if len(args) == 2 {
		opt = csvOptionsFromValue(name, args[1])
	}
	return path, opt
}

func csvSourceAndOptions(name string, args []Value) (Value, csvlib.Options) {
	if len(args) != 1 && len(args) != 2 {
		panic(fmt.Sprintf("%s expects 1 or 2 arguments, got %d", name, len(args)))
	}
	opt := csvlib.DefaultOptions()
	if len(args) == 2 {
		opt = csvOptionsFromValue(name, args[1])
	}
	return args[0], opt
}

func csvReadDispatch(source Value, opt csvlib.Options) ([][]string, error) {
	// Path-like.
	if _, ok := source.impl.asIdent(source); ok {
		return csvlib.ReadFile(goArgString("csv/read-csv", 0, source), opt)
	}
	if TypeOf(source) == typeFile {
		return csvlib.ReadAll(goArgReader("csv/read-csv", 0, source), opt)
	}

	// Host reader (os.File wrapped as any, etc.).
	if r, ok := ValueToAny(source).(io.Reader); ok {
		return csvlib.ReadAll(r, opt)
	}

	// Sequence of lines.
	if isSeqValue(source) {
		lines := csvLinesFromValue("csv/read-csv", source)
		return csvlib.ReadLines(lines, opt)
	}

	return nil, fmt.Errorf("csv/read-csv expects a path, reader, or line sequence; got %s", source.String())
}

func isSeqValue(v Value) bool {
	return v.impl.seqQ(v)
}

func csvLinesFromValue(name string, v Value) []string {
	if IsNil(v) {
		return []string{}
	}
	cursor := newSeqCursor(v)
	out := make([]string, 0, 16)
	for {
		item, ok := cursor.nextOrDone()
		if !ok {
			return out
		}
		out = append(out, goArgString(name, 0, item))
	}
}

// csvOptionsFromValue maps a FLAG option map to csv.Options.
// Supported keys (keywords or strings):
//
//	:fields-per-record  int   (csv.Reader.FieldsPerRecord)
//	:lazy-quotes        bool
//	:trim-leading-space bool
//	:comma              char/string length 1
//	:comment            char/string length 1
func csvOptionsFromValue(name string, v Value) csvlib.Options {
	opt := csvlib.DefaultOptions()
	if IsNil(v) {
		return opt
	}
	if TypeOf(v) != typeMap {
		panic(fmt.Sprintf("%s options: expected map, got %s", name, v.String()))
	}
	for _, e := range v.MapEntries() {
		key := optionKeyName(e.Key)
		switch key {
		case "fields-per-record", "fieldsPerRecord", "FieldsPerRecord":
			opt.FieldsPerRecord = int(goArgInt64(name, 0, e.Value))
		case "lazy-quotes", "lazyQuotes", "LazyQuotes":
			opt.LazyQuotes = goArgBool(name, 0, e.Value)
		case "trim-leading-space", "trimLeadingSpace", "TrimLeadingSpace":
			opt.TrimLeadingSpace = goArgBool(name, 0, e.Value)
		case "comma", "Comma":
			opt.Comma = goArgRune(name, 0, e.Value)
		case "comment", "Comment":
			opt.Comment = goArgRune(name, 0, e.Value)
		default:
			panic(fmt.Sprintf("%s options: unknown key %q", name, key))
		}
	}
	return opt
}

func optionKeyName(k Value) string {
	if s, ok := k.impl.asIdent(k); ok {
		return s
	}
	if s, ok := ValueToAny(k).(string); ok {
		return s
	}
	panic(fmt.Sprintf("option key must be keyword/string, got %s", k.String()))
}

func goArgBool(name string, idx int, v Value) bool {
	if !isType[boolImpl](v) {
		panic(fmt.Sprintf("%s argument %d: cannot convert %s to bool", name, idx+1, v.String()))
	}
	return v.Bool()
}

func goArgRune(name string, idx int, v Value) rune {
	if isType[*stringImpl](v) {
		s := v.StringValue()
		if len([]rune(s)) != 1 {
			panic(fmt.Sprintf("%s argument %d: expected single character string, got %q", name, idx+1, s))
		}
		return []rune(s)[0]
	}
	if isType[longImpl](v) {
		return rune(v.Long())
	}
	panic(fmt.Sprintf("%s argument %d: cannot convert %s to rune", name, idx+1, v.String()))
}
