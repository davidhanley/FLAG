package runtime

import "testing"

func TestMapAndSetStringRendering(t *testing.T) {
	m := NewMap(
		NewKeyword("b"), NewLong(2),
		NewKeyword("a"), NewLong(1),
	)
	if got := m.String(); got != "{:a 1 :b 2}" {
		t.Fatalf("unexpected map string: %q", got)
	}

	s := NewSet(NewLong(2), NewLong(1), NewLong(1))
	if got := s.String(); got != "#{1 2}" {
		t.Fatalf("unexpected set string: %q", got)
	}
}

func TestNewMapSupportsHashMapSemantics(t *testing.T) {
	if got := NewMap(); !isMap(got) || got.MapLen() != 0 {
		t.Fatalf("expected empty map from zero-arg NewMap, got %#v", got)
	}

	m := NewMap(NewKeyword("a"), NewLong(1), NewKeyword("b"), NewLong(2))
	if got := m.String(); got != "{:a 1 :b 2}" {
		t.Fatalf("unexpected hash-map result: %q", got)
	}

	assertPanics(t, func() {
		NewMap(NewKeyword("a"))
	})
}

func nKeywordMap(n int) Value {
	entries := make([]Value, 0, n*2)
	for i := 0; i < n; i++ {
		entries = append(entries, NewKeyword(string(rune('a'+i))), NewLong(int64(i)))
	}
	return NewMap(entries...)
}

func TestArrayMapPromotesAtNine(t *testing.T) {
	empty := NewMap()
	if !isType[*arrayMapImpl](empty) {
		t.Fatalf("expected empty array-map, got %T", empty.impl)
	}

	eight := nKeywordMap(8)
	if !isType[*arrayMapImpl](eight) || eight.MapLen() != 8 {
		t.Fatalf("expected 8-key array-map, got %T len=%d", eight.impl, eight.MapLen())
	}

	nine := nKeywordMap(9)
	if !isType[*mapImpl](nine) || nine.MapLen() != 9 {
		t.Fatalf("expected 9-key HAMT, got %T len=%d", nine.impl, nine.MapLen())
	}

	promoted := Assoc(eight, NewKeyword("i"), NewLong(8))
	if !isType[*mapImpl](promoted) {
		t.Fatalf("expected assoc of 9th key to promote, got %T", promoted.impl)
	}
	if !Eq(promoted, nine) {
		t.Fatalf("promoted map should equal 9-key NewMap: %s vs %s", promoted, nine)
	}
	if Hash(promoted) != Hash(nine) {
		t.Fatalf("promoted hash %d != nine hash %d", Hash(promoted), Hash(nine))
	}

	dup := NewMap(NewKeyword("a"), NewLong(1), NewKeyword("a"), NewLong(2))
	if !isType[*arrayMapImpl](dup) || dup.MapLen() != 1 {
		t.Fatalf("expected last-wins compact array-map, got %T len=%d", dup.impl, dup.MapLen())
	}
	if got := Get(dup, NewKeyword("a")); got.Long() != 2 {
		t.Fatalf("expected last-wins value 2, got %s", got)
	}

	hamt := nKeywordMap(9)
	shrank := Dissoc(hamt, NewKeyword("i"))
	if !isType[*mapImpl](shrank) {
		t.Fatalf("dissoc must not demote HAMT, got %T", shrank.impl)
	}
	if !Eq(shrank, nKeywordMap(8)) {
		t.Fatalf("8-key HAMT should equal 8-key array-map: %s vs %s", shrank, nKeywordMap(8))
	}
}

