package hamt

import (
	"fmt"
	"testing"
)

func intHash(k int) uint64 { return uint64(k) * 0x9e3779b97f4a7c15 }
func intEq(a, b int) bool  { return a == b }

func collideHash(int) uint64 { return 0xdeadbeef }
func prefixHash(k int) uint64 {
	return uint64(k) << 10
}

func mustGet[K, V any](t *testing.T, m Map[K, V], k K) V {
	t.Helper()
	v, ok := m.Get(k)
	if !ok {
		t.Fatalf("missing key %#v", k)
	}
	return v
}

func TestMapAssocGetDissoc(t *testing.T) {
	m := NewMap[int, string](intHash, intEq)
	if m.Len() != 0 {
		t.Fatalf("empty len %d", m.Len())
	}
	if _, ok := m.Get(1); ok {
		t.Fatal("empty get")
	}
	m1 := m.Assoc(1, "a")
	m2 := m1.Assoc(2, "b")
	m3 := m2.Assoc(1, "A")
	if m.Len() != 0 || m1.Len() != 1 || m2.Len() != 2 || m3.Len() != 2 {
		t.Fatalf("lens %d %d %d %d", m.Len(), m1.Len(), m2.Len(), m3.Len())
	}
	if mustGet(t, m1, 1) != "a" || mustGet(t, m3, 1) != "A" || mustGet(t, m2, 2) != "b" {
		t.Fatal("get mismatch")
	}
	if _, ok := m1.Get(2); ok {
		t.Fatal("snapshot grew")
	}
	d := m2.Dissoc(1)
	if d.Len() != 1 || mustGet(t, d, 2) != "b" {
		t.Fatal("dissoc")
	}
	if m2.Len() != 2 {
		t.Fatal("dissoc mutated")
	}
	if m2.Dissoc(9).Len() != 2 {
		t.Fatal("missing dissoc")
	}
}

func TestMapManyKeys(t *testing.T) {
	m := NewMap[int, int](intHash, intEq)
	const n = 2000
	for i := 0; i < n; i++ {
		m = m.Assoc(i, i*i)
	}
	if m.Len() != n {
		t.Fatalf("len %d", m.Len())
	}
	for i := 0; i < n; i++ {
		if mustGet(t, m, i) != i*i {
			t.Fatalf("get %d", i)
		}
	}
	for i := 0; i < n; i += 2 {
		m = m.Dissoc(i)
	}
	if m.Len() != n/2 {
		t.Fatalf("after dissoc len %d", m.Len())
	}
	for i := 1; i < n; i += 2 {
		if mustGet(t, m, i) != i*i {
			t.Fatalf("odd %d", i)
		}
	}
	for i := 0; i < n; i += 2 {
		if _, ok := m.Get(i); ok {
			t.Fatalf("even still present %d", i)
		}
	}
}

func TestHashCollision(t *testing.T) {
	m := NewMap[int, string](collideHash, intEq)
	m = m.Assoc(1, "a").Assoc(2, "b").Assoc(3, "c").Assoc(2, "B")
	if m.Len() != 3 {
		t.Fatalf("len %d", m.Len())
	}
	if mustGet(t, m, 1) != "a" || mustGet(t, m, 2) != "B" || mustGet(t, m, 3) != "c" {
		t.Fatal("collision get")
	}
	m = m.Dissoc(2)
	if m.Len() != 2 {
		t.Fatal("collision dissoc len")
	}
	if _, ok := m.Get(2); ok {
		t.Fatal("2 still there")
	}
	if mustGet(t, m, 1) != "a" || mustGet(t, m, 3) != "c" {
		t.Fatal("collision after dissoc")
	}
}

func TestSharedPrefixHashes(t *testing.T) {
	m := NewMap[int, int](prefixHash, intEq)
	for i := 0; i < 64; i++ {
		m = m.Assoc(i, i)
	}
	if m.Len() != 64 {
		t.Fatalf("len %d", m.Len())
	}
	for i := 0; i < 64; i++ {
		if mustGet(t, m, i) != i {
			t.Fatalf("prefix %d", i)
		}
	}
	m = m.Dissoc(0).Dissoc(63)
	if m.Len() != 62 {
		t.Fatalf("len %d", m.Len())
	}
}

func TestRangeStops(t *testing.T) {
	m := NewMap[int, int](intHash, intEq).Assoc(1, 1).Assoc(2, 2).Assoc(3, 3)
	n := 0
	m.Range(func(k, v int) bool {
		n++
		return n < 2
	})
	if n != 2 {
		t.Fatalf("stopped at %d", n)
	}
	seen := map[int]int{}
	m.Range(func(k, v int) bool {
		seen[k] = v
		return true
	})
	if len(seen) != 3 || seen[1] != 1 || seen[2] != 2 || seen[3] != 3 {
		t.Fatalf("range %v", seen)
	}
}

