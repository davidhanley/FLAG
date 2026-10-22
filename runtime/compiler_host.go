package runtime

import (
	"fmt"
	"go/format"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func init() {
	RegisterGoFunction("compiler/format-go-source", FormatGoSource)
	RegisterGoFunction("compiler/abs-path", AbsPath)
	RegisterGoFunction("compiler/cwd-path", CwdPath)
	RegisterGoFunction("compiler/path-join", PathJoin)
	RegisterGoFunction("compiler/path-dir", PathDir)
	RegisterGoFunction("compiler/path-ext", PathExt)
	RegisterGoFunction("compiler/path-abs?", PathIsAbs)
	RegisterGoFunction("compiler/path-stat-kind", PathStatKind)
	RegisterGoFunction("compiler/write-text-file", WriteTextFile)
	RegisterGoFunction("compiler/read-text-file", ReadTextFile)
	RegisterGoFunction("compiler/mkdir-all", MkdirAll)
	RegisterGoFunction("compiler/go-mod-tidy", GoModTidy)
}

// FormatGoSource runs go/format on generated Go source.
func FormatGoSource(src string) (string, error) {
	formatted, err := format.Source([]byte(src))
	if err != nil {
		return "", fmt.Errorf("format generated Go: %w", err)
	}
	return string(formatted), nil
}

// AbsPath is filepath.Abs for FLAG compiler host helpers.
func AbsPath(path string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return "", fmt.Errorf("empty path")
	}
	return filepath.Abs(path)
}

// CwdPath is os.Getwd for FLAG compiler host helpers.
func CwdPath() (string, error) {
	return os.Getwd()
}

// PathJoin joins two path elements.
func PathJoin(a, b string) string {
	return filepath.Join(a, b)
}

// PathDir is filepath.Dir.
func PathDir(path string) string {
	return filepath.Dir(path)
}

// PathExt is filepath.Ext.
func PathExt(path string) string {
	return filepath.Ext(path)
}

// PathIsAbs is filepath.IsAbs.
func PathIsAbs(path string) bool {
	return filepath.IsAbs(path)
}

// PathStatKind returns "file", "dir", or "missing".
func PathStatKind(path string) string {
	st, err := os.Stat(path)
	if err != nil {
		return "missing"
	}
	if st.IsDir() {
		return "dir"
	}
	return "file"
}

// WriteTextFile writes contents to path (0644).
func WriteTextFile(path, contents string) error {
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("empty path")
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

// ReadTextFile reads a whole file as a string.
func ReadTextFile(path string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return "", fmt.Errorf("empty path")
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read %s: %w", path, err)
	}
	return string(b), nil
}

// MkdirAll creates dir and parents (0755).
func MkdirAll(path string) error {
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("empty path")
	}
	if err := os.MkdirAll(path, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", path, err)
	}
	return nil
}

// GoModTidy runs `go mod tidy` in dir.
func GoModTidy(dir string) error {
	if strings.TrimSpace(dir) == "" {
		return fmt.Errorf("empty path")
	}
	tidy := exec.Command("go", "mod", "tidy")
	tidy.Dir = dir
	if output, err := tidy.CombinedOutput(); err != nil {
		return fmt.Errorf("go mod tidy: %w\n%s", err, strings.TrimSpace(string(output)))
	}
	return nil
}
