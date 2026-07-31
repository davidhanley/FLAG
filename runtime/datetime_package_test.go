package runtime

import "testing"

func TestDateTimePackageFunctionsAreRegistered(t *testing.T) {
	formatter := Call(GoFunction("datetime/formatter"), NewString("yyyy-MM-dd"))
	if formatter.tag != TagString || formatter.StringValue() != "2006-01-02" {
		t.Fatalf("unexpected datetime/formatter result: %#v", formatter)
	}

	formatted := Call(GoFunction("datetime/unparse"), formatter, NewDateFromYMD(2026, 3, 8))
	if formatted.tag != TagString || formatted.StringValue() != "2026-03-08" {
		t.Fatalf("unexpected datetime/unparse result: %#v", formatted)
	}

	now := Call(GoFunction("datetime/now"))
	if now.tag != TagDate {
		t.Fatalf("expected datetime/now to return date, got %#v", now)
	}
}