func TestTransientBuild(t *testing.T) {
	base := NewMap[int, int](intHash, intEq).Assoc(1, 1)
	tr := base.Transient()
	for i := 2; i <= 100; i++ {
		tr.Assoc(i, i)
	}
	if tr.Len() != 100 {
		t.Fatalf("transient len %d", tr.Len())
	}
	if v, ok := tr.Get(50); !ok || v != 50 {
		t.Fatal("transient get")
	}
	out := tr.Persistent()
	if out.Len() != 100 || mustGet(t, out, 1) != 1 || mustGet(t, out, 100) != 100 {
		t.Fatal("persistent result")
	}
	if base.Len() != 1 || mustGet(t, base, 1) != 1 {
		t.Fatal("transient mutated base")
	}
	if _, ok := base.Get(2); ok {
		t.Fatal("base grew")
	}
}

func TestTransientDoesNotMutateSnapshot(t *testing.T) {
	m := NewMap[int, int](intHash, intEq)
	for i := 0; i < 40; i++ {
		m = m.Assoc(i, i)
	}
	snap := m
	tr := m.Transient()
	tr.Assoc(1000, 1)
	tr.Dissoc(0)
	p := tr.Persistent()
	if snap.Len() != 40 {
		t.Fatalf("snap len %d", snap.Len())
	}
	if mustGet(t, snap, 0) != 0 {
		t.Fatal("snap lost 0")
	}
	if _, ok := snap.Get(1000); ok {
		t.Fatal("snap gained 1000")
	}
	if p.Len() != 40 {
		t.Fatalf("p len %d", p.Len())
	}
	if _, ok := p.Get(0); ok {
		t.Fatal("p still has 0")
	}
	if mustGet(t, p, 1000) != 1 {
		t.Fatal("p missing 1000")
	}
}

func TestTransientAfterPersistentPanics(t *testing.T) {
	tr := NewMap[int, int](intHash, intEq).Transient()
	tr.Assoc(1, 1)
	_ = tr.Persistent()
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	tr.Assoc(2, 2)
}

func TestTransientCollisionAndDissoc(t *testing.T) {
	tr := NewMap[int, int](collideHash, intEq).Transient()
	for i := 0; i < 8; i++ {
		tr.Assoc(i, i)
	}
	tr.Dissoc(3)
	p := tr.Persistent()
	if p.Len() != 7 {
		t.Fatalf("len %d", p.Len())
	}
	if _, ok := p.Get(3); ok {
		t.Fatal("3 present")
	}
	if mustGet(t, p, 7) != 7 {
		t.Fatal("7 missing")
	}
}

func TestSetAddHasDelete(t *testing.T) {
	s := NewSet[int](intHash, intEq)
	s1 := s.Add(1).Add(2).Add(1)
	if s.Len() != 0 || s1.Len() != 2 {
		t.Fatalf("lens %d %d", s.Len(), s1.Len())
	}
	if !s1.Has(1) || !s1.Has(2) || s1.Has(3) {
		t.Fatal("has")
	}
	s2 := s1.Delete(1)
	if s2.Has(1) || !s2.Has(2) || s1.Len() != 2 {
		t.Fatal("delete")
	}
}

func TestSetTransient(t *testing.T) {
	base := NewSet[int](intHash, intEq).Add(0)
	tr := base.Transient()
	for i := 1; i < 50; i++ {
		tr.Add(i)
	}
	tr.Delete(0)
	out := tr.Persistent()
	if out.Len() != 49 || out.Has(0) || !out.Has(49) {
		t.Fatalf("set persist %#v len %d", out.Has(49), out.Len())
	}
	if base.Len() != 1 || !base.Has(0) {
		t.Fatal("set base mutated")
	}
	n := 0
	out.Range(func(int) bool { n++; return true })
	if n != 49 {
		t.Fatalf("range %d", n)
	}
}

func TestNewMapNilFuncsPanic(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	NewMap[int, int](nil, intEq)
}

func TestOverwriteSameKeyLen(t *testing.T) {
	m := NewMap[int, int](intHash, intEq).Assoc(1, 1).Assoc(1, 2)
	if m.Len() != 1 || mustGet(t, m, 1) != 2 {
		t.Fatal("overwrite")
	}
}

func TestDissocLast(t *testing.T) {
	m := NewMap[int, int](intHash, intEq).Assoc(1, 1).Dissoc(1)
	if m.Len() != 0 {
		t.Fatalf("len %d", m.Len())
	}
	if _, ok := m.Get(1); ok {
		t.Fatal("still there")
	}
}

func TestStringKeys(t *testing.T) {
	hash := func(s string) uint64 {
		var h uint64 = 14695981039346656037
		for i := 0; i < len(s); i++ {
			h ^= uint64(s[i])
			h *= 1099511628211
		}
		return h
	}
	eq := func(a, b string) bool { return a == b }
	tr := NewMap[string, int](hash, eq).Transient()
	for i := 0; i < 100; i++ {
		tr.Assoc(fmt.Sprintf("k%d", i), i)
	}
	m := tr.Persistent()
	if mustGet(t, m, "k42") != 42 {
		t.Fatal("string key")
	}
}
