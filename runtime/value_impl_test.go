package runtime

import (
	"os"
	"testing"
)

func TestValueImplPresentAndMatchesHelpers(t *testing.T) {
	tmp, err := os.CreateTemp("", "flag-value-impl")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()

	values := []Value{
		NewLong(42),
		NewBigInt(99),
		NewDouble(1.5),
		NewRatio(3, 4),
		NewBool(true),
		NewBool(false),
		NewString("hi"),
		NewDateFromYMD(2020, 1, 2),
		NewFile(tmp),
		NewSymbol("sym"),
		NewKeyword("kw"),
		NewFunction(func(args ...Value) Value { return NilValue }),
		NewMap(NewKeyword("a"), NewLong(1)),
		NewSet(NewLong(1)),
		NilValue,
		NewList(NewLong(1), NewLong(2)),
		NewArray(NewLong(1)),
		NewLazyList(func() (Value, bool) { return Value{}, false }),
		MakeChannel(),
		NewRecur(NewLong(1)),
		NewRecord(struct{ X int }{X: 1}),
		NewVector(NewLong(1)),
		Atom(NewLong(1)),
	}
	seen := map[ValueTag]bool{}
	for _, v := range values {
		if v.impl == nil {
			t.Fatalf("constructed value tag %v has nil impl", v.tag)
		}
		if got, want := v.impl.typ(v), TypeOf(v); !Eq(got, want) {
			t.Fatalf("typ %v: got %s want %s", v.tag, got.String(), want.String())
		}
		if got, want := v.impl.String(v), v.String(); got != want {
			t.Fatalf("String %v: got %q want %q", v.tag, got, want)
		}
		if got, want := v.impl.hash(v), valueIdentity(v); got != want {
			t.Fatalf("hash %v: got %q want %q", v.tag, got, want)
		}
		seen[v.tag] = true
	}
	for tag := TagLong; tag <= TagAtom; tag++ {
		if !seen[tag] {
			t.Fatalf("no constructed value for tag %v", tag)
		}
	}
}

func TestValueImplTypDistinguishesKeyword(t *testing.T) {
	if got := TypeOf(NewKeyword("a")); !Eq(got, NewKeyword("keyword")) {
		t.Fatalf("keyword type-of %s", got.String())
	}
	if got := TypeOf(NewSymbol("a")); !Eq(got, NewKeyword("symbol")) {
		t.Fatalf("symbol type-of %s", got.String())
	}
}

func TestValueImplUninitializedPanics(t *testing.T) {
	assertPanics(t, func() { TypeOf(Value{}) })
	assertPanics(t, func() { implFor(0) })
}

func TestValueImplHashStableForEqualImmediates(t *testing.T) {
	if valueIdentity(NewLong(1)) != valueIdentity(NewLong(1)) {
		t.Fatal("long hash should be stable")
	}
	if valueIdentity(NewDouble(2.0)) != valueIdentity(NewLong(2)) {
		t.Fatal("integral double should hash like long")
	}
}
