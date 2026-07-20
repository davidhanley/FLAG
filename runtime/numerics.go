package runtime

import (
	"math/big"
	"unsafe"
)

type ValueTag uint8

const (
	TagLong ValueTag = iota + 1
	TagDouble
	TagRatio
	TagList
)

type Value struct {
	d   float64
	p   unsafe.Pointer
	tag ValueTag
}

func NewLong(v int64) Value {
	out := Value{tag: TagLong}
	*(*int64)(unsafe.Pointer(&out.d)) = v
	return out
}

func NewDouble(v float64) Value {
	return Value{d: v, tag: TagDouble}
}

func NewRatio(numerator, denominator int64) Value {
	rat := big.NewRat(numerator, denominator)
	return Value{p: unsafe.Pointer(rat), tag: TagRatio}
}

func NewRatioFromRat(rat *big.Rat) Value {
	return Value{p: unsafe.Pointer(rat), tag: TagRatio}
}

func (v Value) Long() int64 {
	return *(*int64)(unsafe.Pointer(&v.d))
}

func (v Value) Double() float64 {
	return v.d
}

func (v Value) Ratio() *big.Rat {
	if v.tag != TagRatio {
		panic("Ratio called on non-ratio Value")
	}
	return v.ratioPointer()
}

func Add(lhs, rhs Value) Value {
	switch lhs.tag {
	case TagLong:
		left := lhs.Long()
		switch rhs.tag {
		case TagLong:
			return NewLong(left + rhs.Long())
		case TagDouble:
			return NewDouble(float64(left) + rhs.Double())
		default:
			panic("unknown rhs tag for Add")
		}
	case TagDouble:
		left := lhs.Double()
		switch rhs.tag {
		case TagLong:
			return NewDouble(left + float64(rhs.Long()))
		case TagDouble:
			return NewDouble(left + rhs.Double())
		default:
			panic("unknown rhs tag for Add")
		}
	case TagRatio:
		left := new(big.Rat).Set(lhs.Ratio())
		switch rhs.tag {
		case TagLong:
			left.Add(left, big.NewRat(rhs.Long(), 1))
			return NewRatioFromRat(left)
		case TagDouble:
			return NewDouble(ratToFloat64(left) + rhs.Double())
		case TagRatio:
			left.Add(left, rhs.Ratio())
			return NewRatioFromRat(left)
		default:
			panic("unknown rhs tag for Add")
		}
	default:
		panic("unknown lhs tag for Add")
	}
}

func Mul(lhs, rhs Value) Value {
	switch lhs.tag {
	case TagLong:
		left := lhs.Long()
		switch rhs.tag {
		case TagLong:
			return NewLong(left * rhs.Long())
		case TagDouble:
			return NewDouble(float64(left) * rhs.Double())
		default:
			panic("unknown rhs tag for Mul")
		}
	case TagDouble:
		left := lhs.Double()
		switch rhs.tag {
		case TagLong:
			return NewDouble(left * float64(rhs.Long()))
		case TagDouble:
			return NewDouble(left * rhs.Double())
		default:
			panic("unknown rhs tag for Mul")
		}
	case TagRatio:
		left := new(big.Rat).Set(lhs.Ratio())
		switch rhs.tag {
		case TagLong:
			left.Mul(left, big.NewRat(rhs.Long(), 1))
			return NewRatioFromRat(left)
		case TagDouble:
			return NewDouble(ratToFloat64(left) * rhs.Double())
		case TagRatio:
			left.Mul(left, rhs.Ratio())
			return NewRatioFromRat(left)
		default:
			panic("unknown rhs tag for Mul")
		}
	default:
		panic("unknown lhs tag for Mul")
	}
}

func Sub(lhs, rhs Value) Value {
	switch lhs.tag {
	case TagLong:
		left := lhs.Long()
		switch rhs.tag {
		case TagLong:
			return NewLong(left - rhs.Long())
		case TagDouble:
			return NewDouble(float64(left) - rhs.Double())
		default:
			panic("unknown rhs tag for Sub")
		}
	case TagDouble:
		left := lhs.Double()
		switch rhs.tag {
		case TagLong:
			return NewDouble(left - float64(rhs.Long()))
		case TagDouble:
			return NewDouble(left - rhs.Double())
		default:
			panic("unknown rhs tag for Sub")
		}
	case TagRatio:
		left := new(big.Rat).Set(lhs.Ratio())
		switch rhs.tag {
		case TagLong:
			left.Sub(left, big.NewRat(rhs.Long(), 1))
			return NewRatioFromRat(left)
		case TagDouble:
			return NewDouble(ratToFloat64(left) - rhs.Double())
		case TagRatio:
			left.Sub(left, rhs.Ratio())
			return NewRatioFromRat(left)
		default:
			panic("unknown rhs tag for Sub")
		}
	default:
		panic("unknown lhs tag for Sub")
	}
}

func Div(lhs, rhs Value) Value {
	switch lhs.tag {
	case TagLong:
		left := lhs.Long()
		switch rhs.tag {
		case TagLong:
			return NewRatio(left, rhs.Long())
		case TagDouble:
			return NewDouble(float64(left) / rhs.Double())
		case TagRatio:
			return NewRatioFromRat(new(big.Rat).Quo(big.NewRat(left, 1), rhs.Ratio()))
		default:
			panic("unknown rhs tag for Div")
		}
	case TagDouble:
		left := lhs.Double()
		switch rhs.tag {
		case TagLong:
			return NewDouble(left / float64(rhs.Long()))
		case TagDouble:
			return NewDouble(left / rhs.Double())
		case TagRatio:
			return NewDouble(left / ratToFloat64(rhs.Ratio()))
		default:
			panic("unknown rhs tag for Div")
		}
	case TagRatio:
		left := new(big.Rat).Set(lhs.Ratio())
		switch rhs.tag {
		case TagLong:
			return NewRatioFromRat(left.Quo(left, big.NewRat(rhs.Long(), 1)))
		case TagDouble:
			return NewDouble(ratToFloat64(left) / rhs.Double())
		case TagRatio:
			return NewRatioFromRat(left.Quo(left, rhs.Ratio()))
		default:
			panic("unknown rhs tag for Div")
		}
	default:
		panic("unknown lhs tag for Div")
	}
}

func ValueToAny(v Value) any {
	switch v.tag {
	case TagLong:
		return v.Long()
	case TagDouble:
		return v.Double()
	case TagRatio:
		return v.Ratio()
	case TagList:
		return listValueToAny(v)
	default:
		panic("unknown Value tag")
	}
}

func (v Value) ratioPointer() *big.Rat {
	if v.p == nil {
		panic("ratio Value does not contain rat pointer")
	}
	return (*big.Rat)(v.p)
}

func ratToFloat64(r *big.Rat) float64 {
	f, _ := r.Float64()
	return f
}
