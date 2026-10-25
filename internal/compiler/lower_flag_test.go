package compiler

import (
	"strings"
	"testing"

	flagrt "flag-lang/runtime"
)

func TestFLAGLiteralLowering(t *testing.T) {
	cases := []struct {
		name string
		expr Expr
		want string
	}{
		{"string", StringExpr{Value: "hi"}, `"hi"`},
		{"char", CharExpr{Value: 'x'}, `flagrt.NewString("x")`},
		{"int", IntExpr{Value: 42}, "flagrt.NewLong(42)"},
		{"negative-int", IntExpr{Value: -7}, "flagrt.NewLong(-7)"},
		{"bigint", BigIntExpr{Value: "10"}, `flagrt.NewBigIntFromString("10")`},
		{"ratio", RatioExpr{Numerator: 5, Denominator: 6}, "flagrt.NewRatio(5, 6)"},
		{"float-raw", FloatExpr{Value: 1.5, Raw: "1.5"}, "flagrt.NewDouble(1.5)"},
		{"float-exp", FloatExpr{Value: 100, Raw: "1e2"}, "flagrt.NewDouble(1e2)"},
		{"keyword", KeywordExpr{Name: "xyz"}, `flagrt.NewKeyword("xyz")`},
		{"quoted-symbol", QuotedSymbolExpr{Name: "abc"}, `flagrt.NewSymbol("abc")`},
		{"true", SymbolExpr{Name: "true"}, "flagrt.NewBool(true)"},
		{"false", SymbolExpr{Name: "false"}, "flagrt.NewBool(false)"},
		{"nil", SymbolExpr{Name: "nil"}, "flagrt.NilValue"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ir, err := flagLiteralToIR(tc.expr)
			if err != nil {
				t.Fatal(err)
			}
			got := renderIRExpr(ir)
			if got != tc.want {
				t.Fatalf("got %q want %q", got, tc.want)
			}
		})
	}
}

func TestFLAGNilLiteralIsSelector(t *testing.T) {
	ir, err := flagLiteralToIR(SymbolExpr{Name: "nil"})
	if err != nil {
		t.Fatal(err)
	}
	sel, ok := ir.(IRSelector)
	if !ok {
		t.Fatalf("nil IR %T, want IRSelector", ir)
	}
	if sel.Pkg != "flagrt" || sel.Name != "NilValue" {
		t.Fatalf("nil selector %+v", sel)
	}
	if _, isCall := ir.(IRCall); isCall {
		t.Fatal("nil must not be IRCall")
	}
}

func TestLiteralExprToGoInternsKeywords(t *testing.T) {
	ctx := compileContext{constants: newConstantInterner()}
	got, err := literalExprToGo(KeywordExpr{Name: "foo"}, ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got.code != "flagKw_foo" {
		t.Fatalf("code %q", got.code)
	}
}

func TestFLAGCompileFormInternsLiterals(t *testing.T) {
	n := 0
	ctx := compileContext{ifTemps: &n, constants: newConstantInterner()}
	got, err := flagCompileForm(KeywordExpr{Name: "foo"}, ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got.code != "flagKw_foo" {
		t.Fatalf("keyword %q", got.code)
	}
	got, err = flagCompileForm(QuotedSymbolExpr{Name: "bar"}, ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got.code != "flagSym_bar" {
		t.Fatalf("symbol %q", got.code)
	}
	got, err = flagCompileForm(QuotedListExpr{Elements: []Expr{IntExpr{Value: 1}, KeywordExpr{Name: "a"}}}, ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got.code != "flagList" {
		t.Fatalf("quoted list %q", got.code)
	}
	decls := ctx.constants.decls()
	if len(decls) != 3 {
		t.Fatalf("decls %d %+v", len(decls), decls)
	}
}

func TestFLAGCompileModuleBodyInternsKeyword(t *testing.T) {
	ast, err := ParseFile("(defn kw [] :foo)")
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	ctx := compileContext{
		namespace:     "compiler",
		ifTemps:       &n,
		constants:     newConstantInterner(),
		functions:     map[string]functionDef{},
		globals:       map[string]exprKind{},
		moduleSymbols: map[string]string{},
	}
	_, _, err = flagCompileModuleBody(ast.Forms, &ctx, false)
	if err != nil {
		t.Fatal(err)
	}
	decls := ctx.constants.decls()
	if len(decls) != 1 || decls[0].goName != "flagKw_foo" {
		t.Fatalf("decls %+v", decls)
	}
}

func TestFLAGQuotedLowering(t *testing.T) {
	cases := []struct {
		name string
		expr Expr
		want string
	}{
		{
			"quoted-list",
			QuotedListExpr{Elements: []Expr{IntExpr{Value: 1}, KeywordExpr{Name: "a"}}},
			`flagrt.NewList(flagrt.NewLong(1), flagrt.NewKeyword("a"))`,
		},
		{
			"empty-quoted-list",
			QuotedListExpr{},
			`flagrt.NewList()`,
		},
		{
			"quoted-true-is-symbol",
			QuotedListExpr{Elements: []Expr{SymbolExpr{Name: "true"}}},
			`flagrt.NewList(flagrt.NewSymbol("true"))`,
		},
		{
			"nested-vector",
			VectorExpr{Elements: []Expr{
				QuotedListExpr{Elements: []Expr{SymbolExpr{Name: "x"}}},
			}},
			`flagrt.NewArray(flagrt.NewList(flagrt.NewSymbol("x")))`,
		},
		{
			"pipe-vector",
			PipeVectorExpr{Elements: []Expr{IntExpr{Value: 1}}},
			`flagrt.NewVector(flagrt.NewLong(1))`,
		},
		{
			"set",
			SetExpr{Elements: []Expr{KeywordExpr{Name: "a"}}},
			`flagrt.NewSet(flagrt.NewKeyword("a"))`,
		},
		{
			"map",
			MapExpr{Entries: []Expr{KeywordExpr{Name: "a"}, IntExpr{Value: 1}}},
			`flagrt.NewMap(flagrt.NewKeyword("a"), flagrt.NewLong(1))`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ir, err := flagQuotedToIR(tc.expr)
			if err != nil {
				t.Fatal(err)
			}
			got := renderIRExpr(ir)
			if got != tc.want {
				t.Fatalf("got %q want %q", got, tc.want)
			}
		})
	}
}

func TestFLAGCallToIR(t *testing.T) {
	ir, err := flagCallToIR(IRIdent{Name: "f"}, []IRExpr{rtCall("NewLong", IRInt{Value: 1})})
	if err != nil {
		t.Fatal(err)
	}
	got := renderIRExpr(ir)
	want := "flagrt.Call(f, flagrt.NewLong(1))"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestFLAGCallAstToIR(t *testing.T) {
	cases := []struct {
		name string
		expr Expr
		want string
	}{
		{
			"keyword-call",
			ListExpr{Elements: []Expr{KeywordExpr{Name: "a"}, IntExpr{Value: 1}}},
			`flagrt.Call(flagrt.NewKeyword("a"), flagrt.NewLong(1))`,
		},
		{
			"no-args",
			ListExpr{Elements: []Expr{KeywordExpr{Name: "a"}}},
			`flagrt.Call(flagrt.NewKeyword("a"))`,
		},
		{
			"nested",
			ListExpr{Elements: []Expr{
				KeywordExpr{Name: "a"},
				ListExpr{Elements: []Expr{KeywordExpr{Name: "b"}, IntExpr{Value: 1}}},
			}},
			`flagrt.Call(flagrt.NewKeyword("a"), flagrt.Call(flagrt.NewKeyword("b"), flagrt.NewLong(1)))`,
		},
		{
			"vector-arg",
			ListExpr{Elements: []Expr{
				KeywordExpr{Name: "a"},
				VectorExpr{Elements: []Expr{IntExpr{Value: 1}}},
			}},
			`flagrt.Call(flagrt.NewKeyword("a"), flagrt.NewArray(flagrt.NewLong(1)))`,
		},
		{
			"quoted-list-arg",
			ListExpr{Elements: []Expr{
				KeywordExpr{Name: "a"},
				QuotedListExpr{Elements: []Expr{IntExpr{Value: 1}}},
			}},
			`flagrt.Call(flagrt.NewKeyword("a"), flagrt.NewList(flagrt.NewLong(1)))`,
		},
		{
			"true-arg",
			ListExpr{Elements: []Expr{KeywordExpr{Name: "a"}, SymbolExpr{Name: "true"}}},
			`flagrt.Call(flagrt.NewKeyword("a"), flagrt.NewBool(true))`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ir, err := flagCallAstToIR(tc.expr)
			if err != nil {
				t.Fatal(err)
			}
			got := renderIRExpr(ir)
			if got != tc.want {
				t.Fatalf("got %q want %q", got, tc.want)
			}
		})
	}
}

func TestFLAGCallAstToIRErrors(t *testing.T) {
	cases := []Expr{
		ListExpr{},
		ListExpr{Elements: []Expr{SymbolExpr{Name: "if"}, SymbolExpr{Name: "true"}}},
		ListExpr{Elements: []Expr{SymbolExpr{Name: "foo"}, IntExpr{Value: 1}}},
	}
	for _, expr := range cases {
		if _, err := flagCallAstToIR(expr); err == nil {
			t.Fatalf("expected error for %#v", expr)
		}
	}
}

func TestExprToGoUsesFLAGCalls(t *testing.T) {
	ctx := compileContext{}
	locals := map[string]exprKind{"f": exprKindValue}
	got, err := exprToGo(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "f"},
		IntExpr{Value: 1},
		KeywordExpr{Name: "a"},
	}}, ctx, locals)
	if err != nil {
		t.Fatal(err)
	}
	want := `flagrt.Call(f, flagrt.NewLong(1), flagrt.NewKeyword("a"))`
	if got.code != want {
		t.Fatalf("code %q want %q", got.code, want)
	}
	if got.kind != exprKindValue {
		t.Fatalf("kind %v", got.kind)
	}
}

func TestExprToGoSelfCallStaysDirect(t *testing.T) {
	ctx := compileContext{
		selfFunctionName: "foo",
		selfArityNames:   map[int]string{1: "flagFoo_1"},
	}
	got, err := exprToGo(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "foo"},
		IntExpr{Value: 1},
	}}, ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	want := "flagFoo_1(flagrt.NewLong(1))"
	if got.code != want {
		t.Fatalf("code %q want %q", got.code, want)
	}
}

