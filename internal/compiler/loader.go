package compiler

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Program is a loaded module graph rooted at an entry file.
type Program struct {
	Entry   *Module
	Modules []*Module // dependency order (providers before dependents), entry last
	byPath  map[string]*Module
}

// LoadProgram loads entryPath and its import closure.
func LoadProgram(entryPath string) (*Program, error) {
	abs, err := filepath.Abs(entryPath)
	if err != nil {
		return nil, fmt.Errorf("resolve %s: %w", entryPath, err)
	}
	info, err := os.Stat(abs)
	if err != nil {
		return nil, fmt.Errorf("stat %s: %w", entryPath, err)
	}
	if info.IsDir() {
		return nil, fmt.Errorf("LoadProgram expects a file entry, got directory %s (pass main.flag or similar)", entryPath)
	}

	prog := &Program{
		byPath: map[string]*Module{},
	}
	stack := map[string]bool{}
	order := make([]*Module, 0, 8)

	var load func(path string) (*Module, error)
	load = func(path string) (*Module, error) {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return nil, fmt.Errorf("resolve %s: %w", path, err)
		}
		if mod, ok := prog.byPath[absPath]; ok {
			return mod, nil
		}
		if stack[absPath] {
			return nil, fmt.Errorf("circular import involving %s", absPath)
		}
		stack[absPath] = true
		defer delete(stack, absPath)

		source, err := os.ReadFile(absPath)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", absPath, err)
		}
		mod, err := parseModuleFile(absPath, string(source))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", absPath, err)
		}
		if mod.HasModuleHeader {
			for _, spec := range mod.Header.Imports {
				resolved, err := resolveImportPath(absPath, spec.Path)
				if err != nil {
					return nil, fmt.Errorf("%s: import %q: %w", absPath, spec.Path, err)
				}
				if _, err := load(resolved); err != nil {
					return nil, err
				}
			}
		} else if len(mod.Header.Imports) > 0 {
			// unreachable: imports only exist with header
		}
		prog.byPath[absPath] = mod
		order = append(order, mod)
		return mod, nil
	}

	entry, err := load(abs)
	if err != nil {
		return nil, err
	}
	prog.Entry = entry
	prog.Modules = order
	return prog, nil
}

// resolveImportPath resolves an import path relative to the importing file,
// the working directory, and project libraries/ directories (for "burp.lib").
func resolveImportPath(importerPath, importPath string) (string, error) {
	importPath = strings.TrimSpace(importPath)
	if importPath == "" {
		return "", fmt.Errorf("empty import path")
	}

	candidates := []string{}
	if filepath.IsAbs(importPath) {
		candidates = append(candidates, importPath)
	} else {
		dir := filepath.Dir(importerPath)
		candidates = append(candidates, filepath.Join(dir, importPath))
		// cwd-relative (project-root style imports).
		if cwd, err := os.Getwd(); err == nil {
			candidates = append(candidates, filepath.Join(cwd, importPath))
		}
		// libraries/ search roots (repo or package that contains libraries/).
		for _, root := range librarySearchRoots(importerPath) {
			candidates = append(candidates, filepath.Join(root, "libraries", importPath))
		}
	}

	var tried []string
	for _, c := range candidates {
		abs, err := filepath.Abs(c)
		if err != nil {
			continue
		}
		tried = append(tried, abs)
		if st, err := os.Stat(abs); err == nil && !st.IsDir() {
			return abs, nil
		}
		// Allow omitting extension: try .flag, .lib, then .flaglib
		ext := filepath.Ext(abs)
		if ext == "" {
			for _, e := range []string{".flag", ".lib", ".flaglib"} {
				withExt := abs + e
				tried = append(tried, withExt)
				if st, err := os.Stat(withExt); err == nil && !st.IsDir() {
					return withExt, nil
				}
			}
		}
	}
	return "", fmt.Errorf("not found (tried %s)", strings.Join(tried, ", "))
}

// librarySearchRoots returns directories that may contain a libraries/ folder.
// Walks from the importer and cwd toward filesystem roots, preferring trees
// that already have libraries/ or a go.mod (project root).
func librarySearchRoots(importerPath string) []string {
	starts := []string{}
	if importerPath != "" {
		starts = append(starts, filepath.Dir(importerPath))
	}
	if cwd, err := os.Getwd(); err == nil {
		starts = append(starts, cwd)
	}

	seen := map[string]bool{}
	var roots []string
	add := func(dir string) {
		abs, err := filepath.Abs(dir)
		if err != nil || seen[abs] {
			return
		}
		seen[abs] = true
		roots = append(roots, abs)
	}

	for _, start := range starts {
		dir := start
		for {
			libDir := filepath.Join(dir, "libraries")
			if st, err := os.Stat(libDir); err == nil && st.IsDir() {
				add(dir)
			}
			if st, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil && !st.IsDir() {
				add(dir)
				break
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}
	return roots
}

// moduleExportSet returns the set of exported local names for a module.
func moduleExportSet(mod *Module) map[string]bool {
	out := map[string]bool{}
	if mod == nil || !mod.HasModuleHeader {
		return out
	}
	for _, name := range mod.Header.Exports {
		out[name] = true
	}
	return out
}

// displayNamespace returns the namespace string used for diagnostics.
func displayNamespace(mod *Module) string {
	if mod == nil {
		return ""
	}
	if mod.HasModuleHeader {
		return mod.Header.Namespace
	}
	return mod.LegacyNS
}

// appendTestForms loads test source files and appends their body forms to the
// entry module so deftest can see private definitions (same-module tests).
func appendTestForms(entry *Module, testPaths []string) error {
	if entry == nil {
		return fmt.Errorf("appendTestForms: nil entry module")
	}
	for _, path := range testPaths {
		abs, err := filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("resolve test %s: %w", path, err)
		}
		source, err := os.ReadFile(abs)
		if err != nil {
			return fmt.Errorf("read test %s: %w", abs, err)
		}
		testMod, err := parseModuleFile(abs, string(source))
		if err != nil {
			return fmt.Errorf("%s: %w", abs, err)
		}
		if testMod.HasModuleHeader && len(testMod.Header.Imports) > 0 {
			return fmt.Errorf("%s: test modules cannot declare :imports (they run in the entry module)", abs)
		}
		entry.Forms = append(entry.Forms, testMod.Forms...)
	}
	return nil
}

// ModuleHasImports reports whether a source file's header declares :imports.
func ModuleHasImports(path string) (bool, error) {
	source, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	mod, err := parseModuleFile(path, string(source))
	if err != nil {
		return false, err
	}
	return mod.HasModuleHeader && len(mod.Header.Imports) > 0, nil
}
