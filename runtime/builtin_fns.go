package runtime

// FLAG builtin implementations as direct Go function pointers.

// FnPlus implements builtin +.
func FnPlus(args ...Value) Value {
	switch len(args) {
	case 0:
		return NewLong(0)
	case 1:
		if !args[0].impl.isNumeric(args[0]) {
			panic("+ expects numeric Value arguments")
		}
		return args[0]
	default:
		return foldNumericBuiltin("+", Add, args...)
	}
}

// FnMul implements builtin *.
func FnMul(args ...Value) Value {
	return foldNumericBuiltin("*", Mul, args...)
}

// FnSub implements builtin -.
func FnSub(args ...Value) Value {
	return foldNumericBuiltin("-", Sub, args...)
}

// FnDiv implements builtin /.
func FnDiv(args ...Value) Value {
	return foldNumericBuiltin("/", Div, args...)
}

// FnMod implements builtin %, mod.
func FnMod(args ...Value) Value {
	if len(args) != 2 {
		panic("mod expects exactly two arguments")
	}
	return Mod(args[0], args[1])
}

// FnQuot implements builtin quot.
func FnQuot(args ...Value) Value {
	if len(args) != 2 {
		panic("quot expects exactly two arguments")
	}
	return Quot(args[0], args[1])
}

// FnRem implements builtin rem.
func FnRem(args ...Value) Value {
	if len(args) != 2 {
		panic("rem expects exactly two arguments")
	}
	return Rem(args[0], args[1])
}

// FnCompare implements builtin compare.
func FnCompare(args ...Value) Value {
	if len(args) != 2 {
		panic("compare expects exactly two arguments")
	}
	return Compare(args[0], args[1])
}

// FnNumericEq implements builtin ==.
func FnNumericEq(args ...Value) Value {
	return NewBool(NumericEq(args...))
}

// FnEq implements builtin =.
func FnEq(args ...Value) Value {
	if len(args) < 2 {
		panic("= expects at least two arguments")
	}
	for i := 0; i < len(args)-1; i++ {
		if !Eq(args[i], args[i+1]) {
			return NewBool(false)
		}
	}
	return NewBool(true)
}

// FnLt implements builtin <.
func FnLt(args ...Value) Value {
	if len(args) < 2 {
		panic("< expects at least two arguments")
	}
	for i := 0; i < len(args)-1; i++ {
		if !Lt(args[i], args[i+1]) {
			return NewBool(false)
		}
	}
	return NewBool(true)
}

// FnLe implements builtin <=.
func FnLe(args ...Value) Value {
	if len(args) < 2 {
		panic("<= expects at least two arguments")
	}
	for i := 0; i < len(args)-1; i++ {
		if !Le(args[i], args[i+1]) {
			return NewBool(false)
		}
	}
	return NewBool(true)
}

// FnGt implements builtin >.
func FnGt(args ...Value) Value {
	if len(args) < 2 {
		panic("> expects at least two arguments")
	}
	for i := 0; i < len(args)-1; i++ {
		if !Gt(args[i], args[i+1]) {
			return NewBool(false)
		}
	}
	return NewBool(true)
}

// FnGe implements builtin >=.
func FnGe(args ...Value) Value {
	if len(args) < 2 {
		panic(">= expects at least two arguments")
	}
	for i := 0; i < len(args)-1; i++ {
		if !Ge(args[i], args[i+1]) {
			return NewBool(false)
		}
	}
	return NewBool(true)
}

// FnMax implements builtin max.
func FnMax(args ...Value) Value {
	if len(args) < 1 {
		panic("max expects at least one argument")
	}
	return Max(args...)
}

// FnMin implements builtin min.
func FnMin(args ...Value) Value {
	if len(args) < 1 {
		panic("min expects at least one argument")
	}
	return Min(args...)
}

// FnBitAnd implements builtin bit-and.
func FnBitAnd(args ...Value) Value {
	return BitAnd(args...)
}

