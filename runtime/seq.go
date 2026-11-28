package runtime

import (
	goruntime "runtime"
	"sort"
	"sync"
	"unicode/utf8"
)

func Map(fn Value, seqs ...Value) Value {
	if len(seqs) == 0 {
		panic("map expects at least one sequence")
	}

	if allLazySeqs(seqs) {
		return lazyMap(fn, seqs...)
	}

	cursors := make([]seqCursor, len(seqs))
	shortestKnown := -1
	allKnown := true
	for i, seq := range seqs {
		cursor := newSeqCursor(seq)
		cursors[i] = cursor
		if remaining, ok := cursor.remainingKnown(); ok {
			if shortestKnown == -1 || remaining < shortestKnown {
				shortestKnown = remaining
			}
		} else {
			allKnown = false
		}
	}
	if shortestKnown == 0 {
		return NewArray()
	}

	initialCap := 0
	switch {
	case allKnown:
		initialCap = shortestKnown
	case shortestKnown > 0:
		initialCap = shortestKnown
	}
	result := make([]Value, 0, initialCap)
	callArgs := make([]Value, len(seqs))
	for {
		for s := range cursors {
			next, ok := cursors[s].nextOrDone()
			if !ok {
				return NewArray(result...)
			}
			callArgs[s] = next
		}
		result = append(result, Call(fn, callArgs...))
	}
}

func Apply(fn Value, args ...Value) Value {
	return Call(fn, spreadApplyArgs(args)...)
}

func ApplyFn(fn func(...Value) Value, args ...Value) Value {
	if fn == nil {
		panic("apply expects a function")
	}
	return fn(spreadApplyArgs(args)...)
}

func spreadApplyArgs(args []Value) []Value {
	if len(args) == 0 {
		panic("apply expects function and sequence arguments")
	}
	fixed := args[:len(args)-1]
	tail := args[len(args)-1]
	cursor := newSeqCursor(tail)
	callArgs := make([]Value, 0, len(fixed))
	callArgs = append(callArgs, fixed...)
	for {
		next, ok := cursor.nextOrDone()
		if !ok {
			break
		}
		callArgs = append(callArgs, next)
	}
	return callArgs
}

func Concat(seqs ...Value) Value {
	seqCopy := append([]Value(nil), seqs...)
	hint := int64(0)
	for _, seq := range seqCopy {
		cursor := newSeqCursor(seq)
		remaining, ok := cursor.remainingKnown()
		if !ok {
			hint = -1
			break
		}
		hint += int64(remaining)
	}

	index := 0
	var (
		cursor     seqCursor
		haveCursor bool
	)
	state := &lazyListState{
		next: func() (Value, bool) {
			for index < len(seqCopy) {
				if !haveCursor {
					cursor = newSeqCursor(seqCopy[index])
					haveCursor = true
				}
				next, ok := cursor.nextOrDone()
				if ok {
					return next, true
				}
				index++
				haveCursor = false
			}
			return Value{}, false
		},
	}
	return newLazyListValue(state, hint, 0)
}

func MapCat(fn Value, seq Value) Value {
	cursor := newSeqCursor(seq)
	var innerCursor seqCursor
	hasInner := false

	state := &lazyListState{
		next: func() (Value, bool) {
			for {
				if hasInner {
					next, ok := innerCursor.nextOrDone()
					if ok {
						return next, true
					}
					hasInner = false
				}

				nextItem, ok := cursor.nextOrDone()
				if !ok {
					return Value{}, false
				}
				inner := Call(fn, nextItem)
				if IsNil(inner) {
					continue
				}
				innerCursor = newSeqCursor(inner)
				hasInner = true
			}
		},
	}
	return newLazyListValue(state, -1, 0)
}

func lazyMap(fn Value, seqs ...Value) Value {
	cursors := make([]seqCursor, len(seqs))
	for i, seq := range seqs {
		cursors[i] = newSeqCursor(seq)
	}

	return NewLazyList(func() (Value, bool) {
		callArgs := make([]Value, len(seqs))
		for i := range cursors {
			next, ok := cursors[i].nextOrDone()
			if !ok {
				return Value{}, false
			}
			callArgs[i] = next
		}
		return Call(fn, callArgs...), true
	})
}

