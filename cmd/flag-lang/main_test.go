package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"flag-lang/internal/compiler"
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

func TestRunCompileWritesDefaultGoFile(t *testing.T) {
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "hello.flag")
	defaultOutput := filepath.Join(dir, "hello.go")

	source := `(ns hello.core)
(println "Hello from FLAG")
`
	if err := os.WriteFile(inputPath, []byte(source), 0o644); err != nil {
		t.Fatalf("WriteFile input: %v", err)
	}

	if err := run([]string{"compile", inputPath}); err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	generated, err := os.ReadFile(defaultOutput)
	if err != nil {
		t.Fatalf("ReadFile output: %v", err)
	}

	if !strings.Contains(string(generated), `fmt.Println(flagrt.Str("Hello from FLAG"))`) {
		t.Fatalf("unexpected output:\n%s", generated)
	}
}

func TestRunBuildMapcatConcatenatesMappedCollections(t *testing.T) {
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "mapcat.flag")
	outputPath := filepath.Join(dir, "mapcatbin")

	source := `(println (vec (mapcat (fn [x] [x (* x 10)]) [1 2 3])))
(println (vec (mapcat (fn [a b] [a b]) [1 2] [10 20])))
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
	got := strings.TrimSpace(string(result))
	want := "[1 10 2 20 3 30]\n[1 10 2 20]"
	if got != want {
		t.Fatalf("unexpected mapcat output: %q", got)
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

func TestRunBuildGoFormRunsAsync(t *testing.T) {
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "async.flag")
	outputPath := filepath.Join(dir, "async-bin")
	// Requires libraries/async.lib on the search path (repo root / go.mod walk).
	source := `
{:namespace "demo"
 :imports [["async.lib" :refer [go sleep]]]}
(go (do (sleep 30) (println "from-go")))
(println "from-main")
(sleep 100)
`
	if err := os.WriteFile(inputPath, []byte(source), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if err := run([]string{"build", inputPath, "-o", outputPath}); err != nil {
		t.Fatalf("build: %v", err)
	}
	result, err := exec.Command(outputPath).CombinedOutput()
	if err != nil {
		t.Fatalf("run: %v\n%s", err, result)
	}
	out := string(result)
	if !strings.Contains(out, "from-main") {
		t.Fatalf("expected from-main in output:\n%s", out)
	}
	if !strings.Contains(out, "from-go") {
		t.Fatalf("expected from-go in output:\n%s", out)
	}
}

func TestRunBuildCompilerTokenizerLibrary(t *testing.T) {
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "main.flag")
	sourcePath := filepath.Join(dir, "sample.flag")
	outputPath := filepath.Join(dir, "tokenizer-bin")

	source := "(+ x 10)\nfoo[bar]\n"
	if err := os.WriteFile(sourcePath, []byte(source), 0o644); err != nil {
		t.Fatalf("WriteFile source: %v", err)
	}

	program := `
{:namespace "main"
 :imports [["compiler/tokenizer.lib" :as "c"]
           ["async.lib" :refer [channel-receive]]]}
(defn drain [ch]
  (let [token (channel-receive ch)]
    (if (nil? token)
      nil
      (do
        (println (:token token) ":" (:line token) ":" (:offset token))
        (drain ch)))))
(defn main [& _args]
  (drain (c/tokenize-file "` + strings.ReplaceAll(sourcePath, "\\", "\\\\") + `")))
`

	if err := os.WriteFile(inputPath, []byte(program), 0o644); err != nil {
		t.Fatalf("WriteFile program: %v", err)
	}

	if err := run([]string{"build", inputPath, "-o", outputPath}); err != nil {
		t.Fatalf("build: %v", err)
	}

	result, err := exec.Command(outputPath).CombinedOutput()
	if err != nil {
		t.Fatalf("tokenizer binary failed: %v\n%s", err, result)
	}
	out := string(result)
	for _, want := range []string{
		"(:1:1",
		"+:1:2",
		"x:1:4",
		"10:1:6",
		"):1:8",
		"foo:2:1",
		"[:2:4",
		"bar:2:5",
		"]:2:8",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in output:\n%s", want, out)
		}
	}
}

func TestRunBuildCompilerTokenizerMatchesGoTokenizer(t *testing.T) {
	exampleDir, err := filepath.Abs(filepath.Join("..", "..", "examples", "compiler_tokenizer"))
	if err != nil {
		t.Fatalf("Abs exampleDir: %v", err)
	}

	fixtures, err := filepath.Glob(filepath.Join(exampleDir, "*.txt"))
	if err != nil {
		t.Fatalf("Glob fixtures: %v", err)
	}
	sort.Strings(fixtures)
	if len(fixtures) == 0 {
		t.Fatal("no compiler_tokenizer fixtures found")
	}

	dir := t.TempDir()
	inputPath := filepath.Join(dir, "tokenizer_parity.flag")
	outputPath := filepath.Join(dir, "tokenizer-parity-bin")
	program := `
{:namespace "main"
 :imports [["compiler/tokenizer.lib" :as "c"]
           ["async.lib" :refer [channel-receive]]]}