func TestFLAGQuotedLoweringOddMap(t *testing.T) {
	_, err := flagQuotedToIR(MapExpr{Entries: []Expr{KeywordExpr{Name: "a"}}})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestFLAGCtorAndRuntimeOps(t *testing.T) {
	ir, err := flagCtorToIR("NewArray", []IRExpr{rtCall("NewLong", IRInt{Value: 1})})
	if err != nil {
		t.Fatal(err)
	}
	if got := renderIRExpr(ir); got != "flagrt.NewArray(flagrt.NewLong(1))" {
		t.Fatalf("ctor %q", got)
	}
	ir, err = flagRuntimeCallToIR("First", []IRExpr{IRIdent{Name: "xs"}})
	if err != nil {
		t.Fatal(err)
	}
	if got := renderIRExpr(ir); got != "flagrt.First(xs)" {
		t.Fatalf("runtime %q", got)
	}
	ir, err = flagFoldCallToIR("Add", []IRExpr{
		rtCall("NewLong", IRInt{Value: 1}),
		rtCall("NewLong", IRInt{Value: 2}),
		rtCall("NewLong", IRInt{Value: 3}),
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := renderIRExpr(ir); got != "flagrt.Add(flagrt.Add(flagrt.NewLong(1), flagrt.NewLong(2)), flagrt.NewLong(3))" {
		t.Fatalf("fold %q", got)
	}
	ir, err = flagEqToIR([]IRExpr{
		rtCall("NewLong", IRInt{Value: 1}),
		rtCall("NewLong", IRInt{Value: 2}),
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := renderIRExpr(ir); got != "flagrt.Eq(flagrt.NewLong(1), flagrt.NewLong(2))" {
		t.Fatalf("eq %q", got)
	}
	ir, err = flagEqToIR([]IRExpr{IRIdent{Name: "a"}, IRIdent{Name: "b"}, IRIdent{Name: "c"}})
	if err != nil {
		t.Fatal(err)
	}
	if got := renderIRExpr(ir); got != "flagrt.Eq(a, b) && flagrt.Eq(b, c)" {
		t.Fatalf("eq-chain %q", got)
	}
	_, err = flagEqToIR([]IRExpr{IRIdent{Name: "a"}})
	if err == nil {
		t.Fatal("expected arity error")
	}
}

func TestFLAGSymbolToIR(t *testing.T) {
	ctx := compileContext{}
	got, err := flagSymbolToGoExpr("x", "x", ctx, map[string]exprKind{"x": exprKindValue})
	if err != nil {
		t.Fatal(err)
	}
	if got.code != "x" || got.kind != exprKindValue {
		t.Fatalf("local %#v", got)
	}
	got, err = flagSymbolToGoExpr("first", "first", ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got.code != `flagrt.NewFunction(flagrt.FnFirst)` {
		t.Fatalf("builtin %q", got.code)
	}
	if _, err := flagSymbolToGoExpr("nope", "nope", ctx, nil); err == nil {
		t.Fatal("expected unknown symbol")
	}
}

func TestFLAGEvalAstCollectionsAndOps(t *testing.T) {
	ctx := flagrt.NewMap(
		flagrt.NewKeyword("locals"),
		flagrt.NewMap(flagrt.NewString("x"), flagrt.NewKeyword("value")),
	)
	ir, err := flagEvalAstToIR(VectorExpr{Elements: []Expr{SymbolExpr{Name: "x"}, IntExpr{Value: 1}}}, ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got := renderIRExpr(ir); got != "flagrt.NewArray(x, flagrt.NewLong(1))" {
		t.Fatalf("vector %q", got)
	}
	ir, err = flagEvalAstToIR(ListExpr{Elements: []Expr{SymbolExpr{Name: "first"}, SymbolExpr{Name: "x"}}}, ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got := renderIRExpr(ir); got != "flagrt.First(x)" {
		t.Fatalf("first %q", got)
	}
	ir, err = flagEvalAstToIR(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "+"}, IntExpr{Value: 1}, IntExpr{Value: 2}, IntExpr{Value: 3},
	}}, flagrt.NewMap())
	if err != nil {
		t.Fatal(err)
	}
	if got := renderIRExpr(ir); got != "flagrt.Add(flagrt.Add(flagrt.NewLong(1), flagrt.NewLong(2)), flagrt.NewLong(3))" {
		t.Fatalf("add %q", got)
	}
}

func TestExprToGoUsesFLAGCollectionsAndSymbols(t *testing.T) {
	ctx := compileContext{}
	locals := map[string]exprKind{"x": exprKindValue}
	got, err := exprToGo(VectorExpr{Elements: []Expr{SymbolExpr{Name: "x"}, IntExpr{Value: 1}}}, ctx, locals)
	if err != nil {
		t.Fatal(err)
	}
	if got.code != "flagrt.NewArray(x, flagrt.NewLong(1))" {
		t.Fatalf("vector %q", got.code)
	}
	got, err = exprToGo(ListExpr{Elements: []Expr{SymbolExpr{Name: "first"}, SymbolExpr{Name: "x"}}}, ctx, locals)
	if err != nil {
		t.Fatal(err)
	}
	if got.code != "flagrt.First(x)" {
		t.Fatalf("first %q", got.code)
	}
	got, err = exprToGo(SymbolExpr{Name: "x"}, ctx, locals)
	if err != nil {
		t.Fatal(err)
	}
	if got.code != "x" {
		t.Fatalf("symbol %q", got.code)
	}
}

func TestExprToGoUsesFLAGLiterals(t *testing.T) {
	ctx := compileContext{}
	got, err := exprToGo(IntExpr{Value: 9}, ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got.code != "flagrt.NewLong(9)" {
		t.Fatalf("code %q", got.code)
	}
	if got.kind != exprKindValue {
		t.Fatalf("kind %v", got.kind)
	}
}

func TestFLAGIfToIR(t *testing.T) {
	stmts, err := flagIfToIR(
		"if_result_1",
		"flagrt.Value",
		IRIdent{Name: "ok"},
		nil,
		rtCall("NewLong", IRInt{Value: 1}),
		nil,
		rtCall("NewLong", IRInt{Value: 2}),
	)
	if err != nil {
		t.Fatal(err)
	}
	got := renderIRStmts(stmts, "\t")
	want := "\tvar if_result_1 flagrt.Value\n\tif ok {\n\t\tif_result_1 = flagrt.NewLong(1)\n\t} else {\n\t\tif_result_1 = flagrt.NewLong(2)\n\t}\n"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestIfExprToGoUsesResultVar(t *testing.T) {
	n := 0
	ctx := compileContext{ifTemps: &n}
	locals := map[string]exprKind{"x": exprKindValue}
	got, err := exprToGo(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "if"},
		SymbolExpr{Name: "x"},
		IntExpr{Value: 1},
		IntExpr{Value: 2},
	}}, ctx, locals)
	if err != nil {
		t.Fatal(err)
	}
	if got.code != "if_result_1" {
		t.Fatalf("code %q", got.code)
	}
	rendered := renderIRStmts(got.stmts, "\t")
	if strings.Contains(rendered, "func()") {
		t.Fatalf("if still used an IIFE:\n%s", rendered)
	}
	if !strings.Contains(rendered, "var if_result_1 flagrt.Value") {
		t.Fatalf("missing result var:\n%s", rendered)
	}
	if !strings.Contains(rendered, "if flagrt.IsTruthy(x)") {
		t.Fatalf("missing cond:\n%s", rendered)
	}
}

func TestFLAGLetToIR(t *testing.T) {
	stmts, err := flagLetToIR(
		"let_result_1",
		"flagrt.Value",
		[]IRStmt{IRVar{Name: "x", Expr: rtCall("NewLong", IRInt{Value: 1})}},
		[]IRStmt{IRExprStmt{Expr: IRIdent{Name: "x"}, Discard: true}},
		IRIdent{Name: "x"},
	)
	if err != nil {
		t.Fatal(err)
	}
	got := renderIRStmts(stmts, "\t")
	want := "\tvar let_result_1 flagrt.Value\n\t{\n\t\tvar x = flagrt.NewLong(1)\n\t\t_ = x\n\t\tlet_result_1 = x\n\t}\n"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestFLAGLetToIREmptyStmts(t *testing.T) {
	stmts, err := flagLetToIR("let_result_1", "flagrt.Value", nil, nil, rtCall("NewLong", IRInt{Value: 1}))
	if err != nil {
		t.Fatal(err)
	}
	got := renderIRStmts(stmts, "\t")
	want := "\tvar let_result_1 flagrt.Value\n\t{\n\t\tlet_result_1 = flagrt.NewLong(1)\n\t}\n"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestFLAGDoPreludeStmts(t *testing.T) {
	stmts, err := flagDoPreludeStmts([]goExpr{
		fromIR(IRIdent{Name: "a"}, exprKindValue),
		fromIR(IRIdent{Name: "b"}, exprKindValue),
	})
	if err != nil {
		t.Fatal(err)
	}
	got := renderIRStmts(stmts, "\t")
	want := "\t_ = a\n"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestFLAGDoDeferToIR(t *testing.T) {
	deferForm := fromStmt(IRDefer{Expr: rtCall("Call", IRIdent{Name: "f"})}, exprKindDefer)
	ir, err := flagDoDeferToIR("flagrt.Value", []goExpr{
		fromIR(IRIdent{Name: "a"}, exprKindValue),
		deferForm,
	})
	if err != nil {
		t.Fatal(err)
	}
	got := renderIRExpr(ir)
	if !strings.Contains(got, "func() flagrt.Value") {
		t.Fatalf("missing IIFE:\n%s", got)
	}
	if !strings.Contains(got, "_ = a") {
		t.Fatalf("missing prelude:\n%s", got)
	}
	if !strings.Contains(got, "defer flagrt.Call(f)") {
		t.Fatalf("missing defer:\n%s", got)
	}
	if !strings.Contains(got, "return flagrt.NilValue") {
		t.Fatalf("missing nil return:\n%s", got)
	}
}

func TestDoExprToGoFlattens(t *testing.T) {
	ctx := compileContext{}
	locals := map[string]exprKind{"a": exprKindValue, "b": exprKindValue}
	got, err := exprToGo(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "do"},
		SymbolExpr{Name: "a"},
		SymbolExpr{Name: "b"},
	}}, ctx, locals)
	if err != nil {
		t.Fatal(err)
	}
	if got.code != "b" {
		t.Fatalf("code %q", got.code)
	}
	rendered := renderIRStmts(got.stmts, "\t")
	if strings.Contains(rendered, "func()") {
		t.Fatalf("do used an IIFE:\n%s", rendered)
	}
	if rendered != "\t_ = a\n" {
		t.Fatalf("prelude %q", rendered)
	}
}

func TestFLAGRecurToIR(t *testing.T) {
	ir, err := flagRecurToIR([]IRExpr{IRIdent{Name: "x"}, rtCall("NewLong", IRInt{Value: 1})})
	if err != nil {
		t.Fatal(err)
	}
	got := renderIRExpr(ir)
	want := "flagrt.NewRecur(x, flagrt.NewLong(1))"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestFLAGLoopToIR(t *testing.T) {
	ir, err := flagLoopToIR(
		nil,
		[]loopBindingIR{{Name: "n", Init: rtCall("NewLong", IRInt{Value: 1})}},
		nil,
		IRIdent{Name: "n"},
	)
	if err != nil {
		t.Fatal(err)
	}
	got := renderIRExpr(ir)
	for _, want := range []string{
		"func() flagrt.Value",
		"var n = flagrt.NewLong(1)",
		"for {",
		"__loopResult := n",
		"UnwrapRecur(__loopResult)",
		"n = __recurValues[0]",
		"continue",
		"return __loopResult",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in:\n%s", want, got)
		}
	}
}

func TestLoopExprToGoUsesFLAG(t *testing.T) {
	ctx := compileContext{}
	got, err := exprToGo(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "loop"},
		VectorExpr{Elements: []Expr{SymbolExpr{Name: "n"}, IntExpr{Value: 0}}},
		ListExpr{Elements: []Expr{
			SymbolExpr{Name: "if"},
			ListExpr{Elements: []Expr{SymbolExpr{Name: "="}, SymbolExpr{Name: "n"}, IntExpr{Value: 2}}},
			SymbolExpr{Name: "n"},
			ListExpr{Elements: []Expr{SymbolExpr{Name: "recur"}, ListExpr{Elements: []Expr{SymbolExpr{Name: "+"}, SymbolExpr{Name: "n"}, IntExpr{Value: 1}}}}},
		}},
	}}, ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got.code, "UnwrapRecur") {
		t.Fatalf("loop IIFE missing unwrap:\n%s", got.code)
	}
	if !strings.Contains(got.code, "flagrt.NewRecur") {
		t.Fatalf("missing recur:\n%s", got.code)
	}
}

func TestFLAGDeferToIR(t *testing.T) {
	stmt, err := flagDeferToIR(IRIdent{Name: "f"})
	if err != nil {
		t.Fatal(err)
	}
	got := renderIRStmt(stmt, "\t")
	want := "\tdefer flagrt.Call(f)\n"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestFLAGThrowToIR(t *testing.T) {
	ir, err := flagThrowToIR(nil, IRIdent{Name: "err"})
	if err != nil {
		t.Fatal(err)
	}
	got := renderIRExpr(ir)
	want := "func() flagrt.Value {\n\tflagrt.Throw(err)\n\treturn flagrt.NilValue\n}()"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestFLAGFutureToIR(t *testing.T) {
	ir, err := flagFutureToIR(nil, IRIdent{Name: "x"})
	if err != nil {
		t.Fatal(err)
	}
	got := renderIRExpr(ir)
	want := "flagrt.NewFuture(func() flagrt.Value {\n\treturn x\n})"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestThrowExprToGoUsesFLAG(t *testing.T) {
	ctx := compileContext{}
	locals := map[string]exprKind{"err": exprKindValue}
	got, err := exprToGo(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "throw"},
		SymbolExpr{Name: "err"},
	}}, ctx, locals)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got.code, "flagrt.Throw(err)") {
		t.Fatalf("missing throw:\n%s", got.code)
	}
}

func TestFLAGDotoToIR(t *testing.T) {
	ir, err := flagDotoToIR(nil, IRIdent{Name: "obj"}, []goExpr{fromIR(IRIdent{Name: "step"}, exprKindValue)})
	if err != nil {
		t.Fatal(err)
	}
	got := renderIRExpr(ir)
	for _, want := range []string{
		"__doto := obj",
		"flagrt.Call(step, __doto)",
		"return __doto",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in:\n%s", want, got)
		}
	}
}

func TestFLAGUpdateBangToIR(t *testing.T) {
	ir, err := flagUpdateBangToIR("n", nil, rtCall("NewLong", IRInt{Value: 1}))
	if err != nil {
		t.Fatal(err)
	}
	got := renderIRExpr(ir)
	want := "func() flagrt.Value {\n\tn = flagrt.NewLong(1)\n\treturn n\n}()"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestFLAGCatchHandlerToIR(t *testing.T) {
	ir, err := flagCatchHandlerToIR("e", nil, IRIdent{Name: "e"})
	if err != nil {
		t.Fatal(err)
	}
	got := renderIRExpr(ir)
	for _, want := range []string{
		"var e = __flag_thrown",
		"_ = e",
		"return e",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in:\n%s", want, got)
		}
	}
}

func TestFLAGTryToIRFinallyOnly(t *testing.T) {
	finally := fromIR(IRIdent{Name: "cleanup"}, exprKindValue)
	ir, err := flagTryToIR(nil, IRIdent{Name: "body"}, nil, &finally)
	if err != nil {
		t.Fatal(err)
	}
	got := renderIRExpr(ir)
	for _, want := range []string{
		"defer func() {",
		"_ = cleanup",
		"return body",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in:\n%s", want, got)
		}
	}
	if strings.Contains(got, "__flag_try_result") {
		t.Fatalf("finally-only try should not use result var:\n%s", got)
	}
}

func TestFLAGTryToIRWithCatch(t *testing.T) {
	handler, err := flagCatchHandlerToIR("e", nil, IRIdent{Name: "e"})
	if err != nil {
		t.Fatal(err)
	}
	ir, err := flagTryToIR(nil, IRIdent{Name: "body"}, []tryCatchIR{{
		Class:   "Exception",
		Handler: handler,
	}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	got := renderIRExpr(ir)
	for _, want := range []string{
		"var __flag_try_result flagrt.Value",
		"r := recover()",
		"flagrt.CatchMatches(\"Exception\", __flag_thrown)",
		"__flag_try_result = body",
		"return __flag_try_result",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in:\n%s", want, got)
		}
	}
}

func TestFLAGDoseqBodyToIR(t *testing.T) {
	ir, err := flagDoseqBodyToIR(nil, IRIdent{Name: "x"})
	if err != nil {
		t.Fatal(err)
	}
	got := renderIRExpr(ir)
	want := "func() flagrt.Value {\n\t_ = x\n\treturn flagrt.NewArray()\n}()"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestFLAGDoseqToIR(t *testing.T) {
	ir, err := flagDoseqToIR(IRIdent{Name: "loop"})
	if err != nil {
		t.Fatal(err)
	}
	got := renderIRExpr(ir)
	want := "func() flagrt.Value {\n\t_ = flagrt.DoAll(loop)\n\treturn flagrt.NilValue\n}()"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestFLAGMapCatBindingToIR(t *testing.T) {
	ir, err := flagMapCatBindingToIR("x", false, "for binding expects exactly one value", IRIdent{Name: "rest"}, nil, IRIdent{Name: "xs"})
	if err != nil {
		t.Fatal(err)
	}
	got := renderIRExpr(ir)
	for _, want := range []string{
		"flagrt.MapCat",
		"x := args[0]",
		"return rest",
		"flagrt.NewFunction",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in:\n%s", want, got)
		}
	}
}

func TestFLAGFnToIR(t *testing.T) {
	ir, err := flagFnToIR("fn", []string{"x"}, false, nil, nil, IRIdent{Name: "x"})
	if err != nil {
		t.Fatal(err)
	}
	got := renderIRExpr(ir)
	for _, want := range []string{
		"flagrt.NewFunction",
		"func(args ...flagrt.Value) flagrt.Value",
		"if len(args) != 1",
		`panic("fn expects exactly 1 arguments")`,
		"x := args[0]",
		"return x",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in:\n%s", want, got)
		}
	}
}

func TestFLAGFnToIRRestAndDiscard(t *testing.T) {
	ir, err := flagFnToIR("fn", []string{"_"}, true, []IRStmt{
		IRVar{Name: "__rest0", Expr: rtCall("NewArray", IRSpread{Expr: IRSlice{X: IRIdent{Name: "args"}, Low: IRInt{Value: 1}}})},
		IRExprStmt{Expr: IRIdent{Name: "__rest0"}, Discard: true},
	}, nil, IRIdent{Name: "__rest0"})
	if err != nil {
		t.Fatal(err)
	}
	got := renderIRExpr(ir)
	for _, want := range []string{
		"if len(args) < 1",
		`panic("fn expects at least 1 arguments")`,
		"_ = args[0]",
		"var __rest0 = flagrt.NewArray(args[1:]...)",
		"_ = __rest0",
		"return __rest0",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in:\n%s", want, got)
		}
	}
}

func TestFLAGFnToIRZeroArityWithBodyStmts(t *testing.T) {
	ir, err := flagFnToIR("#()", nil, false, nil, []IRStmt{
		IRVar{Name: "let_result_1", Type: "flagrt.Value"},
		IRAssign{Name: "let_result_1", Expr: IRIdent{Name: "flagrt.NewLong(1)"}},
	}, IRIdent{Name: "let_result_1"})
	if err != nil {
		t.Fatal(err)
	}
	got := renderIRExpr(ir)
	for _, want := range []string{
		"if len(args) != 0",
		`panic("#() expects exactly 0 arguments")`,
		"var let_result_1 flagrt.Value",
		"return let_result_1",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in:\n%s", want, got)
		}
	}
}

func TestFnExprToGoUsesFLAGNewFunction(t *testing.T) {
	got, err := exprToGo(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "fn"},
		VectorExpr{Elements: []Expr{SymbolExpr{Name: "x"}}},
		SymbolExpr{Name: "x"},
	}}, compileContext{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	src := renderIRExpr(irFromGoExpr(got))
	for _, want := range []string{
		"flagrt.NewFunction",
		"x := args[0]",
		`panic("fn expects exactly 1 arguments")`,
		"return x",
	} {
		if !strings.Contains(src, want) {
			t.Fatalf("missing %q in:\n%s", want, src)
		}
	}
}

