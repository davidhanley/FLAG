package runtime

// arrayMapMax is Clojure PersistentArrayMap's pair limit (HASHTABLE_THRESHOLD/2).
const arrayMapMax = 8
const arrayMapSlots = arrayMapMax * 2

type arrayMapImpl struct {
	defaultImpl
	n   int
	kvs [arrayMapSlots]Value
}

func newArrayMap(kvs []Value) Value {
	if len(kvs) > arrayMapSlots {
		panic("newArrayMap overflow")
	}
	m := &arrayMapImpl{n: len(kvs)}
	copy(m.kvs[:], kvs)
	return Value{impl: m}
}

func (m *arrayMapImpl) pairs() []Value {
	return m.kvs[:m.n]
}

func isMap(v Value) bool {
	return isType[*arrayMapImpl](v) || isType[*mapImpl](v)
}

func arrayMapLookup(kvs []Value, key Value) (Value, bool) {
	i := arrayMapKeyIndex(kvs, key)
	if i < 0 {
		return Value{}, false
	}
	return kvs[i+1], true
}

func arrayMapKeysEq(k, key Value) bool {
	if _, ok := key.impl.(*symbolImpl); ok {
		return k.impl == key.impl
	}
	return valueSlotEq(k, key) || Eq(k, key)
}

func arrayMapKeyIndex(kvs []Value, key Value) int {
	for i := 0; i < len(kvs); i += 2 {
		if arrayMapKeysEq(kvs[i], key) {
			return i
		}
	}
	return -1
}

// arrayMapAssoc applies all key/value pairs on one new arrayMapImpl. If the
// result has more than arrayMapMax unique keys it becomes a HAMT.
func arrayMapAssoc(coll Value, entries []Value) Value {
	m := asImpl[*arrayMapImpl](coll, "arrayMapAssoc expects array-map")
	out := &arrayMapImpl{n: m.n}
	out.kvs = m.kvs
	kvs := out.kvs[:out.n]
	extra := []Value(nil)
	changed := false
	for i := 0; i < len(entries); i += 2 {
		key, val := entries[i], entries[i+1]
		src := kvs
		if extra != nil {
			src = extra
		}
		if j := arrayMapKeyIndex(src, key); j >= 0 {
			if Eq(src[j+1], val) {
				continue
			}
			src[j+1] = val
			changed = true
			continue
		}
		if extra == nil && out.n+2 <= arrayMapSlots {
			out.kvs[out.n] = key
			out.kvs[out.n+1] = val
			out.n += 2
			kvs = out.kvs[:out.n]
			changed = true
			continue
		}
		if extra == nil {
			extra = make([]Value, out.n, out.n+len(entries))
			copy(extra, out.kvs[:out.n])
		}
		extra = append(extra, key, val)
		changed = true
	}
	if !changed {
		return coll
	}
	if extra != nil {
		if len(extra)/2 > arrayMapMax {
			return hashMapFromKVs(extra)
		}
		return newArrayMap(extra)
	}
	return Value{impl: out}
}

func (m *arrayMapImpl) hashImpl(v Value) uint64 {
	return mapsHash(v)
}

func (m *arrayMapImpl) eq(v Value, other Value) bool {
	return mapsEq(v, other)
}

func (arrayMapImpl) typ(Value) Value { return typeMap }

func (arrayMapImpl) String(v Value) string {
	return mapString(v)
}

func (m *arrayMapImpl) assoc(v Value, entries ...Value) Value {
	if len(entries) < 2 || len(entries)%2 != 0 {
		panic("assoc expects collection and key/value pairs")
	}
	return arrayMapAssoc(v, entries)
}

func (m *arrayMapImpl) dissoc(v Value, keys ...Value) Value {
	out := &arrayMapImpl{}
	w := 0
	for i := 0; i < m.n; i += 2 {
		k := m.kvs[i]
		drop := false
		for _, key := range keys {
			if arrayMapKeysEq(k, key) {
				drop = true
				break
			}
		}
		if drop {
			continue
		}
		out.kvs[w] = k
		out.kvs[w+1] = m.kvs[i+1]
		w += 2
	}
	if w == m.n {
		return v
	}
	out.n = w
	return Value{impl: out}
}

func (m *arrayMapImpl) get(_ Value, key Value, missing Value) Value {
	if value, ok := arrayMapLookup(m.pairs(), key); ok {
		return value
	}
	return missing
}

func (m *arrayMapImpl) contains(_ Value, key Value) bool {
	_, ok := arrayMapLookup(m.pairs(), key)
	return ok
}

func (m *arrayMapImpl) find(_ Value, key Value) Value {
	value, ok := arrayMapLookup(m.pairs(), key)
	return findEntry(key, value, ok)
}

func (arrayMapImpl) keys(v Value) Value { return mapEntryKeys(v.MapEntries()) }
func (arrayMapImpl) vals(v Value) Value { return mapEntryVals(v.MapEntries()) }
func (arrayMapImpl) seqCursor(v Value) seqCursor {
	return mapEntriesSeqCursor(v.MapEntries())
}
func (arrayMapImpl) toNative(v Value) any {
	return nativeMap(v.MapEntries())
}
func (arrayMapImpl) toJSON(v Value) (any, error) {
	return mapToJSONValue(v)
}
func (m *arrayMapImpl) exMessage(v Value) Value {
	if msg, ok := arrayMapLookup(m.pairs(), exMessageKey); ok {
		return msg
	}
	return NilValue
}
func (m *arrayMapImpl) count(Value) int    { return m.n / 2 }
func (m *arrayMapImpl) isEmpty(Value) bool { return m.n == 0 }
func (m *arrayMapImpl) notEmpty(v Value) Value {
	return notEmptyIf(v, m.n == 0)
}
func (arrayMapImpl) apply(v Value, args ...Value) Value {
	return lookupApply(v, args, "map")
}
func (m *arrayMapImpl) conj(v Value, entries ...Value) Value {
	out := v
	for _, entry := range entries {
		if isMap(entry) {
			for _, mapEntry := range unsortedMapEntries(entry) {
				out = Assoc(out, mapEntry.Key, mapEntry.Value)
			}
			continue
		}
		key, value := conjMapEntry(entry)
		out = Assoc(out, key, value)
	}
	return out
}
