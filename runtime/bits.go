package runtime

func requireBitLong(name string, v Value) int64 {
	switch v.tag {
	case TagLong:
		return v.Long()
	case TagBigInt:
		if v.BigInt().IsInt64() {
			return v.BigInt().Int64()
		}
		panic(name + " value out of range for long")
	default:
		panic(name + " expects an integer")
	}
}

func bitShiftCount(n int64) uint {
	return uint(uint64(n) & 63)
}

func bitMask(n int64) int64 {
	return int64(uint64(1) << bitShiftCount(n))
}

func foldBitOp(name string, op func(int64, int64) int64, values ...Value) Value {
	if len(values) < 2 {
		panic(name + " expects at least two arguments")
	}
	acc := requireBitLong(name, values[0])
	for _, v := range values[1:] {
		acc = op(acc, requireBitLong(name, v))
	}
	return NewLong(acc)
}

func BitAnd(values ...Value) Value {
	return foldBitOp("bit-and", func(a, b int64) int64 { return a & b }, values...)
}

func BitOr(values ...Value) Value {
	return foldBitOp("bit-or", func(a, b int64) int64 { return a | b }, values...)
}

func BitXor(values ...Value) Value {
	return foldBitOp("bit-xor", func(a, b int64) int64 { return a ^ b }, values...)
}

func BitNot(value Value) Value {
	return NewLong(^requireBitLong("bit-not", value))
}

func BitShiftLeft(x, n Value) Value {
	xv := requireBitLong("bit-shift-left", x)
	nv := requireBitLong("bit-shift-left", n)
	return NewLong(xv << bitShiftCount(nv))
}

func BitShiftRight(x, n Value) Value {
	xv := requireBitLong("bit-shift-right", x)
	nv := requireBitLong("bit-shift-right", n)
	return NewLong(xv >> bitShiftCount(nv))
}

func UnsignedBitShiftRight(x, n Value) Value {
	xv := uint64(requireBitLong("unsigned-bit-shift-right", x))
	nv := requireBitLong("unsigned-bit-shift-right", n)
	return NewLong(int64(xv >> bitShiftCount(nv)))
}

func BitTest(x, n Value) Value {
	xv := requireBitLong("bit-test", x)
	nv := requireBitLong("bit-test", n)
	return NewBool(uint64(xv)&uint64(bitMask(nv)) != 0)
}

func BitSet(x, n Value) Value {
	xv := requireBitLong("bit-set", x)
	nv := requireBitLong("bit-set", n)
	return NewLong(xv | bitMask(nv))
}

func BitClear(x, n Value) Value {
	xv := requireBitLong("bit-clear", x)
	nv := requireBitLong("bit-clear", n)
	return NewLong(xv &^ bitMask(nv))
}

func BitFlip(x, n Value) Value {
	xv := requireBitLong("bit-flip", x)
	nv := requireBitLong("bit-flip", n)
	return NewLong(xv ^ bitMask(nv))
}
