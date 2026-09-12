package runtime

import (
	"fmt"
	"math"
	"sync"
	"unsafe"
)

// AtomObject is an uncoordinated, synchronous FLAG atom. No watches.
type AtomObject struct {
	mu sync.Mutex
	v  Value
}

// Atom creates an atom. Zero args → nil; one arg is the initial value.
func Atom(args ...Value) Value {
	var init Value
	switch len(args) {
	case 0:
		init = NilValue()
	case 1:
		init = args[0]
	default:
		panic("atom expects 0 or 1 arguments")
	}
	a := &AtomObject{v: init}
	return Value{p: unsafe.Pointer(a), tag: TagAtom}
}

func (v Value) AtomObject() *AtomObject {
	if v.tag != TagAtom {
		panic(fmt.Sprintf("expected atom, got %s", ValueToString(v)))
	}
	if v.p == nil {
		panic("atom Value does not contain atom pointer")
	}
	return (*AtomObject)(v.p)
}

func valueSlotEq(a, b Value) bool {
	return a.tag == b.tag && a.p == b.p && math.Float64bits(a.d) == math.Float64bits(b.d)
}

// Deref returns the current value of atom.
func Deref(atom Value) Value {
	if atom.tag != TagAtom {
		panic(fmt.Sprintf("deref expects an atom, got %s", ValueToString(atom)))
	}
	a := atom.AtomObject()
	a.mu.Lock()
	v := a.v
	a.mu.Unlock()
	return v
}

// Reset sets atom to v and returns v.
func Reset(atom, v Value) Value {
	if atom.tag != TagAtom {
		panic(fmt.Sprintf("reset! expects an atom, got %s", ValueToString(atom)))
	}
	a := atom.AtomObject()
	a.mu.Lock()
	a.v = v
	a.mu.Unlock()
	return v
}

// Swap applies f to the current value plus extra args, CAS-retrying until it
// stores the result. Returns the new value. f may run more than once.
func Swap(args ...Value) Value {
	if len(args) < 2 {
		panic("swap! expects atom, function, and optional extra arguments")
	}
	if args[0].tag != TagAtom {
		panic(fmt.Sprintf("swap! expects an atom, got %s", ValueToString(args[0])))
	}
	a := args[0].AtomObject()
	f := args[1]
	extra := args[2:]
	for {
		a.mu.Lock()
		old := a.v
		a.mu.Unlock()
		callArgs := make([]Value, 1+len(extra))
		callArgs[0] = old
		copy(callArgs[1:], extra)
		neu := Call(f, callArgs...)
		a.mu.Lock()
		if valueSlotEq(a.v, old) {
			a.v = neu
			a.mu.Unlock()
			return neu
		}
		a.mu.Unlock()
	}
}
