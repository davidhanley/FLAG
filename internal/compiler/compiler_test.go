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
		`fmt.Println(flagrt.Str("Hello"))`,
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
		`fmt.Println(flagrt.Str("Hello"))`,
		"fmt.Print(flagrt.ValueToAny(flagrt.NewLong(42)))",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go did not contain %q:\n%s", want, got)
		}
	}
}

func TestCompileRejectsUnsupportedForms(t *testing.T) {
	_, err := Compile("(loop [x 42] x)")
	if err == nil {
		t.Fatal("Compile succeeded for unsupported form")
	}
}

func TestCompileRejectsMultipleArgumentsForPrint(t *testing.T) {
	_, err := Compile(`(print "hello" "world")`)
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
	if !strings.Contains(got, "fmt.Println(flagrt.Str(flagrt.Add(flagrt.Add(flagrt.NewLong(1), flagrt.NewLong(2)), flagrt.NewDouble(2.0))))") {
		t.Fatalf("generated Go did not contain expected addition expression:\n%s", got)
	}
}

func TestCompilePrintlnWithSubAndDivExpressions(t *testing.T) {
	output, err := Compile(`
(println (- 10 3 2))
(println (/ 3 2))
(println (% -5 3))
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}

	got := string(output)
	for _, want := range []string{
		"fmt.Println(flagrt.Str(flagrt.Sub(flagrt.Sub(flagrt.NewLong(10), flagrt.NewLong(3)), flagrt.NewLong(2))))",
		"fmt.Println(flagrt.Str(flagrt.Div(flagrt.NewLong(3), flagrt.NewLong(2))))",
		"fmt.Println(flagrt.Str(flagrt.Mod(flagrt.NewLong(-5), flagrt.NewLong(3))))",
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
	want := "fmt.Println(flagrt.Str(flagrt.NewBool(flagrt.Eq(flagrt.NewLong(1), flagrt.NewDouble(1.0)) && flagrt.Eq(flagrt.NewDouble(1.0), flagrt.NewLong(1)))))"
	if !strings.Contains(got, want) {
		t.Fatalf("generated Go did not contain expected equals expression:\n%s", got)
	}
}

func TestCompilePrintlnWithComparisonExpressions(t *testing.T) {
	output, err := Compile(`
(println (< 1 2 3.0))
(println (> 3 2 1))
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}

	got := string(output)
	for _, want := range []string{
		"fmt.Println(flagrt.Str(flagrt.Lt(flagrt.NewLong(1), flagrt.NewLong(2)) && flagrt.Lt(flagrt.NewLong(2), flagrt.NewDouble(3.0))))",
		"fmt.Println(flagrt.Str(flagrt.Gt(flagrt.NewLong(3), flagrt.NewLong(2)) && flagrt.Gt(flagrt.NewLong(2), flagrt.NewLong(1))))",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go did not contain expected comparison expression %q:\n%s", want, got)
		}
	}
}

func TestCompilePrintlnWithIfExpression(t *testing.T) {
	output, err := Compile(`
(println (if (= 1 1.0) 42 0))
(println (if (= 1 2) 42))
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}

	got := string(output)
	for _, want := range []string{
		"if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagrt.NewLong(1), flagrt.NewDouble(1.0)))) {",
		"return flagrt.NewLong(42)",
		"return flagrt.NewLong(0)",
		"if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagrt.NewLong(1), flagrt.NewLong(2)))) {",
		"return flagrt.NilValue()",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go did not contain %q:\n%s", want, got)
		}
	}
}

func TestCompilePrintlnWithLetExpression(t *testing.T) {
	output, err := Compile(`
(println (let [a 1
               b (+ 1 a)]))
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}

	got := string(output)
	for _, want := range []string{
		"func() flagrt.Value {",
		"a := flagrt.NewLong(1)",
		"b := flagrt.Add(flagrt.NewLong(1), a)",
		"return b",
		"fmt.Println(flagrt.Str(func() flagrt.Value {",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go did not contain %q:\n%s", want, got)
		}
	}
}

