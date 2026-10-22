package compiler

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	flagrt "flag-lang/runtime"
)

func init() {
	flagrt.RegisterGoFunction("compiler/module-import-paths", ModuleImportPaths)
	flagrt.RegisterGoFunction("compiler/expand-forms", ExpandFormsHost)
	flagrt.RegisterGoFunction("compiler/go-bind-ident", GoBindIdent)
	flagrt.RegisterGoFunction("compiler/prologue-path", PrologueFilePath)
}

// PrologueFilePath is the language prelude module every program implicitly imports.
func PrologueFilePath() (string, error) {
	candidates := []string{}
	if _, file, _, ok := runtime.Caller(0); ok {
		candidates = append(candidates, filepath.Join(filepath.Dir(file), "prologue.flag"))
	}
	if cwd, err := os.Getwd(); err == nil {
		candidates = append(candidates, filepath.Join(cwd, "internal", "compiler", "prologue.flag"))
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
	}
	return "", fmt.Errorf("prologue.flag not found (tried %s)", strings.Join(tried, ", "))
}

// ModuleImportPaths returns :imports path strings from a FLAG source file.
func ModuleImportPaths(path string) ([]string, error) {
	mod, err := parseModuleTokenStream(path, TokenizeFileToChannel(path))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	if !mod.HasModuleHeader {
		return []string{}, nil
	}
	out := make([]string, 0, len(mod.Header.Imports))
	for _, spec := range mod.Header.Imports {
		out = append(out, spec.Path)
	}
	return out, nil
}

// ExpandFormsHost expands FLAG AST maps with seed defmacros, matching expandFLAGForms.
func ExpandFormsHost(seed, forms flagrt.Value) (flagrt.Value, error) {
	seedExprs, err := flagExprSeq(seed)
	if err != nil {
		return flagrt.Value{}, err
	}
	formExprs, err := flagExprSeq(forms)
	if err != nil {
		return flagrt.Value{}, err
	}
	expanded, err := expandFLAGForms(seedExprs, formExprs)
	if err != nil {
		return flagrt.Value{}, err
	}
	return exprsToFlagVector(expanded), nil
}

// GoBindIdent returns the flagrt.GoBind_* identifier for a :go-exports host key.
func GoBindIdent(host string) (string, error) {
	ident, ok := libraryGoBinds[host]
	if !ok {
		return "", fmt.Errorf("unknown :go-exports host bind %q", host)
	}
	return ident, nil
}
