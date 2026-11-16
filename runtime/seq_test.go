package runtime

import (
	goruntime "runtime"
	"sync"
	"testing"
	"time"
)

func TestMapSingleSequence(t *testing.T) {
	double := NewFunction(func(args ...Value) Value {
		return Add(args[0], args[0])
	})

	mapped := Map(double, NewArray(NewLong(1), NewLong(2), NewLong(3)))
	if !isType[*arrayImpl](mapped) {
		t.Fatalf("expected map result array, got %v", TypeOf(mapped))
	}
	if mapped.ArrayLen() != 3 {
		t.Fatalf("expected length 3, got %d", mapped.ArrayLen())
	}
	if ArrayGet(mapped, 0).Long() != 2 || ArrayGet(mapped, 1).Long() != 4 || ArrayGet(mapped, 2).Long() != 6 {
		t.Fatalf("unexpected mapped values: %#v", mapped.ArrayValues())
	}
}

func TestMapMultipleSequencesUsesShortestLength(t *testing.T) {
	sum2 := NewFunction(func(args ...Value) Value {
		return Add(args[0], args[1])
	})

	mapped := Map(sum2, NewArray(NewLong(1), NewLong(2), NewLong(3)), NewList(NewLong(10), NewLong(20)))
	if mapped.ArrayLen() != 2 {
		t.Fatalf("expected shortest length 2, got %d", mapped.ArrayLen())
	}
	if ArrayGet(mapped, 0).Long() != 11 || ArrayGet(mapped, 1).Long() != 22 {
		t.Fatalf("unexpected mapped values: %#v", mapped.ArrayValues())
	}
}

func TestMapMultipleListsUsesShortestLength(t *testing.T) {
	sum2 := NewFunction(func(args ...Value) Value {
		return Add(args[0], args[1])
	})

	mapped := Map(sum2, NewList(NewLong(1), NewLong(2), NewLong(3)), NewList(NewLong(10), NewLong(20)))
	if mapped.ArrayLen() != 2 {
		t.Fatalf("expected shortest length 2, got %d", mapped.ArrayLen())
	}
	if ArrayGet(mapped, 0).Long() != 11 || ArrayGet(mapped, 1).Long() != 22 {
		t.Fatalf("unexpected mapped values: %#v", mapped.ArrayValues())
	}
}

func TestMapNoSequencesPanics(t *testing.T) {
	assertPanics(t, func() {
		Map(NewFunction(func(args ...Value) Value { return NilValue }))
	})
}

func TestLastAcrossSequenceTypes(t *testing.T) {
	if got := Last(NewList()); !IsNil(got) {
		t.Fatalf("expected nil for empty list, got %#v", got)
	}
	if got := Last(NewList(NewLong(1), NewLong(2), NewLong(3))); got.Long() != 3 {
		t.Fatalf("expected last list element 3, got %#v", got)
	}

	if got := Last(NewArray()); !IsNil(got) {
		t.Fatalf("expected nil for empty array, got %#v", got)
	}
	if got := Last(NewArray(NewLong(4), NewLong(5), NewLong(6))); got.Long() != 6 {
		t.Fatalf("expected last array element 6, got %#v", got)
	}

	count := 0
	lazy := NewLazyList(func() (Value, bool) {
		if count >= 3 {
			return Value{}, false
		}
		count++
		return NewLong(int64(count)), true
	})
	if got := Last(lazy); got.Long() != 3 {
		t.Fatalf("expected last lazy-list element 3, got %#v", got)
	}
	if got := First(lazy); got.Long() != 1 {
		t.Fatalf("expected lazy list to remain readable after last, got %#v", got)
	}
}

func TestSeqPredicateAcrossSequenceTypes(t *testing.T) {
	if got := SeqPredicate(NewList()); !isType[boolImpl](got) || !got.Bool() {
		t.Fatalf("expected seq? true for empty list, got %#v", got)
	}
	if got := SeqPredicate(NewArray()); !isType[boolImpl](got) || !got.Bool() {
		t.Fatalf("expected seq? true for empty array, got %#v", got)
	}
	if got := SeqPredicate(NewLazyList(func() (Value, bool) { return Value{}, false })); !isType[boolImpl](got) || !got.Bool() {
		t.Fatalf("expected seq? true for empty lazy list, got %#v", got)
	}
	if got := SeqPredicate(NilValue); !isType[boolImpl](got) || got.Bool() {
		t.Fatalf("expected seq? false for nil, got %#v", got)
	}
	if got := SeqPredicate(NewString("x")); !isType[boolImpl](got) || got.Bool() {
		t.Fatalf("expected seq? false for string, got %#v", got)
	}
}

func TestReverseAcrossSequenceTypes(t *testing.T) {
	if got := (Reverse(NewList(NewLong(1), NewLong(2), NewLong(3)))).String(); got != "[3 2 1]" {
		t.Fatalf("unexpected reverse list result: %q", got)
	}
	if got := (Reverse(NewArray(NewLong(4), NewLong(5), NewLong(6)))).String(); got != "[6 5 4]" {
		t.Fatalf("unexpected reverse array result: %q", got)
	}

	count := 0
	lazy := NewLazyList(func() (Value, bool) {
		if count >= 3 {
			return Value{}, false
		}
		count++
		return NewLong(int64(count)), true
	})
	if got := Reverse(lazy).String(); got != "[3 2 1]" {
		t.Fatalf("unexpected reverse lazy-list result: %q", got)
	}
	if got := First(lazy); got.Long() != 1 {
		t.Fatalf("expected lazy list to remain readable after reverse, got %#v", got)
	}
	if got := Reverse(NilValue).String(); got != "[]" {
		t.Fatalf("unexpected reverse nil result: %q", got)
	}
}

func TestConsBuildsLists(t *testing.T) {
	if got := (Cons(NewLong(1), NilValue)).String(); got != "(1)" {
		t.Fatalf("unexpected cons onto nil result: %q", got)
	}
	if got := (Cons(NewLong(1), NewList(NewLong(2), NewLong(3)))).String(); got != "(1 2 3)" {
		t.Fatalf("unexpected cons onto list result: %q", got)
	}
	if got := (Cons(NewLong(1), NewArray())).String(); got != "(1)" {
		t.Fatalf("unexpected cons onto empty array result: %q", got)
	}
	if got := (Cons(NewLong(1), NewArray(NewLong(2), NewLong(3)))).String(); got != "(1 2 3)" {
		t.Fatalf("unexpected cons onto array result: %q", got)
	}
}