func TestFLAGNsDefDefnToIR(t *testing.T) {
	ns, err := flagNsToIR("hello.core")
	if err != nil {
		t.Fatal(err)
	}
	if got := renderIRStmt(ns, ""); got != "// Source namespace: hello.core\n" {
		t.Fatalf("ns %q", got)
	}

	defStmt, err := flagDefToIR("x", rtCall("NewLong", IRInt{Value: 1}))
	if err != nil {
		t.Fatal(err)
	}
	if got := renderIRStmt(defStmt, ""); got != "var x = flagrt.NewLong(1)\n" {
		t.Fatalf("def %q", got)
	}

	bind, err := flagDefnBindingToIR("foo", "foo_variadic")
	if err != nil {
		t.Fatal(err)
	}
	if got := renderIRStmt(bind, ""); got != "var foo = flagrt.NewFunction(foo_variadic)\n" {
		t.Fatalf("binding %q", got)
	}

	decls, err := flagDefnToIR("foo_arity_1", "foo_variadic", "foo", []string{"x"}, false, nil, nil, IRIdent{Name: "x"})
	if err != nil {
		t.Fatal(err)
	}
	got := renderIRStmts(decls, "")
	for _, want := range []string{
		"func foo_arity_1(x flagrt.Value) flagrt.Value",
		"return x",
		"func foo_variadic(args ...flagrt.Value) flagrt.Value",
		`panic("foo expects exactly 1 arguments")`,
		"return foo_arity_1(args[0])",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in:\n%s", want, got)
		}
	}

	rest, err := flagDefnToIR("ignored", "foo_variadic", "foo", []string{"x"}, true, nil, nil, IRIdent{Name: "x"})
	if err != nil {
		t.Fatal(err)
	}
	restSrc := renderIRStmts(rest, "")
	for _, want := range []string{
		"func foo_variadic(args ...flagrt.Value) flagrt.Value",
		`panic("foo expects at least 1 arguments")`,
		"x := args[0]",
		"return x",
	} {
		if !strings.Contains(restSrc, want) {
			t.Fatalf("missing %q in rest:\n%s", want, restSrc)
		}
	}
	if strings.Contains(restSrc, "foo_arity_1") {
		t.Fatalf("rest defn should not emit arity func:\n%s", restSrc)
	}

	multi, err := flagDefnMultiToIR("add_variadic", "add expects 1 or 2 arguments", []defnArityIR{
		{Name: "add_arity_1", Params: []string{"x"}, BodyExpr: IRIdent{Name: "x"}},
		{Name: "add_arity_2", Params: []string{"x", "y"}, BodyExpr: IRIdent{Name: "x"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	multiSrc := renderIRStmts(multi, "")
	for _, want := range []string{
		"func add_arity_1(x flagrt.Value) flagrt.Value",
		"func add_arity_2(x flagrt.Value, y flagrt.Value) flagrt.Value",
		"len(args) == 1",
		"return add_arity_1(args[0])",
		"len(args) == 2",
		"return add_arity_2(args[0], args[1])",
		`panic("add expects 1 or 2 arguments")`,
	} {
		if !strings.Contains(multiSrc, want) {
			t.Fatalf("missing %q in multi:\n%s", want, multiSrc)
		}
	}
}

func TestHashFnExprToGoUsesFLAGNewFunction(t *testing.T) {
	got, err := exprToGo(HashFnExpr{Body: SymbolExpr{Name: "%"}}, compileContext{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	src := renderIRExpr(irFromGoExpr(got))
	for _, want := range []string{
		"flagrt.NewFunction",
		"__p1 := args[0]",
		`panic("#() expects exactly 1 arguments")`,
		"return __p1",
	} {
		if !strings.Contains(src, want) {
			t.Fatalf("missing %q in:\n%s", want, src)
		}
	}
}

func TestFLAGGoIdent(t *testing.T) {
	cases := []struct {
		name string
		want string
	}{
		{"foo", "foo"},
		{"foo-bar", "foo_bar"},
		{"foo?", "foo_q"},
		{"set!", "set_bang"},
		{"pred>", "pred_gt"},
		{"pred<", "pred_lt"},
		{"main", "flag_main"},
		{"var", "var_"},
		{"type", "type_"},
		{"_x", "_x"},
		{"a1", "a1"},
		{"->", "__gt"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := flagGoIdent(tc.name)
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Fatalf("got %q want %q", got, tc.want)
			}
		})
	}

	if _, err := flagGoIdent(""); err == nil {
		t.Fatal("expected empty symbol error")
	}
	if _, err := flagGoIdent("1abc"); err == nil {
		t.Fatal("expected unsupported symbol error")
	}
	if _, err := flagGoIdent("foo.bar"); err == nil {
		t.Fatal("expected unsupported symbol error")
	}
}

func TestFLAGBindParams(t *testing.T) {
	params, kinds, stmts, hasRest, err := flagBindParams(
		VectorExpr{Elements: []Expr{SymbolExpr{Name: "x"}, SymbolExpr{Name: "y?"}}},
		"fn",
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	if hasRest {
		t.Fatal("expected no rest")
	}
	if len(params) != 2 || params[0] != "x" || params[1] != "y_q" {
		t.Fatalf("params %v", params)
	}
	if kinds["x"] != exprKindValue || kinds["y_q"] != exprKindValue {
		t.Fatalf("kinds %v", kinds)
	}
	if len(stmts) != 0 {
		t.Fatalf("unexpected init stmts %#v", stmts)
	}

	params, kinds, stmts, hasRest, err = flagBindParams(
		VectorExpr{Elements: []Expr{SymbolExpr{Name: "_"}}},
		"fn",
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(params) != 1 || params[0] != "_" || hasRest {
		t.Fatalf("unused params %v hasRest %v", params, hasRest)
	}
	if _, ok := kinds["_"]; ok {
		t.Fatalf("underscore should not be a local: %v", kinds)
	}

	params, kinds, stmts, hasRest, err = flagBindParams(
		VectorExpr{Elements: []Expr{
			SymbolExpr{Name: "x"},
			SymbolExpr{Name: "&"},
			SymbolExpr{Name: "more"},
		}},
		"fn",
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	if !hasRest || len(params) != 1 || params[0] != "x" {
		t.Fatalf("rest params %v hasRest %v", params, hasRest)
	}
	if kinds["more"] != exprKindValue || kinds["x"] != exprKindValue {
		t.Fatalf("rest kinds %v", kinds)
	}
	src := renderIRStmts(stmts, "")
	for _, want := range []string{
		"__rest0",
		"flagrt.NewArray(args[1:]...)",
		"var more = __rest0",
	} {
		if !strings.Contains(src, want) {
			t.Fatalf("missing %q in rest stmts:\n%s", want, src)
		}
	}

	_, _, _, _, err = flagBindParams(
		VectorExpr{Elements: []Expr{SymbolExpr{Name: "x"}, SymbolExpr{Name: "x"}}},
		"fn",
		nil,
	)
	if err == nil {
		t.Fatal("expected duplicate parameter error")
	}

	mapParams, kinds, stmts, hasRest, err := flagBindParams(
		VectorExpr{Elements: []Expr{MapExpr{Entries: []Expr{KeywordExpr{Name: "a"}, SymbolExpr{Name: "a"}}}}},
		"fn",
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	if hasRest || len(mapParams) != 1 || mapParams[0] != "__arg0" {
		t.Fatalf("map params %v hasRest %v", mapParams, hasRest)
	}
	if kinds["a"] != exprKindValue {
		t.Fatalf("map kinds %v", kinds)
	}
	src = renderIRStmts(stmts, "")
	for _, want := range []string{
		"var a = flagrt.Get(__arg0, flagrt.NewKeyword(\"a\"))",
	} {
		if !strings.Contains(src, want) {
			t.Fatalf("missing %q in map stmts:\n%s", want, src)
		}
	}

	vec, kinds, stmts, hasRest, err := flagBindParams(
		VectorExpr{Elements: []Expr{
			VectorExpr{Elements: []Expr{SymbolExpr{Name: "a"}, SymbolExpr{Name: "b"}}},
		}},
		"fn",
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	if hasRest || len(vec) != 1 || vec[0] != "__arg0" {
		t.Fatalf("vector params %v hasRest %v", vec, hasRest)
	}
	if kinds["a"] != exprKindValue || kinds["b"] != exprKindValue {
		t.Fatalf("vector kinds %v", kinds)
	}
	src = renderIRStmts(stmts, "")
	for _, want := range []string{
		"__dseq0",
		"flagrt.SeqFirst(__arg0)",
		"var a = __dseq0",
		"flagrt.SeqRest(__arg0)",
		"var b = __dseq1",
	} {
		if !strings.Contains(src, want) {
			t.Fatalf("missing %q in vector stmts:\n%s", want, src)
		}
	}
}

func TestFLAGBindParamsPreservesMutableLocal(t *testing.T) {
	_, kinds, _, _, err := flagBindParams(
		VectorExpr{Elements: []Expr{SymbolExpr{Name: "_n"}}},
		"fn",
		map[string]exprKind{"v": exprKindMutableValue},
	)
	if err != nil {
		t.Fatal(err)
	}
	if kinds["v"] != exprKindMutableValue {
		t.Fatalf("lost mutable local: %v", kinds)
	}
	if kinds["_n"] != exprKindValue {
		t.Fatalf("param kinds %v", kinds)
	}
}

func TestFLAGCompileFormLet(t *testing.T) {
	n := 0
	ctx := compileContext{
		globals:       map[string]exprKind{},
		functions:     map[string]functionDef{},
		moduleSymbols: map[string]string{},
		ifTemps:       &n,
	}
	form := ListExpr{Elements: []Expr{
		SymbolExpr{Name: "let"},
		VectorExpr{Elements: []Expr{SymbolExpr{Name: "x"}, IntExpr{Value: 1}}},
		SymbolExpr{Name: "x"},
	}}
	got, err := flagCompileForm(form, ctx, nil)
	if err != nil {
		t.Fatalf("flagCompileForm let: %v\nnode=%s", err, flagrt.ValueToString(exprToFlagValue(form)))
	}
	if got.code != "let_result_1" {
		t.Fatalf("code %q", got.code)
	}
}

func TestFLAGCompileFormTrueFalse(t *testing.T) {
	n := 0
	ctx := compileContext{ifTemps: &n, globals: map[string]exprKind{}, functions: map[string]functionDef{}, moduleSymbols: map[string]string{}}
	got, err := flagCompileForm(SymbolExpr{Name: "true"}, ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got.kind != exprKindBool {
		t.Fatalf("true kind %v", got.kind)
	}
	if renderIRExpr(irFromGoExpr(got)) != "true" {
		t.Fatalf("true %q", renderIRExpr(irFromGoExpr(got)))
	}
	got, err = flagCompileForm(SymbolExpr{Name: "false"}, ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got.kind != exprKindBool {
		t.Fatalf("false kind %v", got.kind)
	}
	if renderIRExpr(irFromGoExpr(got)) != "false" {
		t.Fatalf("false %q", renderIRExpr(irFromGoExpr(got)))
	}
}

func TestFLAGCompileFormIf(t *testing.T) {
	n := 0
	ctx := compileContext{ifTemps: &n, globals: map[string]exprKind{}, functions: map[string]functionDef{}, moduleSymbols: map[string]string{}}
	got, err := flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "if"},
		SymbolExpr{Name: "true"},
		IntExpr{Value: 1},
		IntExpr{Value: 2},
	}}, ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got.code != "if_result_1" {
		t.Fatalf("code %q", got.code)
	}
	rendered := renderIRStmts(got.stmts, "")
	for _, want := range []string{
		"var if_result_1 flagrt.Value",
		"if true",
		"if_result_1 = flagrt.NewLong(1)",
		"if_result_1 = flagrt.NewLong(2)",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("missing %q in:\n%s", want, rendered)
		}
	}
}

func TestFLAGCompileFormIfEq(t *testing.T) {
	n := 0
	ctx := compileContext{ifTemps: &n, globals: map[string]exprKind{}, functions: map[string]functionDef{}, moduleSymbols: map[string]string{}}
	got, err := flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "if"},
		ListExpr{Elements: []Expr{
			SymbolExpr{Name: "="},
			SymbolExpr{Name: "tmp"},
			SymbolExpr{Name: "nil"},
		}},
		IntExpr{Value: 1},
		IntExpr{Value: 2},
	}}, ctx, map[string]exprKind{"tmp": exprKindValue})
	if err != nil {
		t.Fatal(err)
	}
	rendered := renderIRStmts(got.stmts, "")
	if strings.Contains(rendered, "IsTruthy") || strings.Contains(rendered, "NewBool") {
		t.Fatalf("if (=) still boxed:\n%s", rendered)
	}
	if !strings.Contains(rendered, "if flagrt.Eq(tmp, flagrt.NilValue)") {
		t.Fatalf("missing native bool test:\n%s", rendered)
	}
}

func TestFLAGCompileFormIfTwoArgBoxesBool(t *testing.T) {
	n := 0
	ctx := compileContext{ifTemps: &n, globals: map[string]exprKind{}, functions: map[string]functionDef{}, moduleSymbols: map[string]string{}}
	got, err := flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "if"},
		SymbolExpr{Name: "tmp"},
		SymbolExpr{Name: "true"},
	}}, ctx, map[string]exprKind{"tmp": exprKindValue})
	if err != nil {
		t.Fatal(err)
	}
	rendered := renderIRStmts(got.stmts, "")
	if !strings.Contains(rendered, "if flagrt.IsTruthy(tmp)") {
		t.Fatalf("missing value test:\n%s", rendered)
	}
	if !strings.Contains(rendered, "flagrt.NewBool(true)") {
		t.Fatalf("2-arg if should box bool then:\n%s", rendered)
	}
}

func TestFLAGCompileFormEq(t *testing.T) {
	n := 0
	ctx := compileContext{ifTemps: &n, globals: map[string]exprKind{}, functions: map[string]functionDef{}, moduleSymbols: map[string]string{}}
	got, err := flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "="},
		IntExpr{Value: 1},
		IntExpr{Value: 2},
	}}, ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got.kind != exprKindBool {
		t.Fatalf("eq kind %v", got.kind)
	}
	if renderIRExpr(irFromGoExpr(got)) != "flagrt.Eq(flagrt.NewLong(1), flagrt.NewLong(2))" {
		t.Fatalf("eq %q", renderIRExpr(irFromGoExpr(got)))
	}

	got, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "="},
		SymbolExpr{Name: "a"},
		SymbolExpr{Name: "b"},
		SymbolExpr{Name: "c"},
	}}, ctx, map[string]exprKind{"a": exprKindValue, "b": exprKindValue, "c": exprKindValue})
	if err != nil {
		t.Fatal(err)
	}
	if renderIRExpr(irFromGoExpr(got)) != "flagrt.Eq(a, b) && flagrt.Eq(b, c)" {
		t.Fatalf("eq-chain %q", renderIRExpr(irFromGoExpr(got)))
	}

	got, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "="},
		IntExpr{Value: 1},
		StringExpr{Value: "x"},
	}}, ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if renderIRExpr(irFromGoExpr(got)) != `flagrt.Eq(flagrt.NewLong(1), flagrt.NewString("x"))` {
		t.Fatalf("eq-string %q", renderIRExpr(irFromGoExpr(got)))
	}

	_, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "="},
		IntExpr{Value: 1},
	}}, ctx, nil)
	if err == nil {
		t.Fatal("expected arity error")
	}

	got, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "="},
		SymbolExpr{Name: "a"},
		SymbolExpr{Name: "b"},
	}}, ctx, map[string]exprKind{"a": exprKindBool, "b": exprKindValue})
	if err != nil {
		t.Fatal(err)
	}
	if renderIRExpr(irFromGoExpr(got)) != "flagrt.Eq(flagrt.NewBool(a), b)" {
		t.Fatalf("eq-bool %q", renderIRExpr(irFromGoExpr(got)))
	}

	ctx.selfFunctionName = "countdown"
	ctx.selfFunctionRest = true
	ctx.selfFunctionArity = 1
	ctx.selfVariadicName = "countdown_variadic"
	ctx.selfArityName = "countdown_variadic"
	got, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "countdown"},
		ListExpr{Elements: []Expr{SymbolExpr{Name: "-"}, SymbolExpr{Name: "n"}, IntExpr{Value: 1}}},
	}}, ctx, map[string]exprKind{"n": exprKindValue})
	if err != nil {
		t.Fatal(err)
	}
	if renderIRExpr(irFromGoExpr(got)) != "countdown_variadic(flagrt.Sub(n, flagrt.NewLong(1)))" {
		t.Fatalf("self-call %q", renderIRExpr(irFromGoExpr(got)))
	}
}

