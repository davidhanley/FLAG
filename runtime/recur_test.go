package runtime

import "testing"

func TestUnwrapRecurRoundTrip(t *testing.T) {
	marker := NewRecur(NewLong(1), NewString("x"))
	values, ok := UnwrapRecur(marker)
	if !ok {
		t.Fatal("expected recur marker to unwrap")
	}
	if len(values) != 2 {
		t.Fatalf("expected 2 recur values, got %d", len(values))
	}
	if values[0].tag != TagLong || values[0].Long() != 1 {
		t.Fatalf("unexpected first recur value: %#v", values[0])
	}
	if values[1].tag != TagString || values[1].StringValue() != "x" {
		t.Fatalf("unexpected second recur value: %#v", values[1])
	}
}

func TestUnwrapRecurNonMarker(t *testing.T) {
	if _, ok := UnwrapRecur(NewLong(9)); ok {
		t.Fatal("expected non-recur Value not to unwrap")
	}
}
