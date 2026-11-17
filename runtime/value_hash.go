package runtime

import (
	"math"
	"math/big"
	"math/bits"
)

const (
	hashSeedNil     uint64 = 0x1111111111111111
	hashSeedNum     uint64 = 0x2222222222222222
	hashSeedBool    uint64 = 0x3333333333333333
	hashSeedString  uint64 = 0x4444444444444444
	hashSeedKeyword uint64 = 0x5555555555555555
	hashSeedSymbol  uint64 = 0x6666666666666666
	hashSeedList    uint64 = 0x7777777777777777
	hashSeedArray   uint64 = 0x8888888888888888
	hashSeedVector  uint64 = 0x9999999999999999
	hashSeedMap     uint64 = 0xaaaaaaaaaaaaaaaa
	hashSeedSet     uint64 = 0xbbbbbbbbbbbbbbbb
	hashSeedDate    uint64 = 0xcccccccccccccccc
	hashSeedFn      uint64 = 0xdddddddddddddddd
	hashSeedAtom    uint64 = 0xeeeeeeeeeeeeeeee
	hashSeedChannel uint64 = 0xf1f1f1f1f1f1f1f1
	hashSeedFile    uint64 = 0xf2f2f2f2f2f2f2f2
	hashSeedLazy    uint64 = 0xf3f3f3f3f3f3f3f3
	hashSeedRange   uint64 = 0xf6f6f6f6f6f6f6f6
	hashSeedRecord  uint64 = 0xf4f4f4f4f4f4f4f4
	hashSeedRecur   uint64 = 0xf5f5f5f5f5f5f5f5
	hashMixConst    uint64 = 0x9e3779b97f4a7c15
)

type hashCacher interface {
	hashCachePtr() *uint64
}

// Hash is the FLAG hash of v. Heap values cache it on impl after the first
// call. The low bit is always set, so 0 is reserved as “not yet hashed”.
func Hash(v Value) uint64 {
	if d, ok := v.impl.(hashCacher); ok {
		slot := d.hashCachePtr()
		if h := *slot; h != 0 {
			return h
		}
		h := v.impl.hashImpl(v) | 1
		*slot = h
		return h
	}
	return v.impl.hashImpl(v) | 1
}

func mix(h, x uint64) uint64 {
	x ^= bits.RotateLeft64(x, 25) ^ bits.RotateLeft64(x, 50)
	h ^= x
	h *= hashMixConst
	h = bits.RotateLeft64(h, 17)
	return h
}

func hashBytes(b []byte) uint64 {
	h := uint64(14695981039346656037)
	for _, c := range b {
		h ^= uint64(c)
		h *= 1099511628211
	}
	return h
}

func hashInt(n *big.Int) uint64 {
	if n == nil || n.Sign() == 0 {
		return mix(hashSeedNum, 0)
	}
	h := mix(hashSeedNum, hashBytes(n.Bytes()))
	if n.Sign() < 0 {
		h = mix(h, ^uint64(0))
	}
	return h
}

func hashRat(r *big.Rat) uint64 {
	if r == nil {
		return mix(hashSeedNum, 0)
	}
	if r.IsInt() {
		return hashInt(r.Num())
	}
	return mix(mix(hashSeedNum, hashInt(r.Num())), hashInt(r.Denom()))
}

func hashFloat(f float64) uint64 {
	if math.IsNaN(f) {
		return mix(hashSeedNum, 0x7ff8000000000000)
	}
	if r := new(big.Rat).SetFloat64(f); r != nil {
		return hashRat(r)
	}
	return mix(hashSeedNum, math.Float64bits(f))
}

func hashOrdered(seed uint64, values []Value) uint64 {
	h := seed
	for _, value := range values {
		h = mix(h, Hash(value))
	}
	return mix(h, uint64(len(values)))
}

func hashPointer(seed uint64, p uintptr) uint64 {
	return mix(seed, uint64(p))
}

func valuesEq(a, b []Value) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !Eq(a[i], b[i]) {
			return false
		}
	}
	return true
}

type valueHasher struct{}

func (valueHasher) Hash(key Value) uint32 {
	if s, ok := key.impl.(*symbolImpl); ok {
		h := s.hashCache
		if h == 0 {
			h = Hash(key)
		}
		return uint32(h) ^ uint32(h>>32)
	}
	h := Hash(key)
	return uint32(h) ^ uint32(h>>32)
}

func (valueHasher) Equal(a, b Value) bool {
	if a.impl == b.impl {
		return math.Float64bits(a.d) == math.Float64bits(b.d)
	}
	return Eq(a, b)
}

func valueLess(a, b Value) bool {
	return a.String() < b.String()
}
