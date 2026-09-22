package runtime

import "math"

func mathFloat1(name string, value Value, fn func(float64) float64) Value {
	if !isNumericTag(value.tag) {
		panic(name + " expects numeric Value argument")
	}
	return NewDouble(fn(numericToFloat64(value)))
}

func mathFloat2(name string, a, b Value, fn func(float64, float64) float64) Value {
	if !isNumericTag(a.tag) || !isNumericTag(b.tag) {
		panic(name + " expects numeric Value arguments")
	}
	return NewDouble(fn(numericToFloat64(a), numericToFloat64(b)))
}

func Sqrt(value Value) Value {
	return mathFloat1("sqrt", value, math.Sqrt)
}

func Pow(base, exp Value) Value {
	return mathFloat2("pow", base, exp, math.Pow)
}

func Exp(value Value) Value {
	return mathFloat1("exp", value, math.Exp)
}

func Log(value Value) Value {
	return mathFloat1("log", value, math.Log)
}

func Log10(value Value) Value {
	return mathFloat1("log10", value, math.Log10)
}

func Sin(value Value) Value {
	return mathFloat1("sin", value, math.Sin)
}

func Cos(value Value) Value {
	return mathFloat1("cos", value, math.Cos)
}

func Tan(value Value) Value {
	return mathFloat1("tan", value, math.Tan)
}

func Floor(value Value) Value {
	return mathFloat1("floor", value, math.Floor)
}

func Ceil(value Value) Value {
	return mathFloat1("ceil", value, math.Ceil)
}

// Round matches clojure.math/round / java.lang.Math.round: nearest long,
// with ties rounding toward +∞. NaN becomes 0; infinities saturate.
func Round(value Value) Value {
	if !isNumericTag(value.tag) {
		panic("round expects numeric Value argument")
	}
	return NewLong(javaMathRound(numericToFloat64(value)))
}

func IEEERemainder(dividend, divisor Value) Value {
	return mathFloat2("IEEE-remainder", dividend, divisor, math.Remainder)
}

func javaMathRound(x float64) int64 {
	switch {
	case math.IsNaN(x):
		return 0
	case math.IsInf(x, 1) || x >= float64(math.MaxInt64):
		return math.MaxInt64
	case math.IsInf(x, -1) || x <= float64(math.MinInt64):
		return math.MinInt64
	default:
		return int64(math.Floor(x + 0.5))
	}
}