func TestAssocAndDissocMap(t *testing.T) {
	m := NewMap(NewKeyword("a"), NewLong(1))
	m = Assoc(m, NewKeyword("b"), NewLong(2))
	m = Assoc(m, NewKeyword("a"), NewLong(5))
	if got := m.String(); got != "{:a 5 :b 2}" {
		t.Fatalf("unexpected map after assoc: %q", got)
	}

	m = Dissoc(m, NewKeyword("b"))
	if got := m.String(); got != "{:a 5}" {
		t.Fatalf("unexpected map after dissoc: %q", got)
	}

	if got := (Dissoc(NewMap(NewKeyword("a"), NewLong(1)))).String(); got != "{:a 1}" {
		t.Fatalf("unexpected no-op dissoc result: %q", got)
	}
	if got := Dissoc(NilValue); !IsNil(got) {
		t.Fatalf("expected nil dissoc on nil, got %v", ValueToAny(got))
	}
	if got := Dissoc(NilValue, NewKeyword("a"), NewKeyword("b")); !IsNil(got) {
		t.Fatalf("expected nil dissoc with keys on nil, got %v", ValueToAny(got))
	}
}

func TestAssocArrayAndVector(t *testing.T) {
	arr := NewArray(NewLong(1), NewLong(2), NewLong(3))
	updated := Assoc(arr, NewLong(1), NewLong(9))
	if got := updated.String(); got != "[1 9 3]" {
		t.Fatalf("unexpected array assoc: %q", got)
	}
	if got := arr.String(); got != "[1 2 3]" {
		t.Fatalf("assoc should not mutate original array: %q", got)
	}
	if got := (Assoc(NewArray(NewLong(1), NewLong(2)), NewLong(2), NewLong(3))).String(); got != "[1 2 3]" {
		t.Fatalf("unexpected array assoc append: %q", got)
	}
	if got := (Assoc(NewArray(NewLong(1), NewLong(2)), NewLong(0), NewKeyword("a"), NewLong(1), NewKeyword("b"))).String(); got != "[:a :b]" {
		t.Fatalf("unexpected multi-pair array assoc: %q", got)
	}

	vec := NewVector(NewLong(1), NewLong(2), NewLong(3))
	updatedVec := Assoc(vec, NewLong(1), NewLong(9))
	if got := updatedVec.String(); got != "| 1 9 3 |" {
		t.Fatalf("unexpected vector assoc: %q", got)
	}
	if got := vec.String(); got != "| 1 2 3 |" {
		t.Fatalf("assoc should not mutate original vector: %q", got)
	}
	if got := (Assoc(NewVector(NewLong(1), NewLong(2)), NewLong(2), NewLong(3))).String(); got != "| 1 2 3 |" {
		t.Fatalf("unexpected vector assoc append: %q", got)
	}

	assertPanics(t, func() { Assoc(NewArray(NewLong(1)), NewLong(3), NewLong(0)) })
	assertPanics(t, func() { Assoc(NewVector(NewLong(1)), NewLong(3), NewLong(0)) })
	assertPanics(t, func() { Assoc(NewList(NewLong(1)), NewLong(0), NewLong(2)) })
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
	if got := Get(m, NewKeyword("missing")); !IsNil(got) {
		t.Fatalf("expected nil for missing key, got %v", ValueToAny(got))
	}
	if got := Get(m, NewKeyword("missing"), NewLong(42)); got.Long() != 42 {
		t.Fatalf("expected default 42 for missing key, got %v", ValueToAny(got))
	}
}

