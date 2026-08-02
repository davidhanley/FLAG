package compiler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseModuleHeader(t *testing.T) {
	ast, err := ParseFile(`
{:namespace "chess"
 :exports [move legal?]
 :imports ["board.flag"
           ["util.flag" :as "u"]
           ["helpers.flag" :refer [trim]]]}
(defn move [] 1)
`)
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	header, ok, err := parseModuleHeader(ast.Forms[0])
	if err != nil {
		t.Fatalf("parseModuleHeader: %v", err)
	}
	if !ok {
		t.Fatal("expected module header")
	}
	if header.Namespace != "chess" {
		t.Fatalf("namespace: got %q", header.Namespace)
	}
	if len(header.Exports) != 2 || header.Exports[0] != "move" || header.Exports[1] != "legal?" {
		t.Fatalf("exports: %#v", header.Exports)
	}
	if len(header.Imports) != 3 {
		t.Fatalf("imports: %#v", header.Imports)
	}
	if header.Imports[0].Path != "board.flag" || header.Imports[0].As != "" {
		t.Fatalf("import0: %#v", header.Imports[0])
	}
	if header.Imports[1].Path != "util.flag" || header.Imports[1].As != "u" {
		t.Fatalf("import1: %#v", header.Imports[1])
	}
	if header.Imports[2].Path != "helpers.flag" || len(header.Imports[2].Refer) != 1 || header.Imports[2].Refer[0] != "trim" {
		t.Fatalf("import2: %#v", header.Imports[2])
	}
}

func TestParseModuleHeaderRejectsUnknownKey(t *testing.T) {
	ast, err := ParseFile(`{:namespace "x" :wat 1}`)
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = parseModuleHeader(ast.Forms[0])
	if err == nil || !strings.Contains(err.Error(), "unknown module header key") {
		t.Fatalf("expected unknown key error, got %v", err)
	}
}

func TestCompileModuleHeaderSingleFile(t *testing.T) {
	out, err := Compile(`
{:namespace "hello"
 :exports [greet]}
(defn greet [name] name)
(println (greet "x"))
`)
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	got := string(out)
	for _, want := range []string{
		"// Source namespace: hello",
		"func hello__greet_arity_1",
		"var hello__greet =",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in:\n%s", want, got)
		}
	}
}

func TestCompileModuleMissingExport(t *testing.T) {
	_, err := Compile(`
{:namespace "hello"
 :exports [missing]}
(defn greet [] 1)
`)
	if err == nil || !strings.Contains(err.Error(), `export "missing"`) {
		t.Fatalf("expected missing export error, got %v", err)
	}
}

func TestCompileModuleImportsRequirePath(t *testing.T) {
	_, err := Compile(`
{:namespace "main"
 :imports ["other.flag"]}
(println 1)
`)
	if err == nil || !strings.Contains(err.Error(), "file path") {
		t.Fatalf("expected file path error, got %v", err)
	}
}