func TestFLAGCompileFormDirectCalls(t *testing.T) {
	n := 0
	ctx := compileContext{ifTemps: &n, globals: map[string]exprKind{}, functions: map[string]functionDef{"foo": {}}, moduleSymbols: map[string]string{}}
	got, err := flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "foo"},
		IntExpr{Value: 1},
	}}, ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if renderIRExpr(irFromGoExpr(got)) != "foo_variadic(flagrt.NewLong(1))" {
		t.Fatalf("known fn %q", renderIRExpr(irFromGoExpr(got)))
	}

	got, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "str/replace"},
		StringExpr{Value: "hello world"},
		StringExpr{Value: "world"},
		StringExpr{Value: "FLAG"},
	}}, ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	want := `flagrt.GoBind_packages_StringReplace.FunctionObject().Fn(flagrt.NewString("hello world"), flagrt.NewString("world"), flagrt.NewString("FLAG"))`
	if renderIRExpr(irFromGoExpr(got)) != want {
		t.Fatalf("gobind %q", renderIRExpr(irFromGoExpr(got)))
	}

	got, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "f"},
		IntExpr{Value: 1},
	}}, ctx, map[string]exprKind{"f": exprKindValue})
	if err != nil {
		t.Fatal(err)
	}
	if renderIRExpr(irFromGoExpr(got)) != "flagrt.Call(f, flagrt.NewLong(1))" {
		t.Fatalf("local still Call %q", renderIRExpr(irFromGoExpr(got)))
	}
}

func TestFLAGCompileFormBuiltinDirectCalls(t *testing.T) {
	n := 0
	ctx := compileContext{ifTemps: &n, globals: map[string]exprKind{}, functions: map[string]functionDef{}, moduleSymbols: map[string]string{}}
	got, err := flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "close-file"},
		SymbolExpr{Name: "x"},
	}}, ctx, map[string]exprKind{"x": exprKindValue})
	if err != nil {
		t.Fatal(err)
	}
	if renderIRExpr(irFromGoExpr(got)) != "flagrt.FnCloseFile(x)" {
		t.Fatalf("close-file %q", renderIRExpr(irFromGoExpr(got)))
	}

	got, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "apply"},
		SymbolExpr{Name: "max"},
		SymbolExpr{Name: "xs"},
	}}, ctx, map[string]exprKind{"xs": exprKindValue})
	if err != nil {
		t.Fatal(err)
	}
	if renderIRExpr(irFromGoExpr(got)) != "flagrt.ApplyFn(flagrt.FnMax, xs)" {
		t.Fatalf("apply max %q", renderIRExpr(irFromGoExpr(got)))
	}

	got, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "apply"},
		SymbolExpr{Name: "f"},
		SymbolExpr{Name: "xs"},
	}}, ctx, map[string]exprKind{"f": exprKindValue, "xs": exprKindValue})
	if err != nil {
		t.Fatal(err)
	}
	if renderIRExpr(irFromGoExpr(got)) != "flagrt.Apply(f, xs)" {
		t.Fatalf("apply local %q", renderIRExpr(irFromGoExpr(got)))
	}
}

func TestFLAGCompileFormNilEmptyCount(t *testing.T) {
	n := 0
	ctx := compileContext{ifTemps: &n, globals: map[string]exprKind{}, functions: map[string]functionDef{}, moduleSymbols: map[string]string{}}
	got, err := flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "nil?"},
		SymbolExpr{Name: "nil"},
	}}, ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got.kind != exprKindBool {
		t.Fatalf("nil? kind %v", got.kind)
	}
	if renderIRExpr(irFromGoExpr(got)) != "flagrt.IsNil(flagrt.NilValue)" {
		t.Fatalf("nil? %q", renderIRExpr(irFromGoExpr(got)))
	}

	got, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "empty?"},
		VectorExpr{},
	}}, ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if renderIRExpr(irFromGoExpr(got)) != "flagrt.IsEmpty(flagrt.NewArray())" {
		t.Fatalf("empty? %q", renderIRExpr(irFromGoExpr(got)))
	}

	got, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "empty?"},
		StringExpr{Value: "x"},
	}}, ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if renderIRExpr(irFromGoExpr(got)) != `flagrt.IsEmpty(flagrt.NewString("x"))` {
		t.Fatalf("empty?-string %q", renderIRExpr(irFromGoExpr(got)))
	}

	got, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "count"},
		VectorExpr{Elements: []Expr{IntExpr{Value: 1}, IntExpr{Value: 2}}},
	}}, ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if renderIRExpr(irFromGoExpr(got)) != "flagrt.NewLong(int64(flagrt.Count(flagrt.NewArray(flagrt.NewLong(1), flagrt.NewLong(2)))))" {
		t.Fatalf("count %q", renderIRExpr(irFromGoExpr(got)))
	}

	_, err = flagCompileForm(ListExpr{Elements: []Expr{SymbolExpr{Name: "nil?"}}}, ctx, nil)
	if err == nil {
		t.Fatal("expected nil? arity error")
	}
	_, err = flagCompileForm(ListExpr{Elements: []Expr{SymbolExpr{Name: "count"}, SymbolExpr{Name: "a"}, SymbolExpr{Name: "b"}}}, ctx, map[string]exprKind{"a": exprKindValue, "b": exprKindValue})
	if err == nil {
		t.Fatal("expected count arity error")
	}
	got, err = flagCompileForm(ListExpr{Elements: []Expr{SymbolExpr{Name: "nil?"}, SymbolExpr{Name: "a"}}}, ctx, map[string]exprKind{"a": exprKindBool})
	if err != nil {
		t.Fatal(err)
	}
	if renderIRExpr(irFromGoExpr(got)) != "flagrt.IsNil(flagrt.NewBool(a))" {
		t.Fatalf("nil?-bool %q", renderIRExpr(irFromGoExpr(got)))
	}
}

func TestFLAGCompileFormStr(t *testing.T) {
	n := 0
	ctx := compileContext{ifTemps: &n, globals: map[string]exprKind{}, functions: map[string]functionDef{}, moduleSymbols: map[string]string{}}
	got, err := flagCompileForm(ListExpr{Elements: []Expr{SymbolExpr{Name: "str"}}}, ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got.kind != exprKindString {
		t.Fatalf("kind %v", got.kind)
	}
	if renderIRExpr(irFromGoExpr(got)) != "flagrt.Str()" {
		t.Fatalf("empty str %q", renderIRExpr(irFromGoExpr(got)))
	}

	got, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "str"},
		StringExpr{Value: "a"},
		IntExpr{Value: 1},
	}}, ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if renderIRExpr(irFromGoExpr(got)) != `flagrt.Str("a", flagrt.NewLong(1))` {
		t.Fatalf("str %q", renderIRExpr(irFromGoExpr(got)))
	}
}

func TestFLAGCompileFormExInfo(t *testing.T) {
	n := 0
	ctx := compileContext{ifTemps: &n, globals: map[string]exprKind{}, functions: map[string]functionDef{}, moduleSymbols: map[string]string{}}
	got, err := flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "ex-info"},
		StringExpr{Value: "boom"},
		MapExpr{},
	}}, ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	want := `flagrt.NewMap(flagrt.NewKeyword("message"), flagrt.NewString("boom"), flagrt.NewKeyword("data"), flagrt.NewMap())`
	if renderIRExpr(irFromGoExpr(got)) != want {
		t.Fatalf("ex-info %q", renderIRExpr(irFromGoExpr(got)))
	}

	got, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "ex-info"},
		StringExpr{Value: "boom"},
		MapExpr{},
		SymbolExpr{Name: "c"},
	}}, ctx, map[string]exprKind{"c": exprKindValue})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(renderIRExpr(irFromGoExpr(got)), `flagrt.NewKeyword("cause"), c`) {
		t.Fatalf("ex-info cause %q", renderIRExpr(irFromGoExpr(got)))
	}

	_, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "ex-info"},
		IntExpr{Value: 1},
		MapExpr{},
	}}, ctx, nil)
	if err == nil {
		t.Fatal("expected string message error")
	}
	_, err = flagCompileForm(ListExpr{Elements: []Expr{SymbolExpr{Name: "ex-info"}, StringExpr{Value: "x"}}}, ctx, nil)
	if err == nil {
		t.Fatal("expected arity error")
	}
}

func TestFLAGCompileFormAssocConj(t *testing.T) {
	n := 0
	ctx := compileContext{ifTemps: &n, globals: map[string]exprKind{}, functions: map[string]functionDef{}, moduleSymbols: map[string]string{}}
	got, err := flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "assoc"},
		SymbolExpr{Name: "m"},
		KeywordExpr{Name: "k"},
		IntExpr{Value: 1},
	}}, ctx, map[string]exprKind{"m": exprKindValue})
	if err != nil {
		t.Fatal(err)
	}
	if renderIRExpr(irFromGoExpr(got)) != `flagrt.Assoc(m, flagrt.NewKeyword("k"), flagrt.NewLong(1))` {
		t.Fatalf("assoc %q", renderIRExpr(irFromGoExpr(got)))
	}

	_, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "assoc"},
		SymbolExpr{Name: "m"},
		StringExpr{Value: "k"},
		IntExpr{Value: 1},
	}}, ctx, map[string]exprKind{"m": exprKindValue})
	if err == nil {
		t.Fatal("expected assoc string error")
	}
	_, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "assoc"},
		SymbolExpr{Name: "m"},
		KeywordExpr{Name: "k"},
	}}, ctx, map[string]exprKind{"m": exprKindValue})
	if err == nil {
		t.Fatal("expected assoc arity error")
	}

	got, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "conj"},
		SymbolExpr{Name: "xs"},
		StringExpr{Value: "x"},
	}}, ctx, map[string]exprKind{"xs": exprKindValue})
	if err != nil {
		t.Fatal(err)
	}
	if renderIRExpr(irFromGoExpr(got)) != `flagrt.Conj(xs, flagrt.NewString("x"))` {
		t.Fatalf("conj %q", renderIRExpr(irFromGoExpr(got)))
	}

	_, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "conj"},
		SymbolExpr{Name: "xs"},
	}}, ctx, map[string]exprKind{"xs": exprKindValue})
	if err == nil {
		t.Fatal("expected conj arity error")
	}
}

func TestFLAGCompileFormSeqDropTakeCons(t *testing.T) {
	n := 0
	ctx := compileContext{ifTemps: &n, globals: map[string]exprKind{}, functions: map[string]functionDef{}, moduleSymbols: map[string]string{}}
	got, err := flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "seq"},
		SymbolExpr{Name: "xs"},
	}}, ctx, map[string]exprKind{"xs": exprKindValue})
	if err != nil {
		t.Fatal(err)
	}
	if renderIRExpr(irFromGoExpr(got)) != "flagrt.Seq(xs)" {
		t.Fatalf("seq %q", renderIRExpr(irFromGoExpr(got)))
	}

	got, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "seq"},
		StringExpr{Value: "ab"},
	}}, ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if renderIRExpr(irFromGoExpr(got)) != `flagrt.Seq(flagrt.NewString("ab"))` {
		t.Fatalf("seq string %q", renderIRExpr(irFromGoExpr(got)))
	}
	_, err = flagCompileForm(ListExpr{Elements: []Expr{SymbolExpr{Name: "seq"}}}, ctx, nil)
	if err == nil {
		t.Fatal("expected seq arity error")
	}

	got, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "drop"},
		IntExpr{Value: 1},
		SymbolExpr{Name: "xs"},
	}}, ctx, map[string]exprKind{"xs": exprKindValue})
	if err != nil {
		t.Fatal(err)
	}
	if renderIRExpr(irFromGoExpr(got)) != "flagrt.Drop(flagrt.NewLong(1), xs)" {
		t.Fatalf("drop %q", renderIRExpr(irFromGoExpr(got)))
	}
	_, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "drop"},
		StringExpr{Value: "1"},
		SymbolExpr{Name: "xs"},
	}}, ctx, map[string]exprKind{"xs": exprKindValue})
	if err == nil {
		t.Fatal("expected drop string error")
	}
	_, err = flagCompileForm(ListExpr{Elements: []Expr{SymbolExpr{Name: "drop"}, IntExpr{Value: 1}}}, ctx, nil)
	if err == nil {
		t.Fatal("expected drop arity error")
	}

	got, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "take"},
		IntExpr{Value: 2},
		SymbolExpr{Name: "xs"},
	}}, ctx, map[string]exprKind{"xs": exprKindValue})
	if err != nil {
		t.Fatal(err)
	}
	if renderIRExpr(irFromGoExpr(got)) != "flagrt.Take(flagrt.NewLong(2), xs)" {
		t.Fatalf("take %q", renderIRExpr(irFromGoExpr(got)))
	}

	got, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "cons"},
		SymbolExpr{Name: "x"},
		SymbolExpr{Name: "xs"},
	}}, ctx, map[string]exprKind{"x": exprKindValue, "xs": exprKindValue})
	if err != nil {
		t.Fatal(err)
	}
	if renderIRExpr(irFromGoExpr(got)) != "flagrt.Cons(x, xs)" {
		t.Fatalf("cons %q", renderIRExpr(irFromGoExpr(got)))
	}
	_, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "cons"},
		StringExpr{Value: "x"},
		SymbolExpr{Name: "xs"},
	}}, ctx, map[string]exprKind{"xs": exprKindValue})
	if err == nil {
		t.Fatal("expected cons string error")
	}
}

func TestFLAGCompileFormGetNth(t *testing.T) {
	n := 0
	ctx := compileContext{ifTemps: &n, globals: map[string]exprKind{}, functions: map[string]functionDef{}, moduleSymbols: map[string]string{}}
	got, err := flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "get"},
		SymbolExpr{Name: "m"},
		KeywordExpr{Name: "k"},
	}}, ctx, map[string]exprKind{"m": exprKindValue})
	if err != nil {
		t.Fatal(err)
	}
	if renderIRExpr(irFromGoExpr(got)) != `flagrt.Get(m, flagrt.NewKeyword("k"))` {
		t.Fatalf("get %q", renderIRExpr(irFromGoExpr(got)))
	}

	got, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "get"},
		SymbolExpr{Name: "m"},
		StringExpr{Value: "k"},
		IntExpr{Value: 0},
	}}, ctx, map[string]exprKind{"m": exprKindValue})
	if err != nil {
		t.Fatal(err)
	}
	if renderIRExpr(irFromGoExpr(got)) != `flagrt.Get(m, flagrt.NewString("k"), flagrt.NewLong(0))` {
		t.Fatalf("get default %q", renderIRExpr(irFromGoExpr(got)))
	}
	_, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "get"},
		SymbolExpr{Name: "m"},
	}}, ctx, map[string]exprKind{"m": exprKindValue})
	if err == nil {
		t.Fatal("expected get arity error")
	}

	got, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "nth"},
		SymbolExpr{Name: "xs"},
		IntExpr{Value: 0},
	}}, ctx, map[string]exprKind{"xs": exprKindValue})
	if err != nil {
		t.Fatal(err)
	}
	if renderIRExpr(irFromGoExpr(got)) != "flagrt.Nth(xs, flagrt.NewLong(0))" {
		t.Fatalf("nth %q", renderIRExpr(irFromGoExpr(got)))
	}

	got, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "nth"},
		SymbolExpr{Name: "xs"},
		IntExpr{Value: 0},
		StringExpr{Value: "missing"},
	}}, ctx, map[string]exprKind{"xs": exprKindValue})
	if err != nil {
		t.Fatal(err)
	}
	if renderIRExpr(irFromGoExpr(got)) != `flagrt.Nth(xs, flagrt.NewLong(0), flagrt.NewString("missing"))` {
		t.Fatalf("nth default %q", renderIRExpr(irFromGoExpr(got)))
	}
	_, err = flagCompileForm(ListExpr{Elements: []Expr{SymbolExpr{Name: "nth"}}}, ctx, nil)
	if err == nil {
		t.Fatal("expected nth arity error")
	}
}

