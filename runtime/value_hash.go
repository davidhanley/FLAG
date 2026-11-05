package runtime

import (
	"hash/fnv"
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
	return v.impl.hash(v)
}
