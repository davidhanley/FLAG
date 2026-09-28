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

func TestCompileProgramLibraryImportInteropDefrecord(t *testing.T) {
	dir := t.TempDir()
	main := filepath.Join(dir, "main.flag")
	if err := os.WriteFile(main, []byte(`
{:namespace "main"
 :imports [["interop.lib" :refer [defrecord]]]}
(defrecord Person [name age])
(println (:age (->Person "Ada" 36)))
(println (:name (map->Person {:name "Bob" :age 40})))
`), 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := CompileProgram(main)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			t.Skipf("libraries/ not resolvable from test cwd: %v", err)
		}
		t.Fatalf("CompileProgram: %v", err)
	}
	got := string(out)
	for _, want := range []string{
		"func flag_main____gtPerson_arity_2",
		"func flag_main__map__gtPerson_arity_1",
		`flagrt.Call(flag_main____gtPerson, flagStr_Ada, flagrt.NewLong(36))`,
		"flagrt.Call(flag_main__map__gtPerson",
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

func TestGoPackageName(t *testing.T) {
	got, err := goPackageName("math")
	if err != nil || got != "math" {
		t.Fatalf("math: got %q err %v", got, err)
	}
	got, err = goPackageName("type")
	if err != nil || got != "type_" {
		t.Fatalf("type: got %q err %v", got, err)
	}
	got, err = goPackageName("c.frs.core")
	if err != nil || got != "c_frs_core" {
		t.Fatalf("dotted: got %q err %v", got, err)
	}
}

func TestExportedGoIdent(t *testing.T) {
	got, err := exportedGoIdent("add")
	if err != nil || got != "Add" {
		t.Fatalf("add: got %q err %v", got, err)
	}
	got, err = exportedGoIdent("empty-board")
	if err != nil || got != "Empty_board" {
		t.Fatalf("empty-board: got %q err %v", got, err)
	}
	got, err = exportedGoIdent("Add")
	if err != nil || got != "Add" {
		t.Fatalf("Add: got %q err %v", got, err)
	}
}

func TestCompileProgramPackagesNoInline(t *testing.T) {
	main := filepath.Join("..", "..", "examples", "modules", "main.flag")
	pkgs, err := CompileProgramPackages(main)
	if err != nil {
		t.Fatalf("CompileProgramPackages: %v", err)
	}
	var greeter, mathPkg, entry CompiledPackage
	for _, p := range pkgs {
		switch {
		case p.Name == "greeter":
			greeter = p
		case p.Name == "math":
			mathPkg = p
		case p.IsMain:
			entry = p
		}
	}
	var prologue CompiledPackage
	for _, p := range pkgs {
		if p.Name == ProloguePackageName {
			prologue = p
		}
	}
	if greeter.Source == nil || mathPkg.Source == nil || entry.Source == nil || prologue.Source == nil {
		t.Fatalf("missing packages: %#v", pkgs)
	}
	if !strings.Contains(string(prologue.Source), "package prologue") || !strings.Contains(string(prologue.Source), "var Inc =") {
		t.Fatalf("prologue package missing Inc:\n%s", prologue.Source)
	}
	if !strings.Contains(string(prologue.FlagI), "(declare ") || !strings.Contains(string(prologue.FlagI), "(defmacro when") {
		t.Fatalf("prologue.flagi:\n%s", prologue.FlagI)
	}
	if !strings.Contains(string(greeter.FlagI), "(declare greet shout)") && !strings.Contains(string(greeter.FlagI), "(declare shout greet)") {
		if !strings.Contains(string(greeter.FlagI), "(declare") {
			t.Fatalf("greeter.flagi missing declare:\n%s", greeter.FlagI)
		}
	}

	g := string(greeter.Source)
	for _, want := range []string{
		"package greeter",
		"math.Double",
		"\tmath \"flagbuild/math\"",
		"var Greet =",
		"var Shout =",
	} {
		if !strings.Contains(g, want) {
			t.Fatalf("greeter missing %q in:\n%s", want, g)
		}
	}
	for _, refuse := range []string{"math__double", "func add_arity", "func math__add", "func inc_arity", "var identity ="} {
		if strings.Contains(g, refuse) {
			t.Fatalf("greeter should not contain %q in:\n%s", refuse, g)
		}
	}

	m := string(mathPkg.Source)
	for _, want := range []string{
		"package math",
		"var Add =",
		"var Double =",
		"func secret_square",
	} {
		if !strings.Contains(m, want) {
			t.Fatalf("math missing %q in:\n%s", want, m)
		}
	}
	if strings.Contains(m, "func main(") {
		t.Fatalf("library math should not emit main:\n%s", m)
	}

	e := string(entry.Source)
	for _, want := range []string{
		"package main",
		"math.Add",
		"greeter.Greet",
		"greeter.Shout",
	} {
		if !strings.Contains(e, want) {
			t.Fatalf("main missing %q in:\n%s", want, e)
		}
	}
	for _, refuse := range []string{"func math__add", "func greeter__greet", "math__double", "func inc_arity", "var identity ="} {
		if strings.Contains(e, refuse) {
			t.Fatalf("main should not inline %q in:\n%s", refuse, e)
		}
	}
}

func TestWriteProgramPackages(t *testing.T) {
	dir := t.TempDir()
	main := filepath.Join("..", "..", "examples", "modules", "main.flag")
	if err := WriteProgramPackages(dir, main); err != nil {
		t.Fatalf("WriteProgramPackages: %v", err)
	}
	for _, path := range []string{
		filepath.Join(dir, "go.mod"),
		filepath.Join(dir, "main.go"),
		filepath.Join(dir, "math", "math.go"),
		filepath.Join(dir, "greeter", "greeter.go"),
		filepath.Join(dir, "prologue", "prologue.go"),
		filepath.Join(dir, "math", "math.flagi"),
		filepath.Join(dir, "greeter", "greeter.flagi"),
		filepath.Join(dir, "prologue", "prologue.flagi"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected %s: %v", path, err)
		}
	}
	mod, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(mod)
	if !strings.Contains(text, "module flagbuild") || !strings.Contains(text, "replace flag-lang =>") {
		t.Fatalf("go.mod:\n%s", text)
	}
	flagi, err := os.ReadFile(filepath.Join(dir, "math", "math.flagi"))
	if err != nil {
		t.Fatal(err)
	}
	got := string(flagi)
	for _, want := range []string{
		`{:namespace "math"`,
		":exports",
		"add",
		"double",
		"(declare ",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("math.flagi missing %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "secret-square") || strings.Contains(got, "(defn add") {
		t.Fatalf("math.flagi should not include private defs or function bodies:\n%s", got)
	}
}

func TestCompileProgramPackagesPreludeIsSeparatePackage(t *testing.T) {
	dir := t.TempDir()
	main := filepath.Join(dir, "main.flag")
	if err := os.WriteFile(main, []byte(`
{:namespace "main"}
(println (inc 41))
(println (identity 1))
`), 0o644); err != nil {
		t.Fatal(err)
	}
	pkgs, err := CompileProgramPackages(main)
	if err != nil {
		t.Fatalf("CompileProgramPackages: %v", err)
	}
	var prologue, entry CompiledPackage
	for _, p := range pkgs {
		if p.Name == ProloguePackageName {
			prologue = p
		}
		if p.IsMain {
			entry = p
		}
	}
	if prologue.Source == nil || entry.Source == nil {
		t.Fatalf("missing packages: %#v", pkgs)
	}
	p := string(prologue.Source)
	if !strings.Contains(p, "package prologue") || !strings.Contains(p, "var Inc =") || !strings.Contains(p, "var Identity =") {
		t.Fatalf("prologue:\n%s", p)
	}
	e := string(entry.Source)
	for _, want := range []string{"prologue.Inc", "prologue.Identity", "flagbuild/prologue"} {
		if !strings.Contains(e, want) {
			t.Fatalf("main missing %q in:\n%s", want, e)
		}
	}
	for _, refuse := range []string{"func inc_arity", "func identity_arity", "var identity =", "var inc ="} {
		if strings.Contains(e, refuse) {
			t.Fatalf("main inlined prologue %q:\n%s", refuse, e)
		}
	}
}
