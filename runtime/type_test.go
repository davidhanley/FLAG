package runtime

import "testing"

func TestTypeOfKeywords(t *testing.T) {
	cases := []struct {
		name string
		v    Value
		want string
	}{
		{name: "int", v: NewLong(1), want: ":int"},
		{name: "bigint", v: NewBigInt(10), want: ":bigint"},
		{name: "float", v: NewDouble(1.5), want: ":float"},
		{name: "ratio", v: NewRatio(1, 2), want: ":ratio"},
		{name: "bool", v: NewBool(true), want: ":bool"},
		{name: "string", v: NewString("hi"), want: ":string"},
		{name: "symbol", v: NewSymbol("abc"), want: ":symbol"},
		{name: "keyword", v: NewKeyword("kw"), want: ":keyword"},
		{name: "fn", v: NewFunction(func(args ...Value) Value { return NilValue() }), want: ":fn"},
		{name: "map", v: NewMap(), want: ":map"},
		{name: "set", v: NewSet(), want: ":set"},
		{name: "nil", v: NilValue(), want: ":nil"},
		{name: "list", v: NewList(NewLong(1)), want: ":list"},
		{name: "array", v: NewArray(NewLong(1)), want: ":array"},
		{name: "vector", v: NewVector(NewLong(1)), want: ":vector"},
	}
	for _, tc := range cases {
		got := ValueToString(TypeOf(tc.v))
		if got != tc.want {
			t.Fatalf("%s: expected %s, got %s", tc.name, tc.want, got)
		}
	}
}