func TestNextAcrossSequenceTypes(t *testing.T) {
	if got := Next(NilValue); !IsNil(got) {
		t.Fatalf("expected nil next for nil input, got %#v", got)
	}
	if got := Next(NewList()); !IsNil(got) {
		t.Fatalf("expected nil next for empty list, got %#v", got)
	}
	if got := Next(NewList(NewLong(1))); !IsNil(got) {
		t.Fatalf("expected nil next for singleton list, got %#v", got)
	}
	if got := Next(NewList(NewLong(1), NewLong(2))); !isType[*listImpl](got) || got.String() != "(2)" {
		t.Fatalf("expected (2) next for list, got %#v / %s", got, got.String())
	}
	if got := Next(NewArray(NewLong(1))); !IsNil(got) {
		t.Fatalf("expected nil next for singleton array, got %#v", got)
	}
	if got := Next(NewArray(NewLong(1), NewLong(2))); !isType[*arrayImpl](got) || got.String() != "[2]" {
		t.Fatalf("expected [2] next for array, got %#v / %s", got, got.String())
	}

	emptyLazy := NewLazyList(func() (Value, bool) { return Value{}, false })
	if got := Next(emptyLazy); !IsNil(got) {
		t.Fatalf("expected nil next for empty lazy list, got %#v", got)
	}
	oneLazyCount := 0
	oneLazy := NewLazyList(func() (Value, bool) {
		if oneLazyCount > 0 {
			return Value{}, false
		}
		oneLazyCount++
		return NewLong(1), true
	})
	if got := Next(oneLazy); !IsNil(got) {
		t.Fatalf("expected nil next for single-item lazy list, got %#v", got)
	}
	count := 0
	manyLazy := NewLazyList(func() (Value, bool) {
		count++
		switch count {
		case 1:
			return NewLong(1), true
		case 2:
			return NewLong(2), true
		default:
			return Value{}, false
		}
	})
	if got := Next(manyLazy); !isType[*lazyListImpl](got) || First(got).Long() != 2 {
		t.Fatalf("expected lazy next starting at 2, got %#v", got)
	}
}

func TestPMapPreservesOrder(t *testing.T) {
	delayedDouble := NewFunction(func(args ...Value) Value {
		v := args[0].Long()
		time.Sleep(time.Duration(6-v) * 2 * time.Millisecond)
		return Mul(args[0], NewLong(2))
	})

	mapped := PMap(delayedDouble, NewArray(NewLong(1), NewLong(2), NewLong(3), NewLong(4), NewLong(5)))
	if !isType[*arrayImpl](mapped) {
		t.Fatalf("expected pmap result array, got %v", TypeOf(mapped))
	}
	if mapped.ArrayLen() != 5 {
		t.Fatalf("expected length 5, got %d", mapped.ArrayLen())
	}
	want := []int64{2, 4, 6, 8, 10}
	for i := range want {
		if got := ArrayGet(mapped, i).Long(); got != want[i] {
			t.Fatalf("unexpected pmap value at %d: got %d want %d", i, got, want[i])
		}
	}
}

func TestPMapWorkerCount(t *testing.T) {
	if got := pmapWorkerCount(0); got != 0 {
		t.Fatalf("expected 0 workers for no items, got %d", got)
	}
	if got := pmapWorkerCount(1); got != 1 {
		t.Fatalf("expected 1 worker for one item, got %d", got)
	}

	expected := goruntime.NumCPU() * 2
	if expected < 1 {
		expected = 1
	}
	large := expected + 10
	if got := pmapWorkerCount(large); got != expected {
		t.Fatalf("expected %d workers for large input, got %d", expected, got)
	}
}

func TestFilterArray(t *testing.T) {
	keepEven := NewFunction(func(args ...Value) Value {
		if args[0].Long()%2 == 0 {
			return NewLong(1)
		}
		return NilValue
	})

	filtered := Filter(keepEven, NewArray(NewLong(1), NewLong(2), NewLong(3), NewLong(4)))
	if !isType[*arrayImpl](filtered) {
		t.Fatalf("expected filter result array, got %v", TypeOf(filtered))
	}
	if filtered.ArrayLen() != 2 {
		t.Fatalf("expected length 2, got %d", filtered.ArrayLen())
	}
	if ArrayGet(filtered, 0).Long() != 2 || ArrayGet(filtered, 1).Long() != 4 {
		t.Fatalf("unexpected filtered values: %#v", filtered.ArrayValues())
	}
}

func TestFilterList(t *testing.T) {
	keepPositive := NewFunction(func(args ...Value) Value {
		if Gt(args[0], NewLong(0)) {
			return NewLong(1)
		}
		return NilValue
	})

	filtered := Filter(keepPositive, NewList(NewLong(-1), NewLong(0), NewLong(2), NewLong(3)))
	if filtered.ArrayLen() != 2 {
		t.Fatalf("expected length 2, got %d", filtered.ArrayLen())
	}
	if ArrayGet(filtered, 0).Long() != 2 || ArrayGet(filtered, 1).Long() != 3 {
		t.Fatalf("unexpected filtered values: %#v", filtered.ArrayValues())
	}
}

func TestReduceWithoutInitialValue(t *testing.T) {
	sum2 := NewFunction(func(args ...Value) Value {
		return Add(args[0], args[1])
	})

	reduced := Reduce(sum2, NewArray(NewLong(1), NewLong(2), NewLong(3), NewLong(4)))
	if reduced.Long() != 10 {
		t.Fatalf("expected 10, got %v", ValueToAny(reduced))
	}
}

func TestReduceWithInitialValue(t *testing.T) {
	sum2 := NewFunction(func(args ...Value) Value {
		return Add(args[0], args[1])
	})

	reduced := Reduce(sum2, NewLong(10), NewList(NewLong(1), NewLong(2), NewLong(3)))
	if reduced.Long() != 16 {
		t.Fatalf("expected 16, got %v", ValueToAny(reduced))
	}
}

func TestReduceWithoutInitialOnEmptyCallsFnZeroArity(t *testing.T) {
	zero := NewFunction(func(args ...Value) Value {
		if len(args) != 0 {
			t.Fatalf("expected zero-arity call, got %d args", len(args))
		}
		return NewLong(42)
	})

	reduced := Reduce(zero, NewArray())
	if reduced.Long() != 42 {
		t.Fatalf("expected 42, got %v", ValueToAny(reduced))
	}
}

