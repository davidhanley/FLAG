package runtime

import "unsafe"

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
	out := Value{p: unsafe.Pointer(head), tag: TagList}
	*(*int64)(unsafe.Pointer(&out.d)) = int64(length)
	return out
}

func (v Value) listPointer() *listNode {
	if v.tag != TagList {
		panic("listPointer called on non-list Value")
	}
	if v.p == nil {
		return nil
	}
	return (*listNode)(v.p)
}
