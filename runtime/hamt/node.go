package hamt

import "math/bits"

type node[K, V any] interface {
	assoc(e *session, hash func(K) uint64, eq func(K, K) bool, shift uint, h uint64, k K, v V, added *bool) node[K, V]
	without(e *session, hash func(K) uint64, eq func(K, K) bool, shift uint, h uint64, k K, removed *bool) node[K, V]
	find(eq func(K, K) bool, shift uint, h uint64, k K) (V, bool)
	rangeAll(f func(K, V) bool) bool
}

type slot[K, V any] struct {
	n node[K, V]
	k K
	v V
}

type bitmapNode[K, V any] struct {
	edit   *session
	bitmap uint32
	slots  []slot[K, V]
}

type collisionNode[K, V any] struct {
	edit  *session
	hash  uint64
	slots []slot[K, V]
}

func bitpos(h uint64, shift uint) uint32 {
	if shift >= 64 {
		return 1
	}
	return 1 << ((h >> shift) & mask)
}

func index(bitmap, bit uint32) int {
	return bits.OnesCount32(bitmap & (bit - 1))
}

func (n *bitmapNode[K, V]) editable(e *session) *bitmapNode[K, V] {
	if e != nil && e.alive && n.edit == e {
		return n
	}
	slots := make([]slot[K, V], len(n.slots), len(n.slots)+1)
	copy(slots, n.slots)
	return &bitmapNode[K, V]{edit: e, bitmap: n.bitmap, slots: slots}
}

func (n *collisionNode[K, V]) editable(e *session) *collisionNode[K, V] {
	if e != nil && e.alive && n.edit == e {
		return n
	}
	slots := make([]slot[K, V], len(n.slots), len(n.slots)+1)
	copy(slots, n.slots)
	return &collisionNode[K, V]{edit: e, hash: n.hash, slots: slots}
}

func insertSlot[K, V any](s []slot[K, V], idx int, sl slot[K, V]) []slot[K, V] {
	if cap(s) > len(s) {
		s = s[:len(s)+1]
		copy(s[idx+1:], s[idx:len(s)-1])
		s[idx] = sl
		return s
	}
	out := make([]slot[K, V], len(s)+1)
	copy(out, s[:idx])
	out[idx] = sl
	copy(out[idx+1:], s[idx:])
	return out
}

func removeSlot[K, V any](s []slot[K, V], idx int) []slot[K, V] {
	copy(s[idx:], s[idx+1:])
	return s[:len(s)-1]
}

func (n *bitmapNode[K, V]) assoc(e *session, hash func(K) uint64, eq func(K, K) bool, shift uint, h uint64, k K, v V, added *bool) node[K, V] {
	bit := bitpos(h, shift)
	idx := index(n.bitmap, bit)
	if n.bitmap&bit == 0 {
		*added = true
		out := n.editable(e)
		out.bitmap |= bit
		out.slots = insertSlot(out.slots, idx, slot[K, V]{k: k, v: v})
		return out
	}
	sl := n.slots[idx]
	if sl.n != nil {
		child := sl.n.assoc(e, hash, eq, shift+shiftBy, h, k, v, added)
		if child == sl.n {
			return n
		}
		out := n.editable(e)
		out.slots[idx].n = child
		return out
	}
	if eq(sl.k, k) {
		out := n.editable(e)
		out.slots[idx].v = v
		return out
	}
	*added = true
	child := mergeLeaves(e, shift+shiftBy, hash(sl.k), sl.k, sl.v, h, k, v)
	out := n.editable(e)
	out.slots[idx] = slot[K, V]{n: child}
	return out
}

func (n *bitmapNode[K, V]) without(e *session, hash func(K) uint64, eq func(K, K) bool, shift uint, h uint64, k K, removed *bool) node[K, V] {
	bit := bitpos(h, shift)
	if n.bitmap&bit == 0 {
		return n
	}
	idx := index(n.bitmap, bit)
	sl := n.slots[idx]
	if sl.n != nil {
		child := sl.n.without(e, hash, eq, shift+shiftBy, h, k, removed)
		if child == sl.n {
			return n
		}
		if child == nil {
			return n.drop(e, bit, idx)
		}
		out := n.editable(e)
		out.slots[idx].n = child
		return out
	}
	if !eq(sl.k, k) {
		return n
	}
	*removed = true
	return n.drop(e, bit, idx)
}