func TestCompilePrintlnWithSymbolLiterals(t *testing.T) {
	output, err := Compile(`(println 'abc :xyz)`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}

	got := string(output)
	want := `fmt.Println(flagrt.Str(flagrt.NewSymbol("abc"), flagrt.NewKeyword("xyz")))`
	if !strings.Contains(got, want) {
		t.Fatalf("generated Go did not contain expected symbol literals:\n%s", got)
	}
}

func TestCompilePrintlnWithMapAndSetLiterals(t *testing.T) {
	output, err := Compile(`
(println {:a 1 :b 2})
(println #{1 2 3})
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}

	got := string(output)
	for _, want := range []string{
		`fmt.Println(flagrt.Str(flagrt.NewMap(flagrt.NewKeyword("a"), flagrt.NewLong(1), flagrt.NewKeyword("b"), flagrt.NewLong(2))))`,
		`fmt.Println(flagrt.Str(flagrt.NewSet(flagrt.NewLong(1), flagrt.NewLong(2), flagrt.NewLong(3))))`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go did not contain %q:\n%s", want, got)
		}
	}
}

func TestCompileSymbolAndNameFunctions(t *testing.T) {
	output, err := Compile(`
(println (name :xyz))
(println (name (symbol "abc")))
(println (symbol :xyz))
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}

	got := string(output)
	for _, want := range []string{
		`fmt.Println(flagrt.Str(flagrt.Name(flagrt.NewKeyword("xyz"))))`,
		`fmt.Println(flagrt.Str(flagrt.Name(flagrt.Symbol("abc"))))`,
		`fmt.Println(flagrt.Str(flagrt.Symbol(flagrt.NewKeyword("xyz"))))`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go did not contain %q:\n%s", want, got)
		}
	}
}

func TestCompileAssocAndDissocFunctions(t *testing.T) {
	output, err := Compile(`
(println (assoc {:a 1} :b 2))
(println (dissoc {:a 1 :b 2} :b))
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}

	got := string(output)
	for _, want := range []string{
		`fmt.Println(flagrt.Str(flagrt.MapAssoc(flagrt.NewMap(flagrt.NewKeyword("a"), flagrt.NewLong(1)), flagrt.NewKeyword("b"), flagrt.NewLong(2))))`,
		`fmt.Println(flagrt.Str(flagrt.MapDissoc(flagrt.NewMap(flagrt.NewKeyword("a"), flagrt.NewLong(1), flagrt.NewKeyword("b"), flagrt.NewLong(2)), flagrt.NewKeyword("b"))))`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go did not contain %q:\n%s", want, got)
		}
	}
}

func TestCompileFirstAndRestFunctions(t *testing.T) {
	output, err := Compile(`
(println (first [1 2 3]))
(println (fist [1 2 3]))
(println (rest [1 2 3]))
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}

	got := string(output)
	for _, want := range []string{
		`fmt.Println(flagrt.Str(flagrt.First(flagrt.NewArray(flagrt.NewLong(1), flagrt.NewLong(2), flagrt.NewLong(3)))))`,
		`fmt.Println(flagrt.Str(flagrt.Rest(flagrt.NewArray(flagrt.NewLong(1), flagrt.NewLong(2), flagrt.NewLong(3)))))`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go did not contain %q:\n%s", want, got)
		}
	}
}

