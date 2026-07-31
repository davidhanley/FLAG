package runtime

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCSVReadCSVThroughBridgeFromPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.csv")
	if err := os.WriteFile(path, []byte("name,age\n\"Hanley, David\",52\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	records := Call(GoFunction("csv/read-csv"), NewSymbol(path))
	if records.tag != TagArray || records.ArrayLen() != 2 {
		t.Fatalf("unexpected csv record shape: %#v", records)
	}
}

func TestCSVReadCSVThroughBridgeFromOpenFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.csv")
	if err := os.WriteFile(path, []byte("name,age\n\"Hanley, David\",52\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	file := OpenFile(path)
	defer func() { _ = file.Close() }()

	records := Call(GoFunction("csv/read-csv"), file)
	if records.tag != TagArray || records.ArrayLen() != 2 {
		t.Fatalf("unexpected csv record shape from open file: %#v", records)
	}
}
