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

func TestCompileDefnAllowsHyphenatedNames(t *testing.T) {
	output, err := Compile(`
(defn get-scores-list-base [base] base)
(get-scores-list-base 7)
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}

	got := string(output)
	if !strings.Contains(got, "func get_scores_list_base_arity_1(") {
		t.Fatalf("generated Go did not contain hyphenated identifier rewrite:\n%s", got)
	}
	if !strings.Contains(got, `_ = flagrt.Call(get_scores_list_base, flagrt.NewLong(7))`) {
		t.Fatalf("generated Go did not contain expected call:\n%s", got)
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

func TestCompileDefnAllowsPredicateAndBangNames(t *testing.T) {
	output, err := Compile(`
(defn foreign-name? [x] x)
(defn mutate! [x] x)
(foreign-name? 1)
(mutate! 2)
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}
	got := string(output)
	for _, want := range []string{
		"func foreign_name_q_arity_1(",
		"func mutate_bang_arity_1(",
		"_ = flagrt.Call(foreign_name_q, flagrt.NewLong(1))",
		"_ = flagrt.Call(mutate_bang, flagrt.NewLong(2))",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go did not contain %q:\n%s", want, got)
		}
	}
}

func TestCompileDefnDashAliasWorks(t *testing.T) {
	output, err := Compile(`
(defn- hidden-helper [x] (+ x 1))
(hidden-helper 2)
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}

	got := string(output)
	if !strings.Contains(got, "func hidden_helper_arity_1(") {
		t.Fatalf("generated Go did not contain defn- lowering:\n%s", got)
	}
	if !strings.Contains(got, `_ = flagrt.Call(hidden_helper, flagrt.NewLong(2))`) {
		t.Fatalf("generated Go did not contain defn- call:\n%s", got)
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

func TestCompilePrintlnWithRatioLiteral(t *testing.T) {
	output, err := Compile(`(println 5/6)`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}

	got := string(output)
	if !strings.Contains(got, `fmt.Println(flagrt.Str(flagrt.NewRatio(5, 6)))`) {
		t.Fatalf("generated Go did not contain ratio literal:\n%s", got)
	}
}

func TestCompilePrintlnWithBigIntLiteral(t *testing.T) {
	output, err := Compile(`(println 10N)`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}

	got := string(output)
	if !strings.Contains(got, `fmt.Println(flagrt.Str(flagrt.NewBigIntFromString("10")))`) {
		t.Fatalf("generated Go did not contain bigint literal:\n%s", got)
	}
}

func TestCompilePrintlnWithCharLiteral(t *testing.T) {
	output, err := Compile(`(println \M)`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}

	got := string(output)
	if !strings.Contains(got, `fmt.Println(flagrt.Str(flagrt.NewString("M")))`) {
		t.Fatalf("generated Go did not contain char literal:\n%s", got)
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

func TestCompileMapLiteralWithCharKeys(t *testing.T) {
	output, err := Compile(`(println {\M :male \F :female})`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}

	got := string(output)
	if !strings.Contains(got, `flagrt.NewString("M")`) || !strings.Contains(got, `flagrt.NewString("F")`) {
		t.Fatalf("generated Go did not contain char map keys:\n%s", got)
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
		`flagrt.Remove(`,
		`flagrt.GoFunction("str/blank?")`,
		`flagrt.GoFunction("str/upper-case")`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go did not contain %q:\n%s", want, got)
		}
	}
}

func TestCompileDefnAllowsMultipleBodyForms(t *testing.T) {
	output, err := Compile(`
(defn greet [name]
  (println "hi")
  name)
(println (greet "Ada"))
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}
	got := string(output)
	for _, want := range []string{
		`_ = flagrt.Println("hi")`,
		`return name`,
		`fmt.Println(flagrt.Str(flagrt.Call(greet, flagrt.NewString("Ada"))))`,
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
		`flagrt.GoFunction("io/reader")`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go did not contain %q:\n%s", want, got)
		}
	}
}

func TestCompileLastExpression(t *testing.T) {
	output, err := Compile(`(println (last [1 2 3]))`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}
	got := string(output)
	if !strings.Contains(got, `flagrt.Last(flagrt.NewArray(flagrt.NewLong(1), flagrt.NewLong(2), flagrt.NewLong(3)))`) {
		t.Fatalf("generated Go did not contain last helper:\n%s", got)
	}
}

func TestCompileReverseExpression(t *testing.T) {
	output, err := Compile(`(println (reverse [1 2 3]))`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}
	got := string(output)
	if !strings.Contains(got, `flagrt.Reverse(flagrt.NewArray(flagrt.NewLong(1), flagrt.NewLong(2), flagrt.NewLong(3)))`) {
		t.Fatalf("generated Go did not contain reverse helper:\n%s", got)
	}
}

func TestCompileConsExpression(t *testing.T) {
	output, err := Compile(`(println (cons 1 nil))`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}
	got := string(output)
	if !strings.Contains(got, `flagrt.Cons(flagrt.NewLong(1), flagrt.NilValue())`) {
		t.Fatalf("generated Go did not contain cons helper:\n%s", got)
	}
}

