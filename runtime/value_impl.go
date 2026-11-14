package runtime

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"unsafe"
)

func implFor(tag ValueTag) impl {
	if tag == 0 || int(tag) >= len(implByTag) || implByTag[tag] == nil {
		panic("missing Value impl")
	}
	return implByTag[tag]
}

// defaultImpl panics for ops the concrete type does not override.
type defaultImpl struct{}

func (defaultImpl) assoc(Value, ...Value) Value {
	panic("assoc expects map, array, or vector Value")
}
func (defaultImpl) dissoc(Value, ...Value) Value {
	panic("dissoc expects map Value")
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

type longImpl struct{ defaultImpl }

func (longImpl) hash(v Value) string {
	return "L:" + strconv.FormatInt(v.Long(), 10)
}
func (longImpl) typ(Value) Value { return typeInt }
func (longImpl) String(v Value) string {
	return strconv.FormatInt(v.Long(), 10)
}

type bigIntImpl struct{ defaultImpl }

func (bigIntImpl) hash(v Value) string {
	return "I:" + v.BigInt().String()
}
func (bigIntImpl) typ(Value) Value { return typeBigInt }
func (bigIntImpl) String(v Value) string {
	return v.BigInt().String()
}

type doubleImpl struct{ defaultImpl }

func (doubleImpl) hash(v Value) string {
	if math.Trunc(v.Double()) == v.Double() {
		return "L:" + strconv.FormatInt(int64(v.Double()), 10)
	}
	return "D:" + strconv.FormatFloat(v.Double(), 'g', -1, 64)
}
func (doubleImpl) typ(Value) Value { return typeFloat }
func (doubleImpl) String(v Value) string {
	return strconv.FormatFloat(v.Double(), 'g', -1, 64)
}

type ratioImpl struct{ defaultImpl }

func (ratioImpl) hash(v Value) string {
	if v.Ratio().IsInt() {
		num := v.Ratio().Num()
		if num.IsInt64() {
			return "L:" + strconv.FormatInt(num.Int64(), 10)
		}
		return "I:" + num.String()
	}
	return "R:" + v.Ratio().RatString()
}
func (ratioImpl) typ(Value) Value { return typeRatio }
func (ratioImpl) String(v Value) string {
	return v.Ratio().RatString()
}

type boolImpl struct{ defaultImpl }

func (boolImpl) hash(v Value) string {
	if v.Bool() {
		return "B:true"
	}
	return "B:false"
}
func (boolImpl) typ(Value) Value { return typeBool }
func (boolImpl) String(v Value) string {
	return strconv.FormatBool(v.Bool())
}

type stringImpl struct{ defaultImpl }

func (stringImpl) hash(v Value) string {
	return "Q:" + v.StringValue()
}
func (stringImpl) typ(Value) Value { return typeString }
func (stringImpl) String(v Value) string {
	return strconv.Quote(v.StringValue())
}

type dateImpl struct{ defaultImpl }

func (dateImpl) hash(v Value) string {
	return "D:" + v.DateTime().UTC().Format("2006-01-02")
}
func (dateImpl) typ(Value) Value { return typeDate }
func (dateImpl) String(v Value) string {
	return dateString(v)
}

type fileImpl struct{ defaultImpl }

func (fileImpl) hash(v Value) string {
	return "H:" + v.FileObject().Path + ":" + strconv.FormatUint(uint64(uintptr(unsafe.Pointer(v.FileObject()))), 16)
}
func (fileImpl) typ(Value) Value { return typeFile }
func (fileImpl) String(v Value) string {
	return "#<file " + v.FileObject().Path + ">"
}

type symbolImpl struct{ defaultImpl }

func (symbolImpl) hash(v Value) string {
	symbol := v.SymbolObject()
	if symbol.IsKeyword {
		return "K:" + symbol.Name
	}
	return "Y:" + symbol.Name
}
func (symbolImpl) typ(v Value) Value {
	if v.SymbolObject().IsKeyword {
		return typeKeyword
	}
	return typeSymbol
}
func (symbolImpl) String(v Value) string {
	symbol := v.SymbolObject()
	if symbol.IsKeyword {
		return ":" + symbol.Name
	}
	return symbol.Name
}

type functionImpl struct{ defaultImpl }

func (functionImpl) hash(v Value) string {
	return "F:" + strconv.FormatUint(uint64(uintptr(unsafe.Pointer(v.FunctionObject()))), 16)
}
func (functionImpl) typ(Value) Value { return typeFn }
func (functionImpl) String(Value) string {
	return "#<fn>"
}

type mapImpl struct{ defaultImpl }

func (mapImpl) hash(v Value) string {
	entries := v.MapEntries()
	parts := make([]string, 0, len(entries))
	for _, entry := range entries {
		parts = append(parts, valueIdentity(entry.Key)+"="+valueIdentity(entry.Value))
	}
	sort.Strings(parts)
	return "M:{" + strings.Join(parts, ",") + "}"
}
func (mapImpl) typ(Value) Value { return typeMap }
func (mapImpl) String(v Value) string {
	entries := v.MapEntries()
	var out strings.Builder
	out.WriteByte('{')
	for i, entry := range entries {
		if i > 0 {
			out.WriteByte(' ')
		}
		out.WriteString(entry.Key.String())
		out.WriteByte(' ')
		out.WriteString(entry.Value.String())
	}
	out.WriteByte('}')
	return out.String()
}
func (mapImpl) assoc(v Value, entries ...Value) Value {
	return MapAssoc(v, entries...)
}
func (mapImpl) dissoc(v Value, keys ...Value) Value {
	items := v.mapPointer().items
	for _, key := range keys {
		items = items.Delete(key)
	}
	return newMapValue(items)
}

type setImpl struct{ defaultImpl }

func (setImpl) hash(v Value) string {
	values := v.SetValues()
	parts := make([]string, 0, len(values))
	for _, value := range values {
		parts = append(parts, valueIdentity(value))
	}
	sort.Strings(parts)
	return "S:#{" + strings.Join(parts, ",") + "}"
}
func (setImpl) typ(Value) Value { return typeSet }
func (setImpl) String(v Value) string {
	values := v.SetValues()
	var out strings.Builder
	out.WriteString("#{")
	for i, value := range values {
		if i > 0 {
			out.WriteByte(' ')
		}
		out.WriteString(value.String())
	}
	out.WriteByte('}')
	return out.String()
}

type nilImpl struct{ defaultImpl }

func (nilImpl) hash(Value) string { return "N:" }
func (nilImpl) typ(Value) Value   { return typeNil }
func (nilImpl) String(Value) string  { return "" }
func (nilImpl) dissoc(Value, ...Value) Value {
	return NilValue
}

type listImpl struct{ defaultImpl }

func (listImpl) hash(v Value) string {
	values := v.ListValues()
	parts := make([]string, 0, len(values))
	for _, value := range values {
		parts = append(parts, valueIdentity(value))
	}
	return "T:(" + strings.Join(parts, ",") + ")"
}
func (listImpl) typ(Value) Value { return typeList }
func (listImpl) String(v Value) string {
	values := v.ListValues()
	var out strings.Builder
	out.WriteByte('(')
	for i, value := range values {
		if i > 0 {
			out.WriteByte(' ')
		}
		out.WriteString(value.String())
	}
	out.WriteByte(')')
	return out.String()
}

type arrayImpl struct{ defaultImpl }

func (arrayImpl) hash(v Value) string {
	values := v.ArrayValues()
	parts := make([]string, 0, len(values))
	for _, value := range values {
		parts = append(parts, valueIdentity(value))
	}
	return "A:[" + strings.Join(parts, ",") + "]"
}
func (arrayImpl) typ(Value) Value { return typeArray }
func (arrayImpl) String(v Value) string {
	values := v.ArrayValues()
	var out strings.Builder
	out.WriteByte('[')
	for i, value := range values {
		if i > 0 {
			out.WriteByte(' ')
		}
		out.WriteString(value.String())
	}
	out.WriteByte(']')
	return out.String()
}
func (arrayImpl) assoc(v Value, entries ...Value) Value {
	return assocIndexed(v, entries, func(coll Value) int { return coll.ArrayLen() }, ArrayAssoc, ArrayAppend)
}

type lazyListImpl struct{ defaultImpl }

func (lazyListImpl) hash(v Value) string {
	return "Z:" + strconv.FormatUint(uint64(uintptr(unsafe.Pointer(v.lazyListPointer()))), 16)
}
func (lazyListImpl) typ(Value) Value { return typeLazyList }
func (lazyListImpl) String(Value) string {
	return "#<lazy-list>"
}

type channelImpl struct{ defaultImpl }

func (channelImpl) hash(v Value) string {
	return "C:" + strconv.FormatUint(uint64(uintptr(unsafe.Pointer(v.ChannelObject()))), 16)
}
func (channelImpl) typ(Value) Value { return typeChannel }
func (channelImpl) String(Value) string {
	return "#<channel>"
}

type recurImpl struct{ defaultImpl }

func (recurImpl) hash(v Value) string {
	return "U:" + strconv.FormatUint(uint64(uintptr(v.p)), 16)
}
func (recurImpl) typ(Value) Value { return typeRecur }
func (recurImpl) String(Value) string {
	return "#<recur>"
}

type recordImpl struct{ defaultImpl }

func (recordImpl) hash(v Value) string {
	return "R:" + strconv.FormatUint(uint64(uintptr(v.p)), 16)
}
func (recordImpl) typ(Value) Value { return typeRecord }
func (recordImpl) String(v Value) string {
	return fmt.Sprintf("%#v", v.recordBox().value)
}

type vectorImpl struct{ defaultImpl }

func (vectorImpl) hash(v Value) string {
	values := v.VectorValues()
	parts := make([]string, 0, len(values))
	for _, value := range values {
		parts = append(parts, valueIdentity(value))
	}
	return "V:|" + strings.Join(parts, ",") + "|"
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

type atomImpl struct{ defaultImpl }

func (atomImpl) hash(v Value) string {
	return "O:" + strconv.FormatUint(uint64(uintptr(v.p)), 16)
}
func (atomImpl) typ(Value) Value { return typeAtom }
func (atomImpl) String(Value) string {
	return "#<atom>"
}
