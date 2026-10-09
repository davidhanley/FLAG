package compiler

import (
	"testing"

	flagrt "flag-lang/runtime"
)

func flagIRMap(kind string, extra ...flagrt.Value) flagrt.Value {
	args := append([]flagrt.Value{flagrt.NewKeyword("kind"), flagrt.NewKeyword(kind)}, extra...)
	return flagrt.NewMap(args...)
}

func TestFLAGIRMapRendersLikeGo(t *testing.T) {
	cases := []struct {
		name string
		node flagrt.Value
		want string
	}{
		{"ident", flagIRMap("ident", flagrt.NewKeyword("name"), flagrt.NewString("x")), "x"},
		{"string", flagIRMap("string", flagrt.NewKeyword("value"), flagrt.NewString("hi")), `"hi"`},
		{"int", flagIRMap("int", flagrt.NewKeyword("value"), flagrt.NewLong(42)), "42"},
		{"selector", flagIRMap("selector",
			flagrt.NewKeyword("pkg"), flagrt.NewString("flagrt"),
			flagrt.NewKeyword("name"), flagrt.NewString("NewLong"),
		), "flagrt.NewLong"},
		{
			"new-long",
			flagIRMap("call",
				flagrt.NewKeyword("fun"), flagIRMap("selector",
					flagrt.NewKeyword("pkg"), flagrt.NewString("flagrt"),
					flagrt.NewKeyword("name"), flagrt.NewString("NewLong"),
				),
				flagrt.NewKeyword("args"), flagrt.NewArray(flagIRMap("int", flagrt.NewKeyword("value"), flagrt.NewLong(42))),
			),
			"flagrt.NewLong(42)",
		},
		{
			"index",
			flagIRMap("index",
				flagrt.NewKeyword("x"), flagIRMap("ident", flagrt.NewKeyword("name"), flagrt.NewString("args")),
				flagrt.NewKeyword("index"), flagIRMap("int", flagrt.NewKeyword("value"), flagrt.NewLong(0)),
			),
			"args[0]",
		},
		{
			"binary",
			flagIRMap("binary",
				flagrt.NewKeyword("op"), flagrt.NewString("!="),
				flagrt.NewKeyword("left"), flagIRMap("call",
					flagrt.NewKeyword("fun"), flagIRMap("ident", flagrt.NewKeyword("name"), flagrt.NewString("len")),
					flagrt.NewKeyword("args"), flagrt.NewArray(flagIRMap("ident", flagrt.NewKeyword("name"), flagrt.NewString("args"))),
				),
				flagrt.NewKeyword("right"), flagIRMap("int", flagrt.NewKeyword("value"), flagrt.NewLong(1)),
			),
			"len(args) != 1",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ir, err := flagValueToIRExpr(tc.node)
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

func TestIRExprToFlagValueRoundtrip(t *testing.T) {
	ir := rtCall("Call", IRIdent{Name: "f"}, rtCall("NewLong", IRInt{Value: 1}))
	got, err := flagValueToIRExpr(irExprToFlagValue(ir))
	if err != nil {
		t.Fatal(err)
	}
	if renderIRExpr(got) != renderIRExpr(ir) {
		t.Fatalf("got %q want %q", renderIRExpr(got), renderIRExpr(ir))
	}
}

func TestFLAGIRStmtMapRendersLikeGo(t *testing.T) {
	node := flagIRMap("var",
		flagrt.NewKeyword("name"), flagrt.NewString("x"),
		flagrt.NewKeyword("type"), flagrt.NewString("flagrt.Value"),
	)
	stmt, err := flagValueToIRStmt(node)
	if err != nil {
		t.Fatal(err)
	}
	got := renderIRStmt(stmt, "\t")
	want := "\tvar x flagrt.Value\n"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
