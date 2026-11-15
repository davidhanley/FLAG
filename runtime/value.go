package runtime

import (
	"math/big"
	"unsafe"
)

// impl is the per-type vtable on Value. Methods take Value so immediates can
// still read d. typ is type-of; type is a Go keyword.
type impl interface {
	hashImpl(v Value) uint64
	typ(v Value) Value
	String(v Value) string
	assoc(v Value, entries ...Value) Value
	dissoc(v Value, keys ...Value) Value
	get(v Value, key Value, missing Value) Value
	contains(v Value, key Value) bool
	find(v Value, key Value) Value
	keys(v Value) Value
	vals(v Value) Value
	first(v Value) Value
	rest(v Value) Value
	next(v Value) Value
	last(v Value) Value
	conj(v Value, entries ...Value) Value
	count(v Value) int
	isEmpty(v Value) bool
	notEmpty(v Value) Value
	seq(v Value) Value
	nth(v Value, i int, missing Value, hasMissing bool) Value
	slowNth(v Value, i int, missing Value, hasMissing bool) Value
	drop(v Value, n int) Value
	cons(v Value, item Value) Value
	seqQ(v Value) bool
	seqCursor(v Value) seqCursor
	eq(v Value, other Value) bool
	isTruthy(v Value) bool
	isNil(v Value) bool
	compare(v Value, other Value) Value
	toAny(v Value) any
	toNative(v Value) any
	asString(v Value) string
	asIdent(v Value) (string, bool)
	apply(v Value, args ...Value) Value
	asIndex(v Value) (int, bool)
	asCount(v Value, name string) int
	toJSON(v Value) (any, error)
	name(v Value) string
	exMessage(v Value) Value
	goSeq(v Value) []Value
	isNumeric(v Value) bool
	add(v, other Value) Value
	sub(v, other Value) Value
	mul(v, other Value) Value
	div(v, other Value) Value
	quot(v, other Value) Value
	cmpNum(v, other Value) int
	abs(v Value) Value
	toDouble(v Value) Value
	numerator(v Value) Value
	denominator(v Value) Value
	toFloat64(v Value) float64
	toRat(v Value) *big.Rat
	toBigInt(v Value) *big.Int
	bitLong(v Value, name string) int64
	addLong(v Value, n int64) Value
	addDouble(v Value, n float64) Value
	addRatio(v Value, n *big.Rat) Value
	addBigInt(v Value, n *big.Int) Value
	subLong(v Value, n int64) Value
	subDouble(v Value, n float64) Value
	subRatio(v Value, n *big.Rat) Value
	subBigInt(v Value, n *big.Int) Value
	mulLong(v Value, n int64) Value
	mulDouble(v Value, n float64) Value
	mulRatio(v Value, n *big.Rat) Value
	mulBigInt(v Value, n *big.Int) Value
	divLong(v Value, n int64) Value
	divDouble(v Value, n float64) Value
	divRatio(v Value, n *big.Rat) Value
	divBigInt(v Value, n *big.Int) Value
	quotLong(v Value, n int64) Value
	quotDouble(v Value, n float64) Value
	quotRatio(v Value, n *big.Rat) Value
	quotBigInt(v Value, n *big.Int) Value
	cmpLong(v Value, n int64) int
	cmpDouble(v Value, n float64) int
	cmpRatio(v Value, n *big.Rat) int
	cmpBigInt(v Value, n *big.Int) int
}

type Value struct {
	d    float64
	impl impl
}

func packLong(n int64) float64 {
	var d float64
	*(*int64)(unsafe.Pointer(&d)) = n
	return d
}

func asImpl[T impl](v Value, msg string) T {
	x, ok := v.impl.(T)
	if !ok {
		panic(msg)
	}
	return x
}

func isType[T impl](v Value) bool {
	_, ok := v.impl.(T)
	return ok
}

func (v Value) String() string {
	return v.impl.String(v)
}
