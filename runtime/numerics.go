package runtime

import (
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

// Immediate types are a single iface value so NewLong/NewDouble/NewBool do not
// allocate. All longs share longType; equality still uses packed d.
var (
	longType   impl = longImpl{}
	doubleType impl = doubleImpl{}
	boolType   impl = boolImpl{}
	TrueValue       = Value{d: 1, impl: boolType}
	FalseValue      = Value{d: 0, impl: boolType}
)

func NewLong(v int64) Value {
	return Value{d: packLong(v), impl: longType}
}

func NewDouble(v float64) Value {
	return Value{d: v, impl: doubleType}
}

func Double(v Value) Value {
	return v.impl.toDouble(v)
}

func Numerator(v Value) Value {
	return v.impl.numerator(v)
}

func Denominator(v Value) Value {
	return v.impl.denominator(v)
}

func NewRatio(numerator, denominator int64) Value {
	rat := big.NewRat(numerator, denominator)
	return NewRatioFromRat(rat)
}

func NewRatioFromRat(rat *big.Rat) Value {
	return Value{impl: &ratioImpl{rat: rat}}
}

func NewBigInt(v int64) Value {
	return NewBigIntFromBigInt(big.NewInt(v))
}

func NewBigIntFromString(raw string) Value {
	parsed, ok := new(big.Int).SetString(raw, 10)
	if !ok {
		panic("NewBigIntFromString expects a valid base-10 integer")
	}
	return NewBigIntFromBigInt(parsed)
}

func NewBigIntFromBigInt(v *big.Int) Value {
	if v == nil {
		panic("NewBigIntFromBigInt expects non-nil big.Int")
	}
	copied := new(big.Int).Set(v)
	return Value{impl: &bigIntImpl{n: copied}}
}

func NewBool(v bool) Value {
	if v {
		return TrueValue
	}
	return FalseValue
}

func (v Value) Long() int64 {
	return *(*int64)(unsafe.Pointer(&v.d))
}

func (v Value) Double() float64 {
	return v.d
}

func (v Value) Ratio() *big.Rat {
	if !isType[*ratioImpl](v) {
		panic("Ratio called on non-ratio Value")
	}
	return v.ratioPointer()
}

func (v Value) BigInt() *big.Int {
	if !isType[*bigIntImpl](v) {
		panic("BigInt called on non-bigint Value")
	}
	return v.bigIntPointer()
}

func (v Value) Bool() bool {
	if !isType[boolImpl](v) {
		panic("Bool called on non-bool Value")
	}
	return v.d != 0
}

func Add(lhs, rhs Value) Value {
	if _, ok := lhs.impl.(longImpl); ok {
		if _, ok := rhs.impl.(longImpl); ok {
			return addLongs(lhs.Long(), rhs.Long())
		}
		return rhs.impl.addLong(rhs, lhs.Long())
	}
	return lhs.impl.add(lhs, rhs)
}

func Inc(v Value) Value {
	if _, ok := v.impl.(longImpl); ok {
		return addLongs(v.Long(), 1)
	}
	return v.impl.addLong(v, 1)
}

func Dec(v Value) Value {
	if _, ok := v.impl.(longImpl); ok {
		return addLongs(v.Long(), -1)
	}
	return v.impl.addLong(v, -1)
}

func Mul(lhs, rhs Value) Value {
	return lhs.impl.mul(lhs, rhs)
}

func Sub(lhs, rhs Value) Value {
	return lhs.impl.sub(lhs, rhs)
}

func Div(lhs, rhs Value) Value {
	return lhs.impl.div(lhs, rhs)
}

func Abs(value Value) Value {
	return value.impl.abs(value)
}

func Max(values ...Value) Value {
	if len(values) == 0 {
		panic("max expects at least one argument")
	}
	best := values[0]
	if !best.impl.isNumeric(best) {
		panic("max expects numeric Value arguments")
	}
	for _, value := range values[1:] {
		if !value.impl.isNumeric(value) {
			panic("max expects numeric Value arguments")
		}
		if compareNumeric(best, value) <= 0 {
			best = value
		}
	}
	return best
}

func Min(values ...Value) Value {
	if len(values) == 0 {
		panic("min expects at least one argument")
	}
	best := values[0]
	if !best.impl.isNumeric(best) {
		panic("min expects numeric Value arguments")
	}
	for _, value := range values[1:] {
		if !value.impl.isNumeric(value) {
			panic("min expects numeric Value arguments")
		}
		if compareNumeric(best, value) >= 0 {
			best = value
		}
	}
	return best
}

func Quot(lhs, rhs Value) Value {
	return lhs.impl.quot(lhs, rhs)
}

func Rem(lhs, rhs Value) Value {
	ensureNumericOperands("rem", lhs, rhs)
	panicIfExactZeroDivisor("rem", lhs, rhs)
	return Sub(lhs, Mul(Quot(lhs, rhs), rhs))
}

func Mod(lhs, rhs Value) Value {
	ensureNumericOperands("mod", lhs, rhs)
	panicIfExactZeroDivisor("mod", lhs, rhs)
	m := Rem(lhs, rhs)
	if isNumericZero(m) {
		return m
	}
	if (compareNumeric(m, NewLong(0)) < 0) == (compareNumeric(rhs, NewLong(0)) < 0) {
		return m
	}
	return Add(m, rhs)
}

func Eq(lhs, rhs Value) bool {
	if lhs.impl.isNumeric(lhs) && rhs.impl.isNumeric(rhs) {
		return compareNumeric(lhs, rhs) == 0
	}
	return lhs.impl.eq(lhs, rhs)
}

func isNumeric(v Value) bool {
	return v.impl.isNumeric(v)
}

func Lt(lhs, rhs Value) bool {
	return compareOrdered(lhs, rhs) < 0
}

func Gt(lhs, rhs Value) bool {
	return compareOrdered(lhs, rhs) > 0
}

func Le(lhs, rhs Value) bool {
	return compareOrdered(lhs, rhs) <= 0
}

func Ge(lhs, rhs Value) bool {
	return compareOrdered(lhs, rhs) >= 0
}

func Compare(lhs, rhs Value) Value {
	return lhs.impl.compare(lhs, rhs)
}

func NumericEq(values ...Value) bool {
	if len(values) <= 1 {
		for _, value := range values {
			if !value.impl.isNumeric(value) {
				panic("== expects numeric Value arguments")
			}
		}
		return true
	}
	for _, value := range values {
		if !value.impl.isNumeric(value) {
			panic("== expects numeric Value arguments")
		}
	}
	for i := 1; i < len(values); i++ {
		if compareNumeric(values[i-1], values[i]) != 0 {
			return false
		}
	}
	return true
}

func compareOrdered(lhs, rhs Value) int {
	if _, ok := lhs.impl.(longImpl); ok {
		if _, ok := rhs.impl.(longImpl); ok {
			a, b := lhs.Long(), rhs.Long()
			if a < b {
				return -1
			}
			if a > b {
				return 1
			}
			return 0
		}
	}
	if lhs.impl.isNumeric(lhs) && rhs.impl.isNumeric(rhs) {
		return compareNumeric(lhs, rhs)
	}
	if TypeOf(lhs) == typeString && TypeOf(rhs) == typeString {
		return strings.Compare(lhs.StringValue(), rhs.StringValue())
	}
	panic("comparison expects numeric or string Value arguments")
}

func IsTruthy(v Value) bool {
	return v.impl.isTruthy(v)
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
	return NilValue
}

func Print(args ...any) Value {
	fmt.Print(printJoin(args...))
	return NilValue
}

func printJoin(args ...any) string {
	if len(args) == 0 {
		return ""
	}
	var out strings.Builder
	for i, arg := range args {
		if i > 0 {
			out.WriteByte(' ')
		}
		out.WriteString(printArgString(arg))
	}
	return out.String()
}

func printArgString(arg any) string {
	switch value := arg.(type) {
	case Value:
		if IsNil(value) {
			return "nil"
		}
		if isType[*stringImpl](value) {
			return value.StringValue()
		}
		return value.String()
	case string:
		return value
	case nil:
		return "nil"
	default:
		return anyToString(arg)
	}
}

func anyToString(arg any) string {
	switch value := arg.(type) {
	case Value:
		if isType[*stringImpl](value) {
			return value.StringValue()
		}
		return value.String()
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
	case time.Time:
		return dateString(NewDate(value))
	default:
		return fmt.Sprint(value)
	}
}

func ValueToAny(v Value) any {
	return v.impl.toAny(v)
}

func (v Value) ratioPointer() *big.Rat {
	r := asImpl[*ratioImpl](v, "ratio Value does not contain rat pointer").rat
	if r == nil {
		panic("ratio Value does not contain rat pointer")
	}
	return r
}

func ratToFloat64(r *big.Rat) float64 {
	f, _ := r.Float64()
	return f
}

func compareNumeric(lhs, rhs Value) int {
	return lhs.impl.cmpNum(lhs, rhs)
}

func compareInt64(lhs, rhs int64) int {
	switch {
	case lhs < rhs:
		return -1
	case lhs > rhs:
		return 1
	default:
		return 0
	}
}

func compareLongAndBigInt(lhs int64, rhs *big.Int) int {
	if rhs.IsInt64() {
		return compareInt64(lhs, rhs.Int64())
	}
	if rhs.Sign() > 0 {
		return -1
	}
	return 1
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
	n := asImpl[*bigIntImpl](v, "bigint Value does not contain big.Int pointer").n
	if n == nil {
		panic("bigint Value does not contain big.Int pointer")
	}
	return n
}

func ensureNumericOperands(op string, lhs, rhs Value) {
	if !lhs.impl.isNumeric(lhs) || !rhs.impl.isNumeric(rhs) {
		panic(op + " expects numeric Value arguments")
	}
}

func isNumericZero(v Value) bool {
	return v.impl.cmpLong(v, 0) == 0
}

func panicIfExactZeroDivisor(op string, lhs, rhs Value) {
	if TypeOf(lhs) == typeFloat || TypeOf(rhs) == typeFloat {
		return
	}
	if isNumericZero(rhs) {
		panic(op + " by zero")
	}
}

func numericToFloat64(v Value) float64 {
	return v.impl.toFloat64(v)
}

func valueToRat(v Value) *big.Rat {
	return v.impl.toRat(v)
}

func valueToBigInt(v Value) *big.Int {
	return v.impl.toBigInt(v)
}

func newIntegerValueFromBigInt(v *big.Int) Value {
	if v.IsInt64() {
		return NewLong(v.Int64())
	}
	return NewBigIntFromBigInt(v)
}

func addLongs(left, right int64) Value {
	sum := left + right
	if (right > 0 && sum < left) || (right < 0 && sum > left) {
		return NewBigIntFromBigInt(new(big.Int).Add(big.NewInt(left), big.NewInt(right)))
	}
	return NewLong(sum)
}

func subLongs(left, right int64) Value {
	diff := left - right
	if (right < 0 && diff < left) || (right > 0 && diff > left) {
		return NewBigIntFromBigInt(new(big.Int).Sub(big.NewInt(left), big.NewInt(right)))
	}
	return NewLong(diff)
}

func mulLongs(left, right int64) Value {
	if left == 0 || right == 0 {
		return NewLong(0)
	}
	if (left == minInt64 && right == -1) || (left == -1 && right == minInt64) {
		return NewBigIntFromBigInt(new(big.Int).Mul(big.NewInt(left), big.NewInt(right)))
	}
	product := left * right
	if product/right != left {
		return NewBigIntFromBigInt(new(big.Int).Mul(big.NewInt(left), big.NewInt(right)))
	}
	return NewLong(product)
}

func addInts(left, right *big.Int) Value {
	return newIntegerValueFromBigInt(new(big.Int).Add(left, right))
}

func subInts(left, right *big.Int) Value {
	return newIntegerValueFromBigInt(new(big.Int).Sub(left, right))
}

func mulInts(left, right *big.Int) Value {
	return newIntegerValueFromBigInt(new(big.Int).Mul(left, right))
}

func addRats(left, right *big.Rat) Value {
	return NewRatioFromRat(new(big.Rat).Add(left, right))
}

func subRats(left, right *big.Rat) Value {
	return NewRatioFromRat(new(big.Rat).Sub(left, right))
}

func mulRats(left, right *big.Rat) Value {
	return NewRatioFromRat(new(big.Rat).Mul(left, right))
}

func divRats(left, right *big.Rat) Value {
	return NewRatioFromRat(new(big.Rat).Quo(left, right))
}

func longRat(n int64) *big.Rat {
	return big.NewRat(n, 1)
}

func intRat(n *big.Int) *big.Rat {
	return new(big.Rat).SetInt(n)
}

func copyRat(n *big.Rat) *big.Rat {
	return new(big.Rat).Set(n)
}

func bigIntToFloat64(n *big.Int) float64 {
	f, _ := new(big.Float).SetInt(n).Float64()
	return f
}

func quotFloat(left, right float64) Value {
	return NewDouble(math.Trunc(left / right))
}

func quotRats(left, right *big.Rat) Value {
	if right.Sign() == 0 {
		panic("quot by zero")
	}
	n := new(big.Int).Mul(left.Num(), right.Denom())
	d := new(big.Int).Mul(left.Denom(), right.Num())
	return newIntegerValueFromBigInt(new(big.Int).Quo(n, d))
}

const (
	maxInt64 = int64(^uint64(0) >> 1)
	minInt64 = -maxInt64 - 1
)
