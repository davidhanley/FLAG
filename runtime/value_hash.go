package runtime

import (
	"hash/fnv"
	"math"
	"sort"
	"strconv"
	"strings"
	"unsafe"
)

type valueHasher struct{}

func (valueHasher) Hash(key Value) uint32 {
	sum := fnv.New32a()
	_, _ = sum.Write([]byte(valueIdentity(key)))
	return sum.Sum32()
}

func (valueHasher) Equal(a, b Value) bool {
	return valueIdentity(a) == valueIdentity(b)
}

func valueIdentity(v Value) string {
	switch v.tag {
	case TagLong:
		return "L:" + strconv.FormatInt(v.Long(), 10)
	case TagBigInt:
		return "I:" + v.BigInt().String()
	case TagDouble:
		if math.Trunc(v.Double()) == v.Double() {
			return "L:" + strconv.FormatInt(int64(v.Double()), 10)
		}
		return "D:" + strconv.FormatFloat(v.Double(), 'g', -1, 64)
	case TagRatio:
		if v.Ratio().IsInt() {
			num := v.Ratio().Num()
			if num.IsInt64() {
				return "L:" + strconv.FormatInt(num.Int64(), 10)
			}
			return "I:" + num.String()
		}
		return "R:" + v.Ratio().RatString()
	case TagBool:
		if v.Bool() {
			return "B:true"
		}
		return "B:false"
	case TagString:
		return "Q:" + v.StringValue()
	case TagDate:
		return "D:" + v.DateTime().UTC().Format("2006-01-02")
	case TagFile:
		return "H:" + v.FileObject().Path + ":" + strconv.FormatUint(uint64(uintptr(unsafe.Pointer(v.FileObject()))), 16)
	case TagChannel:
		return "C:" + strconv.FormatUint(uint64(uintptr(unsafe.Pointer(v.ChannelObject()))), 16)
	case TagSymbol:
		symbol := v.SymbolObject()
		if symbol.IsKeyword {
			return "K:" + symbol.Name
		}
		return "Y:" + symbol.Name
	case TagFunction:
		return "F:" + strconv.FormatUint(uint64(uintptr(unsafe.Pointer(v.FunctionObject()))), 16)
	case TagMap:
		entries := v.MapEntries()
		parts := make([]string, 0, len(entries))
		for _, entry := range entries {
			parts = append(parts, valueIdentity(entry.Key)+"="+valueIdentity(entry.Value))
		}
		sort.Strings(parts)
		return "M:{" + strings.Join(parts, ",") + "}"
	case TagSet:
		values := v.SetValues()
		parts := make([]string, 0, len(values))
		for _, value := range values {
			parts = append(parts, valueIdentity(value))
		}
		sort.Strings(parts)
		return "S:#{" + strings.Join(parts, ",") + "}"
	case TagNil:
		return "N:"
	case TagList:
		values := v.ListValues()
		parts := make([]string, 0, len(values))
		for _, value := range values {
			parts = append(parts, valueIdentity(value))
		}
		return "T:(" + strings.Join(parts, ",") + ")"
	case TagArray:
		values := v.ArrayValues()
		parts := make([]string, 0, len(values))
		for _, value := range values {
			parts = append(parts, valueIdentity(value))
		}
		return "A:[" + strings.Join(parts, ",") + "]"
	case TagLazyList:
		return "Z:" + strconv.FormatUint(uint64(uintptr(unsafe.Pointer(v.lazyListPointer()))), 16)
	default:
		panic("unknown Value tag")
	}
}
