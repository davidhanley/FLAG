package main

import (
	"os"
	"os/exec"
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

	if !strings.Contains(string(generated), `fmt.Println(flagrt.Str("Hello from FLAG"))`) {
		t.Fatalf("unexpected output:\n%s", generated)
	}
}

func TestRunBuildCreatesRunnableBinary(t *testing.T) {
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "fib.flag")
	outputPath := filepath.Join(dir, "fibbin")

	source := `(defn fib [x] (if (< x 3) 1 (+ (fib (- x 1)) (fib (- x 2)))))
(println (fib 7))
`
	if err := os.WriteFile(inputPath, []byte(source), 0o644); err != nil {
		t.Fatalf("WriteFile input: %v", err)
	}

	if err := run([]string{"build", inputPath, "-o", outputPath}); err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	result, err := exec.Command(outputPath).CombinedOutput()
	if err != nil {
		t.Fatalf("built binary failed: %v\n%s", err, string(result))
	}
	if strings.TrimSpace(string(result)) != "13" {
		t.Fatalf("unexpected output from built binary: %q", result)
	}
}

func TestRunBuildDirectoryCreatesRunnableBinary(t *testing.T) {
	dir := t.TempDir()
	sourceDir := filepath.Join(dir, "src")
	if err := os.MkdirAll(sourceDir, 0o755); err != nil {
		t.Fatalf("MkdirAll sourceDir: %v", err)
	}

	file1 := filepath.Join(sourceDir, "a.flag")
	file2 := filepath.Join(sourceDir, "b.clj")
	file3 := filepath.Join(sourceDir, "ignore.txt")
	outputPath := filepath.Join(dir, "appbin")

	if err := os.WriteFile(file1, []byte("(defn sq [x] (* x x))\n"), 0o644); err != nil {
		t.Fatalf("WriteFile %s: %v", file1, err)
	}
	if err := os.WriteFile(file2, []byte("(println (sq 7))\n"), 0o644); err != nil {
		t.Fatalf("WriteFile %s: %v", file2, err)
	}
	if err := os.WriteFile(file3, []byte("not code"), 0o644); err != nil {
		t.Fatalf("WriteFile %s: %v", file3, err)
	}

	if err := run([]string{"build", sourceDir, "-o", outputPath}); err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	result, err := exec.Command(outputPath).CombinedOutput()
	if err != nil {
		t.Fatalf("built binary failed: %v\n%s", err, string(result))
	}
	if strings.TrimSpace(string(result)) != "49" {
		t.Fatalf("unexpected output from built binary: %q", result)
	}
}

func TestRunBuildExampleProjectDirectory(t *testing.T) {
	exampleDir, err := filepath.Abs(filepath.Join("..", "..", "examples", "hello"))
	if err != nil {
		t.Fatalf("Abs exampleDir: %v", err)
	}

	outputPath := filepath.Join(t.TempDir(), "hello-bin")
	if err := run([]string{"build", exampleDir, "-o", outputPath}); err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	result, err := exec.Command(outputPath).CombinedOutput()
	if err != nil {
		t.Fatalf("built example binary failed: %v\n%s", err, string(result))
	}
	if strings.TrimSpace(string(result)) != "Hello from FLAG" {
		t.Fatalf("unexpected output from example binary: %q", result)
	}
}
