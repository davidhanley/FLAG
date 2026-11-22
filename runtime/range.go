package runtime

type rangeImpl struct {
	defaultImpl
	end    int64
	step   int64
	finite bool
}

func newRangeValue(start, end, step int64, finite bool) Value {
	return Value{d: packLong(start), impl: &rangeImpl{end: end, step: step, finite: finite}}
}

func (r *rangeImpl) start(v Value) int64 { return v.Long() }

func (r *rangeImpl) done(start int64) bool {
	if !r.finite {
		return false
	}
	if r.step > 0 {
		return start >= r.end
	}
	return start <= r.end
}

func (r *rangeImpl) remaining(start int64) int64 {
	if !r.finite {
		return -1
	}
	return rangeCount(start, r.end, r.step)
}

func (r *rangeImpl) hashImpl(v Value) uint64 {
	h := mix(hashSeedRange, uint64(r.start(v)))
	h = mix(h, uint64(r.end))
	h = mix(h, uint64(r.step))
	if r.finite {
		return mix(h, 1)
	}
	return mix(h, 0)
}

func (r *rangeImpl) eq(v Value, other Value) bool {
	o, ok := other.impl.(*rangeImpl)
	if !ok {
		return false
	}
	return r.start(v) == o.start(other) && r.end == o.end && r.step == o.step && r.finite == o.finite
}

func (rangeImpl) typ(Value) Value { return typeRange }

func (rangeImpl) String(Value) string { return "#<range>" }

func (r *rangeImpl) first(v Value) Value {
	start := r.start(v)
	if r.done(start) {
		return NilValue
	}
	return NewLong(start)
}

func (r *rangeImpl) rest(v Value) Value {
	start := r.start(v)
	if r.done(start) {
		return NewArray()
	}
	next := start + r.step
	if r.done(next) {
		return NewArray()
	}
	return newRangeValue(next, r.end, r.step, r.finite)
}

func (r *rangeImpl) next(v Value) Value {
	rest := r.rest(v)
	return nextIfEmpty(rest, IsEmpty(rest))
}

func (r *rangeImpl) last(v Value) Value {
	start := r.start(v)
	if !r.finite {
		cursor := r.seqCursor(v)
		last := NilValue
		for {
			next, ok := cursor.nextOrDone()
			if !ok {
				return last
			}
			last = next
		}
	}
	n := r.remaining(start)
	if n <= 0 {
		return NilValue
	}
	return NewLong(start + (n-1)*r.step)
}

func (r *rangeImpl) count(v Value) int {
	start := r.start(v)
	if !r.finite {
		n := 0
		cursor := r.seqCursor(v)
		for {
			_, ok := cursor.nextOrDone()
			if !ok {
				return n
			}
			n++
		}
	}
	return int(r.remaining(start))
}

func (r *rangeImpl) isEmpty(v Value) bool { return r.done(r.start(v)) }

func (r *rangeImpl) notEmpty(v Value) Value { return notEmptyIf(v, r.isEmpty(v)) }

func (rangeImpl) nth(Value, int, Value, bool) Value {
	panic("nth does not support list or lazy-list; use slow-nth")
}

func (r *rangeImpl) slowNth(v Value, i int, missing Value, hasMissing bool) Value {
	start := r.start(v)
	if r.finite {
		n := r.remaining(start)
		if int64(i) >= n {
			if hasMissing {
				return missing
			}
			panic("slow-nth index out of range")
		}
	}
	return NewLong(start + int64(i)*r.step)
}

func (r *rangeImpl) drop(v Value, n int) Value {
	start := r.start(v)
	if r.finite {
		cnt := r.remaining(start)
		if int64(n) >= cnt {
			return NewArray()
		}
	}
	return newRangeValue(start+int64(n)*r.step, r.end, r.step, r.finite)
}

func (rangeImpl) seqQ(Value) bool { return true }

func (r *rangeImpl) seqCursor(v Value) seqCursor {
	start := r.start(v)
	if !r.finite {
		return rangeSeqCursor(start, r.step, 0, false)
	}
	return rangeSeqCursor(start, r.step, int(r.remaining(start)), true)
}

func (r *rangeImpl) toNative(v Value) any {
	cursor := r.seqCursor(v)
	out := make([]any, 0)
	if remaining, ok := cursor.remainingKnown(); ok && remaining > 0 {
		out = make([]any, 0, remaining)
	}
	for {
		next, ok := cursor.nextOrDone()
		if !ok {
			return out
		}
		out = append(out, nativeAny(next))
	}
}

func (rangeImpl) toJSON(Value) (any, error) {
	return "#<range>", nil
}

func (r *rangeImpl) goSeq(v Value) []Value {
	cursor := r.seqCursor(v)
	out := make([]Value, 0)
	if remaining, ok := cursor.remainingKnown(); ok && remaining > 0 {
		out = make([]Value, 0, remaining)
	}
	for {
		next, ok := cursor.nextOrDone()
		if !ok {
			return out
		}
		out = append(out, next)
	}
}
