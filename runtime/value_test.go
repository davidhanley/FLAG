package runtime

import (
	"math/big"
	"testing"
)

func TestSubAndDiv(t *testing.T) {
	sub := Sub(NewLong(10), NewDouble(2.5))
	if got := sub.Double(); got != 7.5 {
		t.Fatalf("expected subtraction result 7.5, got %v", got)
	}

	div := Div(NewLong(3), NewLong(2))
	if div.tag != TagRatio {
		t.Fatalf("expected ratio result, got tag %v", div.tag)
	}

	rat := div.Ratio()
	if rat.Cmp(big.NewRat(3, 2)) != 0 {
		t.Fatalf("expected division result 3/2, got %s", rat.RatString())
	}
}

func TestValueToAnyForRatio(t *testing.T) {
	div := Div(NewLong(3), NewLong(2))
	rat, ok := ValueToAny(div).(*big.Rat)
	if !ok {
		t.Fatalf("expected *big.Rat conversion, got %T", ValueToAny(div))
	}
	if rat.Cmp(big.NewRat(3, 2)) != 0 {
		t.Fatalf("expected converted ratio 3/2, got %s", rat.RatString())
	}
}

func TestEqNumeric(t *testing.T) {
	cases := []struct {
		name string
		lhs  Value
		rhs  Value
		want bool
	}{
		{name: "long eq long", lhs: NewLong(2), rhs: NewLong(2), want: true},
		{name: "long eq double", lhs: NewLong(3), rhs: NewDouble(3.0), want: true},
		{name: "ratio eq long", lhs: NewRatio(4, 2), rhs: NewLong(2), want: true},
		{name: "ratio eq double", lhs: NewRatio(3, 2), rhs: NewDouble(1.5), want: true},
		{name: "not equal", lhs: NewLong(1), rhs: NewLong(2), want: false},
	}

	for _, tc := range cases {
		if got := Eq(tc.lhs, tc.rhs); got != tc.want {
			t.Fatalf("%s: expected %v, got %v", tc.name, tc.want, got)
		}
	}
}