func TestCompileRegexCompileAndMatches(t *testing.T) {
	output, err := Compile(`
(deftest regex-test
  (is ((regex/compile "^he.*o$") "hello"))
  (is (re-matches (re-pattern "^he.*o$") "hello")))
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}

	got := string(output)
	for _, want := range []string{
		`flagrt.GoFunction("regex/compile")`,
		`flagrt.BuiltinFunction("re-pattern")`,
		`flagrt.BuiltinFunction("re-matches")`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go did not contain %q:\n%s", want, got)
		}
	}
}

func TestCompileSomeExpression(t *testing.T) {
	output, err := Compile(`
(some (fn [x] (when (> x 2) x)) [1 2 3 4])
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}
	got := string(output)
	for _, want := range []string{
		`flagrt.Some(`,
		`flagrt.NewFunction`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go did not contain %q:\n%s", want, got)
		}
	}
}

func TestCompileSomePredicateExpression(t *testing.T) {
	output, err := Compile(`
(let [x 1]
  (some? x))
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}
	if !strings.Contains(string(output), `flagrt.SomePredicate(`) {
		t.Fatalf("generated Go did not contain some? lowering:\n%s", string(output))
	}
}

func TestCompileSeqPredicateExpression(t *testing.T) {
	output, err := Compile(`
(seq? [1 2 3])
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}
	got := string(output)
	for _, want := range []string{
		`flagrt.SeqPredicate(`,
		`flagrt.NewArray`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go did not contain %q:\n%s", want, got)
		}
	}
}

