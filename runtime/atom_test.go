package runtime

import (
	"sync"
	"testing"
)

func TestAtomDerefReset(t *testing.T) {
	a := Atom()
	if !IsNil(Deref(a)) {
		t.Fatalf("empty atom: got %v", ValueToString(Deref(a)))
	}
	a = Atom(NewLong(7))
	if got := Deref(a); got.tag != TagLong || got.Long() != 7 {
		t.Fatalf("deref init: got %v", ValueToString(got))
	}
	if got := Reset(a, NewString("ok")); got.tag != TagString || got.StringValue() != "ok" {
		t.Fatalf("reset! return: got %v", ValueToString(got))
	}
	if got := Deref(a); got.tag != TagString || got.StringValue() != "ok" {
		t.Fatalf("deref after reset: got %v", ValueToString(got))
	}
}

func TestAtomSwap(t *testing.T) {
	a := Atom(NewLong(1))
	inc := NewFunction(func(args ...Value) Value {
		return Add(args[0], NewLong(1))
	})
	if got := Swap(a, inc); got.tag != TagLong || got.Long() != 2 {
		t.Fatalf("swap! inc: got %v", ValueToString(got))
	}
	if got := Swap(a, BuiltinFunction("+"), NewLong(3)); got.Long() != 5 {
		t.Fatalf("swap! + 3: got %v", ValueToString(got))
	}
}

func TestAtomSwapConcurrent(t *testing.T) {
	a := Atom(NewLong(0))
	inc := NewFunction(func(args ...Value) Value {
		return Add(args[0], NewLong(1))
	})
	const n = 100
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			_ = Swap(a, inc)
		}()
	}
	wg.Wait()
	if got := Deref(a); got.tag != TagLong || got.Long() != n {
		t.Fatalf("expected %d, got %v", n, ValueToString(got))
	}
}

func TestAtomTypeOf(t *testing.T) {
	if got := ValueToString(TypeOf(Atom())); got != ":atom" {
		t.Fatalf("type-of atom: %s", got)
	}
}

func TestAtomRejectsNonAtom(t *testing.T) {
	assertPanics(t, func() { _ = Deref(NewLong(1)) })
	assertPanics(t, func() { _ = Reset(NewLong(1), NewLong(2)) })
	assertPanics(t, func() { _ = Swap(NewLong(1), BuiltinFunction("+")) })
	assertPanics(t, func() { _ = Atom(NewLong(1), NewLong(2)) })
}
