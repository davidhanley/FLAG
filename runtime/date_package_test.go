package runtime

import "testing"

func TestDatePackageFunctionsAreRegistered(t *testing.T) {
	for _, name := range []string{"date/from-string", "c/from-string"} {
		got := Call(GoFunction(name), NewString("2026-3-8"))
		if got.tag != TagDate {
			t.Fatalf("expected date result from %s, got %#v", name, got)
		}
		if ValueToString(got) != "{:year 2026 :month 3 :day 8}" {
			t.Fatalf("unexpected date rendering from %s: %q", name, ValueToString(got))
		}
	}
}
