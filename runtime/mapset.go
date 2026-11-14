package runtime

import (
	"sort"
	"unsafe"

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
		if set.tag == TagNil {
			continue
		}
		if set.tag != TagSet {
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
		if set.tag != TagSet && set.tag != TagNil {
			panic("intersection expects set arguments")
		}
		if set.tag == TagNil {
			return NewSet()
		}
	}
	base := sets[0]
	if base.tag == TagNil {
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
	if first.tag == TagNil {
		return NewSet()
	}
	if first.tag != TagSet {
		panic("difference expects set arguments")
	}
	items := first.setPointer().items
	for _, set := range sets[1:] {
		if set.tag == TagNil {
			continue
		}
		if set.tag != TagSet {
			panic("difference expects set arguments")
		}
		for _, value := range set.SetValues() {
			items = items.Delete(value)
		}
	}
	return newSetValue(items)
}

func Subset(set1 Value, set2 Value) bool {
	if set1.tag == TagNil {
		return true
	}
	if set1.tag != TagSet {
		panic("subset? expects set arguments")
	}
	if set2.tag == TagNil {
		return set1.SetLen() == 0
	}
	if set2.tag != TagSet {
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
	if set1.tag == TagNil || set2.tag == TagNil {
		return true
	}
	if set1.tag != TagSet || set2.tag != TagSet {
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
	if coll.tag == TagNil {
		return NilValue
	}
	if coll.tag != TagMap {
		panic("rename-keys expects map and key-map")
	}
	if keyMap.tag == TagNil {
		return coll
	}
	if keyMap.tag != TagMap {
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
	if coll.tag == TagNil {
		return NilValue
	}
	if coll.tag != TagMap {
		panic("map-invert expects map")
	}
	out := NewMap()
	for _, entry := range coll.MapEntries() {
		out = MapAssoc(out, entry.Value, entry.Key)
	}
	return out
}

func SetSelect(pred Value, set Value) Value {
	if set.tag == TagNil {
		return NewSet()
	}
	if set.tag != TagSet {
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
	if rel.tag == TagNil {
		return NewSet()
	}
	if rel.tag != TagSet {
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
	if rel.tag == TagNil {
		return NewSet()
	}
	if rel.tag != TagSet {
		panic("rename expects relation set and key-map")
	}
	items := immutable.NewSet[Value](valueHasher{})
	for _, row := range rel.SetValues() {
		items = items.Add(RenameKeys(row, keyMap))
	}
	return newSetValue(items)
}

func projectRow(row Value, keys []Value) Value {
	if row.tag != TagMap && row.tag != TagDate && row.tag != TagRecord {
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
	return coll.impl.get(coll, key, missing)
}

func associativeIndex(key Value) (int, bool) {
	switch key.tag {
	case TagLong:
		if key.Long() < 0 {
			return 0, false
		}
		return int(key.Long()), true
	case TagBigInt:
		bi := key.BigInt()
		if bi.Sign() < 0 || bi.BitLen() > 62 {
			return 0, false
		}
		return int(bi.Int64()), true
	default:
		return 0, false
	}
}

func Contains(coll Value, key Value) bool {
	return coll.impl.contains(coll, key)
}

func MapDissoc(coll Value, keys ...Value) Value {
	return Dissoc(coll, keys...)
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
	return makeValue(TagMap, 0, unsafe.Pointer(&mapData{items: items}))
}

func newSetValue(items immutable.Set[Value]) Value {
	return makeValue(TagSet, 0, unsafe.Pointer(&setData{items: items}))
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
func (setImpl) get(v Value, key Value, missing Value) Value {
	if v.setPointer().items.Has(key) {
		return key
	}
	return missing
}
func (setImpl) contains(v Value, key Value) bool {
	return v.setPointer().items.Has(key)
}
