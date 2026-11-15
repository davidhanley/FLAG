package runtime

import (
	"fmt"
	"math/big"
)

func implFor(tag ValueTag) impl {
	if tag == 0 || int(tag) >= len(implByTag) || implByTag[tag] == nil {
		panic("missing Value impl")
	}
	return implByTag[tag]
}

// defaultImpl panics for optional vtable ops. Each type file must implement
// hash, typ, and String.
type defaultImpl struct{}

func (defaultImpl) assoc(Value, ...Value) Value {
	panic("assoc expects map, array, or vector Value")
}
func (defaultImpl) dissoc(Value, ...Value) Value {
	panic("dissoc expects map Value")
}
func (defaultImpl) get(Value, Value, Value) Value {
	panic("get expects associative Value")
}
func (defaultImpl) contains(Value, Value) bool {
	panic("contains? expects map or set Value")
}
func (defaultImpl) find(Value, Value) Value {
	panic("find expects map Value")
}
func (defaultImpl) keys(Value) Value {
	panic("keys expects map Value")
}
func (defaultImpl) vals(Value) Value {
	panic("vals expects map Value")
}
func (defaultImpl) first(Value) Value {
	panic("first expects nil, list, array, vector, lazy-list, set, or string Value")
}
func (defaultImpl) rest(Value) Value {
	panic("rest expects list, array, vector, lazy-list, set, or string Value")
}
func (defaultImpl) next(Value) Value {
	panic("next expects list, array, vector, or lazy-list Value")
}
func (defaultImpl) last(Value) Value {
	panic("last expects list, array, vector, or lazy-list Value")
}
func (defaultImpl) conj(Value, ...Value) Value {
	panic("conj expects list, array, vector, set, or map Value")
}
func (defaultImpl) count(Value) int {
	panic("count expects collection Value")
}
func (defaultImpl) isEmpty(Value) bool {
	panic("empty? expects collection Value")
}
func (defaultImpl) notEmpty(Value) Value {
	panic("not-empty expects list, array, vector, lazy-list, map, set, or string Value")
}
func (defaultImpl) seq(v Value) Value {
	return NotEmpty(v)
}
func (defaultImpl) nth(Value, int, Value, bool) Value {
	panic("nth expects array, vector, string, or nil Value")
}
func (defaultImpl) slowNth(Value, int, Value, bool) Value {
	panic("slow-nth expects list, array, vector, lazy-list, string, or nil Value")
}
func (defaultImpl) drop(Value, int) Value {
	panic("drop expects list, array, vector, or lazy-list Value")
}
func (defaultImpl) cons(v Value, item Value) Value {
	return ListCons(seqToList(v), item)
}
func (defaultImpl) seqQ(Value) bool { return false }
func (defaultImpl) seqCursor(Value) seqCursor {
	panic("sequence expects list, array, vector, lazy-list, map, set, string, or nil Value")
}
func (defaultImpl) eq(v Value, other Value) bool {
	return valueIdentity(v) == valueIdentity(other)
}
func (defaultImpl) isTruthy(Value) bool { return true }
func (defaultImpl) isNil(Value) bool    { return false }
func (defaultImpl) compare(v Value, other Value) Value {
	if other.impl.isNil(other) {
		return NewLong(1)
	}
	return NewLong(int64(compareOrdered(v, other)))
}
func (defaultImpl) toAny(v Value) any    { return v }
func (defaultImpl) toNative(v Value) any { return v }
func (defaultImpl) asString(v Value) string {
	return v.String()
}
func (defaultImpl) asIdent(Value) (string, bool) { return "", false }
func (defaultImpl) toJSON(v Value) (any, error) {
	return nil, fmt.Errorf("unsupported value type for JSON: %s", TypeOf(v).String())
}
func (defaultImpl) name(Value) string {
	panic("name expects symbol or keyword")
}
func (defaultImpl) exMessage(Value) Value { return NilValue }
func (defaultImpl) goSeq(Value) []Value {
	panic("ListValues called on non-list Value")
}
func (defaultImpl) isNumeric(Value) bool { return false }
func (defaultImpl) add(Value, Value) Value {
	panic("Add expects numeric Value arguments")
}
func (defaultImpl) sub(Value, Value) Value {
	panic("Sub expects numeric Value arguments")
}
func (defaultImpl) mul(Value, Value) Value {
	panic("Mul expects numeric Value arguments")
}
func (defaultImpl) div(Value, Value) Value {
	panic("Div expects numeric Value arguments")
}
func (defaultImpl) quot(Value, Value) Value {
	panic("quot expects numeric Value arguments")
}
func (defaultImpl) cmpNum(Value, Value) int {
	panic("compare expects numeric Value arguments")
}
func (defaultImpl) abs(Value) Value {
	panic("abs expects numeric Value argument")
}
func (defaultImpl) toDouble(Value) Value {
	panic("double expects numeric Value")
}
func (defaultImpl) numerator(Value) Value {
	panic("numerator expects an integer or ratio")
}
func (defaultImpl) denominator(Value) Value {
	panic("denominator expects an integer or ratio")
}
func (defaultImpl) toFloat64(Value) float64 {
	panic("numericToFloat64 called on non-numeric Value")
}
func (defaultImpl) toRat(Value) *big.Rat {
	panic("valueToRat expects integer or ratio Value")
}
func (defaultImpl) toBigInt(Value) *big.Int {
	panic("valueToBigInt expects integer Value")
}
func (defaultImpl) bitLong(_ Value, name string) int64 {
	panic(name + " expects an integer")
}
func (defaultImpl) addLong(Value, int64) Value {
	panic("Add expects numeric Value arguments")
}
func (defaultImpl) addDouble(Value, float64) Value {
	panic("Add expects numeric Value arguments")
}
func (defaultImpl) addRatio(Value, *big.Rat) Value {
	panic("Add expects numeric Value arguments")
}
func (defaultImpl) addBigInt(Value, *big.Int) Value {
	panic("Add expects numeric Value arguments")
}
func (defaultImpl) subLong(Value, int64) Value {
	panic("Sub expects numeric Value arguments")
}
func (defaultImpl) subDouble(Value, float64) Value {
	panic("Sub expects numeric Value arguments")
}
func (defaultImpl) subRatio(Value, *big.Rat) Value {
	panic("Sub expects numeric Value arguments")
}
func (defaultImpl) subBigInt(Value, *big.Int) Value {
	panic("Sub expects numeric Value arguments")
}
func (defaultImpl) mulLong(Value, int64) Value {
	panic("Mul expects numeric Value arguments")
}
func (defaultImpl) mulDouble(Value, float64) Value {
	panic("Mul expects numeric Value arguments")
}
func (defaultImpl) mulRatio(Value, *big.Rat) Value {
	panic("Mul expects numeric Value arguments")
}
func (defaultImpl) mulBigInt(Value, *big.Int) Value {
	panic("Mul expects numeric Value arguments")
}
func (defaultImpl) divLong(Value, int64) Value {
	panic("Div expects numeric Value arguments")
}
func (defaultImpl) divDouble(Value, float64) Value {
	panic("Div expects numeric Value arguments")
}
func (defaultImpl) divRatio(Value, *big.Rat) Value {
	panic("Div expects numeric Value arguments")
}
func (defaultImpl) divBigInt(Value, *big.Int) Value {
	panic("Div expects numeric Value arguments")
}
func (defaultImpl) quotLong(Value, int64) Value {
	panic("quot expects numeric Value arguments")
}
func (defaultImpl) quotDouble(Value, float64) Value {
	panic("quot expects numeric Value arguments")
}
func (defaultImpl) quotRatio(Value, *big.Rat) Value {
	panic("quot expects numeric Value arguments")
}
func (defaultImpl) quotBigInt(Value, *big.Int) Value {
	panic("quot expects numeric Value arguments")
}
func (defaultImpl) cmpLong(Value, int64) int {
	panic("compare expects numeric Value arguments")
}
func (defaultImpl) cmpDouble(Value, float64) int {
	panic("compare expects numeric Value arguments")
}
func (defaultImpl) cmpRatio(Value, *big.Rat) int {
	panic("compare expects numeric Value arguments")
}
func (defaultImpl) cmpBigInt(Value, *big.Int) int {
	panic("compare expects numeric Value arguments")
}

var implByTag = [TagAtom + 1]impl{
	TagLong:     longImpl{},
	TagBigInt:   bigIntImpl{},
	TagDouble:   doubleImpl{},
	TagRatio:    ratioImpl{},
	TagBool:     boolImpl{},
	TagString:   stringImpl{},
	TagDate:     dateImpl{},
	TagFile:     fileImpl{},
	TagSymbol:   symbolImpl{},
	TagFunction: functionImpl{},
	TagMap:      mapImpl{},
	TagSet:      setImpl{},
	TagNil:      nilImpl{},
	TagList:     listImpl{},
	TagArray:    arrayImpl{},
	TagLazyList: lazyListImpl{},
	TagChannel:  channelImpl{},
	TagRecur:    recurImpl{},
	TagRecord:   recordImpl{},
	TagVector:   vectorImpl{},
	TagAtom:     atomImpl{},
}