func TestReduceWithInitialOnEmptyReturnsInitial(t *testing.T) {
	sum2 := NewFunction(func(args ...Value) Value {
		return Add(args[0], args[1])
	})

	reduced := Reduce(sum2, NewLong(7), NewList())
	if reduced.Long() != 7 {
		t.Fatalf("expected 7, got %v", ValueToAny(reduced))
	}
}

func TestReduceArityPanics(t *testing.T) {
	assertPanics(t, func() {
		Reduce(NewFunction(func(args ...Value) Value { return NilValue }))
	})
}

func TestReduceWithBuiltinPlusFunction(t *testing.T) {
	reduced := Reduce(BuiltinFunction("+"), NewArray(NewLong(1), NewLong(2), NewLong(3), NewLong(4)))
	if reduced.Long() != 10 {
		t.Fatalf("expected 10, got %v", ValueToAny(reduced))
	}
}

func TestReduceWithBuiltinPlusOnEmptyReturnsZero(t *testing.T) {
	reduced := Reduce(BuiltinFunction("+"), NewArray())
	if !isType[longImpl](reduced) || reduced.Long() != 0 {
		t.Fatalf("expected 0 from reduce + on empty collection, got %#v", reduced)
	}
}

func TestReduceWithoutInitialOnSingletonReturnsElement(t *testing.T) {
	reduced := Reduce(BuiltinFunction("+"), NewArray(NewLong(9)))
	if !isType[longImpl](reduced) || reduced.Long() != 9 {
		t.Fatalf("expected singleton reduce to return the element, got %#v", reduced)
	}
}

func TestLazyListPeekAndConsume(t *testing.T) {
	values := []Value{NewLong(10), NewLong(11), NewLong(12)}
	index := 0
	lazy := NewLazyList(func() (Value, bool) {
		if index >= len(values) {
			return Value{}, false
		}
		value := values[index]
		index++
		return value, true
	})

	if got := First(lazy); got.Long() != 10 {
		t.Fatalf("expected first 10, got %v", ValueToAny(got))
	}
	if got := First(lazy); got.Long() != 10 {
		t.Fatalf("expected repeated first to stay 10, got %v", ValueToAny(got))
	}
	if index != 1 {
		t.Fatalf("expected producer to run once after peek, got %d", index)
	}

	lazy = Rest(lazy)
	if got := First(lazy); got.Long() != 11 {
		t.Fatalf("expected second value 11, got %v", ValueToAny(got))
	}
}

func TestFirstStringReturnsFirstRune(t *testing.T) {
	if got := First(NilValue); !IsNil(got) {
		t.Fatalf("expected nil for first nil, got %#v", got)
	}
	if got := First(NewString("hello")); !isType[*stringImpl](got) || got.StringValue() != "h" {
		t.Fatalf("expected first string rune, got %#v", got)
	}
	if got := First(NewString("")); !IsNil(got) {
		t.Fatalf("expected nil for empty string, got %#v", got)
	}
}

func TestFirstRestAndVecForSet(t *testing.T) {
	if got := First(NewSet()); !IsNil(got) {
		t.Fatalf("expected nil for empty set, got %#v", got)
	}
	if got := Rest(NewSet()); !isType[*arrayImpl](got) || got.ArrayLen() != 0 {
		t.Fatalf("expected empty array rest for empty set, got %#v", got)
	}

	one := NewSet(NewLong(10))
	if got := First(one); !isType[longImpl](got) || got.Long() != 10 {
		t.Fatalf("expected singleton set first 10, got %#v", got)
	}
	if got := Rest(one); !isType[*arrayImpl](got) || got.ArrayLen() != 0 {
		t.Fatalf("expected empty rest for singleton set, got %#v", got)
	}
	if got := SeqFirst(one); !isType[longImpl](got) || got.Long() != 10 {
		t.Fatalf("expected SeqFirst 10, got %#v", got)
	}

	two := NewSet(NewLong(1), NewLong(2))
	first := First(two)
	rest := Rest(two)
	if !isType[longImpl](first) || (first.Long() != 1 && first.Long() != 2) {
		t.Fatalf("expected set first to be 1 or 2, got %#v", first)
	}
	if !isType[*arrayImpl](rest) || rest.ArrayLen() != 1 {
		t.Fatalf("expected set rest length 1, got %#v", rest)
	}
	if Eq(ArrayGet(rest, 0), first) {
		t.Fatalf("rest should not contain first: first=%#v rest=%#v", first, rest)
	}
	if got := Vec(two); !isType[*arrayImpl](got) || got.ArrayLen() != 2 {
		t.Fatalf("expected vec of set length 2, got %#v", got)
	}
}

func TestDoAllRealizesLazyList(t *testing.T) {
	count := int64(0)
	lazy := NewLazyList(func() (Value, bool) {
		if count >= 3 {
			return Value{}, false
		}
		count++
		return NewLong(count), true
	})

	realized := DoAll(lazy)
	if !isType[*arrayImpl](realized) || realized.ArrayLen() != 3 {
		t.Fatalf("unexpected doall result: %#v", realized)
	}
	if got := realized.ArrayValues(); got[0].Long() != 1 || got[2].Long() != 3 {
		t.Fatalf("unexpected realized values: %#v", realized.ArrayValues())
	}
}

func TestSomeReturnsFirstTruthyResult(t *testing.T) {
	fn := NewFunction(func(args ...Value) Value {
		if args[0].Long() > 2 {
			return args[0]
		}
		return NilValue
	})

	if got := Some(fn, NewArray(NewLong(1), NewLong(2), NewLong(3), NewLong(4))); !isType[longImpl](got) || got.Long() != 3 {
		t.Fatalf("unexpected some result: %#v", got)
	}
	if got := Some(fn, NilValue); !IsNil(got) {
		t.Fatalf("expected nil for nil collection, got %#v", got)
	}
}

func TestSetRealizesSequenceIntoSet(t *testing.T) {
	seen := 0
	lazy := NewLazyList(func() (Value, bool) {
		if seen >= 4 {
			return Value{}, false
		}
		seen++
		return NewLong(int64(seen % 2)), true
	})

	s := Set(lazy)
	if !isType[*setImpl](s) {
		t.Fatalf("expected set result, got %v", TypeOf(s))
	}
	if s.SetLen() != 2 {
		t.Fatalf("expected 2 unique values, got %d", s.SetLen())
	}
}

