package packages

import "testing"

func TestStringTrim(t *testing.T) {
	if got := StringTrim("  hello \n"); got != "hello" {
		t.Fatalf("expected trimmed string, got %q", got)
	}
}

func TestStringReplace(t *testing.T) {
	if got := StringReplace("hello world", "world", "FLAG"); got != "hello FLAG" {
		t.Fatalf("expected replaced string, got %q", got)
	}
}

func TestStringEscape(t *testing.T) {
	cmap := map[any]any{
		"&": "&amp;",
		"<": "&lt;",
		">": "&gt;",
	}
	if got := StringEscape("3 < 5 & 8 > 2", cmap); got != "3 &lt; 5 &amp; 8 &gt; 2" {
		t.Fatalf("unexpected escaped string: %q", got)
	}
	if got := StringEscape("", cmap); got != "" {
		t.Fatalf("expected empty escaped string, got %q", got)
	}
}

func TestStringSplit(t *testing.T) {
	if got := StringSplit("a,b,c", ","); len(got) != 3 || got[0] != "a" || got[2] != "c" {
		t.Fatalf("unexpected split result: %#v", got)
	}
	if got := StringSplit("a,b,c", ",", 2); len(got) != 2 || got[0] != "a" || got[1] != "b,c" {
		t.Fatalf("unexpected limited split result: %#v", got)
	}
}

func TestStringJoin(t *testing.T) {
	if got := StringJoin([]any{"a", "b", "c"}); got != "abc" {
		t.Fatalf("unexpected joined result without separator: %q", got)
	}
	if got := StringJoin("-", []any{"male", "18-34", "overall"}); got != "male-18-34-overall" {
		t.Fatalf("unexpected joined result with separator: %q", got)
	}
	if got := StringJoin("-", []string{"a", "b"}); got != "a-b" {
		t.Fatalf("unexpected []string joined result: %q", got)
	}
	if got := StringJoin("-", nil); got != "" {
		t.Fatalf("unexpected joined result for nil collection: %q", got)
	}
}

func TestStringBlankAndPrefix(t *testing.T) {
	if !StringBlank("   ") || StringBlank("hello") {
		t.Fatalf("unexpected blank checks")
	}
	if !StringStartsWith("hello world", "hello") || StringStartsWith("hello world", "world") {
		t.Fatalf("unexpected prefix checks")
	}
	if !StringEndsWith("hello world", "world") || StringEndsWith("hello world", "hello") {
		t.Fatalf("unexpected suffix checks")
	}
}

func TestStringUpperCase(t *testing.T) {
	if got := StringUpperCase("hello"); got != "HELLO" {
		t.Fatalf("expected upper case string, got %q", got)
	}
}

func TestStringCapitalize(t *testing.T) {
	if got := StringCapitalize("hELLO"); got != "Hello" {
		t.Fatalf("expected capitalized string, got %q", got)
	}
	if got := StringCapitalize(""); got != "" {
		t.Fatalf("expected empty string, got %q", got)
	}
}