func TestCompileMapFunction(t *testing.T) {
	output, err := Compile(`
(println (map (fn [x] (* x x)) [1 2 3]))
(defn add2 [a b] (+ a b))
(println (map add2 [1 2 3] [10 20 30]))
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}

	got := string(output)
	for _, want := range []string{
		`fmt.Println(flagrt.Str(flagrt.Map(flagrt.NewFunction(func(args ...flagrt.Value) flagrt.Value {`,
		`fmt.Println(flagrt.Str(flagrt.Map(flagrt.NewFunction(add2), flagrt.NewArray(flagrt.NewLong(1), flagrt.NewLong(2), flagrt.NewLong(3)), flagrt.NewArray(flagrt.NewLong(10), flagrt.NewLong(20), flagrt.NewLong(30)))))`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go did not contain %q:\n%s", want, got)
		}
	}
}

func TestCompileFilterFunction(t *testing.T) {
	output, err := Compile(`
(println (filter (fn [x] (if (> x 1) 1)) [1 2 3]))
(defn gt2 [x] (if (> x 2) 1))
(println (filter gt2 [1 2 3 4]))
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}

	got := string(output)
	for _, want := range []string{
		`fmt.Println(flagrt.Str(flagrt.Filter(flagrt.NewFunction(func(args ...flagrt.Value) flagrt.Value {`,
		`fmt.Println(flagrt.Str(flagrt.Filter(flagrt.NewFunction(gt2), flagrt.NewArray(flagrt.NewLong(1), flagrt.NewLong(2), flagrt.NewLong(3), flagrt.NewLong(4)))))`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go did not contain %q:\n%s", want, got)
		}
	}
}

func TestCompileWhenMacro(t *testing.T) {
	output, err := Compile(`
(println (when true 42))
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}
	got := string(output)
	for _, want := range []string{
		"if flagrt.IsTruthy(flagrt.NewBool(true)) {",
		"return flagrt.NewLong(42)",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go did not contain %q:\n%s", want, got)
		}
	}
}

func TestCompileNotEqualsMacro(t *testing.T) {
	output, err := Compile(`(println (not= 1 2))`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}
	got := string(output)
	for _, want := range []string{
		"if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagrt.NewLong(1), flagrt.NewLong(2)))) {",
		"return flagrt.NewBool(false)",
		"return flagrt.NewBool(true)",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go did not contain %q:\n%s", want, got)
		}
	}
}

func TestCompileCondMacro(t *testing.T) {
	output, err := Compile(`(println (cond false 1 (= 1 1) 2 :else 3))`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}
	got := string(output)
	for _, want := range []string{
		"if flagrt.IsTruthy(flagrt.NewBool(false)) {",
		"if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagrt.NewLong(1), flagrt.NewLong(1)))) {",
		"return flagrt.NewLong(2)",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go did not contain %q:\n%s", want, got)
		}
	}
}

func TestCompileThreadFirstMacro(t *testing.T) {
	output, err := Compile(`(println (-> 5 (+ 1) (* 2)))`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}
	got := string(output)
	want := "fmt.Println(flagrt.Str(flagrt.Mul(flagrt.Add(flagrt.NewLong(5), flagrt.NewLong(1)), flagrt.NewLong(2))))"
	if !strings.Contains(got, want) {
		t.Fatalf("generated Go did not contain %q:\n%s", want, got)
	}
}

func TestCompileThreadLastMacro(t *testing.T) {
	output, err := Compile(`(println (->> 5 (- 10) (/ 3)))`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}
	got := string(output)
	want := "fmt.Println(flagrt.Str(flagrt.Div(flagrt.NewLong(3), flagrt.Sub(flagrt.NewLong(10), flagrt.NewLong(5)))))"
	if !strings.Contains(got, want) {
		t.Fatalf("generated Go did not contain %q:\n%s", want, got)
	}
}

func TestCompileSomeThreadFirstMacro(t *testing.T) {
	output, err := Compile(`(println (some-> 5 (+ 1) (* 2)))`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}
	got := string(output)
	for _, want := range []string{
		"__some_arrow_0 :=",
		"__some_arrow_1 :=",
		"flagrt.Eq(__some_arrow_0, flagrt.NilValue())",
		"flagrt.Mul(__some_arrow_1, flagrt.NewLong(2))",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go did not contain %q:\n%s", want, got)
		}
	}
}

func TestCompileFnLiteralAndCall(t *testing.T) {
	output, err := Compile(`(println ((fn [x] (* x 3)) 4))`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}

	got := string(output)
	for _, want := range []string{
		"flagrt.NewFunction(func(args ...flagrt.Value) flagrt.Value {",
		`panic("fn expects exactly 1 arguments")`,
		"x := args[0]",
		"return flagrt.Mul(x, flagrt.NewLong(3))",
		"flagrt.Call(flagrt.NewFunction(func(args ...flagrt.Value) flagrt.Value {",
		"flagrt.NewLong(4)",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go did not contain %q:\n%s", want, got)
		}
	}
}

func TestCompileHashFnLiteralAndCall(t *testing.T) {
	output, err := Compile(`(println (#(* % 3) 4))`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}

	got := string(output)
	for _, want := range []string{
		"flagrt.NewFunction(func(args ...flagrt.Value) flagrt.Value {",
		"__p1 := args[0]",
		"return flagrt.Mul(__p1, flagrt.NewLong(3))",
		"flagrt.Call(flagrt.NewFunction(func(args ...flagrt.Value) flagrt.Value {",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go did not contain %q:\n%s", want, got)
		}
	}
}

func TestCompileDefLambdaAndCall(t *testing.T) {
	output, err := Compile(`
(def triple #(* % 3))
(println (triple 5))
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}

	got := string(output)
	for _, want := range []string{
		"var triple = flagrt.NewFunction(func(args ...flagrt.Value) flagrt.Value {",
		"__p1 := args[0]",
		"return flagrt.Mul(__p1, flagrt.NewLong(3))",
		"fmt.Println(flagrt.Str(flagrt.Call(triple, flagrt.NewLong(5))))",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go did not contain %q:\n%s", want, got)
		}
	}
}

func TestCompilePrintlnWithStrExpression(t *testing.T) {
	output, err := Compile(`(println (str 1 2 (/ 3 2)))`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}

	got := string(output)
	want := "fmt.Println(flagrt.Str(flagrt.Str(flagrt.NewLong(1), flagrt.NewLong(2), flagrt.Div(flagrt.NewLong(3), flagrt.NewLong(2)))))"
	if !strings.Contains(got, want) {
		t.Fatalf("generated Go did not contain expected str expression:\n%s", got)
	}
}

func TestCompilePrintlnWithMultipleArguments(t *testing.T) {
	output, err := Compile(`(println "a" 1 true)`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}

	got := string(output)
	want := `fmt.Println(flagrt.Str("a", flagrt.NewLong(1), flagrt.NewBool(true)))`
	if !strings.Contains(got, want) {
		t.Fatalf("generated Go did not contain expected println call:\n%s", got)
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
		"fmt.Println(flagrt.Str(flagrt.Call(flagrt.NewFunction(sq), flagrt.NewLong(47))))",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go did not contain %q:\n%s", want, got)
		}
	}
}

func TestCompileRecursiveDefnFib(t *testing.T) {
	output, err := Compile(`
(defn fib [x] (if (< x 3) 1 (+ (fib (- x 1)) (fib (- x 2)))))
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}

	got := string(output)
	for _, want := range []string{
		"func fib(args ...flagrt.Value) flagrt.Value {",
		"return flagrt.Add(flagrt.Call(flagrt.NewFunction(fib), flagrt.Sub(x, flagrt.NewLong(1))), flagrt.Call(flagrt.NewFunction(fib), flagrt.Sub(x, flagrt.NewLong(2))))",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go did not contain %q:\n%s", want, got)
		}
	}
}

func TestCompilePredicateDefnEven(t *testing.T) {
	output, err := Compile(`
(defn even [x] (= (% x 2) 0))
(println (even 4))
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}

	got := string(output)
	for _, want := range []string{
		"func even(args ...flagrt.Value) flagrt.Value {",
		"return flagrt.NewBool(flagrt.Eq(flagrt.Mod(x, flagrt.NewLong(2)), flagrt.NewLong(0)))",
		"fmt.Println(flagrt.Str(flagrt.Call(flagrt.NewFunction(even), flagrt.NewLong(4))))",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go did not contain %q:\n%s", want, got)
		}
	}
}

func TestCompileDefnCanReturnAnyValueType(t *testing.T) {
	output, err := Compile(`
(defn mklist [x] (rest [0 x]))
(defn mkarray [x] [x 2])
(defn mksym [x] 'foo)
(defn mkmap [x] {:a x})
(defn mkset [x] #{x})
(defn mkfn [x] (fn [y] (+ x y)))
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}

	got := string(output)
	for _, want := range []string{
		`func mklist(args ...flagrt.Value) flagrt.Value {`,
		`return flagrt.Rest(flagrt.NewArray(flagrt.NewLong(0), x))`,
		`func mkarray(args ...flagrt.Value) flagrt.Value {`,
		`return flagrt.NewArray(x, flagrt.NewLong(2))`,
		`func mksym(args ...flagrt.Value) flagrt.Value {`,
		`return flagrt.NewSymbol("foo")`,
		`func mkmap(args ...flagrt.Value) flagrt.Value {`,
		`return flagrt.NewMap(flagrt.NewKeyword("a"), x)`,
		`func mkset(args ...flagrt.Value) flagrt.Value {`,
		`return flagrt.NewSet(x)`,
		`func mkfn(args ...flagrt.Value) flagrt.Value {`,
		`return flagrt.NewFunction(func(args ...flagrt.Value) flagrt.Value {`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go did not contain %q:\n%s", want, got)
		}
	}
}

func TestCompileTopLevelExpressionExecutesInMain(t *testing.T) {
	output, err := Compile(`
(defn fib [x] (if (< x 3) 1 (+ (fib (- x 1)) (fib (- x 2)))))
(fib 7)
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}

	got := string(output)
	want := "_ = flagrt.Call(flagrt.NewFunction(fib), flagrt.NewLong(7))"
	if !strings.Contains(got, want) {
		t.Fatalf("generated Go did not contain expected top-level execution statement %q:\n%s", want, got)
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
		"fmt.Println(flagrt.Str(flagrt.Add(x, flagrt.NewLong(2))))",
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

func TestCompileExpressionStr(t *testing.T) {
	got, err := CompileExpression(`(str 1 2 (/ 3 2))`)
	if err != nil {
		t.Fatalf("CompileExpression returned error: %v", err)
	}

	want := "flagrt.Str(flagrt.NewLong(1), flagrt.NewLong(2), flagrt.Div(flagrt.NewLong(3), flagrt.NewLong(2)))"
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
	if first.Setup != "var x flagrt.Value;;x = flagrt.NewLong(1)" {
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
	if call.ResultExpr != "flagrt.ValueToAny(flagrt.Call(flagrt.NewFunction(sq), flagrt.NewLong(4)))" {
		t.Fatalf("unexpected function call expression: %s", call.ResultExpr)
	}
}

func TestReplCompilerDefLambdaBinding(t *testing.T) {
	c := NewReplCompiler()

	def, err := c.CompileLine(`(def triple #(* % 3))`)
	if err != nil {
		t.Fatalf("CompileLine def lambda returned error: %v", err)
	}
	if !strings.Contains(def.Setup, "var triple flagrt.Value;;triple = flagrt.NewFunction(func(args ...flagrt.Value) flagrt.Value {") {
		t.Fatalf("unexpected lambda setup: %s", def.Setup)
	}
	if def.ResultExpr != "flagrt.ValueToAny(triple)" {
		t.Fatalf("unexpected lambda result expr: %s", def.ResultExpr)
	}

	call, err := c.CompileLine(`(triple 4)`)
	if err != nil {
		t.Fatalf("CompileLine lambda call returned error: %v", err)
	}
	if call.ResultExpr != "flagrt.ValueToAny(flagrt.Call(triple, flagrt.NewLong(4)))" {
		t.Fatalf("unexpected lambda call expression: %s", call.ResultExpr)
	}
}
