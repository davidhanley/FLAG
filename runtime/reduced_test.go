package runtime

import "testing"

func TestReducedBoxAndPredicates(t *testing.T) {
	box := Reduced(NewLong(1))
	if !IsReduced(box) {
		t.Fatal("expected reduced box")
	}
	if IsReduced(NewLong(1)) {
		t.Fatal("plain value should not be reduced")
	}
	if TypeOf(box).String() != ":reduced" {
		t.Fatalf("type-of reduced: %s", TypeOf(box))
	}
	if box.String() != "#<reduced>" {
		t.Fatalf("string: %s", box.String())
	}
	if Unreduced(box).Long() != 1 {
		t.Fatalf("unreduced: %s", Unreduced(box))
	}
	if Unreduced(NewLong(7)).Long() != 7 {
		t.Fatal("unreduced identity")
	}
	ensured := EnsureReduced(NewLong(2))
	if !IsReduced(ensured) || Unreduced(ensured).Long() != 2 {
		t.Fatalf("ensure-reduced wrap: %s", ensured)
	}
	if EnsureReduced(box) != box {
		t.Fatal("ensure-reduced should not re-box")
	}
	if Eq(Reduced(NewLong(1)), Reduced(NewLong(1))) {
		t.Fatal("distinct reduced boxes should not be equal")
	}
	if !Eq(box, box) {
		t.Fatal("reduced box should equal itself")
	}
}

func TestReduceStopsOnReduced(t *testing.T) {
	calls := 0
	fn := NewFunction(func(args ...Value) Value {
		calls++
		sum := Add(args[0], args[1])
		if sum.Long() >= 6 {
			return Reduced(sum)
		}
		return sum
	})
	got := Reduce(fn, NewLong(0), NewArray(NewLong(1), NewLong(2), NewLong(3), NewLong(4), NewLong(5)))
	if !isType[longImpl](got) || got.Long() != 6 {
		t.Fatalf("expected 6, got %#v", got)
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}

func TestReduceEmptyInitStaysReduced(t *testing.T) {
	init := Reduced(NewLong(1))
	got := Reduce(BuiltinFunction("+"), init, NewArray())
	if !IsReduced(got) || Unreduced(got).Long() != 1 {
		t.Fatalf("empty coll should keep reduced init, got %#v", got)
	}
}

func TestReduceZeroArityUnwrapsReduced(t *testing.T) {
	fn := NewFunction(func(args ...Value) Value {
		return Reduced(NewLong(42))
	})
	got := Reduce(fn, NewArray())
	if IsReduced(got) || !isType[longImpl](got) || got.Long() != 42 {
		t.Fatalf("expected unwrapped 42, got %#v", got)
	}
}

func TestFnReducedArityPanics(t *testing.T) {
	assertPanics(t, func() { FnReduced() })
	assertPanics(t, func() { FnReducedQ() })
	assertPanics(t, func() { FnEnsureReduced(NewLong(1), NewLong(2)) })
	assertPanics(t, func() { FnUnreduced() })
}