func TestGetFromArrayVectorSetAndString(t *testing.T) {
	arr := NewArray(NewLong(1), NewLong(2), NewLong(3))
	if got := Get(arr, NewLong(1)); got.Long() != 2 {
		t.Fatalf("expected array get 2, got %v", ValueToAny(got))
	}
	if got := Get(arr, NewLong(9)); !IsNil(got) {
		t.Fatalf("expected nil for out-of-range array index, got %v", ValueToAny(got))
	}
	if got := Get(arr, NewLong(9), NewKeyword("missing")); !Eq(got, NewKeyword("missing")) {
		t.Fatalf("expected default for out-of-range array index, got %v", ValueToAny(got))
	}
	if got := Get(arr, NewKeyword("a")); !IsNil(got) {
		t.Fatalf("expected nil for non-index array key, got %v", ValueToAny(got))
	}

	vec := NewVector(NewLong(1), NewLong(9), NewLong(3))
	if got := Get(vec, NewLong(1)); got.Long() != 9 {
		t.Fatalf("expected vector get 9, got %v", ValueToAny(got))
	}

	s := NewSet(NewLong(1), NewLong(2))
	if got := Get(s, NewLong(2)); got.Long() != 2 {
		t.Fatalf("expected set get member, got %v", ValueToAny(got))
	}
	if got := Get(s, NewLong(3)); !IsNil(got) {
		t.Fatalf("expected nil for missing set member, got %v", ValueToAny(got))
	}

	if got := Get(NewString("hi"), NewLong(1)); got.StringValue() != "i" {
		t.Fatalf("expected string get i, got %v", ValueToAny(got))
	}
	assertPanics(t, func() { Get(NewLong(1), NewLong(0)) })
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

	if Contains(NilValue, NewKeyword("a")) {
		t.Fatal("expected contains? nil to be false")
	}
	assertPanics(t, func() {
		Contains(NewArray(NewLong(1)), NewLong(0))
	})
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
	if got := Keys(m).String(); got != "[:a :b]" {
		t.Fatalf("unexpected keys: %q", got)
	}
	if got := Vals(m).String(); got != "[1 2]" {
		t.Fatalf("unexpected vals: %q", got)
	}
	if got := Vals(NilValue); !IsNil(got) {
		t.Fatalf("expected nil vals for nil map, got %v", ValueToAny(got))
	}
}

func TestFind(t *testing.T) {
	m := NewMap(NewKeyword("a"), NewLong(1), NewKeyword("b"), NewLong(2))
	if got := (Find(m, NewKeyword("a"))).String(); got != "[:a 1]" {
		t.Fatalf("unexpected find entry: %q", got)
	}
	if got := Find(m, NewKeyword("z")); !IsNil(got) {
		t.Fatalf("expected nil for missing find key, got %v", ValueToAny(got))
	}
	if got := Find(NilValue, NewKeyword("a")); !IsNil(got) {
		t.Fatalf("expected nil find on nil map, got %v", ValueToAny(got))
	}
}

func TestSetAndSetMapHelpers(t *testing.T) {
	if got := (Union(NewSet(NewLong(1), NewLong(2)), NewSet(NewLong(2), NewLong(3)))).String(); got != "#{1 2 3}" {
		t.Fatalf("unexpected union: %q", got)
	}
	if got := (Intersection(NewSet(NewLong(1), NewLong(2), NewLong(3)), NewSet(NewLong(2), NewLong(3), NewLong(4)))).String(); got != "#{2 3}" {
		t.Fatalf("unexpected intersection: %q", got)
	}
	if got := (Difference(NewSet(NewLong(1), NewLong(2), NewLong(3)), NewSet(NewLong(2)))).String(); got != "#{1 3}" {
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
	if got := (RenameKeys(m, NewMap(NewKeyword("a"), NewKeyword("x")))).String(); got != "{:b 2 :x 1}" {
		t.Fatalf("unexpected rename-keys: %q", got)
	}
	if got := (MapInvert(NewMap(NewKeyword("a"), NewLong(1), NewKeyword("b"), NewLong(2)))).String(); got != "{1 :a 2 :b}" {
		t.Fatalf("unexpected map-invert: %q", got)
	}
	evenFn := NewFunction(func(args ...Value) Value {
		if len(args) != 1 || !isType[longImpl](args[0]) {
			panic("evenFn expects one long")
		}
		return NewBool(args[0].Long()%2 == 0)
	})
	if got := (SetSelect(evenFn, NewSet(NewLong(1), NewLong(2), NewLong(3), NewLong(4)))).String(); got != "#{2 4}" {
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
		t.Fatalf("project missing expected rows: %s", projected.String())
	}
	renamed := SetRename(rel, NewMap(NewKeyword("a"), NewKeyword("x")))
	if renamed.SetLen() != 2 {
		t.Fatalf("unexpected rename set size: %d", renamed.SetLen())
	}
	if !Contains(renamed, NewMap(NewKeyword("x"), NewLong(1), NewKeyword("b"), NewLong(9))) ||
		!Contains(renamed, NewMap(NewKeyword("x"), NewLong(2), NewKeyword("b"), NewLong(8))) {
		t.Fatalf("rename missing expected rows: %s", renamed.String())
	}
}
