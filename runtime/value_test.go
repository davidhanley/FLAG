package runtime

import (
	"math/big"
	"testing"
	"time"
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

func TestModIntegers(t *testing.T) {
	cases := []struct {
		name string
		lhs  int64
		rhs  int64
		want int64
	}{
		{name: "positive", lhs: 5, rhs: 3, want: 2},
		{name: "negative dividend", lhs: -5, rhs: 3, want: 1},
		{name: "negative divisor", lhs: 5, rhs: -3, want: -1},
		{name: "both negative", lhs: -5, rhs: -3, want: -2},
	}

	for _, tc := range cases {
		got := Mod(NewLong(tc.lhs), NewLong(tc.rhs))
		if got.tag != TagLong || got.Long() != tc.want {
			t.Fatalf("%s: expected %d, got %#v", tc.name, tc.want, got)
		}
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

func TestLtGtNumeric(t *testing.T) {
	if !Lt(NewLong(1), NewLong(2)) {
		t.Fatal("expected 1 < 2")
	}
	if Lt(NewLong(2), NewLong(1)) {
		t.Fatal("expected 2 !< 1")
	}
	if !Gt(NewRatio(5, 2), NewDouble(2.0)) {
		t.Fatal("expected 5/2 > 2.0")
	}
	if Gt(NewDouble(1.25), NewRatio(3, 2)) {
		t.Fatal("expected 1.25 !> 1.5")
	}
}

func TestOrderedComparisonsNumericAndString(t *testing.T) {
	if !Le(NewLong(2), NewLong(2)) {
		t.Fatal("expected 2 <= 2")
	}
	if !Ge(NewLong(3), NewLong(2)) {
		t.Fatal("expected 3 >= 2")
	}
	if !Lt(NewString("apple"), NewString("banana")) {
		t.Fatal(`expected "apple" < "banana"`)
	}
	if !Le(NewString("apple"), NewString("apple")) {
		t.Fatal(`expected "apple" <= "apple"`)
	}
	if !Gt(NewString("cat"), NewString("banana")) {
		t.Fatal(`expected "cat" > "banana"`)
	}
	if !Ge(NewString("cat"), NewString("cat")) {
		t.Fatal(`expected "cat" >= "cat"`)
	}
}

func TestAbsAcrossNumericTypes(t *testing.T) {
	if got := Abs(NewLong(-7)); got.tag != TagLong || got.Long() != 7 {
		t.Fatalf("expected abs long 7, got %#v", got)
	}

	if got := Abs(NewDouble(-3.5)); got.tag != TagDouble || got.Double() != 3.5 {
		t.Fatalf("expected abs double 3.5, got %#v", got)
	}

	if got := Abs(NewRatio(-3, 2)); got.tag != TagRatio || got.Ratio().Cmp(big.NewRat(3, 2)) != 0 {
		t.Fatalf("expected abs ratio 3/2, got %#v", got)
	}

	minAbs := Abs(NewLong(minInt64))
	if minAbs.tag != TagBigInt || minAbs.BigInt().Cmp(new(big.Int).Abs(big.NewInt(minInt64))) != 0 {
		t.Fatalf("expected bigint abs for min int64, got %#v", minAbs)
	}
}

func TestNilValueTruthiness(t *testing.T) {
	if ValueToAny(NilValue()) != nil {
		t.Fatal("expected NilValue to convert to nil")
	}
	if IsTruthy(NilValue()) {
		t.Fatal("expected NilValue to be falsey")
	}
	if !IsTruthy(NewLong(0)) {
		t.Fatal("expected numeric zero to be truthy")
	}
	if IsTruthy(NewBool(false)) {
		t.Fatal("expected false boolean Value to be falsey")
	}
	if !IsTruthy(NewBool(true)) {
		t.Fatal("expected true boolean Value to be truthy")
	}
	if got, ok := ValueToAny(NewBool(true)).(bool); !ok || !got {
		t.Fatalf("expected ValueToAny(bool) to return true, got %#v", ValueToAny(NewBool(true)))
	}
}

func TestStr(t *testing.T) {
	if got := Str(); got != "" {
		t.Fatalf("expected empty str() result, got %q", got)
	}

	value := Str(
		"a",
		NewLong(1),
		NewDouble(2.5),
		NewRatio(3, 2),
		NewBool(false),
		NewSymbol("s"),
		NewKeyword("k"),
		true,
		NewList(NewLong(4), NewLong(5)),
		NewArray(NewLong(6), NewLong(7)),
		NilValue(),
	)
	if value != "a12.53/2falses:ktrue(4 5)[6 7]" {
		t.Fatalf("unexpected str result: %q", value)
	}
}

func TestSymbolAndName(t *testing.T) {
	sym := Symbol("abc")
	if sym.tag != TagSymbol {
		t.Fatalf("expected symbol tag, got %v", sym.tag)
	}

	if got := Name(sym); got != "abc" {
		t.Fatalf("expected symbol name abc, got %q", got)
	}
	if got := Name(NewKeyword("kw")); got != "kw" {
		t.Fatalf("expected keyword name kw, got %q", got)
	}

	kwAsSymbol := Symbol(NewKeyword("kw"))
	if got := ValueToString(kwAsSymbol); got != "kw" {
		t.Fatalf("expected symbol conversion to drop keyword marker, got %q", got)
	}

	keywordAny, ok := ValueToAny(NewKeyword("key")).(Value)
	if !ok {
		t.Fatalf("expected Value from ValueToAny, got %T", ValueToAny(NewKeyword("key")))
	}
	if got := ValueToString(keywordAny); got != ":key" {
		t.Fatalf("unexpected keyword value: %q", got)
	}
}

func TestIntegerOverflowPromotesToBigInt(t *testing.T) {
	added := Add(NewLong(9223372036854775807), NewLong(1))
	if added.tag != TagBigInt {
		t.Fatalf("expected bigint from add overflow, got %v", added.tag)
	}
	if added.BigInt().Cmp(big.NewInt(0).Add(big.NewInt(9223372036854775807), big.NewInt(1))) != 0 {
		t.Fatalf("unexpected add overflow value: %s", added.BigInt().String())
	}

	subbed := Sub(NewLong(-9223372036854775808), NewLong(1))
	if subbed.tag != TagBigInt {
		t.Fatalf("expected bigint from sub overflow, got %v", subbed.tag)
	}

	mulled := Mul(NewLong(3037000500), NewLong(3037000500))
	if mulled.tag != TagBigInt {
		t.Fatalf("expected bigint from mul overflow, got %v", mulled.tag)
	}
}

func TestBigIntNumericInterop(t *testing.T) {
	huge := Add(NewLong(9223372036854775807), NewLong(1))
	if !Eq(Sub(huge, NewLong(1)), NewLong(9223372036854775807)) {
		t.Fatal("expected bigint-long arithmetic interop to preserve value")
	}
	if !Gt(huge, NewLong(9223372036854775807)) {
		t.Fatal("expected bigint compare to work")
	}

	mod := Mod(huge, NewLong(2))
	if mod.tag != TagLong || mod.Long() != 0 {
		t.Fatalf("expected bigint mod 2 => 0, got %#v", mod)
	}

	ratio := Div(huge, NewLong(2))
	if ratio.tag != TagRatio {
		t.Fatalf("expected ratio from bigint division, got %v", ratio.tag)
	}
}

func TestNewBigIntFromString(t *testing.T) {
	got := NewBigIntFromString("9223372036854775808")
	if got.tag != TagBigInt {
		t.Fatalf("expected bigint tag, got %v", got.tag)
	}
	if got.BigInt().String() != "9223372036854775808" {
		t.Fatalf("unexpected bigint value: %s", got.BigInt().String())
	}
}

func TestStringValues(t *testing.T) {
	got := NewString("hello")
	if got.tag != TagString {
		t.Fatalf("expected string tag, got %v", got.tag)
	}
	if ValueToString(got) != `"hello"` {
		t.Fatalf("expected quoted string value, got %q", ValueToString(got))
	}
	if Str(got) != "hello" {
		t.Fatalf("expected raw string from Str, got %q", Str(got))
	}
	if native, ok := ValueToAny(got).(string); !ok || native != "hello" {
		t.Fatalf("expected native string from ValueToAny, got %#v", ValueToAny(got))
	}
}

func TestDateValues(t *testing.T) {
	got := NewDateFromYMD(2026, 3, 8)
	if got.tag != TagDate {
		t.Fatalf("expected date tag, got %v", got.tag)
	}
	if ValueToString(got) != "{:year 2026 :month 3 :day 8}" {
		t.Fatalf("unexpected date string: %q", ValueToString(got))
	}
	if year := Get(got, NewKeyword("year")); year.Long() != 2026 {
		t.Fatalf("expected year 2026, got %v", ValueToAny(year))
	}
	if month := Call(NewKeyword("month"), got); month.Long() != 3 {
		t.Fatalf("expected month 3, got %v", ValueToAny(month))
	}
	if native, ok := ValueToAny(got).(time.Time); !ok || native.UTC().Year() != 2026 || native.UTC().Month() != 3 || native.UTC().Day() != 8 {
		t.Fatalf("expected native time.Time from ValueToAny, got %#v", ValueToAny(got))
	}
	if got := anyToValue(time.Date(2026, 3, 8, 12, 0, 0, 0, time.UTC)); got.tag != TagDate {
		t.Fatalf("expected anyToValue(time.Time) to produce date, got %v", got.tag)
	}
}