func allLazySeqs(seqs []Value) bool {
	for _, seq := range seqs {
		if isType[*lazyListImpl](seq) {
			continue
		}
		if isType[*rangeImpl](seq) && !asImpl[*rangeImpl](seq, "allLazySeqs").finite {
			continue
		}
		return false
	}
	return true
}

func PMap(fn Value, seqs ...Value) Value {
	if len(seqs) == 0 {
		panic("pmap expects function and at least one sequence")
	}

	cursors := make([]seqCursor, len(seqs))
	shortestKnown := -1
	allKnown := true
	for i, seq := range seqs {
		cursor := newSeqCursor(seq)
		cursors[i] = cursor
		if remaining, ok := cursor.remainingKnown(); ok {
			if shortestKnown == -1 || remaining < shortestKnown {
				shortestKnown = remaining
			}
		} else {
			allKnown = false
		}
	}
	if shortestKnown == 0 {
		return NewArray()
	}

	initialCap := 0
	switch {
	case allKnown:
		initialCap = shortestKnown
	case shortestKnown > 0:
		initialCap = shortestKnown
	}

	inputs := make([][]Value, 0, initialCap)
	for {
		callArgs := make([]Value, len(seqs))
		for s := range cursors {
			next, ok := cursors[s].nextOrDone()
			if !ok {
				return pmapApply(fn, inputs)
			}
			callArgs[s] = next
		}
		inputs = append(inputs, callArgs)
	}
}

func pmapApply(fn Value, inputs [][]Value) Value {
	if len(inputs) == 0 {
		return NewArray()
	}

	workers := pmapWorkerCount(len(inputs))
	results := make([]Value, len(inputs))

	var (
		wg       sync.WaitGroup
		panicVal any
		panicMu  sync.Mutex
	)
	recordPanic := func(v any) {
		panicMu.Lock()
		if panicVal == nil {
			panicVal = v
		}
		panicMu.Unlock()
	}

	wg.Add(workers)
	for worker := 0; worker < workers; worker++ {
		start := worker
		go func() {
			defer wg.Done()
			defer func() {
				if v := recover(); v != nil {
					recordPanic(v)
				}
			}()
			for i := start; i < len(inputs); i += workers {
				results[i] = Call(fn, inputs[i]...)
			}
		}()
	}
	wg.Wait()
	if panicVal != nil {
		panic(panicVal)
	}

	return NewArray(results...)
}

func pmapWorkerCount(itemCount int) int {
	if itemCount <= 0 {
		return 0
	}
	workers := goruntime.NumCPU() * 2
	if workers < 1 {
		workers = 1
	}
	if workers > itemCount {
		workers = itemCount
	}
	return workers
}

func Filter(fn Value, seq Value) Value {
	cursor := newSeqCursor(seq)
	if remaining, ok := cursor.remainingKnown(); ok && remaining == 0 {
		return NewArray()
	}

	initialCap := 0
	if remaining, ok := cursor.remainingKnown(); ok && remaining > 0 {
		initialCap = remaining
	}
	filtered := make([]Value, 0, initialCap)
	for {
		item, ok := cursor.nextOrDone()
		if !ok {
			break
		}
		if IsTruthy(Call(fn, item)) {
			filtered = append(filtered, item)
		}
	}
	return NewArray(filtered...)
}

func Some(fn Value, seq Value) Value {
	if IsNil(seq) {
		return NilValue
	}
	cursor := newSeqCursor(seq)
	for {
		item, ok := cursor.nextOrDone()
		if !ok {
			return NilValue
		}
		result := Call(fn, item)
		if IsTruthy(result) {
			return result
		}
	}
}

func NotEmpty(coll Value) Value {
	return coll.impl.notEmpty(coll)
}

func Seq(coll Value) Value {
	return coll.impl.seq(coll)
}

func IsEmpty(coll Value) bool {
	return coll.impl.isEmpty(coll)
}

func IsNil(v Value) bool {
	return v.impl.isNil(v)
}

func Set(coll Value) Value {
	cursor := newSeqCursor(coll)
	items := make([]Value, 0)
	if remaining, ok := cursor.remainingKnown(); ok && remaining > 0 {
		items = make([]Value, 0, remaining)
	}
	for {
		next, ok := cursor.nextOrDone()
		if !ok {
			break
		}
		items = append(items, next)
	}
	return NewSet(items...)
}

