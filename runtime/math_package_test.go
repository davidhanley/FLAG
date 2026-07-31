package runtime

import "testing"

func TestMathAbsFunctionIsRegistered(t *testing.T) {
	if got := Call(GoFunction("math/abs"), NewLong(-42)); got.tag != TagLong || got.Long() != 42 {
		t.Fatalf("unexpected math/abs long result: %#v", got)
	}
	if got := Call(GoFunction("math/abs"), NewDouble(-2.5)); got.tag != TagDouble || got.Double() != 2.5 {
		t.Fatalf("unexpected math/abs double result: %#v", got)
	}
}
