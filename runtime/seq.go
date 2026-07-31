package runtime

import (
	goruntime "runtime"
	"sort"
	"sync"
)

const eagerRangeThreshold = 1000

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

func Juxt(fns ...Value) Value {
	if len(fns) == 0 {
		panic("juxt expects at least one function")
	}
	return NewFunction(func(args ...Value) Value {
		out := make([]Value, len(fns))
		for i, fn := range fns {
			out[i] = Call(fn, args...)
		}
		return NewArray(out...)
	})
}

func Partial(fn Value, boundArgs ...Value) Value {
	prefix := make([]Value, len(boundArgs))
	copy(prefix, boundArgs)
	return NewFunction(func(args ...Value) Value {
		callArgs := make([]Value, 0, len(prefix)+len(args))
		callArgs = append(callArgs, prefix...)
		callArgs = append(callArgs, args...)
		return Call(fn, callArgs...)
	})
}

func Apply(fn Value, args ...Value) Value {
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
	return Call(fn, callArgs...)
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
				if inner.tag == TagNil {
					continue
				}
				innerCursor = newSeqCursor(inner)
				hasInner = true
			}
		},
	}
	return newLazyListValue(state, -1, 0)
}

func MaxKey(keyFn Value, values ...Value) Value {
	if len(values) == 0 {
		panic("max-key expects key function and at least one value")
	}
	best := values[0]
	bestKey := Call(keyFn, best)
	for _, candidate := range values[1:] {
		candidateKey := Call(keyFn, candidate)
		if maxKeyTakeCandidate(bestKey, candidateKey) {
			best = candidate
			bestKey = candidateKey
		}
	}
	return best
}

func maxKeyTakeCandidate(currentKey, candidateKey Value) bool {
	if currentKey.tag == TagNil {
		return true
	}
	if candidateKey.tag == TagNil {
		return false
	}
	if Gt(currentKey, candidateKey) {
		return false
	}
	return true
}

