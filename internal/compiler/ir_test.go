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
