package compiler

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// CompiledPackage is one Go file produced by CompileProgramPackages.
type CompiledPackage struct {
	Dir    string // subdirectory under the generated module; empty for package main
	Name   string // Go package clause
	Source []byte
	FlagI  []byte // macros + (declare …) interface; nil for package main
	IsMain bool
}

// ProloguePackageName is the Go package for internal/compiler/prologue.flag.
const ProloguePackageName = "prologue"

// CompileProgramPackages loads entryPath and emits one Go package per FLAG module.
// The entry module is always package main; libraries are imported as
// flagbuild/<pkg>. The language prelude is one more library package
// (flagbuild/prologue); it is not copied into importers.
func CompileProgramPackages(entryPath string) ([]CompiledPackage, error) {
	prog, err := LoadProgram(entryPath)
	if err != nil {
		return nil, err
	}

	prologueCtx, err := newPackagePrologueContext()
	if err != nil {
		return nil, err
	}
	prologueExports := prologueExportNames(prologueCtx)
	prologueResult := withPrologue(compileResult{namespace: ProloguePackageName}, prologueCtx)
	prologueResult.vars = append(prologueResult.vars, prologueCtx.constants.decls()...)
	prologueSrc, err := emitGoFile(ProloguePackageName, prologueResult, false)
	if err != nil {
		return nil, err
	}

	shared, err := newCompileContext()
	if err != nil {
		return nil, err
	}
	shared.packageMode = true

	moduleDefs := map[string]map[string]string{}
	moduleMacros := map[string]map[string]Expr{}
	byPath := prog.byPath
	usedPkgs := map[string]string{ProloguePackageName: "<prologue>"}
	out := []CompiledPackage{{
		Dir:    ProloguePackageName,
		Name:   ProloguePackageName,
		Source: prologueSrc,
		FlagI:  emitPrologueFlagI(prologueCtx),
	}}

	for _, mod := range prog.Modules {
		ctx := copyCompileContext(shared)
		ctx.constants = newConstantInterner()
		ctx.packageMode = true
		if mod.HasModuleHeader {
			ctx.namespace = mod.Header.Namespace
			ctx.exportedNames = moduleExportSet(mod)
			if err := seedImports(&ctx, mod, byPath, moduleDefs, moduleMacros); err != nil {
				return nil, err
			}
		} else if mod.LegacyNS != "" {
			ctx.namespace = mod.LegacyNS
		}
		stripInlinedPrologue(&ctx, shared)
		if err := seedProloguePackageRefs(&ctx, prologueExports); err != nil {
			return nil, err
		}

		allowTopLevel := mod == prog.Entry
		partial, defined, definedMacros, err := compileModuleBody(mod, &ctx, allowTopLevel)
		if err != nil {
			if mod.Path != "" {
				return nil, fmt.Errorf("%s: %w", mod.Path, err)
			}
			return nil, err
		}
		if mod.HasModuleHeader {
			if err := validateExports(mod, defined, definedMacros); err != nil {
				return nil, fmt.Errorf("%s: %w", mod.Path, err)
			}
		}

		defs := map[string]string{}
		for name, goName := range defined {
			defs[name] = goName
		}
		moduleDefs[mod.Path] = defs
		moduleMacros[mod.Path] = definedMacros

		partial.vars = append(partial.vars, ctx.constants.decls()...)
		if mod.HasModuleHeader {
			partial.namespace = mod.Header.Namespace
		} else if mod.LegacyNS != "" {
			partial.namespace = mod.LegacyNS
		}
		if resultRefsPackage(partial, ProloguePackageName+".") {
			ctx.addGoImport(ProloguePackageName)
		}
		partial.extraImports = append([]goImport(nil), ctx.goImports...)

		isEntry := mod == prog.Entry
		pkgName := "main"
		dir := ""
		if !isEntry {
			pkgName, err = modulePackageName(mod)
			if err != nil {
				return nil, err
			}
			if prev, ok := usedPkgs[pkgName]; ok {
				return nil, fmt.Errorf("duplicate Go package name %q from %s and %s", pkgName, prev, mod.Path)
			}
			usedPkgs[pkgName] = mod.Path
			dir = pkgName
		}

		src, err := emitGoFile(pkgName, partial, isEntry)
		if err != nil {
			return nil, err
		}
		pkg := CompiledPackage{Dir: dir, Name: pkgName, Source: src, IsMain: isEntry}
		if !isEntry {
			pkg.FlagI = emitModuleFlagI(mod, defined, definedMacros)
		}
		out = append(out, pkg)
	}
	return out, nil
}

