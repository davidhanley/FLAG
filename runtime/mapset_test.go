package runtime

import "testing"

func TestMapAndSetStringRendering(t *testing.T) {
	m := NewMap(
		NewKeyword("b"), NewLong(2),
		NewKeyword("a"), NewLong(1),
	)
	if got := ValueToString(m); got != "{:a 1 :b 2}" {
		t.Fatalf("unexpected map string: %q", got)
	}

	s := NewSet(NewLong(2), NewLong(1), NewLong(1))
	if got := ValueToString(s); got != "#{1 2}" {
		t.Fatalf("unexpected set string: %q", got)
	}
}

func TestNewMapSupportsHashMapSemantics(t *testing.T) {
	if got := NewMap(); got.tag != TagMap || got.MapLen() != 0 {
		t.Fatalf("expected empty map from zero-arg NewMap, got %#v", got)
	}

	m := NewMap(NewKeyword("a"), NewLong(1), NewKeyword("b"), NewLong(2))
	if got := ValueToString(m); got != "{:a 1 :b 2}" {
		t.Fatalf("unexpected hash-map result: %q", got)
	}

	assertPanics(t, func() {
		NewMap(NewKeyword("a"))
	})
}

func TestAssocAndDissocMap(t *testing.T) {
	m := NewMap(NewKeyword("a"), NewLong(1))
	m = Assoc(m, NewKeyword("b"), NewLong(2))
	m = Assoc(m, NewKeyword("a"), NewLong(5))
	if got := ValueToString(m); got != "{:a 5 :b 2}" {
		t.Fatalf("unexpected map after assoc: %q", got)
	}

	m = Dissoc(m, NewKeyword("b"))
	if got := ValueToString(m); got != "{:a 5}" {
		t.Fatalf("unexpected map after dissoc: %q", got)
	}
}

func TestMergeMaps(t *testing.T) {
	if got := Merge(); got.tag != TagNil {
		t.Fatalf("expected nil from merge with no args, got %#v", got)
	}

	single := NewMap(NewKeyword("a"), NewLong(1))
	if got := Merge(single); got.tag != TagMap || ValueToString(got) != "{:a 1}" {
		t.Fatalf("unexpected single-arg merge: %#v / %s", got, ValueToString(got))
	}

	merged := Merge(
		NewMap(NewKeyword("a"), NewLong(1)),
		NilValue(),
		NewMap(NewKeyword("b"), NewLong(2)),
		NewMap(NewKeyword("a"), NewLong(3), NewKeyword("c"), NewLong(4)),
	)
	if got := ValueToString(merged); got != "{:a 3 :b 2 :c 4}" {
		t.Fatalf("unexpected merged map: %q", got)
	}
}

func TestSelectKeys(t *testing.T) {
	selected := SelectKeys(
		NewMap(NewKeyword("a"), NewLong(1), NewKeyword("b"), NilValue(), NewKeyword("c"), NewLong(3)),
		NewArray(NewKeyword("b"), NewKeyword("missing"), NewKeyword("a")),
	)
	if got := ValueToString(selected); got != "{:a 1 :b }" {
		t.Fatalf("unexpected select-keys result: %q", got)
	}

	if got := SelectKeys(NilValue(), NewArray(NewKeyword("a"))); got.tag != TagMap || got.MapLen() != 0 {
		t.Fatalf("expected empty map from select-keys nil input, got %#v", got)
	}
}

func TestUpdateMap(t *testing.T) {
	m := NewMap(NewKeyword("a"), NewLong(1))
	add := NewFunction(func(args ...Value) Value { return Add(args[0], args[1]) })
	updated := Update(m, NewKeyword("a"), add, NewLong(2))
	if got := ValueToString(updated); got != "{:a 3}" {
		t.Fatalf("unexpected map after update: %q", got)
	}

	incFromNil := NewFunction(func(args ...Value) Value {
		if args[0].tag == TagNil {
			return NewLong(1)
		}
		return Add(args[0], NewLong(1))
	})
	updatedNil := Update(NilValue(), NewKeyword("b"), incFromNil)
	if got := ValueToString(updatedNil); got != "{:b 1}" {
		t.Fatalf("unexpected map after update on nil: %q", got)
	}
}

func TestUpdateArray(t *testing.T) {
	double := NewFunction(func(args ...Value) Value { return Mul(args[0], NewLong(2)) })
	updated := Update(NewArray(NewLong(1), NewLong(2), NewLong(3)), NewLong(1), double)
	if got := ValueToString(updated); got != "[1 4 3]" {
		t.Fatalf("unexpected array after update: %q", got)
	}
}

func TestEqMapAndSet(t *testing.T) {
	mapA := NewMap(NewKeyword("a"), NewLong(1), NewKeyword("b"), NewLong(2))
	mapB := NewMap(NewKeyword("b"), NewLong(2), NewKeyword("a"), NewLong(1))
	if !Eq(mapA, mapB) {
		t.Fatal("expected maps with same entries to be equal")
	}

	setA := NewSet(NewLong(2), NewLong(1))
	setB := NewSet(NewLong(1), NewLong(2))
	if !Eq(setA, setB) {
		t.Fatal("expected sets with same entries to be equal")
	}
}

func TestGetFromMap(t *testing.T) {
	m := NewMap(NewKeyword("a"), NewLong(1))
	if got := Get(m, NewKeyword("a")); got.Long() != 1 {
		t.Fatalf("expected map get 1, got %v", ValueToAny(got))
	}
	if got := Get(m, NewKeyword("missing")); got.tag != TagNil {
		t.Fatalf("expected nil for missing key, got %v", ValueToAny(got))
	}
	if got := Get(m, NewKeyword("missing"), NewLong(42)); got.Long() != 42 {
		t.Fatalf("expected default 42 for missing key, got %v", ValueToAny(got))
	}
}

func TestContainsOnMapAndSet(t *testing.T) {
	m := NewMap(NewKeyword("a"), NewLong(1))
	if !Contains(m, NewKeyword("a")) {
		t.Fatal("expected contains? to find map key")
	}
	if Contains(m, NewKeyword("missing")) {
		t.Fatal("expected contains? to miss map key")
	}

	s := NewSet(NewLong(1), NewLong(2))
	if !Contains(s, NewLong(2)) {
		t.Fatal("expected contains? to find set member")
	}
	if Contains(s, NewLong(3)) {
		t.Fatal("expected contains? to miss set member")
	}
}

func TestMapAndKeywordInvocation(t *testing.T) {
	m := NewMap(NewKeyword("a"), NewLong(7))
	if got := Call(m, NewKeyword("a")); got.Long() != 7 {
		t.Fatalf("expected map invocation to return 7, got %v", ValueToAny(got))
	}
	if got := Call(NewKeyword("a"), m); got.Long() != 7 {
		t.Fatalf("expected keyword invocation to return 7, got %v", ValueToAny(got))
	}
	if got := Call(m, NewKeyword("missing"), NewLong(9)); got.Long() != 9 {
		t.Fatalf("expected map invocation default 9, got %v", ValueToAny(got))
	}
	if got := Call(NewKeyword("missing"), m, NewLong(11)); got.Long() != 11 {
		t.Fatalf("expected keyword invocation default 11, got %v", ValueToAny(got))
	}
}
