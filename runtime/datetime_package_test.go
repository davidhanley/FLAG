package runtime

import "testing"

func TestDateTimePackageFunctionsAreRegistered(t *testing.T) {
	formatter := Call(GoFunction("datetime/formatter"), NewString("yyyy-MM-dd"))
	if !isType[*stringImpl](formatter) || formatter.StringValue() != "2006-01-02" {
		t.Fatalf("unexpected datetime/formatter result: %#v", formatter)
	}

	formatted := Call(GoFunction("datetime/unparse"), formatter, NewDateFromYMD(2026, 3, 8))
	if !isType[*stringImpl](formatted) || formatted.StringValue() != "2026-03-08" {
		t.Fatalf("unexpected datetime/unparse result: %#v", formatted)
	}

	now := Call(GoFunction("datetime/now"))
	if !isType[*dateImpl](now) {
		t.Fatalf("expected datetime/now to return date, got %#v", now)
	}
}

func TestTimeNanoTimeIsMonotonicLong(t *testing.T) {
	a := TimeNanoTime()
	b := TimeNanoTime()
	if !isType[longImpl](a) || !isType[longImpl](b) {
		t.Fatalf("time/nano-time should return longs, got %#v %#v", a, b)
	}
	if b.Long() < a.Long() {
		t.Fatalf("time/nano-time moved backwards: %d then %d", a.Long(), b.Long())
	}
	viaName := Call(GoFunction("time/nano-time"))
	if !isType[longImpl](viaName) {
		t.Fatalf("registered time/nano-time should return long, got %#v", viaName)
	}
	bound := GoBind_runtime_TimeNanoTime.Fn()
	if !isType[longImpl](bound) {
		t.Fatalf("GoBind time/nano-time should return long, got %#v", bound)
	}
}