func Vec(coll Value) Value {
	cursor := newSeqCursor(coll)
	items := make([]Value, 0)
	if remaining, ok := cursor.remainingKnown(); ok && remaining > 0 {
		items = make([]Value, 0, remaining)
	}
	for {
		next, ok := cursor.nextOrDone()
		if !ok {
			break
		}
		items = append(items, next)
	}
	return NewArray(items...)
}

func SortBy(keyFn Value, args ...Value) Value {
	var (
		cmpFn Value
		coll  Value
	)
	switch len(args) {
	case 1:
		coll = args[0]
	case 2:
		cmpFn = args[0]
		coll = args[1]
	default:
		panic("sort-by expects key function, optional comparator, and collection")
	}

	cursor := newSeqCursor(coll)
	type sortItem struct {
		value Value
		key   Value
	}
	items := make([]sortItem, 0)
	if remaining, ok := cursor.remainingKnown(); ok && remaining > 0 {
		items = make([]sortItem, 0, remaining)
	}
	for {
		value, ok := cursor.nextOrDone()
		if !ok {
			break
		}
		items = append(items, sortItem{
			value: value,
			key:   Call(keyFn, value),
		})
	}

	if cmpFn.impl != nil {
		sort.SliceStable(items, func(i, j int) bool {
			return IsTruthy(Call(cmpFn, items[i].key, items[j].key))
		})
	} else {
		sort.SliceStable(items, func(i, j int) bool {
			left := items[i].key
			right := items[j].key
			if IsNil(left) {
				return !IsNil(right)
			}
			if IsNil(right) {
				return false
			}
			return Lt(left, right)
		})
	}

	out := make([]Value, len(items))
	for i := range items {
		out[i] = items[i].value
	}
	return NewArray(out...)
}

func Conj(coll Value, entries ...Value) Value {
	if len(entries) == 0 {
		panic("conj expects collection and at least one item")
	}
	return coll.impl.conj(coll, entries...)
}

func Into(coll Value, from Value) Value {
	if isType[*arrayImpl](coll) || isType[*listImpl](coll) || isType[*vectorImpl](coll) || isMap(coll) || isType[*setImpl](coll) {
		out := Transient(coll)
		cursor := newSeqCursor(from)
		for {
			next, ok := cursor.nextOrDone()
			if !ok {
				return PersistentBang(out)
			}
			out = ConjBang(out, next)
		}
	}
	cursor := newSeqCursor(from)
	out := coll
	for {
		next, ok := cursor.nextOrDone()
		if !ok {
			return out
		}
		out = Conj(out, next)
	}
}

func conjMapEntry(entry Value) (Value, Value) {
	if !entry.impl.seqQ(entry) {
		panic("conj on map expects pair sequences")
	}
	cursor := newSeqCursor(entry)
	key, ok := cursor.nextOrDone()
	if !ok {
		panic("conj map entry sequence must contain exactly two items")
	}
	value, ok := cursor.nextOrDone()
	if !ok {
		panic("conj map entry sequence must contain exactly two items")
	}
	if _, ok := cursor.nextOrDone(); ok {
		panic("conj map entry sequence must contain exactly two items")
	}
	return key, value
}

func Count(coll Value) int {
	return coll.impl.count(coll)
}

func notEmptyIf(coll Value, empty bool) Value {
	if empty {
		return NilValue
	}
	return coll
}

func DoAll(coll Value) Value {
	if !coll.impl.seqQ(coll) {
		return coll
	}
	cursor := newSeqCursor(coll)
	items := make([]Value, 0)
	if remaining, ok := cursor.remainingKnown(); ok && remaining > 0 {
		items = make([]Value, 0, remaining)
	}
	for {
		next, ok := cursor.nextOrDone()
		if !ok {
			break
		}
		items = append(items, next)
	}
	return NewArray(items...)
}

func DoRun(coll Value) Value {
	_ = DoAll(coll)
	return NilValue
}

func Juxt(fns ...Value) Value {
	copied := make([]Value, len(fns))
	copy(copied, fns)
	return NewFunction(func(args ...Value) Value {
		out := make([]Value, len(copied))
		for i, f := range copied {
			if s, ok := f.impl.(*symbolImpl); ok && s.obj.IsKeyword && len(args) == 1 {
				out[i] = KeywordGet(args[0], f)
				continue
			}
			out[i] = Call(f, args...)
		}
		return NewArray(out...)
	})
}

