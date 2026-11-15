package runtime

import (
	"sync"
	"testing"
)

func TestAtomDerefReset(t *testing.T) {
	a := Atom()
	if !IsNil(Deref(a)) {
		t.Fatalf("empty atom: got %v", Deref(a).String())
	}
	a = Atom(NewLong(7))
	if got := Deref(a); !isType[longImpl](got) || got.Long() != 7 {
		t.Fatalf("deref init: got %v", got.String())
	}
	if got := Reset(a, NewString("ok")); !isType[*stringImpl](got) || got.StringValue() != "ok" {
		t.Fatalf("reset! return: got %v", got.String())
	}
	if got := Deref(a); !isType[*stringImpl](got) || got.StringValue() != "ok" {
		t.Fatalf("deref after reset: got %v", got.String())
	}
}

func TestAtomSwap(t *testing.T) {
	a := Atom(NewLong(1))
	inc := NewFunction(func(args ...Value) Value {
		return Add(args[0], NewLong(1))
	})
	if got := Swap(a, inc); !isType[longImpl](got) || got.Long() != 2 {
		t.Fatalf("swap! inc: got %v", got.String())
	}
	if got := Swap(a, BuiltinFunction("+"), NewLong(3)); got.Long() != 5 {
		t.Fatalf("swap! + 3: got %v", got.String())
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
	if got := Deref(a); !isType[longImpl](got) || got.Long() != n {
		t.Fatalf("expected %d, got %v", n, got.String())
	}
}

func TestAtomTypeOf(t *testing.T) {
	if got := TypeOf(Atom()).String(); got != ":atom" {
		t.Fatalf("type-of atom: %s", got)
	}
}

func TestAtomRejectsNonAtom(t *testing.T) {
	assertPanics(t, func() { _ = Deref(NewLong(1)) })
	assertPanics(t, func() { _ = Reset(NewLong(1), NewLong(2)) })
	assertPanics(t, func() { _ = Swap(NewLong(1), BuiltinFunction("+")) })
	assertPanics(t, func() { _ = Atom(NewLong(1), NewLong(2)) })
}
