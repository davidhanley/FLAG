package compiler

import (
	"fmt"
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

func TestCompileTokensPrintProgram(t *testing.T) {
	output, err := CompileTokens(TokenizeSourceToChannel(`
(ns hello.core)
(println "Hello")
(print 42)
`))
	if err != nil {
		t.Fatalf("CompileTokens returned error: %v", err)
	}

	got := string(output)
	for _, want := range []string{
		"package main",
		`fmt.Println(flagrt.Str("Hello"))`,
		"fmt.Print(flagrt.ValueToAny(flagrt.NewLong(42)))",
		"// Source namespace: hello.core",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go did not contain %q:\n%s", want, got)
		}
	}
}

func TestCompileDocstringsArePreservedAsComments(t *testing.T) {
	output, err := Compile(`
(defn greet "Say hi" [name] name)
(def msg "A greeting" 42)
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}

	got := string(output)
	for _, want := range []string{
		"// Say hi",
		"// A greeting",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go did not contain %q:\n%s", want, got)
		}
	}
}

func TestCompileDocstringDefmacroStillExpands(t *testing.T) {
	output, err := Compile(`
(defmacro identity "Return the argument unchanged" [x] x)
(println (identity 7))
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}

	got := string(output)
	if !strings.Contains(got, `fmt.Println(flagrt.Str(flagrt.NewLong(7)))`) {
		t.Fatalf("generated Go did not contain expected macro expansion:\n%s", got)
	}
}

func TestCompileDeftestAndAssertions(t *testing.T) {
	output, err := Compile(`
(deftest sample-test
  (testing "inner"
    (let [x 1]
      (is (= x 1))
      (is (= x 2) "x should be 2"))))
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}

	got := string(output)
	for _, want := range []string{
		"func sample_test() flagrt.Value {",
		"func runFlagTestCase(tc flagTestCase) (passed bool) {",
		`panic("at 5:7: (= x 1)")`,
		`panic("at 6:7: x should be 2")`,
		"return flagrt.NilValue()",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go did not contain %q:\n%s", want, got)
		}
	}
}

func TestCompileMainDefnBecomesProgramEntryPoint(t *testing.T) {
	output, err := Compile(`
(defn main [& _args]
  (println "running"))
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}

	got := string(output)
	for _, want := range []string{
		"func flag_main_variadic(args ...flagrt.Value) flagrt.Value {",
		`args := make([]flagrt.Value, 0, len(os.Args)-1)`,
		`_ = flagrt.Call(flag_main, args...)`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go did not contain %q:\n%s", want, got)
		}
	}
}

func TestCompileGoWithoutAsyncImportFails(t *testing.T) {
	_, err := Compile(`
{:namespace "x"}
(go (println "nope"))
`)
	if err == nil {
		t.Fatal("expected unknown symbol without async import")
	}
}

func TestCompileModuleMainIsProgramEntryPoint(t *testing.T) {
	output, err := Compile(`
{:namespace "app"}
(defn main [& _args]
  (println "running"))
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}
	got := string(output)
	for _, want := range []string{
		"func app__flag_main_variadic",
		`_ = flagrt.Call(app__flag_main, args...)`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go did not contain %q:\n%s", want, got)
		}
	}
}

func TestCompileDefnDashRejected(t *testing.T) {
	_, err := Compile(`
(defn- hidden-helper [x] (+ x 1))
`)
	if err == nil {
		t.Fatal("expected defn- to be rejected")
	}
	if !strings.Contains(err.Error(), "defn- is not supported") {
		t.Fatalf("unexpected error: %v", err)
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
	_, err := Compile("(try 1)")
	if err == nil {
		t.Fatal("Compile succeeded for unsupported form")
	}
}

func TestCompileLoopRecurGeneratesIterativeCode(t *testing.T) {
	output, err := Compile(`
(println
  (loop [n 3 acc 0]
    (if (= n 0)
      acc
      (recur (- n 1) (+ acc n)))))
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}

	got := string(output)
	for _, want := range []string{
		"for {",
		"flagrt.UnwrapRecur(__loopResult)",
		"flagrt.NewRecur(",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go did not contain %q:\n%s", want, got)
		}
	}
}

