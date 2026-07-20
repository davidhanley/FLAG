package runtime

import "testing"

func TestNewListStoresLengthInNumericField(t *testing.T) {
	listValue := NewList(NewLong(1), NewDouble(2.5), NewLong(3))

	if listValue.tag != TagList {
		t.Fatalf("expected TagList, got %v", listValue.tag)
	}
	if got := listValue.ListLen(); got != 3 {
		t.Fatalf("expected ListLen 3, got %d", got)
	}
	if got := listValue.Long(); got != 3 {
		t.Fatalf("expected numeric field to hold length 3, got %d", got)
	}
}

func TestListAppendUpdatesLength(t *testing.T) {
	listValue := NewList(NewLong(1))
	listValue = ListAppend(listValue, NewLong(2))
	listValue = ListAppend(listValue, NewLong(3))

	if got := listValue.ListLen(); got != 3 {
		t.Fatalf("expected ListLen 3, got %d", got)
	}
}

func TestListRestAndConsPreserveOriginal(t *testing.T) {
	original := NewList(NewLong(1), NewLong(2), NewLong(3))
	rest := ListRest(original)
	extended := ListCons(rest, NewLong(0))

	if got := original.ListLen(); got != 3 {
		t.Fatalf("expected original length 3, got %d", got)
	}
	if got := rest.ListLen(); got != 2 {
		t.Fatalf("expected rest length 2, got %d", got)
	}
	if got := extended.ListLen(); got != 3 {
		t.Fatalf("expected extended length 3, got %d", got)
	}

	wantOriginal := []Value{NewLong(1), NewLong(2), NewLong(3)}
	wantRest := []Value{NewLong(2), NewLong(3)}
	wantExtended := []Value{NewLong(0), NewLong(2), NewLong(3)}

	checkListValues(t, original, wantOriginal)
	checkListValues(t, rest, wantRest)
	checkListValues(t, extended, wantExtended)
}

func TestListRestOfSingletonIsEmpty(t *testing.T) {
	rest := ListRest(NewList(NewLong(1)))
	if got := rest.ListLen(); got != 0 {
		t.Fatalf("expected empty rest, got length %d", got)
	}
	if values := rest.ListValues(); len(values) != 0 {
		t.Fatalf("expected empty rest values, got %#v", values)
	}
}

func TestValueToAnyForList(t *testing.T) {
	listValue := NewList(NewLong(1), NewDouble(2.5))

	converted, ok := ValueToAny(listValue).([]any)
	if !ok {
		t.Fatalf("expected []any conversion, got %T", ValueToAny(listValue))
	}
	if len(converted) != 2 {
		t.Fatalf("expected 2 list items, got %d", len(converted))
	}
	if converted[0] != int64(1) {
		t.Fatalf("expected first item int64(1), got %#v", converted[0])
	}
	if converted[1] != 2.5 {
		t.Fatalf("expected second item 2.5, got %#v", converted[1])
	}
}

func checkListValues(t *testing.T, listValue Value, want []Value) {
	t.Helper()

	got := listValue.ListValues()
	if len(got) != len(want) {
		t.Fatalf("expected %d values, got %d", len(want), len(got))
	}
	for i := range want {
		if got[i].tag != want[i].tag || got[i].Long() != want[i].Long() {
			t.Fatalf("value %d mismatch: got %#v want %#v", i, got[i], want[i])
		}
	}
}
