package runtime

import (
	"sync"
	"unsafe"
)

type lazyNextFn func() (Value, bool)

type lazyListNode struct {
	value Value
	next  *lazyListNode
}

type lazyListState struct {
	mu   sync.Mutex
	next lazyNextFn
	head *lazyListNode
	tail *lazyListNode
	done bool
}

type lazyListView struct {
	state *lazyListState
	index int
}

func NewLazyList(next lazyNextFn) Value {
	if next == nil {
		panic("NewLazyList expects non-nil producer")
	}
	return newLazyListValue(&lazyListState{next: next}, -1, 0)
}

func newLazyListValue(state *lazyListState, remaining int64, index int) Value {
	if state == nil {
		panic("lazy list expects non-nil state")
	}
	view := &lazyListView{state: state, index: index}
	out := Value{p: unsafe.Pointer(view), tag: TagLazyList}
	*(*int64)(unsafe.Pointer(&out.d)) = remaining
	return out
}

func (v Value) lazyListViewPointer() *lazyListView {
	if v.tag != TagLazyList {
		panic("lazyListViewPointer called on non-lazy-list Value")
	}
	if v.p == nil {
		panic("lazy list Value does not contain view pointer")
	}
	return (*lazyListView)(v.p)
}

func (v Value) lazyListPointer() *lazyListState {
	return v.lazyListViewPointer().state
}

func (v Value) lazyListIndex() int {
	return v.lazyListViewPointer().index
}

func (v Value) lazyListRemainingHint() int64 {
	if v.tag != TagLazyList {
		panic("lazyListRemainingHint called on non-lazy-list Value")
	}
	return v.Long()
}

func lazyPeek(list Value) (Value, bool) {
	return list.lazyListPointer().peekAt(list.lazyListIndex())
}

func lazyTail(list Value) Value {
	remaining := int64(-1)
	if hint := list.lazyListRemainingHint(); hint >= 0 {
		if hint > 0 {
			remaining = hint - 1
		} else {
			remaining = 0
		}
	}
	return newLazyListValue(list.lazyListPointer(), remaining, list.lazyListIndex()+1)
}

func (s *lazyListState) peekAt(index int) (Value, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	node, ok := s.ensureNodeAt(index)
	if !ok {
		return Value{}, false
	}
	return node.value, true
}

func (s *lazyListState) ensureNodeAt(index int) (*lazyListNode, bool) {
	if index < 0 {
		panic("lazy list index must be non-negative")
	}
	for s.head == nil && !s.done {
		if !s.produceOne() {
			break
		}
	}
	if s.head == nil {
		return nil, false
	}
	node := s.head
	for i := 0; i < index; i++ {
		for node.next == nil && !s.done {
			if !s.produceOne() {
				break
			}
		}
		node = node.next
		if node == nil {
			return nil, false
		}
	}
	return node, true
}

func (s *lazyListState) produceOne() bool {
	if s.done {
		return false
	}
	next, ok := s.next()
	if !ok {
		s.done = true
		return false
	}
	node := &lazyListNode{value: next}
	if s.head == nil {
		s.head = node
		s.tail = node
		return true
	}
	s.tail.next = node
	s.tail = node
	return true
}
