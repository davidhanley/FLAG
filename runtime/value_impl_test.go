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
	seen := map[string]bool{}
	for _, v := range values {
		if v.impl == nil {
			t.Fatalf("constructed value %v has nil impl", TypeOf(v))
		}
		if got, want := v.impl.typ(v), TypeOf(v); !Eq(got, want) {
			t.Fatalf("typ %v: got %s want %s", TypeOf(v), got.String(), want.String())
		}
		if got, want := v.impl.String(v), v.String(); got != want {
			t.Fatalf("String %v: got %q want %q", TypeOf(v), got, want)
		}
		h := Hash(v)
		if h&1 == 0 {
			t.Fatalf("hash %v is even: %d", TypeOf(v), h)
		}
		if Hash(v) != h {
			t.Fatalf("hash %v was not stable", TypeOf(v))
		}
		seen[TypeOf(v).String()] = true
	}
	if len(seen) < 20 {
		t.Fatalf("expected many distinct types, got %v", seen)
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
}

func TestValueImplHashStableForEqualImmediates(t *testing.T) {
	if Hash(NewLong(1)) != Hash(NewLong(1)) {
		t.Fatal("long hash should be stable")
	}
	if Hash(NewDouble(2.0)) != Hash(NewLong(2)) {
		t.Fatal("integral double should hash like long")
	}
}

func TestValueImplHashCachesOnHeap(t *testing.T) {
	s := NewString("hello")
	h := Hash(s)
	if h&1 == 0 {
		t.Fatalf("expected odd hash, got %d", h)
	}
	cached := asImpl[*stringImpl](s, "string").hashCache
	if cached != h {
		t.Fatalf("cached hash %d want %d", cached, h)
	}
	if Hash(s) != h {
		t.Fatal("second hash should use cache")
	}
}
