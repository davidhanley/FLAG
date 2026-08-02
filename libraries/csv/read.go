// Package csv is a pure Go CSV reader. It has no dependency on flag-lang/runtime.
// FLAG code reaches it only through static adapters (see runtime/csv_bind.go)
// and libraries/csv.lib.
package csv

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"
)

// ReadAll reads all records from r using opt.
func ReadAll(r io.Reader, opt Options) ([][]string, error) {
	if r == nil {
		return nil, fmt.Errorf("csv: nil reader")
	}
	reader := newReader(r, opt)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("csv read failed: %w", err)
	}
	return records, nil
}

// ReadFile opens path and reads all records using opt.
func ReadFile(path string, opt Options) ([][]string, error) {
	if strings.TrimSpace(path) == "" {
		return nil, fmt.Errorf("csv: empty path")
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("csv open failed: %w", err)
	}
	defer func() { _ = f.Close() }()
	return ReadAll(f, opt)
}

// ReadString parses CSV text in s using opt.
func ReadString(s string, opt Options) ([][]string, error) {
	return ReadAll(strings.NewReader(s), opt)
}

// ReadLines joins lines with newlines and parses the result using opt.
// Each element is one physical line of input (as from line-seq).
func ReadLines(lines []string, opt Options) ([][]string, error) {
	if lines == nil {
		lines = []string{}
	}
	return ReadString(strings.Join(lines, "\n"), opt)
}

func newReader(r io.Reader, opt Options) *csv.Reader {
	reader := csv.NewReader(r)
	reader.FieldsPerRecord = opt.FieldsPerRecord
	reader.LazyQuotes = opt.LazyQuotes
	reader.TrimLeadingSpace = opt.TrimLeadingSpace
	if opt.Comma != 0 {
		reader.Comma = opt.Comma
	}
	if opt.Comment != 0 {
		reader.Comment = opt.Comment
	}
	return reader
}
