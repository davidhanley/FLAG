package compiler

import (
	"strings"
	"testing"
)

func TestCompilePrintProgram(t *testing.T) {
	output, err := Compile(`
(ns hello.core)
(println "Hello")
(print 42)
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}

	got := string(output)
	for _, want := range []string{
		"package main",
		`flagrt "flag-lang/runtime"`,
		`fmt.Println("Hello")`,
		"fmt.Print(flagrt.ValueToAny(flagrt.NewLong(42)))",
		"// Source namespace: hello.core",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go did not contain %q:\n%s", want, got)
		}
	}
}

func TestCompilePrintProgramWithMixedWhitespace(t *testing.T) {
	output, err := Compile(`
		(ns   hello.core)
		(println
			"Hello")
		(print   42)
	`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}

	got := string(output)
	for _, want := range []string{
		`fmt.Println("Hello")`,
		"fmt.Print(flagrt.ValueToAny(flagrt.NewLong(42)))",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go did not contain %q:\n%s", want, got)
		}
	}
}

func TestCompileRejectsUnsupportedForms(t *testing.T) {
	_, err := Compile("(let [x 42] x)")
	if err == nil {
		t.Fatal("Compile succeeded for unsupported form")
	}

	if !strings.Contains(err.Error(), "unsupported form") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCompileRejectsMultipleArguments(t *testing.T) {
	_, err := Compile(`(println "hello" "world")`)
	if err == nil {
		t.Fatal("Compile succeeded for unsupported argument list")
	}

	if !strings.Contains(err.Error(), "expected one argument") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCompilePrintlnWithAdditionExpression(t *testing.T) {
	output, err := Compile(`(println (+ 1 2 2.0))`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}

	got := string(output)
	if !strings.Contains(got, "fmt.Println(flagrt.ValueToAny(flagrt.Add(flagrt.Add(flagrt.NewLong(1), flagrt.NewLong(2)), flagrt.NewDouble(2.0))))") {
		t.Fatalf("generated Go did not contain expected addition expression:\n%s", got)
	}
}

func TestCompilePrintlnWithSubAndDivExpressions(t *testing.T) {
	output, err := Compile(`
(println (- 10 3 2))
(println (/ 3 2))
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}

	got := string(output)
	for _, want := range []string{
		"flagrt.ValueToAny(flagrt.Sub(flagrt.Sub(flagrt.NewLong(10), flagrt.NewLong(3)), flagrt.NewLong(2)))",
		"flagrt.ValueToAny(flagrt.Div(flagrt.NewLong(3), flagrt.NewLong(2)))",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go did not contain %q:\n%s", want, got)
		}
	}
}

func TestCompilePrintlnWithEqualsExpression(t *testing.T) {
	output, err := Compile(`(println (= 1 1.0 1))`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}

	got := string(output)
	want := "fmt.Println(flagrt.Eq(flagrt.NewLong(1), flagrt.NewDouble(1.0)) && flagrt.Eq(flagrt.NewDouble(1.0), flagrt.NewLong(1)))"
	if !strings.Contains(got, want) {
		t.Fatalf("generated Go did not contain expected equals expression:\n%s", got)
	}
}

func TestCompileDefnAndCall(t *testing.T) {
	output, err := Compile(`
(defn sq [x] (* x x))
(println (sq 47))
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}

	got := string(output)
	for _, want := range []string{
		"func sq(args ...flagrt.Value) flagrt.Value {",
		"if len(args) != 1 {",
		`panic("sq expects exactly 1 arguments")`,
		"x := args[0]",
		"return flagrt.Mul(x, x)",
		"fmt.Println(flagrt.ValueToAny(sq(flagrt.NewLong(47))))",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go did not contain %q:\n%s", want, got)
		}
	}
}

func TestCompileDefAndUse(t *testing.T) {
	output, err := Compile(`
(def x 1)
(println (+ x 2))
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}

	got := string(output)
	for _, want := range []string{
		"var x = flagrt.NewLong(1)",
		"fmt.Println(flagrt.ValueToAny(flagrt.Add(x, flagrt.NewLong(2))))",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go did not contain %q:\n%s", want, got)
		}
	}
}

func TestCompileExpression(t *testing.T) {
	got, err := CompileExpression(`(+ 1 2 2.0)`)
	if err != nil {
		t.Fatalf("CompileExpression returned error: %v", err)
	}

	want := "flagrt.ValueToAny(flagrt.Add(flagrt.Add(flagrt.NewLong(1), flagrt.NewLong(2)), flagrt.NewDouble(2.0)))"
	if got != want {
		t.Fatalf("unexpected expression:\nwant: %s\ngot:  %s", want, got)
	}
}

func TestReplCompilerDefBinding(t *testing.T) {
	c := NewReplCompiler()

	first, err := c.CompileLine(`(def x 1)`)
	if err != nil {
		t.Fatalf("CompileLine def returned error: %v", err)
	}
	if first.Setup != "var x flagrt.Value = flagrt.NewLong(1)" {
		t.Fatalf("unexpected setup for first def: %s", first.Setup)
	}
	if first.ResultExpr != "flagrt.ValueToAny(x)" {
		t.Fatalf("unexpected result expr for first def: %s", first.ResultExpr)
	}

	second, err := c.CompileLine(`(def x 2)`)
	if err != nil {
		t.Fatalf("CompileLine redef returned error: %v", err)
	}
	if second.Setup != "x = flagrt.NewLong(2)" {
		t.Fatalf("unexpected setup for redef: %s", second.Setup)
	}

	expr, err := c.CompileLine(`(+ x 3)`)
	if err != nil {
		t.Fatalf("CompileLine expr returned error: %v", err)
	}
	wantExpr := "flagrt.ValueToAny(flagrt.Add(x, flagrt.NewLong(3)))"
	if expr.ResultExpr != wantExpr {
		t.Fatalf("unexpected result expr:\nwant: %s\ngot:  %s", wantExpr, expr.ResultExpr)
	}
}

func TestReplCompilerDefnBinding(t *testing.T) {
	c := NewReplCompiler()

	defn, err := c.CompileLine(`(defn sq [x] (* x x))`)
	if err != nil {
		t.Fatalf("CompileLine defn returned error: %v", err)
	}
	for _, want := range []string{
		"var sq = func(args ...flagrt.Value) flagrt.Value {",
		"if len(args) != 1 {",
		`panic("sq expects exactly 1 arguments")`,
		"x := args[0]",
		"return flagrt.Mul(x, x)",
	} {
		if !strings.Contains(defn.Setup, want) {
			t.Fatalf("defn setup did not contain %q:\n%s", want, defn.Setup)
		}
	}
	if defn.ResultExpr != `"sq"` {
		t.Fatalf("unexpected defn result expr: %s", defn.ResultExpr)
	}

	call, err := c.CompileLine(`(sq 4)`)
	if err != nil {
		t.Fatalf("CompileLine function call returned error: %v", err)
	}
	if call.ResultExpr != "flagrt.ValueToAny(sq(flagrt.NewLong(4)))" {
		t.Fatalf("unexpected function call expression: %s", call.ResultExpr)
	}
}
