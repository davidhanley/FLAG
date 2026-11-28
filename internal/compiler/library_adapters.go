package compiler

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const libraryAdapterPkg = "flagbind"

func goBindPkg(host string) string {
	if strings.HasPrefix(host, "csv/") {
		return libraryAdapterPkg
	}
	return runtimeAlias
}

func goBindSelector(host string) (string, error) {
	ident, ok := libraryGoBinds[host]
	if !ok {
		return "", fmt.Errorf("unknown :go-exports host bind %q", host)
	}
	return goBindPkg(host) + "." + ident, nil
}

func ensureLibraryAdapterImports(pkgs []CompiledPackage) []CompiledPackage {
	out := make([]CompiledPackage, len(pkgs))
	copy(out, pkgs)
	for i, pkg := range out {
		if !bytes.Contains(pkg.Source, []byte(libraryAdapterPkg+".")) {
			continue
		}
		if bytes.Contains(pkg.Source, []byte(GeneratedModulePath+"/"+libraryAdapterPkg)) {
			src := pkg.Source
			if !bytes.Contains(src, []byte("flagrt.")) {
				src = keepFlagrtImport(src)
			}
			out[i].Source = src
			continue
		}
		old := []byte("import (\n")
		neu := []byte("import (\n\t" + libraryAdapterPkg + " \"" + GeneratedModulePath + "/" + libraryAdapterPkg + "\"\n")
		if bytes.Contains(pkg.Source, old) {
			src := bytes.Replace(pkg.Source, old, neu, 1)
			if !bytes.Contains(src, []byte("flagrt.")) {
				src = keepFlagrtImport(src)
			}
			out[i].Source = src
		}
	}
	return out
}

func keepFlagrtImport(src []byte) []byte {
	needle := []byte(")\n\n")
	idx := bytes.Index(src, needle)
	if idx < 0 {
		return src
	}
	insert := []byte(")\n\nvar _ = flagrt.NilValue\n")
	return append(append([]byte{}, src[:idx]...), append(insert, src[idx+len(needle):]...)...)
}

// CopyLibraryAdapters copies MUSE library flagbind/*.go sources into the
// generated module so :go-exports can reference flagbind.GoBind_*.
func CopyLibraryAdapters(buildDir string) error {
	if err := appendLibraryGoModuleReplaces(buildDir); err != nil {
		return err
	}
	dest := filepath.Join(buildDir, libraryAdapterPkg)
	for _, root := range libraryAdapterSearchRoots() {
		src := filepath.Join(root, libraryAdapterPkg)
		ok, err := copyGoPackage(src, dest)
		if err != nil {
			return err
		}
		if ok {
			return nil
		}
	}
	return nil
}

func appendLibraryGoModuleReplaces(buildDir string) error {
	root, err := findFlagLangRoot(buildDir)
	if err != nil {
		cwd, cwdErr := os.Getwd()
		if cwdErr != nil {
			return nil
		}
		root, err = findFlagLangRoot(cwd)
		if err != nil {
			return nil
		}
	}
	extras := libraryGoModuleReplaces(root)
	if extras == "" {
		return nil
	}
	modPath := filepath.Join(buildDir, "go.mod")
	data, err := os.ReadFile(modPath)
	if err != nil {
		return nil
	}
	if err := os.WriteFile(modPath, append(data, []byte(extras)...), 0o644); err != nil {
		return fmt.Errorf("write go.mod library replaces: %w", err)
	}
	return nil
}

func libraryAdapterSearchRoots() []string {
	var roots []string
	if cwd, err := os.Getwd(); err == nil {
		if flagRoot, err := findFlagLangRoot(cwd); err == nil {
			roots = append(roots, filepath.Join(flagRoot, "testdata", "flag-csv"))
		}
	}
	roots = append(roots, ExtraLibraryRoots()...)
	return roots
}

func ExtraLibraryRoots() []string {
	extraLibraryRootsMu.Lock()
	defer extraLibraryRootsMu.Unlock()
	return append([]string(nil), extraLibraryRoots...)
}

func copyGoPackage(src, dest string) (bool, error) {
	st, err := os.Stat(src)
	if err != nil || !st.IsDir() {
		return false, nil
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return false, err
	}
	copied := false
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(src, name))
		if err != nil {
			return false, err
		}
		if err := os.MkdirAll(dest, 0o755); err != nil {
			return false, err
		}
		if err := os.WriteFile(filepath.Join(dest, name), data, 0o644); err != nil {
			return false, err
		}
		copied = true
	}
	return copied, nil
}
