package runtime

import (
	"strings"
)

func NewVector(values ...Value) Value {
	if len(values) == 0 {
		return emptyVectorValue
	}
	tr := emptyVecTree.asTransient()
	for _, v := range values {
		tr.conj(v)
	}
	return tr.persistent()
}

func newVectorFromTree(tree *vecTree, origin, length int) Value {
	if tree == nil {
		panic("invalid vector")
	}
	if length < 0 || origin < 0 || origin+length > tree.cnt {
		panic("invalid vector length")
	}
	if length == 0 {
		return emptyVectorValue
	}
	return Value{d: packLong(int64(length)), impl: &vectorImpl{tree: tree, origin: origin}}
}

func (v Value) VectorLen() int {
	if !isType[*vectorImpl](v) {
		panic("VectorLen called on non-vector Value")
	}
	return int(v.Long())
}

func (v Value) vectorTree() *vectorImpl {
	return asImpl[*vectorImpl](v, "vectorTree called on non-vector Value")
}

func (v Value) VectorValues() []Value {
	if !isType[*vectorImpl](v) {
		panic("VectorValues called on non-vector Value")
	}
	n := v.VectorLen()
	out := make([]Value, n)
	impl := v.vectorTree()
	for i := 0; i < n; {
		arr := impl.tree.arrayFor(impl.origin + i)
		start := (impl.origin + i) & vecMask
		take := vecWidth - start
		if take > n-i {
			take = n - i
		}
		copy(out[i:], arr[start:start+take])
		i += take
	}
	return out
}

func VectorGet(vectorValue Value, index int) Value {
	if !isType[*vectorImpl](vectorValue) {
		panic("VectorGet expects vector Value")
	}
	if index < 0 || index >= vectorValue.VectorLen() {
		panic("vector index out of range")
	}
	impl := vectorValue.vectorTree()
	return impl.tree.get(impl.origin + index)
}

func VectorAssoc(vectorValue Value, index int, item Value) Value {
	if !isType[*vectorImpl](vectorValue) {
		panic("VectorAssoc expects vector Value")
	}
	length := vectorValue.VectorLen()
	if index < 0 || index >= length {
		panic("vector index out of range")
	}
	impl := vectorValue.vectorTree()
	tree := impl.tree.assocN(impl.origin+index, item)
	return newVectorFromTree(tree, impl.origin, length)
}

func VectorRest(vectorValue Value) Value {
	if !isType[*vectorImpl](vectorValue) {
		panic("VectorRest expects vector Value")
	}
	length := vectorValue.VectorLen()
	if length == 0 {
		return emptyVectorValue
	}
	impl := vectorValue.vectorTree()
	return newVectorFromTree(impl.tree, impl.origin+1, length-1)
}

func VectorAppend(vectorValue Value, item Value) Value {
	if !isType[*vectorImpl](vectorValue) {
		panic("VectorAppend expects vector Value")
	}
	impl := vectorValue.vectorTree()
	length := vectorValue.VectorLen()
	tree := impl.tree.assocN(impl.origin+length, item)
	return newVectorFromTree(tree, impl.origin, length+1)
}

func VectorPop(vectorValue Value) Value {
	if !isType[*vectorImpl](vectorValue) {
		panic("VectorPop expects vector Value")
	}
	n := vectorValue.VectorLen()
	if n == 0 {
		panic("Can't pop empty vector")
	}
	impl := vectorValue.vectorTree()
	if impl.origin == 0 && n == impl.tree.cnt {
		tree := impl.tree.pop()
		return newVectorFromTree(tree, 0, tree.cnt)
	}
	return newVectorFromTree(impl.tree, impl.origin, n-1)
}

type vectorImpl struct {
	defaultImpl
	tree   *vecTree
	origin int
}

func (vectorImpl) hashImpl(v Value) uint64 {
	n := v.VectorLen()
	h := hashSeedVector
	for i := 0; i < n; i++ {
		h = mix(h, Hash(VectorGet(v, i)))
	}
	return mix(h, uint64(n))
}
func (vectorImpl) eq(v Value, other Value) bool {
	if !isType[*vectorImpl](other) {
		return false
	}
	n := v.VectorLen()
	if n != other.VectorLen() {
		return false
	}
	a := v.vectorTree()
	b := other.vectorTree()
	if a.tree == b.tree && a.origin == b.origin {
		return true
	}
	for i := 0; i < n; i++ {
		if !Eq(VectorGet(v, i), VectorGet(other, i)) {
			return false
		}
	}
	return true
}
func (vectorImpl) typ(Value) Value { return typeVector }
func (vectorImpl) String(v Value) string {
	values := v.VectorValues()
	var out strings.Builder
	out.WriteByte('|')
	if len(values) == 0 {
		out.WriteString(" |")
		return out.String()
	}
	for _, value := range values {
		out.WriteByte(' ')
		out.WriteString(value.String())
	}
	out.WriteString(" |")
	return out.String()
}
func (vectorImpl) assoc(v Value, entries ...Value) Value {
	return assocIndexed(v, entries, func(coll Value) int { return coll.VectorLen() }, VectorAssoc, VectorAppend)
}
func (vectorImpl) get(v Value, key Value, missing Value) Value {
	if i, ok := associativeIndex(key); ok && i < v.VectorLen() {
		return VectorGet(v, i)
	}
	return missing
}
func (vectorImpl) first(v Value) Value {
	if v.VectorLen() == 0 {
		return NilValue
	}
	return VectorGet(v, 0)
}
func (vectorImpl) rest(v Value) Value {
	return VectorRest(v)
}
func (vectorImpl) next(v Value) Value {
	rest := VectorRest(v)
	return nextIfEmpty(rest, rest.VectorLen() == 0)
}
func (vectorImpl) last(v Value) Value {
	n := v.VectorLen()
	if n == 0 {
		return NilValue
	}
	return VectorGet(v, n-1)
}
func (vectorImpl) conj(v Value, entries ...Value) Value {
	out := v
	for _, item := range entries {
		out = VectorAppend(out, item)
	}
	return out
}
func (vectorImpl) count(v Value) int      { return v.VectorLen() }
func (vectorImpl) isEmpty(v Value) bool   { return v.VectorLen() == 0 }
func (vectorImpl) notEmpty(v Value) Value { return notEmptyIf(v, v.VectorLen() == 0) }
func (vectorImpl) nth(v Value, i int, missing Value, hasMissing bool) Value {
	return nthAt(v.VectorLen(), i, missing, hasMissing, func(idx int) Value { return VectorGet(v, idx) })
}
func (vectorImpl) slowNth(v Value, i int, missing Value, hasMissing bool) Value {
	return slowNthWalk(v, i, missing, hasMissing)
}
func (vectorImpl) drop(v Value, n int) Value {
	length := v.VectorLen()
	if n >= length {
		return emptyVectorValue
	}
	if n <= 0 {
		return v
	}
	impl := v.vectorTree()
	return newVectorFromTree(impl.tree, impl.origin+n, length-n)
}
func (vectorImpl) seqQ(Value) bool { return true }
func (vectorImpl) seqCursor(v Value) seqCursor {
	return arraySeqCursor(v.VectorValues())
}
func (vectorImpl) toNative(v Value) any {
	return nativeSeq(v.VectorValues())
}
func (vectorImpl) toJSON(v Value) (any, error) {
	return arrayToJSONValue(v.VectorValues())
}
func (vectorImpl) goSeq(v Value) []Value {
	return v.VectorValues()
}
