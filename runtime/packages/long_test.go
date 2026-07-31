package packages

import "testing"

func TestLongParse(t *testing.T) {
	if got := LongParse(" 42 "); got == nil || got.(int64) != 42 {
		t.Fatalf("expected parsed long 42, got %#v", got)
	}
	if got := LongParse("nope"); got != nil {
		t.Fatalf("expected nil for invalid parse, got %#v", got)
	}
}
