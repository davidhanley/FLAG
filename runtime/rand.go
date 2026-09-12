package runtime

import (
	"math/rand"
	"sync"
	"time"
)

var (
	randMu     sync.Mutex
	randSource = rand.New(rand.NewSource(time.Now().UnixNano()))
)

func seedRand(seed int64) {
	randMu.Lock()
	defer randMu.Unlock()
	randSource = rand.New(rand.NewSource(seed))
}

func RandInt(max Value) Value {
	if max.tag != TagLong && max.tag != TagBigInt {
		panic("rand-int expects an integer Value")
	}
	n := valueToBigInt(max)
	if n.Sign() <= 0 {
		panic("rand-int expects a positive integer")
	}
	randMu.Lock()
	out := randSource.Int63n(n.Int64())
	randMu.Unlock()
	return NewLong(out)
}

func Rand(args ...Value) Value {
	if len(args) > 1 {
		panic("rand expects 0 or 1 arguments")
	}
	n := 1.0
	if len(args) == 1 {
		if !isNumericTag(args[0].tag) {
			panic("rand expects a numeric Value")
		}
		n = numericToFloat64(args[0])
	}
	randMu.Lock()
	out := randSource.Float64() * n
	randMu.Unlock()
	return NewDouble(out)
}

func RandNth(coll Value) Value {
	items := Vec(coll)
	n := items.ArrayLen()
	if n == 0 {
		panic("rand-nth on empty collection")
	}
	randMu.Lock()
	i := randSource.Intn(n)
	randMu.Unlock()
	return ArrayGet(items, i)
}

func Shuffle(coll Value) Value {
	items := Vec(coll).ArrayValues()
	randMu.Lock()
	randSource.Shuffle(len(items), func(i, j int) {
		items[i], items[j] = items[j], items[i]
	})
	randMu.Unlock()
	return NewArray(items...)
}