func TestFLAGCompileFormNotEmptyTypeOfVec(t *testing.T) {
	n := 0
	ctx := compileContext{ifTemps: &n, globals: map[string]exprKind{}, functions: map[string]functionDef{}, moduleSymbols: map[string]string{}}
	got, err := flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "not-empty"},
		SymbolExpr{Name: "xs"},
	}}, ctx, map[string]exprKind{"xs": exprKindValue})
	if err != nil {
		t.Fatal(err)
	}
	if renderIRExpr(irFromGoExpr(got)) != "flagrt.NotEmpty(xs)" {
		t.Fatalf("not-empty %q", renderIRExpr(irFromGoExpr(got)))
	}
	_, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "not-empty"},
		StringExpr{Value: "x"},
	}}, ctx, nil)
	if err == nil {
		t.Fatal("expected not-empty string error")
	}

	got, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "type-of"},
		StringExpr{Value: "x"},
	}}, ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if renderIRExpr(irFromGoExpr(got)) != `flagrt.TypeOf(flagrt.NewString("x"))` {
		t.Fatalf("type-of %q", renderIRExpr(irFromGoExpr(got)))
	}

	got, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "vec"},
		SymbolExpr{Name: "xs"},
	}}, ctx, map[string]exprKind{"xs": exprKindValue})
	if err != nil {
		t.Fatal(err)
	}
	if renderIRExpr(irFromGoExpr(got)) != "flagrt.Vec(xs)" {
		t.Fatalf("vec %q", renderIRExpr(irFromGoExpr(got)))
	}
	_, err = flagCompileForm(ListExpr{Elements: []Expr{SymbolExpr{Name: "vec"}}}, ctx, nil)
	if err == nil {
		t.Fatal("expected vec arity error")
	}
}

func TestFLAGCompileFormDissocContains(t *testing.T) {
	n := 0
	ctx := compileContext{ifTemps: &n, globals: map[string]exprKind{}, functions: map[string]functionDef{}, moduleSymbols: map[string]string{}}
	got, err := flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "dissoc"},
		SymbolExpr{Name: "m"},
		KeywordExpr{Name: "k"},
	}}, ctx, map[string]exprKind{"m": exprKindValue})
	if err != nil {
		t.Fatal(err)
	}
	if renderIRExpr(irFromGoExpr(got)) != `flagrt.MapDissoc(m, flagrt.NewKeyword("k"))` {
		t.Fatalf("dissoc %q", renderIRExpr(irFromGoExpr(got)))
	}
	_, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "dissoc"},
		SymbolExpr{Name: "m"},
		StringExpr{Value: "k"},
	}}, ctx, map[string]exprKind{"m": exprKindValue})
	if err == nil {
		t.Fatal("expected dissoc string error")
	}
	_, err = flagCompileForm(ListExpr{Elements: []Expr{SymbolExpr{Name: "dissoc"}}}, ctx, nil)
	if err == nil {
		t.Fatal("expected dissoc arity error")
	}

	got, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "contains?"},
		SymbolExpr{Name: "m"},
		KeywordExpr{Name: "k"},
	}}, ctx, map[string]exprKind{"m": exprKindValue})
	if err != nil {
		t.Fatal(err)
	}
	if got.kind != exprKindBool {
		t.Fatalf("contains? kind %v", got.kind)
	}
	if renderIRExpr(irFromGoExpr(got)) != `flagrt.Contains(m, flagrt.NewKeyword("k"))` {
		t.Fatalf("contains? %q", renderIRExpr(irFromGoExpr(got)))
	}
	_, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "contains?"},
		SymbolExpr{Name: "m"},
	}}, ctx, map[string]exprKind{"m": exprKindValue})
	if err == nil {
		t.Fatal("expected contains? arity error")
	}
}

func TestFLAGCompileFormListArrayHashMap(t *testing.T) {
	n := 0
	ctx := compileContext{ifTemps: &n, globals: map[string]exprKind{}, functions: map[string]functionDef{}, moduleSymbols: map[string]string{}}
	got, err := flagCompileForm(ListExpr{Elements: []Expr{SymbolExpr{Name: "list"}}}, ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if renderIRExpr(irFromGoExpr(got)) != "flagrt.NewList()" {
		t.Fatalf("empty list %q", renderIRExpr(irFromGoExpr(got)))
	}

	got, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "list"},
		StringExpr{Value: "a"},
		IntExpr{Value: 1},
	}}, ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if renderIRExpr(irFromGoExpr(got)) != `flagrt.NewList(flagrt.NewString("a"), flagrt.NewLong(1))` {
		t.Fatalf("list %q", renderIRExpr(irFromGoExpr(got)))
	}

	got, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "array"},
		IntExpr{Value: 1},
	}}, ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if renderIRExpr(irFromGoExpr(got)) != "flagrt.NewArray(flagrt.NewLong(1))" {
		t.Fatalf("array %q", renderIRExpr(irFromGoExpr(got)))
	}

	got, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "hash-map"},
		KeywordExpr{Name: "k"},
		IntExpr{Value: 1},
	}}, ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if renderIRExpr(irFromGoExpr(got)) != `flagrt.NewMap(flagrt.NewKeyword("k"), flagrt.NewLong(1))` {
		t.Fatalf("hash-map %q", renderIRExpr(irFromGoExpr(got)))
	}
	_, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "hash-map"},
		KeywordExpr{Name: "k"},
	}}, ctx, nil)
	if err == nil {
		t.Fatal("expected hash-map arity error")
	}
}

func TestFLAGCompileFormIntoConcat(t *testing.T) {
	n := 0
	ctx := compileContext{ifTemps: &n, globals: map[string]exprKind{}, functions: map[string]functionDef{}, moduleSymbols: map[string]string{}}
	got, err := flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "into"},
		SymbolExpr{Name: "to"},
		SymbolExpr{Name: "from"},
	}}, ctx, map[string]exprKind{"to": exprKindValue, "from": exprKindValue})
	if err != nil {
		t.Fatal(err)
	}
	if renderIRExpr(irFromGoExpr(got)) != "flagrt.Into(to, from)" {
		t.Fatalf("into %q", renderIRExpr(irFromGoExpr(got)))
	}
	_, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "into"},
		StringExpr{Value: "to"},
		SymbolExpr{Name: "from"},
	}}, ctx, map[string]exprKind{"from": exprKindValue})
	if err == nil {
		t.Fatal("expected into string error")
	}

	got, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "concat"},
		SymbolExpr{Name: "a"},
		StringExpr{Value: "b"},
	}}, ctx, map[string]exprKind{"a": exprKindValue})
	if err != nil {
		t.Fatal(err)
	}
	if renderIRExpr(irFromGoExpr(got)) != `flagrt.Concat(a, flagrt.NewString("b"))` {
		t.Fatalf("concat %q", renderIRExpr(irFromGoExpr(got)))
	}
}

func TestFLAGCompileFormExFieldsPrintln(t *testing.T) {
	n := 0
	ctx := compileContext{ifTemps: &n, globals: map[string]exprKind{}, functions: map[string]functionDef{}, moduleSymbols: map[string]string{}}
	got, err := flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "ex-message"},
		SymbolExpr{Name: "e"},
	}}, ctx, map[string]exprKind{"e": exprKindValue})
	if err != nil {
		t.Fatal(err)
	}
	if renderIRExpr(irFromGoExpr(got)) != "flagrt.ExMessage(e)" {
		t.Fatalf("ex-message %q", renderIRExpr(irFromGoExpr(got)))
	}

	got, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "ex-data"},
		StringExpr{Value: "x"},
	}}, ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if renderIRExpr(irFromGoExpr(got)) != `flagrt.ExData(flagrt.NewString("x"))` {
		t.Fatalf("ex-data %q", renderIRExpr(irFromGoExpr(got)))
	}

	got, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "ex-cause"},
		SymbolExpr{Name: "e"},
	}}, ctx, map[string]exprKind{"e": exprKindValue})
	if err != nil {
		t.Fatal(err)
	}
	if renderIRExpr(irFromGoExpr(got)) != "flagrt.ExCause(e)" {
		t.Fatalf("ex-cause %q", renderIRExpr(irFromGoExpr(got)))
	}
	_, err = flagCompileForm(ListExpr{Elements: []Expr{SymbolExpr{Name: "ex-message"}}}, ctx, nil)
	if err == nil {
		t.Fatal("expected ex-message arity error")
	}

	got, err = flagCompileForm(ListExpr{Elements: []Expr{SymbolExpr{Name: "println"}}}, ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got.kind != exprKindValue {
		t.Fatalf("println kind %v", got.kind)
	}
	if renderIRExpr(irFromGoExpr(got)) != "flagrt.Println()" {
		t.Fatalf("empty println %q", renderIRExpr(irFromGoExpr(got)))
	}

	got, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "println"},
		StringExpr{Value: "a"},
		IntExpr{Value: 1},
	}}, ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if renderIRExpr(irFromGoExpr(got)) != `flagrt.Println("a", flagrt.NewLong(1))` {
		t.Fatalf("println %q", renderIRExpr(irFromGoExpr(got)))
	}
}

func TestFLAGCompileFormRemainingRuntimeOps(t *testing.T) {
	n := 0
	ctx := compileContext{ifTemps: &n, globals: map[string]exprKind{}, functions: map[string]functionDef{}, moduleSymbols: map[string]string{}}
	vals := map[string]exprKind{"a": exprKindValue, "b": exprKindValue, "f": exprKindValue, "xs": exprKindValue}

	got, err := flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "compare"}, StringExpr{Value: "a"}, StringExpr{Value: "b"},
	}}, ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if renderIRExpr(irFromGoExpr(got)) != `flagrt.Compare(flagrt.NewString("a"), flagrt.NewString("b"))` {
		t.Fatalf("compare %q", renderIRExpr(irFromGoExpr(got)))
	}

	got, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "=="}, SymbolExpr{Name: "a"}, SymbolExpr{Name: "b"},
	}}, ctx, vals)
	if err != nil {
		t.Fatal(err)
	}
	if got.kind != exprKindBool {
		t.Fatalf("== kind %v", got.kind)
	}
	if renderIRExpr(irFromGoExpr(got)) != "flagrt.NumericEq(a, b)" {
		t.Fatalf("== %q", renderIRExpr(irFromGoExpr(got)))
	}
	_, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "=="}, StringExpr{Value: "a"}, SymbolExpr{Name: "b"},
	}}, ctx, vals)
	if err == nil {
		t.Fatal("expected == string error")
	}

	got, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "<"}, SymbolExpr{Name: "a"}, SymbolExpr{Name: "b"}, SymbolExpr{Name: "a"},
	}}, ctx, vals)
	if err != nil {
		t.Fatal(err)
	}
	if got.kind != exprKindBool {
		t.Fatalf("< kind %v", got.kind)
	}
	if renderIRExpr(irFromGoExpr(got)) != "flagrt.Lt(a, b) && flagrt.Lt(b, a)" {
		t.Fatalf("< %q", renderIRExpr(irFromGoExpr(got)))
	}

	got, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "name"}, StringExpr{Value: "x"},
	}}, ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got.kind != exprKindString {
		t.Fatalf("name kind %v", got.kind)
	}
	if renderIRExpr(irFromGoExpr(got)) != `flagrt.Name("x")` {
		t.Fatalf("name %q", renderIRExpr(irFromGoExpr(got)))
	}

	got, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "keyword"}, StringExpr{Value: "k"},
	}}, ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if renderIRExpr(irFromGoExpr(got)) != `flagrt.Keyword("k")` {
		t.Fatalf("keyword %q", renderIRExpr(irFromGoExpr(got)))
	}

	got, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "format"}, StringExpr{Value: "%s"}, IntExpr{Value: 1},
	}}, ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if renderIRExpr(irFromGoExpr(got)) != `flagrt.NewString(flagrt.Format("%s", flagrt.NewLong(1)))` {
		t.Fatalf("format %q", renderIRExpr(irFromGoExpr(got)))
	}

	got, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "map"}, SymbolExpr{Name: "f"}, SymbolExpr{Name: "xs"},
	}}, ctx, vals)
	if err != nil {
		t.Fatal(err)
	}
	if renderIRExpr(irFromGoExpr(got)) != "flagrt.Map(f, xs)" {
		t.Fatalf("map %q", renderIRExpr(irFromGoExpr(got)))
	}

	got, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "filter"}, SymbolExpr{Name: "f"}, SymbolExpr{Name: "xs"},
	}}, ctx, vals)
	if err != nil {
		t.Fatal(err)
	}
	if renderIRExpr(irFromGoExpr(got)) != "flagrt.Filter(f, xs)" {
		t.Fatalf("filter %q", renderIRExpr(irFromGoExpr(got)))
	}

	got, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "reduce"}, SymbolExpr{Name: "f"}, SymbolExpr{Name: "xs"},
	}}, ctx, vals)
	if err != nil {
		t.Fatal(err)
	}
	if renderIRExpr(irFromGoExpr(got)) != "flagrt.Reduce(f, xs)" {
		t.Fatalf("reduce %q", renderIRExpr(irFromGoExpr(got)))
	}

	got, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "to-json"}, SymbolExpr{Name: "a"},
	}}, ctx, vals)
	if err != nil {
		t.Fatal(err)
	}
	if got.kind != exprKindString {
		t.Fatalf("to-json kind %v", got.kind)
	}
	if renderIRExpr(irFromGoExpr(got)) != "flagrt.ToJSON(a)" {
		t.Fatalf("to-json %q", renderIRExpr(irFromGoExpr(got)))
	}

	got, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "from-json"}, StringExpr{Value: "{}"},
	}}, ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if renderIRExpr(irFromGoExpr(got)) != `flagrt.FromJSON("{}")` {
		t.Fatalf("from-json %q", renderIRExpr(irFromGoExpr(got)))
	}

	got, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "open-file"}, StringExpr{Value: "x"},
	}}, ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if renderIRExpr(irFromGoExpr(got)) != `flagrt.OpenFile("x")` {
		t.Fatalf("open-file %q", renderIRExpr(irFromGoExpr(got)))
	}

	got, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "file-to-strings"}, StringExpr{Value: "x"},
	}}, ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if renderIRExpr(irFromGoExpr(got)) != `flagrt.FileToStringsPath("x")` {
		t.Fatalf("file-to-strings %q", renderIRExpr(irFromGoExpr(got)))
	}

	got, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "repeat"}, StringExpr{Value: "x"},
	}}, ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if renderIRExpr(irFromGoExpr(got)) != `flagrt.Repeat(flagrt.NewString("x"))` {
		t.Fatalf("repeat %q", renderIRExpr(irFromGoExpr(got)))
	}

	_, err = flagCompileForm(ListExpr{Elements: []Expr{SymbolExpr{Name: "map"}, SymbolExpr{Name: "f"}}}, ctx, vals)
	if err == nil {
		t.Fatal("expected map arity error")
	}
}

func TestFLAGCompileFormThrow(t *testing.T) {
	n := 0
	ctx := compileContext{ifTemps: &n, globals: map[string]exprKind{}, functions: map[string]functionDef{}, moduleSymbols: map[string]string{}}
	got, err := flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "throw"},
		SymbolExpr{Name: "err"},
	}}, ctx, map[string]exprKind{"err": exprKindValue})
	if err != nil {
		t.Fatal(err)
	}
	rendered := renderIRExpr(irFromGoExpr(got))
	for _, want := range []string{"flagrt.Throw(err)", "return flagrt.NilValue"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("missing %q in:\n%s", want, rendered)
		}
	}

	got, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "throw"},
		StringExpr{Value: "boom"},
	}}, ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(renderIRExpr(irFromGoExpr(got)), `flagrt.Throw(flagrt.NewString("boom"))`) {
		t.Fatalf("throw string:\n%s", renderIRExpr(irFromGoExpr(got)))
	}

	_, err = flagCompileForm(ListExpr{Elements: []Expr{SymbolExpr{Name: "throw"}}}, ctx, nil)
	if err == nil {
		t.Fatal("expected throw arity error")
	}
}