// FnBitOr implements builtin bit-or.
func FnBitOr(args ...Value) Value {
	return BitOr(args...)
}

// FnBitXor implements builtin bit-xor.
func FnBitXor(args ...Value) Value {
	return BitXor(args...)
}

// FnBitNot implements builtin bit-not.
func FnBitNot(args ...Value) Value {
	if len(args) != 1 {
		panic("bit-not expects exactly one argument")
	}
	return BitNot(args[0])
}

// FnBitShiftLeft implements builtin bit-shift-left.
func FnBitShiftLeft(args ...Value) Value {
	if len(args) != 2 {
		panic("bit-shift-left expects exactly two arguments")
	}
	return BitShiftLeft(args[0], args[1])
}

// FnBitShiftRight implements builtin bit-shift-right.
func FnBitShiftRight(args ...Value) Value {
	if len(args) != 2 {
		panic("bit-shift-right expects exactly two arguments")
	}
	return BitShiftRight(args[0], args[1])
}

// FnUnsignedBitShiftRight implements builtin unsigned-bit-shift-right.
func FnUnsignedBitShiftRight(args ...Value) Value {
	if len(args) != 2 {
		panic("unsigned-bit-shift-right expects exactly two arguments")
	}
	return UnsignedBitShiftRight(args[0], args[1])
}

// FnBitTest implements builtin bit-test.
func FnBitTest(args ...Value) Value {
	if len(args) != 2 {
		panic("bit-test expects exactly two arguments")
	}
	return BitTest(args[0], args[1])
}

// FnBitSet implements builtin bit-set.
func FnBitSet(args ...Value) Value {
	if len(args) != 2 {
		panic("bit-set expects exactly two arguments")
	}
	return BitSet(args[0], args[1])
}

// FnBitClear implements builtin bit-clear.
func FnBitClear(args ...Value) Value {
	if len(args) != 2 {
		panic("bit-clear expects exactly two arguments")
	}
	return BitClear(args[0], args[1])
}

// FnBitFlip implements builtin bit-flip.
func FnBitFlip(args ...Value) Value {
	if len(args) != 2 {
		panic("bit-flip expects exactly two arguments")
	}
	return BitFlip(args[0], args[1])
}

// FnRandInt implements builtin rand-int.
func FnRandInt(args ...Value) Value {
	if len(args) != 1 {
		panic("rand-int expects exactly one argument")
	}
	return RandInt(args[0])
}

// FnRand implements builtin rand.
func FnRand(args ...Value) Value {
	return Rand(args...)
}

// FnRandNth implements builtin rand-nth.
func FnRandNth(args ...Value) Value {
	if len(args) != 1 {
		panic("rand-nth expects exactly one argument")
	}
	return RandNth(args[0])
}

// FnShuffle implements builtin shuffle.
func FnShuffle(args ...Value) Value {
	if len(args) != 1 {
		panic("shuffle expects exactly one argument")
	}
	return Shuffle(args[0])
}

// FnFirst implements builtin first, fist.
func FnFirst(args ...Value) Value {
	if len(args) != 1 {
		panic("first expects exactly one argument")
	}
	return First(args[0])
}

// FnPeek implements builtin peek.
func FnPeek(args ...Value) Value {
	if len(args) != 1 {
		panic("peek expects exactly one argument")
	}
	return Peek(args[0])
}

// FnPop implements builtin pop.
func FnPop(args ...Value) Value {
	if len(args) != 1 {
		panic("pop expects exactly one argument")
	}
	return Pop(args[0])
}

// FnRest implements builtin rest.
func FnRest(args ...Value) Value {
	if len(args) != 1 {
		panic("rest expects exactly one argument")
	}
	return Rest(args[0])
}

// FnNext implements builtin next.
func FnNext(args ...Value) Value {
	if len(args) != 1 {
		panic("next expects exactly one argument")
	}
	return Next(args[0])
}

