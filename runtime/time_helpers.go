package runtime

import "time"

// Sleep pauses the current goroutine for the given number of milliseconds
// (truncated toward zero if a float is passed). Returns nil.
func Sleep(ms Value) Value {
	if !isNumericTag(ms.tag) {
		panic("sleep expects milliseconds as a number")
	}
	n := ms.Long()
	if n < 0 {
		panic("sleep expects non-negative milliseconds")
	}
	time.Sleep(time.Duration(n) * time.Millisecond)
	return NilValue()
}

func TimeNow() Value {
	return NewDate(time.Now().UTC())
}

func TimeYears(v Value) Value {
	if !isNumericTag(v.tag) {
		panic("t/years expects a numeric Value")
	}
	return NewLong(int64(v.Long()))
}

func TimeMinus(left Value, right Value) Value {
	if left.tag != TagDate {
		panic("t/minus expects a date as the first argument")
	}
	if !isNumericTag(right.tag) {
		panic("t/minus expects a numeric Value as the second argument")
	}
	return NewDate(left.DateTime().AddDate(-int(right.Long()), 0, 0))
}

func TimeAfter(left Value, right Value) Value {
	if left.tag != TagDate || right.tag != TagDate {
		panic("t/after? expects two dates")
	}
	return NewBool(left.DateTime().After(right.DateTime()))
}
