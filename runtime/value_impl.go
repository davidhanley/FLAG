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