// FnLast implements builtin last.
func FnLast(args ...Value) Value {
	if len(args) != 1 {
		panic("last expects exactly one argument")
	}
	return Last(args[0])
}

// FnReverse implements builtin reverse.
func FnReverse(args ...Value) Value {
	if len(args) != 1 {
		panic("reverse expects exactly one argument")
	}
	return Reverse(args[0])
}

// FnCons implements builtin cons.
func FnCons(args ...Value) Value {
	if len(args) != 2 {
		panic("cons expects exactly two arguments")
	}
	return Cons(args[0], args[1])
}

// FnTake implements builtin take.
func FnTake(args ...Value) Value {
	if len(args) != 2 {
		panic("take expects count and sequence")
	}
	return Take(args[0], args[1])
}

// FnDrop implements builtin drop.
func FnDrop(args ...Value) Value {
	if len(args) != 2 {
		panic("drop expects count and sequence")
	}
	return Drop(args[0], args[1])
}

// FnNth implements builtin nth.
func FnNth(args ...Value) Value {
	if len(args) != 2 && len(args) != 3 {
		panic("nth expects collection, index, and optional default")
	}
	if len(args) == 2 {
		return Nth(args[0], args[1])
	}
	return Nth(args[0], args[1], args[2])
}

// FnSlowNth implements builtin slow-nth.
func FnSlowNth(args ...Value) Value {
	if len(args) != 2 && len(args) != 3 {
		panic("slow-nth expects collection, index, and optional default")
	}
	if len(args) == 2 {
		return SlowNth(args[0], args[1])
	}
	return SlowNth(args[0], args[1], args[2])
}

// FnMap implements builtin map.
func FnMap(args ...Value) Value {
	if len(args) < 2 {
		panic("map expects function and at least one sequence")
	}
	return Map(args[0], args[1:]...)
}

// FnConcat implements builtin concat.
func FnConcat(args ...Value) Value {
	return Concat(args...)
}

// FnInto implements builtin into.
func FnInto(args ...Value) Value {
	if len(args) != 2 {
		panic("into expects collection and sequence")
	}
	return Into(args[0], args[1])
}

// FnSortBy implements builtin sort-by.
func FnSortBy(args ...Value) Value {
	if len(args) != 2 && len(args) != 3 {
		panic("sort-by expects key function, optional comparator, and collection")
	}
	return SortBy(args[0], args[1:]...)
}

// FnApply implements builtin apply.
func FnApply(args ...Value) Value {
	if len(args) < 2 {
		panic("apply expects function and at least one argument sequence")
	}
	return Apply(args[0], args[1:]...)
}

// FnPmap implements builtin pmap.
func FnPmap(args ...Value) Value {
	if len(args) < 2 {
		panic("pmap expects function and at least one sequence")
	}
	return PMap(args[0], args[1:]...)
}

// FnFilter implements builtin filter.
func FnFilter(args ...Value) Value {
	if len(args) != 2 {
		panic("filter expects function and one sequence")
	}
	return Filter(args[0], args[1])
}

// FnJuxt implements builtin juxt.
func FnJuxt(args ...Value) Value {
	return Juxt(args...)
}

// FnGroupBy implements builtin group-by.
func FnGroupBy(args ...Value) Value {
	if len(args) != 2 {
		panic("group-by expects function and collection")
	}
	return GroupBy(args[0], args[1])
}

// FnReduce implements builtin reduce.
func FnReduce(args ...Value) Value {
	if len(args) != 2 && len(args) != 3 {
		panic("reduce expects function and collection, or function, initial value, and collection")
	}
	return Reduce(args[0], args[1:]...)
}

// FnReduced implements builtin reduced.
func FnReduced(args ...Value) Value {
	if len(args) != 1 {
		panic("reduced expects exactly one argument")
	}
	return Reduced(args[0])
}

