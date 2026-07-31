package runtime

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStringPackageFunctionsAreRegistered(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []Value
		want string
	}{
		{name: "string/trim", args: []Value{NewString("  hello  ")}, want: "hello"},
		{name: "str/trim", args: []Value{NewString("  hello  ")}, want: "hello"},
		{name: "string/replace", args: []Value{NewString("hello world"), NewString("world"), NewString("FLAG")}, want: "hello FLAG"},
		{name: "str/replace", args: []Value{NewString("hello world"), NewString("world"), NewString("FLAG")}, want: "hello FLAG"},
		{name: "str/escape", args: []Value{NewString("a & b"), NewMap(NewString("&"), NewString("&amp;"))}, want: "a &amp; b"},
		{name: "string/escape", args: []Value{NewString("a & b"), NewMap(NewString("&"), NewString("&amp;"))}, want: "a &amp; b"},
		{name: "str/join", args: []Value{NewString("-"), NewArray(NewString("male"), NewString("18-34"), NewString("overall"))}, want: "male-18-34-overall"},
		{name: "string/join", args: []Value{NewString("-"), NewArray(NewString("male"), NewString("18-34"), NewString("overall"))}, want: "male-18-34-overall"},
		{name: "str/upper-case", args: []Value{NewString("hello")}, want: "HELLO"},
		{name: "string/upper-case", args: []Value{NewString("hello")}, want: "HELLO"},
		{name: "str/capitalize", args: []Value{NewString("hELLO")}, want: "Hello"},
		{name: "string/capitalize", args: []Value{NewString("hELLO")}, want: "Hello"},
		{name: "str/starts-with?", args: []Value{NewString("hello world"), NewString("hello")}, want: "true"},
		{name: "str/ends-with?", args: []Value{NewString("hello world"), NewString("world")}, want: "true"},
		{name: "str/blank?", args: []Value{NewString("   ")}, want: "true"},
		{name: "character/toUppercase", args: []Value{NewString("hello")}, want: "HELLO"},
		{name: "Character/toUpperCase", args: []Value{NewString("hello")}, want: "HELLO"},
		{name: "long/parse", args: []Value{NewString("42")}, want: "42"},
	} {
		fn := GoFunction(tc.name)
		got := Call(fn, tc.args...)
		switch tc.name {
		case "long/parse":
			if got.tag != TagLong || got.Long() != 42 {
				t.Fatalf("unexpected parse result: %#v", got)
			}
		case "str/starts-with?", "str/ends-with?", "str/blank?":
			if got.tag != TagBool || got.Bool() != (tc.want == "true") {
				t.Fatalf("unexpected predicate result for %s: %#v", tc.name, got)
			}
		default:
			if got.tag != TagString || got.StringValue() != tc.want {
				t.Fatalf("unexpected %s result: %#v", tc.name, got)
			}
		}
	}
	if got := Call(GoFunction("long/parse"), NewString("nope")); got.tag != TagNil {
		t.Fatalf("expected nil for invalid parse, got %#v", got)
	}
}

func TestStringSplitFunctionsAreRegistered(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []Value
		want []string
	}{
		{name: "str/split", args: []Value{NewString("a,b,c"), NewString(",")}, want: []string{"a", "b", "c"}},
		{name: "string/split", args: []Value{NewString("a,b,c"), NewString(","), NewLong(2)}, want: []string{"a", "b,c"}},
	} {
		got := Call(GoFunction(tc.name), tc.args...)
		if got.tag != TagArray || got.ArrayLen() != len(tc.want) {
			t.Fatalf("unexpected split result shape for %s: %#v", tc.name, got)
		}
		for i, want := range tc.want {
			item := ArrayGet(got, i)
			if item.tag != TagString || item.StringValue() != want {
				t.Fatalf("unexpected split result at %d for %s: %#v", i, tc.name, got)
			}
		}
	}
}

func TestCharacterPackageFunctionIsRegistered(t *testing.T) {
	fn := GoFunction("Character/toUpperCase")
	if got := Call(fn, NewString("hello")); got.tag != TagString || got.StringValue() != "HELLO" {
		t.Fatalf("unexpected uppercase result: %#v", got)
	}
}

func TestIOFunctionsAreRegistered(t *testing.T) {
	lineSeqFn := GoFunction("line-seq")
	dir := t.TempDir()
	path := dir + "/sample.txt"
	if err := os.WriteFile(path, []byte("x\ny\n"), 0o600); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	file := OpenFile(path)
	defer func() { _ = file.Close() }()
	got := Call(lineSeqFn, file)
	if got.tag != TagLazyList {
		t.Fatalf("expected lazy list from line-seq, got %#v", got)
	}
	if first := First(got); ValueToString(first) != "x" {
		t.Fatalf("unexpected first line result: %#v", first)
	}

	subdir := filepath.Join(dir, "nested")
	if err := os.MkdirAll(subdir, 0o755); err != nil {
		t.Fatalf("create nested dir: %v", err)
	}
	nestedFile := filepath.Join(subdir, "race.csv")
	if err := os.WriteFile(nestedFile, []byte("a,b\n"), 0o600); err != nil {
		t.Fatalf("write nested file: %v", err)
	}

	scanFn := GoFunction("io/scan-directory")
	scanned := Call(scanFn, NewString(dir))
	if scanned.tag != TagArray {
		t.Fatalf("expected array from io/scan-directory, got %#v", scanned)
	}
	if scanned.ArrayLen() < 2 {
		t.Fatalf("expected at least 2 files from io/scan-directory, got %d", scanned.ArrayLen())
	}

	var foundNested bool
	for _, item := range scanned.ArrayValues() {
		if item.tag != TagMap {
			t.Fatalf("expected file metadata map entries, got %#v", item)
		}
		filename := Get(item, NewKeyword("filename"))
		name := Get(item, NewKeyword("name"))
		size := Get(item, NewKeyword("size"))
		created := Get(item, NewKeyword("created-date-time"))
		updated := Get(item, NewKeyword("updated-date-time"))
		if filename.tag != TagString || name.tag != TagString || size.tag != TagLong || created.tag != TagString || updated.tag != TagString {
			t.Fatalf("unexpected metadata map: %#v", item)
		}
		if filename.StringValue() == nestedFile {
			foundNested = true
		}
	}
	if !foundNested {
		t.Fatalf("expected nested file %q in scan output", nestedFile)
	}
}