func TestFLAGCompileFormLoopRecur(t *testing.T) {
	n := 0
	ctx := compileContext{ifTemps: &n, globals: map[string]exprKind{}, functions: map[string]functionDef{}, moduleSymbols: map[string]string{}}
	got, err := flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "loop"},
		VectorExpr{},
		IntExpr{Value: 1},
	}}, ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	rendered := renderIRExpr(irFromGoExpr(got))
	for _, want := range []string{"UnwrapRecur", "return __loopResult", "flagrt.NewLong(1)"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("missing %q in:\n%s", want, rendered)
		}
	}

	got, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "loop"},
		VectorExpr{Elements: []Expr{SymbolExpr{Name: "n"}, IntExpr{Value: 0}}},
		ListExpr{Elements: []Expr{
			SymbolExpr{Name: "if"},
			ListExpr{Elements: []Expr{SymbolExpr{Name: "="}, SymbolExpr{Name: "n"}, IntExpr{Value: 2}}},
			SymbolExpr{Name: "n"},
			ListExpr{Elements: []Expr{
				SymbolExpr{Name: "recur"},
				ListExpr{Elements: []Expr{SymbolExpr{Name: "+"}, SymbolExpr{Name: "n"}, IntExpr{Value: 1}}},
			}},
		}},
	}}, ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	rendered = renderIRExpr(irFromGoExpr(got))
	for _, want := range []string{
		"var n = flagrt.NewLong(0)",
		"flagrt.NewRecur",
		"UnwrapRecur",
		"n = __recurValues[0]",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("missing %q in:\n%s", want, rendered)
		}
	}

	_, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "recur"},
		IntExpr{Value: 1},
	}}, ctx, nil)
	if err == nil {
		t.Fatal("expected recur outside loop")
	}

	_, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "loop"},
		VectorExpr{Elements: []Expr{SymbolExpr{Name: "n"}, IntExpr{Value: 0}}},
		ListExpr{Elements: []Expr{SymbolExpr{Name: "recur"}}},
	}}, ctx, nil)
	if err == nil {
		t.Fatal("expected recur arity error")
	}

	_, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "loop"},
		VectorExpr{Elements: []Expr{SymbolExpr{Name: "x"}, IntExpr{Value: 1}, SymbolExpr{Name: "x"}, IntExpr{Value: 2}}},
		SymbolExpr{Name: "x"},
	}}, ctx, nil)
	if err == nil {
		t.Fatal("expected duplicate loop binding")
	}

	_, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "loop"},
		IntExpr{Value: 1},
		IntExpr{Value: 2},
	}}, ctx, nil)
	if err == nil {
		t.Fatal("expected binding vector error")
	}

	got, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "loop"},
		VectorExpr{Elements: []Expr{SymbolExpr{Name: "i"}, IntExpr{Value: 0}}},
		ListExpr{Elements: []Expr{
			SymbolExpr{Name: "do"},
			ListExpr{Elements: []Expr{
				SymbolExpr{Name: "fn"},
				VectorExpr{},
				SymbolExpr{Name: "i"},
			}},
			ListExpr{Elements: []Expr{
				SymbolExpr{Name: "recur"},
				ListExpr{Elements: []Expr{SymbolExpr{Name: "+"}, SymbolExpr{Name: "i"}, IntExpr{Value: 1}}},
			}},
		}},
	}}, ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	rendered = renderIRExpr(irFromGoExpr(got))
	if !strings.Contains(rendered, "flagrt.NewRecur") {
		t.Fatalf("fn must not drop loop bindings for sibling recur:\n%s", rendered)
	}
}

func TestFLAGCompileFormMapBoxesBool(t *testing.T) {
	n := 0
	ctx := compileContext{ifTemps: &n, globals: map[string]exprKind{}, functions: map[string]functionDef{}, moduleSymbols: map[string]string{}}
	got, err := flagCompileForm(MapExpr{
		Line: 3,
		Col:  12,
		Entries: []Expr{
			KeywordExpr{Name: "a", Line: 3, Col: 13},
			ListExpr{
				Line: 3,
				Col:  16,
				Elements: []Expr{
					SymbolExpr{Name: "<", Line: 3, Col: 17},
					IntExpr{Value: 1, Line: 3, Col: 19},
					IntExpr{Value: 2, Line: 3, Col: 21},
				},
			},
		},
	}, ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if renderIRExpr(irFromGoExpr(got)) != `flagrt.NewMap(flagrt.NewKeyword("a"), flagrt.NewBool(flagrt.Lt(flagrt.NewLong(1), flagrt.NewLong(2))))` {
		t.Fatalf("map bool %q", renderIRExpr(irFromGoExpr(got)))
	}
}

func TestFLAGCompileFormDefer(t *testing.T) {
	n := 0
	ctx := compileContext{ifTemps: &n, globals: map[string]exprKind{}, functions: map[string]functionDef{}, moduleSymbols: map[string]string{}}
	got, err := flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "defer"},
		SymbolExpr{Name: "f"},
	}}, ctx, map[string]exprKind{"f": exprKindValue})
	if err != nil {
		t.Fatal(err)
	}
	if got.kind != exprKindDefer {
		t.Fatalf("kind %v", got.kind)
	}
	if got.stmt == nil {
		t.Fatal("missing defer stmt")
	}
	if !strings.Contains(renderIRStmt(got.stmt, ""), "defer flagrt.Call(f)") {
		t.Fatalf("stmt:\n%s", renderIRStmt(got.stmt, ""))
	}

	got, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "defer"},
		StringExpr{Value: "nope"},
	}}, ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(renderIRStmt(got.stmt, ""), `defer flagrt.Call(flagrt.NewString("nope"))`) {
		t.Fatalf("string thunk:\n%s", renderIRStmt(got.stmt, ""))
	}

	got, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "do"},
		ListExpr{Elements: []Expr{SymbolExpr{Name: "defer"}, SymbolExpr{Name: "f"}}},
		IntExpr{Value: 1},
	}}, ctx, map[string]exprKind{"f": exprKindValue})
	if err != nil {
		t.Fatal(err)
	}
	if got.kind != exprKindValue {
		t.Fatalf("do+defer kind %v", got.kind)
	}
	rendered := renderIRExpr(irFromGoExpr(got))
	for _, want := range []string{"defer flagrt.Call(f)", "return flagrt.NewLong(1)"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("missing %q in:\n%s", want, rendered)
		}
	}

	_, err = flagCompileForm(ListExpr{Elements: []Expr{SymbolExpr{Name: "defer"}}}, ctx, nil)
	if err == nil {
		t.Fatal("expected defer arity error")
	}
}

func TestFLAGCompileFormDoto(t *testing.T) {
	n := 0
	ctx := compileContext{ifTemps: &n, globals: map[string]exprKind{}, functions: map[string]functionDef{}, moduleSymbols: map[string]string{}}
	got, err := flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "doto"},
		SymbolExpr{Name: "obj"},
		SymbolExpr{Name: "step"},
	}}, ctx, map[string]exprKind{"obj": exprKindValue, "step": exprKindValue})
	if err != nil {
		t.Fatal(err)
	}
	rendered := renderIRExpr(irFromGoExpr(got))
	for _, want := range []string{"__doto := obj", "flagrt.Call(step, __doto)", "return __doto"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("missing %q in:\n%s", want, rendered)
		}
	}

	got, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "doto"},
		StringExpr{Value: "x"},
	}}, ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(renderIRExpr(irFromGoExpr(got)), `__doto := flagrt.NewString("x")`) {
		t.Fatalf("string target:\n%s", renderIRExpr(irFromGoExpr(got)))
	}

	_, err = flagCompileForm(ListExpr{Elements: []Expr{SymbolExpr{Name: "doto"}}}, ctx, nil)
	if err == nil {
		t.Fatal("expected doto arity error")
	}
}

func TestFLAGCompileFormTry(t *testing.T) {
	n := 0
	ctx := compileContext{ifTemps: &n, globals: map[string]exprKind{}, functions: map[string]functionDef{}, moduleSymbols: map[string]string{}}
	got, err := flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "try"},
		IntExpr{Value: 1},
	}}, ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if renderIRExpr(irFromGoExpr(got)) != "flagrt.NewLong(1)" {
		t.Fatalf("bare try %q", renderIRExpr(irFromGoExpr(got)))
	}

	got, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "try"},
		IntExpr{Value: 1},
		ListExpr{Elements: []Expr{
			SymbolExpr{Name: "catch"},
			SymbolExpr{Name: "Exception"},
			SymbolExpr{Name: "e"},
			SymbolExpr{Name: "e"},
		}},
	}}, ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	rendered := renderIRExpr(irFromGoExpr(got))
	for _, want := range []string{
		"var __flag_try_result flagrt.Value",
		`CatchMatches("Exception", __flag_thrown)`,
		"var e = __flag_thrown",
		"return __flag_try_result",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("missing %q in:\n%s", want, rendered)
		}
	}

	got, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "try"},
		IntExpr{Value: 1},
		ListExpr{Elements: []Expr{
			SymbolExpr{Name: "catch"},
			KeywordExpr{Name: "default"},
			SymbolExpr{Name: "e"},
		}},
		ListExpr{Elements: []Expr{
			SymbolExpr{Name: "finally"},
			IntExpr{Value: 2},
		}},
	}}, ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	rendered = renderIRExpr(irFromGoExpr(got))
	for _, want := range []string{
		`CatchMatches(":default", __flag_thrown)`,
		"defer func() {",
		"_ = flagrt.NewLong(2)",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("missing %q in:\n%s", want, rendered)
		}
	}

	_, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "try"},
		IntExpr{Value: 1},
		ListExpr{Elements: []Expr{
			SymbolExpr{Name: "catch"},
			SymbolExpr{Name: "Nope"},
			SymbolExpr{Name: "e"},
		}},
	}}, ctx, nil)
	if err == nil {
		t.Fatal("expected unknown catch type")
	}
	_, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "try"},
		IntExpr{Value: 1},
		ListExpr{Elements: []Expr{SymbolExpr{Name: "finally"}, IntExpr{Value: 2}}},
		IntExpr{Value: 3},
	}}, ctx, nil)
	if err == nil {
		t.Fatal("expected catch/finally last error")
	}
}

func TestFLAGCompileFormDoseqFor(t *testing.T) {
	n := 0
	ctx := compileContext{ifTemps: &n, globals: map[string]exprKind{}, functions: map[string]functionDef{}, moduleSymbols: map[string]string{}}
	got, err := flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "doseq"},
		VectorExpr{Elements: []Expr{SymbolExpr{Name: "x"}, SymbolExpr{Name: "xs"}}},
		SymbolExpr{Name: "x"},
	}}, ctx, map[string]exprKind{"xs": exprKindValue})
	if err != nil {
		t.Fatal(err)
	}
	rendered := renderIRExpr(irFromGoExpr(got))
	for _, want := range []string{"flagrt.DoAll", "flagrt.MapCat", "x := args[0]", "_ = x", "return flagrt.NewArray()"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("missing %q in:\n%s", want, rendered)
		}
	}

	got, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "doseq"},
		VectorExpr{},
		IntExpr{Value: 1},
	}}, ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	rendered = renderIRExpr(irFromGoExpr(got))
	for _, want := range []string{"flagrt.DoAll", "_ = flagrt.NewLong(1)", "return flagrt.NewArray()"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("empty doseq missing %q in:\n%s", want, rendered)
		}
	}

	got, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "for"},
		VectorExpr{Elements: []Expr{SymbolExpr{Name: "x"}, SymbolExpr{Name: "xs"}}},
		SymbolExpr{Name: "x"},
	}}, ctx, map[string]exprKind{"xs": exprKindValue})
	if err != nil {
		t.Fatal(err)
	}
	rendered = renderIRExpr(irFromGoExpr(got))
	if strings.Contains(rendered, "flagrt.DoAll") {
		t.Fatalf("for should not wrap DoAll:\n%s", rendered)
	}
	for _, want := range []string{"flagrt.MapCat", "x := args[0]", "return flagrt.NewArray(x)"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("missing %q in:\n%s", want, rendered)
		}
	}

	got, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "for"},
		VectorExpr{},
		IntExpr{Value: 1},
	}}, ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if renderIRExpr(irFromGoExpr(got)) != "flagrt.NewArray(flagrt.NewLong(1))" {
		t.Fatalf("empty for %q", renderIRExpr(irFromGoExpr(got)))
	}

	got, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "doseq"},
		VectorExpr{Elements: []Expr{SymbolExpr{Name: "_x"}, StringExpr{Value: "ab"}}},
		IntExpr{Value: 1},
	}}, ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	rendered = renderIRExpr(irFromGoExpr(got))
	if !strings.Contains(rendered, `_x := args[0]`) || !strings.Contains(rendered, `_ = _x`) {
		t.Fatalf("unused binding:\n%s", rendered)
	}
	if !strings.Contains(rendered, `flagrt.NewString("ab")`) {
		t.Fatalf("string coll:\n%s", rendered)
	}

	_, err = flagCompileForm(ListExpr{Elements: []Expr{SymbolExpr{Name: "doseq"}, VectorExpr{}}}, ctx, nil)
	if err == nil {
		t.Fatal("expected doseq arity error")
	}
	_, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "for"},
		IntExpr{Value: 1},
		IntExpr{Value: 2},
	}}, ctx, nil)
	if err == nil {
		t.Fatal("expected for binding vector error")
	}
	_, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "doseq"},
		VectorExpr{Elements: []Expr{IntExpr{Value: 1}, SymbolExpr{Name: "xs"}}},
		SymbolExpr{Name: "xs"},
	}}, ctx, map[string]exprKind{"xs": exprKindValue})
	if err == nil {
		t.Fatal("expected doseq binding name error")
	}
}

