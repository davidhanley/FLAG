package runtime

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"unsafe"
)

type Value struct {
	d   float64
	p   unsafe.Pointer
	tag ValueTag
}

func (v Value) String() string {
	return ValueToString(v)
}

func NilValue() Value {
	return Value{tag: TagNil}
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

func NewBool(v bool) Value {
	if v {
		return Value{d: 1, tag: TagBool}
	}
	return Value{tag: TagBool}
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

func (v Value) Bool() bool {
	if v.tag != TagBool {
		panic("Bool called on non-bool Value")
	}
	return v.d != 0
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

func Mod(lhs, rhs Value) Value {
	if lhs.tag != TagLong || rhs.tag != TagLong {
		panic("mod expects integer Value arguments")
	}
	left := lhs.Long()
	right := rhs.Long()
	if right == 0 {
		panic("mod by zero")
	}

	remainder := left % right
	if remainder != 0 && (remainder < 0) != (right < 0) {
		remainder += right
	}
	return NewLong(remainder)
}

func Eq(lhs, rhs Value) bool {
	if isNumericTag(lhs.tag) && isNumericTag(rhs.tag) {
		return compareNumeric(lhs, rhs) == 0
	}
	if lhs.tag != rhs.tag {
		return false
	}
	switch lhs.tag {
	case TagNil:
		return true
	default:
		return valueIdentity(lhs) == valueIdentity(rhs)
	}
}

func isNumericTag(tag ValueTag) bool {
	switch tag {
	case TagLong, TagDouble, TagRatio:
		return true
	default:
		return false
	}
}

func Lt(lhs, rhs Value) bool {
	return compareNumeric(lhs, rhs) < 0
}

func Gt(lhs, rhs Value) bool {
	return compareNumeric(lhs, rhs) > 0
}

func IsTruthy(v Value) bool {
	switch v.tag {
	case TagNil:
		return false
	case TagBool:
		return v.Bool()
	default:
		return true
	}
}

func Str(args ...any) string {
	var out strings.Builder
	for _, arg := range args {
		out.WriteString(anyToString(arg))
	}
	return out.String()
}

func Println(args ...any) Value {
	fmt.Println(Str(args...))
	return NilValue()
}

func ValueToString(v Value) string {
	switch v.tag {
	case TagLong:
		return strconv.FormatInt(v.Long(), 10)
	case TagDouble:
		return strconv.FormatFloat(v.Double(), 'g', -1, 64)
	case TagRatio:
		return v.Ratio().RatString()
	case TagBool:
		return strconv.FormatBool(v.Bool())
	case TagSymbol:
		symbol := v.SymbolObject()
		if symbol.IsKeyword {
			return ":" + symbol.Name
		}
		return symbol.Name
	case TagFunction:
		return "#<fn>"
	case TagMap:
		entries := v.MapEntries()
		var out strings.Builder
		out.WriteByte('{')
		for i, entry := range entries {
			if i > 0 {
				out.WriteByte(' ')
			}
			out.WriteString(ValueToString(entry.Key))
			out.WriteByte(' ')
			out.WriteString(ValueToString(entry.Value))
		}
		out.WriteByte('}')
		return out.String()
	case TagSet:
		values := v.SetValues()
		var out strings.Builder
		out.WriteString("#{")
		for i, value := range values {
			if i > 0 {
				out.WriteByte(' ')
			}
			out.WriteString(ValueToString(value))
		}
		out.WriteByte('}')
		return out.String()
	case TagNil:
		return ""
	case TagList:
		values := v.ListValues()
		var out strings.Builder
		out.WriteByte('(')
		for i, value := range values {
			if i > 0 {
				out.WriteByte(' ')
			}
			out.WriteString(ValueToString(value))
		}
		out.WriteByte(')')
		return out.String()
	case TagArray:
		values := v.ArrayValues()
		var out strings.Builder
		out.WriteByte('[')
		for i, value := range values {
			if i > 0 {
				out.WriteByte(' ')
			}
			out.WriteString(ValueToString(value))
		}
		out.WriteByte(']')
		return out.String()
	default:
		panic("unknown Value tag")
	}
}

func anyToString(arg any) string {
	switch value := arg.(type) {
	case Value:
		return ValueToString(value)
	case string:
		return value
	case *SymbolObject:
		if value == nil {
			return ""
		}
		if value.IsKeyword {
			return ":" + value.Name
		}
		return value.Name
	case bool:
		return strconv.FormatBool(value)
	case nil:
		return ""
	default:
		return fmt.Sprint(value)
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
	case TagBool:
		return v.Bool()
	case TagSymbol, TagFunction, TagMap, TagSet:
		return v
	case TagNil:
		return nil
	case TagList:
		return listValueToAny(v)
	case TagArray:
		return arrayValueToAny(v)
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

func compareNumeric(lhs, rhs Value) int {
	switch lhs.tag {
	case TagLong:
		left := lhs.Long()
		switch rhs.tag {
		case TagLong:
			return compareFloat64(float64(left), float64(rhs.Long()))
		case TagDouble:
			return compareFloat64(float64(left), rhs.Double())
		case TagRatio:
			return big.NewRat(left, 1).Cmp(rhs.Ratio())
		default:
			panic("unknown rhs tag for numeric compare")
		}
	case TagDouble:
		left := lhs.Double()
		switch rhs.tag {
		case TagLong:
			return compareFloat64(left, float64(rhs.Long()))
		case TagDouble:
			return compareFloat64(left, rhs.Double())
		case TagRatio:
			return compareFloat64(left, ratToFloat64(rhs.Ratio()))
		default:
			panic("unknown rhs tag for numeric compare")
		}
	case TagRatio:
		left := lhs.Ratio()
		switch rhs.tag {
		case TagLong:
			return left.Cmp(big.NewRat(rhs.Long(), 1))
		case TagDouble:
			return compareFloat64(ratToFloat64(left), rhs.Double())
		case TagRatio:
			return left.Cmp(rhs.Ratio())
		default:
			panic("unknown rhs tag for numeric compare")
		}
	default:
		panic("unknown lhs tag for numeric compare")
	}
}

func compareFloat64(lhs, rhs float64) int {
	switch {
	case lhs < rhs:
		return -1
	case lhs > rhs:
		return 1
	default:
		return 0
	}
}
