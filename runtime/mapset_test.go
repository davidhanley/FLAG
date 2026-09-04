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

func TestKeysAndVals(t *testing.T) {
	m := NewMap(
		NewKeyword("b"), NewLong(2),
		NewKeyword("a"), NewLong(1),
	)
	if got := ValueToString(Keys(m)); got != "[:a :b]" {
		t.Fatalf("unexpected keys: %q", got)
	}
	if got := ValueToString(Vals(m)); got != "[1 2]" {
		t.Fatalf("unexpected vals: %q", got)
	}
	if got := Vals(NilValue()); got.tag != TagNil {
		t.Fatalf("expected nil vals for nil map, got %v", ValueToAny(got))
	}
}

func TestFind(t *testing.T) {
	m := NewMap(NewKeyword("a"), NewLong(1), NewKeyword("b"), NewLong(2))
	if got := ValueToString(Find(m, NewKeyword("a"))); got != "[:a 1]" {
		t.Fatalf("unexpected find entry: %q", got)
	}
	if got := Find(m, NewKeyword("z")); got.tag != TagNil {
		t.Fatalf("expected nil for missing find key, got %v", ValueToAny(got))
	}
	if got := Find(NilValue(), NewKeyword("a")); got.tag != TagNil {
		t.Fatalf("expected nil find on nil map, got %v", ValueToAny(got))
	}
}

func TestSetAndSetMapHelpers(t *testing.T) {
	if got := ValueToString(Union(NewSet(NewLong(1), NewLong(2)), NewSet(NewLong(2), NewLong(3)))); got != "#{1 2 3}" {
		t.Fatalf("unexpected union: %q", got)
	}
	if got := ValueToString(Intersection(NewSet(NewLong(1), NewLong(2), NewLong(3)), NewSet(NewLong(2), NewLong(3), NewLong(4)))); got != "#{2 3}" {
		t.Fatalf("unexpected intersection: %q", got)
	}
	if got := ValueToString(Difference(NewSet(NewLong(1), NewLong(2), NewLong(3)), NewSet(NewLong(2)))); got != "#{1 3}" {
		t.Fatalf("unexpected difference: %q", got)
	}
	if !Subset(NewSet(NewLong(1), NewLong(2)), NewSet(NewLong(1), NewLong(2), NewLong(3))) {
		t.Fatal("expected subset? true")
	}
	if !Superset(NewSet(NewLong(1), NewLong(2), NewLong(3)), NewSet(NewLong(1), NewLong(2))) {
		t.Fatal("expected superset? true")
	}
	if !Disjoint(NewSet(NewLong(1), NewLong(2)), NewSet(NewLong(3), NewLong(4))) {
		t.Fatal("expected disjoint? true")
	}
	if Disjoint(NewSet(NewLong(1), NewLong(2)), NewSet(NewLong(2), NewLong(3))) {
		t.Fatal("expected disjoint? false")
	}

	m := NewMap(NewKeyword("a"), NewLong(1), NewKeyword("b"), NewLong(2))
	if got := ValueToString(RenameKeys(m, NewMap(NewKeyword("a"), NewKeyword("x")))); got != "{:b 2 :x 1}" {
		t.Fatalf("unexpected rename-keys: %q", got)
	}
	if got := ValueToString(MapInvert(NewMap(NewKeyword("a"), NewLong(1), NewKeyword("b"), NewLong(2)))); got != "{1 :a 2 :b}" {
		t.Fatalf("unexpected map-invert: %q", got)
	}
	evenFn := NewFunction(func(args ...Value) Value {
		if len(args) != 1 || args[0].tag != TagLong {
			panic("evenFn expects one long")
		}
		return NewBool(args[0].Long()%2 == 0)
	})
	if got := ValueToString(SetSelect(evenFn, NewSet(NewLong(1), NewLong(2), NewLong(3), NewLong(4)))); got != "#{2 4}" {
		t.Fatalf("unexpected select: %q", got)
	}

	rel := NewSet(
		NewMap(NewKeyword("a"), NewLong(1), NewKeyword("b"), NewLong(9)),
		NewMap(NewKeyword("a"), NewLong(2), NewKeyword("b"), NewLong(8)),
	)
	projected := SetProject(rel, NewArray(NewKeyword("a")))
	if projected.SetLen() != 2 {
		t.Fatalf("unexpected project set size: %d", projected.SetLen())
	}
	if !Contains(projected, NewMap(NewKeyword("a"), NewLong(1))) || !Contains(projected, NewMap(NewKeyword("a"), NewLong(2))) {
		t.Fatalf("project missing expected rows: %s", ValueToString(projected))
	}
	renamed := SetRename(rel, NewMap(NewKeyword("a"), NewKeyword("x")))
	if renamed.SetLen() != 2 {
		t.Fatalf("unexpected rename set size: %d", renamed.SetLen())
	}
	if !Contains(renamed, NewMap(NewKeyword("x"), NewLong(1), NewKeyword("b"), NewLong(9))) ||
		!Contains(renamed, NewMap(NewKeyword("x"), NewLong(2), NewKeyword("b"), NewLong(8))) {
		t.Fatalf("rename missing expected rows: %s", ValueToString(renamed))
	}
}
