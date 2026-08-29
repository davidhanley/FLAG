package runtime

import (
	"sort"
	"unsafe"

	"github.com/benbjohnson/immutable"
)

type MapEntry struct {
	Key   Value
	Value Value
}

type mapData struct {
	items *immutable.Map[Value, Value]
}

type setData struct {
	items immutable.Set[Value]
}

func NewMap(entries ...Value) Value {
	if len(entries)%2 != 0 {
		panic("NewMap expects key/value pairs")
	}

	items := immutable.NewMap[Value, Value](valueHasher{})
	for i := 0; i < len(entries); i += 2 {
		items = items.Set(entries[i], entries[i+1])
	}
	return newMapValue(items)
}

func NewSet(values ...Value) Value {
	items := immutable.NewSet[Value](valueHasher{}, values...)
	return newSetValue(items)
}

func Assoc(coll Value, entries ...Value) Value {
	return MapAssoc(coll, entries...)
}

func SelectKeys(coll Value, keys Value) Value {
	if coll.tag == TagNil {
		return NewMap()
	}
	if coll.tag != TagMap && coll.tag != TagDate {
		panic("select-keys expects map Value")
	}
	cursor := newSeqCursor(keys)
	result := NewMap()
	for {
		key, ok := cursor.nextOrDone()
		if !ok {
			return result
		}
		if Contains(coll, key) {
			result = MapAssoc(result, key, Get(coll, key))
		}
	}
}

func Keys(coll Value) Value {
	switch coll.tag {
	case TagNil:
		return NilValue()
	case TagMap, TagDate:
		entries := coll.MapEntries()
		keys := make([]Value, 0, len(entries))
		for _, entry := range entries {
			keys = append(keys, entry.Key)
		}
		return NewArray(keys...)
	default:
		panic("keys expects map Value")
	}
}

func GetIn(coll Value, path Value) Value {
	cursor := newSeqCursor(path)
	current := coll
	for {
		key, ok := cursor.nextOrDone()
		if !ok {
			return current
		}
		current = getInStep(current, key)
		if current.tag == TagNil {
			return current
		}
	}
}

func getInStep(coll Value, key Value) Value {
	switch coll.tag {
	case TagNil:
		return NilValue()
	case TagMap, TagDate:
		return Get(coll, key)
	case TagArray:
		index, ok := getInIndex(key)
		if !ok || index < 0 || index >= coll.ArrayLen() {
			return NilValue()
		}
		return ArrayGet(coll, index)
	case TagList:
		index, ok := getInIndex(key)
		if !ok || index < 0 {
			return NilValue()
		}
		cursor := newSeqCursor(coll)
		for i := 0; i <= index; i++ {
			next, ok := cursor.nextOrDone()
			if !ok {
				return NilValue()
			}
			if i == index {
				return next
			}
		}
		return NilValue()
	default:
		panic("get-in expects map, list, or array Value")
	}
}

func getInIndex(key Value) (int, bool) {
	switch key.tag {
	case TagLong:
		return int(key.Long()), true
	case TagBigInt:
		bi := key.BigInt()
		if bi.BitLen() > 62 {
			return 0, false
		}
		return int(bi.Int64()), true
	default:
		return 0, false
	}
}

func Merge(maps ...Value) Value {
	if len(maps) == 0 {
		return NilValue()
	}
	var (
		items   *immutable.Map[Value, Value]
		started bool
	)
	for _, coll := range maps {
		if coll.tag == TagNil {
			continue
		}
		if coll.tag != TagMap {
			panic("merge expects map Value")
		}
		if !started {
			items = coll.mapPointer().items
			started = true
			continue
		}
		for _, entry := range coll.MapEntries() {
			items = items.Set(entry.Key, entry.Value)
		}
	}
	if !started {
		return NilValue()
	}
	return newMapValue(items)
}

func Update(coll Value, key Value, fn Value, args ...Value) Value {
	if coll.tag == TagNil {
		coll = NewMap()
	}
	switch coll.tag {
	case TagMap:
		current := Get(coll, key)
		callArgs := make([]Value, 0, 1+len(args))
		callArgs = append(callArgs, current)
		callArgs = append(callArgs, args...)
		next := Call(fn, callArgs...)
		return MapAssoc(coll, key, next)
	case TagArray:
		index := updateArrayIndex(key)
		current := ArrayGet(coll, index)
		callArgs := make([]Value, 0, 1+len(args))
		callArgs = append(callArgs, current)
		callArgs = append(callArgs, args...)
		next := Call(fn, callArgs...)
		return ArrayAssoc(coll, index, next)
	default:
		panic("update expects map or array Value")
	}
}

