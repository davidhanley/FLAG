package runtime

import "testing"

func TestArrayStoresLengthInNumericField(t *testing.T) {
	arrayValue := NewArray(NewLong(1), NewLong(2), NewLong(3))
	if arrayValue.tag != TagArray {
		t.Fatalf("expected TagArray, got %v", arrayValue.tag)
	}
	if got := arrayValue.ArrayLen(); got != 3 {
		t.Fatalf("expected ArrayLen 3, got %d", got)
	}
	if got := arrayValue.Long(); got != 3 {
		t.Fatalf("expected numeric field to hold length 3, got %d", got)
	}
}

func TestArrayAssocPreservesOriginal(t *testing.T) {
	original := NewArray(NewLong(1), NewLong(2), NewLong(3))
	updated := ArrayAssoc(original, 1, NewLong(99))

	if got := ArrayGet(original, 1).Long(); got != 2 {
		t.Fatalf("expected original index 1 to remain 2, got %d", got)
	}
	if got := ArrayGet(updated, 1).Long(); got != 99 {
		t.Fatalf("expected updated index 1 to be 99, got %d", got)
	}
}

func TestArrayAppendReturnsNewArray(t *testing.T) {
	original := NewArray(NewLong(1), NewLong(2))
	appended := ArrayAppend(original, NewLong(3))

	if got := original.ArrayLen(); got != 2 {
		t.Fatalf("expected original length 2, got %d", got)
	}
	if got := appended.ArrayLen(); got != 3 {
		t.Fatalf("expected appended length 3, got %d", got)
	}
	if got := ArrayGet(appended, 2).Long(); got != 3 {
		t.Fatalf("expected appended value 3, got %d", got)
	}
}

func TestArrayAppendReusesBackingWhenCapacityAvailable(t *testing.T) {
	backing := []Value{NewLong(1), NewLong(2), NewLong(0), NewLong(0)}
	original := newArrayValue(backing, 2)
	appended := ArrayAppend(original, NewLong(3))

	if original.p != appended.p {
		t.Fatal("expected append to reuse backing storage")
	}
	if got := original.ArrayLen(); got != 2 {
		t.Fatalf("expected original logical length 2, got %d", got)
	}
	if got := appended.ArrayLen(); got != 3 {
		t.Fatalf("expected appended logical length 3, got %d", got)
	}
	if got := ArrayGet(appended, 2).Long(); got != 3 {
		t.Fatalf("expected appended third element to be 3, got %d", got)
	}
	assertPanics(t, func() { ArrayGet(original, 2) })
}

func TestArrayAppendGrowthPolicy(t *testing.T) {
	empty := newArrayValue([]Value{}, 0)
	one := ArrayAppend(empty, NewLong(1))
	if got := len(one.arrayItems()); got != 2 {
		t.Fatalf("expected capacity 2 after append from 0, got %d", got)
	}

	single := newArrayValue([]Value{NewLong(1)}, 1)
	two := ArrayAppend(single, NewLong(2))
	if got := len(two.arrayItems()); got != 2 {
		t.Fatalf("expected capacity 2 after append from 1, got %d", got)
	}

	fullTwo := newArrayValue([]Value{NewLong(1), NewLong(2)}, 2)
	three := ArrayAppend(fullTwo, NewLong(3))
	if got := len(three.arrayItems()); got != 3 {
		t.Fatalf("expected capacity 3 after append from 2, got %d", got)
	}
}

func TestValueToAnyForArray(t *testing.T) {
	arrayValue := NewArray(NewLong(1), NewDouble(2.5))
	converted, ok := ValueToAny(arrayValue).([]any)
	if !ok {
		t.Fatalf("expected []any conversion, got %T", ValueToAny(arrayValue))
	}
	if len(converted) != 2 {
		t.Fatalf("expected 2 array items, got %d", len(converted))
	}
	if converted[0] != int64(1) {
		t.Fatalf("expected first item int64(1), got %#v", converted[0])
	}
	if converted[1] != 2.5 {
		t.Fatalf("expected second item 2.5, got %#v", converted[1])
	}
}

func TestFirstAndRestForArray(t *testing.T) {
	arrayValue := NewArray(NewLong(1), NewLong(2), NewLong(3))
	if got := First(arrayValue); got.Long() != 1 {
		t.Fatalf("expected first element 1, got %v", got)
	}

	rest := Rest(arrayValue)
	if rest.tag != TagArray {
		t.Fatalf("expected rest to be array, got %v", rest.tag)
	}
	if rest.ArrayLen() != 2 {
		t.Fatalf("expected rest length 2, got %d", rest.ArrayLen())
	}
	if ArrayGet(rest, 0).Long() != 2 || ArrayGet(rest, 1).Long() != 3 {
		t.Fatalf("unexpected rest values: %#v", rest.ArrayValues())
	}

	if got := First(NewArray()); got.tag != TagNil {
		t.Fatalf("expected first of empty array to be nil, got %v", got.tag)
	}
}

func assertPanics(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	fn()
}
