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
	_, err := Compile("(def answer 42)")
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
