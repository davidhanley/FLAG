package runtime

import (
	"sort"

	"github.com/benbjohnson/immutable"
	"strings"
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
	kvs := compactMapKVs(entries)
	if len(kvs)/2 > arrayMapMax {
		return hashMapFromKVs(kvs)
	}
	return newArrayMap(kvs)
}

func compactMapKVs(entries []Value) []Value {
	out := make([]Value, 0, len(entries))
	for i := 0; i < len(entries); i += 2 {
		k, val := entries[i], entries[i+1]
		found := -1
		for j := 0; j < len(out); j += 2 {
			if valueSlotEq(out[j], k) || Eq(out[j], k) {
				found = j
				break
			}
		}
		if found >= 0 {
			out[found+1] = val
			continue
		}
		out = append(out, k, val)
	}
	return out
}

func hashMapFromKVs(kvs []Value) Value {
	items := immutable.NewMap[Value, Value](valueHasher{})
	for i := 0; i < len(kvs); i += 2 {
		items = items.Set(kvs[i], kvs[i+1])
	}
	return newMapValue(items)
}

func mapsHash(v Value) uint64 {
	h := hashSeedMap
	n := uint64(0)
	for _, entry := range unsortedMapEntries(v) {
		h += mix(Hash(entry.Key), Hash(entry.Value))
		n++
	}
	return mix(h, n)
}

func mapsEq(a, b Value) bool {
	if !isMap(b) || a.MapLen() != b.MapLen() {
		return false
	}
	for _, entry := range unsortedMapEntries(a) {
		if !Contains(b, entry.Key) || !Eq(Get(b, entry.Key), entry.Value) {
			return false
		}
	}
	return true
}

