package runtime

import (
	"math"
	"math/big"
	"testing"
)

func TestBitAndOrXor(t *testing.T) {
	requireLong(t, BitAnd(NewLong(1), NewLong(3)), 1)
	requireLong(t, BitAnd(NewLong(1), NewLong(3), NewLong(7)), 1)
	requireLong(t, BitAnd(NewLong(2), NewLong(3), NewLong(7)), 2)
	requireLong(t, BitOr(NewLong(1), NewLong(2), NewLong(4)), 7)
	requireLong(t, BitXor(NewLong(5), NewLong(3)), 6)
	requireLong(t, BitXor(NewLong(1), NewLong(2), NewLong(3)), 0)
	requireLong(t, BitAnd(NewBigInt(1), NewLong(3)), 1)
	assertPanics(t, func() { BitAnd(NewLong(1)) })
	assertPanics(t, func() { BitOr(NewDouble(1), NewLong(1)) })
	assertPanics(t, func() { BitXor(NewRatio(1, 1), NewLong(1)) })
	assertPanics(t, func() { BitAnd(NewString("x"), NewLong(1)) })
	huge := NewBigIntFromBigInt(new(big.Int).Lsh(big.NewInt(1), 70))
	assertPanics(t, func() { BitAnd(huge, NewLong(1)) })
}

func TestBitNot(t *testing.T) {
	requireLong(t, BitNot(NewLong(0)), -1)
	requireLong(t, BitNot(NewLong(-1)), 0)
	requireLong(t, BitNot(NewLong(1)), -2)
	requireLong(t, BitNot(NewBigInt(0)), -1)
	assertPanics(t, func() { BitNot(NewDouble(1)) })
	assertPanics(t, func() { BitNot(NilValue()) })
}

func TestBitShifts(t *testing.T) {
	requireLong(t, BitShiftLeft(NewLong(1), NewLong(2)), 4)
	requireLong(t, BitShiftLeft(NewLong(1), NewLong(0)), 1)
	requireLong(t, BitShiftLeft(NewLong(1), NewLong(64)), 1) // n & 63, Clojure/Java
	requireLong(t, BitShiftLeft(NewLong(1), NewLong(-1)), math.MinInt64)
	requireLong(t, BitShiftRight(NewLong(8), NewLong(1)), 4)
	requireLong(t, BitShiftRight(NewLong(-8), NewLong(2)), -2)
	requireLong(t, BitShiftRight(NewLong(-1), NewLong(1)), -1)
	requireLong(t, UnsignedBitShiftRight(NewLong(-8), NewLong(2)), 4611686018427387902)
	requireLong(t, UnsignedBitShiftRight(NewLong(-1), NewLong(1)), math.MaxInt64)
	requireLong(t, UnsignedBitShiftRight(NewLong(8), NewLong(1)), 4)
	requireLong(t, UnsignedBitShiftRight(NewLong(-1), NewLong(64)), -1)
	assertPanics(t, func() { BitShiftLeft(NewDouble(1), NewLong(1)) })
	assertPanics(t, func() { BitShiftRight(NewLong(1), NewString("n")) })
	assertPanics(t, func() { UnsignedBitShiftRight(NewRatio(1, 2), NewLong(1)) })
}

func TestBitTestSetClearFlip(t *testing.T) {
	if got := BitTest(NewLong(2), NewLong(1)); got.tag != TagBool || !got.Bool() {
		t.Fatalf("expected bit-test 2 1 true, got %#v", got)
	}
	if got := BitTest(NewLong(2), NewLong(0)); got.tag != TagBool || got.Bool() {
		t.Fatalf("expected bit-test 2 0 false, got %#v", got)
	}
	if got := BitTest(NewLong(1), NewLong(64)); got.tag != TagBool || !got.Bool() {
		t.Fatalf("expected bit-test 1 64 true (n & 63), got %#v", got)
	}
	requireLong(t, BitSet(NewLong(0), NewLong(1)), 2)
	requireLong(t, BitSet(NewLong(1), NewLong(0)), 1)
	requireLong(t, BitSet(NewLong(0), NewLong(63)), math.MinInt64)
	requireLong(t, BitClear(NewLong(3), NewLong(0)), 2)
	requireLong(t, BitClear(NewLong(2), NewLong(1)), 0)
	requireLong(t, BitFlip(NewLong(0), NewLong(0)), 1)
	requireLong(t, BitFlip(NewLong(1), NewLong(0)), 0)
	requireLong(t, BitFlip(NewLong(0), NewLong(1)), 2)
	assertPanics(t, func() { BitTest(NewDouble(2), NewLong(1)) })
	assertPanics(t, func() { BitSet(NewLong(0), NewString("n")) })
	assertPanics(t, func() { BitClear(NilValue(), NewLong(0)) })
	assertPanics(t, func() { BitFlip(NewLong(0), NewRatio(1, 2)) })
}

func TestBitOpsAsBuiltinFunctions(t *testing.T) {
	requireLong(t, Call(BuiltinFunction("bit-and"), NewLong(6), NewLong(3)), 2)
	requireLong(t, Call(BuiltinFunction("bit-or"), NewLong(1), NewLong(4)), 5)
	requireLong(t, Call(BuiltinFunction("bit-xor"), NewLong(7), NewLong(1)), 6)
	requireLong(t, Call(BuiltinFunction("bit-not"), NewLong(0)), -1)
	requireLong(t, Call(BuiltinFunction("bit-shift-left"), NewLong(1), NewLong(3)), 8)
	requireLong(t, Call(BuiltinFunction("bit-shift-right"), NewLong(-8), NewLong(2)), -2)
	requireLong(t, Call(BuiltinFunction("unsigned-bit-shift-right"), NewLong(-8), NewLong(2)), 4611686018427387902)
	if got := Call(BuiltinFunction("bit-test"), NewLong(8), NewLong(3)); got.tag != TagBool || !got.Bool() {
		t.Fatalf("expected bit-test builtin true, got %#v", got)
	}
	requireLong(t, Call(BuiltinFunction("bit-set"), NewLong(0), NewLong(3)), 8)
	requireLong(t, Call(BuiltinFunction("bit-clear"), NewLong(8), NewLong(3)), 0)
	requireLong(t, Call(BuiltinFunction("bit-flip"), NewLong(8), NewLong(3)), 0)
	assertPanicContains(t, "exactly one", func() { Call(BuiltinFunction("bit-not")) })
	assertPanicContains(t, "exactly two", func() { Call(BuiltinFunction("bit-set"), NewLong(1)) })
	assertPanicContains(t, "at least two", func() { Call(BuiltinFunction("bit-and"), NewLong(1)) })
}
