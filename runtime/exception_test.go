package runtime

import "testing"

func TestPanicValue(t *testing.T) {
	if got := PanicValue(nil); got.tag != TagNil {
		t.Fatalf("nil panic: %v", ValueToString(got))
	}
	v := NewString("nope")
	if got := PanicValue(v); !Eq(got, v) {
		t.Fatalf("Value panic: %v", ValueToString(got))
	}
	if got := PanicValue("quot by zero"); got.tag != TagString || got.StringValue() != "quot by zero" {
		t.Fatalf("string panic: %v", ValueToString(got))
	}
}

func TestExInfoMessageDataCause(t *testing.T) {
	info := NewMap(exMessageKey, NewString("boom"), exDataKey, NewMap(NewKeyword("a"), NewLong(1)))
	if !IsExInfo(info) {
		t.Fatal("expected ExceptionInfo")
	}
	if got := ExMessage(info); got.tag != TagString || got.StringValue() != "boom" {
		t.Fatalf("ex-message: %v", ValueToString(got))
	}
	if got := ExData(info); !Eq(got, NewMap(NewKeyword("a"), NewLong(1))) {
		t.Fatalf("ex-data: %v", ValueToString(got))
	}
	if got := ExCause(info); got.tag != TagNil {
		t.Fatalf("ex-cause empty: %v", ValueToString(got))
	}
	if !CatchMatches("ExceptionInfo", info) || !CatchMatches("Exception", info) {
		t.Fatal("ExceptionInfo should match ExceptionInfo and Exception")
	}
	if CatchMatches("ExceptionInfo", NewString("x")) {
		t.Fatal("string should not match ExceptionInfo")
	}
	if !CatchMatches(":default", NewString("x")) {
		t.Fatal(":default should match any value")
	}
	if got := ExMessage(NewString("nope")); got.tag != TagString || got.StringValue() != "nope" {
		t.Fatalf("string ex-message: %v", ValueToString(got))
	}
	if got := ExMessage(NilValue()); got.tag != TagNil {
		t.Fatalf("nil ex-message: %v", ValueToString(got))
	}
	if got := ExData(NewString("x")); got.tag != TagNil {
		t.Fatalf("string ex-data: %v", ValueToString(got))
	}
}