// FnReducedQ implements builtin reduced?.
func FnReducedQ(args ...Value) Value {
	if len(args) != 1 {
		panic("reduced? expects exactly one argument")
	}
	return NewBool(IsReduced(args[0]))
}

// FnEnsureReduced implements builtin ensure-reduced.
func FnEnsureReduced(args ...Value) Value {
	if len(args) != 1 {
		panic("ensure-reduced expects exactly one argument")
	}
	return EnsureReduced(args[0])
}

// FnUnreduced implements builtin unreduced.
func FnUnreduced(args ...Value) Value {
	if len(args) != 1 {
		panic("unreduced expects exactly one argument")
	}
	return Unreduced(args[0])
}

// FnNotEmpty implements builtin not-empty.
func FnNotEmpty(args ...Value) Value {
	if len(args) != 1 {
		panic("not-empty expects exactly one argument")
	}
	return NotEmpty(args[0])
}

// FnSeq implements builtin seq.
func FnSeq(args ...Value) Value {
	if len(args) != 1 {
		panic("seq expects exactly one argument")
	}
	return Seq(args[0])
}

// FnEmptyQ implements builtin empty?.
func FnEmptyQ(args ...Value) Value {
	if len(args) != 1 {
		panic("empty? expects exactly one argument")
	}
	return NewBool(IsEmpty(args[0]))
}

// FnNilQ implements builtin nil?.
func FnNilQ(args ...Value) Value {
	if len(args) != 1 {
		panic("nil? expects exactly one argument")
	}
	return NewBool(IsNil(args[0]))
}

// FnTypeOf implements builtin type-of.
func FnTypeOf(args ...Value) Value {
	if len(args) != 1 {
		panic("type-of expects exactly one argument")
	}
	return TypeOf(args[0])
}

// FnSet implements builtin set.
func FnSet(args ...Value) Value {
	if len(args) != 1 {
		panic("set expects exactly one argument")
	}
	return Set(args[0])
}

// FnVec implements builtin vec.
func FnVec(args ...Value) Value {
	if len(args) != 1 {
		panic("vec expects exactly one argument")
	}
	return Vec(args[0])
}

// FnConj implements builtin conj.
func FnConj(args ...Value) Value {
	if len(args) < 2 {
		panic("conj expects collection and at least one item")
	}
	return Conj(args[0], args[1:]...)
}

// FnContainsQ implements builtin contains?.
func FnContainsQ(args ...Value) Value {
	if len(args) != 2 {
		panic("contains? expects collection and key")
	}
	return NewBool(Contains(args[0], args[1]))
}

// FnSome implements builtin some.
func FnSome(args ...Value) Value {
	if len(args) != 2 {
		panic("some expects predicate and collection")
	}
	return Some(args[0], args[1])
}

// FnSeqQ implements builtin seq?.
func FnSeqQ(args ...Value) Value {
	if len(args) != 1 {
		panic("seq? expects exactly one argument")
	}
	return SeqPredicate(args[0])
}

// FnDoall implements builtin doall.
func FnDoall(args ...Value) Value {
	if len(args) != 1 {
		panic("doall expects exactly one argument")
	}
	return DoAll(args[0])
}

// FnDorun implements builtin dorun.
func FnDorun(args ...Value) Value {
	if len(args) != 1 {
		panic("dorun expects exactly one argument")
	}
	return DoRun(args[0])
}

// FnCount implements builtin count.
func FnCount(args ...Value) Value {
	if len(args) != 1 {
		panic("count expects exactly one argument")
	}
	return NewLong(int64(Count(args[0])))
}

// FnDouble implements builtin double.
func FnDouble(args ...Value) Value {
	if len(args) != 1 {
		panic("double expects exactly one argument")
	}
	return Double(args[0])
}

// FnNumerator implements builtin numerator.
func FnNumerator(args ...Value) Value {
	if len(args) != 1 {
		panic("numerator expects exactly one argument")
	}
	return Numerator(args[0])
}