func GroupBy(fn, coll Value) Value {
	type groupBucket struct {
		key Value
		xs  []Value
	}
	byHash := map[uint64][]int{}
	groups := []groupBucket{}
	keyword := false
	if s, ok := fn.impl.(*symbolImpl); ok && s.obj.IsKeyword {
		keyword = true
	}
	cursor := newSeqCursor(coll)
	for {
		item, ok := cursor.nextOrDone()
		if !ok {
			break
		}
		var key Value
		if keyword {
			key = KeywordGet(item, fn)
		} else {
			key = Call(fn, item)
		}
		h := Hash(key)
		found := -1
		for _, gi := range byHash[h] {
			if Eq(groups[gi].key, key) {
				found = gi
				break
			}
		}
		if found < 0 {
			found = len(groups)
			groups = append(groups, groupBucket{key: key})
			byHash[h] = append(byHash[h], found)
		}
		groups[found].xs = append(groups[found].xs, item)
	}
	entries := make([]Value, 0, len(groups)*2)
	for _, g := range groups {
		entries = append(entries, g.key, newArrayValue(g.xs, len(g.xs)))
	}
	return NewMap(entries...)
}

func Reduce(fn Value, args ...Value) Value {
	if len(args) != 1 && len(args) != 2 {
		panic("reduce expects function, optional initial value, and one sequence")
	}

	var (
		acc    Value
		cursor seqCursor
	)

	if len(args) == 1 {
		cursor = newSeqCursor(args[0])
		next, ok := cursor.nextOrDone()
		if !ok {
			return Unreduced(Call(fn))
		}
		acc = next
	} else {
		acc = args[0]
		cursor = newSeqCursor(args[1])
	}

	for {
		next, ok := cursor.nextOrDone()
		if !ok {
			break
		}
		acc = Call(fn, acc, next)
		if IsReduced(acc) {
			return Unreduced(acc)
		}
	}
	return acc
}

func Take(n Value, coll Value) Value {
	count := nonNegativeCount("take", n)
	if count == 0 {
		return NewArray()
	}

	cursor := newSeqCursor(coll)
	if remaining, ok := cursor.remainingKnown(); ok && remaining < count {
		count = remaining
	}
	if count <= 0 {
		return NewArray()
	}

	out := make([]Value, 0, count)
	for i := 0; i < count; i++ {
		next, ok := cursor.nextOrDone()
		if !ok {
			break
		}
		out = append(out, next)
	}
	return NewArray(out...)
}

func Drop(n Value, coll Value) Value {
	count := nonNegativeCount("drop", n)
	if count == 0 {
		return coll
	}
	return coll.impl.drop(coll, count)
}

func Nth(coll Value, index Value, notFound ...Value) Value {
	if len(notFound) > 1 {
		panic("nth expects collection, index, and optional default")
	}
	i := nonNegativeCount("nth", index)
	missing, hasMissing := nthMissing(notFound)
	return coll.impl.nth(coll, i, missing, hasMissing)
}

func SlowNth(coll Value, index Value, notFound ...Value) Value {
	if len(notFound) > 1 {
		panic("slow-nth expects collection, index, and optional default")
	}
	i := nonNegativeCount("slow-nth", index)
	missing, hasMissing := nthMissing(notFound)
	return coll.impl.slowNth(coll, i, missing, hasMissing)
}

func nthAt(length int, i int, missing Value, hasMissing bool, get func(int) Value) Value {
	if i >= length {
		if hasMissing {
			return missing
		}
		panic("nth index out of range")
	}
	return get(i)
}

func slowNthWalk(coll Value, i int, missing Value, hasMissing bool) Value {
	return walkIndex(coll, i, missing, hasMissing, "slow-nth")
}

func walkIndex(coll Value, i int, missing Value, hasMissing bool, name string) Value {
	cursor := newSeqCursor(coll)
	for pos := 0; pos <= i; pos++ {
		next, ok := cursor.nextOrDone()
		if !ok {
			if hasMissing {
				return missing
			}
			panic(name + " index out of range")
		}
		if pos == i {
			return next
		}
	}
	panic("unreachable")
}

func nthMissing(notFound []Value) (Value, bool) {
	if len(notFound) == 0 {
		return Value{}, false
	}
	return notFound[0], true
}

func First(coll Value) Value {
	return coll.impl.first(coll)
}

