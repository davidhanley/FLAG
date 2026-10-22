package compiler

import (
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

func TestRewriteGoPredeclaredShadowsRenamesClose(t *testing.T) {
	src := `package compiler
func close_arity_1(x int) int { return x }
var close = close_arity_1
func use() { close(1); _ = other.close }
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "t.go", src, 0)
	if err != nil {
		t.Fatal(err)
	}
	RewriteGoPredeclaredShadows(file)

	var buf strings.Builder
	if err := format.Node(&buf, fset, file); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if strings.Contains(got, "var close ") {
		t.Fatalf("close still a package name:\n%s", got)
	}
	if !strings.Contains(got, "prologue_close") {
		t.Fatalf("missing prologue_close:\n%s", got)
	}
	if !strings.Contains(got, "other.close") {
		t.Fatalf("rewrote selector:\n%s", got)
	}
}

func TestRewriteGoPredeclaredShadowsNoOpWithoutCollision(t *testing.T) {
	src := `package compiler
func inc(x int) int { return x + len(nil) }
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "t.go", src, 0)
	if err != nil {
		t.Fatal(err)
	}
	RewriteGoPredeclaredShadows(file)
	ast.Inspect(file, func(n ast.Node) bool {
		id, ok := n.(*ast.Ident)
		if ok && (id.Name == "prologue_len" || id.Name == "prologue_inc") {
			t.Fatalf("unexpected rename %s", id.Name)
		}
		return true
	})
}
