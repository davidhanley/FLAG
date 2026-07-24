package runtime

import (
	"math/big"
	"testing"
)

func TestSubAndDiv(t *testing.T) {
	sub := Sub(NewLong(10), NewDouble(2.5))
	if got := sub.Double(); got != 7.5 {
		t.Fatalf("expected subtraction result 7.5, got %v", got)
	}

	div := Div(NewLong(3), NewLong(2))
	if div.tag != TagRatio {
		t.Fatalf("expected ratio result, got tag %v", div.tag)
	}

	rat := div.Ratio()
	if rat.Cmp(big.NewRat(3, 2)) != 0 {
		t.Fatalf("expected division result 3/2, got %s", rat.RatString())
	}
}

func TestValueToAnyForRatio(t *testing.T) {
	div := Div(NewLong(3), NewLong(2))
	rat, ok := ValueToAny(div).(*big.Rat)
	if !ok {
		t.Fatalf("expected *big.Rat conversion, got %T", ValueToAny(div))
	}
	if rat.Cmp(big.NewRat(3, 2)) != 0 {
		t.Fatalf("expected converted ratio 3/2, got %s", rat.RatString())
	}
}

func TestEqNumeric(t *testing.T) {
	cases := []struct {
		name string
		lhs  Value
		rhs  Value
		want bool
	}{
		{name: "long eq long", lhs: NewLong(2), rhs: NewLong(2), want: true},
		{name: "long eq double", lhs: NewLong(3), rhs: NewDouble(3.0), want: true},
		{name: "ratio eq long", lhs: NewRatio(4, 2), rhs: NewLong(2), want: true},
		{name: "ratio eq double", lhs: NewRatio(3, 2), rhs: NewDouble(1.5), want: true},
		{name: "not equal", lhs: NewLong(1), rhs: NewLong(2), want: false},
	}

	for _, tc := range cases {
		if got := Eq(tc.lhs, tc.rhs); got != tc.want {
			t.Fatalf("%s: expected %v, got %v", tc.name, tc.want, got)
		}
	}
}

func TestLtGtNumeric(t *testing.T) {
	if !Lt(NewLong(1), NewLong(2)) {
		t.Fatal("expected 1 < 2")
	}
	if Lt(NewLong(2), NewLong(1)) {
		t.Fatal("expected 2 !< 1")
	}
	if !Gt(NewRatio(5, 2), NewDouble(2.0)) {
		t.Fatal("expected 5/2 > 2.0")
	}
	if Gt(NewDouble(1.25), NewRatio(3, 2)) {
		t.Fatal("expected 1.25 !> 1.5")
	}
}

func TestNilValueTruthiness(t *testing.T) {
	if ValueToAny(NilValue()) != nil {
		t.Fatal("expected NilValue to convert to nil")
	}
	if IsTruthy(NilValue()) {
		t.Fatal("expected NilValue to be falsey")
	}
	if !IsTruthy(NewLong(0)) {
		t.Fatal("expected numeric zero to be truthy")
	}
}

func TestStr(t *testing.T) {
	if got := Str(); got != "" {
		t.Fatalf("expected empty str() result, got %q", got)
	}

	value := Str(
		"a",
		NewLong(1),
		NewDouble(2.5),
		NewRatio(3, 2),
		NewSymbol("s"),
		NewKeyword("k"),
		true,
		NewList(NewLong(4), NewLong(5)),
		NewArray(NewLong(6), NewLong(7)),
		NilValue(),
	)
	if value != "a12.53/2s:ktrue(4 5)[6 7]" {
		t.Fatalf("unexpected str result: %q", value)
	}
}

func TestSymbolAndName(t *testing.T) {
	sym := Symbol("abc")
	if sym.tag != TagSymbol {
		t.Fatalf("expected symbol tag, got %v", sym.tag)
	}

	if got := Name(sym); got != "abc" {
		t.Fatalf("expected symbol name abc, got %q", got)
	}
	if got := Name(NewKeyword("kw")); got != "kw" {
		t.Fatalf("expected keyword name kw, got %q", got)
	}

	kwAsSymbol := Symbol(NewKeyword("kw"))
	if got := ValueToString(kwAsSymbol); got != "kw" {
		t.Fatalf("expected symbol conversion to drop keyword marker, got %q", got)
	}

	keywordAny, ok := ValueToAny(NewKeyword("key")).(Value)
	if !ok {
		t.Fatalf("expected Value from ValueToAny, got %T", ValueToAny(NewKeyword("key")))
	}
	if got := ValueToString(keywordAny); got != ":key" {
		t.Fatalf("unexpected keyword value: %q", got)
	}
}
