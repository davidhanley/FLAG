// Package muse is FLAG’s build manifest: a .muse file lists git libraries to
// fetch, cache, and put on the compiler’s import path.
package muse

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"flag-lang/internal/compiler"
)

// Manifest is the FLAG map stored in a .muse file.
//
//	{:libraries ["csv" "https://github.com/acme/flag-http.git"]}
type Manifest struct {
	Path      string
	Libraries []string
}

// Prepare parses musePath, fetches each library into the local cache, and
// materializes *.lib files under <project>/libraries so import resolution
// finds them. It returns the project directory (the .muse file’s parent).
func Prepare(musePath string) (string, error) {
	abs, err := filepath.Abs(musePath)
	if err != nil {
		return "", fmt.Errorf("resolve %s: %w", musePath, err)
	}
	man, err := ParseFile(abs)
	if err != nil {
		return "", err
	}
	project := filepath.Dir(abs)
	var roots []string
	for _, lib := range man.Libraries {
		checkout, err := Fetch(lib)
		if err != nil {
			return "", fmt.Errorf("library %q: %w", lib, err)
		}
		roots = append(roots, checkout)
		if err := materialize(project, checkout, libraryName(lib)); err != nil {
			return "", fmt.Errorf("library %q: %w", lib, err)
		}
	}
	compiler.SetExtraLibraryRoots(roots)
	return project, nil
}

// Find returns the .muse file that applies to path, if any.
// A .muse path is used as-is. A directory (or a FLAG file’s directory) may
// contain project.muse or exactly one *.muse file.
func Find(path string) (string, bool) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", false
	}
	st, err := os.Stat(abs)
	if err != nil {
		return "", false
	}
	if !st.IsDir() {
		if strings.EqualFold(filepath.Ext(abs), ".muse") {
			return abs, true
		}
		abs = filepath.Dir(abs)
	}
	return findInDir(abs)
}

func findInDir(dir string) (string, bool) {
	preferred := filepath.Join(dir, "project.muse")
	if st, err := os.Stat(preferred); err == nil && !st.IsDir() {
		return preferred, true
	}
	matches, err := filepath.Glob(filepath.Join(dir, "*.muse"))
	if err != nil || len(matches) == 0 {
		return "", false
	}
	if len(matches) == 1 {
		return matches[0], true
	}
	return "", false
}

// ParseFile reads a .muse FLAG map.
func ParseFile(path string) (*Manifest, error) {
	source, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	man, err := Parse(string(source))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	man.Path = path
	return man, nil
}

// Parse interprets source as a single FLAG map.
func Parse(source string) (*Manifest, error) {
	ast, err := compiler.ParseFile(source)
	if err != nil {
		return nil, err
	}
	if len(ast.Forms) == 0 {
		return nil, fmt.Errorf("empty muse file")
	}
	if len(ast.Forms) > 1 {
		return nil, fmt.Errorf("muse file must contain a single map")
	}
	m, ok := ast.Forms[0].(compiler.MapExpr)
	if !ok {
		return nil, fmt.Errorf("muse file must be a FLAG map")
	}
	if len(m.Entries)%2 != 0 {
		return nil, fmt.Errorf("muse map expects even number of forms")
	}
	man := &Manifest{}
	for i := 0; i < len(m.Entries); i += 2 {
		keyExpr := m.Entries[i]
		valExpr := m.Entries[i+1]
		key, ok := keyExpr.(compiler.KeywordExpr)
		if !ok {
			return nil, fmt.Errorf("muse keys must be keywords")
		}
		switch key.Name {
		case "libraries":
			libs, err := parseStringVector(valExpr)
			if err != nil {
				return nil, err
			}
			man.Libraries = libs
		default:
			return nil, fmt.Errorf("unknown muse key :%s", key.Name)
		}
	}
	return man, nil
}

func parseStringVector(expr compiler.Expr) ([]string, error) {
	vec, ok := expr.(compiler.VectorExpr)
	if !ok {
		return nil, fmt.Errorf(":libraries must be a vector of strings")
	}
	out := make([]string, 0, len(vec.Elements))
	for _, item := range vec.Elements {
		s, ok := item.(compiler.StringExpr)
		if !ok {
			return nil, fmt.Errorf(":libraries entries must be strings")
		}
		name := strings.TrimSpace(s.Value)
		if name == "" {
			return nil, fmt.Errorf(":libraries entry cannot be empty")
		}
		out = append(out, name)
	}
	return out, nil
}

// Fetch clones or updates spec into the muse cache and returns the checkout path.
func Fetch(spec string) (string, error) {
	url, err := ResolveGitURL(spec)
	if err != nil {
		return "", err
	}
	if !isGitDir(url) && hasLibTree(url, libraryName(spec)) {
		return url, nil
	}
	name := libraryName(spec)
	dest := filepath.Join(CacheDir(), "libs", name)
	if isGitDir(dest) {
		if err := git(dest, "fetch", "--quiet", "--all"); err != nil {
			return "", err
		}
		if err := git(dest, "pull", "--ff-only", "--quiet"); err != nil {
			_ = err
		}
		return dest, nil
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return "", err
	}
	if err := git("", "clone", "--quiet", "--depth", "1", url, dest); err != nil {
		return "", fmt.Errorf("git clone %s: %w", url, err)
	}
	return dest, nil
}