func newPackagePrologueContext() (compileContext, error) {
	ifTemps := 0
	ctx := compileContext{
		functions:     make(map[string]functionDef),
		globals:       make(map[string]exprKind),
		moduleSymbols: make(map[string]string),
		recordTypes:   make(map[string]string),
		ifTemps:       &ifTemps,
		packageMode:   true,
		namespace:     ProloguePackageName,
		exportedNames: map[string]bool{},
		constants:     newConstantInterner(),
	}
	if err := loadStandardPrologue(&ctx); err != nil {
		return compileContext{}, err
	}
	return ctx, nil
}

func prologueExportNames(ctx compileContext) map[string]string {
	out := map[string]string{}
	for _, fn := range ctx.prologueFns {
		out[fn.flagName] = fn.goName
	}
	for _, v := range ctx.prologueVars {
		if v.flagName != "" {
			out[v.flagName] = v.goName
		}
	}
	return out
}

func stripInlinedPrologue(ctx *compileContext, shared compileContext) {
	clearName := func(flagName, goName string) {
		delete(ctx.functions, goName)
		delete(ctx.globals, goName)
		if flagName != "" {
			// Keep :refer / :as bindings (pkg.Name). Prelude also defines
			// second/third; deleting those keys would rebind them to prologue.
			if prev, ok := ctx.moduleSymbols[flagName]; !ok || !strings.Contains(prev, ".") {
				delete(ctx.moduleSymbols, flagName)
			}
			if ctx.namespace != "" {
				qual := ctx.namespace + "/" + flagName
				if prev, ok := ctx.moduleSymbols[qual]; !ok || !strings.Contains(prev, ".") {
					delete(ctx.moduleSymbols, qual)
				}
			}
		}
	}
	for _, fn := range shared.prologueFns {
		clearName(fn.flagName, fn.goName)
	}
	for _, v := range shared.prologueVars {
		clearName(v.flagName, v.goName)
	}
	ctx.prologueFns = nil
	ctx.prologueVars = nil
}

func seedProloguePackageRefs(ctx *compileContext, exports map[string]string) error {
	for name, goName := range exports {
		if prev, ok := ctx.moduleSymbols[name]; ok && strings.Contains(prev, ".") {
			continue
		}
		ref := ProloguePackageName + "." + goName
		if ctx.moduleSymbols == nil {
			ctx.moduleSymbols = map[string]string{}
		}
		ctx.moduleSymbols[name] = ref
		if ctx.namespace != "" {
			ctx.moduleSymbols[ctx.namespace+"/"+name] = ref
		}
	}
	return nil
}

func resultRefsPackage(result compileResult, prefix string) bool {
	for _, fn := range result.functions {
		if strings.Contains(renderFunctionDef(fn), prefix) {
			return true
		}
	}
	for _, v := range result.vars {
		if strings.Contains(v.expr, prefix) {
			return true
		}
	}
	for _, stmt := range result.stmts {
		if strings.Contains(stmt.code, prefix) || strings.Contains(stmt.prelude, prefix) {
			return true
		}
	}
	for _, decl := range result.typeDecls {
		if strings.Contains(decl, prefix) {
			return true
		}
	}
	return false
}

func modulePackageName(mod *Module) (string, error) {
	ns := displayNamespace(mod)
	if ns == "" {
		return "", fmt.Errorf("%s: package emit requires :namespace", mod.Path)
	}
	return goPackageName(ns)
}

func emitModuleFlagI(mod *Module, defined map[string]string, macros map[string]Expr) []byte {
	ns := displayNamespace(mod)
	exports := []string{}
	if mod.HasModuleHeader {
		exports = append(exports, mod.Header.Exports...)
	} else {
		for name := range defined {
			exports = append(exports, name)
		}
		for name := range macros {
			exports = append(exports, name)
		}
	}
	var goExports map[string]string
	if mod.HasModuleHeader {
		goExports = mod.Header.GoExports
	}
	return emitFlagI(ns, exports, goExports, macros, defined)
}

func emitPrologueFlagI(ctx compileContext) []byte {
	exports := make([]string, 0, len(ctx.prologueFns)+len(ctx.prologueVars)+len(ctx.macroForms))
	defined := map[string]string{}
	for _, fn := range ctx.prologueFns {
		exports = append(exports, fn.flagName)
		defined[fn.flagName] = fn.goName
	}
	for _, v := range ctx.prologueVars {
		if v.flagName == "" {
			continue
		}
		if _, ok := defined[v.flagName]; ok {
			continue
		}
		exports = append(exports, v.flagName)
		defined[v.flagName] = v.goName
	}
	macros := map[string]Expr{}
	for _, form := range ctx.macroForms {
		if name, ok := defmacroFormName(form); ok {
			macros[name] = form
			exports = append(exports, name)
		}
	}
	return emitFlagI(ProloguePackageName, exports, nil, macros, defined)
}

