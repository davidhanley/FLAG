package csv

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadCSVFromPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.csv")
	if err := os.WriteFile(path, []byte("name,age\n\"Hanley, David\",52\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	records := ReadCSVPath(path)
	if len(records) != 2 {
		t.Fatalf("unexpected record count: %d", len(records))
	}
	if records[1][0] != "Hanley, David" || records[1][1] != "52" {
		t.Fatalf("unexpected quoted row: %#v", records[1])
	}
}

func TestReadCSVFromLineSequence(t *testing.T) {
	records := ReadCSV([]string{
		"race,points",
		"\"2022 T2T TAMPA\",150",
		"\"Line, With, Commas\",200",
	})

	if len(records) != 3 {
		t.Fatalf("unexpected record count: %d", len(records))
	}
	if records[2][0] != "Line, With, Commas" || records[2][1] != "200" {
		t.Fatalf("unexpected parsed row: %#v", records[2])
	}
}

func TestReadCSVFromReader(t *testing.T) {
	records := ReadCSVReader(strings.NewReader("a,b\n\"c,d\",e\n"))
	if len(records) != 2 {
		t.Fatalf("unexpected record count: %d", len(records))
	}
	if records[1][0] != "c,d" || records[1][1] != "e" {
		t.Fatalf("unexpected parsed row: %#v", records[1])
	}
}