func TestConjAcrossCollectionTypes(t *testing.T) {
	list := Conj(NewList(NewLong(2), NewLong(3)), NewLong(1))
	if got := list.String(); got != "(1 2 3)" {
		t.Fatalf("unexpected list conj result: %q", got)
	}

	array := Conj(NewArray(NewLong(1), NewLong(2)), NewLong(3), NewLong(4))
	if got := array.String(); got != "[1 2 3 4]" {
		t.Fatalf("unexpected array conj result: %q", got)
	}

	set := Conj(NewSet(NewLong(1)), NewLong(2), NewLong(2))
	if got := set.String(); got != "#{1 2}" {
		t.Fatalf("unexpected set conj result: %q", got)
	}

	m := Conj(NewMap(NewKeyword("a"), NewLong(1)), NewArray(NewKeyword("b"), NewLong(2)))
	if got := m.String(); got != "{:a 1 :b 2}" {
		t.Fatalf("unexpected map conj result: %q", got)
	}

	merged := Conj(
		NewMap(NewKeyword("a"), NewLong(1)),
		NewMap(NewKeyword("b"), NewLong(2), NewKeyword("c"), NewLong(3)),
	)
	if got := merged.String(); got != "{:a 1 :b 2 :c 3}" {
		t.Fatalf("unexpected map-merge conj result: %q", got)
	}
}

func TestNotEmptyAcrossContainerTypes(t *testing.T) {
	if got := NotEmpty(NewList()); !IsNil(got) {
		t.Fatalf("expected nil for empty list, got %#v", got)
	}
	if got := NotEmpty(NewList(NewLong(1))); !isType[*listImpl](got) {
		t.Fatalf("expected list for non-empty list, got %#v", got)
	}
	if got := NotEmpty(NewArray()); !IsNil(got) {
		t.Fatalf("expected nil for empty array, got %#v", got)
	}
	if got := NotEmpty(NewArray(NewLong(1))); !isType[*arrayImpl](got) {
		t.Fatalf("expected array for non-empty array, got %#v", got)
	}
	if got := NotEmpty(NewMap()); !IsNil(got) {
		t.Fatalf("expected nil for empty map, got %#v", got)
	}
	if got := NotEmpty(NewMap(NewKeyword("a"), NewLong(1))); !isMap(got) {
		t.Fatalf("expected map for non-empty map, got %#v", got)
	}
	if got := NotEmpty(NewSet()); !IsNil(got) {
		t.Fatalf("expected nil for empty set, got %#v", got)
	}
	if got := NotEmpty(NewSet(NewLong(1))); !isType[*setImpl](got) {
		t.Fatalf("expected set for non-empty set, got %#v", got)
	}
	if got := NotEmpty(NewString("")); !IsNil(got) {
		t.Fatalf("expected nil for empty string, got %#v", got)
	}
	if got := NotEmpty(NewString("x")); !isType[*stringImpl](got) || got.StringValue() != "x" {
		t.Fatalf("expected string for non-empty string, got %#v", got)
	}

	emptyLazy := NewLazyList(func() (Value, bool) { return Value{}, false })
	if got := NotEmpty(emptyLazy); !IsNil(got) {
		t.Fatalf("expected nil for empty lazy list, got %#v", got)
	}
	nonEmptyLazy := NewLazyList(func() (Value, bool) { return NewLong(1), true })
	if got := NotEmpty(nonEmptyLazy); !isType[*lazyListImpl](got) {
		t.Fatalf("expected lazy list for non-empty lazy list, got %#v", got)
	}
}

func TestEmptyPredicatesAcrossContainerTypes(t *testing.T) {
	if !IsEmpty(NilValue) {
		t.Fatal("expected nil to be empty")
	}
	if !IsEmpty(NewArray()) || IsEmpty(NewArray(NewLong(1))) {
		t.Fatal("unexpected empty? results for arrays")
	}
	if !IsEmpty(NewList()) || IsEmpty(NewList(NewLong(1))) {
		t.Fatal("unexpected empty? results for lists")
	}
	if !IsEmpty(NewMap()) || IsEmpty(NewMap(NewKeyword("a"), NewLong(1))) {
		t.Fatal("unexpected empty? results for maps")
	}
	if !IsEmpty(NewSet()) || IsEmpty(NewSet(NewLong(1))) {
		t.Fatal("unexpected empty? results for sets")
	}
	if !IsEmpty(NewString("")) || IsEmpty(NewString("x")) {
		t.Fatal("unexpected empty? results for strings")
	}
}

func TestNilPredicate(t *testing.T) {
	if !IsNil(NilValue) {
		t.Fatal("expected nil? true for nil")
	}
	if IsNil(NewLong(1)) {
		t.Fatal("expected nil? false for non-nil value")
	}
}

func TestCountAcrossContainerTypes(t *testing.T) {
	// Test list
	if got := Count(NewList()); got != 0 {
		t.Fatalf("expected count 0 for empty list, got %d", got)
	}
	if got := Count(NewList(NewLong(1), NewLong(2), NewLong(3))); got != 3 {
		t.Fatalf("expected count 3 for list, got %d", got)
	}

	// Test array
	if got := Count(NewArray()); got != 0 {
		t.Fatalf("expected count 0 for empty array, got %d", got)
	}
	if got := Count(NewArray(NewLong(1), NewLong(2), NewLong(3))); got != 3 {
		t.Fatalf("expected count 3 for array, got %d", got)
	}

	// Test map
	if got := Count(NewMap()); got != 0 {
		t.Fatalf("expected count 0 for empty map, got %d", got)
	}
	if got := Count(NewMap(NewKeyword("a"), NewLong(1), NewKeyword("b"), NewLong(2))); got != 2 {
		t.Fatalf("expected count 2 for map, got %d", got)
	}

	// Test set
	if got := Count(NewSet()); got != 0 {
		t.Fatalf("expected count 0 for empty set, got %d", got)
	}
	if got := Count(NewSet(NewLong(1), NewLong(2), NewLong(3))); got != 3 {
		t.Fatalf("expected count 3 for set, got %d", got)
	}

	// Test string
	if got := Count(NewString("")); got != 0 {
		t.Fatalf("expected count 0 for empty string, got %d", got)
	}
	if got := Count(NewString("hello")); got != 5 {
		t.Fatalf("expected count 5 for string 'hello', got %d", got)
	}
	if got := Count(NewString("a😀")); got != 2 {
		t.Fatalf("expected rune count 2 for string 'a😀', got %d", got)
	}

	// Test nil
	if got := Count(NilValue); got != 0 {
		t.Fatalf("expected count 0 for nil, got %d", got)
	}

	// Test lazy list
	count := 0
	lazySeq := NewLazyList(func() (Value, bool) {
		if count >= 3 {
			return Value{}, false
		}
		value := NewLong(int64(count + 1))
		count++
		return value, true
	})
	if got := Count(lazySeq); got != 3 {
		t.Fatalf("expected count 3 for lazy list, got %d", got)
	}
}

