package runtime

func BuiltinFunction(name string) Value {
	switch name {
	case "+":
		return NewFunction(func(args ...Value) Value {
			switch len(args) {
			case 0:
				return NewLong(0)
			case 1:
				if !isNumericTag(args[0].tag) {
					panic("+ expects numeric Value arguments")
				}
				return args[0]
			default:
				return foldNumericBuiltin("+", Add, args...)
			}
		})
	case "*":
		return NewFunction(func(args ...Value) Value {
			return foldNumericBuiltin("*", Mul, args...)
		})
	case "-":
		return NewFunction(func(args ...Value) Value {
			return foldNumericBuiltin("-", Sub, args...)
		})
	case "/":
		return NewFunction(func(args ...Value) Value {
			return foldNumericBuiltin("/", Div, args...)
		})
	case "%":
		return NewFunction(func(args ...Value) Value {
			if len(args) != 2 {
				panic("% expects exactly two arguments")
			}
			return Mod(args[0], args[1])
		})
	case "=":
		return NewFunction(func(args ...Value) Value {
			if len(args) < 2 {
				panic("= expects at least two arguments")
			}
			for i := 0; i < len(args)-1; i++ {
				if !Eq(args[i], args[i+1]) {
					return NewBool(false)
				}
			}
			return NewBool(true)
		})
	case "<":
		return NewFunction(func(args ...Value) Value {
			if len(args) < 2 {
				panic("< expects at least two arguments")
			}
			for i := 0; i < len(args)-1; i++ {
				if !Lt(args[i], args[i+1]) {
					return NewBool(false)
				}
			}
			return NewBool(true)
		})
	case "<=":
		return NewFunction(func(args ...Value) Value {
			if len(args) < 2 {
				panic("<= expects at least two arguments")
			}
			for i := 0; i < len(args)-1; i++ {
				if !Le(args[i], args[i+1]) {
					return NewBool(false)
				}
			}
			return NewBool(true)
		})
	case ">":
		return NewFunction(func(args ...Value) Value {
			if len(args) < 2 {
				panic("> expects at least two arguments")
			}
			for i := 0; i < len(args)-1; i++ {
				if !Gt(args[i], args[i+1]) {
					return NewBool(false)
				}
			}
			return NewBool(true)
		})
	case ">=":
		return NewFunction(func(args ...Value) Value {
			if len(args) < 2 {
				panic(">= expects at least two arguments")
			}
			for i := 0; i < len(args)-1; i++ {
				if !Ge(args[i], args[i+1]) {
					return NewBool(false)
				}
			}
			return NewBool(true)
		})
	case "max":
		return NewFunction(func(args ...Value) Value {
			if len(args) < 1 {
				panic("max expects at least one argument")
			}
			return Max(args...)
		})
	case "min":
		return NewFunction(func(args ...Value) Value {
			if len(args) < 1 {
				panic("min expects at least one argument")
			}
			return Min(args...)
		})
	case "rand-int":
		return NewFunction(func(args ...Value) Value {
			if len(args) != 1 {
				panic("rand-int expects exactly one argument")
			}
			return RandInt(args[0])
		})
	case "first", "fist":
		return NewFunction(func(args ...Value) Value {
			if len(args) != 1 {
				panic("first expects exactly one argument")
			}
			return First(args[0])
		})
	case "rest":
		return NewFunction(func(args ...Value) Value {
			if len(args) != 1 {
				panic("rest expects exactly one argument")
			}
			return Rest(args[0])
		})
	case "next":
		return NewFunction(func(args ...Value) Value {
			if len(args) != 1 {
				panic("next expects exactly one argument")
			}
			return Next(args[0])
		})
	case "last":
		return NewFunction(func(args ...Value) Value {
			if len(args) != 1 {
				panic("last expects exactly one argument")
			}
			return Last(args[0])
		})
	case "reverse":
		return NewFunction(func(args ...Value) Value {
			if len(args) != 1 {
				panic("reverse expects exactly one argument")
			}
			return Reverse(args[0])
		})
	case "cons":
		return NewFunction(func(args ...Value) Value {
			if len(args) != 2 {
				panic("cons expects exactly two arguments")
			}
			return Cons(args[0], args[1])
		})
	case "take":
		return NewFunction(func(args ...Value) Value {
			if len(args) != 2 {
				panic("take expects count and sequence")
			}
			return Take(args[0], args[1])
		})
	case "drop":
		return NewFunction(func(args ...Value) Value {
			if len(args) != 2 {
				panic("drop expects count and sequence")
			}
			return Drop(args[0], args[1])
		})
	case "map":
		return NewFunction(func(args ...Value) Value {
			if len(args) < 2 {
				panic("map expects function and at least one sequence")
			}
			return Map(args[0], args[1:]...)
		})
	case "concat":
		return NewFunction(func(args ...Value) Value {
			return Concat(args...)
		})
	case "into":
		return NewFunction(func(args ...Value) Value {
			if len(args) != 2 {
				panic("into expects collection and sequence")
			}
			return Into(args[0], args[1])
		})
	case "sort-by":
		return NewFunction(func(args ...Value) Value {
			if len(args) != 2 && len(args) != 3 {
				panic("sort-by expects key function, optional comparator, and collection")
			}
			return SortBy(args[0], args[1:]...)
		})
	case "apply":
		return NewFunction(func(args ...Value) Value {
			if len(args) < 2 {
				panic("apply expects function and at least one argument sequence")
			}
			return Apply(args[0], args[1:]...)
		})
	case "pmap":
		return NewFunction(func(args ...Value) Value {
			if len(args) < 2 {
				panic("pmap expects function and at least one sequence")
			}
			return PMap(args[0], args[1:]...)
		})
	case "filter":
		return NewFunction(func(args ...Value) Value {
			if len(args) != 2 {
				panic("filter expects function and one sequence")
			}
			return Filter(args[0], args[1])
		})
	case "reduce":
		return NewFunction(func(args ...Value) Value {
			if len(args) != 2 && len(args) != 3 {
				panic("reduce expects function and collection, or function, initial value, and collection")
			}
			return Reduce(args[0], args[1:]...)
		})
	case "not-empty":
		return NewFunction(func(args ...Value) Value {
			if len(args) != 1 {
				panic("not-empty expects exactly one argument")
			}
			return NotEmpty(args[0])
		})
	case "seq":
		return NewFunction(func(args ...Value) Value {
			if len(args) != 1 {
				panic("seq expects exactly one argument")
			}
			return Seq(args[0])
		})
	case "empty?":
		return NewFunction(func(args ...Value) Value {
			if len(args) != 1 {
				panic("empty? expects exactly one argument")
			}
			return NewBool(IsEmpty(args[0]))
		})
	case "nil?":
		return NewFunction(func(args ...Value) Value {
			if len(args) != 1 {
				panic("nil? expects exactly one argument")
			}
			return NewBool(IsNil(args[0]))
		})
	case "set":
		return NewFunction(func(args ...Value) Value {
			if len(args) != 1 {
				panic("set expects exactly one argument")
			}
			return Set(args[0])
		})
	case "vec":
		return NewFunction(func(args ...Value) Value {
			if len(args) != 1 {
				panic("vec expects exactly one argument")
			}
			return Vec(args[0])
		})
	case "conj":
		return NewFunction(func(args ...Value) Value {
			if len(args) < 2 {
				panic("conj expects collection and at least one item")
			}
			return Conj(args[0], args[1:]...)
		})
	case "contains?":
		return NewFunction(func(args ...Value) Value {
			if len(args) != 2 {
				panic("contains? expects collection and key")
			}
			return NewBool(Contains(args[0], args[1]))
		})
	case "some":
		return NewFunction(func(args ...Value) Value {
			if len(args) != 2 {
				panic("some expects predicate and collection")
			}
			return Some(args[0], args[1])
		})
	case "seq?":
		return NewFunction(func(args ...Value) Value {
			if len(args) != 1 {
				panic("seq? expects exactly one argument")
			}
			return SeqPredicate(args[0])
		})
	case "doall":
		return NewFunction(func(args ...Value) Value {
			if len(args) != 1 {
				panic("doall expects exactly one argument")
			}
			return DoAll(args[0])
		})
	case "dorun":
		return NewFunction(func(args ...Value) Value {
			if len(args) != 1 {
				panic("dorun expects exactly one argument")
			}
			return DoRun(args[0])
		})
	case "count":
		return NewFunction(func(args ...Value) Value {
			if len(args) != 1 {
				panic("count expects exactly one argument")
			}
			return NewLong(int64(Count(args[0])))
		})
	case "double":
		return NewFunction(func(args ...Value) Value {
			if len(args) != 1 {
				panic("double expects exactly one argument")
			}
			return Double(args[0])
		})
	case "format":
		return NewFunction(func(args ...Value) Value {
			if len(args) < 1 {
				panic("format expects a format string and optional values")
			}
			format := valueAsString(args[0])
			return NewString(Format(format, args[1:]...))
		})
	case "keyword":
		return NewFunction(func(args ...Value) Value {
			if len(args) != 1 {
				panic("keyword expects exactly one argument")
			}
			return Keyword(args[0])
		})
	case "get":
		return NewFunction(func(args ...Value) Value {
			if len(args) != 2 && len(args) != 3 {
				panic("get expects map, key, and optional default")
			}
			if len(args) == 2 {
				return Get(args[0], args[1])
			}
			return Get(args[0], args[1], args[2])
		})
	case "keys":
		return NewFunction(func(args ...Value) Value {
			if len(args) != 1 {
				panic("keys expects exactly one argument")
			}
			return Keys(args[0])
		})
	case "hash-map":
		return NewFunction(func(args ...Value) Value {
			return NewMap(args...)
		})
	case "range":
		return NewFunction(func(args ...Value) Value {
			return Range(args...)
		})
	case "repeat":
		return NewFunction(func(args ...Value) Value {
			return Repeat(args...)
		})
	case "assoc":
		return NewFunction(func(args ...Value) Value {
			if len(args) < 3 || len(args)%2 == 0 {
				panic("assoc expects collection and key/value pairs")
			}
			return Assoc(args[0], args[1:]...)
		})
	case "dissoc":
		return NewFunction(func(args ...Value) Value {
			if len(args) < 1 {
				panic("dissoc expects at least one argument")
			}
			return Dissoc(args[0], args[1:]...)
		})
	case "open-file":
		return NewFunction(func(args ...Value) Value {
			if len(args) != 1 {
				panic("open-file expects exactly one argument")
			}
			return OpenFile(Name(args[0]))
		})
	case "file-to-strings":
		return NewFunction(func(args ...Value) Value {
			if len(args) != 1 {
				panic("file-to-strings expects exactly one argument")
			}
			return FileToStrings(args[0])
		})
	case "go-fn":
		return NewFunction(func(args ...Value) Value {
			if len(args) != 1 {
				panic("go-fn expects exactly one argument")
			}
			return GoFunction(Name(args[0]))
		})
	case "go-fn-args":
		return NewFunction(func(args ...Value) Value {
			if len(args) != 1 {
				panic("go-fn-args expects exactly one argument")
			}
			return GoFunctionArgs(Name(args[0]))
		})
	case "re-pattern":
		return NewFunction(func(args ...Value) Value {
			if len(args) != 1 {
				panic("re-pattern expects exactly one argument")
			}
			return Call(GoFunction("regex/compile"), args[0])
		})
	case "re-matches":
		return NewFunction(func(args ...Value) Value {
			if len(args) != 2 {
				panic("re-matches expects pattern and string")
			}
			return NewBool(RegexMatches(args[0], valueAsString(args[1])))
		})
	default:
		panic("unknown builtin function: " + name)
	}
}

func foldNumericBuiltin(name string, op func(lhs, rhs Value) Value, args ...Value) Value {
	if len(args) < 2 {
		panic(name + " expects at least two arguments")
	}
	acc := args[0]
	for i := 1; i < len(args); i++ {
		acc = op(acc, args[i])
	}
	return acc
}
