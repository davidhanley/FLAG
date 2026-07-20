package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunCompileAcceptsOutputFlagAfterInput(t *testing.T) {
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "hello.flag")
	outputPath := filepath.Join(dir, "hello.go")

	source := `(ns hello.core)
(println "Hello from FLAG")
`
	if err := os.WriteFile(inputPath, []byte(source), 0o644); err != nil {
		t.Fatalf("WriteFile input: %v", err)
	}

	if err := run([]string{"compile", inputPath, "-o", outputPath}); err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	generated, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("ReadFile output: %v", err)
	}

	if !strings.Contains(string(generated), `fmt.Println("Hello from FLAG")`) {
		t.Fatalf("unexpected output:\n%s", generated)
	}
}
