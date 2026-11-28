// Package csv is a pure Go CSV reader. It has no dependency on FLAG.
package csv

import (
	"bytes"
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
	src := r
	if opt.KeepBlankLines {
		data, err := io.ReadAll(r)
		if err != nil {
			return nil, fmt.Errorf("csv read failed: %w", err)
		}
		src = bytes.NewReader(keepBlankRecords(data))
	}
	reader := newReader(src, opt)
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

// keepBlankRecords rewrites empty lines that are not inside quoted fields as a
// quoted-empty field. encoding/csv skips fully blank lines, which drops FRS
// race headers of the form name / date / <blank> / points.
func keepBlankRecords(data []byte) []byte {
	var out bytes.Buffer
	out.Grow(len(data) + 16)
	inQuotes := false
	atFieldStart := true
	emptyRecord := true
	i := 0
	for i < len(data) {
		c := data[i]
		if inQuotes {
			out.WriteByte(c)
			if c == '"' {
				if i+1 < len(data) && data[i+1] == '"' {
					out.WriteByte('"')
					i += 2
					continue
				}
				inQuotes = false
			}
			i++
			continue
		}
		if c == '\r' {
			if i+1 < len(data) && data[i+1] == '\n' {
				i++
			}
			c = '\n'
		}
		switch c {
		case '\n':
			if emptyRecord {
				out.WriteString(`""`)
			}
			out.WriteByte('\n')
			atFieldStart = true
			emptyRecord = true
			i++
		case ',':
			out.WriteByte(',')
			atFieldStart = true
			emptyRecord = false
			i++
		case '"':
			if atFieldStart {
				inQuotes = true
			}
			out.WriteByte('"')
			atFieldStart = false
			emptyRecord = false
			i++
		default:
			out.WriteByte(c)
			atFieldStart = false
			emptyRecord = false
			i++
		}
	}
	return out.Bytes()
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