// Peek is Clojure-style stack peek: first of a list, last of an array or vector.
// Empty collections and nil yield nil.
func Peek(coll Value) Value {
	switch {
	case IsNil(coll):
		return NilValue
	case isType[*listImpl](coll):
		return First(coll)
	case isType[*arrayImpl](coll), isType[*vectorImpl](coll):
		return Last(coll)
	default:
		panic("peek expects list, array, or vector")
	}
}

// Pop is Clojure-style stack pop: drop the first of a list, or the last of an
// array or vector. Empty collections and nil panic. Result keeps the input kind.
func Pop(coll Value) Value {
	switch {
	case IsNil(coll):
		panic("Can't pop empty list")
	case isType[*listImpl](coll):
		if coll.ListLen() == 0 {
			panic("Can't pop empty list")
		}
		return ListRest(coll)
	case isType[*arrayImpl](coll):
		n := coll.ArrayLen()
		if n == 0 {
			panic("Can't pop empty array")
		}
		return NewArray(coll.arrayItems()[:n-1]...)
	case isType[*vectorImpl](coll):
		return VectorPop(coll)
	default:
		panic("pop expects list, array, or vector")
	}
}

func SeqPredicate(coll Value) Value {
	return NewBool(coll.impl.seqQ(coll))
}

func Rest(coll Value) Value {
	return coll.impl.rest(coll)
}

func Next(coll Value) Value {
	return coll.impl.next(coll)
}

func Last(coll Value) Value {
	return coll.impl.last(coll)
}

func nextIfEmpty(rest Value, empty bool) Value {
	if empty {
		return NilValue
	}
	return rest
}

func Reverse(coll Value) Value {
	cursor := newSeqCursor(coll)
	items := make([]Value, 0)
	if remaining, ok := cursor.remainingKnown(); ok && remaining > 0 {
		items = make([]Value, 0, remaining)
	}
	for {
		next, ok := cursor.nextOrDone()
		if !ok {
			break
		}
		items = append(items, next)
	}
	for left, right := 0, len(items)-1; left < right; left, right = left+1, right-1 {
		items[left], items[right] = items[right], items[left]
	}
	return NewArray(items...)
}

func Cons(item Value, coll Value) Value {
	return coll.impl.cons(coll, item)
}

func seqToList(coll Value) Value {
	cursor := newSeqCursor(coll)
	items := make([]Value, 0)
	if remaining, ok := cursor.remainingKnown(); ok && remaining > 0 {
		items = make([]Value, 0, remaining)
	}
	for {
		next, ok := cursor.nextOrDone()
		if !ok {
			break
		}
		items = append(items, next)
	}
	return NewList(items...)
}

func SeqFirst(coll Value) Value {
	if IsNil(coll) {
		return NilValue
	}
	return First(coll)
}

func SeqRest(coll Value) Value {
	return Rest(coll)
}

func Range(args ...Value) Value {
	switch len(args) {
	case 0:
		return newRangeValue(0, 0, 1, false)
	case 1:
		end := rangeArgLong("range", args[0])
		return finiteRange(0, end, 1)
	case 2:
		start := rangeArgLong("range", args[0])
		end := rangeArgLong("range", args[1])
		return finiteRange(start, end, 1)
	case 3:
		start := rangeArgLong("range", args[0])
		end := rangeArgLong("range", args[1])
		step := rangeArgLong("range", args[2])
		return finiteRange(start, end, step)
	default:
		panic("range expects zero to three integer arguments")
	}
}

func finiteRange(start, end, step int64) Value {
	if step == 0 {
		if start == end {
			return NewArray()
		}
		return newLazyRepeat(NewLong(start), 0, false)
	}
	count := rangeCount(start, end, step)
	if count <= 0 {
		return NewArray()
	}
	return newRangeValue(start, end, step, true)
}

func rangeCount(start, end, step int64) int64 {
	if step > 0 {
		if start >= end {
			return 0
		}
		return (end - start + step - 1) / step
	}
	if start <= end {
		return 0
	}
	return (start - end - step - 1) / (-step)
}

func Repeat(args ...Value) Value {
	switch len(args) {
	case 1:
		return newLazyRepeat(args[0], 0, false)
	case 2:
		count := nonNegativeCount("repeat", args[0])
		return newLazyRepeat(args[1], int64(count), true)
	default:
		panic("repeat expects one or two arguments")
	}
}