func TestMapOverFiniteRangeIsEager(t *testing.T) {
	src := Range(NewLong(1), NewLong(4))
	double := NewFunction(func(args ...Value) Value { return Add(args[0], args[0]) })
	mapped := Map(double, src)
	if !isType[*arrayImpl](mapped) {
		t.Fatalf("expected mapped finite range array, got %v", TypeOf(mapped))
	}
	if mapped.String() != "[2 4 6]" {
		t.Fatalf("unexpected mapped finite range: %s", mapped.String())
	}
	if got := First(src); got.Long() != 1 {
		t.Fatalf("expected source range unchanged after map, got %v", ValueToAny(got))
	}
}

func TestMapOverInfiniteRangeStaysLazy(t *testing.T) {
	double := NewFunction(func(args ...Value) Value { return Add(args[0], args[0]) })
	mapped := Map(double, Range())
	if !isType[*lazyListImpl](mapped) {
		t.Fatalf("expected mapped infinite range to stay lazy, got %v", TypeOf(mapped))
	}
	if got := First(mapped); got.Long() != 0 {
		t.Fatalf("expected first mapped value 0, got %v", ValueToAny(got))
	}
	if got := First(Rest(mapped)); got.Long() != 2 {
		t.Fatalf("expected second mapped value 2, got %v", ValueToAny(got))
	}
}

func TestConcatReturnsLazySequenceAcrossInputs(t *testing.T) {
	joined := Concat(
		NewArray(NewLong(1), NewLong(2)),
		NilValue,
		NewList(NewLong(3)),
		Repeat(NewLong(2), NewLong(4)),
	)
	if !isType[*lazyListImpl](joined) {
		t.Fatalf("expected lazy list from concat, got %v", TypeOf(joined))
	}
	if got := (Take(NewLong(5), joined)).String(); got != "[1 2 3 4 4]" {
		t.Fatalf("unexpected concat values: %q", got)
	}
	if got := First(Concat()); !IsNil(got) {
		t.Fatalf("expected nil from empty concat head, got %#v", got)
	}
}

func TestIntoFoldsSequenceIntoCollection(t *testing.T) {
	got := Into(NewMap(), NewArray(NewArray(NewKeyword("a"), NewLong(1))))
	if !isMap(got) || got.String() != "{:a 1}" {
		t.Fatalf("unexpected into result: %#v", got)
	}
}

func TestMapCatFlattensMappedSequences(t *testing.T) {
	joined := MapCat(NewFunction(func(args ...Value) Value {
		return NewArray(args[0], Add(args[0], NewLong(10)))
	}), NewArray(NewLong(1), NewLong(2)))
	if !isType[*lazyListImpl](joined) {
		t.Fatalf("expected lazy list from mapcat, got %v", TypeOf(joined))
	}
	if got := (Take(NewLong(4), joined)).String(); got != "[1 11 2 12]" {
		t.Fatalf("unexpected mapcat values: %q", got)
	}
}

func TestSeqReturnsNilForEmptyAndCollectionForNonEmpty(t *testing.T) {
	if got := Seq(NewArray()); !IsNil(got) {
		t.Fatalf("expected nil for empty array seq, got %#v", got)
	}
	if got := Seq(NewList(NewLong(1))); !isType[*listImpl](got) {
		t.Fatalf("expected non-empty list from seq, got %#v", got)
	}
	if got := Seq(NewString("x")); !isType[*arrayImpl](got) || got.ArrayLen() != 1 || ArrayGet(got, 0).StringValue() != "x" {
		t.Fatalf("expected character array from seq(string), got %#v", got)
	}
}

func TestVecRealizesCollectionsToArray(t *testing.T) {
	if got := Vec(NewList(NewLong(1), NewLong(2))); !isType[*arrayImpl](got) || got.String() != "[1 2]" {
		t.Fatalf("unexpected vec result for list: %#v", got)
	}
	if got := Vec(NilValue); !isType[*arrayImpl](got) || got.ArrayLen() != 0 {
		t.Fatalf("unexpected vec result for nil: %#v", got)
	}
	if got := Vec(NewString("a😀")); !isType[*arrayImpl](got) || got.ArrayLen() != 2 || ArrayGet(got, 0).StringValue() != "a" || ArrayGet(got, 1).StringValue() != "😀" {
		t.Fatalf("unexpected vec result for string: %#v", got)
	}
}

func TestMapOverStringYieldsCharacterValues(t *testing.T) {
	identity := NewFunction(func(args ...Value) Value { return args[0] })
	got := Map(identity, NewString("ab😀"))
	if !isType[*arrayImpl](got) || got.ArrayLen() != 3 {
		t.Fatalf("unexpected map result for string: %#v", got)
	}
	if ArrayGet(got, 0).StringValue() != "a" || ArrayGet(got, 1).StringValue() != "b" || ArrayGet(got, 2).StringValue() != "😀" {
		t.Fatalf("unexpected mapped characters: %s", got.String())
	}
}

func TestFormatFormatsValues(t *testing.T) {
	if got := Format("%.2f", NewDouble(12.345)); got != "12.35" {
		t.Fatalf("unexpected format result: %q", got)
	}
}

func TestApplySpreadsFinalSequence(t *testing.T) {
	if got := Apply(BuiltinFunction("+"), NewArray(NewLong(1), NewLong(2), NewLong(3))); !isType[longImpl](got) || got.Long() != 6 {
		t.Fatalf("unexpected apply result: %#v", got)
	}
	if got := Apply(BuiltinFunction("+"), NewLong(10), NewArray(NewLong(1), NewLong(2))); !isType[longImpl](got) || got.Long() != 13 {
		t.Fatalf("unexpected apply with prefix args result: %#v", got)
	}
}