// FnDenominator implements builtin denominator.
func FnDenominator(args ...Value) Value {
	if len(args) != 1 {
		panic("denominator expects exactly one argument")
	}
	return Denominator(args[0])
}

// FnPrint implements builtin print.
func FnPrint(args ...Value) Value {
	anys := make([]any, len(args))
	for i, arg := range args {
		anys[i] = arg
	}
	return Print(anys...)
}

// FnFormat implements builtin format.
func FnFormat(args ...Value) Value {
	if len(args) < 1 {
		panic("format expects a format string and optional values")
	}
	format := valueAsString(args[0])
	return NewString(Format(format, args[1:]...))
}

// FnSubs implements builtin subs.
func FnSubs(args ...Value) Value {
	if len(args) != 2 && len(args) != 3 {
		panic("subs expects 2 or 3 arguments")
	}
	if len(args) == 2 {
		return Subs(args[0], args[1])
	}
	return Subs(args[0], args[1], args[2])
}

// FnKeyword implements builtin keyword.
func FnKeyword(args ...Value) Value {
	if len(args) != 1 {
		panic("keyword expects exactly one argument")
	}
	return Keyword(args[0])
}

// FnGet implements builtin get.
func FnGet(args ...Value) Value {
	if len(args) != 2 && len(args) != 3 {
		panic("get expects map, key, and optional default")
	}
	if len(args) == 2 {
		return Get(args[0], args[1])
	}
	return Get(args[0], args[1], args[2])
}

// FnKeys implements builtin keys.
func FnKeys(args ...Value) Value {
	if len(args) != 1 {
		panic("keys expects exactly one argument")
	}
	return Keys(args[0])
}

// FnVals implements builtin vals.
func FnVals(args ...Value) Value {
	if len(args) != 1 {
		panic("vals expects exactly one argument")
	}
	return Vals(args[0])
}

// FnFind implements builtin find.
func FnFind(args ...Value) Value {
	if len(args) != 2 {
		panic("find expects map and key")
	}
	return Find(args[0], args[1])
}

// FnUnion implements builtin union.
func FnUnion(args ...Value) Value {
	return Union(args...)
}

// FnIntersection implements builtin intersection.
func FnIntersection(args ...Value) Value {
	return Intersection(args...)
}

// FnDifference implements builtin difference.
func FnDifference(args ...Value) Value {
	return Difference(args...)
}

// FnSubsetQ implements builtin subset?.
func FnSubsetQ(args ...Value) Value {
	if len(args) != 2 {
		panic("subset? expects exactly two sets")
	}
	return NewBool(Subset(args[0], args[1]))
}

// FnSupersetQ implements builtin superset?.
func FnSupersetQ(args ...Value) Value {
	if len(args) != 2 {
		panic("superset? expects exactly two sets")
	}
	return NewBool(Superset(args[0], args[1]))
}

// FnDisjointQ implements builtin disjoint?.
func FnDisjointQ(args ...Value) Value {
	if len(args) != 2 {
		panic("disjoint? expects exactly two sets")
	}
	return NewBool(Disjoint(args[0], args[1]))
}

// FnRenameKeys implements builtin rename-keys.
func FnRenameKeys(args ...Value) Value {
	if len(args) != 2 {
		panic("rename-keys expects map and key map")
	}
	return RenameKeys(args[0], args[1])
}

// FnMapInvert implements builtin map-invert.
func FnMapInvert(args ...Value) Value {
	if len(args) != 1 {
		panic("map-invert expects exactly one map")
	}
	return MapInvert(args[0])
}

// FnSelect implements builtin select.
func FnSelect(args ...Value) Value {
	if len(args) != 2 {
		panic("select expects predicate and set")
	}
	return SetSelect(args[0], args[1])
}

// FnProject implements builtin project.
func FnProject(args ...Value) Value {
	if len(args) != 2 {
		panic("project expects relation set and key sequence")
	}
	return SetProject(args[0], args[1])
}

