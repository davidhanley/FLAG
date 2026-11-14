package runtime

import (
	"fmt"
	"time"
	"unsafe"
)

type DateObject struct {
	Time time.Time
}

func NewDate(t time.Time) Value {
	return makeValue(TagDate, 0, unsafe.Pointer(&DateObject{Time: normalizeDateTime(t)}))
}

func NewDateFromYMD(year, month, day int) Value {
	return NewDate(time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC))
}

func (v Value) DateObject() *DateObject {
	if v.tag != TagDate {
		panic("DateObject called on non-date Value")
	}
	if v.p == nil {
		panic("date Value does not contain date pointer")
	}
	return (*DateObject)(v.p)
}

func (v Value) DateTime() time.Time {
	return v.DateObject().Time
}

func (v Value) DateYear() int {
	return v.DateTime().UTC().Year()
}

func (v Value) DateMonth() int {
	return int(v.DateTime().UTC().Month())
}

func (v Value) DateDay() int {
	return v.DateTime().UTC().Day()
}

func dateEntries(v Value) []MapEntry {
	return []MapEntry{
		{Key: NewKeyword("year"), Value: NewLong(int64(v.DateYear()))},
		{Key: NewKeyword("month"), Value: NewLong(int64(v.DateMonth()))},
		{Key: NewKeyword("day"), Value: NewLong(int64(v.DateDay()))},
	}
}

func dateFieldValue(v Value, key Value) (Value, bool) {
	if key.tag != TagSymbol || !key.SymbolObject().IsKeyword {
		return Value{}, false
	}
	switch key.SymbolObject().Name {
	case "year":
		return NewLong(int64(v.DateYear())), true
	case "month":
		return NewLong(int64(v.DateMonth())), true
	case "day":
		return NewLong(int64(v.DateDay())), true
	default:
		return Value{}, false
	}
}

func normalizeDateTime(t time.Time) time.Time {
	return time.Date(t.UTC().Year(), t.UTC().Month(), t.UTC().Day(), 0, 0, 0, 0, time.UTC)
}

func dateString(v Value) string {
	entries := dateEntries(v)
	return fmt.Sprintf("{%s %s %s %s %s %s}",
		entries[0].Key.String(), entries[0].Value.String(),
		entries[1].Key.String(), entries[1].Value.String(),
		entries[2].Key.String(), entries[2].Value.String(),
	)
}

type dateImpl struct{ defaultImpl }

func (dateImpl) hash(v Value) string {
	return "D:" + v.DateTime().UTC().Format("2006-01-02")
}
func (dateImpl) typ(Value) Value { return typeDate }
func (dateImpl) String(v Value) string {
	return dateString(v)
}
func (dateImpl) get(v Value, key Value, missing Value) Value {
	if value, ok := dateFieldValue(v, key); ok {
		return value
	}
	return missing
}
func (dateImpl) contains(v Value, key Value) bool {
	_, ok := dateFieldValue(v, key)
	return ok
}
func (dateImpl) find(v Value, key Value) Value {
	value, ok := dateFieldValue(v, key)
	return findEntry(key, value, ok)
}
func (dateImpl) keys(v Value) Value     { return mapEntryKeys(dateEntries(v)) }
func (dateImpl) vals(v Value) Value     { return mapEntryVals(dateEntries(v)) }
func (dateImpl) count(v Value) int      { return len(dateEntries(v)) }
func (dateImpl) isEmpty(v Value) bool   { return len(dateEntries(v)) == 0 }
func (dateImpl) notEmpty(v Value) Value { return notEmptyIf(v, len(dateEntries(v)) == 0) }
