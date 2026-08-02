package runtime

import (
	"math/rand"
	"time"
)

var randSource = rand.New(rand.NewSource(time.Now().UnixNano()))

func RandInt(max Value) Value {
	if max.tag != TagLong && max.tag != TagBigInt {
		panic("rand-int expects an integer Value")
	}
	n := valueToBigInt(max)
	if n.Sign() <= 0 {
		panic("rand-int expects a positive integer")
	}
	return NewLong(randSource.Int63n(n.Int64()))
}
