package runtime

type ValueTag uint8

const (
	TagLong ValueTag = iota + 1
	TagDouble
	TagRatio
	TagSymbol
	TagFunction
	TagMap
	TagSet
	TagNil
	TagList
	TagArray
)
