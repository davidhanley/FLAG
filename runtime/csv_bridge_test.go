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

	records := Call(GoBind_csv_ReadCSV, NewString(path))
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

	records := Call(GoBind_csv_ReadCSV, file)
	if records.tag != TagArray || records.ArrayLen() != 2 {
		t.Fatalf("unexpected csv record shape from open file: %#v", records)
	}
}

func TestCSVReadCSVPathWithOptions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.csv")
	if err := os.WriteFile(path, []byte("a,b,c\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	// Fixed width 2 should reject a 3-field row.
	opts := NewMap(NewKeyword("fields-per-record"), NewLong(2))
	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("expected panic for fields-per-record mismatch")
			}
		}()
		_ = Call(GoBind_csv_ReadCSVPath, NewString(path), opts)
	}()

	// Variable width still works.
	okOpts := NewMap(NewKeyword("fields-per-record"), NewLong(-1), NewKeyword("lazy-quotes"), NewBool(true))
	records := Call(GoBind_csv_ReadCSVPath, NewString(path), okOpts)
	if records.tag != TagArray || records.ArrayLen() != 1 {
		t.Fatalf("unexpected records: %#v", records)
	}
}
