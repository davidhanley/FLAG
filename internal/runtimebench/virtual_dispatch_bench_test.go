package runtimebench

import (
	"testing"
	"unsafe"
)

type virtualFns struct {
	add       func(lhs, rhs virtualValue) virtualValue
	addLong   func(receiver virtualValue, lhs int64) virtualValue
	addDouble func(receiver virtualValue, lhs float64) virtualValue
}

type virtualValue struct {
	d   float64
	p   unsafe.Pointer
	fns *virtualFns
}

var (
	virtualLongFns        virtualFns
	virtualDoubleFns      virtualFns
	virtualScalarLongLHS  virtualValue
	virtualScalarLongRHS  virtualValue
	virtualScalarMixedLHS virtualValue
	virtualScalarMixedRHS virtualValue
	virtualLongArray      []virtualValue
	virtualMixedArray     []virtualValue
	virtualSink           virtualValue
)

func init() {
	virtualLongFns = virtualFns{
		add: func(lhs, rhs virtualValue) virtualValue {
			return rhs.fns.addLong(rhs, lhs.longValue())
		},
		addLong: func(receiver virtualValue, lhs int64) virtualValue {
			return newVirtualLong(lhs + receiver.longValue())
		},
		addDouble: func(receiver virtualValue, lhs float64) virtualValue {
			return newVirtualDouble(lhs + float64(receiver.longValue()))
		},
	}

	virtualDoubleFns = virtualFns{
		add: func(lhs, rhs virtualValue) virtualValue {
			return rhs.fns.addDouble(rhs, lhs.doubleValue())
		},
		addLong: func(receiver virtualValue, lhs int64) virtualValue {
			return newVirtualDouble(float64(lhs) + receiver.doubleValue())
		},
		addDouble: func(receiver virtualValue, lhs float64) virtualValue {
			return newVirtualDouble(lhs + receiver.doubleValue())
		},
	}

	virtualScalarLongLHS = newVirtualLong(41)
	virtualScalarLongRHS = newVirtualLong(1)
	virtualScalarMixedLHS = newVirtualLong(41)
	virtualScalarMixedRHS = newVirtualDouble(1.5)
	virtualLongArray = buildVirtualLongArray(1024)
	virtualMixedArray = buildVirtualMixedArray(1024)
}

func newVirtualLong(v int64) virtualValue {
	out := virtualValue{fns: &virtualLongFns}
	*(*int64)(unsafe.Pointer(&out.d)) = v
	return out
}

func newVirtualDouble(v float64) virtualValue {
	return virtualValue{d: v, fns: &virtualDoubleFns}
}

func (v virtualValue) add(rhs virtualValue) virtualValue {
	return v.fns.add(v, rhs)
}

func (v virtualValue) longValue() int64 {
	return *(*int64)(unsafe.Pointer(&v.d))
}

func (v virtualValue) doubleValue() float64 {
	return v.d
}

func buildVirtualLongArray(size int) []virtualValue {
	values := make([]virtualValue, size)
	for i := range values {
		values[i] = newVirtualLong(int64(i % 97))
	}
	return values
}

func buildVirtualMixedArray(size int) []virtualValue {
	values := make([]virtualValue, size)
	for i := range values {
		if i%2 == 0 {
			values[i] = newVirtualLong(int64(i % 97))
			continue
		}
		values[i] = newVirtualDouble(float64(i%97) + 0.5)
	}
	return values
}

func BenchmarkVirtualScalarLongAdd(b *testing.B) {
	lhs := virtualScalarLongLHS
	rhs := virtualScalarLongRHS
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		virtualSink = lhs.add(rhs)
	}
}

func BenchmarkVirtualScalarMixedAdd(b *testing.B) {
	lhs := virtualScalarMixedLHS
	rhs := virtualScalarMixedRHS
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		virtualSink = lhs.add(rhs)
	}
}

func BenchmarkVirtualArrayLongReduce(b *testing.B) {
	values := virtualLongArray
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		acc := newVirtualLong(0)
		for _, value := range values {
			acc = acc.add(value)
		}
		virtualSink = acc
	}
}

func BenchmarkVirtualArrayMixedReduce(b *testing.B) {
	values := virtualMixedArray
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		acc := newVirtualLong(0)
		for _, value := range values {
			acc = acc.add(value)
		}
		virtualSink = acc
	}
}