func emitFlagI(ns string, exports []string, goExports map[string]string, macros map[string]Expr, defined map[string]string) []byte {
	seen := map[string]bool{}
	unique := make([]string, 0, len(exports))
	for _, name := range exports {
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		unique = append(unique, name)
	}
	sort.Strings(unique)

	var b strings.Builder
	b.WriteString("{:namespace ")
	b.WriteString(strconv.Quote(ns))
	if len(unique) > 0 {
		b.WriteString("\n :exports   [")
		b.WriteString(strings.Join(unique, " "))
		b.WriteString("]")
	}
	if len(goExports) > 0 {
		keys := make([]string, 0, len(goExports))
		for k := range goExports {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		b.WriteString("\n :go-exports {")
		for i, k := range keys {
			if i > 0 {
				b.WriteByte(' ')
			}
			b.WriteString(k)
			b.WriteByte(' ')
			b.WriteString(strconv.Quote(goExports[k]))
		}
		b.WriteByte('}')
	}
	b.WriteString("}\n")

	macroNames := make([]string, 0, len(macros))
	for name := range macros {
		macroNames = append(macroNames, name)
	}
	sort.Strings(macroNames)
	for _, name := range macroNames {
		b.WriteByte('\n')
		b.WriteString(exprToSourceString(macros[name]))
		b.WriteByte('\n')
	}

	decls := make([]string, 0, len(unique))
	for _, name := range unique {
		if _, isMacro := macros[name]; isMacro {
			continue
		}
		if _, ok := defined[name]; ok || goExports[name] != "" {
			decls = append(decls, name)
		}
	}
	if len(decls) > 0 {
		b.WriteString("\n(declare ")
		b.WriteString(strings.Join(decls, " "))
		b.WriteString(")\n")
	}
	return []byte(b.String())
}

// WriteProgramPackages compiles entryPath to one Go package per FLAG module
// under dir, plus a go.mod that requires flag-lang via a replace directive.
func WriteProgramPackages(dir, entryPath string) error {
	pkgs, err := CompileProgramPackages(entryPath)
	if err != nil {
		return err
	}
	root, err := findFlagLangRoot(entryPath)
	if err != nil {
		return err
	}
	goMod := fmt.Sprintf("module %s\n\ngo 1.26.4\n\nrequire flag-lang v0.0.0\n\nreplace flag-lang => %s\n",
		GeneratedModulePath, root)
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goMod), 0o644); err != nil {
		return fmt.Errorf("write go.mod: %w", err)
	}
	sumSrc := filepath.Join(root, "go.sum")
	if data, err := os.ReadFile(sumSrc); err == nil {
		if err := os.WriteFile(filepath.Join(dir, "go.sum"), data, 0o644); err != nil {
			return fmt.Errorf("write go.sum: %w", err)
		}
	}
	for _, pkg := range pkgs {
		if pkg.IsMain {
			if err := os.WriteFile(filepath.Join(dir, "main.go"), pkg.Source, 0o644); err != nil {
				return fmt.Errorf("write main.go: %w", err)
			}
			continue
		}
		pdir := filepath.Join(dir, pkg.Dir)
		if err := os.MkdirAll(pdir, 0o755); err != nil {
			return fmt.Errorf("mkdir %s: %w", pdir, err)
		}
		path := filepath.Join(pdir, pkg.Name+".go")
		if err := os.WriteFile(path, pkg.Source, 0o644); err != nil {
			return fmt.Errorf("write %s: %w", path, err)
		}
		if len(pkg.FlagI) > 0 {
			ipath := filepath.Join(pdir, pkg.Name+".flagi")
			if err := os.WriteFile(ipath, pkg.FlagI, 0o644); err != nil {
				return fmt.Errorf("write %s: %w", ipath, err)
			}
		}
	}
	tidy := exec.Command("go", "mod", "tidy")
	tidy.Dir = dir
	if output, err := tidy.CombinedOutput(); err != nil {
		return fmt.Errorf("go mod tidy: %w\n%s", err, strings.TrimSpace(string(output)))
	}
	return nil
}

func findFlagLangRoot(start string) (string, error) {
	abs, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}
	dir := abs
	if info, err := os.Stat(abs); err == nil && !info.IsDir() {
		dir = filepath.Dir(abs)
	}
	seen := map[string]bool{}
	search := func(d string) string {
		for {
			if seen[d] {
				return ""
			}
			seen[d] = true
			b, err := os.ReadFile(filepath.Join(d, "go.mod"))
			if err == nil {
				line, _, _ := strings.Cut(string(b), "\n")
				if strings.TrimSpace(line) == "module flag-lang" {
					return d
				}
			}
			parent := filepath.Dir(d)
			if parent == d {
				return ""
			}
			d = parent
		}
	}
	if root := search(dir); root != "" {
		return root, nil
	}
	cwd, err := os.Getwd()
	if err == nil {
		if root := search(cwd); root != "" {
			return root, nil
		}
	}
	return "", fmt.Errorf("could not find flag-lang module root (go.mod with module flag-lang)")
}
