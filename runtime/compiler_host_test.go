package runtime

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFormatGoSourceFormatsPackage(t *testing.T) {
	src := "package p\nfunc F( x int ){ _ = x }\n"
	got, err := FormatGoSource(src)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "func F(x int)") {
		t.Fatalf("formatted:\n%s", got)
	}
	if strings.Contains(got, "F( x int )") {
		t.Fatalf("still unformatted:\n%s", got)
	}
}

func TestFormatGoSourceRejectsInvalidGo(t *testing.T) {
	_, err := FormatGoSource("this is not go")
	if err == nil || !strings.Contains(err.Error(), "format generated Go") {
		t.Fatalf("got %v", err)
	}
}

func TestAbsPathAndCwd(t *testing.T) {
	cwd, err := CwdPath()
	if err != nil {
		t.Fatal(err)
	}
	abs, err := AbsPath(".")
	if err != nil {
		t.Fatal(err)
	}
	if abs != cwd {
		t.Fatalf("AbsPath(.) = %q cwd = %q", abs, cwd)
	}
	if _, err := AbsPath("  "); err == nil {
		t.Fatal("expected empty path error")
	}
}

func TestPathJoinDirExtAbs(t *testing.T) {
	joined := PathJoin("a", "b.flag")
	if joined != filepath.Join("a", "b.flag") {
		t.Fatalf("join %q", joined)
	}
	if PathDir(joined) != "a" {
		t.Fatalf("dir %q", PathDir(joined))
	}
	if PathExt(joined) != ".flag" {
		t.Fatalf("ext %q", PathExt(joined))
	}
	if PathExt("noext") != "" {
		t.Fatalf("ext empty %q", PathExt("noext"))
	}
	if PathIsAbs(joined) {
		t.Fatal("relative should not be abs")
	}
	if abs := filepath.Join("/", "tmp"); PathIsAbs(abs) != filepath.IsAbs(abs) {
		t.Fatalf("abs mismatch for %q", abs)
	}
}

func TestPathStatKind(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "x.flag")
	if err := os.WriteFile(file, []byte("{:namespace \"x\"}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := PathStatKind(file); got != "file" {
		t.Fatalf("file kind %q", got)
	}
	if got := PathStatKind(dir); got != "dir" {
		t.Fatalf("dir kind %q", got)
	}
	if got := PathStatKind(filepath.Join(dir, "missing")); got != "missing" {
		t.Fatalf("missing kind %q", got)
	}
}

func TestWriteReadTextFileAndMkdir(t *testing.T) {
	dir := t.TempDir()
	nested := filepath.Join(dir, "a", "b")
	if err := MkdirAll(nested); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(nested, "x.txt")
	if err := WriteTextFile(path, "hello"); err != nil {
		t.Fatal(err)
	}
	got, err := ReadTextFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got != "hello" {
		t.Fatalf("got %q", got)
	}
	if err := WriteTextFile("", "x"); err == nil {
		t.Fatal("expected empty path error")
	}
	if _, err := ReadTextFile(filepath.Join(dir, "missing")); err == nil {
		t.Fatal("expected missing file error")
	}
}

func TestGoModTidyRejectsEmpty(t *testing.T) {
	if err := GoModTidy(""); err == nil || !strings.Contains(err.Error(), "empty path") {
		t.Fatalf("got %v", err)
	}
}

func TestCompilerHostGoFnDispatch(t *testing.T) {
	formatted := Call(GoFunction("compiler/format-go-source"), NewString("package p\nfunc F(){}\n"))
	if formatted.tag != TagString || !strings.Contains(formatted.StringValue(), "func F()") {
		t.Fatalf("go-fn format %#v", formatted)
	}
	kind := Call(GoFunction("compiler/path-stat-kind"), NewString("."))
	if kind.StringValue() != "dir" {
		t.Fatalf("stat . => %q", kind.StringValue())
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "w.txt")
	Call(GoFunction("compiler/write-text-file"), NewString(path), NewString("ok"))
	got := Call(GoFunction("compiler/read-text-file"), NewString(path))
	if got.StringValue() != "ok" {
		t.Fatalf("read-text-file %q", got.StringValue())
	}
}