func Val(entry Value) Value {
	switch entry.tag {
	case TagList, TagArray, TagLazyList:
		// valid map-entry-like sequence input
	default:
		panic("val expects map-entry sequence Value")
	}
	cursor := newSeqCursor(entry)
	if _, ok := cursor.nextOrDone(); !ok {
		return NilValue()
	}
	value, ok := cursor.nextOrDone()
	if !ok {
		return NilValue()
	}
	return value
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
		if seq.tag != TagLazyList {
			return false
		}
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

func Remove(fn Value, seq Value) Value {
	negate := NewFunction(func(args ...Value) Value {
		if IsTruthy(Call(fn, args...)) {
			return NilValue()
		}
		return NewBool(true)
	})
	return Filter(negate, seq)
}

func Keep(fn Value, seq Value) Value {
	cursor := newSeqCursor(seq)
	results := make([]Value, 0)
	if remaining, ok := cursor.remainingKnown(); ok && remaining > 0 {
		results = make([]Value, 0, remaining)
	}
	for {
		item, ok := cursor.nextOrDone()
		if !ok {
			break
		}
		mapped := Call(fn, item)
		if mapped.tag != TagNil {
			results = append(results, mapped)
		}
	}
	return NewArray(results...)
}

func Some(fn Value, seq Value) Value {
	if seq.tag == TagNil {
		return NilValue()
	}
	cursor := newSeqCursor(seq)
	for {
		item, ok := cursor.nextOrDone()
		if !ok {
			return NilValue()
		}
		result := Call(fn, item)
		if IsTruthy(result) {
			return result
		}
	}
}

func SomePredicate(v Value) Value {
	return NewBool(!IsNil(v))
}

func NotAny(fn Value, seq Value) Value {
	return NewBool(!IsTruthy(Some(fn, seq)))
}

func Every(fn Value, seq Value) Value {
	cursor := newSeqCursor(seq)
	for {
		item, ok := cursor.nextOrDone()
		if !ok {
			return NewBool(true)
		}
		if !IsTruthy(Call(fn, item)) {
			return NewBool(false)
		}
	}
}

func NotEmpty(coll Value) Value {
	switch coll.tag {
	case TagList:
		if coll.ListLen() == 0 {
			return NilValue()
		}
		return coll
	case TagArray:
		if coll.ArrayLen() == 0 {
			return NilValue()
		}
		return coll
	case TagLazyList:
		if _, ok := lazyPeek(coll); !ok {
			return NilValue()
		}
		return coll
	case TagMap:
		if coll.MapLen() == 0 {
			return NilValue()
		}
		return coll
	case TagDate:
		if coll.MapLen() == 0 {
			return NilValue()
		}
		return coll
	case TagSet:
		if coll.SetLen() == 0 {
			return NilValue()
		}
		return coll
	case TagString:
		if coll.StringValue() == "" {
			return NilValue()
		}
		return coll
	default:
		panic("not-empty expects list, array, lazy-list, map, set, or string Value")
	}
}

func Seq(coll Value) Value {
	return NotEmpty(coll)
}

func IsEmpty(coll Value) bool {
	switch coll.tag {
	case TagNil:
		return true
	case TagList:
		return coll.ListLen() == 0
	case TagArray:
		return coll.ArrayLen() == 0
	case TagLazyList:
		_, ok := lazyPeek(coll)
		return !ok
	case TagMap, TagDate:
		return coll.MapLen() == 0
	case TagSet:
		return coll.SetLen() == 0
	case TagString:
		return coll.StringValue() == ""
	default:
		panic("empty? expects collection Value")
	}
}

func IsNotEmpty(coll Value) bool {
	return !IsEmpty(coll)
}

func IsNil(v Value) bool {
	return v.tag == TagNil
}

func Not(v Value) Value {
	return NewBool(!IsTruthy(v))
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

func GroupBy(fn Value, coll Value) Value {
	cursor := newSeqCursor(coll)
	grouped := NewMap()
	for {
		item, ok := cursor.nextOrDone()
		if !ok {
			break
		}
		key := Call(fn, item)
		existing := Get(grouped, key)
		if existing.tag == TagNil {
			grouped = MapAssoc(grouped, key, NewArray(item))
			continue
		}
		if existing.tag != TagArray {
			panic("group-by internal error: expected array bucket")
		}
		grouped = MapAssoc(grouped, key, ArrayAppend(existing, item))
	}
	return grouped
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

	if cmpFn.tag != 0 {
		sort.SliceStable(items, func(i, j int) bool {
			return IsTruthy(Call(cmpFn, items[i].key, items[j].key))
		})
	} else {
		sort.SliceStable(items, func(i, j int) bool {
			left := items[i].key
			right := items[j].key
			if left.tag == TagNil {
				return right.tag != TagNil
			}
			if right.tag == TagNil {
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
	switch coll.tag {
	case TagList:
		out := coll
		for _, item := range entries {
			out = ListCons(out, item)
		}
		return out
	case TagArray:
		out := coll
		for _, item := range entries {
			out = ArrayAppend(out, item)
		}
		return out
	case TagSet:
		items := coll.setPointer().items
		for _, item := range entries {
			items = items.Add(item)
		}
		return newSetValue(items)
	case TagMap:
		items := coll.mapPointer().items
		for _, entry := range entries {
			if entry.tag == TagMap {
				for _, mapEntry := range entry.MapEntries() {
					items = items.Set(mapEntry.Key, mapEntry.Value)
				}
				continue
			}
			key, value := conjMapEntry(entry)
			items = items.Set(key, value)
		}
		return newMapValue(items)
	default:
		panic("conj expects list, array, set, or map Value")
	}
}

func Into(coll Value, from Value) Value {
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
	switch entry.tag {
	case TagList, TagArray, TagLazyList:
		// valid pair sequence input
	default:
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

func ZipMap(keys Value, vals Value) Value {
	keyCursor := newSeqCursor(keys)
	valCursor := newSeqCursor(vals)
	entries := make([]Value, 0)
	if keyRemaining, keyKnown := keyCursor.remainingKnown(); keyKnown {
		if valRemaining, valKnown := valCursor.remainingKnown(); valKnown {
			capacity := keyRemaining
			if valRemaining < capacity {
				capacity = valRemaining
			}
			if capacity > 0 {
				entries = make([]Value, 0, capacity*2)
			}
		}
	}

	for {
		key, ok := keyCursor.nextOrDone()
		if !ok {
			break
		}
		val, ok := valCursor.nextOrDone()
		if !ok {
			break
		}
		entries = append(entries, key, val)
	}
	return NewMap(entries...)
}

func Count(coll Value) int {
	switch coll.tag {
	case TagList:
		return coll.ListLen()
	case TagArray:
		return coll.ArrayLen()
	case TagLazyList:
		// For lazy lists, count by iterating
		cursor := newSeqCursor(coll)
		count := 0
		for {
			_, ok := cursor.nextOrDone()
			if !ok {
				break
			}
			count++
		}
		return count
	case TagMap:
		return coll.MapLen()
	case TagDate:
		return coll.MapLen()
	case TagSet:
		return coll.SetLen()
	case TagString:
		// Return the length of the string
		return len(coll.StringValue())
	case TagNil:
		return 0
	default:
		panic("count expects collection Value")
	}
}

func DoAll(coll Value) Value {
	switch coll.tag {
	case TagList, TagArray, TagLazyList:
		// Realize seqable collections.
	default:
		// Non-seq values are already realized; keep doall as a no-op.
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
	return NilValue()
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
			return Call(fn)
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

	switch coll.tag {
	case TagArray:
		length := coll.ArrayLen()
		if count >= length {
			return NewArray()
		}
		items := coll.arrayItems()
		return newArrayValue(items[count:length], length-count)
	case TagList:
		out := coll
		for i := 0; i < count && out.ListLen() > 0; i++ {
			out = ListRest(out)
		}
		return out
	case TagLazyList:
		out := coll
		for i := 0; i < count; i++ {
			if got := First(out); got.tag == TagNil {
				return out
			}
			out = lazyTail(out)
		}
		return out
	default:
		panic("drop expects list, array, or lazy-list Value")
	}
}

func First(coll Value) Value {
	switch coll.tag {
	case TagNil:
		return NilValue()
	case TagList:
		if coll.ListLen() == 0 {
			return NilValue()
		}
		return coll.listPointer().value
	case TagArray:
		if coll.ArrayLen() == 0 {
			return NilValue()
		}
		return coll.arrayItems()[0]
	case TagLazyList:
		next, ok := lazyPeek(coll)
		if !ok {
			return NilValue()
		}
		return next
	case TagString:
		value := coll.StringValue()
		if value == "" {
			return NilValue()
		}
		runes := []rune(value)
		return NewString(string(runes[0]))
	default:
		panic("first expects nil, list, array, lazy-list, or string Value")
	}
}

func Second(coll Value) Value {
	return First(Rest(coll))
}

func SeqPredicate(coll Value) Value {
	switch coll.tag {
	case TagList, TagArray, TagLazyList:
		return NewBool(true)
	default:
		return NewBool(false)
	}
}

func Rest(coll Value) Value {
	switch coll.tag {
	case TagList:
		return ListRest(coll)
	case TagArray:
		return ArrayRest(coll)
	case TagLazyList:
		return lazyTail(coll)
	default:
		panic("rest expects list, array, or lazy-list Value")
	}
}

func Next(coll Value) Value {
	if coll.tag == TagNil {
		return NilValue()
	}
	rest := Rest(coll)
	switch rest.tag {
	case TagList:
		if rest.ListLen() == 0 {
			return NilValue()
		}
	case TagArray:
		if rest.ArrayLen() == 0 {
			return NilValue()
		}
	case TagLazyList:
		if _, ok := lazyPeek(rest); !ok {
			return NilValue()
		}
	default:
		panic("next expects list, array, or lazy-list Value")
	}
	return rest
}

func Last(coll Value) Value {
	switch coll.tag {
	case TagList:
		if coll.ListLen() == 0 {
			return NilValue()
		}
		node := coll.listPointer()
		for node.next != nil {
			node = node.next
		}
		return node.value
	case TagArray:
		if coll.ArrayLen() == 0 {
			return NilValue()
		}
		items := coll.arrayItems()
		return items[coll.ArrayLen()-1]
	case TagLazyList:
		cursor := newSeqCursor(coll)
		last := NilValue()
		for {
			next, ok := cursor.nextOrDone()
			if !ok {
				return last
			}
			last = next
		}
	default:
		panic("last expects list, array, or lazy-list Value")
	}
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
	switch coll.tag {
	case TagNil:
		return NewList(item)
	case TagList:
		return ListCons(coll, item)
	default:
		panic("cons expects nil or list Value as second argument")
	}
}

func SeqFirst(coll Value) Value {
	if coll.tag == TagNil {
		return NilValue()
	}
	return First(coll)
}

func SeqRest(coll Value) Value {
	if coll.tag == TagNil {
		return NewList()
	}
	return Rest(coll)
}

func Range(args ...Value) Value {
	switch len(args) {
	case 0:
		return newLazyRange(0, 0, false)
	case 1:
		start := rangeArgLong("range", args[0])
		return newLazyRange(start, 0, false)
	case 2:
		start := rangeArgLong("range", args[0])
		end := rangeArgLong("range", args[1])
		if end <= start {
			return NewArray()
		}
		count := end - start
		if count < eagerRangeThreshold {
			values := make([]Value, 0, count)
			for i := start; i < end; i++ {
				values = append(values, NewLong(i))
			}
			return NewArray(values...)
		}
		return newLazyRange(start, end, true)
	default:
		panic("range expects one or two integer arguments")
	}
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

func newLazyRange(start, end int64, finite bool) Value {
	current := start
	state := &lazyListState{
		next: func() (Value, bool) {
			if finite && current >= end {
				return Value{}, false
			}
			value := NewLong(current)
			current++
			return value, true
		},
	}
	remaining := int64(-1)
	if finite {
		remaining = end - start
	}
	return newLazyListValue(state, remaining, 0)
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
	if value.tag != TagLong {
		panic(name + " expects integer arguments")
	}
	return value.Long()
}

func nonNegativeCount(name string, value Value) int {
	switch value.tag {
	case TagLong:
		if value.Long() < 0 {
			panic(name + " expects non-negative count")
		}
		return int(value.Long())
	case TagBigInt:
		bi := value.BigInt()
		if bi.Sign() < 0 {
			panic(name + " expects non-negative count")
		}
		if bi.BitLen() > 62 {
			panic(name + " count is too large")
		}
		return int(bi.Int64())
	default:
		panic(name + " expects integer count")
	}
}

type seqCursor struct {
	kind      ValueTag
	listNode  *listNode
	lazyList  Value
	arrayVals []Value
	arrayIdx  int
	length    int
	hasLength bool
}

func newSeqCursor(coll Value) seqCursor {
	switch coll.tag {
	case TagNil:
		return seqCursor{
			kind:      TagArray,
			arrayVals: nil,
			length:    0,
			hasLength: true,
		}
	case TagList:
		return seqCursor{
			kind:      TagList,
			listNode:  coll.listPointer(),
			length:    coll.ListLen(),
			hasLength: true,
		}
	case TagArray:
		length := coll.ArrayLen()
		return seqCursor{
			kind:      TagArray,
			arrayVals: coll.arrayItems()[:length],
			length:    length,
			hasLength: true,
		}
	case TagLazyList:
		remaining := coll.lazyListRemainingHint()
		hasLength := remaining >= 0
		length := 0
		if hasLength {
			length = int(remaining)
		}
		return seqCursor{
			kind:      TagLazyList,
			lazyList:  coll,
			length:    length,
			hasLength: hasLength,
		}
	case TagMap, TagDate:
		entries := coll.MapEntries()
		pairs := make([]Value, 0, len(entries))
		for _, entry := range entries {
			pairs = append(pairs, NewArray(entry.Key, entry.Value))
		}
		return seqCursor{
			kind:      TagArray,
			arrayVals: pairs,
			length:    len(pairs),
			hasLength: true,
		}
	default:
		panic("sequence expects list, array, lazy-list, map, or nil Value")
	}
}

func (c *seqCursor) remainingKnown() (int, bool) {
	if !c.hasLength {
		return 0, false
	}
	switch c.kind {
	case TagList:
		return c.length, true
	case TagArray:
		return c.length - c.arrayIdx, true
	case TagLazyList:
		return c.length, true
	default:
		panic("unknown cursor kind")
	}
}

func (c *seqCursor) nextOrDone() (Value, bool) {
	switch c.kind {
	case TagList:
		if c.length == 0 {
			return Value{}, false
		}
		value := c.listNode.value
		c.listNode = c.listNode.next
		c.length--
		return value, true
	case TagArray:
		if c.arrayIdx >= c.length {
			return Value{}, false
		}
		value := c.arrayVals[c.arrayIdx]
		c.arrayIdx++
		return value, true
	case TagLazyList:
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
	default:
		panic("unknown cursor kind")
	}
}
