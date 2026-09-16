package runtime

import (
	"math"
	"strings"
	"testing"
)

func requireDouble(t *testing.T, got Value, want float64) {
	t.Helper()
	if got.tag != TagDouble || got.Double() != want {
		t.Fatalf("expected double %v, got %#v (tag=%v d=%v)", want, got, got.tag, got.Double())
	}
}

func requireLong(t *testing.T, got Value, want int64) {
	t.Helper()
	if got.tag != TagLong || got.Long() != want {
		t.Fatalf("expected long %d, got %#v", want, got)
	}
}

func requireNaN(t *testing.T, got Value) {
	t.Helper()
	if got.tag != TagDouble || !math.IsNaN(got.Double()) {
		t.Fatalf("expected NaN double, got %#v", got)
	}
}

func requireInf(t *testing.T, got Value, sign int) {
	t.Helper()
	if got.tag != TagDouble || !math.IsInf(got.Double(), sign) {
		t.Fatalf("expected Inf(sign=%d) double, got %#v", sign, got)
	}
}

func TestMathAbsFunctionIsRegistered(t *testing.T) {
	if got := Call(GoFunction("math/abs"), NewLong(-42)); got.tag != TagLong || got.Long() != 42 {
		t.Fatalf("unexpected math/abs long result: %#v", got)
	}
	if got := Call(GoFunction("math/abs"), NewDouble(-2.5)); got.tag != TagDouble || got.Double() != 2.5 {
		t.Fatalf("unexpected math/abs double result: %#v", got)
	}
}

func TestSqrt(t *testing.T) {
	requireDouble(t, Sqrt(NewLong(4)), 2)
	requireDouble(t, Sqrt(NewDouble(9)), 3)
	requireDouble(t, Sqrt(NewRatio(9, 4)), 1.5)
	requireDouble(t, Sqrt(NewBigInt(16)), 4)
	requireDouble(t, Sqrt(NewLong(0)), 0)
	requireNaN(t, Sqrt(NewLong(-1)))
	requireNaN(t, Sqrt(NewDouble(math.NaN())))
	requireInf(t, Sqrt(NewDouble(math.Inf(1))), 1)
	assertPanics(t, func() { Sqrt(NewString("x")) })
}

func TestPow(t *testing.T) {
	requireDouble(t, Pow(NewLong(2), NewLong(3)), 8)
	requireDouble(t, Pow(NewLong(2), NewLong(0)), 1)
	requireDouble(t, Pow(NewLong(4), NewDouble(0.5)), 2)
	requireDouble(t, Pow(NewRatio(1, 2), NewLong(2)), 0.25)
	requireDouble(t, Pow(NewLong(2), NewLong(-1)), 0.5)
	requireNaN(t, Pow(NewLong(-1), NewDouble(0.5)))
	requireInf(t, Pow(NewLong(0), NewLong(-1)), 1)
	assertPanics(t, func() { Pow(NewString("x"), NewLong(1)) })
	assertPanics(t, func() { Pow(NewLong(1), NewString("x")) })
}

func TestExpLogLog10(t *testing.T) {
	requireDouble(t, Exp(NewLong(0)), 1)
	requireDouble(t, Exp(NewDouble(0)), 1)
	requireDouble(t, Log(NewLong(1)), 0)
	requireDouble(t, Log(NewRatio(1, 1)), 0)
	requireDouble(t, Log10(NewLong(1000)), 3)
	requireDouble(t, Log10(NewDouble(100)), 2)
	requireInf(t, Log(NewLong(0)), -1)
	requireInf(t, Log10(NewDouble(0)), -1)
	requireNaN(t, Log(NewLong(-1)))
	requireNaN(t, Log10(NewDouble(-8)))
	requireInf(t, Exp(NewDouble(math.Inf(1))), 1)
	requireDouble(t, Exp(NewDouble(math.Inf(-1))), 0)
	assertPanics(t, func() { Exp(NewBool(true)) })
	assertPanics(t, func() { Log(NilValue()) })
	assertPanics(t, func() { Log10(NewKeyword("x")) })
}

