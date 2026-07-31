package runtime

import (
	"os"
	"path/filepath"
	"testing"
)

func TestJSONRoundTrip(t *testing.T) {
	original := NewMap(
		NewKeyword("a"), NewLong(1),
		NewKeyword("b"), NewArray(NewLong(2), NewLong(3), NewString("x")),
	)
	roundTrip := FromJSON(ToJSON(original))
	if !Eq(roundTrip, original) {
		t.Fatalf("expected round-trip JSON value to match original; got %s want %s", ValueToString(roundTrip), ValueToString(original))
	}
}

func TestFromJSONLiteralArray(t *testing.T) {
	decoded := FromJSON("[1,2]")
	if !Eq(decoded, NewArray(NewLong(1), NewLong(2))) {
		t.Fatalf("expected decoded array to match expected value; got %s", ValueToString(decoded))
	}
}

func TestFromJSONStringLiteral(t *testing.T) {
	decoded := FromJSON(`"hello"`)
	if !Eq(decoded, NewString("hello")) {
		t.Fatalf("expected decoded string to match expected value; got %s", ValueToString(decoded))
	}
}

func TestJSONReadStrFromStringWithTransforms(t *testing.T) {
	valueFn := NewFunction(func(args ...Value) Value {
		if len(args) != 2 {
			panic("value-fn expects key and value")
		}
		if Eq(args[0], NewKeyword("answer")) {
			return Add(args[1], NewLong(1))
		}
		return args[1]
	})

	decoded := JSONReadStr(
		NewString(`{"answer":41,"nested":{"answer":1}}`),
		NewKeyword("key-fn"), BuiltinFunction("keyword"),
		NewKeyword("value-fn"), valueFn,
	)

	expected := NewMap(
		NewKeyword("answer"), NewLong(42),
		NewKeyword("nested"), NewMap(
			NewKeyword("answer"), NewLong(2),
		),
	)
	if !Eq(decoded, expected) {
		t.Fatalf("expected transformed JSON to match; got %s want %s", ValueToString(decoded), ValueToString(expected))
	}
}

func TestJSONReadFromFileLeavesReaderAtNextObject(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "stream.json")
	if err := os.WriteFile(path, []byte("{\"a\":1}\n{\"b\":2}\n"), 0o600); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	file := OpenFile(path)
	defer func() {
		if err := file.Close(); err != nil {
			t.Fatalf("close file: %v", err)
		}
	}()

	first := JSONRead(file, NewKeyword("key-fn"), BuiltinFunction("keyword"))
	second := JSONReadStr(file, NewKeyword("key-fn"), BuiltinFunction("keyword"))
	third := JSONRead(file, NewKeyword("key-fn"), BuiltinFunction("keyword"))

	if !Eq(first, NewMap(NewKeyword("a"), NewLong(1))) {
		t.Fatalf("unexpected first JSON value: %s", ValueToString(first))
	}
	if !Eq(second, NewMap(NewKeyword("b"), NewLong(2))) {
		t.Fatalf("unexpected second JSON value: %s", ValueToString(second))
	}
	if third.tag != TagNil {
		t.Fatalf("expected nil at eof, got %#v", third)
	}
}

func TestJSONReadFunctionsAreRegistered(t *testing.T) {
	readStr := Call(GoFunction("json/read-str"), NewString(`{"a":1}`), NewKeyword("key-fn"), BuiltinFunction("keyword"))
	read := Call(GoFunction("json/read"), NewString(`{"b":2}`), NewKeyword("key-fn"), BuiltinFunction("keyword"))

	if !Eq(readStr, NewMap(NewKeyword("a"), NewLong(1))) {
		t.Fatalf("unexpected json/read-str result: %s", ValueToString(readStr))
	}
	if !Eq(read, NewMap(NewKeyword("b"), NewLong(2))) {
		t.Fatalf("unexpected json/read result: %s", ValueToString(read))
	}
}
