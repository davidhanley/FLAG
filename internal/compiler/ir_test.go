package compiler

import "testing"

func TestRenderIRExpr(t *testing.T) {
	cases := []struct {
		name string
		ir   IRExpr
		want string
	}{
		{"ident", IRIdent{Name: "x"}, "x"},
		{"string", IRString{Value: "hi"}, `"hi"`},
		{"string-escape", IRString{Value: "a\"b"}, `"a\"b"`},
		{"int", IRInt{Value: 42}, "42"},
		{"negative-int", IRInt{Value: -7}, "-7"},
		{"selector", IRSelector{Pkg: "flagrt", Name: "NewLong"}, "flagrt.NewLong"},
		{"bare-selector", IRSelector{Name: "println"}, "println"},
		{"call-no-args", IRCall{Fun: IRSelector{Pkg: "flagrt", Name: "NilValue"}}, "flagrt.NilValue()"},
		{
			"new-long",
			rtCall("NewLong", IRInt{Value: 42}),
			"flagrt.NewLong(42)",
		},
		{
			"new-bool",
			rtCall("NewBool", IRRaw{Code: "true"}),
			"flagrt.NewBool(true)",
		},
		{
			"new-string",
			rtCall("NewString", IRString{Value: "hi"}),
			`flagrt.NewString("hi")`,
		},
		{
			"new-ratio",
			rtCall("NewRatio", IRInt{Value: 5}, IRInt{Value: 6}),
			"flagrt.NewRatio(5, 6)",
		},
		{"raw", IRRaw{Code: "1e2"}, "1e2"},
		{
			"nested-raw-float",
			rtCall("NewDouble", IRRaw{Code: "1e2"}),
			"flagrt.NewDouble(1e2)",
		},
		{
			"nested-add",
			IRCall{
				Fun: parseGoFun("flagrt.Add"),
				Args: []IRExpr{
					rtCall("Add", IRIdent{Name: "a"}, IRIdent{Name: "b"}),
					IRIdent{Name: "c"},
				},
			},
			"flagrt.Add(flagrt.Add(a, b), c)",
		},
		{
			"runtime-call",
			rtCall("Call", IRIdent{Name: "f"}, IRIdent{Name: "x"}),
			"flagrt.Call(f, x)",
		},
		{
			"self-arity-call",
			identCall("flagFoo_1", IRIdent{Name: "x"}),
			"flagFoo_1(x)",
		},
		{
			"nested-predicate",
			rtCall("NewBool", rtCall("IsEmpty", IRIdent{Name: "xs"})),
			"flagrt.NewBool(flagrt.IsEmpty(xs))",
		},
		{
			"int64-count",
			rtCall("NewLong", identCall("int64", rtCall("Count", IRIdent{Name: "xs"}))),
			"flagrt.NewLong(int64(flagrt.Count(xs)))",
		},
		{
			"strings-has-suffix",
			selectorCall("strings", "HasSuffix", IRIdent{Name: "s"}, IRString{Value: ".go"}),
			`strings.HasSuffix(s, ".go")`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := renderIRExpr(tc.ir)
			if got != tc.want {
				t.Fatalf("got %q want %q", got, tc.want)
			}
		})
	}
}

func TestInternIRHoistsWhenEnabled(t *testing.T) {
	ctx := compileContext{constants: newConstantInterner()}
	ir := ctx.internIR("Kw_foo", rtCall("NewKeyword", IRString{Value: "foo"}))
	ident, ok := ir.(IRIdent)
	if !ok {
		t.Fatalf("expected interned ident, got %#v", ir)
	}
	if ident.Name != "flagKw_foo" {
		t.Fatalf("got ident %q", ident.Name)
	}
	if renderIRExpr(ir) != "flagKw_foo" {
		t.Fatalf("render %q", renderIRExpr(ir))
	}
}

func TestInternIRStaysInlineWithoutInterner(t *testing.T) {
	ctx := compileContext{}
	ir := ctx.internIR("Kw_foo", rtCall("NewKeyword", IRString{Value: "foo"}))
	want := `flagrt.NewKeyword("foo")`
	if renderIRExpr(ir) != want {
		t.Fatalf("got %q want %q", renderIRExpr(ir), want)
	}
}