func TestFLAGCompileFormDotTestingHashDestructure(t *testing.T) {
	n := 0
	ctx := compileContext{ifTemps: &n, globals: map[string]exprKind{}, functions: map[string]functionDef{}, moduleSymbols: map[string]string{}}

	got, err := flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: ".write"},
		SymbolExpr{Name: "f"},
		StringExpr{Value: "hi"},
	}}, ctx, map[string]exprKind{"f": exprKindValue})
	if err != nil {
		t.Fatal(err)
	}
	if renderIRExpr(irFromGoExpr(got)) != `flagrt.Write(f, "hi")` {
		t.Fatalf("write %q", renderIRExpr(irFromGoExpr(got)))
	}

	got, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: ".endsWith"},
		SymbolExpr{Name: "s"},
		StringExpr{Value: "ing"},
	}}, ctx, map[string]exprKind{"s": exprKindValue})
	if err != nil {
		t.Fatal(err)
	}
	if got.kind != exprKindBool {
		t.Fatalf("endsWith kind %v", got.kind)
	}
	if renderIRExpr(irFromGoExpr(got)) != `strings.HasSuffix(flagrt.Str(s), "ing")` {
		t.Fatalf("endsWith %q", renderIRExpr(irFromGoExpr(got)))
	}

	got, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "testing"},
		StringExpr{Value: "label"},
		IntExpr{Value: 1},
	}}, ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	rendered := renderIRExpr(irFromGoExpr(got))
	for _, want := range []string{"_ = flagrt.NewLong(1)", "return flagrt.NilValue"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("testing missing %q in:\n%s", want, rendered)
		}
	}

	got, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "is"},
		SymbolExpr{Name: "true"},
		StringExpr{Value: "ok"},
	}}, ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	rendered = renderIRExpr(irFromGoExpr(got))
	for _, want := range []string{
		"if !(true)",
		`panic("ok")`,
		"return flagrt.NilValue",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("is missing %q in:\n%s", want, rendered)
		}
	}

	got, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "expect-exception"},
		IntExpr{Value: 1},
	}}, ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	rendered = renderIRExpr(irFromGoExpr(got))
	for _, want := range []string{
		"r := recover()",
		"if r == nil",
		`panic("expected exception from 1")`,
		"_ = flagrt.NewLong(1)",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("expect-exception missing %q in:\n%s", want, rendered)
		}
	}

	got, err = flagCompileForm(HashFnExpr{Body: ListExpr{Elements: []Expr{
		SymbolExpr{Name: "+"},
		SymbolExpr{Name: "%"},
		SymbolExpr{Name: "%2"},
	}}}, ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	rendered = renderIRExpr(irFromGoExpr(got))
	for _, want := range []string{
		"__p1 := args[0]",
		"__p2 := args[1]",
		"flagrt.Add(__p1, __p2)",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("hash-fn missing %q in:\n%s", want, rendered)
		}
	}

	got, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "let"},
		VectorExpr{Elements: []Expr{
			MapExpr{Entries: []Expr{
				KeywordExpr{Name: "a"}, SymbolExpr{Name: "a"},
				KeywordExpr{Name: "or"}, MapExpr{Entries: []Expr{
					SymbolExpr{Name: "a"}, IntExpr{Value: 0},
				}},
			}},
			SymbolExpr{Name: "m"},
		}},
		SymbolExpr{Name: "a"},
	}}, ctx, map[string]exprKind{"m": exprKindValue})
	if err != nil {
		t.Fatal(err)
	}
	rendered = renderIRStmts(got.stmts, "")
	for _, want := range []string{
		"__bind0",
		`flagrt.Get(__bind0, flagrt.NewKeyword("a"), flagrt.NewLong(0))`,
		"var a =",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("let map missing %q in:\n%s", want, rendered)
		}
	}

	got, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "let"},
		VectorExpr{Elements: []Expr{
			VectorExpr{Elements: []Expr{SymbolExpr{Name: "x"}, SymbolExpr{Name: "y"}}},
			SymbolExpr{Name: "xs"},
		}},
		SymbolExpr{Name: "x"},
	}}, ctx, map[string]exprKind{"xs": exprKindValue})
	if err != nil {
		t.Fatal(err)
	}
	rendered = renderIRStmts(got.stmts, "")
	for _, want := range []string{
		"__bind0",
		"flagrt.SeqFirst(__bind0)",
		"flagrt.SeqRest(__bind0)",
		"var x =",
		"var y =",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("let vector missing %q in:\n%s", want, rendered)
		}
	}

	_, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: ".write"},
		SymbolExpr{Name: "f"},
	}}, ctx, map[string]exprKind{"f": exprKindValue})
	if err == nil {
		t.Fatal("expected .write arity error")
	}
	_, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "is"},
		SymbolExpr{Name: "true"},
		IntExpr{Value: 1},
	}}, ctx, nil)
	if err == nil {
		t.Fatal("expected is message type error")
	}
}

func TestFLAGCompileFormUnhandled(t *testing.T) {
	n := 0
	ctx := compileContext{ifTemps: &n, globals: map[string]exprKind{}, functions: map[string]functionDef{}, moduleSymbols: map[string]string{}}
	_, err := flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "catch"},
		SymbolExpr{Name: "Exception"},
		SymbolExpr{Name: "e"},
	}}, ctx, nil)
	if err == nil || !strings.Contains(err.Error(), "catch used outside try") {
		t.Fatalf("expected catch outside try, got %v", err)
	}

	_, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: ".Foo"},
		SymbolExpr{Name: "x"},
	}}, ctx, map[string]exprKind{"x": exprKindValue})
	if !flagUnhandled(err) {
		t.Fatalf("expected unhandled method call, got %v", err)
	}
}

func TestFLAGCompileFormQualifiedSymbol(t *testing.T) {
	n := 0
	ctx := compileContext{
		globals:       map[string]exprKind{},
		functions:     map[string]functionDef{},
		moduleSymbols: map[string]string{"str/upper-case": "string__upper_case"},
		ifTemps:       &n,
	}
	got, err := flagCompileForm(SymbolExpr{Name: "str/upper-case"}, ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got.code != "string__upper_case" {
		t.Fatalf("code %q", got.code)
	}
}

func TestFLAGCompileFormVolatileUpdate(t *testing.T) {
	n := 0
	ctx := compileContext{ifTemps: &n, globals: map[string]exprKind{}, functions: map[string]functionDef{}, moduleSymbols: map[string]string{}}
	form := ListExpr{Elements: []Expr{
		SymbolExpr{Name: "let"},
		VectorExpr{Elements: []Expr{
			MetaExpr{
				Meta:   MapExpr{Entries: []Expr{KeywordExpr{Name: "volatile"}, SymbolExpr{Name: "true"}}},
				Target: SymbolExpr{Name: "v"},
			},
			IntExpr{Value: 1},
		}},
		ListExpr{Elements: []Expr{
			SymbolExpr{Name: "update!"},
			SymbolExpr{Name: "v"},
			IntExpr{Value: 2},
		}},
		SymbolExpr{Name: "v"},
	}}
	got, err := flagCompileForm(form, ctx, nil)
	if err != nil {
		t.Fatalf("flagCompileForm volatile update!: %v", err)
	}
	if got.code != "let_result_1" {
		t.Fatalf("code %q", got.code)
	}
	rendered := renderIRStmts(got.stmts, "")
	for _, want := range []string{
		"var __bind0 = flagrt.NewLong(1)",
		"var v = __bind0",
		"v = flagrt.NewLong(2)",
		"return v",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("missing %q in:\n%s", want, rendered)
		}
	}

	_, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "update!"},
		SymbolExpr{Name: "v"},
		IntExpr{Value: 1},
	}}, ctx, nil)
	if err == nil {
		t.Fatal("expected update! mutability error")
	}
}

func TestFLAGCompileFormsDo(t *testing.T) {
	n := 0
	ctx := compileContext{ifTemps: &n, globals: map[string]exprKind{}, functions: map[string]functionDef{}, moduleSymbols: map[string]string{}}
	got, err := flagCompileForms([]Expr{
		IntExpr{Value: 1},
		IntExpr{Value: 2},
	}, ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if renderIRExpr(irFromGoExpr(got)) != "flagrt.NewLong(2)" {
		t.Fatalf("expr %q", renderIRExpr(irFromGoExpr(got)))
	}
	if !strings.Contains(renderIRStmts(got.stmts, ""), "_ = flagrt.NewLong(1)") {
		t.Fatalf("missing discarded form:\n%s", renderIRStmts(got.stmts, ""))
	}
}

func TestFLAGInternIR(t *testing.T) {
	inline := flagrt.Call(compiler__intern_ir, flagrt.NilValue, flagrt.NewString("Kw_foo"), flagrt.NewString(`flagrt.NewKeyword("foo")`))
	if flagString(flagrt.Get(inline, flagrt.NewKeyword("name"))) != `flagrt.NewKeyword("foo")` {
		t.Fatalf("inline name %q", flagString(flagrt.Get(inline, flagrt.NewKeyword("name"))))
	}

	empty := flagrt.NewMap(
		flagrt.NewKeyword("by-code"), flagrt.NewMap(),
		flagrt.NewKeyword("used"), flagrt.NewMap(),
		flagrt.NewKeyword("order"), flagrt.NewArray(),
	)
	first := flagrt.Call(compiler__intern_ir, empty, flagrt.NewString("Kw_foo"), flagrt.NewString(`flagrt.NewKeyword("foo")`))
	if flagString(flagrt.Get(first, flagrt.NewKeyword("name"))) != "flagKw_foo" {
		t.Fatalf("first name %q", flagString(flagrt.Get(first, flagrt.NewKeyword("name"))))
	}
	again := flagrt.Call(compiler__intern_ir, flagrt.Get(first, flagrt.NewKeyword("constants")), flagrt.NewString("Kw_foo"), flagrt.NewString(`flagrt.NewKeyword("foo")`))
	if flagString(flagrt.Get(again, flagrt.NewKeyword("name"))) != "flagKw_foo" {
		t.Fatalf("dedup name %q", flagString(flagrt.Get(again, flagrt.NewKeyword("name"))))
	}
	other := flagrt.Call(compiler__intern_ir, flagrt.Get(again, flagrt.NewKeyword("constants")), flagrt.NewString("Kw_foo"), flagrt.NewString(`flagrt.NewKeyword("bar")`))
	if flagString(flagrt.Get(other, flagrt.NewKeyword("name"))) != "flagKw_foo_1" {
		t.Fatalf("collision name %q", flagString(flagrt.Get(other, flagrt.NewKeyword("name"))))
	}
}

func TestFLAGCompileProgramPrinter(t *testing.T) {
	src, err := flagCompileProgram("main", compileResult{
		functions: []functionDef{{
			flagName: "flag_main",
			goName:   "flag_main",
			irDecls:  []IRStmt{IRRawStmt{Code: "func flag_main() flagrt.Value {\n\treturn flagrt.NilValue\n}\n"}},
		}},
		vars:  []varDef{{goName: "flagKw_foo", expr: `flagrt.NewKeyword("foo")`}},
		stmts: []mainStmt{{code: "flagrt.NewLong(1)", kind: exprKindValue}},
	}, true)
	if err != nil {
		t.Fatal(err)
	}
	got := string(src)
	for _, want := range []string{
		"package main",
		"flagrt \"flag-lang/runtime\"",
		`var flagKw_foo = flagrt.NewKeyword("foo")`,
		"func flag_main()",
		"_ = flagrt.NewLong(1)",
		"_ = flagrt.Call(flag_main)",
		"_ = strings.HasSuffix",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in:\n%s", want, got)
		}
	}

	lib, err := flagCompileProgram("pkg", compileResult{
		functions: []functionDef{{
			doc:     "add two numbers",
			goName:  "Add",
			irDecls: []IRStmt{IRRawStmt{Code: "func Add() flagrt.Value {\n\treturn flagrt.NilValue\n}\n"}},
		}},
	}, false)
	if err != nil {
		t.Fatal(err)
	}
	got = string(lib)
	if strings.Contains(got, "func main()") {
		t.Fatalf("library emit included main:\n%s", got)
	}
	if !strings.Contains(got, "// add two numbers") {
		t.Fatalf("missing doc:\n%s", got)
	}
}

func TestFLAGDefGoIdent(t *testing.T) {
	got := flagString(flagrt.Call(
		compiler__def_go_ident,
		flagrt.NewMap(
			flagKW("namespace"), flagrt.NewString("compiler"),
			flagKW("package-mode"), flagrt.NewBool(false),
		),
		flagrt.NewString("foo-bar"),
	))
	if got != "compiler__foo_bar" {
		t.Fatalf("inlined ident %q", got)
	}
	got = flagString(flagrt.Call(
		compiler__module_go_ident,
		flagrt.NewString("a.b"),
		flagrt.NewString("x"),
	))
	if got != "a_b__x" {
		t.Fatalf("dotted ns ident %q", got)
	}
	got = flagString(flagrt.Call(
		compiler__def_go_ident,
		flagrt.NewMap(
			flagKW("package-mode"), flagrt.NewBool(true),
			flagKW("exported-names"), flagrt.NewMap(flagrt.NewString("add"), flagrt.NewBool(true)),
		),
		flagrt.NewString("add"),
	))
	if got != "Add" {
		t.Fatalf("exported ident %q", got)
	}
}

func TestFLAGCompileModuleBody(t *testing.T) {
	ast, err := ParseFile("(defn add [x] x)\n(def answer 42)\n(deftest ok (is true))\n")
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	ctx := compileContext{
		namespace:     "compiler",
		ifTemps:       &n,
		functions:     map[string]functionDef{},
		globals:       map[string]exprKind{},
		moduleSymbols: map[string]string{},
	}
	result, defined, err := flagCompileModuleBody(ast.Forms, &ctx, false)
	if err != nil {
		t.Fatal(err)
	}
	if defined["add"] != "compiler__add" || defined["answer"] != "compiler__answer" {
		t.Fatalf("defined %v", defined)
	}
	if len(result.functions) != 2 {
		t.Fatalf("functions %d", len(result.functions))
	}
	if result.functions[0].goName != "compiler__add" || result.functions[0].arityName != "compiler__add_arity_1" {
		t.Fatalf("defn %+v", result.functions[0])
	}
	src := renderIRStmts(result.functions[0].irDecls, "")
	if result.functions[0].body != "x" {
		t.Fatalf("defn body %q", result.functions[0].body)
	}
	if !strings.Contains(src, "func compiler__add_arity_1(x flagrt.Value) flagrt.Value") {
		t.Fatalf("missing arity func:\n%s", src)
	}
	if len(result.vars) != 2 {
		t.Fatalf("vars %d", len(result.vars))
	}
	if result.vars[1].goName != "compiler__answer" || result.vars[1].expr != "flagrt.NewLong(42)" {
		t.Fatalf("def %+v", result.vars[1])
	}
	if len(result.tests) != 1 || result.tests[0].label != "ok" || result.functions[1].arityName != "compiler__ok" {
		t.Fatalf("deftest fn=%+v tests=%+v", result.functions[1], result.tests)
	}
	if ctx.moduleSymbols["add"] != "compiler__add" || ctx.moduleSymbols["compiler/add"] != "compiler__add" {
		t.Fatalf("module symbols %v", ctx.moduleSymbols)
	}
}

func TestFLAGCompileModuleBodyDeclare(t *testing.T) {
	ast, err := ParseFile("(declare foo bar)")
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	ctx := compileContext{
		namespace:     "compiler",
		ifTemps:       &n,
		functions:     map[string]functionDef{},
		globals:       map[string]exprKind{},
		moduleSymbols: map[string]string{},
	}
	_, defined, err := flagCompileModuleBody(ast.Forms, &ctx, false)
	if err != nil {
		t.Fatal(err)
	}
	if defined["foo"] != "compiler__foo" || defined["bar"] != "compiler__bar" {
		t.Fatalf("declare defined %v", defined)
	}
	if ctx.moduleSymbols["foo"] != "compiler__foo" || ctx.moduleSymbols["compiler/foo"] != "compiler__foo" {
		t.Fatalf("declare module %v", ctx.moduleSymbols)
	}
}

func TestFLAGCompileModuleBodyDefrecord(t *testing.T) {
	ast, err := ParseFile("(defrecord Point [x y])")
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	ctx := compileContext{
		namespace:     "compiler",
		ifTemps:       &n,
		functions:     map[string]functionDef{},
		globals:       map[string]exprKind{},
		moduleSymbols: map[string]string{},
		constants:     newConstantInterner(),
	}
	result, defined, err := flagCompileModuleBody(ast.Forms, &ctx, false)
	if err != nil {
		t.Fatal(err)
	}
	if defined["->Point"] == "" || defined["map->Point"] == "" {
		t.Fatalf("defined %v", defined)
	}
	if len(result.typeDecls) != 1 || !strings.Contains(result.typeDecls[0], "type Point struct") {
		t.Fatalf("typeDecls %v", result.typeDecls)
	}
	if !strings.Contains(result.typeDecls[0], "`flag:\"x\"`") {
		t.Fatalf("missing field tag:\n%s", result.typeDecls[0])
	}
	if len(result.functions) != 2 {
		t.Fatalf("functions %d", len(result.functions))
	}
	src := renderIRStmts(result.functions[0].irDecls, "")
	if !strings.Contains(src, "flagrt.NewRecord(Point{X: x, Y: y})") {
		t.Fatalf("ctor src:\n%s", src)
	}
}