(defn drain [ch]
  (let [token (channel-receive ch)]
    (if (nil? token)
      nil
      (do
        (println (format "%q|%d|%d"
                         (:token token)
                         (:line token)
                         (:offset token)))
        (drain ch)))))
(defn main [& argv]
  (drain (c/tokenize-file (first argv))))
`
	if err := os.WriteFile(inputPath, []byte(program), 0o644); err != nil {
		t.Fatalf("WriteFile parity program: %v", err)
	}

	if err := run([]string{"build", inputPath, "-o", outputPath}); err != nil {
		t.Fatalf("build parity tokenizer binary: %v", err)
	}

	for _, fixture := range fixtures {
		fixture := fixture
		t.Run(filepath.Base(fixture), func(t *testing.T) {
			source, err := os.ReadFile(fixture)
			if err != nil {
				t.Fatalf("ReadFile fixture: %v", err)
			}

			want := goTokenizerOutput(string(source))
			got := runFlagTokenizerOutput(t, outputPath, fixture)
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("tokenizer mismatch for %s\nwant: %#v\ngot:  %#v", fixture, want, got)
			}
		})
	}
}

func goTokenizerOutput(source string) []compiler.SourceToken {
	tokens := make([]compiler.SourceToken, 0, 32)
	for tok := range compiler.TokenizeSourceToChannel(source) {
		tokens = append(tokens, tok)
	}
	return tokens
}

func runFlagTokenizerOutput(t *testing.T, binaryPath, sourcePath string) []compiler.SourceToken {
	t.Helper()

	result, err := exec.Command(binaryPath, sourcePath).CombinedOutput()
	if err != nil {
		t.Fatalf("flag tokenizer binary failed for %s: %v\n%s", sourcePath, err, result)
	}
	out := strings.TrimSuffix(string(result), "\n")
	if out == "" {
		return []compiler.SourceToken{}
	}

	lines := strings.Split(out, "\n")
	tokens := make([]compiler.SourceToken, 0, len(lines))
	for _, line := range lines {
		token, err := parseFlagTokenizerLine(line)
		if err != nil {
			t.Fatalf("parse tokenizer line %q: %v", line, err)
		}
		tokens = append(tokens, token)
	}
	return tokens
}

func parseFlagTokenizerLine(line string) (compiler.SourceToken, error) {
	var (
		tokenText string
		lineNum   int
		offset    int
	)
	n, err := fmt.Sscanf(line, "%q|%d|%d", &tokenText, &lineNum, &offset)
	if err != nil {
		return compiler.SourceToken{}, fmt.Errorf("scan token line: %w", err)
	}
	if n != 3 {
		return compiler.SourceToken{}, fmt.Errorf("expected 3 parsed fields, got %d", n)
	}
	return compiler.SourceToken{
		Token:  tokenText,
		Line:   int64(lineNum),
		Offset: int64(offset),
	}, nil
}

// Acceptance: examples/compiler_tokenizer FLAG tests for tokenizer behavior.
func TestAcceptanceCompilerTokenizerFlagTests(t *testing.T) {
	exampleDir, err := filepath.Abs(filepath.Join("..", "..", "examples", "compiler_tokenizer"))
	if err != nil {
		t.Fatalf("Abs exampleDir: %v", err)
	}
	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("Abs repoRoot: %v", err)
	}
	prevWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	if err := os.Chdir(repoRoot); err != nil {
		t.Fatalf("Chdir repoRoot: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(prevWD)
	})
	cleanup := func() {
		_ = os.Remove(filepath.Join(exampleDir, "main.go"))
		_ = os.Remove(filepath.Join(exampleDir, "compiler_tokenizer.go"))
	}
	cleanup()
	t.Cleanup(cleanup)

	if err := run([]string{"test", exampleDir}); err != nil {
		t.Fatalf("flag-lang test examples/compiler_tokenizer: %v", err)
	}
}

// Acceptance: examples/concurrency FLAG tests (sleep, go, fib) via `flag-lang test`.
func TestAcceptanceConcurrencyFlagTests(t *testing.T) {
	exampleDir, err := filepath.Abs(filepath.Join("..", "..", "examples", "concurrency"))
	if err != nil {
		t.Fatalf("Abs exampleDir: %v", err)
	}
	cleanup := func() {
		_ = os.Remove(filepath.Join(exampleDir, "main.go"))
		_ = os.Remove(filepath.Join(exampleDir, "concurrency.go"))
	}
	cleanup()
	t.Cleanup(cleanup)

	if err := run([]string{"test", exampleDir}); err != nil {
		t.Fatalf("flag-lang test examples/concurrency: %v", err)
	}
}

// Acceptance: build and run the concurrency demo binary.
func TestAcceptanceConcurrencyDemo(t *testing.T) {
	exampleDir, err := filepath.Abs(filepath.Join("..", "..", "examples", "concurrency"))
	if err != nil {
		t.Fatalf("Abs exampleDir: %v", err)
	}
	entry := filepath.Join(exampleDir, "main.flag")
	outputPath := filepath.Join(t.TempDir(), "concurrency-bin")

	cleanup := func() {
		_ = os.Remove(filepath.Join(exampleDir, "main.go"))
		_ = os.Remove(filepath.Join(exampleDir, "concurrency.go"))
	}
	cleanup()
	t.Cleanup(cleanup)

	if err := run([]string{"build", entry, "-o", outputPath}); err != nil {
		t.Fatalf("build concurrency example: %v", err)
	}
	result, err := exec.Command(outputPath).CombinedOutput()
	if err != nil {
		t.Fatalf("concurrency binary failed: %v\n%s", err, result)
	}
	out := string(result)
	for _, want := range []string{
		"main: start",
		"go: computing fib(28)...",
		"go: fib(28) = 317811",
		"go: later task done",
		"future: computing fib(20)...",
		"main: future started (not waiting yet)",
		"main: future result = 6765",
		"main: future again = 6765",
		"channel: sending 7",
		"channel: received 7",
		"select: processed 2 channels, sum = 30",
		"main: kicked off work",
		"main: end",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in output:\n%s", want, out)
		}
	}
}

// Acceptance: examples/http FLAG tests for local server routing and client calls.
func TestAcceptanceHTTPFlagTests(t *testing.T) {
	exampleDir, err := filepath.Abs(filepath.Join("..", "..", "examples", "http"))
	if err != nil {
		t.Fatalf("Abs exampleDir: %v", err)
	}
	cleanup := func() {
		_ = os.Remove(filepath.Join(exampleDir, "main.go"))
		_ = os.Remove(filepath.Join(exampleDir, "http.go"))
	}
	cleanup()
	t.Cleanup(cleanup)

	if err := run([]string{"test", exampleDir}); err != nil {
		t.Fatalf("flag-lang test examples/http: %v", err)
	}
}

// Acceptance: build and run the HTTP demo binary.
func TestAcceptanceHTTPDemo(t *testing.T) {
	exampleDir, err := filepath.Abs(filepath.Join("..", "..", "examples", "http"))
	if err != nil {
		t.Fatalf("Abs exampleDir: %v", err)
	}
	entry := filepath.Join(exampleDir, "main.flag")
	outputPath := filepath.Join(t.TempDir(), "http-bin")

	cleanup := func() {
		_ = os.Remove(filepath.Join(exampleDir, "main.go"))
		_ = os.Remove(filepath.Join(exampleDir, "http.go"))
	}
	cleanup()
	t.Cleanup(cleanup)

	if err := run([]string{"build", entry, "-o", outputPath}); err != nil {
		t.Fatalf("build http example: %v", err)
	}
	result, err := exec.Command(outputPath).CombinedOutput()
	if err != nil {
		t.Fatalf("http binary failed: %v\n%s", err, result)
	}
	out := string(result)
	for _, want := range []string{
		"server: 127.0.0.1:",
		"GET / -> 200 home",
		"GET /hello/:name -> 200 hello flag from FLAG",
		"POST /echo -> 201 echo payload method=POST",
		"GET /static/*path -> 200 static css/site.css",
		"DELETE /echo -> 405 method not allowed",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in output:\n%s", want, out)
		}
	}
}

func TestRunBuildExampleModules(t *testing.T) {
	// Multi-file module graph: math -> greeter -> main (qualified, :as, :refer).
	exampleDir, err := filepath.Abs(filepath.Join("..", "..", "examples", "modules"))
	if err != nil {
		t.Fatalf("Abs exampleDir: %v", err)
	}
	entry := filepath.Join(exampleDir, "main.flag")
	outputPath := filepath.Join(t.TempDir(), "modules-bin")

	// Remove any inspection .go artifacts so this directory is never a Go package.
	cleanupModulesGo := func() {
		_ = os.Remove(filepath.Join(exampleDir, "main.go"))
		_ = os.Remove(filepath.Join(exampleDir, "modules.go"))
	}
	cleanupModulesGo()
	t.Cleanup(cleanupModulesGo)

	if err := run([]string{"build", entry, "-o", outputPath}); err != nil {
		t.Fatalf("build modules entry: %v", err)
	}

	result, err := exec.Command(outputPath).CombinedOutput()
	if err != nil {
		t.Fatalf("modules binary failed: %v\n%s", err, string(result))
	}
	want := "5\nHello, world x2 = 10\nHI FLAG"
	if strings.TrimSpace(string(result)) != want {
		t.Fatalf("unexpected modules output:\n got: %q\nwant: %q", result, want)
	}
}

func TestRunBuildExampleModulesDirectory(t *testing.T) {
	// Directory build should pick up main.flag as the modular entry.
	exampleDir, err := filepath.Abs(filepath.Join("..", "..", "examples", "modules"))
	if err != nil {
		t.Fatalf("Abs exampleDir: %v", err)
	}
	outputPath := filepath.Join(t.TempDir(), "modules-dir-bin")

	cleanupModulesGo := func() {
		_ = os.Remove(filepath.Join(exampleDir, "main.go"))
		_ = os.Remove(filepath.Join(exampleDir, "modules.go"))
	}
	cleanupModulesGo()
	t.Cleanup(cleanupModulesGo)

	if err := run([]string{"build", exampleDir, "-o", outputPath}); err != nil {
		t.Fatalf("build modules directory: %v", err)
	}

	result, err := exec.Command(outputPath).CombinedOutput()
	if err != nil {
		t.Fatalf("modules directory binary failed: %v\n%s", err, string(result))
	}
	want := "5\nHello, world x2 = 10\nHI FLAG"
	if strings.TrimSpace(string(result)) != want {
		t.Fatalf("unexpected modules directory output:\n got: %q\nwant: %q", result, want)
	}
}

func TestRunTestRunsAssociatedTests(t *testing.T) {
	dir := t.TempDir()
	sourcePath := filepath.Join(dir, "math.flag")
	testPath := filepath.Join(dir, "math_test.flag")

	source := `(ns math.core)
