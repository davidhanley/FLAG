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
	if ValueToString(converted) != "| 1 2 3 |" {
		t.Fatalf("unexpected vector print: %s", ValueToString(converted))
	}
}

func TestVectorNSVectorAndGet(t *testing.T) {
	v := VectorNSVector(NewLong(1), NewKeyword("a"))
	if v.tag != TagVector || v.VectorLen() != 2 {
		t.Fatalf("expected vector of len 2, got %s", ValueToString(v))
	}
	if got := VectorNSGet(v, NewLong(1)); !Eq(got, NewKeyword("a")) {
		t.Fatalf("expected keyword at index 1, got %s", ValueToString(got))
	}
	if got := VectorNSGet(v, NewLong(9), NewKeyword("missing")); !Eq(got, NewKeyword("missing")) {
		t.Fatalf("expected keyword default, got %s", ValueToString(got))
	}
}

func TestVectorNSMutationOps(t *testing.T) {
	v := NewVector(NewLong(1), NewLong(2), NewLong(3))
	if got := VectorNSSet(v, NewLong(1), NewLong(9)); ValueToString(got) != "| 1 9 3 |" {
		t.Fatalf("unexpected set result: %s", ValueToString(got))
	}
	if got := VectorNSSet(v, NewLong(3), NewLong(4)); ValueToString(got) != "| 1 2 3 4 |" {
		t.Fatalf("unexpected set-at-end result: %s", ValueToString(got))
	}
	if ValueToString(v) != "| 1 2 3 |" {
		t.Fatalf("expected original unchanged, got %s", ValueToString(v))
	}
	if got := VectorNSAppend(v, NewLong(4), NewLong(5)); ValueToString(got) != "| 1 2 3 4 5 |" {
		t.Fatalf("unexpected append result: %s", ValueToString(got))
	}
	if got := VectorNSPrepend(v, NewLong(0)); ValueToString(got) != "| 0 1 2 3 |" {
		t.Fatalf("unexpected prepend result: %s", ValueToString(got))
	}
	if got := VectorNSPop(v); ValueToString(got) != "| 1 2 |" {
		t.Fatalf("unexpected pop result: %s", ValueToString(got))
	}
	if got := VectorNSInsert(v, NewLong(1), NewLong(7)); ValueToString(got) != "| 1 7 2 3 |" {
		t.Fatalf("unexpected insert result: %s", ValueToString(got))
	}
	if got := VectorNSRemove(v, NewLong(1)); ValueToString(got) != "| 1 3 |" {
		t.Fatalf("unexpected remove result: %s", ValueToString(got))
	}
}

func TestVectorNSEdgeCases(t *testing.T) {
	if got := VectorNSAppend(NilValue(), NewLong(1), NewLong(2)); ValueToString(got) != "| 1 2 |" {
		t.Fatalf("unexpected append nil result: %s", ValueToString(got))
	}
	if got := VectorNSInsert(NilValue(), NewLong(0), NewLong(1)); ValueToString(got) != "| 1 |" {
		t.Fatalf("unexpected insert nil result: %s", ValueToString(got))
	}
	if got := VectorNSSet(NilValue(), NewLong(0), NewLong(1)); ValueToString(got) != "| 1 |" {
		t.Fatalf("unexpected set nil result: %s", ValueToString(got))
	}
	if ValueToString(NewVector()) != "| |" {
		t.Fatalf("unexpected empty vector print: %s", ValueToString(NewVector()))
	}
	if got := VectorNSGet(NilValue(), NewLong(0)); !Eq(got, NilValue()) {
		t.Fatalf("unexpected get nil result: %s", ValueToString(got))
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
		VectorNSRemove(NilValue(), NewLong(0))
	})
	expectVectorNSPanic(t, "vector/pop expects non-empty vector", func() {
		VectorNSPop(NilValue())
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
