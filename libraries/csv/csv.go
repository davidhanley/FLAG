package csv

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"
)

func ReadCSV(source any) [][]string {
	r, closer := readerFromSource(source)
	if closer != nil {
		defer closer()
	}

        reader :=  csv.NewReader(r)
        reader.FieldsPerRecord = -1
        reader.LazyQuotes = true

	records, err := reader.ReadAll()
	if err != nil {
		panic(fmt.Sprintf("csv read failed: %v", err))
	}
	return records
}

func ReadCSVPath(path string) [][]string {
	if strings.TrimSpace(path) == "" {
		panic("read-csv expects a non-empty path")
	}
	f, err := os.Open(path)
	if err != nil {
		panic(fmt.Sprintf("csv open failed: %v", err))
	}
	defer func() {
		if cerr := f.Close(); cerr != nil {
			panic(fmt.Sprintf("csv close failed: %v", cerr))
		}
	}()

        reader := csv.NewReader(f)
        reader.FieldsPerRecord = -1
        reader.LazyQuotes = true

	records, err := reader.ReadAll()
	if err != nil {
		panic(fmt.Sprintf("csv read failed: %v", err))
	}
	return records
}

func ReadCSVReader(r io.Reader) [][]string {
        reader :=  csv.NewReader(r)
        reader.FieldsPerRecord = -1
        reader.LazyQuotes = true

	records, err := reader.ReadAll()
	if err != nil {
		panic(fmt.Sprintf("csv read failed: %v", err))
	}
	return records
}

func ReadCSVLines(lines any) [][]string {
	return ReadCSV(lines)
}

func readerFromSource(source any) (io.Reader, func()) {
	switch value := source.(type) {
	case nil:
		panic("read-csv expects a filename, reader, or line sequence")
	case string:
		return readerFromPath(value)
	case []string:
		return bytes.NewBufferString(strings.Join(value, "\n")), nil
	case []any:
		return bytes.NewBufferString(joinLineValues(value)), nil
	case io.Reader:
		return value, nil
	default:
		if path, ok := source.(fmt.Stringer); ok {
			return readerFromPath(path.String())
		}
	}

	panic(fmt.Sprintf("read-csv expects a filename, reader, or line sequence; got %T", source))
}

func readerFromPath(path string) (io.Reader, func()) {
	if strings.TrimSpace(path) == "" {
		panic("read-csv expects a non-empty path")
	}
	f, err := os.Open(path)
	if err != nil {
		panic(fmt.Sprintf("csv open failed: %v", err))
	}
	return f, func() {
		if cerr := f.Close(); cerr != nil {
			panic(fmt.Sprintf("csv close failed: %v", cerr))
		}
	}
}

func joinLineValues(lines []any) string {
	parts := make([]string, 0, len(lines))
	for _, line := range lines {
		parts = append(parts, fmt.Sprint(line))
	}
	return strings.Join(parts, "\n")
}