// defaultGitRepos maps short library names to canonical git remotes.
var defaultGitRepos = map[string]string{
	"csv": "https://github.com/davidhanley/CSV.git",
}

// ResolveGitURL maps a :libraries entry to a git remote or local repo.
// Short names use $MUSE_LIB_<NAME>, then a sibling checkout, testdata,
// a known GitHub repo, or https://github.com/davidhanley/flag-<name>.git.
func ResolveGitURL(spec string) (string, error) {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return "", fmt.Errorf("empty library spec")
	}
	if isGitSpec(spec) {
		if abs, err := filepath.Abs(spec); err == nil && isGitDir(abs) {
			return abs, nil
		}
		return spec, nil
	}
	envKey := "MUSE_LIB_" + strings.ReplaceAll(strings.ToUpper(spec), "-", "_")
	if env := strings.TrimSpace(os.Getenv(envKey)); env != "" {
		return env, nil
	}
	if root, err := findFlagRoot(); err == nil {
		for _, sibling := range []string{
			filepath.Join(filepath.Dir(root), spec),
			filepath.Join(filepath.Dir(root), "flag-"+spec),
		} {
			if isGitDir(sibling) || hasLibTree(sibling, spec) {
				return sibling, nil
			}
		}
		bundled := filepath.Join(root, "testdata", "flag-"+spec)
		if isGitDir(bundled) || hasLibTree(bundled, spec) {
			return bundled, nil
		}
	}
	if url, ok := defaultGitRepos[strings.ToLower(spec)]; ok {
		return url, nil
	}
	return "https://github.com/davidhanley/flag-" + spec + ".git", nil
}

// CacheDir is $MUSE_CACHE, else $XDG_CACHE_HOME/flag-muse, else ~/.cache/flag-muse.
func CacheDir() string {
	if d := strings.TrimSpace(os.Getenv("MUSE_CACHE")); d != "" {
		return d
	}
	if xdg := strings.TrimSpace(os.Getenv("XDG_CACHE_HOME")); xdg != "" {
		return filepath.Join(xdg, "flag-muse")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(os.TempDir(), "flag-muse")
	}
	return filepath.Join(home, ".cache", "flag-muse")
}

func materialize(projectDir, checkout, name string) error {
	src, err := findLibFile(checkout, name)
	if err != nil {
		return err
	}
	destDir := filepath.Join(projectDir, "libraries")
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return err
	}
	dest := filepath.Join(destDir, filepath.Base(src))
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dest, data, 0o644)
}

func findLibFile(checkout, name string) (string, error) {
	candidates := []string{
		filepath.Join(checkout, "libraries", name+".lib"),
		filepath.Join(checkout, name+".lib"),
		filepath.Join(checkout, "libraries", name+".flaglib"),
		filepath.Join(checkout, name+".flag"),
	}
	var tried []string
	for _, c := range candidates {
		tried = append(tried, c)
		if st, err := os.Stat(c); err == nil && !st.IsDir() {
			return c, nil
		}
	}
	for _, pattern := range []string{
		filepath.Join(checkout, "libraries", "*.lib"),
		filepath.Join(checkout, "*.lib"),
	} {
		matches, _ := filepath.Glob(pattern)
		tried = append(tried, pattern)
		if len(matches) == 1 {
			return matches[0], nil
		}
	}
	return "", fmt.Errorf("no %s.lib in checkout (tried %s)", name, strings.Join(tried, ", "))
}

func libraryName(spec string) string {
	s := strings.TrimSpace(spec)
	s = strings.TrimSuffix(s, ".git")
	s = strings.ReplaceAll(s, "\\", "/")
	if i := strings.LastIndex(s, "/"); i >= 0 {
		s = s[i+1:]
	}
	s = strings.TrimPrefix(s, "flag-")
	if s == "" {
		return "lib"
	}
	return s
}

func isGitSpec(spec string) bool {
	switch {
	case strings.Contains(spec, "://"),
		strings.HasPrefix(spec, "git@"),
		strings.HasSuffix(spec, ".git"):
		return true
	}
	if filepath.IsAbs(spec) || strings.HasPrefix(spec, ".") || strings.ContainsAny(spec, "/\\") {
		return true
	}
	return false
}

func isGitDir(path string) bool {
	st, err := os.Stat(filepath.Join(path, ".git"))
	return err == nil && (st.IsDir() || st.Mode().IsRegular())
}

func hasLibTree(dir, name string) bool {
	_, err := findLibFile(dir, name)
	return err == nil
}

func git(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git %s: %w\n%s", strings.Join(args, " "), err, strings.TrimSpace(string(out)))
	}
	return nil
}

func findFlagRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	dir := cwd
	for {
		gomod := filepath.Join(dir, "go.mod")
		data, err := os.ReadFile(gomod)
		if err == nil {
			first := strings.TrimSpace(strings.SplitN(string(data), "\n", 2)[0])
			if first == "module flag-lang" {
				return dir, nil
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("flag-lang root not found")
		}
		dir = parent
	}
}
