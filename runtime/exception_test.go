package runtime

import "testing"

func TestPanicValue(t *testing.T) {
	if got := PanicValue(nil); !IsNil(got) {
		t.Fatalf("nil panic: %v", got.String())
	}
	v := NewString("nope")
	if got := PanicValue(v); !Eq(got, v) {
		t.Fatalf("Value panic: %v", got.String())
	}
	if got := PanicValue("quot by zero"); !isType[*stringImpl](got) || got.StringValue() != "quot by zero" {
		t.Fatalf("string panic: %v", got.String())
	}
}

func TestExInfoMessageDataCause(t *testing.T) {
	info := NewMap(exMessageKey, NewString("boom"), exDataKey, NewMap(NewKeyword("a"), NewLong(1)))
	if !IsExInfo(info) {
		t.Fatal("expected ExceptionInfo")
	}
	if got := ExMessage(info); !isType[*stringImpl](got) || got.StringValue() != "boom" {
		t.Fatalf("ex-message: %v", got.String())
	}
	if got := ExData(info); !Eq(got, NewMap(NewKeyword("a"), NewLong(1))) {
		t.Fatalf("ex-data: %v", got.String())
	}
	if got := ExCause(info); !IsNil(got) {
		t.Fatalf("ex-cause empty: %v", got.String())
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
	if got := ExMessage(NewString("nope")); !isType[*stringImpl](got) || got.StringValue() != "nope" {
		t.Fatalf("string ex-message: %v", got.String())
	}
	if got := ExMessage(NilValue); !IsNil(got) {
		t.Fatalf("nil ex-message: %v", got.String())
	}
	if got := ExData(NewString("x")); !IsNil(got) {
		t.Fatalf("string ex-data: %v", got.String())
	}
}