func mapString(v Value) string {
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

func NewSet(values ...Value) Value {
	items := immutable.NewSet[Value](valueHasher{}, values...)
	return newSetValue(items)
}

func Assoc(coll Value, entries ...Value) Value {
	return coll.impl.assoc(coll, entries...)
}

func assocIndexed(coll Value, entries []Value, length func(Value) int, assocAt func(Value, int, Value) Value, appendAt func(Value, Value) Value) Value {
	if len(entries) < 2 || len(entries)%2 != 0 {
		panic("assoc expects collection and key/value pairs")
	}
	out := coll
	for i := 0; i < len(entries); i += 2 {
		index := nonNegativeCount("assoc", entries[i])
		item := entries[i+1]
		n := length(out)
		if index < n {
			out = assocAt(out, index, item)
			continue
		}
		if index == n {
			out = appendAt(out, item)
			continue
		}
		panic("assoc index out of bounds")
	}
	return out
}

func Keys(coll Value) Value {
	return coll.impl.keys(coll)
}

func Vals(coll Value) Value {
	return coll.impl.vals(coll)
}

func mapEntryKeys(entries []MapEntry) Value {
	keys := make([]Value, 0, len(entries))
	for _, entry := range entries {
		keys = append(keys, entry.Key)
	}
	return NewArray(keys...)
}

func mapEntryVals(entries []MapEntry) Value {
	vals := make([]Value, 0, len(entries))
	for _, entry := range entries {
		vals = append(vals, entry.Value)
	}
	return NewArray(vals...)
}

func Find(coll Value, key Value) Value {
	return coll.impl.find(coll, key)
}

func findEntry(key Value, value Value, ok bool) Value {
	if !ok {
		return NilValue
	}
	return NewArray(key, value)
}

func Union(sets ...Value) Value {
	items := immutable.NewSet[Value](valueHasher{})
	for _, set := range sets {
		if IsNil(set) {
			continue
		}
		if !isType[*setImpl](set) {
			panic("union expects set arguments")
		}
		for _, value := range set.SetValues() {
			items = items.Add(value)
		}
	}
	return newSetValue(items)
}

func Intersection(sets ...Value) Value {
	if len(sets) == 0 {
		return NewSet()
	}
	for _, set := range sets {
		if !isType[*setImpl](set) && !IsNil(set) {
			panic("intersection expects set arguments")
		}
		if IsNil(set) {
			return NewSet()
		}
	}
	base := sets[0]
	if IsNil(base) {
		return NewSet()
	}
	items := immutable.NewSet[Value](valueHasher{})
	for _, value := range base.SetValues() {
		keep := true
		for i := 1; i < len(sets); i++ {
			if !sets[i].setPointer().items.Has(value) {
				keep = false
				break
			}
		}
		if keep {
			items = items.Add(value)
		}
	}
	return newSetValue(items)
}

func Difference(sets ...Value) Value {
	if len(sets) == 0 {
		return NewSet()
	}
	first := sets[0]
	if IsNil(first) {
		return NewSet()
	}
	if !isType[*setImpl](first) {
		panic("difference expects set arguments")
	}
	items := first.setPointer().items
	for _, set := range sets[1:] {
		if IsNil(set) {
			continue
		}
		if !isType[*setImpl](set) {
			panic("difference expects set arguments")
		}
		for _, value := range set.SetValues() {
			items = items.Delete(value)
		}
	}
	return newSetValue(items)
}

func Subset(set1 Value, set2 Value) bool {
	if IsNil(set1) {
		return true
	}
	if !isType[*setImpl](set1) {
		panic("subset? expects set arguments")
	}
	if IsNil(set2) {
		return set1.SetLen() == 0
	}
	if !isType[*setImpl](set2) {
		panic("subset? expects set arguments")
	}
	for _, value := range set1.SetValues() {
		if !set2.setPointer().items.Has(value) {
			return false
		}
	}
	return true
}

func Superset(set1 Value, set2 Value) bool {
	return Subset(set2, set1)
}

func Disjoint(set1 Value, set2 Value) bool {
	if IsNil(set1) || IsNil(set2) {
		return true
	}
	if !isType[*setImpl](set1) || !isType[*setImpl](set2) {
		panic("disjoint? expects set arguments")
	}
	left := set1
	right := set2
	if right.SetLen() < left.SetLen() {
		left, right = right, left
	}
	for _, value := range left.SetValues() {
		if right.setPointer().items.Has(value) {
			return false
		}
	}
	return true
}

func RenameKeys(coll Value, keyMap Value) Value {
	if IsNil(coll) {
		return NilValue
	}
	if !isMap(coll) {
		panic("rename-keys expects map and key-map")
	}
	if IsNil(keyMap) {
		return coll
	}
	if !isMap(keyMap) {
		panic("rename-keys expects map and key-map")
	}
	out := coll
	for _, entry := range keyMap.MapEntries() {
		oldKey := entry.Key
		newKey := entry.Value
		if !Contains(out, oldKey) {
			continue
		}
		value := Get(out, oldKey)
		out = MapDissoc(out, oldKey)
		out = MapAssoc(out, newKey, value)
	}
	return out
}

func MapInvert(coll Value) Value {
	if IsNil(coll) {
		return NilValue
	}
	if !isMap(coll) {
		panic("map-invert expects map")
	}
	out := NewMap()
	for _, entry := range coll.MapEntries() {
		out = MapAssoc(out, entry.Value, entry.Key)
	}
	return out
}

func SetSelect(pred Value, set Value) Value {
	if IsNil(set) {
		return NewSet()
	}
	if !isType[*setImpl](set) {
		panic("select expects predicate and set")
	}
	items := immutable.NewSet[Value](valueHasher{})
	for _, value := range set.SetValues() {
		if IsTruthy(Call(pred, value)) {
			items = items.Add(value)
		}
	}
	return newSetValue(items)
}

func SetProject(rel Value, keys Value) Value {
	if IsNil(rel) {
		return NewSet()
	}
	if !isType[*setImpl](rel) {
		panic("project expects relation set and key sequence")
	}
	keyCursor := newSeqCursor(keys)
	keyList := make([]Value, 0)
	for {
		key, ok := keyCursor.nextOrDone()
		if !ok {
			break
		}
		keyList = append(keyList, key)
	}

	items := immutable.NewSet[Value](valueHasher{})
	for _, row := range rel.SetValues() {
		items = items.Add(projectRow(row, keyList))
	}
	return newSetValue(items)
}

func SetRename(rel Value, keyMap Value) Value {
	if IsNil(rel) {
		return NewSet()
	}
	if !isType[*setImpl](rel) {
		panic("rename expects relation set and key-map")
	}
	items := immutable.NewSet[Value](valueHasher{})
	for _, row := range rel.SetValues() {
		items = items.Add(RenameKeys(row, keyMap))
	}
	return newSetValue(items)
}

func projectRow(row Value, keys []Value) Value {
	if !isMap(row) && !isType[*dateImpl](row) && !isType[*recordImpl](row) {
		panic("project expects set of map-like values")
	}
	out := NewMap()
	for _, key := range keys {
		if Contains(row, key) {
			out = MapAssoc(out, key, Get(row, key))
		}
	}
	return out
}

func MapAssoc(coll Value, entries ...Value) Value {
	return coll.impl.assoc(coll, entries...)
}

func hashMapAssoc(coll Value, entries ...Value) Value {
	if len(entries) < 2 || len(entries)%2 != 0 {
		panic("assoc expects collection and key/value pairs")
	}
	items := coll.mapPointer().items
	for i := 0; i < len(entries); i += 2 {
		items = items.Set(entries[i], entries[i+1])
	}
	return newMapValue(items)
}

func Dissoc(coll Value, keys ...Value) Value {
	return coll.impl.dissoc(coll, keys...)
}

func Get(coll Value, key Value, notFound ...Value) Value {
	missing := NilValue
	if len(notFound) > 1 {
		panic("get expects map, key, and optional default")
	}
	if len(notFound) == 1 {
		missing = notFound[0]
	}
	return fastGet(coll, key, missing)
}

func fastGet(coll, key, missing Value) Value {
	if m, ok := coll.impl.(*arrayMapImpl); ok {
		if value, found := arrayMapLookup(m.kvs, key); found {
			return value
		}
		return missing
	}
	if m, ok := coll.impl.(*mapImpl); ok {
		if value, found := m.data.items.Get(key); found {
			return value
		}
		return missing
	}
	return coll.impl.get(coll, key, missing)
}

func associativeIndex(key Value) (int, bool) {
	return key.impl.asIndex(key)
}

func Contains(coll Value, key Value) bool {
	return coll.impl.contains(coll, key)
}

func MapDissoc(coll Value, keys ...Value) Value {
	return Dissoc(coll, keys...)
}

func (v Value) MapLen() int {
	if isType[*arrayMapImpl](v) {
		return len(asImpl[*arrayMapImpl](v, "MapLen").kvs) / 2
	}
	if isType[*mapImpl](v) {
		return v.mapPointer().items.Len()
	}
	if isType[*dateImpl](v) {
		return len(dateEntries(v))
	}
	panic("MapLen called on non-map Value")
}

func unsortedMapEntries(v Value) []MapEntry {
	if am, ok := v.impl.(*arrayMapImpl); ok {
		entries := make([]MapEntry, 0, len(am.kvs)/2)
		for i := 0; i < len(am.kvs); i += 2 {
			entries = append(entries, MapEntry{Key: am.kvs[i], Value: am.kvs[i+1]})
		}
		return entries
	}
	itr := v.mapPointer().items.Iterator()
	entries := make([]MapEntry, 0, v.MapLen())
	for !itr.Done() {
		key, value, _ := itr.Next()
		entries = append(entries, MapEntry{Key: key, Value: value})
	}
	return entries
}

func (v Value) MapEntries() []MapEntry {
	if isType[*dateImpl](v) {
		return dateEntries(v)
	}
	entries := unsortedMapEntries(v)
	sort.Slice(entries, func(i, j int) bool {
		return valueLess(entries[i].Key, entries[j].Key)
	})
	return entries
}

func (v Value) SetLen() int {
	if !isType[*setImpl](v) {
		panic("SetLen called on non-set Value")
	}
	return v.setPointer().items.Len()
}

func (v Value) SetValues() []Value {
	if !isType[*setImpl](v) {
		panic("SetValues called on non-set Value")
	}
	itr := v.setPointer().items.Iterator()
	values := make([]Value, 0, v.SetLen())
	for !itr.Done() {
		value, _ := itr.Next()
		values = append(values, value)
	}
	sort.Slice(values, func(i, j int) bool {
		return valueLess(values[i], values[j])
	})
	return values
}

func newMapValue(items *immutable.Map[Value, Value]) Value {
	return Value{impl: &mapImpl{data: &mapData{items: items}}}
}

func newSetValue(items immutable.Set[Value]) Value {
	return Value{impl: &setImpl{data: &setData{items: items}}}
}

func (v Value) mapPointer() *mapData {
	return asImpl[*mapImpl](v, "mapPointer called on non-map Value").data
}

func (v Value) setPointer() *setData {
	return asImpl[*setImpl](v, "setPointer called on non-set Value").data
}

type mapImpl struct {
	defaultImpl
	data *mapData
}

func (mapImpl) hashImpl(v Value) uint64 {
	return mapsHash(v)
}
func (mapImpl) eq(v Value, other Value) bool {
	return mapsEq(v, other)
}
func (mapImpl) typ(Value) Value { return typeMap }
func (mapImpl) String(v Value) string {
	return mapString(v)
}
func (mapImpl) assoc(v Value, entries ...Value) Value {
	return hashMapAssoc(v, entries...)
}
func (mapImpl) dissoc(v Value, keys ...Value) Value {
	items := v.mapPointer().items
	for _, key := range keys {
		items = items.Delete(key)
	}
	return newMapValue(items)
}
func (mapImpl) get(v Value, key Value, missing Value) Value {
	value, ok := v.mapPointer().items.Get(key)
	if !ok {
		return missing
	}
	return value
}
func (mapImpl) contains(v Value, key Value) bool {
	_, ok := v.mapPointer().items.Get(key)
	return ok
}
func (mapImpl) find(v Value, key Value) Value {
	value, ok := v.mapPointer().items.Get(key)
	return findEntry(key, value, ok)
}
func (mapImpl) keys(v Value) Value { return mapEntryKeys(v.MapEntries()) }
func (mapImpl) vals(v Value) Value { return mapEntryVals(v.MapEntries()) }
func (mapImpl) seqCursor(v Value) seqCursor {
	return mapEntriesSeqCursor(v.MapEntries())
}
func (mapImpl) toNative(v Value) any {
	return nativeMap(v.MapEntries())
}
func (mapImpl) toJSON(v Value) (any, error) {
	return mapToJSONValue(v)
}
func (mapImpl) exMessage(v Value) Value {
	if msg, ok := mapLookup(v, exMessageKey); ok {
		return msg
	}
	return NilValue
}
func (mapImpl) count(v Value) int      { return v.MapLen() }
func (mapImpl) isEmpty(v Value) bool   { return v.MapLen() == 0 }
func (mapImpl) notEmpty(v Value) Value { return notEmptyIf(v, v.MapLen() == 0) }
func (mapImpl) apply(v Value, args ...Value) Value {
	return lookupApply(v, args, "map")
}
func (mapImpl) conj(v Value, entries ...Value) Value {
	items := v.mapPointer().items
	for _, entry := range entries {
		if isMap(entry) {
			for _, mapEntry := range unsortedMapEntries(entry) {
				items = items.Set(mapEntry.Key, mapEntry.Value)
			}
			continue
		}
		key, value := conjMapEntry(entry)
		items = items.Set(key, value)
	}
	return newMapValue(items)
}

type setImpl struct {
	defaultImpl
	data *setData
}

func (setImpl) hashImpl(v Value) uint64 {
	h := hashSeedSet
	n := uint64(0)
	itr := v.setPointer().items.Iterator()
	for !itr.Done() {
		value, _ := itr.Next()
		h += Hash(value)
		n++
	}
	return mix(h, n)
}
func (setImpl) eq(v Value, other Value) bool {
	if !isType[*setImpl](other) || v.SetLen() != other.SetLen() {
		return false
	}
	itr := v.setPointer().items.Iterator()
	for !itr.Done() {
		value, _ := itr.Next()
		if !Contains(other, value) {
			return false
		}
	}
	return true
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
func (setImpl) get(v Value, key Value, missing Value) Value {
	if v.setPointer().items.Has(key) {
		return key
	}
	return missing
}
func (setImpl) contains(v Value, key Value) bool {
	return v.setPointer().items.Has(key)
}
func (setImpl) first(v Value) Value {
	values := v.SetValues()
	if len(values) == 0 {
		return NilValue
	}
	return values[0]
}
func (setImpl) rest(v Value) Value {
	values := v.SetValues()
	if len(values) <= 1 {
		return NewArray()
	}
	return NewArray(values[1:]...)
}
func (setImpl) next(v Value) Value {
	rest := Rest(v)
	return nextIfEmpty(rest, rest.ArrayLen() == 0)
}
func (setImpl) seqCursor(v Value) seqCursor {
	return arraySeqCursor(v.SetValues())
}
func (setImpl) toNative(v Value) any {
	return nativeSeq(v.SetValues())
}
func (setImpl) toJSON(v Value) (any, error) {
	return arrayToJSONValue(v.SetValues())
}
func (setImpl) count(v Value) int      { return v.SetLen() }
func (setImpl) isEmpty(v Value) bool   { return v.SetLen() == 0 }
func (setImpl) notEmpty(v Value) Value { return notEmptyIf(v, v.SetLen() == 0) }
func (setImpl) conj(v Value, entries ...Value) Value {
	items := v.setPointer().items
	for _, item := range entries {
		items = items.Add(item)
	}
	return newSetValue(items)
}
func (setImpl) apply(v Value, args ...Value) Value {
	if len(args) != 1 && len(args) != 2 {
		panic("set invocation expects value and optional default")
	}
	if Contains(v, args[0]) {
		return args[0]
	}
	if len(args) == 2 {
		return args[1]
	}
	return NilValue
}