func TestSinCosTan(t *testing.T) {
	requireDouble(t, Sin(NewLong(0)), 0)
	requireDouble(t, Cos(NewLong(0)), 1)
	requireDouble(t, Tan(NewLong(0)), 0)
	requireDouble(t, Sin(NewRatio(0, 1)), 0)
	requireDouble(t, Cos(NewDouble(math.Pi)), -1)
	requireNaN(t, Sin(NewDouble(math.Inf(1))))
	requireNaN(t, Cos(NewDouble(math.Inf(-1))))
	requireNaN(t, Tan(NewDouble(math.NaN())))
	assertPanics(t, func() { Sin(NewString("x")) })
	assertPanics(t, func() { Cos(NewArray()) })
	assertPanics(t, func() { Tan(NewMap()) })
}

func TestFloorCeil(t *testing.T) {
	requireDouble(t, Floor(NewDouble(2.3)), 2)
	requireDouble(t, Floor(NewDouble(-2.3)), -3)
	requireDouble(t, Floor(NewLong(5)), 5)
	requireDouble(t, Floor(NewRatio(5, 2)), 2)
	requireDouble(t, Ceil(NewDouble(2.3)), 3)
	requireDouble(t, Ceil(NewDouble(-2.3)), -2)
	requireDouble(t, Ceil(NewLong(-5)), -5)
	requireDouble(t, Ceil(NewRatio(-5, 2)), -2)
	requireInf(t, Floor(NewDouble(math.Inf(1))), 1)
	requireInf(t, Ceil(NewDouble(math.Inf(-1))), -1)
	requireNaN(t, Floor(NewDouble(math.NaN())))
	requireNaN(t, Ceil(NewDouble(math.NaN())))
	assertPanics(t, func() { Floor(NewString("x")) })
	assertPanics(t, func() { Ceil(NewString("x")) })
}

func TestRound(t *testing.T) {
	requireLong(t, Round(NewDouble(1.4)), 1)
	requireLong(t, Round(NewDouble(1.5)), 2)
	requireLong(t, Round(NewDouble(2.5)), 3) // Clojure/Java: ties toward +∞
	requireLong(t, Round(NewDouble(-1.5)), -1)
	requireLong(t, Round(NewDouble(-2.5)), -2)
	requireLong(t, Round(NewLong(7)), 7)
	requireLong(t, Round(NewRatio(5, 2)), 3)
	requireLong(t, Round(NewDouble(math.NaN())), 0)
	requireLong(t, Round(NewDouble(math.Inf(1))), math.MaxInt64)
	requireLong(t, Round(NewDouble(math.Inf(-1))), math.MinInt64)
	assertPanics(t, func() { Round(NewString("x")) })
}

func TestIEEERemainder(t *testing.T) {
	requireDouble(t, IEEERemainder(NewDouble(5), NewDouble(3)), -1)
	requireDouble(t, IEEERemainder(NewLong(4), NewLong(3)), 1)
	requireDouble(t, IEEERemainder(NewRatio(5, 1), NewLong(3)), -1)
	requireDouble(t, IEEERemainder(NewLong(1), NewDouble(math.Inf(1))), 1)
	requireNaN(t, IEEERemainder(NewLong(1), NewLong(0)))
	requireNaN(t, IEEERemainder(NewDouble(math.Inf(1)), NewLong(2)))
	requireNaN(t, IEEERemainder(NewDouble(math.NaN()), NewLong(2)))
	assertPanics(t, func() { IEEERemainder(NewString("x"), NewLong(1)) })
	assertPanics(t, func() { IEEERemainder(NewLong(1), NewString("x")) })
}

