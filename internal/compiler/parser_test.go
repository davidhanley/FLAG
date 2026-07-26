package compiler

import (
	"strings"
	"testing"
)

func TestParseFileMultipleFormsWhitespaceInsensitive(t *testing.T) {
	source := `
		; file comment
		(ns   hello.core)

		(println
			"Hello")
		(print,42)
		{:x 1 :y [2 3]}
	`

	ast, err := ParseFile(source)
	if err != nil {
		t.Fatalf("ParseFile returned error: %v", err)
	}

	if len(ast.Forms) != 4 {
		t.Fatalf("expected 4 top-level forms, got %d", len(ast.Forms))
	}

	nsForm, ok := ast.Forms[0].(ListExpr)
	if !ok {
		t.Fatalf("expected first form to be ListExpr, got %T", ast.Forms[0])
	}
	if len(nsForm.Elements) != 2 {
		t.Fatalf("expected ns form with 2 elements, got %d", len(nsForm.Elements))
	}

	sym, ok := nsForm.Elements[1].(SymbolExpr)
	if !ok || sym.Name != "hello.core" {
		t.Fatalf("expected namespace symbol hello.core, got %#v", nsForm.Elements[1])
	}

	mapForm, ok := ast.Forms[3].(MapExpr)
	if !ok {
		t.Fatalf("expected fourth form to be MapExpr, got %T", ast.Forms[3])
	}
	if len(mapForm.Entries) != 4 {
		t.Fatalf("expected map with 4 entry forms, got %d", len(mapForm.Entries))
	}
}

func TestParseFileUnterminatedList(t *testing.T) {
	_, err := ParseFile(`(println "x"`)
	if err == nil {
		t.Fatal("ParseFile succeeded with unterminated list")
	}

	if !strings.Contains(err.Error(), "missing closing") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestParseFileUnterminatedString(t *testing.T) {
	_, err := ParseFile(`(println "x)`)
	if err == nil {
		t.Fatal("ParseFile succeeded with unterminated string")
	}

	if !strings.Contains(err.Error(), "unterminated string literal") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestParseFileQuotedAndKeywordSymbols(t *testing.T) {
	ast, err := ParseFile(`(println 'abc :xyz)`)
	if err != nil {
		t.Fatalf("ParseFile returned error: %v", err)
	}

	call, ok := ast.Forms[0].(ListExpr)
	if !ok {
		t.Fatalf("expected first form to be ListExpr, got %T", ast.Forms[0])
	}
	if len(call.Elements) != 3 {
		t.Fatalf("expected println form with 3 elements, got %d", len(call.Elements))
	}

	quoted, ok := call.Elements[1].(QuotedSymbolExpr)
	if !ok || quoted.Name != "abc" {
		t.Fatalf("expected quoted symbol abc, got %#v", call.Elements[1])
	}
	keyword, ok := call.Elements[2].(KeywordExpr)
	if !ok || keyword.Name != "xyz" {
		t.Fatalf("expected keyword xyz, got %#v", call.Elements[2])
	}
}

func TestParseFileCommentFormTopLevel(t *testing.T) {
	ast, err := ParseFile(`(println 1)
(comment (println 2) (+ 1 2))
(println 3)`)
	if err != nil {
		t.Fatalf("ParseFile returned error: %v", err)
	}

	if len(ast.Forms) != 2 {
		t.Fatalf("expected 2 top-level forms, got %d", len(ast.Forms))
	}
}

func TestParseFileCommentFormNested(t *testing.T) {
	ast, err := ParseFile(`(do 1 (comment (println 2) {:a 1}) 3)`)
	if err != nil {
		t.Fatalf("ParseFile returned error: %v", err)
	}

	form, ok := ast.Forms[0].(ListExpr)
	if !ok {
		t.Fatalf("expected ListExpr, got %T", ast.Forms[0])
	}
	if len(form.Elements) != 3 {
		t.Fatalf("expected comment form to be removed, got %d elements", len(form.Elements))
	}
	if value, ok := form.Elements[2].(IntExpr); !ok || value.Value != 3 {
		t.Fatalf("expected trailing form to remain, got %#v", form.Elements[2])
	}
}

func TestParseFileQuotedList(t *testing.T) {
	ast, err := ParseFile(`(println '(1 2 3))`)
	if err != nil {
		t.Fatalf("ParseFile returned error: %v", err)
	}

	call, ok := ast.Forms[0].(ListExpr)
	if !ok {
		t.Fatalf("expected first form to be ListExpr, got %T", ast.Forms[0])
	}
	if len(call.Elements) != 2 {
		t.Fatalf("expected println form with 2 elements, got %d", len(call.Elements))
	}

	quoted, ok := call.Elements[1].(QuotedListExpr)
	if !ok {
		t.Fatalf("expected quoted list, got %#v", call.Elements[1])
	}
	if len(quoted.Elements) != 3 {
		t.Fatalf("expected quoted list length 3, got %d", len(quoted.Elements))
	}
}

func TestParseFileSetLiteral(t *testing.T) {
	ast, err := ParseFile(`(println #{1 2 3})`)
	if err != nil {
		t.Fatalf("ParseFile returned error: %v", err)
	}

	call, ok := ast.Forms[0].(ListExpr)
	if !ok {
		t.Fatalf("expected first form to be ListExpr, got %T", ast.Forms[0])
	}

	setExpr, ok := call.Elements[1].(SetExpr)
	if !ok {
		t.Fatalf("expected set literal, got %#v", call.Elements[1])
	}
	if len(setExpr.Elements) != 3 {
		t.Fatalf("expected set with 3 elements, got %d", len(setExpr.Elements))
	}
}

func TestParseFileHashFnLiteral(t *testing.T) {
	ast, err := ParseFile(`(println #(* % 3))`)
	if err != nil {
		t.Fatalf("ParseFile returned error: %v", err)
	}

	call, ok := ast.Forms[0].(ListExpr)
	if !ok {
		t.Fatalf("expected first form to be ListExpr, got %T", ast.Forms[0])
	}

	hashFn, ok := call.Elements[1].(HashFnExpr)
	if !ok {
		t.Fatalf("expected hash-fn literal, got %#v", call.Elements[1])
	}

	body, ok := hashFn.Body.(ListExpr)
	if !ok {
		t.Fatalf("expected hash-fn body to be list, got %T", hashFn.Body)
	}
	if len(body.Elements) != 3 {
		t.Fatalf("expected hash-fn body with 3 elements, got %d", len(body.Elements))
	}
}
