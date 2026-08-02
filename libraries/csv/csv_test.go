package csv

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.csv")
	if err := os.WriteFile(path, []byte("name,age\n\"Hanley, David\",52\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	records, err := ReadFile(path, DefaultOptions())
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 2 {
		t.Fatalf("unexpected record count: %d", len(records))
	}
	if records[1][0] != "Hanley, David" || records[1][1] != "52" {
		t.Fatalf("unexpected quoted row: %#v", records[1])
	}
}

func TestReadLines(t *testing.T) {
	records, err := ReadLines([]string{
		"race,points",
		"\"2022 T2T TAMPA\",150",
		"\"Line, With, Commas\",200",
	}, DefaultOptions())
	if err != nil {
		t.Fatal(err)
	}

	if len(records) != 3 {
		t.Fatalf("unexpected record count: %d", len(records))
	}
	if records[2][0] != "Line, With, Commas" || records[2][1] != "200" {
		t.Fatalf("unexpected parsed row: %#v", records[2])
	}
}

func TestReadAllToleratesBareQuotes(t *testing.T) {
	// Real-world race files contain bare double-quotes in unquoted fields.
	records, err := ReadLines([]string{
		"place,name,age,gender",
		",Cody O\"connell,13,M",
		"Big \"D\" 2017",
	}, DefaultOptions())
	if err != nil {
		t.Fatal(err)
	}

	if len(records) != 3 {
		t.Fatalf("unexpected record count: %d", len(records))
	}
	if records[1][1] != "Cody O\"connell" {
		t.Fatalf("unexpected bare-quote field: %#v", records[1])
	}
	if records[2][0] != "Big \"D\" 2017" {
		t.Fatalf("unexpected bare-quote title: %#v", records[2])
	}
}

func TestReadAllFromReader(t *testing.T) {
	records, err := ReadAll(strings.NewReader("a,b\n\"c,d\",e\n"), DefaultOptions())
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 2 {
		t.Fatalf("unexpected record count: %d", len(records))
	}
	if records[1][0] != "c,d" || records[1][1] != "e" {
		t.Fatalf("unexpected parsed row: %#v", records[1])
	}
}

func TestFieldsPerRecordOption(t *testing.T) {
	opt := DefaultOptions()
	opt.FieldsPerRecord = 2
	// Third field should cause an error with fixed width 2.
	_, err := ReadString("a,b,c\n", opt)
	if err == nil {
		t.Fatal("expected error for fields-per-record mismatch")
	}

	opt.FieldsPerRecord = -1
	rows, err := ReadString("a,b,c\n", opt)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || len(rows[0]) != 3 {
		t.Fatalf("unexpected rows: %#v", rows)
	}
}
