package runtime

import "time"

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