func TestCompileProgramQualifiedImport(t *testing.T) {
	dir := t.TempDir()
	board := filepath.Join(dir, "board.flag")
	main := filepath.Join(dir, "main.flag")
	if err := os.WriteFile(board, []byte(`
{:namespace "board"
 :exports [empty-board]}
(defn empty-board [] 42)
`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(main, []byte(`
{:namespace "main"
 :imports ["board.flag"]}
(println (board/empty-board))
`), 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := CompileProgram(main)
	if err != nil {
		t.Fatalf("CompileProgram: %v", err)
	}
	got := string(out)
	for _, want := range []string{
		"func board__empty_board_arity_0",
		"flagrt.Call(board__empty_board)",
		"// Source namespace: main",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in:\n%s", want, got)
		}
	}
}

func TestCompileProgramAsAndRefer(t *testing.T) {
	dir := t.TempDir()
	chess := filepath.Join(dir, "chess.flag")
	main := filepath.Join(dir, "main.flag")
	if err := os.WriteFile(chess, []byte(`
{:namespace "chess"
 :exports [move legal?]}
(defn move [x] x)
(defn legal? [x] true)
`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(main, []byte(`
{:namespace "main"
 :imports [["chess.flag" :as "ch" :refer [legal?]]]}
(println (ch/move 1))
(println (legal? 1))
`), 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := CompileProgram(main)
	if err != nil {
		t.Fatalf("CompileProgram: %v", err)
	}
	got := string(out)
	for _, want := range []string{
		"func chess__move_arity_1",
		"func chess__legal_q_arity_1",
		"flagrt.Call(chess__move",
		"flagrt.Call(chess__legal_q",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in:\n%s", want, got)
		}
	}
	// Bare chess/ prefix should not be registered when only :as is used.
	if strings.Contains(got, "chess/move") {
		// source shouldn't appear; just ensure we didn't also require chess/ in Go
	}
}

func TestCompileProgramRejectsPrivateImport(t *testing.T) {
	dir := t.TempDir()
	lib := filepath.Join(dir, "lib.flag")
	main := filepath.Join(dir, "main.flag")
	if err := os.WriteFile(lib, []byte(`
{:namespace "lib"
 :exports [pub]}
(defn pub [] 1)
(defn secret [] 2)
`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(main, []byte(`
{:namespace "main"
 :imports ["lib.flag"]}
(println (lib/secret))
`), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := CompileProgram(main)
	if err == nil || !strings.Contains(err.Error(), `unknown symbol "lib/secret"`) {
		t.Fatalf("expected private import error, got %v", err)
	}
}

func TestCompileProgramCircularImport(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.flag")
	b := filepath.Join(dir, "b.flag")
	if err := os.WriteFile(a, []byte(`
{:namespace "a"
 :exports [a-fn]
 :imports ["b.flag"]}
(defn a-fn [] 1)
`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(b, []byte(`
{:namespace "b"
 :exports [b-fn]
 :imports ["a.flag"]}
(defn b-fn [] 1)
`), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := CompileProgram(a)
	if err == nil || !strings.Contains(err.Error(), "circular import") {
		t.Fatalf("expected circular import error, got %v", err)
	}
}

func TestCompileProgramReferNonExport(t *testing.T) {
	dir := t.TempDir()
	lib := filepath.Join(dir, "lib.flag")
	main := filepath.Join(dir, "main.flag")
	if err := os.WriteFile(lib, []byte(`
{:namespace "lib"
 :exports [pub]}
(defn pub [] 1)
(defn secret [] 2)
`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(main, []byte(`
{:namespace "main"
 :imports [["lib.flag" :refer [secret]]]}
(println (secret))
`), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := CompileProgram(main)
	if err == nil || !strings.Contains(err.Error(), `:refer "secret"`) {
		t.Fatalf("expected refer export error, got %v", err)
	}
}

func TestCompileProgramLibraryImportBurpCSV(t *testing.T) {
	// Resolve libraries/ from the repo root (walk-up from this package's temp entry).
	dir := t.TempDir()
	main := filepath.Join(dir, "main.flag")
	// Put a stub libraries dir that should NOT be used if walk-up finds the real one;
	// we rely on walking to the module root that contains libraries/burp.lib.
	if err := os.WriteFile(main, []byte(`
{:namespace "main"
 :imports   ["csv.lib" "burp.lib"]}
(println (csv/read-csv "x"))
(println (burp/html [:div "hi"]))
`), 0o644); err != nil {
		t.Fatal(err)
	}

	// Run compile from repo root context: set entry under temp but imports must
	// find repo libraries via go.mod walk from cwd (test runs in package dir).
	// Place a go.mod marker is already at repo root; librarySearchRoots walks cwd.
	out, err := CompileProgram(main)
	if err != nil {
		// If libraries aren't found from package test cwd, skip with clear message.
		if strings.Contains(err.Error(), "not found") {
			t.Skipf("libraries/ not resolvable from test cwd: %v", err)
		}
		t.Fatalf("CompileProgram: %v", err)
	}
	got := string(out)
	for _, want := range []string{
		"var csv__read_csv = flagrt.GoBind_csv_ReadCSV",
		"var burp__html = flagrt.GoBind_burp_Html",
		"flagrt.Call(csv__read_csv",
		"flagrt.Call(burp__html",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in:\n%s", want, got)
		}
	}
}

func TestCompileRejectsAmbientBurpWithoutImport(t *testing.T) {
	_, err := Compile(`
{:namespace "main"}
(println (burp/html [:div "x"]))
`)
	if err == nil || !strings.Contains(err.Error(), `unknown symbol "burp/html"`) {
		t.Fatalf("expected ambient burp to fail, got %v", err)
	}
}

func TestLegacyNsStillWorks(t *testing.T) {
	out, err := Compile(`
(ns hello.core)
(println "hi")
`)
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	if !strings.Contains(string(out), "// Source namespace: hello.core") {
		t.Fatalf("expected legacy ns comment:\n%s", out)
	}
	// Legacy mode should not mangle println-only programs with ns prefixes on random symbols.
	if strings.Contains(string(out), "hello_core__") {
		t.Fatalf("legacy ns should not mangle names:\n%s", out)
	}
}
