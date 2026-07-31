package packages

import (
	"testing"
	"time"
)

func TestDateFromString(t *testing.T) {
	got := DateFromString("2026-3-8")
	ts, ok := got.(time.Time)
	if !ok {
		t.Fatalf("expected time.Time, got %T", got)
	}
	if ts.UTC().Year() != 2026 || ts.UTC().Month() != 3 || ts.UTC().Day() != 8 {
		t.Fatalf("unexpected parsed date: %v", ts)
	}
	if DateFromString("nope") != nil {
		t.Fatal("expected nil for invalid date")
	}
}
