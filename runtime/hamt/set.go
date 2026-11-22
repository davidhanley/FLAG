package hamt

type empty = struct{}

// Set is an immutable hash set. Transient() builds a TSet.
type Set[K any] struct {
	m Map[K, empty]
}

// TSet is a single-threaded set builder. After Persistent every method panics.
type TSet[K any] struct {
	t *Transient[K, empty]
}

func NewSet[K any](hash func(K) uint64, eq func(K, K) bool) Set[K] {
	return Set[K]{m: NewMap[K, empty](hash, eq)}
}

func (s Set[K]) Len() int { return s.m.Len() }

func (s Set[K]) Has(k K) bool {
	_, ok := s.m.Get(k)
	return ok
}

func (s Set[K]) Add(k K) Set[K] {
	return Set[K]{m: s.m.Assoc(k, empty{})}
}

func (s Set[K]) Delete(k K) Set[K] {
	return Set[K]{m: s.m.Dissoc(k)}
}

func (s Set[K]) Range(f func(K) bool) {
	if f == nil {
		return
	}
	s.m.Range(func(k K, _ empty) bool { return f(k) })
}

func (s Set[K]) Transient() *TSet[K] {
	return &TSet[K]{t: s.m.Transient()}
}

func (t *TSet[K]) Len() int {
	return t.t.Len()
}

func (t *TSet[K]) Has(k K) bool {
	_, ok := t.t.Get(k)
	return ok
}

func (t *TSet[K]) Add(k K) *TSet[K] {
	t.t.Assoc(k, empty{})
	return t
}

func (t *TSet[K]) Delete(k K) *TSet[K] {
	t.t.Dissoc(k)
	return t
}

func (t *TSet[K]) Persistent() Set[K] {
	return Set[K]{m: t.t.Persistent()}
}
