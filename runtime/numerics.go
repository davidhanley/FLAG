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

func NewBigInt(v int64) Value {
	return NewBigIntFromBigInt(big.NewInt(v))
}

func NewBigIntFromBigInt(v *big.Int) Value {
	if v == nil {
		panic("NewBigIntFromBigInt expects non-nil big.Int")
	}
	copied := new(big.Int).Set(v)
	return Value{p: unsafe.Pointer(copied), tag: TagBigInt}
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

func (v Value) BigInt() *big.Int {
	if v.tag != TagBigInt {
		panic("BigInt called on non-bigint Value")
	}
	return v.bigIntPointer()
}

func (v Value) Bool() bool {
	if v.tag != TagBool {
		panic("Bool called on non-bool Value")
	}
	return v.d != 0
}

func Add(lhs, rhs Value) Value {
	ensureNumericOperands("Add", lhs, rhs)
	if lhs.tag == TagDouble || rhs.tag == TagDouble {
		return NewDouble(numericToFloat64(lhs) + numericToFloat64(rhs))
	}
	if lhs.tag == TagRatio || rhs.tag == TagRatio {
		left := valueToRat(lhs)
		left.Add(left, valueToRat(rhs))
		return NewRatioFromRat(left)
	}
	return addIntegerValues(lhs, rhs)
}

func Mul(lhs, rhs Value) Value {
	ensureNumericOperands("Mul", lhs, rhs)
	if lhs.tag == TagDouble || rhs.tag == TagDouble {
		return NewDouble(numericToFloat64(lhs) * numericToFloat64(rhs))
	}
	if lhs.tag == TagRatio || rhs.tag == TagRatio {
		left := valueToRat(lhs)
		left.Mul(left, valueToRat(rhs))
		return NewRatioFromRat(left)
	}
	return mulIntegerValues(lhs, rhs)
}

func Sub(lhs, rhs Value) Value {
	ensureNumericOperands("Sub", lhs, rhs)
	if lhs.tag == TagDouble || rhs.tag == TagDouble {
		return NewDouble(numericToFloat64(lhs) - numericToFloat64(rhs))
	}
	if lhs.tag == TagRatio || rhs.tag == TagRatio {
		left := valueToRat(lhs)
		left.Sub(left, valueToRat(rhs))
		return NewRatioFromRat(left)
	}
	return subIntegerValues(lhs, rhs)
}

func Div(lhs, rhs Value) Value {
	ensureNumericOperands("Div", lhs, rhs)
	if lhs.tag == TagDouble || rhs.tag == TagDouble {
		return NewDouble(numericToFloat64(lhs) / numericToFloat64(rhs))
	}
	left := valueToRat(lhs)
	left.Quo(left, valueToRat(rhs))
	return NewRatioFromRat(left)
}

func Mod(lhs, rhs Value) Value {
	if !isIntegerTag(lhs.tag) || !isIntegerTag(rhs.tag) {
		panic("mod expects integer Value arguments")
	}
	left := valueToBigInt(lhs)
	right := valueToBigInt(rhs)
	if right.Sign() == 0 {
		panic("mod by zero")
	}

	remainder := new(big.Int).Rem(left, right)
	if remainder.Sign() != 0 && ((remainder.Sign() < 0) != (right.Sign() < 0)) {
		remainder.Add(remainder, right)
	}
	return newIntegerValueFromBigInt(remainder)
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
	case TagLong, TagDouble, TagRatio, TagBigInt:
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
	case TagBigInt:
		return v.BigInt().String()
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
	case TagLazyList:
		return "#<lazy-list>"
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
	case TagBigInt:
		return v.BigInt()
	case TagBool:
		return v.Bool()
	case TagSymbol, TagFunction, TagMap, TagSet, TagLazyList:
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
	ensureNumericOperands("compare", lhs, rhs)
	if lhs.tag == TagDouble || rhs.tag == TagDouble {
		return compareFloat64(numericToFloat64(lhs), numericToFloat64(rhs))
	}
	return valueToRat(lhs).Cmp(valueToRat(rhs))
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

func (v Value) bigIntPointer() *big.Int {
	if v.p == nil {
		panic("bigint Value does not contain big.Int pointer")
	}
	return (*big.Int)(v.p)
}

func isIntegerTag(tag ValueTag) bool {
	return tag == TagLong || tag == TagBigInt
}

func ensureNumericOperands(op string, lhs, rhs Value) {
	if !isNumericTag(lhs.tag) || !isNumericTag(rhs.tag) {
		panic(op + " expects numeric Value arguments")
	}
}

func numericToFloat64(v Value) float64 {
	switch v.tag {
	case TagLong:
		return float64(v.Long())
	case TagDouble:
		return v.Double()
	case TagRatio:
		return ratToFloat64(v.Ratio())
	case TagBigInt:
		f, _ := new(big.Float).SetInt(v.BigInt()).Float64()
		return f
	default:
		panic("numericToFloat64 called on non-numeric Value")
	}
}

func valueToRat(v Value) *big.Rat {
	switch v.tag {
	case TagLong:
		return big.NewRat(v.Long(), 1)
	case TagBigInt:
		return new(big.Rat).SetInt(v.BigInt())
	case TagRatio:
		return new(big.Rat).Set(v.Ratio())
	default:
		panic("valueToRat expects integer or ratio Value")
	}
}

func valueToBigInt(v Value) *big.Int {
	switch v.tag {
	case TagLong:
		return big.NewInt(v.Long())
	case TagBigInt:
		return new(big.Int).Set(v.BigInt())
	default:
		panic("valueToBigInt expects integer Value")
	}
}

func newIntegerValueFromBigInt(v *big.Int) Value {
	if v.IsInt64() {
		return NewLong(v.Int64())
	}
	return NewBigIntFromBigInt(v)
}

func addIntegerValues(lhs, rhs Value) Value {
	if lhs.tag == TagLong && rhs.tag == TagLong {
		left := lhs.Long()
		right := rhs.Long()
		sum := left + right
		if (right > 0 && sum < left) || (right < 0 && sum > left) {
			bigSum := new(big.Int).Add(big.NewInt(left), big.NewInt(right))
			return NewBigIntFromBigInt(bigSum)
		}
		return NewLong(sum)
	}
	sum := new(big.Int).Add(valueToBigInt(lhs), valueToBigInt(rhs))
	return newIntegerValueFromBigInt(sum)
}

func subIntegerValues(lhs, rhs Value) Value {
	if lhs.tag == TagLong && rhs.tag == TagLong {
		left := lhs.Long()
		right := rhs.Long()
		diff := left - right
		if (right < 0 && diff < left) || (right > 0 && diff > left) {
			bigDiff := new(big.Int).Sub(big.NewInt(left), big.NewInt(right))
			return NewBigIntFromBigInt(bigDiff)
		}
		return NewLong(diff)
	}
	diff := new(big.Int).Sub(valueToBigInt(lhs), valueToBigInt(rhs))
	return newIntegerValueFromBigInt(diff)
}

func mulIntegerValues(lhs, rhs Value) Value {
	if lhs.tag == TagLong && rhs.tag == TagLong {
		left := lhs.Long()
		right := rhs.Long()
		if left == 0 || right == 0 {
			return NewLong(0)
		}
		if (left == minInt64 && right == -1) || (left == -1 && right == minInt64) {
			bigProd := new(big.Int).Mul(big.NewInt(left), big.NewInt(right))
			return NewBigIntFromBigInt(bigProd)
		}
		product := left * right
		if product/right != left {
			bigProd := new(big.Int).Mul(big.NewInt(left), big.NewInt(right))
			return NewBigIntFromBigInt(bigProd)
		}
		return NewLong(product)
	}
	product := new(big.Int).Mul(valueToBigInt(lhs), valueToBigInt(rhs))
	return newIntegerValueFromBigInt(product)
}

const (
	maxInt64 = int64(^uint64(0) >> 1)
	minInt64 = -maxInt64 - 1
)
