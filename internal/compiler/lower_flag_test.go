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
		{"nil", SymbolExpr{Name: "nil"}, "flagrt.NilValue()"},
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
	if got := renderIRExpr(ir); got != "flagrt.NewBool(flagrt.Eq(flagrt.NewLong(1), flagrt.NewLong(2)))" {
		t.Fatalf("eq %q", got)
	}
	ir, err = flagEqToIR([]IRExpr{IRIdent{Name: "a"}, IRIdent{Name: "b"}, IRIdent{Name: "c"}})
	if err != nil {
		t.Fatal(err)
	}
	if got := renderIRExpr(ir); got != "flagrt.NewBool(flagrt.Eq(a, b) && flagrt.Eq(b, c))" {
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
	if got.code != `flagrt.BuiltinFunction("first")` {
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
	if !strings.Contains(got, "return flagrt.NilValue()") {
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
	want := "func() flagrt.Value {\n\tflagrt.Throw(err)\n\treturn flagrt.NilValue()\n}()"
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
	want := "func() flagrt.Value {\n\t_ = flagrt.DoAll(loop)\n\treturn flagrt.NilValue()\n}()"
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

	_, _, _, _, err = flagBindParams(
		VectorExpr{Elements: []Expr{MapExpr{Entries: []Expr{KeywordExpr{Name: "a"}, SymbolExpr{Name: "a"}}}}},
		"fn",
		nil,
	)
	if !flagUnhandled(err) {
		t.Fatalf("expected unhandled map param, got %v", err)
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
		globals:        map[string]exprKind{},
		functions:      map[string]functionDef{},
		moduleSymbols:  map[string]string{},
		ifTemps:        &n,
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
		"if flagrt.IsTruthy(flagrt.NewBool(true))",
		"if_result_1 = flagrt.NewLong(1)",
		"if_result_1 = flagrt.NewLong(2)",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("missing %q in:\n%s", want, rendered)
		}
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
	if renderIRExpr(irFromGoExpr(got)) != "flagrt.NewBool(flagrt.Eq(flagrt.NewLong(1), flagrt.NewLong(2)))" {
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
	if renderIRExpr(irFromGoExpr(got)) != "flagrt.NewBool(flagrt.Eq(a, b) && flagrt.Eq(b, c))" {
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
	if renderIRExpr(irFromGoExpr(got)) != `flagrt.NewBool(flagrt.Eq(flagrt.NewLong(1), flagrt.NewString("x")))` {
		t.Fatalf("eq-string %q", renderIRExpr(irFromGoExpr(got)))
	}

	_, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "="},
		IntExpr{Value: 1},
	}}, ctx, nil)
	if err == nil {
		t.Fatal("expected arity error")
	}

	_, err = flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "="},
		SymbolExpr{Name: "a"},
		SymbolExpr{Name: "b"},
	}}, ctx, map[string]exprKind{"a": exprKindBool, "b": exprKindValue})
	if err == nil {
		t.Fatal("expected value-kind error")
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

func TestFLAGCompileFormUnhandled(t *testing.T) {
	n := 0
	ctx := compileContext{ifTemps: &n, globals: map[string]exprKind{}, functions: map[string]functionDef{}, moduleSymbols: map[string]string{}}
	_, err := flagCompileForm(ListExpr{Elements: []Expr{
		SymbolExpr{Name: "loop"},
		VectorExpr{},
		IntExpr{Value: 1},
	}}, ctx, nil)
	if !flagUnhandled(err) {
		t.Fatalf("expected unhandled loop, got %v", err)
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
