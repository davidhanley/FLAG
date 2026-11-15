package runtime

import (
	"strings"
)

type listNode struct {
	value Value
	next  *listNode
}

func NewList(values ...Value) Value {
	var head *listNode
	for i := len(values) - 1; i >= 0; i-- {
		head = &listNode{
			value: values[i],
			next:  head,
		}
	}
	return newListValue(head, len(values))
}

func ListCons(listValue Value, item Value) Value {
	if listValue.tag != TagList {
		panic("ListCons expects list Value")
	}
	head := &listNode{
		value: item,
		next:  listValue.listPointer(),
	}
	return newListValue(head, listValue.ListLen()+1)
}

func ListRest(listValue Value) Value {
	if listValue.tag != TagList {
		panic("ListRest expects list Value")
	}
	head := listValue.listPointer()
	if head == nil {
		return newListValue(nil, 0)
	}
	return newListValue(head.next, listValue.ListLen()-1)
}

func (v Value) ListLen() int {
	if v.tag != TagList {
		panic("ListLen called on non-list Value")
	}
	return int(v.Long())
}

func (v Value) ListValues() []Value {
	if v.tag != TagList {
		panic("ListValues called on non-list Value")
	}
	head := v.listPointer()
	out := make([]Value, 0, v.ListLen())
	for node := head; node != nil; node = node.next {
		out = append(out, node.value)
	}
	return out
}

func ListAppend(listValue Value, item Value) Value {
	if listValue.tag != TagList {
		panic("ListAppend expects list Value")
	}
	values := append(listValue.ListValues(), item)
	return NewList(values...)
}

func newListValue(head *listNode, length int) Value {
	return makeValue(TagList, packLong(int64(length)), &listImpl{head: head})
}

func (v Value) listPointer() *listNode {
	return asImpl[*listImpl](v, "listPointer called on non-list Value").head
}

type listImpl struct {
	defaultImpl
	head *listNode
}

func (listImpl) hash(v Value) string {
	values := v.ListValues()
	parts := make([]string, 0, len(values))
	for _, value := range values {
		parts = append(parts, valueIdentity(value))
	}
	return "T:(" + strings.Join(parts, ",") + ")"
}
func (listImpl) typ(Value) Value { return typeList }
func (listImpl) String(v Value) string {
	values := v.ListValues()
	var out strings.Builder
	out.WriteByte('(')
	for i, value := range values {
		if i > 0 {
			out.WriteByte(' ')
		}
		out.WriteString(value.String())
	}
	out.WriteByte(')')
	return out.String()
}
func (listImpl) first(v Value) Value {
	if v.ListLen() == 0 {
		return NilValue
	}
	return v.listPointer().value
}
func (listImpl) rest(v Value) Value {
	return ListRest(v)
}
func (listImpl) next(v Value) Value {
	rest := ListRest(v)
	return nextIfEmpty(rest, rest.ListLen() == 0)
}
func (listImpl) last(v Value) Value {
	if v.ListLen() == 0 {
		return NilValue
	}
	node := v.listPointer()
	for node.next != nil {
		node = node.next
	}
	return node.value
}
func (listImpl) conj(v Value, entries ...Value) Value {
	out := v
	for _, item := range entries {
		out = ListCons(out, item)
	}
	return out
}
func (listImpl) count(v Value) int      { return v.ListLen() }
func (listImpl) isEmpty(v Value) bool   { return v.ListLen() == 0 }
func (listImpl) notEmpty(v Value) Value { return notEmptyIf(v, v.ListLen() == 0) }
func (listImpl) nth(Value, int, Value, bool) Value {
	panic("nth does not support list or lazy-list; use slow-nth")
}
func (listImpl) slowNth(v Value, i int, missing Value, hasMissing bool) Value {
	return slowNthWalk(v, i, missing, hasMissing)
}
func (listImpl) drop(v Value, n int) Value {
	out := v
	for i := 0; i < n && out.ListLen() > 0; i++ {
		out = ListRest(out)
	}
	return out
}
func (listImpl) cons(v Value, item Value) Value {
	return ListCons(v, item)
}
func (listImpl) seqQ(Value) bool { return true }
func (listImpl) seqCursor(v Value) seqCursor {
	return listSeqCursor(v.listPointer(), v.ListLen())
}
func (listImpl) toNative(v Value) any {
	return nativeSeq(v.ListValues())
}
func (listImpl) toJSON(v Value) (any, error) {
	return arrayToJSONValue(v.ListValues())
}
func (listImpl) goSeq(v Value) []Value {
	return v.ListValues()
}
