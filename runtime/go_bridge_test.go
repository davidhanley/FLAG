package runtime

import (
	"fmt"
	"testing"

	"github.com/traefik/yaegi/stdlib"
)

func TestGoFunctionCallByName(t *testing.T) {
	RegisterGoFunction("test.add", func(a int64, b int64) int64 { return a + b })

	fn := GoFunction("test.add")
	got := Call(fn, NewLong(2), NewLong(5))
	if got.tag != TagLong || got.Long() != 7 {
		t.Fatalf("expected 7 from go function call, got %#v", got)
	}
}

func TestGoFunctionArgs(t *testing.T) {
	RegisterGoFunction("test.meta", func(a int64, b ...int64) int64 { return a + int64(len(b)) })

	meta := GoFunctionArgs("test.meta")
	if meta.tag != TagMap {
		t.Fatalf("expected map metadata, got %v", meta.tag)
	}

	params := Get(meta, NewKeyword("params"))
	if params.tag != TagArray || params.ArrayLen() != 2 {
		t.Fatalf("expected 2 params, got %#v", params)
	}
	if got := ValueToString(ArrayGet(params, 0)); got != "int64" {
		t.Fatalf("unexpected first param type: %q", got)
	}
	if got := ValueToString(ArrayGet(params, 1)); got != "int64..." {
		t.Fatalf("unexpected variadic param type: %q", got)
	}

	variadic := Get(meta, NewKeyword("variadic"))
	if variadic.tag != TagBool || !variadic.Bool() {
		t.Fatalf("expected variadic=true, got %#v", variadic)
	}
}

func TestGoFunctionVariadicCall(t *testing.T) {
	RegisterGoFunction("test.format", func(format string, args ...any) string {
		return fmt.Sprintf(format, args...)
	})

	fn := GoFunction("test.format")
	got := Call(fn, NewSymbol("david is %d years old"), NewLong(23))
	if got.tag != TagString || got.StringValue() != "david is 23 years old" {
		t.Fatalf("expected formatted string value, got %#v", got)
	}
}

func TestGoFunctionVariadicCallFromStdlibSymbols(t *testing.T) {
	RegisterGoSymbols(stdlib.Symbols)

	fn := GoFunction("fmt.Sprintf")
	got := Call(fn, NewSymbol("david is %d years old"), NewLong(23))
	if got.tag != TagString || got.StringValue() != "david is 23 years old" {
		t.Fatalf("expected formatted string value, got %#v", got)
	}
}

type interopPerson struct {
	Name      string
	FirstName string
	Age       int64
}

func TestGoFunctionMapRecordConvertsToStruct(t *testing.T) {
	RegisterGoFunction("test.struct-age", func(person interopPerson) int64 { return person.Age })

	fn := GoFunction("test.struct-age")
	got := Call(fn, NewMap(
		NewKeyword("name"), NewString("Ada"),
		NewKeyword("age"), NewLong(36),
	))
	if got.tag != TagLong || got.Long() != 36 {
		t.Fatalf("expected age 36, got %#v", got)
	}
}

func TestGoFunctionMapRecordConvertsHyphenatedKeysToStructField(t *testing.T) {
	RegisterGoFunction("test.struct-first-name", func(person interopPerson) string { return person.FirstName })

	fn := GoFunction("test.struct-first-name")
	got := Call(fn, NewMap(
		NewKeyword("first-name"), NewString("Grace"),
	))
	if got.tag != TagString || got.StringValue() != "Grace" {
		t.Fatalf("expected first name Grace, got %#v", got)
	}
}