func TestApplyFnCallsGoFunctionDirectly(t *testing.T) {
	if got := ApplyFn(FnMax, NewArray(NewLong(1), NewLong(3), NewLong(2))); !isType[longImpl](got) || got.Long() != 3 {
		t.Fatalf("unexpected ApplyFn max result: %#v", got)
	}
	if got := ApplyFn(FnPlus, NewLong(10), NewArray(NewLong(1), NewLong(2))); !isType[longImpl](got) || got.Long() != 13 {
		t.Fatalf("unexpected ApplyFn plus result: %#v", got)
	}
}

func TestSortBySortsSequences(t *testing.T) {
	identity := NewFunction(func(args ...Value) Value { return args[0] })
	asc := SortBy(identity, NewList(NewLong(3), NewLong(1), NewLong(2)))
	if got := asc.String(); got != "[1 2 3]" {
		t.Fatalf("unexpected ascending sort-by result: %q", got)
	}

	desc := SortBy(identity, BuiltinFunction(">"), NewArray(NewLong(3), NewLong(1), NewLong(2)))
	if got := desc.String(); got != "[3 2 1]" {
		t.Fatalf("unexpected descending sort-by result: %q", got)
	}

	withNil := SortBy(
		NewKeyword("age"),
		NewArray(
			NewMap(NewKeyword("name"), NewString("A"), NewKeyword("age"), NilValue),
			NewMap(NewKeyword("name"), NewString("B"), NewKeyword("age"), NewLong(30)),
			NewMap(NewKeyword("name"), NewString("C"), NewKeyword("age"), NewLong(20)),
		),
	)
	if !isType[*arrayImpl](withNil) || withNil.ArrayLen() != 3 {
		t.Fatalf("unexpected sort-by nil-key shape: %#v", withNil)
	}
	first := ArrayGet(withNil, 0)
	second := ArrayGet(withNil, 1)
	third := ArrayGet(withNil, 2)
	if !IsNil(Get(first, NewKeyword("age"))) ||
		Get(second, NewKeyword("age")).Long() != 20 ||
		Get(third, NewKeyword("age")).Long() != 30 {
		t.Fatalf("unexpected sort-by nil-key ordering: %s", withNil.String())
	}
}

func TestFilterConsumesLazyList(t *testing.T) {
	count := 0
	source := NewLazyList(func() (Value, bool) {
		if count >= 6 {
			return Value{}, false
		}
		value := count
		count++
		return NewLong(int64(value)), true
	})

	keepEven := NewFunction(func(args ...Value) Value {
		if Mod(args[0], NewLong(2)).Long() == 0 {
			return NewBool(true)
		}
		return NilValue
	})

	filtered := Filter(keepEven, source)
	if filtered.ArrayLen() != 3 {
		t.Fatalf("expected 3 filtered values, got %d", filtered.ArrayLen())
	}
	if ArrayGet(filtered, 0).Long() != 0 || ArrayGet(filtered, 2).Long() != 4 {
		t.Fatalf("unexpected filtered values: %#v", filtered.ArrayValues())
	}
	if got := First(source); got.Long() != 0 {
		t.Fatalf("expected source lazy list unchanged after filter, got %v", ValueToAny(got))
	}
}

func TestReduceConsumesLazyList(t *testing.T) {
	count := int64(0)
	source := NewLazyList(func() (Value, bool) {
		if count >= 5 {
			return Value{}, false
		}
		count++
		return NewLong(count), true
	})

	reduced := Reduce(BuiltinFunction("+"), source)
	if reduced.Long() != 15 {
		t.Fatalf("expected reduced value 15, got %v", ValueToAny(reduced))
	}
	if got := First(source); got.Long() != 1 {
		t.Fatalf("expected source lazy list unchanged after reduce, got %v", ValueToAny(got))
	}
}

func TestReduceWithInitialOnNilCollection(t *testing.T) {
	add := NewFunction(func(args ...Value) Value { return Add(args[0], args[1]) })
	if got := Reduce(add, NewLong(10), NilValue); !isType[longImpl](got) || got.Long() != 10 {
		t.Fatalf("expected reduce init value for nil collection, got %#v", got)
	}
}

func TestDoAllIsNoOpForNonSequences(t *testing.T) {
	m := NewMap(NewKeyword("a"), NewLong(1))
	if got := DoAll(m); !isMap(got) || got.String() != "{:a 1}" {
		t.Fatalf("expected doall map no-op, got %#v", got)
	}
	if got := DoAll(NewString("abc")); !isType[*stringImpl](got) || got.StringValue() != "abc" {
		t.Fatalf("expected doall string no-op, got %#v", got)
	}
}

func TestRangeTwoArgsShortReturnsRange(t *testing.T) {
	r := Range(NewLong(3), NewLong(7))
	if !isType[*rangeImpl](r) {
		t.Fatalf("expected range for short range, got %v", TypeOf(r))
	}
	got := DoAll(r)
	if got.String() != "[3 4 5 6]" {
		t.Fatalf("unexpected short range values: %s", got.String())
	}
}

func TestRangeOneArgIsExclusiveEndFromZero(t *testing.T) {
	r := Range(NewLong(4))
	if !isType[*rangeImpl](r) {
		t.Fatalf("expected range for (range 4), got %v", TypeOf(r))
	}
	if got := DoAll(r).String(); got != "[0 1 2 3]" {
		t.Fatalf("expected (range 4) to be 0..3, got %q", got)
	}
	empty := Range(NewLong(0))
	if !isType[*arrayImpl](empty) || empty.ArrayLen() != 0 {
		t.Fatalf("expected empty (range 0), got %#v", empty)
	}
	neg := Range(NewLong(-3))
	if !isType[*arrayImpl](neg) || neg.ArrayLen() != 0 {
		t.Fatalf("expected empty (range -3), got %#v", neg)
	}
}

func TestRangeNoArgsReturnsInfiniteRangeFromZero(t *testing.T) {
	r := Range()
	if !isType[*rangeImpl](r) {
		t.Fatalf("expected range for zero-arg range, got %v", TypeOf(r))
	}
	for i := int64(0); i < 5; i++ {
		got := First(r)
		if got.Long() != i {
			t.Fatalf("expected %d from range, got %v", i, ValueToAny(got))
		}
		r = Rest(r)
	}
}