func TestFromIRSetsCodeAndKind(t *testing.T) {
	got := fromIR(rtCall("NewLong", IRInt{Value: 7}), exprKindValue)
	if got.code != "flagrt.NewLong(7)" {
		t.Fatalf("code %q", got.code)
	}
	if got.kind != exprKindValue {
		t.Fatalf("kind %v", got.kind)
	}
	if got.ir == nil {
		t.Fatal("expected ir to be set")
	}
}

func TestIRFromGoExprPrefersIR(t *testing.T) {
	src := fromIR(IRIdent{Name: "x"}, exprKindValue)
	src.code = "stale"
	got := irFromGoExpr(src)
	ident, ok := got.(IRIdent)
	if !ok || ident.Name != "x" {
		t.Fatalf("got %#v", got)
	}
}

func TestIRFromGoExprFallsBackToRaw(t *testing.T) {
	got := irFromGoExpr(goExpr{code: "foo(1)", kind: exprKindValue})
	raw, ok := got.(IRRaw)
	if !ok || raw.Code != "foo(1)" {
		t.Fatalf("got %#v", got)
	}
}

func TestRenderCollectionCalls(t *testing.T) {
	got := renderIRExpr(rtCall("NewArray", rtCall("NewLong", IRInt{Value: 1}), rtCall("NewLong", IRInt{Value: 2})))
	want := "flagrt.NewArray(flagrt.NewLong(1), flagrt.NewLong(2))"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
	got = renderIRExpr(rtCall("NewMap", rtCall("NewKeyword", IRString{Value: "a"}), rtCall("NewLong", IRInt{Value: 1})))
	want = `flagrt.NewMap(flagrt.NewKeyword("a"), flagrt.NewLong(1))`
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestRenderIIFEAndStatements(t *testing.T) {
	got := renderIRExpr(iife("bool",
		IRIfStmt{
			Cond: IRIdent{Name: "ok"},
			Then: []IRStmt{IRReturn{Expr: IRRaw{Code: "true"}}},
		},
		IRReturn{Expr: IRRaw{Code: "false"}},
	))
	want := "func() bool {\n\tif ok {\n\t\treturn true\n\t}\n\treturn false\n}()"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestRenderSequentialDo(t *testing.T) {
	got := renderIRExpr(valueIIFE(
		IRExprStmt{Expr: IRIdent{Name: "a"}, Discard: true},
		IRReturn{Expr: IRIdent{Name: "b"}},
	))
	want := "func() flagrt.Value {\n\t_ = a\n\treturn b\n}()"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestRenderDeferAndVar(t *testing.T) {
	got := renderIRStmts([]IRStmt{
		IRVar{Name: "x", Type: "flagrt.Value"},
		IRDefer{Expr: rtCall("Call", IRIdent{Name: "f"})},
		IRAssign{Name: "x", Expr: IRIdent{Name: "y"}},
		IRReturn{Expr: IRIdent{Name: "x"}},
	}, "\t")
	want := "\tvar x flagrt.Value\n\tdefer flagrt.Call(f)\n\tx = y\n\treturn x\n"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestFromStmtDefer(t *testing.T) {
	got := fromStmt(IRDefer{Expr: rtCall("Call", IRIdent{Name: "f"})}, exprKindDefer)
	if got.code != "defer flagrt.Call(f)" {
		t.Fatalf("code %q", got.code)
	}
	if got.kind != exprKindDefer {
		t.Fatalf("kind %v", got.kind)
	}
	if got.stmt == nil {
		t.Fatal("expected stmt")
	}
}

func TestQuotedLiteralToIR(t *testing.T) {
	ir, err := quotedLiteralToIR(QuotedListExpr{Elements: []Expr{IntExpr{Value: 1}, KeywordExpr{Name: "a"}}})
	if err != nil {
		t.Fatal(err)
	}
	got := renderIRExpr(ir)
	want := `flagrt.NewList(flagrt.NewLong(1), flagrt.NewKeyword("a"))`
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestQuotedLiteralToIRNested(t *testing.T) {
	ir, err := quotedLiteralToIR(VectorExpr{Elements: []Expr{
		QuotedListExpr{Elements: []Expr{SymbolExpr{Name: "x"}}},
	}})
	if err != nil {
		t.Fatal(err)
	}
	got := renderIRExpr(ir)
	want := `flagrt.NewArray(flagrt.NewList(flagrt.NewSymbol("x")))`
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestQuotedLiteralToIRUnsupported(t *testing.T) {
	_, err := quotedLiteralToIR(HashFnExpr{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseGoFun(t *testing.T) {
	sel, ok := parseGoFun("flagrt.Add").(IRSelector)
	if !ok || sel.Pkg != "flagrt" || sel.Name != "Add" {
		t.Fatalf("selector %#v", parseGoFun("flagrt.Add"))
	}
	ident, ok := parseGoFun("flagFoo_1").(IRIdent)
	if !ok || ident.Name != "flagFoo_1" {
		t.Fatalf("ident %#v", parseGoFun("flagFoo_1"))
	}
}

func TestRenderIndexSliceSpread(t *testing.T) {
	got := renderIRExpr(IRIndex{X: IRIdent{Name: "args"}, Index: IRInt{Value: 0}})
	if got != "args[0]" {
		t.Fatalf("index %q", got)
	}
	got = renderIRExpr(IRSlice{X: IRIdent{Name: "args"}, Low: IRInt{Value: 2}})
	if got != "args[2:]" {
		t.Fatalf("slice %q", got)
	}
	got = renderIRExpr(rtCall("NewArray", IRSpread{Expr: IRSlice{X: IRIdent{Name: "args"}, Low: IRInt{Value: 1}}}))
	if got != "flagrt.NewArray(args[1:]...)" {
		t.Fatalf("spread %q", got)
	}
}

func TestRenderUnaryBinary(t *testing.T) {
	got := renderIRExpr(IRUnary{Op: "!", X: rtCall("IsTruthy", IRIdent{Name: "x"})})
	if got != "!(flagrt.IsTruthy(x))" {
		t.Fatalf("unary %q", got)
	}
	got = renderIRExpr(IRBinary{
		Op:    "!=",
		Left:  identCall("len", IRIdent{Name: "args"}),
		Right: IRInt{Value: 1},
	})
	if got != "len(args) != 1" {
		t.Fatalf("binary %q", got)
	}
}

func TestRenderFuncLitParams(t *testing.T) {
	got := renderIRExpr(IRFuncLit{
		Params: "args ...flagrt.Value",
		Result: "flagrt.Value",
		Body: []IRStmt{
			IRDefine{Names: []string{"x"}, Expr: IRIndex{X: IRIdent{Name: "args"}, Index: IRInt{Value: 0}}},
			IRReturn{Expr: IRIdent{Name: "x"}},
		},
	})
	want := "func(args ...flagrt.Value) flagrt.Value {\n\tx := args[0]\n\treturn x\n}"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestRenderMapCatBinding(t *testing.T) {
	got := renderIRExpr(mapCatBindingIR("x", "x", "for binding expects exactly one value", IRIdent{Name: "rest"}, IRIdent{Name: "coll"}))
	want := "func() flagrt.Value {\n\treturn flagrt.MapCat(flagrt.NewFunction(func(args ...flagrt.Value) flagrt.Value {\n\tif len(args) != 1 {\n\t\tpanic(\"for binding expects exactly one value\")\n\t}\n\tx := args[0]\n\treturn rest\n}), coll)\n}()"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestTruthyIR(t *testing.T) {
	got, err := truthyIR(fromIR(IRIdent{Name: "ok"}, exprKindBool))
	if err != nil {
		t.Fatal(err)
	}
	if renderIRExpr(got) != "ok" {
		t.Fatalf("bool %q", renderIRExpr(got))
	}
	got, err = truthyIR(fromIR(IRIdent{Name: "v"}, exprKindValue))
	if err != nil {
		t.Fatal(err)
	}
	if renderIRExpr(got) != "flagrt.IsTruthy(v)" {
		t.Fatalf("value %q", renderIRExpr(got))
	}
}
