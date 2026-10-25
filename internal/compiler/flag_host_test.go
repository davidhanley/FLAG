package compiler

import (
	"go/format"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestModuleImportPaths(t *testing.T) {
	paths, err := ModuleImportPaths(filepath.Join("..", "..", "examples", "modules", "main.flag"))
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) != 3 {
		t.Fatalf("imports %#v", paths)
	}
	if paths[0] != "math.flag" || paths[1] != "greeter.flag" || paths[2] != "greeter.flag" {
		t.Fatalf("imports %#v", paths)
	}
}

func TestModuleHasHeader(t *testing.T) {
	dir := t.TempDir()
	header := filepath.Join(dir, "mod.flag")
	if err := os.WriteFile(header, []byte("{:namespace \"demo\"}\n(defn x [] 1)\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	ok, err := ModuleHasHeader(header)
	if err != nil || !ok {
		t.Fatalf("header: ok=%v err=%v", ok, err)
	}
	bare := filepath.Join(dir, "bare.flag")
	if err := os.WriteFile(bare, []byte("(defn x [] 1)\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	ok, err = ModuleHasHeader(bare)
	if err != nil || ok {
		t.Fatalf("bare: ok=%v err=%v", ok, err)
	}
}

func TestModuleImportPathsNoHeader(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bare.flag")
	if err := os.WriteFile(path, []byte("(defn x [] 1)\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	paths, err := ModuleImportPaths(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) != 0 {
		t.Fatalf("got %#v", paths)
	}
}

func TestFLAGResolveImportPathMatchesGo(t *testing.T) {
	importer := filepath.Join("..", "..", "examples", "modules", "main.flag")
	got, err := resolveImportPath(importer, "math.flag")
	if err != nil {
		t.Fatal(err)
	}
	want, err := resolveImportPathGo(importer, "math.flag")
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("FLAG %q Go %q", got, want)
	}
	noExt, err := resolveImportPath(importer, "math")
	if err != nil {
		t.Fatal(err)
	}
	if noExt != want {
		t.Fatalf("no-ext %q want %q", noExt, want)
	}
	libImporter := filepath.Join("..", "..", "examples", "compiler_tokenizer", "main.flag")
	lib, err := resolveImportPath(libImporter, "compiler/lower.lib")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(lib, filepath.Join("libraries", "compiler", "lower.lib")) {
		t.Fatalf("library import %q", lib)
	}
	if _, err := resolveImportPath(importer, ""); err == nil || !strings.Contains(err.Error(), "empty import path") {
		t.Fatalf("empty import: %v", err)
	}
}

func TestPrologueFilePath(t *testing.T) {
	got, err := PrologueFilePath()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(got, "prologue.flag") {
		t.Fatalf("path %q", got)
	}
	if _, err := os.Stat(got); err != nil {
		t.Fatal(err)
	}
}

func TestFLAGLoadProgramMatchesGo(t *testing.T) {
	entry := filepath.Join("..", "..", "examples", "modules", "main.flag")
	got, err := LoadProgram(entry)
	if err != nil {
		t.Fatal(err)
	}
	want, err := loadProgramGo(entry)
	if err != nil {
		t.Fatal(err)
	}
	pro, err := PrologueFilePath()
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Modules) == 0 || got.Modules[0].Path != pro {
		t.Fatalf("FLAG first module %#v want prologue %q", modulePaths(got), pro)
	}
	if len(got.Modules) != len(want.Modules)+1 {
		t.Fatalf("module count FLAG %d Go %d", len(got.Modules), len(want.Modules))
	}
	for i := range want.Modules {
		if got.Modules[i+1].Path != want.Modules[i].Path {
			t.Fatalf("order[%d] FLAG %q Go %q", i, got.Modules[i+1].Path, want.Modules[i].Path)
		}
	}
	if got.Entry.Path != want.Entry.Path {
		t.Fatalf("entry FLAG %q Go %q", got.Entry.Path, want.Entry.Path)
	}
}

func modulePaths(p *Program) []string {
	if p == nil {
		return nil
	}
	out := make([]string, 0, len(p.Modules))
	for _, m := range p.Modules {
		out = append(out, m.Path)
	}
	return out
}

func TestFLAGLoadProgramRejectsDirectory(t *testing.T) {
	_, err := LoadProgram(t.TempDir())
	if err == nil || !strings.Contains(err.Error(), "file entry") {
		t.Fatalf("got %v", err)
	}
}

func TestModuleImportPathsParseError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.flag")
	if err := os.WriteFile(path, []byte("{:namespace\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := ModuleImportPaths(path)
	if err == nil || !strings.Contains(err.Error(), path) {
		t.Fatalf("got %v", err)
	}
}

func TestGoBindIdent(t *testing.T) {
	got, err := GoBindIdent("burp/html")
	if err != nil || got != "GoBind_burp_Html" {
		t.Fatalf("got %q err %v", got, err)
	}
	if _, err := GoBindIdent("not/a-bind"); err == nil || !strings.Contains(err.Error(), "unknown :go-exports host bind") {
		t.Fatalf("unknown bind: %v", err)
	}
}

func TestFLAGGoPackageNameMatchesGo(t *testing.T) {
	for _, ns := range []string{"math", "type", "c.frs.core"} {
		got, err := goPackageName(ns)
		if err != nil {
			t.Fatal(err)
		}
		want, err := goPackageNameGo(ns)
		if err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Fatalf("%s: FLAG %q Go %q", ns, got, want)
		}
	}
}

func TestFLAGValidateExports(t *testing.T) {
	mod := &Module{
		HasModuleHeader: true,
		Header:          ModuleHeader{Namespace: "hello", Exports: []string{"greet"}},
	}
	if err := validateExports(mod, map[string]string{"greet": "hello__greet"}, nil); err != nil {
		t.Fatal(err)
	}
	err := validateExports(mod, map[string]string{"other": "x"}, nil)
	if err == nil || !strings.Contains(err.Error(), `export "greet"`) {
		t.Fatalf("got %v", err)
	}
}

func TestFLAGEmitFlagI(t *testing.T) {
	got := string(emitFlagI("math", []string{"double", "add"}, nil, nil, map[string]string{"add": "Add", "double": "Double"}))
	want := string(emitFlagIGo("math", []string{"double", "add"}, nil, nil, map[string]string{"add": "Add", "double": "Double"}))
	if got != want {
		t.Fatalf("FLAG:\n%s\nGo:\n%s", got, want)
	}
}

func TestFLAGSeedImportsRefer(t *testing.T) {
	entry := filepath.Join("..", "..", "examples", "modules", "main.flag")
	prog, err := LoadProgram(entry)
	if err != nil {
		t.Fatal(err)
	}
	var mathMod, mainMod *Module
	for _, mod := range prog.Modules {
		switch filepath.Base(mod.Path) {
		case "math.flag":
			mathMod = mod
		case "main.flag":
			mainMod = mod
		}
	}
	if mathMod == nil || mainMod == nil {
		t.Fatalf("missing modules")
	}
	ctx := compileContext{namespace: "main", moduleSymbols: map[string]string{}}
	defs := map[string]map[string]string{}
	macros := map[string]map[string]Expr{}
	for _, mod := range prog.Modules {
		if mod == mainMod {
			continue
		}
		local := map[string]string{}
		for _, name := range mod.Header.Exports {
			local[name] = "math__" + name
		}
		defs[mod.Path] = local
		macros[mod.Path] = map[string]Expr{}
	}
	if err := seedImports(&ctx, mainMod, prog.byPath, defs, macros); err != nil {
		t.Fatal(err)
	}
	if ctx.moduleSymbols["math/add"] != "math__add" {
		t.Fatalf("module symbols %#v", ctx.moduleSymbols)
	}
}

func TestFLAGCompileInlineProgram(t *testing.T) {
	entry := filepath.Join("..", "..", "examples", "modules", "main.flag")
	got, err := flagCompileInlineProgram(entry, nil)
	if err != nil {
		t.Fatal(err)
	}
	want, err := compileProgramGo(entry, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.functions) == 0 {
		t.Fatal("FLAG inline compile produced no functions")
	}
	if len(got.functions) != len(want.functions) {
		t.Fatalf("functions FLAG %d Go %d", len(got.functions), len(want.functions))
	}
}

func TestFLAGCompileProgramPackages(t *testing.T) {
	entry := filepath.Join("..", "..", "examples", "modules", "main.flag")
	got, err := flagCompileProgramPackages(entry, nil)
	if err != nil {
		t.Fatal(err)
	}
	want, err := compileProgramPackagesGo(entry)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != len(want) {
		t.Fatalf("packages FLAG %d Go %d", len(got), len(want))
	}
	for i, pkg := range got {
		if pkg.Name != want[i].Name || pkg.Dir != want[i].Dir || pkg.IsMain != want[i].IsMain {
			t.Fatalf("pkg[%d] FLAG %+v Go %+v", i, pkg, want[i])
		}
		if _, err := format.Source(pkg.Source); err != nil {
			t.Fatalf("FLAG package %s format: %v\n%s", pkg.Name, err, pkg.Source)
		}
		if !pkg.IsMain && len(pkg.FlagI) == 0 {
			t.Fatalf("FLAG package %s missing .flagi", pkg.Name)
		}
	}
}

func TestFLAGFindFlagLangRoot(t *testing.T) {
	entry := filepath.Join("..", "..", "examples", "modules", "main.flag")
	got, err := flagFindFlagLangRoot(entry)
	if err != nil {
		t.Fatal(err)
	}
	want, err := findFlagLangRoot(entry)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("FLAG %q Go %q", got, want)
	}
	mod, err := os.ReadFile(filepath.Join(got, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(mod), "module flag-lang") {
		t.Fatalf("root go.mod:\n%s", mod)
	}
}

func TestFLAGWriteCompiledPackages(t *testing.T) {
	dir := t.TempDir()
	pkgs := []CompiledPackage{
		{Name: "main", Source: []byte("package main\n"), IsMain: true},
		{Dir: "math", Name: "math", Source: []byte("package math\n"), FlagI: []byte("{:namespace \"math\"}\n")},
		{Dir: "empty", Name: "empty", Source: []byte("package empty\n")},
	}
	if err := flagWriteCompiledPackages(dir, pkgs); err != nil {
		t.Fatal(err)
	}
	if err := flagWriteCompiledPackages(dir, nil); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{
		filepath.Join(dir, "main.go"),
		filepath.Join(dir, "math", "math.go"),
		filepath.Join(dir, "math", "math.flagi"),
		filepath.Join(dir, "empty", "empty.go"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected %s: %v", path, err)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "empty", "empty.flagi")); !os.IsNotExist(err) {
		t.Fatalf("empty.flagi should be absent: %v", err)
	}
	mainSrc, err := os.ReadFile(filepath.Join(dir, "main.go"))
	if err != nil {
		t.Fatal(err)
	}
	if string(mainSrc) != "package main\n" {
		t.Fatalf("main.go %q", mainSrc)
	}
}