func TestCompileSeqExpression(t *testing.T) {
	output, err := Compile(`
(seq [1 2 3])
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}
	got := string(output)
	for _, want := range []string{
		`flagrt.Seq(`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go did not contain %q:\n%s", want, got)
		}
	}
}

func TestCompileVecExpression(t *testing.T) {
	output, err := Compile(`
(vec (seq [1 2 3]))
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}
	got := string(output)
	if !strings.Contains(got, `flagrt.Vec(`) {
		t.Fatalf("generated Go did not contain vec lowering:\n%s", got)
	}
}

func TestCompileNotExpression(t *testing.T) {
	output, err := Compile(`
(not false)
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}
	if !strings.Contains(string(output), `flagrt.NewBool(!flagrt.IsTruthy(`) {
		t.Fatalf("generated Go did not contain not lowering:\n%s", string(output))
	}
}

func TestCompileFormatExpression(t *testing.T) {
	output, err := Compile(`
(format "%.2f" 1.23)
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}
	if !strings.Contains(string(output), `flagrt.Format(`) {
		t.Fatalf("generated Go did not contain format lowering:\n%s", string(output))
	}
}

func TestCompileDoubleExpression(t *testing.T) {
	output, err := Compile(`
(double 3)
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}
	if !strings.Contains(string(output), `flagrt.Double(`) {
		t.Fatalf("generated Go did not contain double lowering:\n%s", string(output))
	}
}

func TestCompileIntoExpression(t *testing.T) {
	output, err := Compile(`
(into {} (map (fn [x] [x x]) [1 2]))
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}
	if !strings.Contains(string(output), `flagrt.Into(`) {
		t.Fatalf("generated Go did not contain into lowering:\n%s", string(output))
	}
}

func TestCompileForExpression(t *testing.T) {
	output, err := Compile(`
(for [x [1 2]] [:li x])
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}
	got := string(output)
	for _, want := range []string{
		`flagrt.MapCat(`,
		`flagrt.NewFunction`,
		`flagrt.NewArray`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go did not contain %q:\n%s", want, got)
		}
	}
}

func TestCompileForNestedBindings(t *testing.T) {
	output, err := Compile(`
(for [x [1 2] y [10 20]] [x y])
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}
	got := string(output)
	if strings.Count(got, "flagrt.MapCat(") < 2 {
		t.Fatalf("expected nested MapCat in generated Go:\n%s", got)
	}
}

func TestCompileNotAnyExpression(t *testing.T) {
	output, err := Compile(`
(not-any? (fn [x] (> x 2)) [1 2 3 4])
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}
	got := string(output)
	for _, want := range []string{
		`flagrt.NotAny(`,
		`flagrt.NewFunction`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go did not contain %q:\n%s", want, got)
		}
	}
}

func TestCompileEveryExpression(t *testing.T) {
	output, err := Compile(`
(every? (fn [x] (> x 0)) [1 2 3 4])
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}
	got := string(output)
	for _, want := range []string{
		`flagrt.Every(`,
		`flagrt.NewFunction`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go did not contain %q:\n%s", want, got)
		}
	}
}

func TestCompileKeepExpression(t *testing.T) {
	output, err := Compile(`
(keep (fn [x] (when (> x 2) x)) [1 2 3 4])
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}
	got := string(output)
	for _, want := range []string{
		`flagrt.Keep(`,
		`flagrt.NewFunction`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go did not contain %q:\n%s", want, got)
		}
	}
}

func TestCompileSetExpression(t *testing.T) {
	output, err := Compile(`
(set [1 1 2])
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}
	if !strings.Contains(string(output), `flagrt.Set(`) {
		t.Fatalf("generated Go did not contain set helper:\n%s", string(output))
	}
}

func TestCompileConjExpression(t *testing.T) {
	output, err := Compile(`
(conj [1 2] 3)
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}
	if !strings.Contains(string(output), `flagrt.Conj(`) {
		t.Fatalf("generated Go did not contain conj helper:\n%s", string(output))
	}
}

func TestCompileContainsExpression(t *testing.T) {
	output, err := Compile(`
(contains? #{1 2} 2)
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}
	got := string(output)
	if !strings.Contains(got, `flagrt.Contains(`) {
		t.Fatalf("generated Go did not contain contains helper:\n%s", got)
	}
}

func TestCompileNotEmptyExpression(t *testing.T) {
	output, err := Compile(`
(not-empty [1 2])
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}
	if !strings.Contains(string(output), `flagrt.NotEmpty(`) {
		t.Fatalf("generated Go did not contain not-empty helper:\n%s", string(output))
	}
}

func TestCompileNotEmptyAndEmptyPredicates(t *testing.T) {
	output, err := Compile(`
(not-empty? [1 2])
(empty? [])
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}
	got := string(output)
	for _, want := range []string{
		`flagrt.NewBool(flagrt.IsNotEmpty(`,
		`flagrt.NewBool(flagrt.IsEmpty(`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go did not contain %q:\n%s", want, got)
		}
	}
}

func TestCompileNilPredicate(t *testing.T) {
	output, err := Compile(`
(nil? nil)
(nil? "x")
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}
	got := string(output)
	for _, want := range []string{
		`flagrt.NewBool(flagrt.IsNil(flagrt.NilValue()))`,
		`flagrt.NewBool(flagrt.IsNil(flagrt.NewString("x")))`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go did not contain %q:\n%s", want, got)
		}
	}
}

func TestCompileCountExpression(t *testing.T) {
	output, err := Compile(`
(count [1 2 3])
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}
	if !strings.Contains(string(output), `flagrt.Count(`) {
		t.Fatalf("generated Go did not contain count helper:\n%s", string(output))
	}
	if !strings.Contains(string(output), `flagrt.NewLong(int64(flagrt.Count(`) {
		t.Fatalf("expected count to return NewLong(int64(...)):\n%s", string(output))
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

func TestCompilePrintlnWithStringEqualsExpression(t *testing.T) {
	output, err := Compile(`(println (= "a" "a"))`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}

	got := string(output)
	want := `fmt.Println(flagrt.Str(flagrt.NewBool(flagrt.Eq(flagrt.NewString("a"), flagrt.NewString("a")))))`
	if !strings.Contains(got, want) {
		t.Fatalf("generated Go did not contain expected string equals expression:\n%s", got)
	}
}

func TestCompilePrintlnWithComparisonExpressions(t *testing.T) {
	output, err := Compile(`
(println (< 1 2 3.0))
(println (<= 1 2 2))
(println (> 3 2 1))
(println (>= 3 2 2))
(println (< "a" "b"))
(println (>= "b" "a"))
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}

	got := string(output)
	for _, want := range []string{
		"fmt.Println(flagrt.Str(flagrt.Lt(flagrt.NewLong(1), flagrt.NewLong(2)) && flagrt.Lt(flagrt.NewLong(2), flagrt.NewDouble(3.0))))",
		"fmt.Println(flagrt.Str(flagrt.Le(flagrt.NewLong(1), flagrt.NewLong(2)) && flagrt.Le(flagrt.NewLong(2), flagrt.NewLong(2))))",
		"fmt.Println(flagrt.Str(flagrt.Gt(flagrt.NewLong(3), flagrt.NewLong(2)) && flagrt.Gt(flagrt.NewLong(2), flagrt.NewLong(1))))",
		"fmt.Println(flagrt.Str(flagrt.Ge(flagrt.NewLong(3), flagrt.NewLong(2)) && flagrt.Ge(flagrt.NewLong(2), flagrt.NewLong(2))))",
		`fmt.Println(flagrt.Str(flagrt.Lt(flagrt.NewString("a"), flagrt.NewString("b"))))`,
		`fmt.Println(flagrt.Str(flagrt.Ge(flagrt.NewString("b"), flagrt.NewString("a"))))`,
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
		"var a = flagrt.NewLong(1)",
		"var b = flagrt.Add(flagrt.NewLong(1), a)",
		"return b",
		"fmt.Println(flagrt.Str(func() flagrt.Value {",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go did not contain %q:\n%s", want, got)
		}
	}
}

func TestCompileLetAllowsStringAndBoolBindings(t *testing.T) {
	output, err := Compile(`
(println
  (let [c1 "#d8dbff"
        ok (> 3 2)]
    (if ok c1 "no")))
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}

	got := string(output)
	for _, want := range []string{
		`var c1 = "#d8dbff"`,
		`var ok = flagrt.Gt(flagrt.NewLong(3), flagrt.NewLong(2))`,
		`return c1`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go did not contain %q:\n%s", want, got)
		}
	}
}

func TestCompileLetVolatileBindingAndUpdateBang(t *testing.T) {
	output, err := Compile(`
(println
  (let [^{:volatile true} seen {:a 1}]
    (update! seen (assoc seen :b 2))
    seen))
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}
	got := string(output)
	for _, want := range []string{
		"var seen = __bind0",
		"seen = flagrt.MapAssoc(seen, flagrt.NewKeyword(\"b\"), flagrt.NewLong(2))",
		"return seen",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go did not contain %q:\n%s", want, got)
		}
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

func TestCompileLetVectorDestructuring(t *testing.T) {
	output, err := Compile(`
(println (let [[a b & rest :as all] [1 2 3 4]]
  (+ a (first rest))))
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}

	got := string(output)
	for _, want := range []string{
		"flagrt.SeqFirst(__bind0)",
		"flagrt.SeqRest(__bind0)",
		"var a = __dseq",
		"var b = __dseq",
		"var rest = flagrt.SeqRest(flagrt.SeqRest(__bind0))",
		"var all = __bind0",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go did not contain %q:\n%s", want, got)
		}
	}
}

func TestCompileLetMapDestructuring(t *testing.T) {
	output, err := Compile(`
(println (let [{:keys [a b] :or {b 9} :as m} {:a 1}]
  (+ a b)))
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}

	got := string(output)
	for _, want := range []string{
		`var m = __bind0`,
		`flagrt.Get(__bind0, flagrt.NewKeyword("a"))`,
		`flagrt.Get(__bind0, flagrt.NewKeyword("b"), flagrt.NewLong(9))`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go did not contain %q:\n%s", want, got)
		}
	}
}

func TestCompileDefnWithDestructuredParams(t *testing.T) {
	output, err := Compile(`
(defn f [[a b] {:keys [x]}]
  (+ a x))
(println (f [1 2] {:x 3}))
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}

	got := string(output)
	for _, want := range []string{
		"func f_arity_2(__arg0 flagrt.Value, __arg1 flagrt.Value) flagrt.Value {",
		"flagrt.SeqFirst(__arg0)",
		`flagrt.Get(__arg1, flagrt.NewKeyword("x"))`,
		"return flagrt.Add(a, x)",
		"fmt.Println(flagrt.Str(flagrt.Call(f, flagrt.NewArray(flagrt.NewLong(1), flagrt.NewLong(2)), flagrt.NewMap(flagrt.NewKeyword(\"x\"), flagrt.NewLong(3)))))",
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

func TestCompilePrintlnWithQuotedListLiteral(t *testing.T) {
	output, err := Compile(`(println '(1 2 3))`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}

	got := string(output)
	want := `fmt.Println(flagrt.Str(flagrt.NewList(flagrt.NewLong(1), flagrt.NewLong(2), flagrt.NewLong(3))))`
	if !strings.Contains(got, want) {
		t.Fatalf("generated Go did not contain expected quoted list literal:\n%s", got)
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

func TestCompilePrintlnWithStringCollectionLiterals(t *testing.T) {
	output, err := Compile(`
(println {:name "DAVID HANLEY" :age 30})
(println ["hello" {:msg "world"}])
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}

	got := string(output)
	for _, want := range []string{
		`flagrt.NewString("DAVID HANLEY")`,
		`flagrt.NewString("hello")`,
		`flagrt.NewString("world")`,
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
(println (keyword "abc"))
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}

	got := string(output)
	for _, want := range []string{
		`fmt.Println(flagrt.Str(flagrt.Name(flagrt.NewKeyword("xyz"))))`,
		`fmt.Println(flagrt.Str(flagrt.Name(flagrt.Symbol("abc"))))`,
		`fmt.Println(flagrt.Str(flagrt.Symbol(flagrt.NewKeyword("xyz"))))`,
		`fmt.Println(flagrt.Str(flagrt.Keyword("abc")))`,
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

func TestCompileUpdateFunction(t *testing.T) {
	output, err := Compile(`
(println (update {:a 1} :a + 2))
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}
	got := string(output)
	if !strings.Contains(got, `flagrt.Update(`) {
		t.Fatalf("generated Go did not contain update helper:\n%s", got)
	}
}

func TestCompileMapLookupForms(t *testing.T) {
	output, err := Compile(`
(def m {:a 1})
(println (get m :a))
(println (:a m))
(println (m :a))
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}

	got := string(output)
	for _, want := range []string{
		`fmt.Println(flagrt.Str(flagrt.Call(flagrt.BuiltinFunction("get"), m, flagrt.NewKeyword("a"))))`,
		`fmt.Println(flagrt.Str(flagrt.Call(flagrt.NewKeyword("a"), m)))`,
		`fmt.Println(flagrt.Str(flagrt.Call(m, flagrt.NewKeyword("a"))))`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go did not contain %q:\n%s", want, got)
		}
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
		`_ = flagrt.Call(f, flagrt.NewString("dave is %d years old"), flagrt.NewLong(23))`,
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
		`flagrt.Call(flagrt.GoFunction("string/trim"), flagrt.NewString("  hello  "))`,
		`flagrt.Call(flagrt.GoFunction("string/replace"), flagrt.NewString("hello world"), flagrt.NewString("world"), flagrt.NewString("FLAG"))`,
		`flagrt.Call(flagrt.GoFunction("str/join"), flagrt.NewString("-"), flagrt.NewArray(flagrt.NewString("male"), flagrt.NewString("18-34"), flagrt.NewString("overall")))`,
		`flagrt.Call(flagrt.GoFunction("str/capitalize"), flagrt.NewString("hELLO"))`,
		`flagrt.Call(flagrt.GoFunction("character/toUppercase"), flagrt.NewString("hello"))`,
		`flagrt.Call(flagrt.GoFunction("long/parse"), flagrt.NewString("42"))`,
		`flagrt.Call(flagrt.GoFunction("math/abs"), flagrt.NewLong(-42))`,
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
(println (next [1 2 3]))
(println (take 2 [1 2 3]))
(println (drop 1 [1 2 3]))
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}

	got := string(output)
	for _, want := range []string{
		`fmt.Println(flagrt.Str(flagrt.First(flagrt.NewArray(flagrt.NewLong(1), flagrt.NewLong(2), flagrt.NewLong(3)))))`,
		`fmt.Println(flagrt.Str(flagrt.Rest(flagrt.NewArray(flagrt.NewLong(1), flagrt.NewLong(2), flagrt.NewLong(3)))))`,
		`fmt.Println(flagrt.Str(flagrt.Next(flagrt.NewArray(flagrt.NewLong(1), flagrt.NewLong(2), flagrt.NewLong(3)))))`,
		`fmt.Println(flagrt.Str(flagrt.Take(flagrt.NewLong(2), flagrt.NewArray(flagrt.NewLong(1), flagrt.NewLong(2), flagrt.NewLong(3)))))`,
		`fmt.Println(flagrt.Str(flagrt.Drop(flagrt.NewLong(1), flagrt.NewArray(flagrt.NewLong(1), flagrt.NewLong(2), flagrt.NewLong(3)))))`,
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
		`fmt.Println(flagrt.Str(flagrt.Map(add2, flagrt.NewArray(flagrt.NewLong(1), flagrt.NewLong(2), flagrt.NewLong(3)), flagrt.NewArray(flagrt.NewLong(10), flagrt.NewLong(20), flagrt.NewLong(30)))))`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go did not contain %q:\n%s", want, got)
		}
	}
}

func TestCompileMapcatFunction(t *testing.T) {
	output, err := Compile(`
(mapcat rest [1 2 3])
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}
	got := string(output)
	if !strings.Contains(got, `flagrt.MapCat(`) {
		t.Fatalf("generated Go did not contain mapcat lowering:\n%s", got)
	}
}

func TestCompileConcatFunction(t *testing.T) {
	output, err := Compile(`
(println (concat [1 2] [3] nil))
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}

	got := string(output)
	want := `fmt.Println(flagrt.Str(flagrt.Concat(flagrt.NewArray(flagrt.NewLong(1), flagrt.NewLong(2)), flagrt.NewArray(flagrt.NewLong(3)), flagrt.NilValue())))`
	if !strings.Contains(got, want) {
		t.Fatalf("generated Go did not contain %q:\n%s", want, got)
	}
}

func TestCompileGroupByFunction(t *testing.T) {
	output, err := Compile(`
(println (group-by (fn [x] (% x 2)) [1 2 3 4]))
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}

	got := string(output)
	for _, want := range []string{
		`flagrt.GroupBy(`,
		`flagrt.NewFunction(func(args ...flagrt.Value) flagrt.Value {`,
		`flagrt.NewArray(flagrt.NewLong(1), flagrt.NewLong(2), flagrt.NewLong(3), flagrt.NewLong(4))`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go did not contain %q:\n%s", want, got)
		}
	}
}

func TestCompileSortByFunction(t *testing.T) {
	output, err := Compile(`
(println (sort-by identity [3 1 2]))
(println (sort-by identity > [3 1 2]))
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}

	got := string(output)
	for _, want := range []string{
		`flagrt.SortBy(flagrt.BuiltinFunction("identity"), flagrt.NewArray(flagrt.NewLong(3), flagrt.NewLong(1), flagrt.NewLong(2)))`,
		`flagrt.SortBy(flagrt.BuiltinFunction("identity"), flagrt.BuiltinFunction(">"), flagrt.NewArray(flagrt.NewLong(3), flagrt.NewLong(1), flagrt.NewLong(2)))`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go did not contain %q:\n%s", want, got)
		}
	}
}

func TestCompileJuxtFunction(t *testing.T) {
	output, err := Compile(`
(println ((juxt str/trim str/upper-case) " hi "))
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}
	got := string(output)
	for _, want := range []string{
		`flagrt.Juxt(`,
		`flagrt.GoFunction("str/trim")`,
		`flagrt.GoFunction("str/upper-case")`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go did not contain %q:\n%s", want, got)
		}
	}
}

func TestCompilePartialFunction(t *testing.T) {
	output, err := Compile(`
(println ((partial + 10) 7))
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}
	got := string(output)
	for _, want := range []string{
		`flagrt.Partial(`,
		`flagrt.BuiltinFunction("+")`,
		`flagrt.NewLong(10)`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go did not contain %q:\n%s", want, got)
		}
	}
}

func TestCompileApplyFunction(t *testing.T) {
	output, err := Compile(`
(println (apply + [1 2 3]))
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}
	got := string(output)
	for _, want := range []string{
		`flagrt.Apply(`,
		`flagrt.BuiltinFunction("+")`,
		`flagrt.NewArray(flagrt.NewLong(1), flagrt.NewLong(2), flagrt.NewLong(3))`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go did not contain %q:\n%s", want, got)
		}
	}
}

func TestCompileHashMapFunction(t *testing.T) {
	output, err := Compile(`
(println (hash-map :a 1 :b 2))
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}
	got := string(output)
	want := `fmt.Println(flagrt.Str(flagrt.NewMap(flagrt.NewKeyword("a"), flagrt.NewLong(1), flagrt.NewKeyword("b"), flagrt.NewLong(2))))`
	if !strings.Contains(got, want) {
		t.Fatalf("generated Go did not contain %q:\n%s", want, got)
	}
}

func TestCompileMaxAndMinFunctions(t *testing.T) {
	output, err := Compile(`
(println (max 1 3/2 2.5))
(println (min 4 5/2 3))
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}
	got := string(output)
	for _, want := range []string{
		`flagrt.Max(`,
		`flagrt.Min(`,
		`flagrt.NewRatio(3, 2)`,
		`flagrt.NewDouble(2.5)`,
		`flagrt.NewRatio(5, 2)`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go did not contain %q:\n%s", want, got)
		}
	}
}

func TestCompileMergeFunction(t *testing.T) {
	output, err := Compile(`
(println (merge {:a 1} {:b 2} {:a 3}))
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}
	got := string(output)
	want := `fmt.Println(flagrt.Str(flagrt.Merge(flagrt.NewMap(flagrt.NewKeyword("a"), flagrt.NewLong(1)), flagrt.NewMap(flagrt.NewKeyword("b"), flagrt.NewLong(2)), flagrt.NewMap(flagrt.NewKeyword("a"), flagrt.NewLong(3)))))`
	if !strings.Contains(got, want) {
		t.Fatalf("generated Go did not contain %q:\n%s", want, got)
	}
}

func TestCompileSelectKeysFunction(t *testing.T) {
	output, err := Compile(`
(println (select-keys {:a 1 :b 2 :c 3} [:c :a]))
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}
	got := string(output)
	want := `fmt.Println(flagrt.Str(flagrt.SelectKeys(flagrt.NewMap(flagrt.NewKeyword("a"), flagrt.NewLong(1), flagrt.NewKeyword("b"), flagrt.NewLong(2), flagrt.NewKeyword("c"), flagrt.NewLong(3)), flagrt.NewArray(flagrt.NewKeyword("c"), flagrt.NewKeyword("a")))))`
	if !strings.Contains(got, want) {
		t.Fatalf("generated Go did not contain %q:\n%s", want, got)
	}
}

func TestCompileMaxKeyAndValFunctions(t *testing.T) {
	output, err := Compile(`
(println (max-key :score {:score 1} {:score 3}))
(println (val [:a 10]))
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}
	got := string(output)
	for _, want := range []string{
		`flagrt.MaxKey(`,
		`flagrt.Val(`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go did not contain %q:\n%s", want, got)
		}
	}
}

func TestCompileZipMapFunction(t *testing.T) {
	output, err := Compile(`
(println (zipmap [:a :b :c] [10 20]))
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}

	got := string(output)
	want := `fmt.Println(flagrt.Str(flagrt.ZipMap(flagrt.NewArray(flagrt.NewKeyword("a"), flagrt.NewKeyword("b"), flagrt.NewKeyword("c")), flagrt.NewArray(flagrt.NewLong(10), flagrt.NewLong(20)))))`
	if !strings.Contains(got, want) {
		t.Fatalf("generated Go did not contain %q:\n%s", want, got)
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
		`fmt.Println(flagrt.Str(flagrt.PMap(add2, flagrt.NewArray(flagrt.NewLong(1), flagrt.NewLong(2), flagrt.NewLong(3)), flagrt.NewArray(flagrt.NewLong(10), flagrt.NewLong(20), flagrt.NewLong(30)))))`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go did not contain %q:\n%s", want, got)
		}
	}
}

func TestCompileJSONFunctions(t *testing.T) {
	output, err := Compile(`
(println (to-json {:a 1 :b [2 3]}))
(println (from-json "{\"a\":1}"))
(println (from-json (to-json {:x 9})))
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}

	got := string(output)
	for _, want := range []string{
		`fmt.Println(flagrt.Str(flagrt.ToJSON(flagrt.NewMap(flagrt.NewKeyword("a"), flagrt.NewLong(1), flagrt.NewKeyword("b"), flagrt.NewArray(flagrt.NewLong(2), flagrt.NewLong(3))))))`,
		`fmt.Println(flagrt.Str(flagrt.FromJSON("{\"a\":1}")))`,
		`fmt.Println(flagrt.Str(flagrt.FromJSON(flagrt.ToJSON(flagrt.NewMap(flagrt.NewKeyword("x"), flagrt.NewLong(9))))))`,
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
		`fmt.Println(flagrt.Str(flagrt.Filter(gt2, flagrt.NewArray(flagrt.NewLong(1), flagrt.NewLong(2), flagrt.NewLong(3), flagrt.NewLong(4)))))`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go did not contain %q:\n%s", want, got)
		}
	}
}

func TestCompileReduceFunction(t *testing.T) {
	output, err := Compile(`
(defn add2 [a b] (+ a b))
(println (reduce add2 [1 2 3]))
(println (reduce add2 10 [1 2 3]))
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}

	got := string(output)
	for _, want := range []string{
		`fmt.Println(flagrt.Str(flagrt.Reduce(add2, flagrt.NewArray(flagrt.NewLong(1), flagrt.NewLong(2), flagrt.NewLong(3)))))`,
		`fmt.Println(flagrt.Str(flagrt.Reduce(add2, flagrt.NewLong(10), flagrt.NewArray(flagrt.NewLong(1), flagrt.NewLong(2), flagrt.NewLong(3)))))`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go did not contain %q:\n%s", want, got)
		}
	}
}

func TestCompileReduceWithBuiltinPlusFunction(t *testing.T) {
	output, err := Compile(`(println (reduce + [1 2 3 4]))`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}
	got := string(output)
	want := `fmt.Println(flagrt.Str(flagrt.Reduce(flagrt.BuiltinFunction("+"), flagrt.NewArray(flagrt.NewLong(1), flagrt.NewLong(2), flagrt.NewLong(3), flagrt.NewLong(4)))))`
	if !strings.Contains(got, want) {
		t.Fatalf("generated Go did not contain %q:\n%s", want, got)
	}
}

func TestCompileRangeFunction(t *testing.T) {
	output, err := Compile(`
(println (range))
(println (range 1 5))
(println (range 5))
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}
	got := string(output)
	for _, want := range []string{
		`fmt.Println(flagrt.Str(flagrt.Range()))`,
		`fmt.Println(flagrt.Str(flagrt.Range(flagrt.NewLong(1), flagrt.NewLong(5))))`,
		`fmt.Println(flagrt.Str(flagrt.Range(flagrt.NewLong(5))))`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go did not contain %q:\n%s", want, got)
		}
	}
}

func TestCompileRepeatFunction(t *testing.T) {
	output, err := Compile(`
(println (repeat "x"))
(println (repeat 3 "x"))
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}
	got := string(output)
	for _, want := range []string{
		`fmt.Println(flagrt.Str(flagrt.Repeat(flagrt.NewString("x"))))`,
		`fmt.Println(flagrt.Str(flagrt.Repeat(flagrt.NewLong(3), flagrt.NewString("x"))))`,
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

func TestCompileWhenNotMacro(t *testing.T) {
	output, err := Compile(`
(println (when-not false 42))
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}
	got := string(output)
	for _, want := range []string{
		"if flagrt.IsTruthy(flagrt.NewBool(false)) {",
		"return flagrt.NilValue()",
		"return flagrt.NewLong(42)",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go did not contain %q:\n%s", want, got)
		}
	}
}

func TestCompileIncDecMacros(t *testing.T) {
	output, err := Compile(`
(println (inc 41))
(println (dec 41))
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}
	got := string(output)
	for _, want := range []string{
		"flagrt.Add(flagrt.NewLong(41), flagrt.NewLong(1))",
		"flagrt.Sub(flagrt.NewLong(41), flagrt.NewLong(1))",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go did not contain %q:\n%s", want, got)
		}
	}
}

func TestCompileWhenLetMacro(t *testing.T) {
	output, err := Compile(`
(println (when-let [x 42] x))
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}
	got := string(output)
	for _, want := range []string{
		"var __flag_when_let_tmp = flagrt.NewLong(42)",
		"if flagrt.IsTruthy(__flag_when_let_tmp) {",
		"var x = __flag_when_let_tmp",
		"return x",
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

func TestCompileCompMacro(t *testing.T) {
	output, err := Compile(`
(def trim-and-upper (comp str/trim str/upper-case))
(println (trim-and-upper "  hi  "))
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}
	got := string(output)
	for _, want := range []string{
		`flagrt.GoFunction("str/trim")`,
		`flagrt.GoFunction("str/upper-case")`,
		"flagrt.NewFunction(func(args ...flagrt.Value) flagrt.Value {",
		"flagrt.Call(__flag_comp_fn_0, flagrt.Call(__flag_comp_fn_1, __flag_comp_x))",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go did not contain %q:\n%s", want, got)
		}
	}
}

func TestCompileSomeThreadFirstMacro(t *testing.T) {
	output, err := Compile(`(println (some-> 5 (+ 1) (* 2)))`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}
	got := string(output)
	for _, want := range []string{
		"var __some_arrow_0 =",
		"var __some_arrow_1 =",
		"flagrt.Eq(__some_arrow_0, flagrt.NilValue())",
		"flagrt.Mul(__some_arrow_1, flagrt.NewLong(2))",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go did not contain %q:\n%s", want, got)
		}
	}
}

func TestCompileSomeThreadLastMacro(t *testing.T) {
	output, err := Compile(`(println (some->> 5 (- 10) (/ 3)))`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}
	got := string(output)
	for _, want := range []string{
		"var __some_arrow_0 =",
		"var __some_arrow_1 =",
		"flagrt.Eq(__some_arrow_0, flagrt.NilValue())",
		"flagrt.Sub(flagrt.NewLong(10), __some_arrow_0)",
		"flagrt.Div(flagrt.NewLong(3), __some_arrow_1)",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go did not contain %q:\n%s", want, got)
		}
	}
}

func TestCompileCondThreadFirstMacro(t *testing.T) {
	output, err := Compile(`(println (cond-> 5 true (+ 1) false (* 2) true (- 3)))`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}
	got := string(output)
	for _, want := range []string{
		"var __cond_arrow_0 =",
		"var __cond_arrow_1 =",
		"var __cond_arrow_2 =",
		"if flagrt.IsTruthy(flagrt.NewBool(true)) {",
		"if flagrt.IsTruthy(flagrt.NewBool(false)) {",
		"flagrt.Add(__cond_arrow_0, flagrt.NewLong(1))",
		"flagrt.Mul(__cond_arrow_1, flagrt.NewLong(2))",
		"flagrt.Sub(__cond_arrow_2, flagrt.NewLong(3))",
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

func TestCompileFnLiteralWithIgnoredParam(t *testing.T) {
	output, err := Compile(`(println ((fn [_] true) 42))`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}
	got := string(output)
	for _, want := range []string{
		`panic("fn expects exactly 1 arguments")`,
		`_ = args[0]`,
		`return flagrt.NewBool(true)`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go did not contain %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, `_ := args[0]`) {
		t.Fatalf("generated Go must not redeclare blank identifier with :=:\n%s", got)
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
		"func sq_arity_1(x flagrt.Value) flagrt.Value {",
		"func sq_variadic(args ...flagrt.Value) flagrt.Value {",
		"if len(args) != 1 {",
		`panic("sq expects exactly 1 arguments")`,
		"return sq_arity_1(args[0])",
		"return flagrt.Mul(x, x)",
		"var sq = flagrt.NewFunction(sq_variadic)",
		"fmt.Println(flagrt.Str(flagrt.Call(sq, flagrt.NewLong(47))))",
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
		"func fib_arity_1(x flagrt.Value) flagrt.Value {",
		"return flagrt.Add(fib_arity_1(flagrt.Sub(x, flagrt.NewLong(1))), fib_arity_1(flagrt.Sub(x, flagrt.NewLong(2))))",
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
		"func even_arity_1(x flagrt.Value) flagrt.Value {",
		"return flagrt.NewBool(flagrt.Eq(flagrt.Mod(x, flagrt.NewLong(2)), flagrt.NewLong(0)))",
		"var even = flagrt.NewFunction(even_variadic)",
		"fmt.Println(flagrt.Str(flagrt.Call(even, flagrt.NewLong(4))))",
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
(defn mkstr [x] (str "id-" x))
(defn mkbool [x] (< x 3))
`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}

	got := string(output)
	for _, want := range []string{
		`func mklist_arity_1(x flagrt.Value) flagrt.Value {`,
		`return flagrt.Rest(flagrt.NewArray(flagrt.NewLong(0), x))`,
		`func mkarray_arity_1(x flagrt.Value) flagrt.Value {`,
		`return flagrt.NewArray(x, flagrt.NewLong(2))`,
		`func mksym_arity_1(x flagrt.Value) flagrt.Value {`,
		`return flagrt.NewSymbol("foo")`,
		`func mkmap_arity_1(x flagrt.Value) flagrt.Value {`,
		`return flagrt.NewMap(flagrt.NewKeyword("a"), x)`,
		`func mkset_arity_1(x flagrt.Value) flagrt.Value {`,
		`return flagrt.NewSet(x)`,
		`func mkfn_arity_1(x flagrt.Value) flagrt.Value {`,
		`return flagrt.NewFunction(func(args ...flagrt.Value) flagrt.Value {`,
		`func mkstr_arity_1(x flagrt.Value) flagrt.Value {`,
		`return flagrt.NewString(func() string {`,
		`return flagrt.Str("id-", x)`,
		`func mkbool_arity_1(x flagrt.Value) flagrt.Value {`,
		`return flagrt.NewBool(func() bool {`,
		`return flagrt.Lt(x, flagrt.NewLong(3))`,
		`var mklist = flagrt.NewFunction(mklist_variadic)`,
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