// FnRename implements builtin rename.
func FnRename(args ...Value) Value {
	if len(args) != 2 {
		panic("rename expects relation set and key map")
	}
	return SetRename(args[0], args[1])
}

// FnHash implements builtin hash.
func FnHash(args ...Value) Value {
	if len(args) != 1 {
		panic("hash expects 1 argument")
	}
	return NewLong(int64(Hash(args[0])))
}

// FnHashMap implements builtin hash-map.
func FnHashMap(args ...Value) Value {
	return NewMap(args...)
}

// FnList implements builtin list.
func FnList(args ...Value) Value {
	return NewList(args...)
}

// FnArray implements builtin array.
func FnArray(args ...Value) Value {
	return NewArray(args...)
}

// FnRepeat implements builtin repeat.
func FnRepeat(args ...Value) Value {
	return Repeat(args...)
}

// FnRange implements builtin range.
func FnRange(args ...Value) Value {
	return Range(args...)
}

// FnAssoc implements builtin assoc.
func FnAssoc(args ...Value) Value {
	if len(args) < 3 || len(args)%2 == 0 {
		panic("assoc expects collection and key/value pairs")
	}
	return Assoc(args[0], args[1:]...)
}

// FnDissoc implements builtin dissoc.
func FnDissoc(args ...Value) Value {
	if len(args) < 1 {
		panic("dissoc expects at least one argument")
	}
	return Dissoc(args[0], args[1:]...)
}

// FnOpenFile implements builtin open-file.
func FnOpenFile(args ...Value) Value {
	if len(args) != 1 {
		panic("open-file expects exactly one argument")
	}
	return OpenFile(Name(args[0]))
}

// FnCloseFile implements builtin close-file.
func FnCloseFile(args ...Value) Value {
	if len(args) != 1 {
		panic("close-file expects exactly one argument")
	}
	return CloseFile(args[0])
}

// FnCloseChannel implements builtin close-channel.
func FnCloseChannel(args ...Value) Value {
	if len(args) != 1 {
		panic("close-channel expects exactly one argument")
	}
	return ChannelClose(args[0])
}

// FnFileToStrings implements builtin file-to-strings.
func FnFileToStrings(args ...Value) Value {
	if len(args) != 1 {
		panic("file-to-strings expects exactly one argument")
	}
	return FileToStrings(args[0])
}

// FnGoFn implements builtin go-fn.
func FnGoFn(args ...Value) Value {
	if len(args) != 1 {
		panic("go-fn expects exactly one argument")
	}
	return GoFunction(Name(args[0]))
}

// FnGoFnArgs implements builtin go-fn-args.
func FnGoFnArgs(args ...Value) Value {
	if len(args) != 1 {
		panic("go-fn-args expects exactly one argument")
	}
	return GoFunctionArgs(Name(args[0]))
}

// FnRePattern implements builtin re-pattern.
func FnRePattern(args ...Value) Value {
	if len(args) != 1 {
		panic("re-pattern expects exactly one argument")
	}
	return Call(GoFunction("regex/compile"), args[0])
}

// FnReMatches implements builtin re-matches.
func FnReMatches(args ...Value) Value {
	if len(args) != 2 {
		panic("re-matches expects pattern and string")
	}
	return NewBool(RegexMatches(args[0], valueAsString(args[1])))
}

// FnExMessage implements builtin ex-message.
func FnExMessage(args ...Value) Value {
	if len(args) != 1 {
		panic("ex-message expects exactly one argument")
	}
	return ExMessage(args[0])
}

// FnExData implements builtin ex-data.
func FnExData(args ...Value) Value {
	if len(args) != 1 {
		panic("ex-data expects exactly one argument")
	}
	return ExData(args[0])
}

// FnExCause implements builtin ex-cause.
func FnExCause(args ...Value) Value {
	if len(args) != 1 {
		panic("ex-cause expects exactly one argument")
	}
	return ExCause(args[0])
}
