package runtimebench

import (
	"testing"
	"unsafe"
)

type taggedTag uint8

const (
	taggedLongTag taggedTag = iota + 1
	taggedDoubleTag
)

type taggedValue struct {
	d   float64
	p   unsafe.Pointer
	tag taggedTag
}

var (
	taggedScalarLongLHS  = newTaggedLong(41)
	taggedScalarLongRHS  = newTaggedLong(1)
	taggedScalarMixedLHS = newTaggedLong(41)
	taggedScalarMixedRHS = newTaggedDouble(1.5)
	taggedLongArray      = buildTaggedLongArray(1024)
	taggedMixedArray     = buildTaggedMixedArray(1024)
	taggedSink           taggedValue
)

func newTaggedLong(v int64) taggedValue {
	out := taggedValue{tag: taggedLongTag}
	*(*int64)(unsafe.Pointer(&out.d)) = v
	return out
}

func newTaggedDouble(v float64) taggedValue {
	return taggedValue{d: v, tag: taggedDoubleTag}
}

func (v taggedValue) longValue() int64 {
	return *(*int64)(unsafe.Pointer(&v.d))
}

func (v taggedValue) doubleValue() float64 {
	return v.d
}

func addTagged(lhs, rhs taggedValue) taggedValue {
	switch lhs.tag {
	case taggedLongTag:
		left := lhs.longValue()
		switch rhs.tag {
		case taggedLongTag:
			return newTaggedLong(left + rhs.longValue())
		case taggedDoubleTag:
			return newTaggedDouble(float64(left) + rhs.doubleValue())
		default:
			panic("unknown tagged rhs")
		}
	case taggedDoubleTag:
		left := lhs.doubleValue()
		switch rhs.tag {
		case taggedLongTag:
			return newTaggedDouble(left + float64(rhs.longValue()))
		case taggedDoubleTag:
			return newTaggedDouble(left + rhs.doubleValue())
		default:
			panic("unknown tagged rhs")
		}
	default:
		panic("unknown tagged lhs")
	}
}

func buildTaggedLongArray(size int) []taggedValue {
	values := make([]taggedValue, size)
	for i := range values {
		values[i] = newTaggedLong(int64(i % 97))
	}
	return values
}

func buildTaggedMixedArray(size int) []taggedValue {
	values := make([]taggedValue, size)
	for i := range values {
		if i%2 == 0 {
			values[i] = newTaggedLong(int64(i % 97))
			continue
		}
		values[i] = newTaggedDouble(float64(i%97) + 0.5)
	}
	return values
}

func BenchmarkTagScalarLongAdd(b *testing.B) {
	lhs := taggedScalarLongLHS
	rhs := taggedScalarLongRHS
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		taggedSink = addTagged(lhs, rhs)
	}
}

func BenchmarkTagScalarMixedAdd(b *testing.B) {
	lhs := taggedScalarMixedLHS
	rhs := taggedScalarMixedRHS
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		taggedSink = addTagged(lhs, rhs)
	}
}

func BenchmarkTagArrayLongReduce(b *testing.B) {
	values := taggedLongArray
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		acc := newTaggedLong(0)
		for _, value := range values {
			acc = addTagged(acc, value)
		}
		taggedSink = acc
	}
}

func BenchmarkTagArrayMixedReduce(b *testing.B) {
	values := taggedMixedArray
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		acc := newTaggedLong(0)
		for _, value := range values {
			acc = addTagged(acc, value)
		}
		taggedSink = acc
	}
}