func TestMathPackageFunctionsAreRegistered(t *testing.T) {
	cases := []struct {
		name  string
		args  []Value
		check func(t *testing.T, got Value)
	}{
		{"math/sqrt", []Value{NewLong(4)}, func(t *testing.T, got Value) { requireDouble(t, got, 2) }},
		{"math/pow", []Value{NewLong(2), NewLong(3)}, func(t *testing.T, got Value) { requireDouble(t, got, 8) }},
		{"math/exp", []Value{NewLong(0)}, func(t *testing.T, got Value) { requireDouble(t, got, 1) }},
		{"math/log", []Value{NewLong(1)}, func(t *testing.T, got Value) { requireDouble(t, got, 0) }},
		{"math/log10", []Value{NewLong(100)}, func(t *testing.T, got Value) { requireDouble(t, got, 2) }},
		{"math/sin", []Value{NewLong(0)}, func(t *testing.T, got Value) { requireDouble(t, got, 0) }},
		{"math/cos", []Value{NewLong(0)}, func(t *testing.T, got Value) { requireDouble(t, got, 1) }},
		{"math/tan", []Value{NewLong(0)}, func(t *testing.T, got Value) { requireDouble(t, got, 0) }},
		{"math/floor", []Value{NewDouble(-1.5)}, func(t *testing.T, got Value) { requireDouble(t, got, -2) }},
		{"math/ceil", []Value{NewDouble(-1.5)}, func(t *testing.T, got Value) { requireDouble(t, got, -1) }},
		{"math/round", []Value{NewDouble(2.5)}, func(t *testing.T, got Value) { requireLong(t, got, 3) }},
		{"math/IEEE-remainder", []Value{NewDouble(5), NewDouble(3)}, func(t *testing.T, got Value) { requireDouble(t, got, -1) }},
	}
	for _, tc := range cases {
		t.Run(tc.name+"/GoFunction", func(t *testing.T) {
			tc.check(t, Call(GoFunction(tc.name), tc.args...))
		})
	}

	bindChecks := []struct {
		name  string
		fn    Value
		args  []Value
		check func(t *testing.T, got Value)
	}{
		{"sqrt", GoBind_runtime_Sqrt, []Value{NewLong(9)}, func(t *testing.T, got Value) { requireDouble(t, got, 3) }},
		{"pow", GoBind_runtime_Pow, []Value{NewLong(2), NewLong(10)}, func(t *testing.T, got Value) { requireDouble(t, got, 1024) }},
		{"exp", GoBind_runtime_Exp, []Value{NewLong(0)}, func(t *testing.T, got Value) { requireDouble(t, got, 1) }},
		{"log", GoBind_runtime_Log, []Value{NewLong(1)}, func(t *testing.T, got Value) { requireDouble(t, got, 0) }},
		{"log10", GoBind_runtime_Log10, []Value{NewLong(10)}, func(t *testing.T, got Value) { requireDouble(t, got, 1) }},
		{"sin", GoBind_runtime_Sin, []Value{NewLong(0)}, func(t *testing.T, got Value) { requireDouble(t, got, 0) }},
		{"cos", GoBind_runtime_Cos, []Value{NewLong(0)}, func(t *testing.T, got Value) { requireDouble(t, got, 1) }},
		{"tan", GoBind_runtime_Tan, []Value{NewLong(0)}, func(t *testing.T, got Value) { requireDouble(t, got, 0) }},
		{"floor", GoBind_runtime_Floor, []Value{NewDouble(1.1)}, func(t *testing.T, got Value) { requireDouble(t, got, 1) }},
		{"ceil", GoBind_runtime_Ceil, []Value{NewDouble(1.1)}, func(t *testing.T, got Value) { requireDouble(t, got, 2) }},
		{"round", GoBind_runtime_Round, []Value{NewDouble(-1.5)}, func(t *testing.T, got Value) { requireLong(t, got, -1) }},
		{"IEEE-remainder", GoBind_runtime_IEEERemainder, []Value{NewLong(4), NewLong(3)}, func(t *testing.T, got Value) { requireDouble(t, got, 1) }},
	}
	for _, tc := range bindChecks {
		t.Run(tc.name+"/GoBind", func(t *testing.T) {
			tc.check(t, Call(tc.fn, tc.args...))
		})
	}
}

func TestMathPackageArity(t *testing.T) {
	assertPanicContains(t, "exactly 1", func() { Call(GoBind_runtime_Sqrt) })
	assertPanicContains(t, "exactly 2", func() { Call(GoBind_runtime_Pow, NewLong(2)) })
	assertPanicContains(t, "exactly 2", func() { Call(GoBind_runtime_IEEERemainder, NewLong(1)) })
}

func assertPanicContains(t *testing.T, want string, fn func()) {
	t.Helper()
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic")
		}
		if !strings.Contains(toPanicString(r), want) {
			t.Fatalf("expected panic containing %q, got %#v", want, r)
		}
	}()
	fn()
}

func toPanicString(r any) string {
	switch v := r.(type) {
	case string:
		return v
	case interface{ Error() string }:
		return v.Error()
	default:
		return ""
	}
}
