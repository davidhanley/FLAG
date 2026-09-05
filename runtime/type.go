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
	typeChannel  = NewKeyword("channel")
	typeRecur    = NewKeyword("recur")
	typeRecord   = NewKeyword("record")
	typeVector   = NewKeyword("vector")
)

func TypeOf(v Value) Value {
	switch v.tag {
	case TagLong:
		return typeInt
	case TagBigInt:
		return typeBigInt
	case TagDouble:
		return typeFloat
	case TagRatio:
		return typeRatio
	case TagBool:
		return typeBool
	case TagString:
		return typeString
	case TagDate:
		return typeDate
	case TagFile:
		return typeFile
	case TagSymbol:
		if v.SymbolObject().IsKeyword {
			return typeKeyword
		}
		return typeSymbol
	case TagFunction:
		return typeFn
	case TagMap:
		return typeMap
	case TagSet:
		return typeSet
	case TagNil:
		return typeNil
	case TagList:
		return typeList
	case TagArray:
		return typeArray
	case TagLazyList:
		return typeLazyList
	case TagChannel:
		return typeChannel
	case TagRecur:
		return typeRecur
	case TagRecord:
		return typeRecord
	case TagVector:
		return typeVector
	default:
		panic("unknown Value tag")
	}
}
