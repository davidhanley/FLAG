package runtime

import (
	"fmt"
	"strings"
	"testing"
)

func TestValueToAnyKeepsVector(t *testing.T) {
	v := NewVector(NewLong(1), NewLong(2), NewLong(3))
	converted, ok := ValueToAny(v).(Value)
	if !ok {
		t.Fatalf("expected vector Value from ValueToAny, got %T", ValueToAny(v))
	}
	if converted.tag != TagVector {
		t.Fatalf("expected TagVector, got %v", converted.tag)
	}
	if converted.String() != "| 1 2 3 |" {
		t.Fatalf("unexpected vector print: %s", converted.String())
	}
}

func TestVectorNSVectorAndGet(t *testing.T) {
	v := VectorNSVector(NewLong(1), NewKeyword("a"))
	if v.tag != TagVector || v.VectorLen() != 2 {
		t.Fatalf("expected vector of len 2, got %s", v.String())
	}
	if got := VectorNSGet(v, NewLong(1)); !Eq(got, NewKeyword("a")) {
		t.Fatalf("expected keyword at index 1, got %s", got.String())
	}
	if got := VectorNSGet(v, NewLong(9), NewKeyword("missing")); !Eq(got, NewKeyword("missing")) {
		t.Fatalf("expected keyword default, got %s", got.String())
	}
}

func TestVectorNSMutationOps(t *testing.T) {
	v := NewVector(NewLong(1), NewLong(2), NewLong(3))
	if got := VectorNSSet(v, NewLong(1), NewLong(9)); got.String() != "| 1 9 3 |" {
		t.Fatalf("unexpected set result: %s", got.String())
	}
	if got := VectorNSSet(v, NewLong(3), NewLong(4)); got.String() != "| 1 2 3 4 |" {
		t.Fatalf("unexpected set-at-end result: %s", got.String())
	}
	if v.String() != "| 1 2 3 |" {
		t.Fatalf("expected original unchanged, got %s", v.String())
	}
	if got := VectorNSAppend(v, NewLong(4), NewLong(5)); got.String() != "| 1 2 3 4 5 |" {
		t.Fatalf("unexpected append result: %s", got.String())
	}
	if got := VectorNSPrepend(v, NewLong(0)); got.String() != "| 0 1 2 3 |" {
		t.Fatalf("unexpected prepend result: %s", got.String())
	}
	if got := VectorNSPop(v); got.String() != "| 1 2 |" {
		t.Fatalf("unexpected pop result: %s", got.String())
	}
	if got := VectorNSInsert(v, NewLong(1), NewLong(7)); got.String() != "| 1 7 2 3 |" {
		t.Fatalf("unexpected insert result: %s", got.String())
	}
	if got := VectorNSRemove(v, NewLong(1)); got.String() != "| 1 3 |" {
		t.Fatalf("unexpected remove result: %s", got.String())
	}
}

func TestVectorNSEdgeCases(t *testing.T) {
	if got := VectorNSAppend(NilValue, NewLong(1), NewLong(2)); got.String() != "| 1 2 |" {
		t.Fatalf("unexpected append nil result: %s", got.String())
	}
	if got := VectorNSInsert(NilValue, NewLong(0), NewLong(1)); got.String() != "| 1 |" {
		t.Fatalf("unexpected insert nil result: %s", got.String())
	}
	if got := VectorNSSet(NilValue, NewLong(0), NewLong(1)); got.String() != "| 1 |" {
		t.Fatalf("unexpected set nil result: %s", got.String())
	}
	if NewVector().String() != "| |" {
		t.Fatalf("unexpected empty vector print: %s", NewVector().String())
	}
	if got := VectorNSGet(NilValue, NewLong(0)); !Eq(got, NilValue) {
		t.Fatalf("unexpected get nil result: %s", got.String())
	}
}

func TestVectorNSErrors(t *testing.T) {
	expectVectorNSPanic(t, "vector/get expects at most one", func() {
		VectorNSGet(NewVector(NewLong(1)), NewLong(0), NewLong(1), NewLong(2))
	})
	expectVectorNSPanic(t, "vector/set index out of bounds", func() {
		VectorNSSet(NewVector(NewLong(1)), NewLong(3), NewLong(9))
	})
	expectVectorNSPanic(t, "vector/pop expects non-empty vector", func() {
		VectorNSPop(NewVector())
	})
	expectVectorNSPanic(t, "vector/* expects vector Value", func() {
		VectorNSPop(NewArray(NewLong(1)))
	})
	expectVectorNSPanic(t, "vector/remove index out of bounds", func() {
		VectorNSRemove(NilValue, NewLong(0))
	})
	expectVectorNSPanic(t, "vector/pop expects non-empty vector", func() {
		VectorNSPop(NilValue)
	})
}

func expectVectorNSPanic(t *testing.T, contains string, fn func()) {
	t.Helper()
	defer func() {
		recovered := recover()
		if recovered == nil {
			t.Fatalf("expected panic containing %q", contains)
		}
		message := fmt.Sprint(recovered)
		if !strings.Contains(message, contains) {
			t.Fatalf("expected panic containing %q, got %q", contains, message)
		}
	}()
	fn()
}
