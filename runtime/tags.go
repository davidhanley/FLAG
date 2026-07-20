package runtime

type ValueTag uint8

const (
	TagLong ValueTag = iota + 1
	TagDouble
	TagRatio
	TagList
	TagArray
)
