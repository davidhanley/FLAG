package muse

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseLibraries(t *testing.T) {
	man, err := Parse(`{:libraries ["csv" "http"]}`)
	if err != nil {
		t.Fatal(err)
	}
	if len(man.Libraries) != 2 || man.Libraries[0] != "csv" || man.Libraries[1] != "http" {
		t.Fatalf("libraries: %#v", man.Libraries)
	}
}

func TestParseRejectsUnknownKey(t *testing.T) {
	_, err := Parse(`{:libraries ["csv"] :wat 1}`)
	if err == nil || !strings.Contains(err.Error(), "unknown muse key") {
		t.Fatalf("expected unknown key, got %v", err)
	}
}

func TestParseRejectsNonMap(t *testing.T) {
	_, err := Parse(`["csv"]`)
	if err == nil || !strings.Contains(err.Error(), "FLAG map") {
		t.Fatalf("expected map error, got %v", err)
	}
}

func TestDefaultCSVRepo(t *testing.T) {
	if got := defaultGitRepos["csv"]; got != "https://github.com/davidhanley/CSV.git" {
		t.Fatalf("csv remote: %q", got)
	}
}

func TestLibraryName(t *testing.T) {
	cases := map[string]string{
		"csv": "csv",
		"https://github.com/davidhanley/flag-csv.git": "csv",
		"https://github.com/davidhanley/CSV.git":      "CSV",
		"/tmp/flag-http":                              "http",
	}
	for in, want := range cases {
		if got := libraryName(in); got != want {
			t.Errorf("libraryName(%q)=%q want %q", in, got, want)
		}
	}
}

func TestPrepareFetchesGitLibrary(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	cache := t.TempDir()
	t.Setenv("MUSE_CACHE", cache)

	libRepo := t.TempDir()
	libFile := filepath.Join(libRepo, "libraries")
	if err := os.MkdirAll(libFile, 0o755); err != nil {
		t.Fatal(err)
	}
	src := `{:namespace "demo"
 :exports   [ping]}
(defn ping [] :ok)
`
	if err := os.WriteFile(filepath.Join(libFile, "demo.lib"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	initGitRepo(t, libRepo)

	project := t.TempDir()
	musePath := filepath.Join(project, "project.muse")
	body := `{:libraries ["` + libRepo + `"]}
`
	if err := os.WriteFile(musePath, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	dir, err := Prepare(musePath)
	if err != nil {
		t.Fatalf("Prepare: %v", err)
	}
	if dir != project {
		t.Fatalf("project dir: got %q want %q", dir, project)
	}
	got, err := os.ReadFile(filepath.Join(project, "libraries", "demo.lib"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), `:namespace "demo"`) {
		t.Fatalf("materialized lib:\n%s", got)
	}
}

func TestFindMuseFileAndDir(t *testing.T) {
	dir := t.TempDir()
	musePath := filepath.Join(dir, "app.muse")
	if err := os.WriteFile(musePath, []byte(`{:libraries []}`), 0o644); err != nil {
		t.Fatal(err)
	}
	flagPath := filepath.Join(dir, "main.flag")
	if err := os.WriteFile(flagPath, []byte(`{:namespace "main"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	if p, ok := Find(musePath); !ok || p != musePath {
		t.Fatalf("Find muse file: %q %v", p, ok)
	}
	if p, ok := Find(dir); !ok || p != musePath {
		t.Fatalf("Find dir: %q %v", p, ok)
	}
	if p, ok := Find(flagPath); !ok || p != musePath {
		t.Fatalf("Find flag file: %q %v", p, ok)
	}
}

func initGitRepo(t *testing.T, dir string) {
	t.Helper()
	cmds := [][]string{
		{"git", "init", "-q"},
		{"git", "config", "user.email", "muse@example.com"},
		{"git", "config", "user.name", "muse"},
		{"git", "add", "."},
		{"git", "commit", "-qm", "lib"},
	}
	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%s: %v\n%s", strings.Join(args, " "), err, out)
		}
	}
}
