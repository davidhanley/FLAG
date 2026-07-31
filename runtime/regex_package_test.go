package runtime

import "testing"

func TestRegexPackageFunctionsAreRegistered(t *testing.T) {
	compile := GoFunction("regex/compile")
	matcher := Call(compile, NewString("^he.*o$"))
	if matcher.tag != TagFunction {
		t.Fatalf("expected matcher function, got %#v", matcher)
	}
	if got := Call(matcher, NewString("hello")); got.tag != TagBool || !got.Bool() {
		t.Fatalf("expected match for hello, got %#v", got)
	}

	if got := Call(GoFunction("re-pattern"), NewString("^he.*o$")); got.tag != TagFunction {
		t.Fatalf("expected re-pattern matcher function, got %#v", got)
	}
	if got := Call(GoFunction("re-matches"), matcher, NewString("hello")); got.tag != TagBool || !got.Bool() {
		t.Fatalf("expected re-matches truthy result, got %#v", got)
	}
	if got := Call(GoFunction("re-matches"), matcher, NewString("bye")); got.tag != TagBool || got.Bool() {
		t.Fatalf("expected re-matches false result, got %#v", got)
	}

	caseInsensitive := Call(compile, NewString("^hello$"), NewLong(0x02))
	if caseInsensitive.tag != TagFunction {
		t.Fatalf("expected regex/compile with flags to return matcher function, got %#v", caseInsensitive)
	}
	if got := Call(caseInsensitive, NewString("HELLO")); got.tag != TagBool || !got.Bool() {
		t.Fatalf("expected case-insensitive match for HELLO, got %#v", got)
	}

	if got := Call(compile, matcher, NewString("hello")); got.tag != TagBool || !got.Bool() {
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
		return NilValue()
	}), values); got.tag != TagString || got.StringValue() != "hello" {
		t.Fatalf("unexpected some regex result: %#v", got)
	}
}
