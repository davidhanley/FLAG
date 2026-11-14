package runtime

import (
	"strconv"
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

// lazyListView is one position in a shared lazy list.
// init=false is the start (head not yet bound). init=true and node=nil is empty.
type lazyListView struct {
	state *lazyListState
	node  *lazyListNode
	init  bool
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
	return makeValue(TagLazyList, packLong(remaining), unsafe.Pointer(view))
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
	view := list.lazyListViewPointer()
	s := view.state
	s.mu.Lock()
	defer s.mu.Unlock()
	node, ok := s.realize(view)
	if !ok {
		return Value{}, false
	}
	return node.value, true
}

func lazyTail(list Value) Value {
	view := list.lazyListViewPointer()
	s := view.state
	s.mu.Lock()
	node, ok := s.realize(view)
	var next *lazyListNode
	if ok {
		if node.next == nil && !s.done {
			s.produceOne()
		}
		next = node.next
	}
	remaining := int64(-1)
	if hint := list.lazyListRemainingHint(); hint >= 0 {
		if hint > 0 {
			remaining = hint - 1
		} else {
			remaining = 0
		}
	}
	s.mu.Unlock()
	out := newLazyListValue(s, remaining, view.index+1)
	outView := out.lazyListViewPointer()
	outView.init = true
	outView.node = next
	return out
}

func (s *lazyListState) realize(view *lazyListView) (*lazyListNode, bool) {
	if view.init {
		if view.node == nil {
			return nil, false
		}
		return view.node, true
	}
	if s.head == nil && !s.produceOne() {
		view.init = true
		view.node = nil
		return nil, false
	}
	view.init = true
	view.node = s.head
	return view.node, true
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

type lazyListImpl struct{ defaultImpl }

func (lazyListImpl) hash(v Value) string {
	return "Z:" + strconv.FormatUint(uint64(uintptr(unsafe.Pointer(v.lazyListPointer()))), 16)
}
func (lazyListImpl) typ(Value) Value { return typeLazyList }
func (lazyListImpl) String(Value) string {
	return "#<lazy-list>"
}
func (lazyListImpl) first(v Value) Value {
	next, ok := lazyPeek(v)
	if !ok {
		return NilValue
	}
	return next
}
func (lazyListImpl) rest(v Value) Value {
	return lazyTail(v)
}
func (lazyListImpl) next(v Value) Value {
	rest := lazyTail(v)
	_, ok := lazyPeek(rest)
	return nextIfEmpty(rest, !ok)
}
func (lazyListImpl) last(v Value) Value {
	cursor := newSeqCursor(v)
	last := NilValue
	for {
		next, ok := cursor.nextOrDone()
		if !ok {
			return last
		}
		last = next
	}
}
func (lazyListImpl) count(v Value) int {
	if hint := v.lazyListRemainingHint(); hint >= 0 {
		return int(hint)
	}
	cursor := newSeqCursor(v)
	count := 0
	for {
		_, ok := cursor.nextOrDone()
		if !ok {
			break
		}
		count++
	}
	return count
}
func (lazyListImpl) isEmpty(v Value) bool {
	_, ok := lazyPeek(v)
	return !ok
}
func (lazyListImpl) notEmpty(v Value) Value {
	_, ok := lazyPeek(v)
	return notEmptyIf(v, !ok)
}
func (lazyListImpl) nth(Value, int, Value, bool) Value {
	panic("nth does not support list or lazy-list; use slow-nth")
}
func (lazyListImpl) slowNth(v Value, i int, missing Value, hasMissing bool) Value {
	return slowNthWalk(v, i, missing, hasMissing)
}
func (lazyListImpl) drop(v Value, n int) Value {
	out := v
	for i := 0; i < n; i++ {
		if got := First(out); got.tag == TagNil {
			return out
		}
		out = lazyTail(out)
	}
	return out
}
func (lazyListImpl) seqQ(Value) bool { return true }
func (lazyListImpl) seqCursor(v Value) seqCursor {
	remaining := v.lazyListRemainingHint()
	hasLength := remaining >= 0
	length := 0
	if hasLength {
		length = int(remaining)
	}
	return lazySeqCursor(v, length, hasLength)
}
