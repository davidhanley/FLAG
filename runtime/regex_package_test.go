package runtime

import "testing"

func TestRegexPackageFunctionsAreRegistered(t *testing.T) {
	compile := GoFunction("regex/compile")
	matcher := Call(compile, NewString("^he.*o$"))
	if !isType[*functionImpl](matcher) {
		t.Fatalf("expected matcher function, got %#v", matcher)
	}
	if got := Call(matcher, NewString("hello")); !isType[boolImpl](got) || !got.Bool() {
		t.Fatalf("expected match for hello, got %#v", got)
	}

	if got := Call(GoFunction("re-pattern"), NewString("^he.*o$")); !isType[*functionImpl](got) {
		t.Fatalf("expected re-pattern matcher function, got %#v", got)
	}
	if got := Call(GoFunction("re-matches"), matcher, NewString("hello")); !isType[boolImpl](got) || !got.Bool() {
		t.Fatalf("expected re-matches truthy result, got %#v", got)
	}
	if got := Call(GoFunction("re-matches"), matcher, NewString("bye")); !isType[boolImpl](got) || got.Bool() {
		t.Fatalf("expected re-matches false result, got %#v", got)
	}

	caseInsensitive := Call(compile, NewString("^hello$"), NewLong(0x02))
	if !isType[*functionImpl](caseInsensitive) {
		t.Fatalf("expected regex/compile with flags to return matcher function, got %#v", caseInsensitive)
	}
	if got := Call(caseInsensitive, NewString("HELLO")); !isType[boolImpl](got) || !got.Bool() {
		t.Fatalf("expected case-insensitive match for HELLO, got %#v", got)
	}

	if got := Call(compile, matcher, NewString("hello")); !isType[boolImpl](got) || !got.Bool() {
		t.Fatalf("expected two-arity regex/compile matcher+value to return true, got %#v", got)
	}
}

func TestSomeWorksWithRegexMatcher(t *testing.T) {
	matcher := Call(GoFunction("regex/compile"), NewString("^he.*o$"))
	values := NewArray(NewString("bye"), NewString("hello"))
	if got := Some(NewFunction(func(args ...Value) Value {
		if IsTruthy(Call(matcher, args[0])) {
			return args[0]
		}
		return NilValue
	}), values); !isType[*stringImpl](got) || got.StringValue() != "hello" {
		t.Fatalf("unexpected some regex result: %#v", got)
	}
}