func TestRangeLargeReturnsTerminatingRange(t *testing.T) {
	r := Range(NewLong(1), NewLong(1002))
	if !isType[*rangeImpl](r) {
		t.Fatalf("expected range for large range, got %v", TypeOf(r))
	}
	for i := int64(1); i < 1002; i++ {
		got := First(r)
		if got.Long() != i {
			t.Fatalf("expected %d from large range, got %v", i, ValueToAny(got))
		}
		r = Rest(r)
	}
	if got := First(r); !IsNil(got) {
		t.Fatalf("expected exhausted large range to be nil, got %v", ValueToAny(got))
	}
}

func TestRangeThreeArgStep(t *testing.T) {
	if got := (DoAll(Range(NewLong(0), NewLong(8), NewLong(2)))).String(); got != "[0 2 4 6]" {
		t.Fatalf("unexpected positive step range: %q", got)
	}
	if got := (DoAll(Range(NewLong(5), NewLong(1), NewLong(-1)))).String(); got != "[5 4 3 2]" {
		t.Fatalf("unexpected negative step range: %q", got)
	}
	if got := Range(NewLong(4), NewLong(1)); !isType[*arrayImpl](got) || got.ArrayLen() != 0 {
		t.Fatalf("expected empty inverted range, got %#v", got)
	}
	zeros := Range(NewLong(1), NewLong(10), NewLong(0))
	if !isType[*lazyListImpl](zeros) {
		t.Fatalf("expected infinite zero-step range, got %v", TypeOf(zeros))
	}
	if got := First(zeros); got.Long() != 1 {
		t.Fatalf("expected repeated start, got %v", ValueToAny(got))
	}
}

func TestRangePanicsOnInvalidArgs(t *testing.T) {
	assertPanics(t, func() { Range(NewLong(1), NewLong(2), NewLong(3), NewLong(4)) })
	assertPanics(t, func() { Range(NewDouble(1.5)) })
}

func TestRangeCountUsesRemainingHint(t *testing.T) {
	if got := Count(Range(NewLong(4))); got != 4 {
		t.Fatalf("count (range 4) = %d", got)
	}
	if got := Count(Rest(Range(NewLong(4)))); got != 3 {
		t.Fatalf("count (rest (range 4)) = %d", got)
	}
	if got := Count(Range(NewLong(0))); got != 0 {
		t.Fatalf("count (range 0) = %d", got)
	}
}

func TestRangeWalkIsLinear(t *testing.T) {
	const n = 50000
	r := Range(NewLong(n))
	start := time.Now()
	for i := int64(0); i < n; i++ {
		got := First(r)
		if !isType[longImpl](got) || got.Long() != i {
			t.Fatalf("at %d got %v", i, ValueToAny(got))
		}
		r = Rest(r)
	}
	if got := First(r); !IsNil(got) {
		t.Fatalf("expected exhausted range, got %v", ValueToAny(got))
	}
	if elapsed := time.Since(start); elapsed > 2*time.Second {
		t.Fatalf("walk of %d took %s; first/rest is likely still quadratic", n, elapsed)
	}
}

func TestFnRangeMatchesRange(t *testing.T) {
	got := DoAll(FnRange(NewLong(3)))
	if got.String() != "[0 1 2]" {
		t.Fatalf("FnRange(3) = %s", got.String())
	}
}

func TestRepeatOneArgReturnsInfiniteLazyList(t *testing.T) {
	r := Repeat(NewLong(7))
	if !isType[*lazyListImpl](r) {
		t.Fatalf("expected lazy list for one-arg repeat, got %v", TypeOf(r))
	}
	for i := 0; i < 4; i++ {
		got := First(r)
		if !isType[longImpl](got) || got.Long() != 7 {
			t.Fatalf("expected repeated 7, got %#v", got)
		}
		r = Rest(r)
	}
}

func TestRepeatTwoArgsReturnsFiniteLazyList(t *testing.T) {
	r := Repeat(NewLong(3), NewString("x"))
	if !isType[*lazyListImpl](r) {
		t.Fatalf("expected lazy list for two-arg repeat, got %v", TypeOf(r))
	}
	if got := (Take(NewLong(4), r)).String(); got != `["x" "x" "x"]` {
		t.Fatalf("unexpected repeat values: %q", got)
	}
	if got := First(Drop(NewLong(3), r)); !IsNil(got) {
		t.Fatalf("expected finite repeat to exhaust, got %#v", got)
	}
}

func TestRepeatPanicsOnInvalidArgs(t *testing.T) {
	assertPanics(t, func() { Repeat() })
	assertPanics(t, func() { Repeat(NewLong(1), NewLong(2), NewLong(3)) })
	assertPanics(t, func() { Repeat(NewDouble(1.5), NewLong(7)) })
}

func TestTakeOnArrayListLazy(t *testing.T) {
	fromArray := Take(NewLong(2), NewArray(NewLong(1), NewLong(2), NewLong(3)))
	if got := fromArray.ArrayValues(); len(got) != 2 || got[0].Long() != 1 || got[1].Long() != 2 {
		t.Fatalf("unexpected take array values: %#v", got)
	}

	fromList := Take(NewLong(3), NewList(NewLong(4), NewLong(5), NewLong(6), NewLong(7)))
	if got := fromList.ArrayValues(); len(got) != 3 || got[0].Long() != 4 || got[2].Long() != 6 {
		t.Fatalf("unexpected take list values: %#v", got)
	}

	lazy := Range(NewLong(10), NewLong(2000))
	fromLazy := Take(NewLong(4), lazy)
	if got := fromLazy.ArrayValues(); len(got) != 4 || got[0].Long() != 10 || got[3].Long() != 13 {
		t.Fatalf("unexpected take lazy values: %#v", got)
	}
	if got := First(lazy); got.Long() != 10 {
		t.Fatalf("expected original lazy unchanged after take, got %v", ValueToAny(got))
	}
}