func updateArrayIndex(key Value) int {
	switch key.tag {
	case TagLong:
		index := key.Long()
		if index < 0 {
			panic("update array index must be non-negative")
		}
		return int(index)
	case TagBigInt:
		bi := key.BigInt()
		if bi.Sign() < 0 {
			panic("update array index must be non-negative")
		}
		if bi.BitLen() > 62 {
			panic("update array index is too large")
		}
		return int(bi.Int64())
	default:
		panic("update array key must be integer")
	}
}

func MapAssoc(coll Value, entries ...Value) Value {
	if len(entries) < 2 || len(entries)%2 != 0 {
		panic("assoc expects collection and key/value pairs")
	}

	switch coll.tag {
	case TagMap:
		items := coll.mapPointer().items
		for i := 0; i < len(entries); i += 2 {
			items = items.Set(entries[i], entries[i+1])
		}
		return newMapValue(items)
	default:
		panic("assoc expects map Value")
	}
}

func Dissoc(coll Value, keys ...Value) Value {
	return MapDissoc(coll, keys...)
}

func Get(coll Value, key Value, notFound ...Value) Value {
	missing := NilValue()
	if len(notFound) > 1 {
		panic("get expects map, key, and optional default")
	}
	if len(notFound) == 1 {
		missing = notFound[0]
	}

	if coll.tag == TagNil {
		return missing
	}
	if coll.tag == TagDate {
		if value, ok := dateFieldValue(coll, key); ok {
			return value
		}
		return missing
	}
	if coll.tag == TagRecord {
		if value, ok := recordFieldValue(coll, key); ok {
			return value
		}
		return missing
	}
	if coll.tag != TagMap {
		panic("get expects map Value")
	}
	value, ok := coll.mapPointer().items.Get(key)
	if !ok {
		return missing
	}
	return value
}

func Contains(coll Value, key Value) bool {
	switch coll.tag {
	case TagMap:
		_, ok := coll.mapPointer().items.Get(key)
		return ok
	case TagDate:
		_, ok := dateFieldValue(coll, key)
		return ok
	case TagRecord:
		_, ok := recordFieldValue(coll, key)
		return ok
	case TagSet:
		return coll.setPointer().items.Has(key)
	default:
		panic("contains? expects map or set Value")
	}
}

func MapDissoc(coll Value, keys ...Value) Value {
	switch coll.tag {
	case TagMap:
		items := coll.mapPointer().items
		for _, key := range keys {
			items = items.Delete(key)
		}
		return newMapValue(items)
	default:
		panic("dissoc expects map Value")
	}
}

func (v Value) MapLen() int {
	switch v.tag {
	case TagMap:
		return v.mapPointer().items.Len()
	case TagDate:
		return len(dateEntries(v))
	default:
		panic("MapLen called on non-map Value")
	}
}

func (v Value) MapEntries() []MapEntry {
	switch v.tag {
	case TagMap:
		itr := v.mapPointer().items.Iterator()
		entries := make([]MapEntry, 0, v.MapLen())
		for !itr.Done() {
			key, value, _ := itr.Next()
			entries = append(entries, MapEntry{Key: key, Value: value})
		}
		sort.Slice(entries, func(i, j int) bool {
			return valueIdentity(entries[i].Key) < valueIdentity(entries[j].Key)
		})
		return entries
	case TagDate:
		return dateEntries(v)
	default:
		panic("MapEntries called on non-map Value")
	}
}

func (v Value) SetLen() int {
	if v.tag != TagSet {
		panic("SetLen called on non-set Value")
	}
	return v.setPointer().items.Len()
}

func (v Value) SetValues() []Value {
	if v.tag != TagSet {
		panic("SetValues called on non-set Value")
	}
	itr := v.setPointer().items.Iterator()
	values := make([]Value, 0, v.SetLen())
	for !itr.Done() {
		value, _ := itr.Next()
		values = append(values, value)
	}
	sort.Slice(values, func(i, j int) bool {
		return valueIdentity(values[i]) < valueIdentity(values[j])
	})
	return values
}

func newMapValue(items *immutable.Map[Value, Value]) Value {
	return Value{p: unsafe.Pointer(&mapData{items: items}), tag: TagMap}
}

func newSetValue(items immutable.Set[Value]) Value {
	return Value{p: unsafe.Pointer(&setData{items: items}), tag: TagSet}
}

func (v Value) mapPointer() *mapData {
	if v.tag != TagMap {
		panic("mapPointer called on non-map Value")
	}
	if v.p == nil {
		panic("map Value does not contain map pointer")
	}
	return (*mapData)(v.p)
}

func (v Value) setPointer() *setData {
	if v.tag != TagSet {
		panic("setPointer called on non-set Value")
	}
	if v.p == nil {
		panic("set Value does not contain set pointer")
	}
	return (*setData)(v.p)
}
