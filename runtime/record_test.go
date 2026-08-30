package runtime

import "testing"

type testSourceToken struct {
	Token  string `flag:"token"`
	Line   int64  `flag:"line"`
	Offset int64  `flag:"offset"`
}

func TestRecordKeywordLookup(t *testing.T) {
	rec := NewRecord(testSourceToken{Token: "foo", Line: 3, Offset: 7})
	got := Get(rec, NewKeyword("token"))
	if got.StringValue() != "foo" {
		t.Fatalf("token: %q", got.StringValue())
	}
	if Get(rec, NewKeyword("line")).Long() != 3 {
		t.Fatalf("line: %d", Get(rec, NewKeyword("line")).Long())
	}
	native := GoStruct[testSourceToken](rec)
	if native.Token != "foo" || native.Offset != 7 {
		t.Fatalf("GoStruct: %+v", native)
	}
}
