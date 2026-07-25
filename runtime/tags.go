package runtime

type ValueTag uint8

const (
	TagLong ValueTag = iota + 1
	TagBigInt
	TagDouble
	TagRatio
	TagBool
	TagSymbol
	TagFunction
	TagMap
	TagSet
	TagNil
	TagList
	TagArray
	TagLazyList
)