func newLazyRepeat(value Value, remaining int64, finite bool) Value {
	state := &lazyListState{
		next: func() (Value, bool) {
			if finite {
				if remaining <= 0 {
					return Value{}, false
				}
				remaining--
			}
			return value, true
		},
	}
	hint := int64(-1)
	if finite {
		hint = remaining
	}
	return newLazyListValue(state, hint, 0)
}

func rangeArgLong(name string, value Value) int64 {
	if !isType[longImpl](value) {
		panic(name + " expects integer arguments")
	}
	return value.Long()
}

func nonNegativeCount(name string, value Value) int {
	return value.impl.asCount(value, name)
}

type seqCursorKind uint8

const (
	seqKindArray seqCursorKind = iota + 1
	seqKindList
	seqKindLazy
	seqKindString
	seqKindRange
)

type seqCursor struct {
	kind       seqCursorKind
	listNode   *listNode
	lazyList   Value
	arrayVals  []Value
	strValue   string
	strIndex   int
	arrayIdx   int
	length     int
	hasLength  bool
	rangeStart int64
	rangeStep  int64
}

func newSeqCursor(coll Value) seqCursor {
	return coll.impl.seqCursor(coll)
}

func emptySeqCursor() seqCursor {
	return arraySeqCursor(nil)
}

func arraySeqCursor(vals []Value) seqCursor {
	return seqCursor{
		kind:      seqKindArray,
		arrayVals: vals,
		length:    len(vals),
		hasLength: true,
	}
}

func listSeqCursor(node *listNode, length int) seqCursor {
	return seqCursor{
		kind:      seqKindList,
		listNode:  node,
		length:    length,
		hasLength: true,
	}
}

func lazySeqCursor(list Value, length int, hasLength bool) seqCursor {
	return seqCursor{
		kind:      seqKindLazy,
		lazyList:  list,
		length:    length,
		hasLength: hasLength,
	}
}

func stringSeqCursor(s string) seqCursor {
	return seqCursor{
		kind:      seqKindString,
		strValue:  s,
		hasLength: false,
	}
}

func rangeSeqCursor(start, step int64, length int, hasLength bool) seqCursor {
	return seqCursor{
		kind:       seqKindRange,
		rangeStart: start,
		rangeStep:  step,
		length:     length,
		hasLength:  hasLength,
	}
}

func mapEntriesSeqCursor(entries []MapEntry) seqCursor {
	pairs := make([]Value, 0, len(entries))
	for _, entry := range entries {
		pairs = append(pairs, NewArray(entry.Key, entry.Value))
	}
	return arraySeqCursor(pairs)
}

func (c *seqCursor) remainingKnown() (int, bool) {
	if !c.hasLength {
		return 0, false
	}
	switch c.kind {
	case seqKindList:
		return c.length, true
	case seqKindArray:
		return c.length - c.arrayIdx, true
	case seqKindLazy:
		return c.length, true
	case seqKindRange:
		return c.length, true
	default:
		panic("unknown cursor kind")
	}
}

func (c *seqCursor) nextOrDone() (Value, bool) {
	switch c.kind {
	case seqKindList:
		if c.length == 0 {
			return Value{}, false
		}
		value := c.listNode.value
		c.listNode = c.listNode.next
		c.length--
		return value, true
	case seqKindArray:
		if c.arrayIdx >= c.length {
			return Value{}, false
		}
		value := c.arrayVals[c.arrayIdx]
		c.arrayIdx++
		return value, true
	case seqKindLazy:
		value, ok := lazyPeek(c.lazyList)
		if !ok {
			if c.hasLength {
				c.length = 0
			}
			return Value{}, false
		}
		c.lazyList = lazyTail(c.lazyList)
		if c.hasLength && c.length > 0 {
			c.length--
		}
		return value, true
	case seqKindString:
		if c.strIndex >= len(c.strValue) {
			return Value{}, false
		}
		_, size := utf8.DecodeRuneInString(c.strValue[c.strIndex:])
		start := c.strIndex
		c.strIndex += size
		return NewString(c.strValue[start:c.strIndex]), true
	case seqKindRange:
		if c.hasLength {
			if c.length == 0 {
				return Value{}, false
			}
			c.length--
		}
		value := NewLong(c.rangeStart)
		c.rangeStart += c.rangeStep
		return value, true
	default:
		panic("unknown cursor kind")
	}
}