func TestFLAGCompileModuleBodyPrintlnPrint(t *testing.T) {
	ast, err := ParseFile("(println \"Hello\")\n(print 42)\n")
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	ctx := compileContext{
		namespace:     "hello.core",
		ifTemps:       &n,
		functions:     map[string]functionDef{},
		globals:       map[string]exprKind{},
		moduleSymbols: map[string]string{},
		constants:     newConstantInterner(),
	}
	result, _, err := flagCompileModuleBody(ast.Forms, &ctx, true)
	if err != nil {
		t.Fatal(err)
	}
	if !result.needsFmt {
		t.Fatal("expected needsFmt")
	}
	if len(result.stmts) != 2 {
		t.Fatalf("stmts %d", len(result.stmts))
	}
	if result.stmts[0].code != `fmt.Println(flagrt.Str("Hello"))` {
		t.Fatalf("println %q", result.stmts[0].code)
	}
	if result.stmts[1].code != "fmt.Print(flagrt.ValueToAny(flagrt.NewLong(42)))" {
		t.Fatalf("print %q", result.stmts[1].code)
	}
	_, _, err = flagCompileModuleBody(ast.Forms, &ctx, false)
	if err == nil || !strings.Contains(err.Error(), "top-level expressions are only allowed in the entry module") {
		t.Fatalf("expected top-level error, got %v", err)
	}
}

func TestFLAGCompileModuleBodyGoInterface(t *testing.T) {
	ast, err := ParseFile("(defn square [x] (* x x))\n(go-interface square double [double])\n")
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	ctx := compileContext{
		namespace:     "compiler",
		ifTemps:       &n,
		functions:     map[string]functionDef{},
		globals:       map[string]exprKind{},
		moduleSymbols: map[string]string{},
		constants:     newConstantInterner(),
	}
	result, _, err := flagCompileModuleBody(ast.Forms, &ctx, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.functions) != 2 {
		t.Fatalf("functions %d", len(result.functions))
	}
	wrapper := result.functions[1]
	if wrapper.goName != "square_go" || wrapper.goSignature != "(p1 float64) float64" {
		t.Fatalf("wrapper %+v", wrapper)
	}
	src := renderFunctionDef(wrapper)
	for _, want := range []string{
		"func square_go(p1 float64) float64 {",
		"args := make([]flagrt.Value, 0, 1)",
		"args = append(args, flagrt.NewDouble(p1))",
		"result := flagrt.Call(compiler__square, args...)",
		"return result.Double()",
	} {
		if !strings.Contains(src, want) {
			t.Fatalf("missing %q in:\n%s", want, src)
		}
	}
}

func TestFLAGTypeHintAndRecordField(t *testing.T) {
	got := flagString(flagrt.Call(compiler__type_hint_to_go_type, flagrt.NewString("long")))
	if got != "int64" {
		t.Fatalf("long hint %q", got)
	}
	got = flagString(flagrt.Call(compiler__exported_record_field_name, flagrt.NewString("start-line")))
	if got != "StartLine" {
		t.Fatalf("field %q", got)
	}
	got = flagString(flagrt.Call(compiler__exported_record_field_name, flagrt.NewString("eof")))
	if got != "EOF" {
		t.Fatalf("eof field %q", got)
	}
}

func TestFLAGCompileFormCatchOutsideTry(t *testing.T) {
	n := 0
	ctx := compileContext{ifTemps: &n, globals: map[string]exprKind{}, functions: map[string]functionDef{}, moduleSymbols: map[string]string{}}
	_, err := flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "catch"},
		SymbolExpr{Name: "Exception"},
		SymbolExpr{Name: "e"},
		SymbolExpr{Name: "e"},
	}}, ctx, nil)
	if err == nil || !strings.Contains(err.Error(), "catch used outside try") {
		t.Fatalf("expected catch outside try, got %v", err)
	}
	_, err = flagCompileForm(ListExpr{Elements: []Expr{SymbolExpr{Name: "finally"}}}, ctx, nil)
	if err == nil || !strings.Contains(err.Error(), "finally used outside try") {
		t.Fatalf("expected finally outside try, got %v", err)
	}
}

func TestLetExprToGoUsesResultVar(t *testing.T) {
	n := 0
	ctx := compileContext{ifTemps: &n}
	got, err := exprToGo(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "let"},
		VectorExpr{Elements: []Expr{SymbolExpr{Name: "a"}, IntExpr{Value: 1}}},
		SymbolExpr{Name: "a"},
	}}, ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got.code != "let_result_1" {
		t.Fatalf("code %q", got.code)
	}
	rendered := renderIRStmts(got.stmts, "\t")
	if strings.Contains(rendered, "func()") {
		t.Fatalf("let still used an IIFE:\n%s", rendered)
	}
	if !strings.Contains(rendered, "var let_result_1 flagrt.Value") {
		t.Fatalf("missing result var:\n%s", rendered)
	}
	if !strings.Contains(rendered, "{\n") {
		t.Fatalf("missing block:\n%s", rendered)
	}
}

func TestFLAGRenderIRMatchesGo(t *testing.T) {
	cases := []IRExpr{
		IRIdent{Name: "x"},
		IRString{Value: "hi"},
		IRString{Value: "a\"b"},
		IRInt{Value: 42},
		IRInt{Value: -7},
		IRSelector{Pkg: "flagrt", Name: "NewLong"},
		IRSelector{Name: "println"},
		IRSelector{Pkg: "flagrt", Name: "NilValue"},
		rtCall("NewLong", IRInt{Value: 42}),
		rtCall("NewString", IRString{Value: "hi"}),
		IRRaw{Code: "1e2"},
		IRIndex{X: IRIdent{Name: "args"}, Index: IRInt{Value: 0}},
		IRSlice{X: IRIdent{Name: "args"}, Low: IRInt{Value: 2}},
		IRCall{Fun: IRSelector{Pkg: "flagrt", Name: "Call"}, Args: []IRExpr{IRIdent{Name: "f"}, IRSpread{Expr: IRIdent{Name: "xs"}}}},
		IRUnary{Op: "!", X: rtCall("IsTruthy", IRIdent{Name: "x"})},
		IRBinary{Op: "!=", Left: identCall("len", IRIdent{Name: "args"}), Right: IRInt{Value: 1}},
		IRFuncLit{Params: "args ...flagrt.Value", Result: "flagrt.Value", Body: []IRStmt{IRReturn{Expr: IRIdent{Name: "x"}}}},
	}
	for _, ir := range cases {
		got, ok := flagRenderIRExpr(ir)
		if !ok {
			t.Fatalf("FLAG render failed for %#v", ir)
		}
		want := renderIRExprGo(ir)
		if got != want {
			t.Fatalf("FLAG %q want %q for %#v", got, want, ir)
		}
	}

	stmts := []IRStmt{
		IRVar{Name: "x", Type: "flagrt.Value"},
		IRAssign{Name: "x", Expr: IRIdent{Name: "y"}},
		IRReturn{Expr: IRIdent{Name: "x"}},
		IRFuncDecl{Name: "foo_arity_1", Params: "x flagrt.Value", Result: "flagrt.Value", Body: []IRStmt{IRReturn{Expr: IRIdent{Name: "x"}}}},
	}
	got, ok := flagRenderIRStmts(stmts, "")
	if !ok {
		t.Fatal("FLAG render stmts failed")
	}
	want := renderIRStmtsGo(stmts, "")
	if got != want {
		t.Fatalf("FLAG stmts %q want %q", got, want)
	}
}

func TestFLAGRenderFunctionDef(t *testing.T) {
	cases := []struct {
		name string
		fn   functionDef
		want string
	}{
		{
			name: "go-signature",
			fn:   functionDef{goName: "wrap", goSignature: "(p1 float64) float64", body: "\treturn p1\n"},
			want: "func wrap(p1 float64) float64 {\n\treturn p1\n}\n",
		},
		{
			name: "ir-decls",
			fn: functionDef{irDecls: []IRStmt{IRFuncDecl{
				Name: "foo", Result: "flagrt.Value",
				Body: []IRStmt{IRReturn{Expr: IRSelector{Pkg: "flagrt", Name: "NilValue"}}},
			}}},
			want: "func foo() flagrt.Value {\n\treturn flagrt.NilValue\n}\n",
		},
		{
			name: "rest",
			fn:   functionDef{goName: "sum", variadicName: "sum_variadic", hasRest: true, params: []string{"xs"}, body: "xs"},
			want: "func sum_variadic(args ...flagrt.Value) flagrt.Value {\n" +
				"\tif len(args) < 1 {\n" +
				"\t\tpanic(\"sum expects at least 1 arguments\")\n" +
				"\t}\n" +
				"\txs := args[0]\n" +
				"\treturn xs\n" +
				"}\n",
		},
		{
			name: "arity-plus-variadic",
			fn:   functionDef{goName: "add", arityName: "add_arity_1", variadicName: "add_variadic", params: []string{"x"}, body: "x"},
			want: "func add_arity_1(x flagrt.Value) flagrt.Value {\n" +
				"\treturn x\n" +
				"}\n\n" +
				"func add_variadic(args ...flagrt.Value) flagrt.Value {\n" +
				"\tif len(args) != 1 {\n" +
				"\t\tpanic(\"add expects exactly 1 arguments\")\n" +
				"\t}\n" +
				"\treturn add_arity_1(args[0])\n" +
				"}\n",
		},
		{
			name: "multi-arity",
			fn: functionDef{flagName: "add", variadicName: "add_variadic", arities: []arityDef{
				{arityName: "add_arity_0", body: "flagrt.NewLong(0)"},
				{arityName: "add_arity_1", params: []string{"x"}, body: "x"},
			}},
			want: "func add_arity_0() flagrt.Value {\n" +
				"\treturn flagrt.NewLong(0)\n" +
				"}\n\n" +
				"func add_arity_1(x flagrt.Value) flagrt.Value {\n" +
				"\treturn x\n" +
				"}\n\n" +
				"func add_variadic(args ...flagrt.Value) flagrt.Value {\n" +
				"\tswitch len(args) {\n" +
				"\tcase 0:\n" +
				"\t\treturn add_arity_0()\n" +
				"\tcase 1:\n" +
				"\t\treturn add_arity_1(args[0])\n" +
				"\tdefault:\n" +
				"\t\tpanic(\"add expects 0 or 1 arguments\")\n" +
				"\t}\n" +
				"}\n",
		},
		{
			name: "unused-rest-local-init",
			fn:   functionDef{goName: "drop", variadicName: "drop_variadic", hasRest: true, params: []string{"_"}, localInits: []string{"\tvar tmp flagrt.Value\n"}, body: "flagrt.NilValue"},
			want: "func drop_variadic(args ...flagrt.Value) flagrt.Value {\n" +
				"\tif len(args) < 1 {\n" +
				"\t\tpanic(\"drop expects at least 1 arguments\")\n" +
				"\t}\n" +
				"\t_ = args[0]\n" +
				"\tvar tmp flagrt.Value\n" +
				"\t_ = tmp\n" +
				"\treturn flagrt.NilValue\n" +
				"}\n",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := flagRenderFunctionDef(tc.fn)
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Fatalf("got:\n%s\nwant:\n%s", got, tc.want)
			}
			if renderFunctionDef(tc.fn) != tc.want {
				t.Fatalf("production printer:\n%s\nwant:\n%s", renderFunctionDef(tc.fn), tc.want)
			}
		})
	}
}

func TestFLAGRenderFunctionLiterals(t *testing.T) {
	zero := functionDef{goName: "id", arityName: "id_arity_0", variadicName: "id_variadic", body: "flagrt.NilValue"}
	one := functionDef{goName: "add", arityName: "add_arity_1", variadicName: "add_variadic", params: []string{"x"}, body: "x", bodyPrelude: "\t_ = x\n"}
	rest := functionDef{goName: "sum", variadicName: "sum_variadic", hasRest: true, params: []string{"xs"}, body: "xs"}
	unused := functionDef{goName: "drop", variadicName: "drop_variadic", hasRest: true, params: []string{"_"}, localInits: []string{"\tvar tmp flagrt.Value\n"}, body: "flagrt.NilValue"}
	multi := functionDef{flagName: "add", variadicName: "add_variadic", arities: []arityDef{
		{arityName: "add_arity_0", body: "flagrt.NewLong(0)"},
		{arityName: "add_arity_1", params: []string{"x"}, body: "x"},
		{arityName: "add_arity_2", params: []string{"x", "y"}, body: "y"},
	}}
	if got, want := renderDirectFunctionType(zero), "func() flagrt.Value"; got != want {
		t.Fatalf("zero type %q want %q", got, want)
	}
	if got, want := renderDirectFunctionType(one), "func(flagrt.Value) flagrt.Value"; got != want {
		t.Fatalf("one type %q want %q", got, want)
	}
	if got, want := renderDirectFunctionLiteral(one), "func(x flagrt.Value) flagrt.Value {\n\t_ = x\n\treturn x\n}"; got != want {
		t.Fatalf("literal got:\n%s\nwant:\n%s", got, want)
	}
	if got, want := renderDirectFunctionLiteral(zero), "func() flagrt.Value {\n\treturn flagrt.NilValue\n}"; got != want {
		t.Fatalf("zero literal got:\n%s\nwant:\n%s", got, want)
	}
	if got, want := renderVariadicFunctionLiteral(one), "func(args ...flagrt.Value) flagrt.Value {\n"+
		"\tif len(args) != 1 {\n"+
		"\t\tpanic(\"add expects exactly 1 arguments\")\n"+
		"\t}\n"+
		"\treturn add_arity_1(args[0])\n"+
		"}"; got != want {
		t.Fatalf("variadic got:\n%s\nwant:\n%s", got, want)
	}
	if got, want := renderVariadicFunctionLiteral(zero), "func(args ...flagrt.Value) flagrt.Value {\n"+
		"\tif len(args) != 0 {\n"+
		"\t\tpanic(\"id expects exactly 0 arguments\")\n"+
		"\t}\n"+
		"\treturn id_arity_0()\n"+
		"}"; got != want {
		t.Fatalf("zero variadic got:\n%s\nwant:\n%s", got, want)
	}
	if got, want := renderFunctionLiteral(rest), "func(args ...flagrt.Value) flagrt.Value {\n"+
		"\tif len(args) < 1 {\n"+
		"\t\tpanic(\"sum expects at least 1 arguments\")\n"+
		"\t}\n"+
		"\txs := args[0]\n"+
		"\treturn xs\n"+
		"}"; got != want {
		t.Fatalf("fn literal rest got:\n%s\nwant:\n%s", got, want)
	}
	if got, want := renderFunctionLiteral(unused), "func(args ...flagrt.Value) flagrt.Value {\n"+
		"\tif len(args) < 1 {\n"+
		"\t\tpanic(\"drop expects at least 1 arguments\")\n"+
		"\t}\n"+
		"\t_ = args[0]\n"+
		"\tvar tmp flagrt.Value\n"+
		"\t_ = tmp\n"+
		"\treturn flagrt.NilValue\n"+
		"}"; got != want {
		t.Fatalf("fn literal unused got:\n%s\nwant:\n%s", got, want)
	}
	if got, want := renderMultiArityVariadicLiteral(multi), "func(args ...flagrt.Value) flagrt.Value {\n"+
		"\tswitch len(args) {\n"+
		"\tcase 0:\n"+
		"\t\treturn add_arity_0()\n"+
		"\tcase 1:\n"+
		"\t\treturn add_arity_1(args[0])\n"+
		"\tcase 2:\n"+
		"\t\treturn add_arity_2(args[0], args[1])\n"+
		"\tdefault:\n"+
		"\t\tpanic(\"add expects 0, 1, or 2 arguments\")\n"+
		"\t}\n"+
		"}"; got != want {
		t.Fatalf("multi got:\n%s\nwant:\n%s", got, want)
	}
}
