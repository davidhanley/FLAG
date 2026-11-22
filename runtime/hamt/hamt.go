// Package hamt is a generic 32-way bitmap HAMT with persistent and transient
// maps. Sets are maps to struct{}.
//
// A transient copies a node the first time that session touches it, stamps it
// dirty, and updates in place after that. Persistent() flips the session off so
// those nodes freeze without a walk.
package hamt

const (
	fanout  = 32
	shiftBy = 5
	mask    = fanout - 1
)

type session struct {
	alive bool
}

// Map is an immutable hash map unless obtained via Transient.
type Map[K, V any] struct {
	hash func(K) uint64
	eq   func(K, K) bool
	root node[K, V]
	size int
}

// Transient is a single-threaded builder. Do not share it. After Persistent
// every method panics.
type Transient[K, V any] struct {
	hash func(K) uint64
	eq   func(K, K) bool
	root node[K, V]
	size int
	sess *session
}

// NewMap is empty. hash and eq must be non-nil and consistent: equal keys
// have equal hashes.
func NewMap[K, V any](hash func(K) uint64, eq func(K, K) bool) Map[K, V] {
	if hash == nil || eq == nil {
		panic("hamt.NewMap requires hash and eq")
	}
	return Map[K, V]{hash: hash, eq: eq}
}

func (m Map[K, V]) Len() int { return m.size }

func (m Map[K, V]) Get(k K) (V, bool) {
	var zero V
	if m.root == nil {
		return zero, false
	}
	return m.root.find(m.eq, 0, m.hash(k), k)
}

func (m Map[K, V]) Assoc(k K, v V) Map[K, V] {
	added := false
	h := m.hash(k)
	var root node[K, V]
	if m.root == nil {
		root = newLeaf[K, V](nil, h, k, v)
		added = true
	} else {
		root = m.root.assoc(nil, m.hash, m.eq, 0, h, k, v, &added)
	}
	if root == m.root && !added {
		return m
	}
	size := m.size
	if added {
		size++
	}
	return Map[K, V]{hash: m.hash, eq: m.eq, root: root, size: size}
}

func (m Map[K, V]) Dissoc(k K) Map[K, V] {
	if m.root == nil {
		return m
	}
	removed := false
	root := m.root.without(nil, m.hash, m.eq, 0, m.hash(k), k, &removed)
	if !removed {
		return m
	}
	return Map[K, V]{hash: m.hash, eq: m.eq, root: root, size: m.size - 1}
}

// Range calls f for each pair. Order is hash-trie order, not insertion.
// Stop if f returns false.
func (m Map[K, V]) Range(f func(K, V) bool) {
	if m.root == nil || f == nil {
		return
	}
	m.root.rangeAll(f)
}

func (m Map[K, V]) Transient() *Transient[K, V] {
	return &Transient[K, V]{
		hash: m.hash,
		eq:   m.eq,
		root: m.root,
		size: m.size,
		sess: &session{alive: true},
	}
}

func (t *Transient[K, V]) ensure() {
	if t == nil || t.sess == nil || !t.sess.alive {
		panic("transient used after persistent!")
	}
}

func (t *Transient[K, V]) Len() int {
	t.ensure()
	return t.size
}

func (t *Transient[K, V]) Get(k K) (V, bool) {
	t.ensure()
	var zero V
	if t.root == nil {
		return zero, false
	}
	return t.root.find(t.eq, 0, t.hash(k), k)
}

func (t *Transient[K, V]) Assoc(k K, v V) *Transient[K, V] {
	t.ensure()
	added := false
	h := t.hash(k)
	if t.root == nil {
		t.root = newLeaf[K, V](t.sess, h, k, v)
		t.size = 1
		return t
	}
	t.root = t.root.assoc(t.sess, t.hash, t.eq, 0, h, k, v, &added)
	if added {
		t.size++
	}
	return t
}

func (t *Transient[K, V]) Dissoc(k K) *Transient[K, V] {
	t.ensure()
	if t.root == nil {
		return t
	}
	removed := false
	t.root = t.root.without(t.sess, t.hash, t.eq, 0, t.hash(k), k, &removed)
	if removed {
		t.size--
	}
	return t
}

func (t *Transient[K, V]) Persistent() Map[K, V] {
	t.ensure()
	t.sess.alive = false
	return Map[K, V]{hash: t.hash, eq: t.eq, root: t.root, size: t.size}
}

func newLeaf[K, V any](e *session, h uint64, k K, v V) node[K, V] {
	bit := bitpos(h, 0)
	return &bitmapNode[K, V]{
		edit:   e,
		bitmap: bit,
		slots:  []slot[K, V]{{k: k, v: v}},
	}
}