(defn add1 [x] (+ x 1))
`
	tests := `(ns math.core-test)
(deftest add1-test
  (is (= (add1 2) 3)))
`
	if err := os.WriteFile(sourcePath, []byte(source), 0o644); err != nil {
		t.Fatalf("WriteFile source: %v", err)
	}
	if err := os.WriteFile(testPath, []byte(tests), 0o644); err != nil {
		t.Fatalf("WriteFile test: %v", err)
	}

	if err := run([]string{"test", dir}); err != nil {
		t.Fatalf("run returned error: %v", err)
	}
}

func TestRunTestRemapsTestFileErrors(t *testing.T) {
	dir := t.TempDir()
	sourcePath := filepath.Join(dir, "math.flag")
	testPath := filepath.Join(dir, "math_test.flag")

	source := `(ns math.core)
(defn add1 [x] (+ x 1))
`
	tests := `(ns math.core-test)
(deftest bad-test
	  (println {:a (< 1 2)}))
`
	if err := os.WriteFile(sourcePath, []byte(source), 0o644); err != nil {
		t.Fatalf("WriteFile source: %v", err)
	}
	if err := os.WriteFile(testPath, []byte(tests), 0o644); err != nil {
		t.Fatalf("WriteFile test: %v", err)
	}

	err := run([]string{"test", dir})
	if err == nil {
		t.Fatal("run succeeded unexpectedly")
	}
	if !strings.Contains(err.Error(), "at 3:") {
		t.Fatalf("expected remapped line number in error, got: %v", err)
	}
}