func TestDropOnArrayListLazy(t *testing.T) {
	arrayDropped := Drop(NewLong(2), NewArray(NewLong(1), NewLong(2), NewLong(3), NewLong(4)))
	if !isType[*arrayImpl](arrayDropped) {
		t.Fatalf("expected array from drop array, got %v", TypeOf(arrayDropped))
	}
	if got := arrayDropped.ArrayValues(); len(got) != 2 || got[0].Long() != 3 || got[1].Long() != 4 {
		t.Fatalf("unexpected drop array values: %#v", got)
	}

	listDropped := Drop(NewLong(1), NewList(NewLong(5), NewLong(6), NewLong(7)))
	if !isType[*listImpl](listDropped) {
		t.Fatalf("expected list from drop list, got %v", TypeOf(listDropped))
	}
	if got := listDropped.ListValues(); len(got) != 2 || got[0].Long() != 6 || got[1].Long() != 7 {
		t.Fatalf("unexpected drop list values: %#v", got)
	}

	src := Range(NewLong(20), NewLong(2000))
	dropped := Drop(NewLong(3), src)
	if !isType[*rangeImpl](dropped) {
		t.Fatalf("expected range from drop range, got %v", TypeOf(dropped))
	}
	if got := First(dropped); got.Long() != 23 {
		t.Fatalf("expected dropped range head 23, got %v", ValueToAny(got))
	}
	if got := First(src); got.Long() != 20 {
		t.Fatalf("expected original range unchanged after drop, got %v", ValueToAny(got))
	}
}

func TestTakeDropPanicsOnInvalidCount(t *testing.T) {
	assertPanics(t, func() { Take(NewLong(-1), NewArray()) })
	assertPanics(t, func() { Drop(NewLong(-1), NewArray()) })
	assertPanics(t, func() { Take(NewDouble(1.2), NewArray()) })
	assertPanics(t, func() { Drop(NewDouble(1.2), NewArray()) })
}

func TestNthAndSlowNth(t *testing.T) {
	if got := Nth(NewArray(NewLong(10), NewLong(20), NewLong(30)), NewLong(1)); got.Long() != 20 {
		t.Fatalf("unexpected nth array result: %v", ValueToAny(got))
	}
	if got := Nth(NewString("a😀"), NewLong(1)); got.StringValue() != "😀" {
		t.Fatalf("unexpected nth string result: %q", got.StringValue())
	}
	if got := Nth(NewArray(NewLong(1)), NewLong(5), NewKeyword("missing")); !isType[*symbolImpl](got) || !got.SymbolObject().IsKeyword || got.SymbolObject().Name != "missing" {
		t.Fatalf("expected nth not-found keyword, got %v", ValueToAny(got))
	}
	assertPanics(t, func() { Nth(NewList(NewLong(1), NewLong(2)), NewLong(1)) })

	if got := SlowNth(NewList(NewLong(1), NewLong(2), NewLong(3)), NewLong(1)); got.Long() != 2 {
		t.Fatalf("unexpected slow-nth list result: %v", ValueToAny(got))
	}
	if got := SlowNth(Range(NewLong(10), NewLong(2000)), NewLong(2)); got.Long() != 12 {
		t.Fatalf("unexpected slow-nth lazy result: %v", ValueToAny(got))
	}
	if got := SlowNth(NewList(NewLong(1)), NewLong(8), NewKeyword("missing")); !isType[*symbolImpl](got) || !got.SymbolObject().IsKeyword || got.SymbolObject().Name != "missing" {
		t.Fatalf("expected slow-nth not-found keyword, got %v", ValueToAny(got))
	}
}

func TestSubs(t *testing.T) {
	if got := Subs(NewString("hello"), NewLong(1)); got.StringValue() != "ello" {
		t.Fatalf("unexpected 2-arg subs: %q", got.StringValue())
	}
	if got := Subs(NewString("hello"), NewLong(1), NewLong(4)); got.StringValue() != "ell" {
		t.Fatalf("unexpected 3-arg subs: %q", got.StringValue())
	}
	if got := Subs(NewString("a😀b"), NewLong(1), NewLong(2)); got.StringValue() != "😀" {
		t.Fatalf("unexpected rune subs: %q", got.StringValue())
	}
	if got := Subs(NewString("ab"), NewLong(2)); got.StringValue() != "" {
		t.Fatalf("expected empty end-of-string subs, got %q", got.StringValue())
	}
	assertPanics(t, func() { Subs(NewString("ab"), NewLong(3)) })
	assertPanics(t, func() { Subs(NewString("ab"), NewLong(1), NewLong(0)) })
}

func TestLazyListConcurrentRealizationIsSafe(t *testing.T) {
	source := Range(NewLong(1), NewLong(1200))

	var wg sync.WaitGroup
	errCh := make(chan string, 1)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			local := source
			for want := int64(1); want <= 300; want++ {
				value := First(local)
				if IsNil(value) {
					select {
					case errCh <- "unexpected nil while traversing lazy list":
					default:
					}
					return
				}
				if value.Long() != want {
					select {
					case errCh <- "unexpected value while traversing lazy list":
					default:
					}
					return
				}
				local = Rest(local)
			}
		}()
	}
	wg.Wait()
	select {
	case msg := <-errCh:
		t.Fatal(msg)
	default:
	}
}

func TestLazyListRestDoesNotMutateOriginal(t *testing.T) {
	r := Range(NewLong(4), NewLong(2000))
	if got := First(r); got.Long() != 4 {
		t.Fatalf("expected first r = 4, got %v", ValueToAny(got))
	}
	tail := Rest(r)
	if got := First(tail); got.Long() != 5 {
		t.Fatalf("expected first(rest r) = 5, got %v", ValueToAny(got))
	}
	if got := First(Rest(r)); got.Long() != 5 {
		t.Fatalf("expected first(rest r) = 5 from fresh rest, got %v", ValueToAny(got))
	}
	if got := First(r); got.Long() != 4 {
		t.Fatalf("expected first r to remain 4, got %v", ValueToAny(got))
	}
}

func TestLazyListConcurrentTailConsumptionIsSafe(t *testing.T) {
	source := Range(NewLong(1), NewLong(2200))

	var wg sync.WaitGroup
	errCh := make(chan string, 1)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			local := source
			for j := 0; j < 200; j++ {
				value := First(local)
				if IsNil(value) {
					select {
					case errCh <- "unexpected nil during concurrent tail walk":
					default:
					}
					return
				}
				local = Rest(local)
			}
		}()
	}
	wg.Wait()
	select {
	case msg := <-errCh:
		t.Fatal(msg)
	default:
	}
}
