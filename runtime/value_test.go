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