func TestCompileRejectsRecurOutsideLoop(t *testing.T) {
	_, err := Compile(`(recur 1)`)
	if err == nil {
		t.Fatal("expected recur outside loop to fail")
	}
	if !strings.Contains(err.Error(), "recur can only be used within loop") {
		t.Fatalf("unexpected error: %v", err)
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

func TestCompileMapLiteralErrorIncludesLocation(t *testing.T) {
	_, err := Compile("(println {:a (< 1 2)})")
	if err == nil {
		t.Fatal("Compile succeeded for invalid map literal")
	}
	if !strings.Contains(err.Error(), "map literal entries must evaluate to Value at 1:") {
		t.Fatalf("expected location in error, got: %v", err)
	}
}

func TestCompileWithOpenEmitsDefer(t *testing.T) {
	output, err := Compile(`
(with-open [rdr (open-file "sample.txt")]
  (first (file-to-strings rdr)))
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}

	got := string(output)
	for _, want := range []string{
		"defer __bind0.Close()",
		"flagrt.OpenFile(\"sample.txt\")",
		"flagrt.First(flagrt.FileToStrings(rdr))",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go did not contain %q:\n%s", want, got)
		}
	}
}

func TestCompileDoallAndStringPredicates(t *testing.T) {
	output, err := Compile(`
(doall (remove str/blank? (map str/upper-case [" a " "" "b"])))
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}

	got := string(output)
	for _, want := range []string{
		`flagrt.DoAll(`,
		`flagrt.GoBind_packages_StringBlank`,
		`flagrt.GoBind_packages_StringUpperCase`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go did not contain %q:\n%s", want, got)
		}
	}
}

func TestCompileLineSeqAndIOReader(t *testing.T) {
	output, err := Compile(`
(line-seq (io/reader "sample.txt"))
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}
	got := string(output)
	for _, want := range []string{
		`flagrt.LineSeq(`,
		`flagrt.GoBind_runtime_OpenFile`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go did not contain %q:\n%s", want, got)
		}
	}
}

func TestCompileUnknownSymbolIncludesLineNumber(t *testing.T) {
	_, err := Compile(`
(println nope)
`)
	if err == nil {
		t.Fatal("Compile succeeded unexpectedly")
	}
	if !strings.Contains(err.Error(), `unknown symbol "nope"`) || !strings.Contains(err.Error(), "at 2:") {
		t.Fatalf("expected line-numbered unknown symbol error, got: %v", err)
	}
}

func TestCompileDefnArityErrorIncludesLineNumber(t *testing.T) {
	_, err := Compile(`
(defn broken [x])
`)
	if err == nil {
		t.Fatal("Compile succeeded unexpectedly")
	}
	if !strings.Contains(err.Error(), "defn expects name, optional docstring, vector params, and body") || !strings.Contains(err.Error(), "at 2:") {
		t.Fatalf("expected line-numbered defn error, got: %v", err)
	}
}

func TestCompileUpdateBangRejectsNonMutableBinding(t *testing.T) {
	_, err := Compile(`
(let [seen {}]
  (update! seen (assoc seen :a 1)))
`)
	if err == nil {
		t.Fatal("Compile succeeded unexpectedly")
	}
	if !strings.Contains(err.Error(), "update! first argument must be a mutable let binding") || !strings.Contains(err.Error(), "at 3:") {
		t.Fatalf("expected line-numbered update! mutability error, got: %v", err)
	}
}

func TestCompileGoFunctionInteropForms(t *testing.T) {
	output, err := Compile(`
(def f (go-fn "fmt.Sprintf"))
(f "dave is %d years old" 23)
(println (go-fn-args "fmt.Sprintf"))
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}

	got := string(output)
	for _, want := range []string{
		`var f = flagrt.GoFunction("fmt.Sprintf")`,
		`_ = flagrt.Call(f, flagStr_dave_is__d_years_old, flagrt.NewLong(23))`,
		`fmt.Println(flagrt.Str(flagrt.GoFunctionArgs("fmt.Sprintf")))`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go did not contain %q:\n%s", want, got)
		}
	}
}

func TestCompileSlashQualifiedGoFunctions(t *testing.T) {
	output, err := Compile(`
(println (string/trim "  hello  "))
(println (string/replace "hello world" "world" "FLAG"))
(println (str/join "-" ["male" "18-34" "overall"]))
(println (str/capitalize "hELLO"))
(println (character/toUppercase "hello"))
(println (long/parse "42"))
(println (math/abs -42))
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}

	got := string(output)
	for _, want := range []string{
		`flagrt.Call(flagrt.GoBind_packages_StringTrim, flagStr___hello__)`,
		`flagrt.Call(flagrt.GoBind_packages_StringReplace, flagStr_hello_world, flagStr_world, flagStr_FLAG)`,
		`flagrt.Call(flagrt.GoBind_packages_StringJoin, flagStr__, flagVec)`,
		`flagrt.Call(flagrt.GoBind_packages_StringCapitalize, flagStr_hELLO)`,
		`flagrt.Call(flagrt.GoBind_packages_ToUppercase, flagStr_hello)`,
		`flagrt.Call(flagrt.GoBind_packages_LongParse, flagStr_42)`,
		`flagrt.Call(flagrt.GoBind_runtime_Abs, flagrt.NewLong(-42))`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go did not contain %q:\n%s", want, got)
		}
	}

	// Statically bound calls must not fall back to the runtime reflection bridge.
	if strings.Contains(got, "GoFunction(") {
		t.Fatalf("expected no runtime GoFunction lookup for namespaced calls:\n%s", got)
	}
}

func TestCompileUnknownGoFunctionIsCompileError(t *testing.T) {
	_, err := Compile(`(defn oops [s] (str/trmi s))`)
	if err == nil {
		t.Fatal("expected a compile error for an unknown go function, got nil")
	}
	if !strings.Contains(err.Error(), `unknown symbol "str/trmi"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAnnotateExprErrorAddsPosition(t *testing.T) {
	err := annotateExprError(ListExpr{Line: 335, Col: 10}, fmt.Errorf("mapcat expects function and one collection"))
	if err == nil {
		t.Fatal("expected annotated error")
	}
	got := err.Error()
	if got != "mapcat expects function and one collection at 335:10" {
		t.Fatalf("unexpected annotated error: %q", got)
	}
}

func TestCompilePMapFunction(t *testing.T) {
	output, err := Compile(`
(println (pmap (fn [x] (* x x)) [1 2 3]))
(defn add2 [a b] (+ a b))
(println (pmap add2 [1 2 3] [10 20 30]))
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}

	got := string(output)
	for _, want := range []string{
		`fmt.Println(flagrt.Str(flagrt.PMap(flagrt.NewFunction(func(args ...flagrt.Value) flagrt.Value {`,
		`fmt.Println(flagrt.Str(flagrt.PMap(add2, flagVec, flagVec_1)))`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go did not contain %q:\n%s", want, got)
		}
	}
}

func TestCompileOpenFileAndFileToStrings(t *testing.T) {
	output, err := Compile(`(println (first (file-to-strings (open-file "sample.txt"))))`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}

	got := string(output)
	for _, want := range []string{
		`flagrt.OpenFile("sample.txt")`,
		`flagrt.FileToStrings(`,
		`flagrt.First(`,
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
	want := "_ = flagrt.Call(fib, flagrt.NewLong(7))"
	if !strings.Contains(got, want) {
		t.Fatalf("generated Go did not contain expected top-level execution statement %q:\n%s", want, got)
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

func TestCompileDefrecordGoStruct(t *testing.T) {
	output, err := Compile(`
(defrecord SourceToken [^string token ^long line ^long offset])
(println (:token (->SourceToken "x" 1 2)))
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}
	got := string(output)
	for _, want := range []string{
		"type SourceToken struct {",
		"`flag:\"token\"`",
		"`flag:\"line\"`",
		"`flag:\"offset\"`",
		"Line   int64",
		"flagrt.NewRecord(SourceToken{",
		"flagrt.RequireString(token)",
		"flagrt.RequireLong(line)",
		"flagrt.RequireLong(offset)",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go did not contain %q:\n%s", want, got)
		}
	}
}

func TestCompileGoInterface(t *testing.T) {
	output, err := Compile(`
(defn square [x] (* x x))
(go-interface square double [double])
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}

	got := string(output)
	for _, want := range []string{
		"func square_go(p1 float64) float64 {",
		"args := make([]flagrt.Value, 0, 1)",
		"args = append(args, flagrt.NewDouble(p1))",
		"result := flagrt.Call(",
		"return result.Double()",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go did not contain %q:\n%s", want, got)
		}
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
		"var sq_arity_1 func(flagrt.Value) flagrt.Value",
		"sq_arity_1 = func(x flagrt.Value) flagrt.Value {",
		"var sq_variadic func(args ...flagrt.Value) flagrt.Value",
		"sq_variadic = func(args ...flagrt.Value) flagrt.Value {",
		"var sq flagrt.Value",
		"if len(args) != 1 {",
		`panic("sq expects exactly 1 arguments")`,
		"return sq_arity_1(args[0])",
		"return flagrt.Mul(x, x)",
		"sq = flagrt.NewFunction(sq_variadic)",
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
	if call.ResultExpr != "flagrt.ValueToAny(flagrt.Call(sq, flagrt.NewLong(4)))" {
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

func TestReplCompilerDefrecordBinding(t *testing.T) {
	c := NewReplCompiler()

	rec, err := c.CompileLine(`(defrecord Food [weight calories])`)
	if err != nil {
		t.Fatalf("CompileLine defrecord returned error: %v", err)
	}
	for _, want := range []string{
		"type Food struct {",
		"`flag:\"weight\"`",
		"`flag:\"calories\"`",
		"flagrt.NewRecord(Food{",
		"var __gtFood_arity_2 func(flagrt.Value, flagrt.Value) flagrt.Value",
		"__gtFood = flagrt.NewFunction(__gtFood_variadic)",
	} {
		if !strings.Contains(rec.Setup, want) {
			t.Fatalf("defrecord setup did not contain %q:\n%s", want, rec.Setup)
		}
	}

	lookup, err := c.CompileLine(`(:weight (->Food 10 200))`)
	if err != nil {
		t.Fatalf("CompileLine record lookup returned error: %v", err)
	}
	if !strings.Contains(lookup.ResultExpr, "flagrt.Call(__gtFood, flagrt.NewLong(10), flagrt.NewLong(200))") {
		t.Fatalf("unexpected record lookup expression: %s", lookup.ResultExpr)
	}
}
