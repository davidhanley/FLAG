package runtime

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileToStringsReadsLinesLazily(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "lines.txt")
	if err := os.WriteFile(path, []byte("alpha\nbeta\n"), 0o600); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	lines := FileToStringsPath(path)

	if got := ValueToString(First(lines)); got != "alpha" {
		t.Fatalf("expected first line alpha, got %q", got)
	}
	lines = Rest(lines)
	if got := ValueToString(First(lines)); got != "beta" {
		t.Fatalf("expected second line beta, got %q", got)
	}
	lines = Rest(lines)
	if got := First(lines); got.tag != TagNil {
		t.Fatalf("expected nil after exhausting lines, got %#v", got)
	}
}

func TestFileToStringsDelaysOpeningUntilConsumption(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "delayed.txt")
	if err := os.WriteFile(path, []byte("alpha\n"), 0o600); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	lines := FileToStringsPath(path)
	if err := os.Remove(path); err != nil {
		t.Fatalf("remove temp file: %v", err)
	}

	assertPanics(t, func() {
		_ = First(lines)
	})
}

func TestOpenFileAndFileToStringsIntegration(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "joined.txt")
	if err := os.WriteFile(path, []byte("x\ny\n"), 0o600); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	file := OpenFile(path)
	defer func() {
		if err := file.Close(); err != nil {
			t.Fatalf("close file: %v", err)
		}
	}()
	lines := FileToStrings(file)

	if got := ValueToString(First(lines)); got != "x" {
		t.Fatalf("expected first line x, got %q", got)
	}
}

func TestOpenFileCloseIsIdempotent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "close.txt")
	if err := os.WriteFile(path, []byte("x\n"), 0o600); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	file := OpenFile(path)
	if err := file.Close(); err != nil {
		t.Fatalf("first close failed: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("second close failed: %v", err)
	}
}

func TestLineSeqReadsFileLines(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "lines.txt")
	if err := os.WriteFile(path, []byte("a\nb\n"), 0o600); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	lines := LineSeq(OpenFile(path))
	if got := ValueToString(First(lines)); got != "a" {
		t.Fatalf("expected first line a, got %q", got)
	}
}

func TestReadLineReadsSequentialLines(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "readline.txt")
	if err := os.WriteFile(path, []byte("alpha\nbeta\n"), 0o600); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	file := OpenFile(path)
	defer func() {
		if err := file.Close(); err != nil {
			t.Fatalf("close file: %v", err)
		}
	}()

	if got := ReadLine(file); got.tag != TagString || got.StringValue() != "alpha" {
		t.Fatalf("expected first line alpha, got %#v", got)
	}
	if got := ReadLine(file); got.tag != TagString || got.StringValue() != "beta" {
		t.Fatalf("expected second line beta, got %#v", got)
	}
	if got := ReadLine(file); got.tag != TagNil {
		t.Fatalf("expected nil at eof, got %#v", got)
	}
}
