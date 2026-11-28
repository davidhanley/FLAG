package runtime

var (
	typeInt      = NewKeyword("int")
	typeBigInt   = NewKeyword("bigint")
	typeFloat    = NewKeyword("float")
	typeRatio    = NewKeyword("ratio")
	typeBool     = NewKeyword("bool")
	typeString   = NewKeyword("string")
	typeDate     = NewKeyword("date")
	typeFile     = NewKeyword("file")
	typeSymbol   = NewKeyword("symbol")
	typeKeyword  = NewKeyword("keyword")
	typeFn       = NewKeyword("fn")
	typeMap      = NewKeyword("map")
	typeSet      = NewKeyword("set")
	typeNil      = NewKeyword("nil")
	typeList     = NewKeyword("list")
	typeArray    = NewKeyword("array")
	typeLazyList = NewKeyword("lazy-list")
	typeRange    = NewKeyword("range")
	typeChannel  = NewKeyword("channel")
	typeRecur    = NewKeyword("recur")
	typeRecord   = NewKeyword("record")
	typeVector   = NewKeyword("vector")
	typeAtom     = NewKeyword("atom")
	typeReduced  = NewKeyword("reduced")
)

func TypeOf(v Value) Value {
	return v.impl.typ(v)
}