func (n *bitmapNode[K, V]) drop(e *session, bit uint32, idx int) node[K, V] {
	if len(n.slots) == 1 {
		return nil
	}
	out := n.editable(e)
	out.bitmap &^= bit
	out.slots = removeSlot(out.slots, idx)
	return out
}

func (n *bitmapNode[K, V]) find(eq func(K, K) bool, shift uint, h uint64, k K) (V, bool) {
	var zero V
	bit := bitpos(h, shift)
	if n.bitmap&bit == 0 {
		return zero, false
	}
	sl := n.slots[index(n.bitmap, bit)]
	if sl.n != nil {
		return sl.n.find(eq, shift+shiftBy, h, k)
	}
	if eq(sl.k, k) {
		return sl.v, true
	}
	return zero, false
}

func (n *bitmapNode[K, V]) rangeAll(f func(K, V) bool) bool {
	for _, sl := range n.slots {
		if sl.n != nil {
			if !sl.n.rangeAll(f) {
				return false
			}
			continue
		}
		if !f(sl.k, sl.v) {
			return false
		}
	}
	return true
}

func (n *collisionNode[K, V]) assoc(e *session, hash func(K) uint64, eq func(K, K) bool, shift uint, h uint64, k K, v V, added *bool) node[K, V] {
	if h != n.hash {
		if shift >= 64 {
			*added = true
			out := n.editable(e)
			out.slots = append(out.slots, slot[K, V]{k: k, v: v})
			return out
		}
		base := &bitmapNode[K, V]{
			edit:   e,
			bitmap: bitpos(n.hash, shift),
			slots:  []slot[K, V]{{n: n}},
		}
		return base.assoc(e, hash, eq, shift, h, k, v, added)
	}
	for i, sl := range n.slots {
		if eq(sl.k, k) {
			out := n.editable(e)
			out.slots[i].v = v
			return out
		}
	}
	*added = true
	out := n.editable(e)
	out.slots = append(out.slots, slot[K, V]{k: k, v: v})
	return out
}

func (n *collisionNode[K, V]) without(e *session, _ func(K) uint64, eq func(K, K) bool, _ uint, h uint64, k K, removed *bool) node[K, V] {
	if h != n.hash {
		return n
	}
	for i, sl := range n.slots {
		if !eq(sl.k, k) {
			continue
		}
		*removed = true
		if len(n.slots) == 1 {
			return nil
		}
		out := n.editable(e)
		out.slots = removeSlot(out.slots, i)
		return out
	}
	return n
}

func (n *collisionNode[K, V]) find(eq func(K, K) bool, _ uint, h uint64, k K) (V, bool) {
	var zero V
	if h != n.hash {
		return zero, false
	}
	for _, sl := range n.slots {
		if eq(sl.k, k) {
			return sl.v, true
		}
	}
	return zero, false
}

func (n *collisionNode[K, V]) rangeAll(f func(K, V) bool) bool {
	for _, sl := range n.slots {
		if !f(sl.k, sl.v) {
			return false
		}
	}
	return true
}

func mergeLeaves[K, V any](e *session, shift uint, h0 uint64, k0 K, v0 V, h1 uint64, k1 K, v1 V) node[K, V] {
	if h0 == h1 || shift >= 64 {
		return &collisionNode[K, V]{
			edit:  e,
			hash:  h0,
			slots: []slot[K, V]{{k: k0, v: v0}, {k: k1, v: v1}},
		}
	}
	b0 := bitpos(h0, shift)
	b1 := bitpos(h1, shift)
	if b0 == b1 {
		child := mergeLeaves(e, shift+shiftBy, h0, k0, v0, h1, k1, v1)
		return &bitmapNode[K, V]{edit: e, bitmap: b0, slots: []slot[K, V]{{n: child}}}
	}
	var slots []slot[K, V]
	if index(b0|b1, b0) == 0 {
		slots = []slot[K, V]{{k: k0, v: v0}, {k: k1, v: v1}}
	} else {
		slots = []slot[K, V]{{k: k1, v: v1}, {k: k0, v: v0}}
	}
	return &bitmapNode[K, V]{edit: e, bitmap: b0 | b1, slots: slots}
}
