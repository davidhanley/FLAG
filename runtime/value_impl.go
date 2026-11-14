package runtime

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
