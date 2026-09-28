package main

import (
	flagrt "flag-lang/runtime"
	"strings"
)

// Source namespace: compiler-tokenizer
type SourceToken struct {
	Token  string `flag:"token"`
	Line   int64  `flag:"line"`
	Offset int64  `flag:"offset"`
}

type TokenState struct {
	Token       string `flag:"token"`
	StartLine   int64  `flag:"start-line"`
	StartOffset int64  `flag:"start-offset"`
	InString    bool   `flag:"in-string"`
	Triple      bool   `flag:"triple"`
	Escaped     bool   `flag:"escaped"`
}

type ParseToken struct {
	Kind    int64  `flag:"kind"`
	Lexeme  string `flag:"lexeme"`
	String  string `flag:"string"`
	Message string `flag:"message"`
	Line    int64  `flag:"line"`
	Col     int64  `flag:"col"`
}

func inc_arity_1(x flagrt.Value) flagrt.Value {
	return flagrt.Add(x, flagrt.NewLong(1))
}

func inc_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("inc expects exactly 1 arguments")
	}
	return inc_arity_1(args[0])
}

func dec_arity_1(x flagrt.Value) flagrt.Value {
	return flagrt.Sub(x, flagrt.NewLong(1))
}

func dec_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("dec expects exactly 1 arguments")
	}
	return dec_arity_1(args[0])
}

func close_arity_1(x flagrt.Value) flagrt.Value {
	var if_result_2 flagrt.Value
	if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagrt.TypeOf(x), flagrt.NewKeyword("file")))) {
		if_result_2 = flagrt.Call(flagrt.BuiltinFunction("close-file"), x)
	} else {
		var if_result_1 flagrt.Value
		if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagrt.TypeOf(x), flagrt.NewKeyword("channel")))) {
			if_result_1 = flagrt.Call(flagrt.BuiltinFunction("close-channel"), x)
		} else {
			if_result_1 = func() flagrt.Value {
				flagrt.Throw(flagrt.NewString(flagrt.Str("close expects a file or channel, got ", flagrt.TypeOf(x))))
				return flagrt.NilValue()
			}()
		}
		if_result_2 = if_result_1
	}
	return if_result_2
}

func close_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("close expects exactly 1 arguments")
	}
	return close_arity_1(args[0])
}

func identity_arity_1(x flagrt.Value) flagrt.Value {
	return x
}

func identity_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("identity expects exactly 1 arguments")
	}
	return identity_arity_1(args[0])
}

func constantly_arity_1(x flagrt.Value) flagrt.Value {
	return flagrt.NewFunction(func(args ...flagrt.Value) flagrt.Value {
		if len(args) < 0 {
			panic("fn expects at least 0 arguments")
		}
		var __rest0 = flagrt.NewArray(args[0:]...)
		_ = __rest0
		_ = __rest0
		var rest = __rest0
		_ = rest
		_ = rest
		return x
	})
}

func constantly_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("constantly expects exactly 1 arguments")
	}
	return constantly_arity_1(args[0])
}

func second_arity_1(coll flagrt.Value) flagrt.Value {
	return flagrt.First(flagrt.Rest(coll))
}

func second_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("second expects exactly 1 arguments")
	}
	return second_arity_1(args[0])
}

func third_arity_1(coll flagrt.Value) flagrt.Value {
	return flagrt.Call(flagrt.BuiltinFunction("slow-nth"), coll, flagrt.NewLong(2))
}

func third_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("third expects exactly 1 arguments")
	}
	return third_arity_1(args[0])
}

func fourth_arity_1(coll flagrt.Value) flagrt.Value {
	return flagrt.Call(flagrt.BuiltinFunction("slow-nth"), coll, flagrt.NewLong(3))
}

func fourth_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("fourth expects exactly 1 arguments")
	}
	return fourth_arity_1(args[0])
}

func fifth_arity_1(coll flagrt.Value) flagrt.Value {
	return flagrt.Call(flagrt.BuiltinFunction("slow-nth"), coll, flagrt.NewLong(4))
}

func fifth_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("fifth expects exactly 1 arguments")
	}
	return fifth_arity_1(args[0])
}

func sixth_arity_1(coll flagrt.Value) flagrt.Value {
	return flagrt.Call(flagrt.BuiltinFunction("slow-nth"), coll, flagrt.NewLong(5))
}

func sixth_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("sixth expects exactly 1 arguments")
	}
	return sixth_arity_1(args[0])
}

func seventh_arity_1(coll flagrt.Value) flagrt.Value {
	return flagrt.Call(flagrt.BuiltinFunction("slow-nth"), coll, flagrt.NewLong(6))
}

func seventh_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("seventh expects exactly 1 arguments")
	}
	return seventh_arity_1(args[0])
}

func eighth_arity_1(coll flagrt.Value) flagrt.Value {
	return flagrt.Call(flagrt.BuiltinFunction("slow-nth"), coll, flagrt.NewLong(7))
}

func eighth_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("eighth expects exactly 1 arguments")
	}
	return eighth_arity_1(args[0])
}

func ninth_arity_1(coll flagrt.Value) flagrt.Value {
	return flagrt.Call(flagrt.BuiltinFunction("slow-nth"), coll, flagrt.NewLong(8))
}

func ninth_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("ninth expects exactly 1 arguments")
	}
	return ninth_arity_1(args[0])
}

func tenth_arity_1(coll flagrt.Value) flagrt.Value {
	return flagrt.Call(flagrt.BuiltinFunction("slow-nth"), coll, flagrt.NewLong(9))
}

func tenth_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("tenth expects exactly 1 arguments")
	}
	return tenth_arity_1(args[0])
}

func val_arity_1(entry flagrt.Value) flagrt.Value {
	return flagrt.Call(second, entry)
}

func val_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("val expects exactly 1 arguments")
	}
	return val_arity_1(args[0])
}

func some_q_arity_1(x flagrt.Value) flagrt.Value {
	var if_result_3 flagrt.Value
	if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsNil(x))) {
		if_result_3 = flagrt.NewBool(false)
	} else {
		if_result_3 = flagrt.NewBool(true)
	}
	return if_result_3
}

func some_q_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("some_q expects exactly 1 arguments")
	}
	return some_q_arity_1(args[0])
}

func true_q_arity_1(x flagrt.Value) flagrt.Value {
	return flagrt.NewBool(flagrt.Eq(x, flagrt.NewBool(true)))
}

func true_q_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("true_q expects exactly 1 arguments")
	}
	return true_q_arity_1(args[0])
}

func false_q_arity_1(x flagrt.Value) flagrt.Value {
	return flagrt.NewBool(flagrt.Eq(x, flagrt.NewBool(false)))
}

func false_q_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("false_q expects exactly 1 arguments")
	}
	return false_q_arity_1(args[0])
}

func boolean_arity_1(x flagrt.Value) flagrt.Value {
	var if_result_4 flagrt.Value
	if flagrt.IsTruthy(x) {
		if_result_4 = flagrt.NewBool(true)
	} else {
		if_result_4 = flagrt.NewBool(false)
	}
	return if_result_4
}

func boolean_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("boolean expects exactly 1 arguments")
	}
	return boolean_arity_1(args[0])
}

func number_q_arity_1(x flagrt.Value) flagrt.Value {
	var let_result_10 flagrt.Value
	{
		var or_tmp = flagrt.NewBool(flagrt.Eq(flagrt.TypeOf(x), flagrt.NewKeyword("int")))
		var if_result_9 flagrt.Value
		if flagrt.IsTruthy(or_tmp) {
			if_result_9 = or_tmp
		} else {
			var let_result_8 flagrt.Value
			{
				var or_tmp = flagrt.NewBool(flagrt.Eq(flagrt.TypeOf(x), flagrt.NewKeyword("float")))
				var if_result_7 flagrt.Value
				if flagrt.IsTruthy(or_tmp) {
					if_result_7 = or_tmp
				} else {
					var let_result_6 flagrt.Value
					{
						var or_tmp = flagrt.NewBool(flagrt.Eq(flagrt.TypeOf(x), flagrt.NewKeyword("bigint")))
						var if_result_5 flagrt.Value
						if flagrt.IsTruthy(or_tmp) {
							if_result_5 = or_tmp
						} else {
							if_result_5 = flagrt.NewBool(flagrt.Eq(flagrt.TypeOf(x), flagrt.NewKeyword("ratio")))
						}
						let_result_6 = if_result_5
					}
					if_result_7 = let_result_6
				}
				let_result_8 = if_result_7
			}
			if_result_9 = let_result_8
		}
		let_result_10 = if_result_9
	}
	return let_result_10
}

func number_q_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("number_q expects exactly 1 arguments")
	}
	return number_q_arity_1(args[0])
}

func int_q_arity_1(x flagrt.Value) flagrt.Value {
	var let_result_12 flagrt.Value
	{
		var or_tmp = flagrt.NewBool(flagrt.Eq(flagrt.TypeOf(x), flagrt.NewKeyword("int")))
		var if_result_11 flagrt.Value
		if flagrt.IsTruthy(or_tmp) {
			if_result_11 = or_tmp
		} else {
			if_result_11 = flagrt.NewBool(flagrt.Eq(flagrt.TypeOf(x), flagrt.NewKeyword("bigint")))
		}
		let_result_12 = if_result_11
	}
	return let_result_12
}

func int_q_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("int_q expects exactly 1 arguments")
	}
	return int_q_arity_1(args[0])
}

func float_q_arity_1(x flagrt.Value) flagrt.Value {
	return flagrt.NewBool(flagrt.Eq(flagrt.TypeOf(x), flagrt.NewKeyword("float")))
}

func float_q_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("float_q expects exactly 1 arguments")
	}
	return float_q_arity_1(args[0])
}

func string_q_arity_1(x flagrt.Value) flagrt.Value {
	return flagrt.NewBool(flagrt.Eq(flagrt.TypeOf(x), flagrt.NewKeyword("string")))
}

func string_q_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("string_q expects exactly 1 arguments")
	}
	return string_q_arity_1(args[0])
}

func symbol_q_arity_1(x flagrt.Value) flagrt.Value {
	return flagrt.NewBool(flagrt.Eq(flagrt.TypeOf(x), flagrt.NewKeyword("symbol")))
}

func symbol_q_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("symbol_q expects exactly 1 arguments")
	}
	return symbol_q_arity_1(args[0])
}

func keyword_q_arity_1(x flagrt.Value) flagrt.Value {
	return flagrt.NewBool(flagrt.Eq(flagrt.TypeOf(x), flagrt.NewKeyword("keyword")))
}

func keyword_q_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("keyword_q expects exactly 1 arguments")
	}
	return keyword_q_arity_1(args[0])
}

func map_q_arity_1(x flagrt.Value) flagrt.Value {
	var let_result_14 flagrt.Value
	{
		var or_tmp = flagrt.NewBool(flagrt.Eq(flagrt.TypeOf(x), flagrt.NewKeyword("map")))
		var if_result_13 flagrt.Value
		if flagrt.IsTruthy(or_tmp) {
			if_result_13 = or_tmp
		} else {
			if_result_13 = flagrt.NewBool(flagrt.Eq(flagrt.TypeOf(x), flagrt.NewKeyword("record")))
		}
		let_result_14 = if_result_13
	}
	return let_result_14
}

func map_q_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("map_q expects exactly 1 arguments")
	}
	return map_q_arity_1(args[0])
}

func vector_q_arity_1(x flagrt.Value) flagrt.Value {
	return flagrt.NewBool(flagrt.Eq(flagrt.TypeOf(x), flagrt.NewKeyword("vector")))
}

func vector_q_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("vector_q expects exactly 1 arguments")
	}
	return vector_q_arity_1(args[0])
}

func set_q_arity_1(x flagrt.Value) flagrt.Value {
	return flagrt.NewBool(flagrt.Eq(flagrt.TypeOf(x), flagrt.NewKeyword("set")))
}

func set_q_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("set_q expects exactly 1 arguments")
	}
	return set_q_arity_1(args[0])
}

func sequential_q_arity_1(x flagrt.Value) flagrt.Value {
	var let_result_20 flagrt.Value
	{
		var or_tmp = flagrt.NewBool(flagrt.Eq(flagrt.TypeOf(x), flagrt.NewKeyword("list")))
		var if_result_19 flagrt.Value
		if flagrt.IsTruthy(or_tmp) {
			if_result_19 = or_tmp
		} else {
			var let_result_18 flagrt.Value
			{
				var or_tmp = flagrt.NewBool(flagrt.Eq(flagrt.TypeOf(x), flagrt.NewKeyword("array")))
				var if_result_17 flagrt.Value
				if flagrt.IsTruthy(or_tmp) {
					if_result_17 = or_tmp
				} else {
					var let_result_16 flagrt.Value
					{
						var or_tmp = flagrt.NewBool(flagrt.Eq(flagrt.TypeOf(x), flagrt.NewKeyword("vector")))
						var if_result_15 flagrt.Value
						if flagrt.IsTruthy(or_tmp) {
							if_result_15 = or_tmp
						} else {
							if_result_15 = flagrt.NewBool(flagrt.Eq(flagrt.TypeOf(x), flagrt.NewKeyword("lazy-list")))
						}
						let_result_16 = if_result_15
					}
					if_result_17 = let_result_16
				}
				let_result_18 = if_result_17
			}
			if_result_19 = let_result_18
		}
		let_result_20 = if_result_19
	}
	return let_result_20
}

func sequential_q_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("sequential_q expects exactly 1 arguments")
	}
	return sequential_q_arity_1(args[0])
}

func coll_q_arity_1(x flagrt.Value) flagrt.Value {
	var let_result_24 flagrt.Value
	{
		var or_tmp = flagrt.Call(sequential_q, x)
		var if_result_23 flagrt.Value
		if flagrt.IsTruthy(or_tmp) {
			if_result_23 = or_tmp
		} else {
			var let_result_22 flagrt.Value
			{
				var or_tmp = flagrt.Call(map_q, x)
				var if_result_21 flagrt.Value
				if flagrt.IsTruthy(or_tmp) {
					if_result_21 = or_tmp
				} else {
					if_result_21 = flagrt.Call(set_q, x)
				}
				let_result_22 = if_result_21
			}
			if_result_23 = let_result_22
		}
		let_result_24 = if_result_23
	}
	return let_result_24
}

func coll_q_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("coll_q expects exactly 1 arguments")
	}
	return coll_q_arity_1(args[0])
}

func empty_arity_1(coll flagrt.Value) flagrt.Value {
	var if_result_31 flagrt.Value
	if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagrt.TypeOf(coll), flagrt.NewKeyword("list")))) {
		if_result_31 = flagrt.NewList()
	} else {
		var if_result_30 flagrt.Value
		if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagrt.TypeOf(coll), flagrt.NewKeyword("array")))) {
			if_result_30 = flagrt.NewArray()
		} else {
			var if_result_29 flagrt.Value
			if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagrt.TypeOf(coll), flagrt.NewKeyword("vector")))) {
				if_result_29 = flagrt.NewVector()
			} else {
				var if_result_28 flagrt.Value
				if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagrt.TypeOf(coll), flagrt.NewKeyword("map")))) {
					if_result_28 = flagrt.NewMap()
				} else {
					var if_result_27 flagrt.Value
					if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagrt.TypeOf(coll), flagrt.NewKeyword("set")))) {
						if_result_27 = flagrt.NewSet()
					} else {
						var if_result_26 flagrt.Value
						if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagrt.TypeOf(coll), flagrt.NewKeyword("lazy-list")))) {
							if_result_26 = flagrt.NewList()
						} else {
							var if_result_25 flagrt.Value
							if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagrt.TypeOf(coll), flagrt.NewKeyword("record")))) {
								if_result_25 = flagrt.NewMap()
							} else {
								if_result_25 = flagrt.NilValue()
							}
							if_result_26 = if_result_25
						}
						if_result_27 = if_result_26
					}
					if_result_28 = if_result_27
				}
				if_result_29 = if_result_28
			}
			if_result_30 = if_result_29
		}
		if_result_31 = if_result_30
	}
	return if_result_31
}

func empty_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("empty expects exactly 1 arguments")
	}
	return empty_arity_1(args[0])
}

func zero_q_arity_1(x flagrt.Value) flagrt.Value {
	var if_result_32 flagrt.Value
	if flagrt.IsTruthy(flagrt.Call(number_q, x)) {
		if_result_32 = flagrt.NewBool(flagrt.Eq(x, flagrt.NewLong(0)))
	} else {
		if_result_32 = func() flagrt.Value {
			flagrt.Throw(flagrt.NewString("zero? expects a number"))
			return flagrt.NilValue()
		}()
	}
	return if_result_32
}

func zero_q_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("zero_q expects exactly 1 arguments")
	}
	return zero_q_arity_1(args[0])
}

func pos_q_arity_1(x flagrt.Value) flagrt.Value {
	var if_result_33 flagrt.Value
	if flagrt.IsTruthy(flagrt.Call(number_q, x)) {
		if_result_33 = flagrt.NewBool(flagrt.Gt(x, flagrt.NewLong(0)))
	} else {
		if_result_33 = func() flagrt.Value {
			flagrt.Throw(flagrt.NewString("pos? expects a number"))
			return flagrt.NilValue()
		}()
	}
	return if_result_33
}

func pos_q_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("pos_q expects exactly 1 arguments")
	}
	return pos_q_arity_1(args[0])
}

func neg_q_arity_1(x flagrt.Value) flagrt.Value {
	var if_result_34 flagrt.Value
	if flagrt.IsTruthy(flagrt.Call(number_q, x)) {
		if_result_34 = flagrt.NewBool(flagrt.Lt(x, flagrt.NewLong(0)))
	} else {
		if_result_34 = func() flagrt.Value {
			flagrt.Throw(flagrt.NewString("neg? expects a number"))
			return flagrt.NilValue()
		}()
	}
	return if_result_34
}

func neg_q_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("neg_q expects exactly 1 arguments")
	}
	return neg_q_arity_1(args[0])
}

func even_q_arity_1(n flagrt.Value) flagrt.Value {
	var if_result_35 flagrt.Value
	if flagrt.IsTruthy(flagrt.Call(int_q, n)) {
		if_result_35 = flagrt.Call(zero_q, flagrt.Mod(n, flagrt.NewLong(2)))
	} else {
		if_result_35 = func() flagrt.Value {
			flagrt.Throw(flagrt.NewString("even? expects an integer"))
			return flagrt.NilValue()
		}()
	}
	return if_result_35
}

func even_q_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("even_q expects exactly 1 arguments")
	}
	return even_q_arity_1(args[0])
}

func odd_q_arity_1(n flagrt.Value) flagrt.Value {
	var if_result_37 flagrt.Value
	if flagrt.IsTruthy(flagrt.Call(int_q, n)) {
		var if_result_36 flagrt.Value
		if flagrt.IsTruthy(flagrt.Call(even_q, n)) {
			if_result_36 = flagrt.NewBool(false)
		} else {
			if_result_36 = flagrt.NewBool(true)
		}
		if_result_37 = if_result_36
	} else {
		if_result_37 = func() flagrt.Value {
			flagrt.Throw(flagrt.NewString("odd? expects an integer"))
			return flagrt.NilValue()
		}()
	}
	return if_result_37
}

func odd_q_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("odd_q expects exactly 1 arguments")
	}
	return odd_q_arity_1(args[0])
}

func not_empty_q_arity_1(coll flagrt.Value) flagrt.Value {
	var if_result_38 flagrt.Value
	if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsEmpty(coll))) {
		if_result_38 = flagrt.NewBool(false)
	} else {
		if_result_38 = flagrt.NewBool(true)
	}
	return if_result_38
}

func not_empty_q_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("not_empty_q expects exactly 1 arguments")
	}
	return not_empty_q_arity_1(args[0])
}

func not_any_q_arity_2(pred flagrt.Value, coll flagrt.Value) flagrt.Value {
	var if_result_39 flagrt.Value
	if flagrt.IsTruthy(flagrt.Some(pred, coll)) {
		if_result_39 = flagrt.NewBool(false)
	} else {
		if_result_39 = flagrt.NewBool(true)
	}
	return if_result_39
}

func not_any_q_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 2 {
		panic("not_any_q expects exactly 2 arguments")
	}
	return not_any_q_arity_2(args[0], args[1])
}

func remove_arity_2(pred flagrt.Value, coll flagrt.Value) flagrt.Value {
	return flagrt.Filter(flagrt.NewFunction(func(args ...flagrt.Value) flagrt.Value {
		if len(args) != 1 {
			panic("fn expects exactly 1 arguments")
		}
		x := args[0]
		var if_result_40 flagrt.Value
		if flagrt.IsTruthy(flagrt.Call(pred, x)) {
			if_result_40 = flagrt.NewBool(false)
		} else {
			if_result_40 = flagrt.NewBool(true)
		}
		return if_result_40
	}), coll)
}

func remove_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 2 {
		panic("remove expects exactly 2 arguments")
	}
	return remove_arity_2(args[0], args[1])
}

func update_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) < 3 {
		panic("update expects at least 3 arguments")
	}
	m := args[0]
	k := args[1]
	f := args[2]
	var __rest0 = flagrt.NewArray(args[3:]...)
	_ = __rest0
	_ = __rest0
	var extra = __rest0
	_ = extra
	return flagrt.Assoc(m, k, flagrt.Apply(f, flagrt.Call(flagrt.BuiltinFunction("get"), m, k), extra))
}

func get_in_arity_2(m flagrt.Value, ks flagrt.Value) flagrt.Value {
	return flagrt.Reduce(flagrt.BuiltinFunction("get"), m, ks)
}

func get_in_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 2 {
		panic("get_in expects exactly 2 arguments")
	}
	return get_in_arity_2(args[0], args[1])
}

func update_in_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) < 3 {
		panic("update_in expects at least 3 arguments")
	}
	m := args[0]
	ks := args[1]
	f := args[2]
	var __rest0 = flagrt.NewArray(args[3:]...)
	_ = __rest0
	_ = __rest0
	var extra = __rest0
	_ = extra
	var if_result_44 flagrt.Value
	if flagrt.IsTruthy(flagrt.Seq(ks)) {
		var let_result_43 flagrt.Value
		{
			var k = flagrt.First(ks)
			var rest_ks = flagrt.Rest(ks)
			var if_result_41 flagrt.Value
			if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsNil(m))) {
				if_result_41 = flagrt.NewMap()
			} else {
				if_result_41 = m
			}
			var base = if_result_41
			var if_result_42 flagrt.Value
			if flagrt.IsTruthy(flagrt.Seq(rest_ks)) {
				if_result_42 = flagrt.Assoc(base, k, flagrt.Apply(flagrt.NewFunction(update_in_variadic), flagrt.Call(flagrt.BuiltinFunction("get"), base, k), rest_ks, f, extra))
			} else {
				if_result_42 = flagrt.Assoc(base, k, flagrt.Apply(f, flagrt.Call(flagrt.BuiltinFunction("get"), base, k), extra))
			}
			let_result_43 = if_result_42
		}
		if_result_44 = let_result_43
	} else {
		if_result_44 = flagrt.Apply(f, m, extra)
	}
	return if_result_44
}

func assoc_in_arity_3(m flagrt.Value, ks flagrt.Value, v flagrt.Value) flagrt.Value {
	var if_result_48 flagrt.Value
	if flagrt.IsTruthy(flagrt.Seq(ks)) {
		var let_result_47 flagrt.Value
		{
			var k = flagrt.First(ks)
			var rest_ks = flagrt.Rest(ks)
			var if_result_45 flagrt.Value
			if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsNil(m))) {
				if_result_45 = flagrt.NewMap()
			} else {
				if_result_45 = m
			}
			var base = if_result_45
			var if_result_46 flagrt.Value
			if flagrt.IsTruthy(flagrt.Seq(rest_ks)) {
				if_result_46 = flagrt.Assoc(base, k, assoc_in_arity_3(flagrt.Call(flagrt.BuiltinFunction("get"), base, k), rest_ks, v))
			} else {
				if_result_46 = flagrt.Assoc(base, k, v)
			}
			let_result_47 = if_result_46
		}
		if_result_48 = let_result_47
	} else {
		if_result_48 = m
	}
	return if_result_48
}

func assoc_in_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 3 {
		panic("assoc_in expects exactly 3 arguments")
	}
	return assoc_in_arity_3(args[0], args[1], args[2])
}

func dissoc_in_arity_2(m flagrt.Value, ks flagrt.Value) flagrt.Value {
	var if_result_56 flagrt.Value
	if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsNil(m))) {
		if_result_56 = flagrt.NilValue()
	} else {
		var if_result_55 flagrt.Value
		if flagrt.IsTruthy(flagrt.Seq(ks)) {
			var let_result_54 flagrt.Value
			{
				var k = flagrt.First(ks)
				var rest_ks = flagrt.Rest(ks)
				var if_result_53 flagrt.Value
				if flagrt.IsTruthy(flagrt.Seq(rest_ks)) {
					var let_result_52 flagrt.Value
					{
						var next = dissoc_in_arity_2(flagrt.Call(flagrt.BuiltinFunction("get"), m, k), rest_ks)
						var let_result_50 flagrt.Value
						{
							var or_tmp = flagrt.NewBool(flagrt.IsNil(next))
							var if_result_49 flagrt.Value
							if flagrt.IsTruthy(or_tmp) {
								if_result_49 = or_tmp
							} else {
								if_result_49 = flagrt.NewBool(flagrt.IsEmpty(next))
							}
							let_result_50 = if_result_49
						}
						var if_result_51 flagrt.Value
						if flagrt.IsTruthy(let_result_50) {
							if_result_51 = flagrt.MapDissoc(m, k)
						} else {
							if_result_51 = flagrt.Assoc(m, k, next)
						}
						let_result_52 = if_result_51
					}
					if_result_53 = let_result_52
				} else {
					if_result_53 = flagrt.MapDissoc(m, k)
				}
				let_result_54 = if_result_53
			}
			if_result_55 = let_result_54
		} else {
			if_result_55 = m
		}
		if_result_56 = if_result_55
	}
	return if_result_56
}

func dissoc_in_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 2 {
		panic("dissoc_in expects exactly 2 arguments")
	}
	return dissoc_in_arity_2(args[0], args[1])
}

func juxt_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) < 0 {
		panic("juxt expects at least 0 arguments")
	}
	var __rest0 = flagrt.NewArray(args[0:]...)
	_ = __rest0
	_ = __rest0
	var fns = __rest0
	_ = fns
	return flagrt.NewFunction(func(args ...flagrt.Value) flagrt.Value {
		if len(args) < 0 {
			panic("fn expects at least 0 arguments")
		}
		var __rest0 = flagrt.NewArray(args[0:]...)
		_ = __rest0
		_ = __rest0
		var xs = __rest0
		_ = xs
		return flagrt.Vec(flagrt.Map(flagrt.NewFunction(func(args ...flagrt.Value) flagrt.Value {
			if len(args) != 1 {
				panic("fn expects exactly 1 arguments")
			}
			f := args[0]
			return flagrt.Apply(f, xs)
		}), fns))
	})
}

func partial_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) < 1 {
		panic("partial expects at least 1 arguments")
	}
	f := args[0]
	var __rest0 = flagrt.NewArray(args[1:]...)
	_ = __rest0
	_ = __rest0
	var bound = __rest0
	_ = bound
	return flagrt.NewFunction(func(args ...flagrt.Value) flagrt.Value {
		if len(args) < 0 {
			panic("fn expects at least 0 arguments")
		}
		var __rest0 = flagrt.NewArray(args[0:]...)
		_ = __rest0
		_ = __rest0
		var more = __rest0
		_ = more
		return flagrt.Apply(f, flagrt.Concat(bound, more))
	})
}

func complement_arity_1(f flagrt.Value) flagrt.Value {
	return flagrt.NewFunction(func(args ...flagrt.Value) flagrt.Value {
		if len(args) < 0 {
			panic("fn expects at least 0 arguments")
		}
		var __rest0 = flagrt.NewArray(args[0:]...)
		_ = __rest0
		_ = __rest0
		var xs = __rest0
		_ = xs
		var if_result_57 flagrt.Value
		if flagrt.IsTruthy(flagrt.Apply(f, xs)) {
			if_result_57 = flagrt.NewBool(false)
		} else {
			if_result_57 = flagrt.NewBool(true)
		}
		return if_result_57
	})
}

func complement_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("complement expects exactly 1 arguments")
	}
	return complement_arity_1(args[0])
}

func fnil_arity_2(f flagrt.Value, x flagrt.Value) flagrt.Value {
	return flagrt.NewFunction(func(args ...flagrt.Value) flagrt.Value {
		if len(args) < 0 {
			panic("fn expects at least 0 arguments")
		}
		var __rest0 = flagrt.NewArray(args[0:]...)
		_ = __rest0
		_ = __rest0
		var xs = __rest0
		_ = xs
		var if_result_59 flagrt.Value
		if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsEmpty(xs))) {
			if_result_59 = flagrt.Call(f)
		} else {
			var if_result_58 flagrt.Value
			if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsNil(flagrt.First(xs)))) {
				if_result_58 = x
			} else {
				if_result_58 = flagrt.First(xs)
			}
			if_result_59 = flagrt.Apply(f, if_result_58, flagrt.Rest(xs))
		}
		return if_result_59
	})
}

func fnil_arity_3(f flagrt.Value, x flagrt.Value, y flagrt.Value) flagrt.Value {
	return flagrt.NewFunction(func(args ...flagrt.Value) flagrt.Value {
		if len(args) < 0 {
			panic("fn expects at least 0 arguments")
		}
		var __rest0 = flagrt.NewArray(args[0:]...)
		_ = __rest0
		_ = __rest0
		var xs = __rest0
		_ = xs
		var let_result_65 flagrt.Value
		{
			var n = flagrt.NewLong(int64(flagrt.Count(xs)))
			var if_result_64 flagrt.Value
			if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(n, flagrt.NewLong(0)))) {
				if_result_64 = flagrt.Call(f)
			} else {
				var if_result_63 flagrt.Value
				if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(n, flagrt.NewLong(1)))) {
					var if_result_60 flagrt.Value
					if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsNil(flagrt.First(xs)))) {
						if_result_60 = x
					} else {
						if_result_60 = flagrt.First(xs)
					}
					if_result_63 = flagrt.Call(f, if_result_60)
				} else {
					var if_result_61 flagrt.Value
					if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsNil(flagrt.First(xs)))) {
						if_result_61 = x
					} else {
						if_result_61 = flagrt.First(xs)
					}
					var if_result_62 flagrt.Value
					if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsNil(flagrt.Call(second, xs)))) {
						if_result_62 = y
					} else {
						if_result_62 = flagrt.Call(second, xs)
					}
					if_result_63 = flagrt.Apply(f, if_result_61, if_result_62, flagrt.Drop(flagrt.NewLong(2), xs))
				}
				if_result_64 = if_result_63
			}
			let_result_65 = if_result_64
		}
		return let_result_65
	})
}

func fnil_arity_4(f flagrt.Value, x flagrt.Value, y flagrt.Value, z flagrt.Value) flagrt.Value {
	return flagrt.NewFunction(func(args ...flagrt.Value) flagrt.Value {
		if len(args) < 0 {
			panic("fn expects at least 0 arguments")
		}
		var __rest0 = flagrt.NewArray(args[0:]...)
		_ = __rest0
		_ = __rest0
		var xs = __rest0
		_ = xs
		var let_result_75 flagrt.Value
		{
			var n = flagrt.NewLong(int64(flagrt.Count(xs)))
			var if_result_74 flagrt.Value
			if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(n, flagrt.NewLong(0)))) {
				if_result_74 = flagrt.Call(f)
			} else {
				var if_result_73 flagrt.Value
				if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(n, flagrt.NewLong(1)))) {
					var if_result_66 flagrt.Value
					if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsNil(flagrt.First(xs)))) {
						if_result_66 = x
					} else {
						if_result_66 = flagrt.First(xs)
					}
					if_result_73 = flagrt.Call(f, if_result_66)
				} else {
					var if_result_72 flagrt.Value
					if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(n, flagrt.NewLong(2)))) {
						var if_result_67 flagrt.Value
						if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsNil(flagrt.First(xs)))) {
							if_result_67 = x
						} else {
							if_result_67 = flagrt.First(xs)
						}
						var if_result_68 flagrt.Value
						if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsNil(flagrt.Call(second, xs)))) {
							if_result_68 = y
						} else {
							if_result_68 = flagrt.Call(second, xs)
						}
						if_result_72 = flagrt.Call(f, if_result_67, if_result_68)
					} else {
						var if_result_69 flagrt.Value
						if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsNil(flagrt.First(xs)))) {
							if_result_69 = x
						} else {
							if_result_69 = flagrt.First(xs)
						}
						var if_result_70 flagrt.Value
						if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsNil(flagrt.Call(second, xs)))) {
							if_result_70 = y
						} else {
							if_result_70 = flagrt.Call(second, xs)
						}
						var if_result_71 flagrt.Value
						if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsNil(flagrt.Call(flagrt.BuiltinFunction("nth"), xs, flagrt.NewLong(2))))) {
							if_result_71 = z
						} else {
							if_result_71 = flagrt.Call(flagrt.BuiltinFunction("nth"), xs, flagrt.NewLong(2))
						}
						if_result_72 = flagrt.Apply(f, if_result_69, if_result_70, if_result_71, flagrt.Drop(flagrt.NewLong(3), xs))
					}
					if_result_73 = if_result_72
				}
				if_result_74 = if_result_73
			}
			let_result_75 = if_result_74
		}
		return let_result_75
	})
}

func fnil_variadic(args ...flagrt.Value) flagrt.Value {
	switch len(args) {
	case 2:
		return fnil_arity_2(args[0], args[1])
	case 3:
		return fnil_arity_3(args[0], args[1], args[2])
	case 4:
		return fnil_arity_4(args[0], args[1], args[2], args[3])
	default:
		panic("fnil expects 2, 3, or 4 arguments")
	}
}

func comp_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) < 0 {
		panic("comp expects at least 0 arguments")
	}
	var __rest0 = flagrt.NewArray(args[0:]...)
	_ = __rest0
	_ = __rest0
	var fs = __rest0
	_ = fs
	var if_result_77 flagrt.Value
	if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsEmpty(fs))) {
		if_result_77 = identity
	} else {
		var if_result_76 flagrt.Value
		if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsEmpty(flagrt.Rest(fs)))) {
			if_result_76 = flagrt.First(fs)
		} else {
			if_result_76 = flagrt.Reduce(flagrt.NewFunction(func(args ...flagrt.Value) flagrt.Value {
				if len(args) != 2 {
					panic("fn expects exactly 2 arguments")
				}
				a := args[0]
				b := args[1]
				return flagrt.NewFunction(func(args ...flagrt.Value) flagrt.Value {
					if len(args) < 0 {
						panic("fn expects at least 0 arguments")
					}
					var __rest0 = flagrt.NewArray(args[0:]...)
					_ = __rest0
					_ = __rest0
					var xs = __rest0
					_ = xs
					return flagrt.Call(a, flagrt.Apply(b, xs))
				})
			}), fs)
		}
		if_result_77 = if_result_76
	}
	return if_result_77
}

func iterate_arity_2(f flagrt.Value, x flagrt.Value) flagrt.Value {
	var let_result_79 flagrt.Value
	{
		var __bind0 = x
		var v = __bind0
		let_result_79 = flagrt.Map(flagrt.NewFunction(func(args ...flagrt.Value) flagrt.Value {
			if len(args) != 1 {
				panic("fn expects exactly 1 arguments")
			}
			_n := args[0]
			_ = _n
			var let_result_78 flagrt.Value
			{
				var cur = v
				_ = func() flagrt.Value {
					v = flagrt.Call(f, cur)
					return v
				}()
				let_result_78 = cur
			}
			return let_result_78
		}), flagrt.Repeat(flagrt.NewLong(0)))
	}
	return let_result_79
}

func iterate_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 2 {
		panic("iterate expects exactly 2 arguments")
	}
	return iterate_arity_2(args[0], args[1])
}

func range__arity_0() flagrt.Value {
	return flagrt.Call(iterate, inc, flagrt.NewLong(0))
}

func range__arity_1(end flagrt.Value) flagrt.Value {
	return range__arity_3(flagrt.NewLong(0), end, flagrt.NewLong(1))
}

func range__arity_2(start flagrt.Value, end flagrt.Value) flagrt.Value {
	return range__arity_3(start, end, flagrt.NewLong(1))
}

func range__arity_3(start flagrt.Value, end flagrt.Value, step flagrt.Value) flagrt.Value {
	var if_result_86 flagrt.Value
	if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(step, flagrt.NewLong(0)))) {
		var if_result_80 flagrt.Value
		if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(start, end))) {
			if_result_80 = flagrt.NewArray()
		} else {
			if_result_80 = flagrt.Repeat(start)
		}
		if_result_86 = if_result_80
	} else {
		var let_result_85 flagrt.Value
		{
			var if_result_83 flagrt.Value
			if flagrt.Gt(step, flagrt.NewLong(0)) {
				var if_result_81 flagrt.Value
				if flagrt.Ge(start, end) {
					if_result_81 = flagrt.NewLong(0)
				} else {
					if_result_81 = flagrt.Quot(flagrt.Add(flagrt.Sub(end, start), flagrt.Call(dec, step)), step)
				}
				if_result_83 = if_result_81
			} else {
				var if_result_82 flagrt.Value
				if flagrt.Le(start, end) {
					if_result_82 = flagrt.NewLong(0)
				} else {
					if_result_82 = flagrt.Quot(flagrt.Add(flagrt.Sub(start, end), flagrt.Call(dec, flagrt.Sub(flagrt.NewLong(0), step))), flagrt.Sub(flagrt.NewLong(0), step))
				}
				if_result_83 = if_result_82
			}
			var n = if_result_83
			var if_result_84 flagrt.Value
			if flagrt.Le(n, flagrt.NewLong(0)) {
				if_result_84 = flagrt.NewArray()
			} else {
				if_result_84 = flagrt.Take(n, flagrt.Call(iterate, flagrt.NewFunction(func(args ...flagrt.Value) flagrt.Value {
					if len(args) != 1 {
						panic("fn expects exactly 1 arguments")
					}
					x := args[0]
					return flagrt.Add(x, step)
				}), start))
			}
			let_result_85 = if_result_84
		}
		if_result_86 = let_result_85
	}
	return if_result_86
}

func range__variadic(args ...flagrt.Value) flagrt.Value {
	switch len(args) {
	case 0:
		return range__arity_0()
	case 1:
		return range__arity_1(args[0])
	case 2:
		return range__arity_2(args[0], args[1])
	case 3:
		return range__arity_3(args[0], args[1], args[2])
	default:
		panic("range expects 0, 1, 2, or 3 arguments")
	}
}

func repeatedly_arity_1(f flagrt.Value) flagrt.Value {
	return flagrt.Map(flagrt.NewFunction(func(args ...flagrt.Value) flagrt.Value {
		if len(args) != 1 {
			panic("fn expects exactly 1 arguments")
		}
		_n := args[0]
		_ = _n
		return flagrt.Call(f)
	}), flagrt.Call(range_))
}

func repeatedly_arity_2(n flagrt.Value, f flagrt.Value) flagrt.Value {
	return flagrt.Take(n, repeatedly_arity_1(f))
}

func repeatedly_variadic(args ...flagrt.Value) flagrt.Value {
	switch len(args) {
	case 1:
		return repeatedly_arity_1(args[0])
	case 2:
		return repeatedly_arity_2(args[0], args[1])
	default:
		panic("repeatedly expects 1 or 2 arguments")
	}
}

func take_while_arity_2(pred flagrt.Value, coll flagrt.Value) flagrt.Value {
	return func() flagrt.Value {
		var xs = coll
		var acc = flagrt.NewArray()
		for {
			var let_result_88 flagrt.Value
			{
				var or_tmp = flagrt.NewBool(flagrt.IsNil(xs))
				var if_result_87 flagrt.Value
				if flagrt.IsTruthy(or_tmp) {
					if_result_87 = or_tmp
				} else {
					if_result_87 = flagrt.NewBool(flagrt.IsEmpty(xs))
				}
				let_result_88 = if_result_87
			}
			var if_result_91 flagrt.Value
			if flagrt.IsTruthy(let_result_88) {
				if_result_91 = acc
			} else {
				var let_result_90 flagrt.Value
				{
					var x = flagrt.First(xs)
					var if_result_89 flagrt.Value
					if flagrt.IsTruthy(flagrt.Call(pred, x)) {
						if_result_89 = flagrt.NewRecur(flagrt.Rest(xs), flagrt.Conj(acc, x))
					} else {
						if_result_89 = acc
					}
					let_result_90 = if_result_89
				}
				if_result_91 = let_result_90
			}
			__loopResult := if_result_91
			if __recurValues, __isRecur := flagrt.UnwrapRecur(__loopResult); __isRecur {
				if len(__recurValues) != 2 {
					panic("internal error: recur arity mismatch")
				}
				xs = __recurValues[0]
				acc = __recurValues[1]
				continue
			}
			return __loopResult
		}
	}()
}

func take_while_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 2 {
		panic("take_while expects exactly 2 arguments")
	}
	return take_while_arity_2(args[0], args[1])
}

func drop_while_arity_2(pred flagrt.Value, coll flagrt.Value) flagrt.Value {
	return func() flagrt.Value {
		var xs = coll
		for {
			var let_result_93 flagrt.Value
			{
				var or_tmp = flagrt.NewBool(flagrt.IsNil(xs))
				var if_result_92 flagrt.Value
				if flagrt.IsTruthy(or_tmp) {
					if_result_92 = or_tmp
				} else {
					if_result_92 = flagrt.NewBool(flagrt.IsEmpty(xs))
				}
				let_result_93 = if_result_92
			}
			var if_result_95 flagrt.Value
			if flagrt.IsTruthy(let_result_93) {
				if_result_95 = xs
			} else {
				var if_result_94 flagrt.Value
				if flagrt.IsTruthy(flagrt.Call(pred, flagrt.First(xs))) {
					if_result_94 = flagrt.NewRecur(flagrt.Rest(xs))
				} else {
					if_result_94 = xs
				}
				if_result_95 = if_result_94
			}
			__loopResult := if_result_95
			if __recurValues, __isRecur := flagrt.UnwrapRecur(__loopResult); __isRecur {
				if len(__recurValues) != 1 {
					panic("internal error: recur arity mismatch")
				}
				xs = __recurValues[0]
				continue
			}
			return __loopResult
		}
	}()
}

func drop_while_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 2 {
		panic("drop_while expects exactly 2 arguments")
	}
	return drop_while_arity_2(args[0], args[1])
}

func take_last_arity_2(n flagrt.Value, coll flagrt.Value) flagrt.Value {
	var let_result_100 flagrt.Value
	{
		var or_tmp = flagrt.NewBool(flagrt.IsNil(coll))
		var if_result_99 flagrt.Value
		if flagrt.IsTruthy(or_tmp) {
			if_result_99 = or_tmp
		} else {
			var let_result_98 flagrt.Value
			{
				var or_tmp = flagrt.NewBool(flagrt.IsEmpty(coll))
				var if_result_97 flagrt.Value
				if flagrt.IsTruthy(or_tmp) {
					if_result_97 = or_tmp
				} else {
					var if_result_96 flagrt.Value
					if flagrt.IsTruthy(flagrt.Call(pos_q, n)) {
						if_result_96 = flagrt.NewBool(false)
					} else {
						if_result_96 = flagrt.NewBool(true)
					}
					if_result_97 = if_result_96
				}
				let_result_98 = if_result_97
			}
			if_result_99 = let_result_98
		}
		let_result_100 = if_result_99
	}
	var if_result_104 flagrt.Value
	if flagrt.IsTruthy(let_result_100) {
		if_result_104 = flagrt.NilValue()
	} else {
		if_result_104 = func() flagrt.Value {
			var s = coll
			var lead = flagrt.Drop(n, coll)
			for {
				var let_result_102 flagrt.Value
				{
					var or_tmp = flagrt.NewBool(flagrt.IsNil(lead))
					var if_result_101 flagrt.Value
					if flagrt.IsTruthy(or_tmp) {
						if_result_101 = or_tmp
					} else {
						if_result_101 = flagrt.NewBool(flagrt.IsEmpty(lead))
					}
					let_result_102 = if_result_101
				}
				var if_result_103 flagrt.Value
				if flagrt.IsTruthy(let_result_102) {
					if_result_103 = s
				} else {
					if_result_103 = flagrt.NewRecur(flagrt.Next(s), flagrt.Next(lead))
				}
				__loopResult := if_result_103
				if __recurValues, __isRecur := flagrt.UnwrapRecur(__loopResult); __isRecur {
					if len(__recurValues) != 2 {
						panic("internal error: recur arity mismatch")
					}
					s = __recurValues[0]
					lead = __recurValues[1]
					continue
				}
				return __loopResult
			}
		}()
	}
	return if_result_104
}

func take_last_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 2 {
		panic("take_last expects exactly 2 arguments")
	}
	return take_last_arity_2(args[0], args[1])
}

func drop_last_arity_1(coll flagrt.Value) flagrt.Value {
	return drop_last_arity_2(flagrt.NewLong(1), coll)
}

func drop_last_arity_2(n flagrt.Value, coll flagrt.Value) flagrt.Value {
	var if_result_107 flagrt.Value
	if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsNil(coll))) {
		if_result_107 = flagrt.NewArray()
	} else {
		var if_result_105 flagrt.Value
		if flagrt.IsTruthy(flagrt.Call(pos_q, n)) {
			if_result_105 = flagrt.NewBool(false)
		} else {
			if_result_105 = flagrt.NewBool(true)
		}
		var if_result_106 flagrt.Value
		if flagrt.IsTruthy(if_result_105) {
			if_result_106 = coll
		} else {
			if_result_106 = flagrt.Map(flagrt.NewFunction(func(args ...flagrt.Value) flagrt.Value {
				if len(args) != 2 {
					panic("fn expects exactly 2 arguments")
				}
				x := args[0]
				_ = args[1]
				return x
			}), coll, flagrt.Drop(n, coll))
		}
		if_result_107 = if_result_106
	}
	return if_result_107
}

func drop_last_variadic(args ...flagrt.Value) flagrt.Value {
	switch len(args) {
	case 1:
		return drop_last_arity_1(args[0])
	case 2:
		return drop_last_arity_2(args[0], args[1])
	default:
		panic("drop-last expects 1 or 2 arguments")
	}
}

func take_nth_arity_2(n flagrt.Value, coll flagrt.Value) flagrt.Value {
	var let_result_109 flagrt.Value
	{
		var or_tmp = flagrt.NewBool(flagrt.IsNil(coll))
		var if_result_108 flagrt.Value
		if flagrt.IsTruthy(or_tmp) {
			if_result_108 = or_tmp
		} else {
			if_result_108 = flagrt.NewBool(flagrt.IsEmpty(coll))
		}
		let_result_109 = if_result_108
	}
	var if_result_115 flagrt.Value
	if flagrt.IsTruthy(let_result_109) {
		if_result_115 = flagrt.NewArray()
	} else {
		var if_result_110 flagrt.Value
		if flagrt.IsTruthy(flagrt.Call(pos_q, n)) {
			if_result_110 = flagrt.NewBool(false)
		} else {
			if_result_110 = flagrt.NewBool(true)
		}
		var if_result_114 flagrt.Value
		if flagrt.IsTruthy(if_result_110) {
			if_result_114 = flagrt.Repeat(flagrt.First(coll))
		} else {
			if_result_114 = func() flagrt.Value {
				var xs = coll
				var acc = flagrt.NewArray()
				for {
					var let_result_112 flagrt.Value
					{
						var or_tmp = flagrt.NewBool(flagrt.IsNil(xs))
						var if_result_111 flagrt.Value
						if flagrt.IsTruthy(or_tmp) {
							if_result_111 = or_tmp
						} else {
							if_result_111 = flagrt.NewBool(flagrt.IsEmpty(xs))
						}
						let_result_112 = if_result_111
					}
					var if_result_113 flagrt.Value
					if flagrt.IsTruthy(let_result_112) {
						if_result_113 = acc
					} else {
						if_result_113 = flagrt.NewRecur(flagrt.Drop(n, xs), flagrt.Conj(acc, flagrt.First(xs)))
					}
					__loopResult := if_result_113
					if __recurValues, __isRecur := flagrt.UnwrapRecur(__loopResult); __isRecur {
						if len(__recurValues) != 2 {
							panic("internal error: recur arity mismatch")
						}
						xs = __recurValues[0]
						acc = __recurValues[1]
						continue
					}
					return __loopResult
				}
			}()
		}
		if_result_115 = if_result_114
	}
	return if_result_115
}

func take_nth_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 2 {
		panic("take_nth expects exactly 2 arguments")
	}
	return take_nth_arity_2(args[0], args[1])
}

func keep_arity_2(f flagrt.Value, coll flagrt.Value) flagrt.Value {
	return flagrt.Filter(some_q, flagrt.Map(f, coll))
}

func keep_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 2 {
		panic("keep expects exactly 2 arguments")
	}
	return keep_arity_2(args[0], args[1])
}

func map_indexed_arity_2(f flagrt.Value, coll flagrt.Value) flagrt.Value {
	return flagrt.Map(f, flagrt.Call(range_), coll)
}

func map_indexed_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 2 {
		panic("map_indexed expects exactly 2 arguments")
	}
	return map_indexed_arity_2(args[0], args[1])
}

func keep_indexed_arity_2(f flagrt.Value, coll flagrt.Value) flagrt.Value {
	return flagrt.Call(keep, identity, flagrt.Call(map_indexed, f, coll))
}

func keep_indexed_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 2 {
		panic("keep_indexed expects exactly 2 arguments")
	}
	return keep_indexed_arity_2(args[0], args[1])
}

func reduce_kv_arity_3(f flagrt.Value, init flagrt.Value, coll flagrt.Value) flagrt.Value {
	var if_result_121 flagrt.Value
	if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsNil(coll))) {
		if_result_121 = init
	} else {
		var if_result_120 flagrt.Value
		if flagrt.IsTruthy(flagrt.Call(map_q, coll)) {
			if_result_120 = flagrt.Reduce(flagrt.NewFunction(func(args ...flagrt.Value) flagrt.Value {
				if len(args) != 2 {
					panic("fn expects exactly 2 arguments")
				}
				acc := args[0]
				pair := args[1]
				return flagrt.Call(f, acc, flagrt.First(pair), flagrt.Call(val, pair))
			}), init, coll)
		} else {
			var let_result_117 flagrt.Value
			{
				var or_tmp = flagrt.NewBool(flagrt.Eq(flagrt.TypeOf(coll), flagrt.NewKeyword("array")))
				var if_result_116 flagrt.Value
				if flagrt.IsTruthy(or_tmp) {
					if_result_116 = or_tmp
				} else {
					if_result_116 = flagrt.Call(vector_q, coll)
				}
				let_result_117 = if_result_116
			}
			var if_result_119 flagrt.Value
			if flagrt.IsTruthy(let_result_117) {
				if_result_119 = func() flagrt.Value {
					var i = flagrt.NewLong(0)
					var acc = init
					for {
						var if_result_118 flagrt.Value
						if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(i, flagrt.NewLong(int64(flagrt.Count(coll)))))) {
							if_result_118 = acc
						} else {
							if_result_118 = flagrt.NewRecur(flagrt.Call(inc, i), flagrt.Call(f, acc, i, flagrt.Call(flagrt.BuiltinFunction("nth"), coll, i)))
						}
						__loopResult := if_result_118
						if __recurValues, __isRecur := flagrt.UnwrapRecur(__loopResult); __isRecur {
							if len(__recurValues) != 2 {
								panic("internal error: recur arity mismatch")
							}
							i = __recurValues[0]
							acc = __recurValues[1]
							continue
						}
						return __loopResult
					}
				}()
			} else {
				if_result_119 = func() flagrt.Value {
					flagrt.Throw(flagrt.NewString("reduce-kv expects a map, array, or vector"))
					return flagrt.NilValue()
				}()
			}
			if_result_120 = if_result_119
		}
		if_result_121 = if_result_120
	}
	return if_result_121
}

func reduce_kv_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 3 {
		panic("reduce_kv expects exactly 3 arguments")
	}
	return reduce_kv_arity_3(args[0], args[1], args[2])
}

func every_q_arity_2(pred flagrt.Value, coll flagrt.Value) flagrt.Value {
	return func() flagrt.Value {
		var xs = coll
		for {
			var if_result_123 flagrt.Value
			if flagrt.IsTruthy(flagrt.Seq(xs)) {
				var if_result_122 flagrt.Value
				if flagrt.IsTruthy(flagrt.Call(pred, flagrt.First(xs))) {
					if_result_122 = flagrt.NewRecur(flagrt.Rest(xs))
				} else {
					if_result_122 = flagrt.NewBool(false)
				}
				if_result_123 = if_result_122
			} else {
				if_result_123 = flagrt.NewBool(true)
			}
			__loopResult := if_result_123
			if __recurValues, __isRecur := flagrt.UnwrapRecur(__loopResult); __isRecur {
				if len(__recurValues) != 1 {
					panic("internal error: recur arity mismatch")
				}
				xs = __recurValues[0]
				continue
			}
			return __loopResult
		}
	}()
}

func every_q_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 2 {
		panic("every_q expects exactly 2 arguments")
	}
	return every_q_arity_2(args[0], args[1])
}

func every_pred_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) < 1 {
		panic("every_pred expects at least 1 arguments")
	}
	p := args[0]
	var __rest0 = flagrt.NewArray(args[1:]...)
	_ = __rest0
	_ = __rest0
	var more = __rest0
	_ = more
	return flagrt.NewFunction(func(args ...flagrt.Value) flagrt.Value {
		if len(args) < 0 {
			panic("fn expects at least 0 arguments")
		}
		var __rest0 = flagrt.NewArray(args[0:]...)
		_ = __rest0
		_ = __rest0
		var xs = __rest0
		_ = xs
		return func() flagrt.Value {
			var ps = flagrt.Cons(p, more)
			for {
				var if_result_125 flagrt.Value
				if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsEmpty(ps))) {
					if_result_125 = flagrt.NewBool(true)
				} else {
					var if_result_124 flagrt.Value
					if flagrt.IsTruthy(flagrt.Call(every_q, flagrt.First(ps), xs)) {
						if_result_124 = flagrt.NewRecur(flagrt.Rest(ps))
					} else {
						if_result_124 = flagrt.NewBool(false)
					}
					if_result_125 = if_result_124
				}
				__loopResult := if_result_125
				if __recurValues, __isRecur := flagrt.UnwrapRecur(__loopResult); __isRecur {
					if len(__recurValues) != 1 {
						panic("internal error: recur arity mismatch")
					}
					ps = __recurValues[0]
					continue
				}
				return __loopResult
			}
		}()
	})
}

func some_fn_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) < 1 {
		panic("some_fn expects at least 1 arguments")
	}
	p := args[0]
	var __rest0 = flagrt.NewArray(args[1:]...)
	_ = __rest0
	_ = __rest0
	var more = __rest0
	_ = more
	return flagrt.NewFunction(func(args ...flagrt.Value) flagrt.Value {
		if len(args) < 0 {
			panic("fn expects at least 0 arguments")
		}
		var __rest0 = flagrt.NewArray(args[0:]...)
		_ = __rest0
		_ = __rest0
		var xs = __rest0
		_ = xs
		return func() flagrt.Value {
			var ps = flagrt.Cons(p, more)
			for {
				var if_result_128 flagrt.Value
				if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsEmpty(ps))) {
					if_result_128 = flagrt.NilValue()
				} else {
					var let_result_127 flagrt.Value
					{
						var r = flagrt.Some(flagrt.First(ps), xs)
						var if_result_126 flagrt.Value
						if flagrt.IsTruthy(r) {
							if_result_126 = r
						} else {
							if_result_126 = flagrt.NewRecur(flagrt.Rest(ps))
						}
						let_result_127 = if_result_126
					}
					if_result_128 = let_result_127
				}
				__loopResult := if_result_128
				if __recurValues, __isRecur := flagrt.UnwrapRecur(__loopResult); __isRecur {
					if len(__recurValues) != 1 {
						panic("internal error: recur arity mismatch")
					}
					ps = __recurValues[0]
					continue
				}
				return __loopResult
			}
		}()
	})
}

func max_key_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) < 2 {
		panic("max_key expects at least 2 arguments")
	}
	kfn := args[0]
	x := args[1]
	var __rest0 = flagrt.NewArray(args[2:]...)
	_ = __rest0
	_ = __rest0
	var more = __rest0
	_ = more
	return flagrt.Reduce(flagrt.NewFunction(func(args ...flagrt.Value) flagrt.Value {
		if len(args) != 2 {
			panic("fn expects exactly 2 arguments")
		}
		best := args[0]
		cand := args[1]
		var if_result_129 flagrt.Value
		if flagrt.Gt(flagrt.Call(kfn, cand), flagrt.Call(kfn, best)) {
			if_result_129 = cand
		} else {
			if_result_129 = best
		}
		return if_result_129
	}), x, more)
}

func zipmap_arity_2(ks flagrt.Value, vs flagrt.Value) flagrt.Value {
	return func() flagrt.Value {
		var ks = ks
		var vs = vs
		var out = flagrt.NewMap()
		for {
			var let_result_131 flagrt.Value
			{
				var and_tmp = flagrt.Seq(ks)
				var if_result_130 flagrt.Value
				if flagrt.IsTruthy(and_tmp) {
					if_result_130 = flagrt.Seq(vs)
				} else {
					if_result_130 = and_tmp
				}
				let_result_131 = if_result_130
			}
			var if_result_132 flagrt.Value
			if flagrt.IsTruthy(let_result_131) {
				if_result_132 = flagrt.NewRecur(flagrt.Rest(ks), flagrt.Rest(vs), flagrt.Assoc(out, flagrt.First(ks), flagrt.First(vs)))
			} else {
				if_result_132 = out
			}
			__loopResult := if_result_132
			if __recurValues, __isRecur := flagrt.UnwrapRecur(__loopResult); __isRecur {
				if len(__recurValues) != 3 {
					panic("internal error: recur arity mismatch")
				}
				ks = __recurValues[0]
				vs = __recurValues[1]
				out = __recurValues[2]
				continue
			}
			return __loopResult
		}
	}()
}

func zipmap_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 2 {
		panic("zipmap expects exactly 2 arguments")
	}
	return zipmap_arity_2(args[0], args[1])
}

func group_by_arity_2(f flagrt.Value, coll flagrt.Value) flagrt.Value {
	return flagrt.Reduce(flagrt.NewFunction(func(args ...flagrt.Value) flagrt.Value {
		if len(args) != 2 {
			panic("fn expects exactly 2 arguments")
		}
		acc := args[0]
		x := args[1]
		var let_result_135 flagrt.Value
		{
			var k = flagrt.Call(f, x)
			var let_result_134 flagrt.Value
			{
				var or_tmp = flagrt.Call(flagrt.BuiltinFunction("get"), acc, k)
				var if_result_133 flagrt.Value
				if flagrt.IsTruthy(or_tmp) {
					if_result_133 = or_tmp
				} else {
					if_result_133 = flagrt.NewArray()
				}
				let_result_134 = if_result_133
			}
			let_result_135 = flagrt.Assoc(acc, k, flagrt.Conj(let_result_134, x))
		}
		return let_result_135
	}), flagrt.NewMap(), coll)
}

func group_by_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 2 {
		panic("group_by expects exactly 2 arguments")
	}
	return group_by_arity_2(args[0], args[1])
}

func select_keys_arity_2(m flagrt.Value, ks flagrt.Value) flagrt.Value {
	return flagrt.Reduce(flagrt.NewFunction(func(args ...flagrt.Value) flagrt.Value {
		if len(args) != 2 {
			panic("fn expects exactly 2 arguments")
		}
		acc := args[0]
		k := args[1]
		var if_result_136 flagrt.Value
		if flagrt.IsTruthy(flagrt.NewBool(flagrt.Contains(m, k))) {
			if_result_136 = flagrt.Assoc(acc, k, flagrt.Call(flagrt.BuiltinFunction("get"), m, k))
		} else {
			if_result_136 = acc
		}
		return if_result_136
	}), flagrt.NewMap(), ks)
}

func select_keys_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 2 {
		panic("select_keys expects exactly 2 arguments")
	}
	return select_keys_arity_2(args[0], args[1])
}

func merge_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) < 0 {
		panic("merge expects at least 0 arguments")
	}
	var __rest0 = flagrt.NewArray(args[0:]...)
	_ = __rest0
	_ = __rest0
	var maps = __rest0
	_ = maps
	return flagrt.Reduce(flagrt.NewFunction(func(args ...flagrt.Value) flagrt.Value {
		if len(args) != 2 {
			panic("fn expects exactly 2 arguments")
		}
		out := args[0]
		m := args[1]
		var if_result_138 flagrt.Value
		if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsNil(m))) {
			if_result_138 = out
		} else {
			var if_result_137 flagrt.Value
			if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsNil(out))) {
				if_result_137 = flagrt.NewMap()
			} else {
				if_result_137 = out
			}
			if_result_138 = flagrt.Reduce(flagrt.NewFunction(func(args ...flagrt.Value) flagrt.Value {
				if len(args) != 2 {
					panic("fn expects exactly 2 arguments")
				}
				acc := args[0]
				pair := args[1]
				return flagrt.Assoc(acc, flagrt.First(pair), flagrt.Call(val, pair))
			}), if_result_137, m)
		}
		return if_result_138
	}), flagrt.NilValue(), maps)
}

func merge_with_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) < 1 {
		panic("merge_with expects at least 1 arguments")
	}
	f := args[0]
	var __rest0 = flagrt.NewArray(args[1:]...)
	_ = __rest0
	_ = __rest0
	var maps = __rest0
	_ = maps
	return flagrt.Reduce(flagrt.NewFunction(func(args ...flagrt.Value) flagrt.Value {
		if len(args) != 2 {
			panic("fn expects exactly 2 arguments")
		}
		out := args[0]
		m := args[1]
		var if_result_142 flagrt.Value
		if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsNil(m))) {
			if_result_142 = out
		} else {
			var if_result_141 flagrt.Value
			if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsNil(out))) {
				if_result_141 = flagrt.NewMap()
			} else {
				if_result_141 = out
			}
			if_result_142 = flagrt.Reduce(flagrt.NewFunction(func(args ...flagrt.Value) flagrt.Value {
				if len(args) != 2 {
					panic("fn expects exactly 2 arguments")
				}
				acc := args[0]
				pair := args[1]
				var let_result_140 flagrt.Value
				{
					var k = flagrt.First(pair)
					var v = flagrt.Call(val, pair)
					var if_result_139 flagrt.Value
					if flagrt.IsTruthy(flagrt.NewBool(flagrt.Contains(acc, k))) {
						if_result_139 = flagrt.Assoc(acc, k, flagrt.Call(f, flagrt.Call(flagrt.BuiltinFunction("get"), acc, k), v))
					} else {
						if_result_139 = flagrt.Assoc(acc, k, v)
					}
					let_result_140 = if_result_139
				}
				return let_result_140
			}), if_result_141, m)
		}
		return if_result_142
	}), flagrt.NilValue(), maps)
}

func mapcat_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) < 1 {
		panic("mapcat expects at least 1 arguments")
	}
	f := args[0]
	var __rest0 = flagrt.NewArray(args[1:]...)
	_ = __rest0
	_ = __rest0
	var colls = __rest0
	_ = colls
	return flagrt.Apply(flagrt.BuiltinFunction("concat"), flagrt.Apply(flagrt.BuiltinFunction("map"), f, colls))
}

func distinct_arity_1(coll flagrt.Value) flagrt.Value {
	return func() flagrt.Value {
		var xs = coll
		var seen = flagrt.NewSet()
		var out = flagrt.NewArray()
		for {
			var if_result_145 flagrt.Value
			if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsEmpty(xs))) {
				if_result_145 = out
			} else {
				var let_result_144 flagrt.Value
				{
					var x = flagrt.First(xs)
					var if_result_143 flagrt.Value
					if flagrt.IsTruthy(flagrt.NewBool(flagrt.Contains(seen, x))) {
						if_result_143 = flagrt.NewRecur(flagrt.Rest(xs), seen, out)
					} else {
						if_result_143 = flagrt.NewRecur(flagrt.Rest(xs), flagrt.Conj(seen, x), flagrt.Conj(out, x))
					}
					let_result_144 = if_result_143
				}
				if_result_145 = let_result_144
			}
			__loopResult := if_result_145
			if __recurValues, __isRecur := flagrt.UnwrapRecur(__loopResult); __isRecur {
				if len(__recurValues) != 3 {
					panic("internal error: recur arity mismatch")
				}
				xs = __recurValues[0]
				seen = __recurValues[1]
				out = __recurValues[2]
				continue
			}
			return __loopResult
		}
	}()
}

func distinct_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("distinct expects exactly 1 arguments")
	}
	return distinct_arity_1(args[0])
}

func flatten_arity_1(x flagrt.Value) flagrt.Value {
	var if_result_147 flagrt.Value
	if flagrt.IsTruthy(flagrt.Call(sequential_q, x)) {
		if_result_147 = flagrt.Vec(flagrt.Call(mapcat, flagrt.NewFunction(func(args ...flagrt.Value) flagrt.Value {
			if len(args) != 1 {
				panic("fn expects exactly 1 arguments")
			}
			item := args[0]
			var if_result_146 flagrt.Value
			if flagrt.IsTruthy(flagrt.Call(sequential_q, item)) {
				if_result_146 = flatten_arity_1(item)
			} else {
				if_result_146 = flagrt.NewArray(item)
			}
			return if_result_146
		}), x))
	} else {
		if_result_147 = flagrt.NewArray()
	}
	return if_result_147
}

func flatten_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("flatten expects exactly 1 arguments")
	}
	return flatten_arity_1(args[0])
}

func interpose_arity_2(sep flagrt.Value, coll flagrt.Value) flagrt.Value {
	var if_result_149 flagrt.Value
	if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsEmpty(coll))) {
		if_result_149 = flagrt.NewArray()
	} else {
		if_result_149 = func() flagrt.Value {
			var xs = flagrt.Rest(coll)
			var out = flagrt.NewArray(flagrt.First(coll))
			for {
				var if_result_148 flagrt.Value
				if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsEmpty(xs))) {
					if_result_148 = out
				} else {
					if_result_148 = flagrt.NewRecur(flagrt.Rest(xs), flagrt.Conj(out, sep, flagrt.First(xs)))
				}
				__loopResult := if_result_148
				if __recurValues, __isRecur := flagrt.UnwrapRecur(__loopResult); __isRecur {
					if len(__recurValues) != 2 {
						panic("internal error: recur arity mismatch")
					}
					xs = __recurValues[0]
					out = __recurValues[1]
					continue
				}
				return __loopResult
			}
		}()
	}
	return if_result_149
}

func interpose_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 2 {
		panic("interpose expects exactly 2 arguments")
	}
	return interpose_arity_2(args[0], args[1])
}

func interleave_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) < 0 {
		panic("interleave expects at least 0 arguments")
	}
	var __rest0 = flagrt.NewArray(args[0:]...)
	_ = __rest0
	_ = __rest0
	var colls = __rest0
	_ = colls
	var if_result_151 flagrt.Value
	if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsEmpty(colls))) {
		if_result_151 = flagrt.NewArray()
	} else {
		if_result_151 = func() flagrt.Value {
			var cols = colls
			var out = flagrt.NewArray()
			for {
				var if_result_150 flagrt.Value
				if flagrt.IsTruthy(flagrt.Call(every_q, not_empty_q, cols)) {
					if_result_150 = flagrt.NewRecur(flagrt.Map(flagrt.BuiltinFunction("rest"), cols), flagrt.Reduce(flagrt.BuiltinFunction("conj"), out, flagrt.Map(flagrt.BuiltinFunction("first"), cols)))
				} else {
					if_result_150 = out
				}
				__loopResult := if_result_150
				if __recurValues, __isRecur := flagrt.UnwrapRecur(__loopResult); __isRecur {
					if len(__recurValues) != 2 {
						panic("internal error: recur arity mismatch")
					}
					cols = __recurValues[0]
					out = __recurValues[1]
					continue
				}
				return __loopResult
			}
		}()
	}
	return if_result_151
}

func partition_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) < 1 {
		panic("partition expects at least 1 arguments")
	}
	n := args[0]
	var __rest0 = flagrt.NewArray(args[1:]...)
	_ = __rest0
	_ = __rest0
	var more = __rest0
	_ = more
	var let_result_155 flagrt.Value
	{
		var or_tmp = flagrt.NewBool(flagrt.Eq(flagrt.NewLong(int64(flagrt.Count(more))), flagrt.NewLong(1)))
		var if_result_154 flagrt.Value
		if flagrt.IsTruthy(or_tmp) {
			if_result_154 = or_tmp
		} else {
			var let_result_153 flagrt.Value
			{
				var or_tmp = flagrt.NewBool(flagrt.Eq(flagrt.NewLong(int64(flagrt.Count(more))), flagrt.NewLong(2)))
				var if_result_152 flagrt.Value
				if flagrt.IsTruthy(or_tmp) {
					if_result_152 = or_tmp
				} else {
					if_result_152 = flagrt.NewBool(flagrt.Eq(flagrt.NewLong(int64(flagrt.Count(more))), flagrt.NewLong(3)))
				}
				let_result_153 = if_result_152
			}
			if_result_154 = let_result_153
		}
		let_result_155 = if_result_154
	}
	var if_result_165 flagrt.Value
	if flagrt.IsTruthy(let_result_155) {
		var let_result_164 flagrt.Value
		{
			var c = flagrt.NewLong(int64(flagrt.Count(more)))
			var if_result_156 flagrt.Value
			if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(c, flagrt.NewLong(1)))) {
				if_result_156 = n
			} else {
				if_result_156 = flagrt.First(more)
			}
			var step = if_result_156
			var if_result_158 flagrt.Value
			if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(c, flagrt.NewLong(1)))) {
				if_result_158 = flagrt.First(more)
			} else {
				var if_result_157 flagrt.Value
				if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(c, flagrt.NewLong(2)))) {
					if_result_157 = flagrt.Call(second, more)
				} else {
					if_result_157 = flagrt.Call(flagrt.BuiltinFunction("nth"), more, flagrt.NewLong(2))
				}
				if_result_158 = if_result_157
			}
			var coll = if_result_158
			var if_result_159 flagrt.Value
			if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(c, flagrt.NewLong(3)))) {
				if_result_159 = flagrt.Call(second, more)
			} else {
				if_result_159 = flagrt.NilValue()
			}
			var pad = if_result_159
			var use_pad = flagrt.NewBool(flagrt.Eq(c, flagrt.NewLong(3)))
			let_result_164 = func() flagrt.Value {
				var xs = coll
				var out = flagrt.NewArray()
				for {
					var if_result_163 flagrt.Value
					if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsEmpty(xs))) {
						if_result_163 = out
					} else {
						var let_result_162 flagrt.Value
						{
							var part = flagrt.Vec(flagrt.Take(n, xs))
							var if_result_161 flagrt.Value
							if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagrt.NewLong(int64(flagrt.Count(part))), n))) {
								if_result_161 = flagrt.NewRecur(flagrt.Drop(step, xs), flagrt.Conj(out, part))
							} else {
								var if_result_160 flagrt.Value
								if flagrt.IsTruthy(use_pad) {
									if_result_160 = flagrt.Conj(out, flagrt.Vec(flagrt.Take(n, flagrt.Concat(part, pad))))
								} else {
									if_result_160 = out
								}
								if_result_161 = if_result_160
							}
							let_result_162 = if_result_161
						}
						if_result_163 = let_result_162
					}
					__loopResult := if_result_163
					if __recurValues, __isRecur := flagrt.UnwrapRecur(__loopResult); __isRecur {
						if len(__recurValues) != 2 {
							panic("internal error: recur arity mismatch")
						}
						xs = __recurValues[0]
						out = __recurValues[1]
						continue
					}
					return __loopResult
				}
			}()
		}
		if_result_165 = let_result_164
	} else {
		if_result_165 = func() flagrt.Value {
			flagrt.Throw(flagrt.NewString("partition expects 2, 3, or 4 arguments"))
			return flagrt.NilValue()
		}()
	}
	return if_result_165
}

func partition_all_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) < 1 {
		panic("partition_all expects at least 1 arguments")
	}
	n := args[0]
	var __rest0 = flagrt.NewArray(args[1:]...)
	_ = __rest0
	_ = __rest0
	var more = __rest0
	_ = more
	var let_result_167 flagrt.Value
	{
		var or_tmp = flagrt.NewBool(flagrt.Eq(flagrt.NewLong(int64(flagrt.Count(more))), flagrt.NewLong(1)))
		var if_result_166 flagrt.Value
		if flagrt.IsTruthy(or_tmp) {
			if_result_166 = or_tmp
		} else {
			if_result_166 = flagrt.NewBool(flagrt.Eq(flagrt.NewLong(int64(flagrt.Count(more))), flagrt.NewLong(2)))
		}
		let_result_167 = if_result_166
	}
	var if_result_172 flagrt.Value
	if flagrt.IsTruthy(let_result_167) {
		var let_result_171 flagrt.Value
		{
			var if_result_168 flagrt.Value
			if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagrt.NewLong(int64(flagrt.Count(more))), flagrt.NewLong(1)))) {
				if_result_168 = n
			} else {
				if_result_168 = flagrt.First(more)
			}
			var step = if_result_168
			var if_result_169 flagrt.Value
			if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagrt.NewLong(int64(flagrt.Count(more))), flagrt.NewLong(1)))) {
				if_result_169 = flagrt.First(more)
			} else {
				if_result_169 = flagrt.Call(second, more)
			}
			var coll = if_result_169
			let_result_171 = func() flagrt.Value {
				var xs = coll
				var out = flagrt.NewArray()
				for {
					var if_result_170 flagrt.Value
					if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsEmpty(xs))) {
						if_result_170 = out
					} else {
						if_result_170 = flagrt.NewRecur(flagrt.Drop(step, xs), flagrt.Conj(out, flagrt.Vec(flagrt.Take(n, xs))))
					}
					__loopResult := if_result_170
					if __recurValues, __isRecur := flagrt.UnwrapRecur(__loopResult); __isRecur {
						if len(__recurValues) != 2 {
							panic("internal error: recur arity mismatch")
						}
						xs = __recurValues[0]
						out = __recurValues[1]
						continue
					}
					return __loopResult
				}
			}()
		}
		if_result_172 = let_result_171
	} else {
		if_result_172 = func() flagrt.Value {
			flagrt.Throw(flagrt.NewString("partition-all expects 2 or 3 arguments"))
			return flagrt.NilValue()
		}()
	}
	return if_result_172
}

func partition_by_arity_2(f flagrt.Value, coll flagrt.Value) flagrt.Value {
	return func() flagrt.Value {
		var xs = coll
		var out = flagrt.NewArray()
		for {
			var if_result_178 flagrt.Value
			if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsEmpty(xs))) {
				if_result_178 = out
			} else {
				var let_result_177 flagrt.Value
				{
					var fv = flagrt.Call(f, flagrt.First(xs))
					var run = func() flagrt.Value {
						var ys = flagrt.Rest(xs)
						var acc = flagrt.NewArray(flagrt.First(xs))
						for {
							var let_result_175 flagrt.Value
							{
								var or_tmp = flagrt.NewBool(flagrt.IsEmpty(ys))
								var if_result_174 flagrt.Value
								if flagrt.IsTruthy(or_tmp) {
									if_result_174 = or_tmp
								} else {
									var if_result_173 flagrt.Value
									if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(fv, flagrt.Call(f, flagrt.First(ys))))) {
										if_result_173 = flagrt.NewBool(false)
									} else {
										if_result_173 = flagrt.NewBool(true)
									}
									if_result_174 = if_result_173
								}
								let_result_175 = if_result_174
							}
							var if_result_176 flagrt.Value
							if flagrt.IsTruthy(let_result_175) {
								if_result_176 = acc
							} else {
								if_result_176 = flagrt.NewRecur(flagrt.Rest(ys), flagrt.Conj(acc, flagrt.First(ys)))
							}
							__loopResult := if_result_176
							if __recurValues, __isRecur := flagrt.UnwrapRecur(__loopResult); __isRecur {
								if len(__recurValues) != 2 {
									panic("internal error: recur arity mismatch")
								}
								ys = __recurValues[0]
								acc = __recurValues[1]
								continue
							}
							return __loopResult
						}
					}()
					let_result_177 = flagrt.NewRecur(flagrt.Drop(flagrt.NewLong(int64(flagrt.Count(run))), xs), flagrt.Conj(out, run))
				}
				if_result_178 = let_result_177
			}
			__loopResult := if_result_178
			if __recurValues, __isRecur := flagrt.UnwrapRecur(__loopResult); __isRecur {
				if len(__recurValues) != 2 {
					panic("internal error: recur arity mismatch")
				}
				xs = __recurValues[0]
				out = __recurValues[1]
				continue
			}
			return __loopResult
		}
	}()
}

func partition_by_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 2 {
		panic("partition_by expects exactly 2 arguments")
	}
	return partition_by_arity_2(args[0], args[1])
}

func sort_arity_1(coll flagrt.Value) flagrt.Value {
	return flagrt.SortBy(identity, coll)
}

func sort_arity_2(comp flagrt.Value, coll flagrt.Value) flagrt.Value {
	return flagrt.SortBy(identity, comp, coll)
}

func sort_variadic(args ...flagrt.Value) flagrt.Value {
	switch len(args) {
	case 1:
		return sort_arity_1(args[0])
	case 2:
		return sort_arity_2(args[0], args[1])
	default:
		panic("sort expects 1 or 2 arguments")
	}
}

// True if pred is truthy for every value. On first failure, close ch and return false.
func async__channel_every_q_arity_2(pred flagrt.Value, ch flagrt.Value) flagrt.Value {
	var let_result_181 flagrt.Value
	{
		var v = flagrt.Call(async__channel_receive, ch)
		var if_result_180 flagrt.Value
		if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsNil(v))) {
			if_result_180 = flagrt.NewBool(true)
		} else {
			var if_result_179 flagrt.Value
			if flagrt.IsTruthy(flagrt.Call(pred, v)) {
				if_result_179 = async__channel_every_q_arity_2(pred, ch)
			} else {
				_ = flagrt.Call(close, ch)
				if_result_179 = flagrt.NewBool(false)
			}
			if_result_180 = if_result_179
		}
		let_result_181 = if_result_180
	}
	return let_result_181
}

func async__channel_every_q_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 2 {
		panic("async__channel_every_q expects exactly 2 arguments")
	}
	return async__channel_every_q_arity_2(args[0], args[1])
}

// First value for which pred is truthy, or nil. On match, close ch and return the value.
func async__channel_some_q_arity_2(pred flagrt.Value, ch flagrt.Value) flagrt.Value {
	var let_result_184 flagrt.Value
	{
		var v = flagrt.Call(async__channel_receive, ch)
		var if_result_183 flagrt.Value
		if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsNil(v))) {
			if_result_183 = flagrt.NilValue()
		} else {
			var if_result_182 flagrt.Value
			if flagrt.IsTruthy(flagrt.Call(pred, v)) {
				_ = flagrt.Call(close, ch)
				if_result_182 = v
			} else {
				if_result_182 = async__channel_some_q_arity_2(pred, ch)
			}
			if_result_183 = if_result_182
		}
		let_result_184 = if_result_183
	}
	return let_result_184
}

func async__channel_some_q_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 2 {
		panic("async__channel_some_q expects exactly 2 arguments")
	}
	return async__channel_some_q_arity_2(args[0], args[1])
}

func stdlib__second_arity_1(coll flagrt.Value) flagrt.Value {
	return flagrt.First(flagrt.Rest(coll))
}

func stdlib__second_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("stdlib__second expects exactly 1 arguments")
	}
	return stdlib__second_arity_1(args[0])
}

func stdlib__third_arity_1(coll flagrt.Value) flagrt.Value {
	return flagrt.First(flagrt.Rest(flagrt.Rest(coll)))
}

func stdlib__third_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("stdlib__third expects exactly 1 arguments")
	}
	return stdlib__third_arity_1(args[0])
}

func compiler__whitespace_char_q_arity_1(ch flagrt.Value) flagrt.Value {
	return flagrt.NewBool(flagrt.Contains(compiler__whitespace_chars, ch))
}

func compiler__whitespace_char_q_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__whitespace_char_q expects exactly 1 arguments")
	}
	return compiler__whitespace_char_q_arity_1(args[0])
}

func compiler__delimiter_char_q_arity_1(ch flagrt.Value) flagrt.Value {
	return flagrt.NewBool(flagrt.Contains(compiler__delimiter_chars, ch))
}

func compiler__delimiter_char_q_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__delimiter_char_q expects exactly 1 arguments")
	}
	return compiler__delimiter_char_q_arity_1(args[0])
}

func compiler____gtSourceToken_arity_3(token flagrt.Value, line flagrt.Value, offset flagrt.Value) flagrt.Value {
	return flagrt.NewRecord(SourceToken{Token: flagrt.RequireString(token), Line: flagrt.RequireLong(line), Offset: flagrt.RequireLong(offset)})
}

func compiler____gtSourceToken_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 3 {
		panic("compiler____gtSourceToken expects exactly 3 arguments")
	}
	return compiler____gtSourceToken_arity_3(args[0], args[1], args[2])
}

func compiler__map__gtSourceToken_arity_1(m flagrt.Value) flagrt.Value {
	return flagrt.NewRecord(SourceToken{Token: flagrt.RequireString(flagrt.Get(m, flagKw_token)), Line: flagrt.RequireLong(flagrt.Get(m, flagKw_line)), Offset: flagrt.RequireLong(flagrt.Get(m, flagKw_offset))})
}

func compiler__map__gtSourceToken_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__map__gtSourceToken expects exactly 1 arguments")
	}
	return compiler__map__gtSourceToken_arity_1(args[0])
}

func compiler____gtTokenState_arity_6(token flagrt.Value, start_line flagrt.Value, start_offset flagrt.Value, in_string flagrt.Value, triple flagrt.Value, escaped flagrt.Value) flagrt.Value {
	return flagrt.NewRecord(TokenState{Token: flagrt.RequireString(token), StartLine: flagrt.RequireLong(start_line), StartOffset: flagrt.RequireLong(start_offset), InString: flagrt.RequireBool(in_string), Triple: flagrt.RequireBool(triple), Escaped: flagrt.RequireBool(escaped)})
}

func compiler____gtTokenState_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 6 {
		panic("compiler____gtTokenState expects exactly 6 arguments")
	}
	return compiler____gtTokenState_arity_6(args[0], args[1], args[2], args[3], args[4], args[5])
}

func compiler__map__gtTokenState_arity_1(m flagrt.Value) flagrt.Value {
	return flagrt.NewRecord(TokenState{Token: flagrt.RequireString(flagrt.Get(m, flagKw_token)), StartLine: flagrt.RequireLong(flagrt.Get(m, flagKw_start_line)), StartOffset: flagrt.RequireLong(flagrt.Get(m, flagKw_start_offset)), InString: flagrt.RequireBool(flagrt.Get(m, flagKw_in_string)), Triple: flagrt.RequireBool(flagrt.Get(m, flagKw_triple)), Escaped: flagrt.RequireBool(flagrt.Get(m, flagKw_escaped))})
}

func compiler__map__gtTokenState_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__map__gtTokenState expects exactly 1 arguments")
	}
	return compiler__map__gtTokenState_arity_1(args[0])
}

func compiler____gtParseToken_arity_6(kind flagrt.Value, lexeme flagrt.Value, string_ flagrt.Value, message flagrt.Value, line flagrt.Value, col flagrt.Value) flagrt.Value {
	return flagrt.NewRecord(ParseToken{Kind: flagrt.RequireLong(kind), Lexeme: flagrt.RequireString(lexeme), String: flagrt.RequireString(string_), Message: flagrt.RequireString(message), Line: flagrt.RequireLong(line), Col: flagrt.RequireLong(col)})
}

func compiler____gtParseToken_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 6 {
		panic("compiler____gtParseToken expects exactly 6 arguments")
	}
	return compiler____gtParseToken_arity_6(args[0], args[1], args[2], args[3], args[4], args[5])
}

func compiler__map__gtParseToken_arity_1(m flagrt.Value) flagrt.Value {
	return flagrt.NewRecord(ParseToken{Kind: flagrt.RequireLong(flagrt.Get(m, flagKw_kind)), Lexeme: flagrt.RequireString(flagrt.Get(m, flagKw_lexeme)), String: flagrt.RequireString(flagrt.Get(m, flagKw_string)), Message: flagrt.RequireString(flagrt.Get(m, flagKw_message)), Line: flagrt.RequireLong(flagrt.Get(m, flagKw_line)), Col: flagrt.RequireLong(flagrt.Get(m, flagKw_col))})
}

func compiler__map__gtParseToken_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__map__gtParseToken expects exactly 1 arguments")
	}
	return compiler__map__gtParseToken_arity_1(args[0])
}

func compiler__emit_token_bang_arity_4(out flagrt.Value, token flagrt.Value, line flagrt.Value, offset flagrt.Value) flagrt.Value {
	var if_result_185 flagrt.Value
	if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsEmpty(token))) {
		if_result_185 = flagrt.NilValue()
	} else {
		if_result_185 = flagrt.Call(async__channel_send, out, flagrt.Call(compiler____gtSourceToken, token, line, offset))
	}
	return if_result_185
}

func compiler__emit_token_bang_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 4 {
		panic("compiler__emit_token_bang expects exactly 4 arguments")
	}
	return compiler__emit_token_bang_arity_4(args[0], args[1], args[2], args[3])
}

func compiler__make_state_arity_6(token flagrt.Value, start_line flagrt.Value, start_offset flagrt.Value, in_string flagrt.Value, triple flagrt.Value, escaped flagrt.Value) flagrt.Value {
	return flagrt.Call(compiler____gtTokenState, token, start_line, start_offset, in_string, triple, escaped)
}

func compiler__make_state_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 6 {
		panic("compiler__make_state expects exactly 6 arguments")
	}
	return compiler__make_state_arity_6(args[0], args[1], args[2], args[3], args[4], args[5])
}

func compiler__starts_triple_quote_q_arity_1(chars flagrt.Value) flagrt.Value {
	var let_result_189 flagrt.Value
	{
		var and_tmp = flagrt.NewBool(flagrt.Eq(flagrt.NewString("\""), flagrt.First(chars)))
		var if_result_188 flagrt.Value
		if flagrt.IsTruthy(and_tmp) {
			var let_result_187 flagrt.Value
			{
				var and_tmp = flagrt.NewBool(flagrt.Eq(flagrt.NewString("\""), flagrt.Call(stdlib__second, chars)))
				var if_result_186 flagrt.Value
				if flagrt.IsTruthy(and_tmp) {
					if_result_186 = flagrt.NewBool(flagrt.Eq(flagrt.NewString("\""), flagrt.Call(stdlib__third, chars)))
				} else {
					if_result_186 = and_tmp
				}
				let_result_187 = if_result_186
			}
			if_result_188 = let_result_187
		} else {
			if_result_188 = and_tmp
		}
		let_result_189 = if_result_188
	}
	return let_result_189
}

func compiler__starts_triple_quote_q_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__starts_triple_quote_q expects exactly 1 arguments")
	}
	return compiler__starts_triple_quote_q_arity_1(args[0])
}

func compiler__starts_dispatch_set_q_arity_1(chars flagrt.Value) flagrt.Value {
	var let_result_191 flagrt.Value
	{
		var and_tmp = flagrt.NewBool(flagrt.Eq(flagrt.NewString("#"), flagrt.First(chars)))
		var if_result_190 flagrt.Value
		if flagrt.IsTruthy(and_tmp) {
			if_result_190 = flagrt.NewBool(flagrt.Eq(flagrt.NewString("{"), flagrt.Call(stdlib__second, chars)))
		} else {
			if_result_190 = and_tmp
		}
		let_result_191 = if_result_190
	}
	return let_result_191
}

func compiler__starts_dispatch_set_q_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__starts_dispatch_set_q expects exactly 1 arguments")
	}
	return compiler__starts_dispatch_set_q_arity_1(args[0])
}

func compiler__starts_dispatch_fn_q_arity_1(chars flagrt.Value) flagrt.Value {
	var let_result_193 flagrt.Value
	{
		var and_tmp = flagrt.NewBool(flagrt.Eq(flagrt.NewString("#"), flagrt.First(chars)))
		var if_result_192 flagrt.Value
		if flagrt.IsTruthy(and_tmp) {
			if_result_192 = flagrt.NewBool(flagrt.Eq(flagrt.NewString("("), flagrt.Call(stdlib__second, chars)))
		} else {
			if_result_192 = and_tmp
		}
		let_result_193 = if_result_192
	}
	return let_result_193
}

func compiler__starts_dispatch_fn_q_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__starts_dispatch_fn_q expects exactly 1 arguments")
	}
	return compiler__starts_dispatch_fn_q_arity_1(args[0])
}

func compiler__strip_quoted_token_arity_3(token flagrt.Value, prefix flagrt.Value, suffix flagrt.Value) flagrt.Value {
	var let_result_196 flagrt.Value
	{
		var chars = flagrt.Seq(token)
		var total = flagrt.NewLong(int64(flagrt.Count(chars)))
		var if_result_195 flagrt.Value
		if flagrt.Le(total, flagrt.Add(prefix, suffix)) {
			if_result_195 = flagStr_
		} else {
			if_result_195 = func() flagrt.Value {
				var remaining = flagrt.Take(flagrt.Sub(flagrt.Sub(total, prefix), suffix), flagrt.Drop(prefix, chars))
				var out = flagStr_
				for {
					var if_result_194 flagrt.Value
					if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsEmpty(remaining))) {
						if_result_194 = out
					} else {
						if_result_194 = flagrt.NewRecur(flagrt.Rest(remaining), flagrt.NewString(flagrt.Str(out, flagrt.First(remaining))))
					}
					__loopResult := if_result_194
					if __recurValues, __isRecur := flagrt.UnwrapRecur(__loopResult); __isRecur {
						if len(__recurValues) != 2 {
							panic("internal error: recur arity mismatch")
						}
						remaining = __recurValues[0]
						out = __recurValues[1]
						continue
					}
					return __loopResult
				}
			}()
		}
		let_result_196 = if_result_195
	}
	return let_result_196
}

func compiler__strip_quoted_token_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 3 {
		panic("compiler__strip_quoted_token expects exactly 3 arguments")
	}
	return compiler__strip_quoted_token_arity_3(args[0], args[1], args[2])
}

func compiler__token__gtstring_value_arity_1(token flagrt.Value) flagrt.Value {
	var let_result_212 flagrt.Value
	{
		var chars = flagrt.Seq(token)
		var n = flagrt.NewLong(int64(flagrt.Count(chars)))
		var let_result_204 flagrt.Value
		{
			var and_tmp = flagrt.Ge(n, flagrt.NewLong(6))
			var if_result_203 flagrt.Value
			if and_tmp {
				var let_result_202 flagrt.Value
				{
					var and_tmp = flagrt.Call(compiler__starts_triple_quote_q, chars)
					var if_result_201 flagrt.Value
					if flagrt.IsTruthy(and_tmp) {
						var let_result_200 flagrt.Value
						{
							var and_tmp = flagrt.NewBool(flagrt.Eq(flagrt.NewString("\""), flagrt.First(flagrt.Drop(flagrt.Sub(n, flagrt.NewLong(3)), chars))))
							var if_result_199 flagrt.Value
							if flagrt.IsTruthy(and_tmp) {
								var let_result_198 flagrt.Value
								{
									var and_tmp = flagrt.NewBool(flagrt.Eq(flagrt.NewString("\""), flagrt.First(flagrt.Drop(flagrt.Sub(n, flagrt.NewLong(2)), chars))))
									var if_result_197 flagrt.Value
									if flagrt.IsTruthy(and_tmp) {
										if_result_197 = flagrt.NewBool(flagrt.Eq(flagrt.NewString("\""), flagrt.First(flagrt.Drop(flagrt.Sub(n, flagrt.NewLong(1)), chars))))
									} else {
										if_result_197 = and_tmp
									}
									let_result_198 = if_result_197
								}
								if_result_199 = let_result_198
							} else {
								if_result_199 = and_tmp
							}
							let_result_200 = if_result_199
						}
						if_result_201 = let_result_200
					} else {
						if_result_201 = and_tmp
					}
					let_result_202 = if_result_201
				}
				if_result_203 = let_result_202
			} else {
				if_result_203 = flagrt.NewBool(and_tmp)
			}
			let_result_204 = if_result_203
		}
		var if_result_211 flagrt.Value
		if flagrt.IsTruthy(let_result_204) {
			if_result_211 = flagrt.Call(compiler__strip_quoted_token, token, flagrt.NewLong(3), flagrt.NewLong(3))
		} else {
			var let_result_208 flagrt.Value
			{
				var and_tmp = flagrt.Ge(n, flagrt.NewLong(2))
				var if_result_207 flagrt.Value
				if and_tmp {
					var let_result_206 flagrt.Value
					{
						var and_tmp = flagrt.NewBool(flagrt.Eq(flagrt.NewString("\""), flagrt.First(chars)))
						var if_result_205 flagrt.Value
						if flagrt.IsTruthy(and_tmp) {
							if_result_205 = flagrt.NewBool(flagrt.Eq(flagrt.NewString("\""), flagrt.First(flagrt.Drop(flagrt.Sub(n, flagrt.NewLong(1)), chars))))
						} else {
							if_result_205 = and_tmp
						}
						let_result_206 = if_result_205
					}
					if_result_207 = let_result_206
				} else {
					if_result_207 = flagrt.NewBool(and_tmp)
				}
				let_result_208 = if_result_207
			}
			var if_result_210 flagrt.Value
			if flagrt.IsTruthy(let_result_208) {
				if_result_210 = flagrt.Call(compiler__strip_quoted_token, token, flagrt.NewLong(1), flagrt.NewLong(1))
			} else {
				var if_result_209 flagrt.Value
				if flagrt.IsTruthy(flagrt.NewBool(true)) {
					if_result_209 = token
				} else {
					if_result_209 = flagrt.NilValue()
				}
				if_result_210 = if_result_209
			}
			if_result_211 = if_result_210
		}
		let_result_212 = if_result_211
	}
	return let_result_212
}

func compiler__token__gtstring_value_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__token__gtstring_value expects exactly 1 arguments")
	}
	return compiler__token__gtstring_value_arity_1(args[0])
}

func compiler__source_token__gtparse_token_arity_1(__arg0 flagrt.Value) flagrt.Value {
	var token = flagrt.Get(__arg0, flagKw_token)
	_ = token
	var line = flagrt.Get(__arg0, flagKw_line)
	_ = line
	var offset = flagrt.Get(__arg0, flagKw_offset)
	_ = offset
	var if_result_230 flagrt.Value
	if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(token, flagrt.NewString("(")))) {
		if_result_230 = flagrt.Call(compiler____gtParseToken, compiler__token_list_open, flagStr_, flagStr_, flagStr_, line, offset)
	} else {
		var if_result_229 flagrt.Value
		if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(token, flagrt.NewString(")")))) {
			if_result_229 = flagrt.Call(compiler____gtParseToken, compiler__token_list_close, flagStr_, flagStr_, flagStr_, line, offset)
		} else {
			var if_result_228 flagrt.Value
			if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(token, flagrt.NewString("[")))) {
				if_result_228 = flagrt.Call(compiler____gtParseToken, compiler__token_vector_open, flagStr_, flagStr_, flagStr_, line, offset)
			} else {
				var if_result_227 flagrt.Value
				if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(token, flagrt.NewString("]")))) {
					if_result_227 = flagrt.Call(compiler____gtParseToken, compiler__token_vector_close, flagStr_, flagStr_, flagStr_, line, offset)
				} else {
					var if_result_226 flagrt.Value
					if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(token, flagrt.NewString("|")))) {
						if_result_226 = flagrt.Call(compiler____gtParseToken, compiler__token_pipe, flagStr_, flagStr_, flagStr_, line, offset)
					} else {
						var if_result_225 flagrt.Value
						if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(token, flagrt.NewString("{")))) {
							if_result_225 = flagrt.Call(compiler____gtParseToken, compiler__token_map_open, flagStr_, flagStr_, flagStr_, line, offset)
						} else {
							var if_result_224 flagrt.Value
							if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(token, flagrt.NewString("}")))) {
								if_result_224 = flagrt.Call(compiler____gtParseToken, compiler__token_map_close, flagStr_, flagStr_, flagStr_, line, offset)
							} else {
								var if_result_223 flagrt.Value
								if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(token, flagrt.NewString("^")))) {
									if_result_223 = flagrt.Call(compiler____gtParseToken, compiler__token_metadata, flagStr_, flagStr_, flagStr_, line, offset)
								} else {
									var if_result_222 flagrt.Value
									if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(token, flagrt.NewString("'")))) {
										if_result_222 = flagrt.Call(compiler____gtParseToken, compiler__token_quote, flagStr_, flagStr_, flagStr_, line, offset)
									} else {
										var if_result_221 flagrt.Value
										if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(token, flagrt.NewString("#{")))) {
											if_result_221 = flagrt.Call(compiler____gtParseToken, compiler__token_dispatch_set_open, flagStr_, flagStr_, flagStr_, line, offset)
										} else {
											var if_result_220 flagrt.Value
											if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(token, flagrt.NewString("#(")))) {
												if_result_220 = flagrt.Call(compiler____gtParseToken, compiler__token_dispatch_fn_open, flagStr_, flagStr_, flagStr_, line, offset)
											} else {
												var if_result_219 flagrt.Value
												if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(token, flagrt.NewString("#")))) {
													if_result_219 = flagrt.Call(compiler____gtParseToken, compiler__token_error, flagStr_, flagStr_, flagStr_unexpected_end_after__, line, offset)
												} else {
													var let_result_216 flagrt.Value
													{
														var and_tmp = flagrt.Ge(flagrt.NewLong(int64(flagrt.Count(token))), flagrt.NewLong(2))
														var if_result_215 flagrt.Value
														if and_tmp {
															var let_result_214 flagrt.Value
															{
																var and_tmp = flagrt.NewBool(flagrt.Eq(flagrt.NewString("\""), flagrt.First(flagrt.Seq(token))))
																var if_result_213 flagrt.Value
																if flagrt.IsTruthy(and_tmp) {
																	if_result_213 = flagrt.NewBool(flagrt.Eq(flagrt.NewString("\""), flagrt.Last(flagrt.Seq(token))))
																} else {
																	if_result_213 = and_tmp
																}
																let_result_214 = if_result_213
															}
															if_result_215 = let_result_214
														} else {
															if_result_215 = flagrt.NewBool(and_tmp)
														}
														let_result_216 = if_result_215
													}
													var if_result_218 flagrt.Value
													if flagrt.IsTruthy(let_result_216) {
														if_result_218 = flagrt.Call(compiler____gtParseToken, compiler__token_string, flagStr_, flagrt.Call(compiler__token__gtstring_value, token), flagStr_, line, offset)
													} else {
														var if_result_217 flagrt.Value
														if flagrt.IsTruthy(flagrt.NewBool(true)) {
															if_result_217 = flagrt.Call(compiler____gtParseToken, compiler__token_atom, token, flagStr_, flagStr_, line, offset)
														} else {
															if_result_217 = flagrt.NilValue()
														}
														if_result_218 = if_result_217
													}
													if_result_219 = if_result_218
												}
												if_result_220 = if_result_219
											}
											if_result_221 = if_result_220
										}
										if_result_222 = if_result_221
									}
									if_result_223 = if_result_222
								}
								if_result_224 = if_result_223
							}
							if_result_225 = if_result_224
						}
						if_result_226 = if_result_225
					}
					if_result_227 = if_result_226
				}
				if_result_228 = if_result_227
			}
			if_result_229 = if_result_228
		}
		if_result_230 = if_result_229
	}
	return if_result_230
}

func compiler__source_token__gtparse_token_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__source_token__gtparse_token expects exactly 1 arguments")
	}
	return compiler__source_token__gtparse_token_arity_1(args[0])
}

func compiler__token_end_col_arity_1(source_token flagrt.Value) flagrt.Value {
	return flagrt.Add(flagrt.Call(flagKw_offset, source_token), flagrt.NewLong(int64(flagrt.Count(flagrt.Call(flagKw_token, source_token)))))
}

func compiler__token_end_col_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__token_end_col expects exactly 1 arguments")
	}
	return compiler__token_end_col_arity_1(args[0])
}

func compiler__tokenize_line_step_bang_arity_5(out flagrt.Value, chars flagrt.Value, line flagrt.Value, column flagrt.Value, state flagrt.Value) flagrt.Value {
	return func() flagrt.Value {
		var remaining = chars
		var col = column
		var current = state
		for {
			var if_result_251 flagrt.Value
			if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsEmpty(remaining))) {
				var if_result_231 flagrt.Value
				if flagrt.IsTruthy(flagrt.Call(flagKw_in_string, current)) {
					if_result_231 = current
				} else {
					_ = flagrt.Call(compiler__emit_token_bang, out, flagrt.Call(flagKw_token, current), flagrt.Call(flagKw_start_line, current), flagrt.Call(flagKw_start_offset, current))
					if_result_231 = compiler__empty_token_state
				}
				if_result_251 = if_result_231
			} else {
				var let_result_250 flagrt.Value
				{
					var ch = flagrt.First(remaining)
					var if_result_249 flagrt.Value
					if flagrt.IsTruthy(flagrt.Call(flagKw_in_string, current)) {
						var if_result_239 flagrt.Value
						if flagrt.IsTruthy(flagrt.Call(flagKw_triple, current)) {
							var if_result_233 flagrt.Value
							if flagrt.IsTruthy(flagrt.Call(compiler__starts_triple_quote_q, remaining)) {
								var let_result_232 flagrt.Value
								{
									var token = flagrt.Str(flagrt.Call(flagKw_token, current), "\"\"\"")
									_ = flagrt.Call(compiler__emit_token_bang, out, flagrt.NewString(token), flagrt.Call(flagKw_start_line, current), flagrt.Call(flagKw_start_offset, current))
									let_result_232 = flagrt.NewRecur(flagrt.Drop(flagrt.NewLong(3), remaining), flagrt.Add(col, flagrt.NewLong(3)), compiler__empty_token_state)
								}
								if_result_233 = let_result_232
							} else {
								if_result_233 = flagrt.NewRecur(flagrt.Rest(remaining), flagrt.Call(inc, col), flagrt.Call(compiler__make_state, flagrt.NewString(flagrt.Str(flagrt.Call(flagKw_token, current), ch)), flagrt.Call(flagKw_start_line, current), flagrt.Call(flagKw_start_offset, current), flagrt.NewBool(true), flagrt.NewBool(true), flagrt.NewBool(false)))
							}
							if_result_239 = if_result_233
						} else {
							var let_result_238 flagrt.Value
							{
								var token = flagrt.Str(flagrt.Call(flagKw_token, current), ch)
								var if_result_237 flagrt.Value
								if flagrt.IsTruthy(flagrt.Call(flagKw_escaped, current)) {
									if_result_237 = flagrt.NewRecur(flagrt.Rest(remaining), flagrt.Call(inc, col), flagrt.Call(compiler__make_state, flagrt.NewString(token), flagrt.Call(flagKw_start_line, current), flagrt.Call(flagKw_start_offset, current), flagrt.NewBool(true), flagrt.NewBool(false), flagrt.NewBool(false)))
								} else {
									var if_result_236 flagrt.Value
									if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagrt.NewString("\\"), ch))) {
										if_result_236 = flagrt.NewRecur(flagrt.Rest(remaining), flagrt.Call(inc, col), flagrt.Call(compiler__make_state, flagrt.NewString(token), flagrt.Call(flagKw_start_line, current), flagrt.Call(flagKw_start_offset, current), flagrt.NewBool(true), flagrt.NewBool(false), flagrt.NewBool(true)))
									} else {
										var if_result_235 flagrt.Value
										if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagrt.NewString("\""), ch))) {
											_ = flagrt.Call(compiler__emit_token_bang, out, flagrt.NewString(token), flagrt.Call(flagKw_start_line, current), flagrt.Call(flagKw_start_offset, current))
											if_result_235 = flagrt.NewRecur(flagrt.Rest(remaining), flagrt.Call(inc, col), compiler__empty_token_state)
										} else {
											var if_result_234 flagrt.Value
											if flagrt.IsTruthy(flagrt.NewBool(true)) {
												if_result_234 = flagrt.NewRecur(flagrt.Rest(remaining), flagrt.Call(inc, col), flagrt.Call(compiler__make_state, flagrt.NewString(token), flagrt.Call(flagKw_start_line, current), flagrt.Call(flagKw_start_offset, current), flagrt.NewBool(true), flagrt.NewBool(false), flagrt.NewBool(false)))
											} else {
												if_result_234 = flagrt.NilValue()
											}
											if_result_235 = if_result_234
										}
										if_result_236 = if_result_235
									}
									if_result_237 = if_result_236
								}
								let_result_238 = if_result_237
							}
							if_result_239 = let_result_238
						}
						if_result_249 = if_result_239
					} else {
						var if_result_248 flagrt.Value
						if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagrt.NewString(";"), ch))) {
							_ = flagrt.Call(compiler__emit_token_bang, out, flagrt.Call(flagKw_token, current), flagrt.Call(flagKw_start_line, current), flagrt.Call(flagKw_start_offset, current))
							if_result_248 = compiler__empty_token_state
						} else {
							var if_result_247 flagrt.Value
							if flagrt.IsTruthy(flagrt.Call(compiler__whitespace_char_q, ch)) {
								_ = flagrt.Call(compiler__emit_token_bang, out, flagrt.Call(flagKw_token, current), flagrt.Call(flagKw_start_line, current), flagrt.Call(flagKw_start_offset, current))
								if_result_247 = flagrt.NewRecur(flagrt.Rest(remaining), flagrt.Call(inc, col), compiler__empty_token_state)
							} else {
								var if_result_246 flagrt.Value
								if flagrt.IsTruthy(flagrt.Call(compiler__delimiter_char_q, ch)) {
									_ = flagrt.Call(compiler__emit_token_bang, out, flagrt.Call(flagKw_token, current), flagrt.Call(flagKw_start_line, current), flagrt.Call(flagKw_start_offset, current))
									_ = flagrt.Call(compiler__emit_token_bang, out, ch, line, col)
									if_result_246 = flagrt.NewRecur(flagrt.Rest(remaining), flagrt.Call(inc, col), compiler__empty_token_state)
								} else {
									var if_result_245 flagrt.Value
									if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagrt.NewString("\""), ch))) {
										var if_result_240 flagrt.Value
										if flagrt.IsTruthy(flagrt.Call(compiler__starts_triple_quote_q, remaining)) {
											_ = flagrt.Call(compiler__emit_token_bang, out, flagrt.Call(flagKw_token, current), flagrt.Call(flagKw_start_line, current), flagrt.Call(flagKw_start_offset, current))
											if_result_240 = flagrt.NewRecur(flagrt.Drop(flagrt.NewLong(3), remaining), flagrt.Add(col, flagrt.NewLong(3)), flagrt.Call(compiler__make_state, flagStr____, line, col, flagrt.NewBool(true), flagrt.NewBool(true), flagrt.NewBool(false)))
										} else {
											_ = flagrt.Call(compiler__emit_token_bang, out, flagrt.Call(flagKw_token, current), flagrt.Call(flagKw_start_line, current), flagrt.Call(flagKw_start_offset, current))
											if_result_240 = flagrt.NewRecur(flagrt.Rest(remaining), flagrt.Call(inc, col), flagrt.Call(compiler__make_state, flagStr__, line, col, flagrt.NewBool(true), flagrt.NewBool(false), flagrt.NewBool(false)))
										}
										if_result_245 = if_result_240
									} else {
										var if_result_244 flagrt.Value
										if flagrt.IsTruthy(flagrt.Call(compiler__starts_dispatch_set_q, remaining)) {
											_ = flagrt.Call(compiler__emit_token_bang, out, flagrt.Call(flagKw_token, current), flagrt.Call(flagKw_start_line, current), flagrt.Call(flagKw_start_offset, current))
											_ = flagrt.Call(compiler__emit_token_bang, out, flagStr___, line, col)
											if_result_244 = flagrt.NewRecur(flagrt.Drop(flagrt.NewLong(2), remaining), flagrt.Add(col, flagrt.NewLong(2)), compiler__empty_token_state)
										} else {
											var if_result_243 flagrt.Value
											if flagrt.IsTruthy(flagrt.Call(compiler__starts_dispatch_fn_q, remaining)) {
												_ = flagrt.Call(compiler__emit_token_bang, out, flagrt.Call(flagKw_token, current), flagrt.Call(flagKw_start_line, current), flagrt.Call(flagKw_start_offset, current))
												_ = flagrt.Call(compiler__emit_token_bang, out, flagStr____1, line, col)
												if_result_243 = flagrt.NewRecur(flagrt.Drop(flagrt.NewLong(2), remaining), flagrt.Add(col, flagrt.NewLong(2)), compiler__empty_token_state)
											} else {
												var if_result_242 flagrt.Value
												if flagrt.IsTruthy(flagrt.NewBool(true)) {
													var if_result_241 flagrt.Value
													if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsEmpty(flagrt.Call(flagKw_token, current)))) {
														if_result_241 = flagrt.NewRecur(flagrt.Rest(remaining), flagrt.Call(inc, col), flagrt.Call(compiler__make_state, ch, line, col, flagrt.NewBool(false), flagrt.NewBool(false), flagrt.NewBool(false)))
													} else {
														if_result_241 = flagrt.NewRecur(flagrt.Rest(remaining), flagrt.Call(inc, col), flagrt.Call(compiler__make_state, flagrt.NewString(flagrt.Str(flagrt.Call(flagKw_token, current), ch)), flagrt.Call(flagKw_start_line, current), flagrt.Call(flagKw_start_offset, current), flagrt.NewBool(false), flagrt.NewBool(false), flagrt.NewBool(false)))
													}
													if_result_242 = if_result_241
												} else {
													if_result_242 = flagrt.NilValue()
												}
												if_result_243 = if_result_242
											}
											if_result_244 = if_result_243
										}
										if_result_245 = if_result_244
									}
									if_result_246 = if_result_245
								}
								if_result_247 = if_result_246
							}
							if_result_248 = if_result_247
						}
						if_result_249 = if_result_248
					}
					let_result_250 = if_result_249
				}
				if_result_251 = let_result_250
			}
			__loopResult := if_result_251
			if __recurValues, __isRecur := flagrt.UnwrapRecur(__loopResult); __isRecur {
				if len(__recurValues) != 3 {
					panic("internal error: recur arity mismatch")
				}
				remaining = __recurValues[0]
				col = __recurValues[1]
				current = __recurValues[2]
				continue
			}
			return __loopResult
		}
	}()
}

func compiler__tokenize_line_step_bang_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 5 {
		panic("compiler__tokenize_line_step_bang expects exactly 5 arguments")
	}
	return compiler__tokenize_line_step_bang_arity_5(args[0], args[1], args[2], args[3], args[4])
}

func compiler__tokenize_lines_bang_arity_3(out flagrt.Value, lines flagrt.Value, line_number flagrt.Value) flagrt.Value {
	return func() flagrt.Value {
		var remaining = lines
		var line = line_number
		var state = compiler__empty_token_state
		for {
			var if_result_257 flagrt.Value
			if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsEmpty(remaining))) {
				if_result_257 = flagrt.Call(compiler__emit_token_bang, out, flagrt.Call(flagKw_token, state), flagrt.Call(flagKw_start_line, state), flagrt.Call(flagKw_start_offset, state))
			} else {
				var let_result_256 flagrt.Value
				{
					var next_state = flagrt.Call(compiler__tokenize_line_step_bang, out, flagrt.Seq(flagrt.NewString(flagrt.Str(flagrt.First(remaining)))), line, flagrt.NewLong(1), state)
					var if_result_252 flagrt.Value
					if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsEmpty(flagrt.Rest(remaining)))) {
						if_result_252 = flagrt.NewBool(false)
					} else {
						if_result_252 = flagrt.NewBool(true)
					}
					var has_more = if_result_252
					var let_result_254 flagrt.Value
					{
						var and_tmp = has_more
						var if_result_253 flagrt.Value
						if flagrt.IsTruthy(and_tmp) {
							if_result_253 = flagrt.Call(flagKw_in_string, next_state)
						} else {
							if_result_253 = and_tmp
						}
						let_result_254 = if_result_253
					}
					var if_result_255 flagrt.Value
					if flagrt.IsTruthy(let_result_254) {
						if_result_255 = flagrt.Call(compiler__make_state, flagrt.NewString(flagrt.Str(flagrt.Call(flagKw_token, next_state), "\n")), flagrt.Call(flagKw_start_line, next_state), flagrt.Call(flagKw_start_offset, next_state), flagrt.NewBool(true), flagrt.Call(flagKw_triple, next_state), flagrt.Call(flagKw_escaped, next_state))
					} else {
						if_result_255 = next_state
					}
					var carry_state = if_result_255
					let_result_256 = flagrt.NewRecur(flagrt.Rest(remaining), flagrt.Add(line, flagrt.NewLong(1)), carry_state)
				}
				if_result_257 = let_result_256
			}
			__loopResult := if_result_257
			if __recurValues, __isRecur := flagrt.UnwrapRecur(__loopResult); __isRecur {
				if len(__recurValues) != 3 {
					panic("internal error: recur arity mismatch")
				}
				remaining = __recurValues[0]
				line = __recurValues[1]
				state = __recurValues[2]
				continue
			}
			return __loopResult
		}
	}()
}

func compiler__tokenize_lines_bang_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 3 {
		panic("compiler__tokenize_lines_bang expects exactly 3 arguments")
	}
	return compiler__tokenize_lines_bang_arity_3(args[0], args[1], args[2])
}

func compiler__split_lines_arity_1(source flagrt.Value) flagrt.Value {
	return func() flagrt.Value {
		var chars = flagrt.Seq(source)
		var line = flagStr_
		var lines = flagVec
		for {
			var if_result_261 flagrt.Value
			if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsEmpty(chars))) {
				var if_result_258 flagrt.Value
				if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsEmpty(line))) {
					if_result_258 = lines
				} else {
					if_result_258 = flagrt.Conj(lines, line)
				}
				if_result_261 = if_result_258
			} else {
				var let_result_260 flagrt.Value
				{
					var ch = flagrt.First(chars)
					var if_result_259 flagrt.Value
					if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagrt.NewString("\n"), ch))) {
						if_result_259 = flagrt.NewRecur(flagrt.Rest(chars), flagStr_, flagrt.Conj(lines, line))
					} else {
						if_result_259 = flagrt.NewRecur(flagrt.Rest(chars), flagrt.NewString(flagrt.Str(line, ch)), lines)
					}
					let_result_260 = if_result_259
				}
				if_result_261 = let_result_260
			}
			__loopResult := if_result_261
			if __recurValues, __isRecur := flagrt.UnwrapRecur(__loopResult); __isRecur {
				if len(__recurValues) != 3 {
					panic("internal error: recur arity mismatch")
				}
				chars = __recurValues[0]
				line = __recurValues[1]
				lines = __recurValues[2]
				continue
			}
			return __loopResult
		}
	}()
}

func compiler__split_lines_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__split_lines expects exactly 1 arguments")
	}
	return compiler__split_lines_arity_1(args[0])
}

// Tokenize an in-memory source string and return a channel of SourceToken maps {:token :line :offset}.
func compiler__tokenize_source_arity_1(source flagrt.Value) flagrt.Value {
	var let_result_262 flagrt.Value
	{
		var out = flagrt.Call(async__make_channel, flagrt.NewLong(64))
		_ = flagrt.Call(async__go_run, flagrt.NewFunction(func(args ...flagrt.Value) flagrt.Value {
			if len(args) != 0 {
				panic("fn expects exactly 0 arguments")
			}
			_ = flagrt.Call(compiler__tokenize_lines_bang, out, flagrt.Call(compiler__split_lines, source), flagrt.NewLong(1))
			return flagrt.Call(async__channel_close, out)
		}))
		let_result_262 = out
	}
	return let_result_262
}

func compiler__tokenize_source_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__tokenize_source expects exactly 1 arguments")
	}
	return compiler__tokenize_source_arity_1(args[0])
}

// Read a source file and return a channel of token maps {:token :line :offset}.
func compiler__tokenize_file_arity_1(path flagrt.Value) flagrt.Value {
	var let_result_263 flagrt.Value
	{
		var out = flagrt.Call(async__make_channel, flagrt.NewLong(64))
		_ = flagrt.Call(async__go_run, flagrt.NewFunction(func(args ...flagrt.Value) flagrt.Value {
			if len(args) != 0 {
				panic("fn expects exactly 0 arguments")
			}
			return func() flagrt.Value {
				var rdr = flagrt.Call(flagrt.GoBind_runtime_OpenFile, path)
				defer flagrt.Call(flagrt.NewFunction(func(args ...flagrt.Value) flagrt.Value {
					if len(args) != 0 {
						panic("fn expects exactly 0 arguments")
					}
					return flagrt.Call(close, rdr)
				}))
				_ = flagrt.Call(compiler__tokenize_lines_bang, out, flagrt.LineSeq(rdr), flagrt.NewLong(1))
				return flagrt.Call(async__channel_close, out)
			}()
		}))
		let_result_263 = out
	}
	return let_result_263
}

func compiler__tokenize_file_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__tokenize_file expects exactly 1 arguments")
	}
	return compiler__tokenize_file_arity_1(args[0])
}

// Read a source file and return a channel of ParseToken-shaped maps matching Go ParseToken fields.
func compiler__tokenize_file_parse_tokens_arity_1(path flagrt.Value) flagrt.Value {
	var let_result_269 flagrt.Value
	{
		var out = flagrt.Call(async__make_channel, flagrt.NewLong(64))
		var in = flagrt.Call(compiler__tokenize_file, path)
		_ = flagrt.Call(async__go_run, flagrt.NewFunction(func(args ...flagrt.Value) flagrt.Value {
			if len(args) != 0 {
				panic("fn expects exactly 0 arguments")
			}
			return func() flagrt.Value {
				var last_token = flagrt.NilValue()
				for {
					var let_result_268 flagrt.Value
					{
						var source_token = flagrt.Call(async__channel_receive, in)
						var if_result_267 flagrt.Value
						if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsNil(source_token))) {
							var let_result_266 flagrt.Value
							{
								var if_result_264 flagrt.Value
								if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsNil(last_token))) {
									if_result_264 = flagrt.NewLong(1)
								} else {
									if_result_264 = flagrt.Call(flagKw_line, last_token)
								}
								var eof_line = if_result_264
								var if_result_265 flagrt.Value
								if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsNil(last_token))) {
									if_result_265 = flagrt.NewLong(1)
								} else {
									if_result_265 = flagrt.Call(compiler__token_end_col, last_token)
								}
								var eof_col = if_result_265
								_ = flagrt.Call(async__channel_send, out, flagrt.Call(compiler____gtParseToken, compiler__token_eof, flagStr_, flagStr_, flagStr_, eof_line, eof_col))
								let_result_266 = flagrt.Call(async__channel_close, out)
							}
							if_result_267 = let_result_266
						} else {
							_ = flagrt.Call(async__channel_send, out, flagrt.Call(compiler__source_token__gtparse_token, source_token))
							if_result_267 = flagrt.NewRecur(source_token)
						}
						let_result_268 = if_result_267
					}
					__loopResult := let_result_268
					if __recurValues, __isRecur := flagrt.UnwrapRecur(__loopResult); __isRecur {
						if len(__recurValues) != 1 {
							panic("internal error: recur arity mismatch")
						}
						last_token = __recurValues[0]
						continue
					}
					return __loopResult
				}
			}()
		}))
		let_result_269 = out
	}
	return let_result_269
}

func compiler__tokenize_file_parse_tokens_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__tokenize_file_parse_tokens expects exactly 1 arguments")
	}
	return compiler__tokenize_file_parse_tokens_arity_1(args[0])
}

func compiler__parse_error_arity_3(line flagrt.Value, col flagrt.Value, msg flagrt.Value) flagrt.Value {
	return func() flagrt.Value {
		flagrt.Throw(flagrt.NewString(flagrt.Str("parse error at ", line, ":", col, ": ", msg)))
		return flagrt.NilValue()
	}()
}

func compiler__parse_error_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 3 {
		panic("compiler__parse_error expects exactly 3 arguments")
	}
	return compiler__parse_error_arity_3(args[0], args[1], args[2])
}

func compiler__token_line_arity_1(tok flagrt.Value) flagrt.Value {
	return flagrt.Call(flagKw_line, tok)
}

func compiler__token_line_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__token_line expects exactly 1 arguments")
	}
	return compiler__token_line_arity_1(args[0])
}

func compiler__token_col_arity_1(tok flagrt.Value) flagrt.Value {
	return flagrt.Call(flagKw_offset, tok)
}

func compiler__token_col_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__token_col expects exactly 1 arguments")
	}
	return compiler__token_col_arity_1(args[0])
}

func compiler__make_parser_arity_1(in flagrt.Value) flagrt.Value {
	return flagrt.NewMap(flagKw_in, in, flagKw_la, flagrt.Call(async__atom, flagKw_none), flagKw_last, flagrt.Call(async__atom, flagrt.NilValue()))
}

func compiler__make_parser_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__make_parser expects exactly 1 arguments")
	}
	return compiler__make_parser_arity_1(args[0])
}

func compiler__peek_arity_1(p flagrt.Value) flagrt.Value {
	var let_result_273 flagrt.Value
	{
		var la = flagrt.Call(async__deref, flagrt.Call(flagKw_la, p))
		var if_result_272 flagrt.Value
		if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagKw_none, la))) {
			var let_result_271 flagrt.Value
			{
				var t = flagrt.Call(async__channel_receive, flagrt.Call(flagKw_in, p))
				var if_result_270 flagrt.Value
				if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsNil(t))) {
					_ = flagrt.Call(async__reset_bang, flagrt.Call(flagKw_la, p), flagKw_eof)
					if_result_270 = flagKw_eof
				} else {
					_ = flagrt.Call(async__reset_bang, flagrt.Call(flagKw_last, p), t)
					_ = flagrt.Call(async__reset_bang, flagrt.Call(flagKw_la, p), t)
					if_result_270 = t
				}
				let_result_271 = if_result_270
			}
			if_result_272 = let_result_271
		} else {
			if_result_272 = la
		}
		let_result_273 = if_result_272
	}
	return let_result_273
}

func compiler__peek_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__peek expects exactly 1 arguments")
	}
	return compiler__peek_arity_1(args[0])
}

func compiler__next_tok_arity_1(p flagrt.Value) flagrt.Value {
	var let_result_274 flagrt.Value
	{
		var t = flagrt.Call(compiler__peek, p)
		_ = flagrt.Call(async__reset_bang, flagrt.Call(flagKw_la, p), flagKw_none)
		let_result_274 = t
	}
	return let_result_274
}

func compiler__next_tok_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__next_tok expects exactly 1 arguments")
	}
	return compiler__next_tok_arity_1(args[0])
}

func compiler__eof_line_col_arity_1(p flagrt.Value) flagrt.Value {
	var let_result_276 flagrt.Value
	{
		var last = flagrt.Call(async__deref, flagrt.Call(flagKw_last, p))
		var if_result_275 flagrt.Value
		if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsNil(last))) {
			if_result_275 = flagVec_1
		} else {
			if_result_275 = flagrt.NewArray(flagrt.Call(flagKw_line, last), flagrt.Add(flagrt.Call(flagKw_offset, last), flagrt.NewLong(int64(flagrt.Count(flagrt.Call(flagKw_token, last))))))
		}
		let_result_276 = if_result_275
	}
	return let_result_276
}

func compiler__eof_line_col_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__eof_line_col expects exactly 1 arguments")
	}
	return compiler__eof_line_col_arity_1(args[0])
}

func compiler__starts_with_char_q_arity_2(token flagrt.Value, ch flagrt.Value) flagrt.Value {
	var let_result_279 flagrt.Value
	{
		var if_result_277 flagrt.Value
		if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsEmpty(token))) {
			if_result_277 = flagrt.NewBool(false)
		} else {
			if_result_277 = flagrt.NewBool(true)
		}
		var and_tmp = if_result_277
		var if_result_278 flagrt.Value
		if flagrt.IsTruthy(and_tmp) {
			if_result_278 = flagrt.NewBool(flagrt.Eq(ch, flagrt.First(flagrt.Seq(token))))
		} else {
			if_result_278 = and_tmp
		}
		let_result_279 = if_result_278
	}
	return let_result_279
}

func compiler__starts_with_char_q_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 2 {
		panic("compiler__starts_with_char_q expects exactly 2 arguments")
	}
	return compiler__starts_with_char_q_arity_2(args[0], args[1])
}

func compiler__ends_with_char_q_arity_2(token flagrt.Value, ch flagrt.Value) flagrt.Value {
	var let_result_282 flagrt.Value
	{
		var if_result_280 flagrt.Value
		if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsEmpty(token))) {
			if_result_280 = flagrt.NewBool(false)
		} else {
			if_result_280 = flagrt.NewBool(true)
		}
		var and_tmp = if_result_280
		var if_result_281 flagrt.Value
		if flagrt.IsTruthy(and_tmp) {
			if_result_281 = flagrt.NewBool(flagrt.Eq(ch, flagrt.Last(flagrt.Seq(token))))
		} else {
			if_result_281 = and_tmp
		}
		let_result_282 = if_result_281
	}
	return let_result_282
}

func compiler__ends_with_char_q_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 2 {
		panic("compiler__ends_with_char_q expects exactly 2 arguments")
	}
	return compiler__ends_with_char_q_arity_2(args[0], args[1])
}

func compiler__contains_float_mark_q_arity_1(token flagrt.Value) flagrt.Value {
	return func() flagrt.Value {
		var cs = flagrt.Seq(token)
		for {
			var if_result_289 flagrt.Value
			if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsEmpty(cs))) {
				if_result_289 = flagrt.NewBool(false)
			} else {
				var let_result_288 flagrt.Value
				{
					var ch = flagrt.First(cs)
					var let_result_286 flagrt.Value
					{
						var or_tmp = flagrt.NewBool(flagrt.Eq(ch, flagrt.NewString(".")))
						var if_result_285 flagrt.Value
						if flagrt.IsTruthy(or_tmp) {
							if_result_285 = or_tmp
						} else {
							var let_result_284 flagrt.Value
							{
								var or_tmp = flagrt.NewBool(flagrt.Eq(ch, flagrt.NewString("e")))
								var if_result_283 flagrt.Value
								if flagrt.IsTruthy(or_tmp) {
									if_result_283 = or_tmp
								} else {
									if_result_283 = flagrt.NewBool(flagrt.Eq(ch, flagrt.NewString("E")))
								}
								let_result_284 = if_result_283
							}
							if_result_285 = let_result_284
						}
						let_result_286 = if_result_285
					}
					var if_result_287 flagrt.Value
					if flagrt.IsTruthy(let_result_286) {
						if_result_287 = flagrt.NewBool(true)
					} else {
						if_result_287 = flagrt.NewRecur(flagrt.Rest(cs))
					}
					let_result_288 = if_result_287
				}
				if_result_289 = let_result_288
			}
			__loopResult := if_result_289
			if __recurValues, __isRecur := flagrt.UnwrapRecur(__loopResult); __isRecur {
				if len(__recurValues) != 1 {
					panic("internal error: recur arity mismatch")
				}
				cs = __recurValues[0]
				continue
			}
			return __loopResult
		}
	}()
}

func compiler__contains_float_mark_q_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__contains_float_mark_q expects exactly 1 arguments")
	}
	return compiler__contains_float_mark_q_arity_1(args[0])
}

func compiler__slash_count_arity_1(token flagrt.Value) flagrt.Value {
	return func() flagrt.Value {
		var cs = flagrt.Seq(token)
		var n = flagrt.NewLong(0)
		for {
			var if_result_291 flagrt.Value
			if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsEmpty(cs))) {
				if_result_291 = n
			} else {
				var if_result_290 flagrt.Value
				if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagrt.NewString("/"), flagrt.First(cs)))) {
					if_result_290 = flagrt.Add(n, flagrt.NewLong(1))
				} else {
					if_result_290 = n
				}
				if_result_291 = flagrt.NewRecur(flagrt.Rest(cs), if_result_290)
			}
			__loopResult := if_result_291
			if __recurValues, __isRecur := flagrt.UnwrapRecur(__loopResult); __isRecur {
				if len(__recurValues) != 2 {
					panic("internal error: recur arity mismatch")
				}
				cs = __recurValues[0]
				n = __recurValues[1]
				continue
			}
			return __loopResult
		}
	}()
}

func compiler__slash_count_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__slash_count expects exactly 1 arguments")
	}
	return compiler__slash_count_arity_1(args[0])
}

func compiler__split_once_arity_2(token flagrt.Value, sep flagrt.Value) flagrt.Value {
	return func() flagrt.Value {
		var cs = flagrt.Seq(token)
		var left = flagStr_
		for {
			var if_result_294 flagrt.Value
			if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsEmpty(cs))) {
				if_result_294 = flagrt.NilValue()
			} else {
				var if_result_293 flagrt.Value
				if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(sep, flagrt.First(cs)))) {
					if_result_293 = func() flagrt.Value {
						var r = flagrt.Rest(cs)
						var right = flagStr_
						for {
							var if_result_292 flagrt.Value
							if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsEmpty(r))) {
								if_result_292 = flagrt.NewArray(left, right)
							} else {
								if_result_292 = flagrt.NewRecur(flagrt.Rest(r), flagrt.NewString(flagrt.Str(right, flagrt.First(r))))
							}
							__loopResult := if_result_292
							if __recurValues, __isRecur := flagrt.UnwrapRecur(__loopResult); __isRecur {
								if len(__recurValues) != 2 {
									panic("internal error: recur arity mismatch")
								}
								r = __recurValues[0]
								right = __recurValues[1]
								continue
							}
							return __loopResult
						}
					}()
				} else {
					if_result_293 = flagrt.NewRecur(flagrt.Rest(cs), flagrt.NewString(flagrt.Str(left, flagrt.First(cs))))
				}
				if_result_294 = if_result_293
			}
			__loopResult := if_result_294
			if __recurValues, __isRecur := flagrt.UnwrapRecur(__loopResult); __isRecur {
				if len(__recurValues) != 2 {
					panic("internal error: recur arity mismatch")
				}
				cs = __recurValues[0]
				left = __recurValues[1]
				continue
			}
			return __loopResult
		}
	}()
}

func compiler__split_once_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 2 {
		panic("compiler__split_once expects exactly 2 arguments")
	}
	return compiler__split_once_arity_2(args[0], args[1])
}

func compiler__unescape_char_arity_1(ch flagrt.Value) flagrt.Value {
	var if_result_298 flagrt.Value
	if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(ch, flagrt.NewString("n")))) {
		if_result_298 = flagStr___3
	} else {
		var if_result_297 flagrt.Value
		if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(ch, flagrt.NewString("t")))) {
			if_result_297 = flagStr___2
		} else {
			var if_result_296 flagrt.Value
			if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(ch, flagrt.NewString("r")))) {
				if_result_296 = flagStr___1
			} else {
				var if_result_295 flagrt.Value
				if flagrt.IsTruthy(flagrt.NewBool(true)) {
					if_result_295 = ch
				} else {
					if_result_295 = flagrt.NilValue()
				}
				if_result_296 = if_result_295
			}
			if_result_297 = if_result_296
		}
		if_result_298 = if_result_297
	}
	return if_result_298
}

func compiler__unescape_char_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__unescape_char expects exactly 1 arguments")
	}
	return compiler__unescape_char_arity_1(args[0])
}

func compiler__unquote_go_string_arity_1(token flagrt.Value) flagrt.Value {
	var let_result_312 flagrt.Value
	{
		var chars = flagrt.Seq(token)
		var n = flagrt.NewLong(int64(flagrt.Count(chars)))
		var let_result_304 flagrt.Value
		{
			var or_tmp = flagrt.Lt(n, flagrt.NewLong(2))
			var if_result_303 flagrt.Value
			if or_tmp {
				if_result_303 = flagrt.NewBool(or_tmp)
			} else {
				var let_result_302 flagrt.Value
				{
					var if_result_299 flagrt.Value
					if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagrt.NewString("\""), flagrt.First(chars)))) {
						if_result_299 = flagrt.NewBool(false)
					} else {
						if_result_299 = flagrt.NewBool(true)
					}
					var or_tmp = if_result_299
					var if_result_301 flagrt.Value
					if flagrt.IsTruthy(or_tmp) {
						if_result_301 = or_tmp
					} else {
						var if_result_300 flagrt.Value
						if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagrt.NewString("\""), flagrt.Last(chars)))) {
							if_result_300 = flagrt.NewBool(false)
						} else {
							if_result_300 = flagrt.NewBool(true)
						}
						if_result_301 = if_result_300
					}
					let_result_302 = if_result_301
				}
				if_result_303 = let_result_302
			}
			let_result_304 = if_result_303
		}
		var if_result_311 flagrt.Value
		if flagrt.IsTruthy(let_result_304) {
			if_result_311 = flagKw_error
		} else {
			if_result_311 = func() flagrt.Value {
				var remaining = flagrt.Take(flagrt.Sub(n, flagrt.NewLong(2)), flagrt.Drop(flagrt.NewLong(1), chars))
				var out = flagStr_
				var escaped = flagrt.NewBool(false)
				for {
					var if_result_310 flagrt.Value
					if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsEmpty(remaining))) {
						var if_result_305 flagrt.Value
						if flagrt.IsTruthy(escaped) {
							if_result_305 = flagKw_error
						} else {
							if_result_305 = out
						}
						if_result_310 = if_result_305
					} else {
						var let_result_309 flagrt.Value
						{
							var ch = flagrt.First(remaining)
							var if_result_308 flagrt.Value
							if flagrt.IsTruthy(escaped) {
								if_result_308 = flagrt.NewRecur(flagrt.Rest(remaining), flagrt.NewString(flagrt.Str(out, flagrt.Call(compiler__unescape_char, ch))), flagrt.NewBool(false))
							} else {
								var if_result_307 flagrt.Value
								if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(ch, flagrt.NewString("\\")))) {
									if_result_307 = flagrt.NewRecur(flagrt.Rest(remaining), out, flagrt.NewBool(true))
								} else {
									var if_result_306 flagrt.Value
									if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(ch, flagrt.NewString("\"")))) {
										if_result_306 = flagKw_error
									} else {
										if_result_306 = flagrt.NewRecur(flagrt.Rest(remaining), flagrt.NewString(flagrt.Str(out, ch)), flagrt.NewBool(false))
									}
									if_result_307 = if_result_306
								}
								if_result_308 = if_result_307
							}
							let_result_309 = if_result_308
						}
						if_result_310 = let_result_309
					}
					__loopResult := if_result_310
					if __recurValues, __isRecur := flagrt.UnwrapRecur(__loopResult); __isRecur {
						if len(__recurValues) != 3 {
							panic("internal error: recur arity mismatch")
						}
						remaining = __recurValues[0]
						out = __recurValues[1]
						escaped = __recurValues[2]
						continue
					}
					return __loopResult
				}
			}()
		}
		let_result_312 = if_result_311
	}
	return let_result_312
}

func compiler__unquote_go_string_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__unquote_go_string expects exactly 1 arguments")
	}
	return compiler__unquote_go_string_arity_1(args[0])
}

func compiler__string_token_q_arity_1(tok flagrt.Value) flagrt.Value {
	return flagrt.Call(compiler__starts_with_char_q, flagrt.Call(flagKw_token, tok), flagStr__)
}

func compiler__string_token_q_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__string_token_q expects exactly 1 arguments")
	}
	return compiler__string_token_q_arity_1(args[0])
}

func compiler__strip_edges_arity_3(token flagrt.Value, left flagrt.Value, right flagrt.Value) flagrt.Value {
	var let_result_315 flagrt.Value
	{
		var chars = flagrt.Seq(token)
		var n = flagrt.NewLong(int64(flagrt.Count(chars)))
		var if_result_314 flagrt.Value
		if flagrt.Le(n, flagrt.Add(left, right)) {
			if_result_314 = flagStr_
		} else {
			if_result_314 = func() flagrt.Value {
				var remaining = flagrt.Take(flagrt.Sub(flagrt.Sub(n, left), right), flagrt.Drop(left, chars))
				var out = flagStr_
				for {
					var if_result_313 flagrt.Value
					if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsEmpty(remaining))) {
						if_result_313 = out
					} else {
						if_result_313 = flagrt.NewRecur(flagrt.Rest(remaining), flagrt.NewString(flagrt.Str(out, flagrt.First(remaining))))
					}
					__loopResult := if_result_313
					if __recurValues, __isRecur := flagrt.UnwrapRecur(__loopResult); __isRecur {
						if len(__recurValues) != 2 {
							panic("internal error: recur arity mismatch")
						}
						remaining = __recurValues[0]
						out = __recurValues[1]
						continue
					}
					return __loopResult
				}
			}()
		}
		let_result_315 = if_result_314
	}
	return let_result_315
}

func compiler__strip_edges_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 3 {
		panic("compiler__strip_edges expects exactly 3 arguments")
	}
	return compiler__strip_edges_arity_3(args[0], args[1], args[2])
}

func compiler__triple_quoted_q_arity_1(token flagrt.Value) flagrt.Value {
	var let_result_322 flagrt.Value
	{
		var chars = flagrt.Seq(token)
		var n = flagrt.NewLong(int64(flagrt.Count(chars)))
		var let_result_321 flagrt.Value
		{
			var and_tmp = flagrt.Ge(n, flagrt.NewLong(6))
			var if_result_320 flagrt.Value
			if and_tmp {
				var let_result_319 flagrt.Value
				{
					var and_tmp = flagrt.NewBool(flagrt.Eq(flagrt.NewString("\""), flagrt.Call(flagrt.BuiltinFunction("nth"), chars, flagrt.NewLong(0))))
					var if_result_318 flagrt.Value
					if flagrt.IsTruthy(and_tmp) {
						var let_result_317 flagrt.Value
						{
							var and_tmp = flagrt.NewBool(flagrt.Eq(flagrt.NewString("\""), flagrt.Call(flagrt.BuiltinFunction("nth"), chars, flagrt.NewLong(1))))
							var if_result_316 flagrt.Value
							if flagrt.IsTruthy(and_tmp) {
								if_result_316 = flagrt.NewBool(flagrt.Eq(flagrt.NewString("\""), flagrt.Call(flagrt.BuiltinFunction("nth"), chars, flagrt.NewLong(2))))
							} else {
								if_result_316 = and_tmp
							}
							let_result_317 = if_result_316
						}
						if_result_318 = let_result_317
					} else {
						if_result_318 = and_tmp
					}
					let_result_319 = if_result_318
				}
				if_result_320 = let_result_319
			} else {
				if_result_320 = flagrt.NewBool(and_tmp)
			}
			let_result_321 = if_result_320
		}
		let_result_322 = let_result_321
	}
	return let_result_322
}

func compiler__triple_quoted_q_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__triple_quoted_q expects exactly 1 arguments")
	}
	return compiler__triple_quoted_q_arity_1(args[0])
}

func compiler__triple_quoted_closed_q_arity_1(token flagrt.Value) flagrt.Value {
	var let_result_329 flagrt.Value
	{
		var chars = flagrt.Seq(token)
		var n = flagrt.NewLong(int64(flagrt.Count(chars)))
		var let_result_328 flagrt.Value
		{
			var and_tmp = flagrt.Ge(n, flagrt.NewLong(6))
			var if_result_327 flagrt.Value
			if and_tmp {
				var let_result_326 flagrt.Value
				{
					var and_tmp = flagrt.NewBool(flagrt.Eq(flagrt.NewString("\""), flagrt.Call(flagrt.BuiltinFunction("nth"), chars, flagrt.Sub(n, flagrt.NewLong(1)))))
					var if_result_325 flagrt.Value
					if flagrt.IsTruthy(and_tmp) {
						var let_result_324 flagrt.Value
						{
							var and_tmp = flagrt.NewBool(flagrt.Eq(flagrt.NewString("\""), flagrt.Call(flagrt.BuiltinFunction("nth"), chars, flagrt.Sub(n, flagrt.NewLong(2)))))
							var if_result_323 flagrt.Value
							if flagrt.IsTruthy(and_tmp) {
								if_result_323 = flagrt.NewBool(flagrt.Eq(flagrt.NewString("\""), flagrt.Call(flagrt.BuiltinFunction("nth"), chars, flagrt.Sub(n, flagrt.NewLong(3)))))
							} else {
								if_result_323 = and_tmp
							}
							let_result_324 = if_result_323
						}
						if_result_325 = let_result_324
					} else {
						if_result_325 = and_tmp
					}
					let_result_326 = if_result_325
				}
				if_result_327 = let_result_326
			} else {
				if_result_327 = flagrt.NewBool(and_tmp)
			}
			let_result_328 = if_result_327
		}
		let_result_329 = let_result_328
	}
	return let_result_329
}

func compiler__triple_quoted_closed_q_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__triple_quoted_closed_q expects exactly 1 arguments")
	}
	return compiler__triple_quoted_closed_q_arity_1(args[0])
}

func compiler__parse_char_token_arity_3(token flagrt.Value, line flagrt.Value, col flagrt.Value) flagrt.Value {
	var let_result_335 flagrt.Value
	{
		var text = flagrt.Call(flagrt.BuiltinFunction("subs"), token, flagrt.NewLong(1))
		var if_result_334 flagrt.Value
		if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(text, flagrt.NewString("space")))) {
			if_result_334 = flagrt.NewMap(flagKw_kind, flagKw_char, flagKw_value, flagrt.NewString(" "), flagKw_line, line, flagKw_col, col)
		} else {
			var if_result_333 flagrt.Value
			if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(text, flagrt.NewString("newline")))) {
				if_result_333 = flagrt.NewMap(flagKw_kind, flagKw_char, flagKw_value, flagrt.NewString("\n"), flagKw_line, line, flagKw_col, col)
			} else {
				var if_result_332 flagrt.Value
				if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(text, flagrt.NewString("tab")))) {
					if_result_332 = flagrt.NewMap(flagKw_kind, flagKw_char, flagKw_value, flagrt.NewString("\t"), flagKw_line, line, flagKw_col, col)
				} else {
					var if_result_331 flagrt.Value
					if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagrt.NewLong(1), flagrt.NewLong(int64(flagrt.Count(flagrt.Seq(text))))))) {
						if_result_331 = flagrt.NewMap(flagKw_kind, flagKw_char, flagKw_value, text, flagKw_line, line, flagKw_col, col)
					} else {
						var if_result_330 flagrt.Value
						if flagrt.IsTruthy(flagrt.NewBool(true)) {
							if_result_330 = flagrt.Call(compiler__parse_error, line, col, flagStr_unsupported_character_li)
						} else {
							if_result_330 = flagrt.NilValue()
						}
						if_result_331 = if_result_330
					}
					if_result_332 = if_result_331
				}
				if_result_333 = if_result_332
			}
			if_result_334 = if_result_333
		}
		let_result_335 = if_result_334
	}
	return let_result_335
}

func compiler__parse_char_token_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 3 {
		panic("compiler__parse_char_token expects exactly 3 arguments")
	}
	return compiler__parse_char_token_arity_3(args[0], args[1], args[2])
}

func compiler__signed_decimal_int_q_arity_1(token flagrt.Value) flagrt.Value {
	var let_result_348 flagrt.Value
	{
		var if_result_336 flagrt.Value
		if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsEmpty(token))) {
			if_result_336 = flagrt.NewBool(false)
		} else {
			if_result_336 = flagrt.NewBool(true)
		}
		var and_tmp = if_result_336
		var if_result_347 flagrt.Value
		if flagrt.IsTruthy(and_tmp) {
			if_result_347 = func() flagrt.Value {
				var cs = flagrt.Seq(token)
				var idx = flagrt.NewLong(0)
				for {
					var if_result_346 flagrt.Value
					if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsEmpty(cs))) {
						if_result_346 = flagrt.NewBool(flagrt.Gt(idx, flagrt.NewLong(0)))
					} else {
						var let_result_345 flagrt.Value
						{
							var ch = flagrt.First(cs)
							var let_result_340 flagrt.Value
							{
								var and_tmp = flagrt.NewBool(flagrt.Eq(idx, flagrt.NewLong(0)))
								var if_result_339 flagrt.Value
								if flagrt.IsTruthy(and_tmp) {
									var let_result_338 flagrt.Value
									{
										var or_tmp = flagrt.NewBool(flagrt.Eq(ch, flagrt.NewString("+")))
										var if_result_337 flagrt.Value
										if flagrt.IsTruthy(or_tmp) {
											if_result_337 = or_tmp
										} else {
											if_result_337 = flagrt.NewBool(flagrt.Eq(ch, flagrt.NewString("-")))
										}
										let_result_338 = if_result_337
									}
									if_result_339 = let_result_338
								} else {
									if_result_339 = and_tmp
								}
								let_result_340 = if_result_339
							}
							var if_result_344 flagrt.Value
							if flagrt.IsTruthy(let_result_340) {
								var if_result_341 flagrt.Value
								if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsEmpty(flagrt.Rest(cs)))) {
									if_result_341 = flagrt.NewBool(false)
								} else {
									if_result_341 = flagrt.NewRecur(flagrt.Rest(cs), flagrt.Add(idx, flagrt.NewLong(1)))
								}
								if_result_344 = if_result_341
							} else {
								var if_result_342 flagrt.Value
								if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsNil(flagrt.Call(flagrt.GoBind_packages_LongParse, ch)))) {
									if_result_342 = flagrt.NewBool(false)
								} else {
									if_result_342 = flagrt.NewBool(true)
								}
								var if_result_343 flagrt.Value
								if flagrt.IsTruthy(if_result_342) {
									if_result_343 = flagrt.NewRecur(flagrt.Rest(cs), flagrt.Add(idx, flagrt.NewLong(1)))
								} else {
									if_result_343 = flagrt.NewBool(false)
								}
								if_result_344 = if_result_343
							}
							let_result_345 = if_result_344
						}
						if_result_346 = let_result_345
					}
					__loopResult := if_result_346
					if __recurValues, __isRecur := flagrt.UnwrapRecur(__loopResult); __isRecur {
						if len(__recurValues) != 2 {
							panic("internal error: recur arity mismatch")
						}
						cs = __recurValues[0]
						idx = __recurValues[1]
						continue
					}
					return __loopResult
				}
			}()
		} else {
			if_result_347 = and_tmp
		}
		let_result_348 = if_result_347
	}
	return let_result_348
}

func compiler__signed_decimal_int_q_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__signed_decimal_int_q expects exactly 1 arguments")
	}
	return compiler__signed_decimal_int_q_arity_1(args[0])
}

func compiler__parse_atom_token_arity_1(tok flagrt.Value) flagrt.Value {
	var let_result_379 flagrt.Value
	{
		var token = flagrt.Call(flagKw_token, tok)
		var line = flagrt.Call(compiler__token_line, tok)
		var col = flagrt.Call(compiler__token_col, tok)
		var let_result_350 flagrt.Value
		{
			var and_tmp = flagrt.Call(compiler__starts_with_char_q, token, flagStr___4)
			var if_result_349 flagrt.Value
			if flagrt.IsTruthy(and_tmp) {
				if_result_349 = flagrt.NewBool(flagrt.Gt(flagrt.NewLong(int64(flagrt.Count(token))), flagrt.NewLong(1)))
			} else {
				if_result_349 = and_tmp
			}
			let_result_350 = if_result_349
		}
		var if_result_378 flagrt.Value
		if flagrt.IsTruthy(let_result_350) {
			if_result_378 = flagrt.NewMap(flagKw_kind, flagKw_keyword, flagKw_name, flagrt.Call(flagrt.BuiltinFunction("subs"), token, flagrt.NewLong(1)), flagKw_line, line, flagKw_col, col)
		} else {
			var let_result_352 flagrt.Value
			{
				var and_tmp = flagrt.Call(compiler__starts_with_char_q, token, flagStr___5)
				var if_result_351 flagrt.Value
				if flagrt.IsTruthy(and_tmp) {
					if_result_351 = flagrt.NewBool(flagrt.Gt(flagrt.NewLong(int64(flagrt.Count(token))), flagrt.NewLong(1)))
				} else {
					if_result_351 = and_tmp
				}
				let_result_352 = if_result_351
			}
			var if_result_377 flagrt.Value
			if flagrt.IsTruthy(let_result_352) {
				if_result_377 = flagrt.Call(compiler__parse_char_token, token, line, col)
			} else {
				var let_result_358 flagrt.Value
				{
					var and_tmp = flagrt.NewBool(flagrt.Eq(flagrt.NewLong(1), flagrt.Call(compiler__slash_count, token)))
					var if_result_357 flagrt.Value
					if flagrt.IsTruthy(and_tmp) {
						var let_result_356 flagrt.Value
						{
							var if_result_353 flagrt.Value
							if flagrt.IsTruthy(flagrt.Call(compiler__starts_with_char_q, token, flagStr___6)) {
								if_result_353 = flagrt.NewBool(false)
							} else {
								if_result_353 = flagrt.NewBool(true)
							}
							var and_tmp = if_result_353
							var if_result_355 flagrt.Value
							if flagrt.IsTruthy(and_tmp) {
								var if_result_354 flagrt.Value
								if flagrt.IsTruthy(flagrt.Call(compiler__ends_with_char_q, token, flagStr___6)) {
									if_result_354 = flagrt.NewBool(false)
								} else {
									if_result_354 = flagrt.NewBool(true)
								}
								if_result_355 = if_result_354
							} else {
								if_result_355 = and_tmp
							}
							let_result_356 = if_result_355
						}
						if_result_357 = let_result_356
					} else {
						if_result_357 = and_tmp
					}
					let_result_358 = if_result_357
				}
				var if_result_376 flagrt.Value
				if flagrt.IsTruthy(let_result_358) {
					var let_result_365 flagrt.Value
					{
						var parts = flagrt.Call(compiler__split_once, token, flagStr___6)
						var numerator = flagrt.Call(flagrt.GoBind_packages_LongParse, flagrt.First(parts))
						var denominator = flagrt.Call(flagrt.GoBind_packages_LongParse, flagrt.First(flagrt.Rest(parts)))
						var let_result_362 flagrt.Value
						{
							var if_result_359 flagrt.Value
							if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsNil(numerator))) {
								if_result_359 = flagrt.NewBool(false)
							} else {
								if_result_359 = flagrt.NewBool(true)
							}
							var and_tmp = if_result_359
							var if_result_361 flagrt.Value
							if flagrt.IsTruthy(and_tmp) {
								var if_result_360 flagrt.Value
								if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsNil(denominator))) {
									if_result_360 = flagrt.NewBool(false)
								} else {
									if_result_360 = flagrt.NewBool(true)
								}
								if_result_361 = if_result_360
							} else {
								if_result_361 = and_tmp
							}
							let_result_362 = if_result_361
						}
						var if_result_364 flagrt.Value
						if flagrt.IsTruthy(let_result_362) {
							var if_result_363 flagrt.Value
							if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagrt.NewLong(0), denominator))) {
								if_result_363 = flagrt.Call(compiler__parse_error, line, col, flagStr_ratio_denominator_cannot)
							} else {
								if_result_363 = flagrt.NewMap(flagKw_kind, flagKw_ratio, flagKw_numerator, numerator, flagKw_denominator, denominator, flagKw_line, line, flagKw_col, col)
							}
							if_result_364 = if_result_363
						} else {
							if_result_364 = flagrt.NewMap(flagKw_kind, flagKw_symbol, flagKw_name, token, flagKw_line, line, flagKw_col, col)
						}
						let_result_365 = if_result_364
					}
					if_result_376 = let_result_365
				} else {
					var let_result_367 flagrt.Value
					{
						var and_tmp = flagrt.Call(compiler__ends_with_char_q, token, flagStr_N)
						var if_result_366 flagrt.Value
						if flagrt.IsTruthy(and_tmp) {
							if_result_366 = flagrt.Call(compiler__signed_decimal_int_q, flagrt.Call(flagrt.BuiltinFunction("subs"), token, flagrt.NewLong(0), flagrt.Sub(flagrt.NewLong(int64(flagrt.Count(token))), flagrt.NewLong(1))))
						} else {
							if_result_366 = and_tmp
						}
						let_result_367 = if_result_366
					}
					var if_result_375 flagrt.Value
					if flagrt.IsTruthy(let_result_367) {
						if_result_375 = flagrt.NewMap(flagKw_kind, flagKw_bigint, flagKw_value, flagrt.Call(flagrt.BuiltinFunction("subs"), token, flagrt.NewLong(0), flagrt.Sub(flagrt.NewLong(int64(flagrt.Count(token))), flagrt.NewLong(1))), flagKw_line, line, flagKw_col, col)
					} else {
						var if_result_374 flagrt.Value
						if flagrt.IsTruthy(flagrt.NewBool(true)) {
							var let_result_373 flagrt.Value
							{
								var as_int = flagrt.Call(flagrt.GoBind_packages_LongParse, token)
								var let_result_370 flagrt.Value
								{
									var if_result_368 flagrt.Value
									if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsNil(as_int))) {
										if_result_368 = flagrt.NewBool(false)
									} else {
										if_result_368 = flagrt.NewBool(true)
									}
									var and_tmp = if_result_368
									var if_result_369 flagrt.Value
									if flagrt.IsTruthy(and_tmp) {
										if_result_369 = flagrt.Call(compiler__signed_decimal_int_q, token)
									} else {
										if_result_369 = and_tmp
									}
									let_result_370 = if_result_369
								}
								var if_result_372 flagrt.Value
								if flagrt.IsTruthy(let_result_370) {
									if_result_372 = flagrt.NewMap(flagKw_kind, flagKw_int, flagKw_value, as_int, flagKw_line, line, flagKw_col, col)
								} else {
									var if_result_371 flagrt.Value
									if flagrt.IsTruthy(flagrt.Call(compiler__contains_float_mark_q, token)) {
										if_result_371 = func() flagrt.Value {
											var __flag_try_result flagrt.Value
											func() {
												defer func() {
													r := recover()
													if r == nil {
														return
													}
													__flag_thrown := flagrt.PanicValue(r)
													if flagrt.CatchMatches("Exception", __flag_thrown) {
														__flag_try_result = func() flagrt.Value {
															var e = __flag_thrown
															_ = e
															return flagrt.NewMap(flagKw_kind, flagKw_symbol, flagKw_name, token, flagKw_line, line, flagKw_col, col)
														}()
														return
													}
													panic(r)
												}()
												__flag_try_result = flagrt.NewMap(flagKw_kind, flagKw_float, flagKw_value, flagrt.Double(flagrt.Call(flagrt.GoBind_runtime_JSONReadStr, token)), flagKw_raw, token, flagKw_line, line, flagKw_col, col)
											}()
											return __flag_try_result
										}()
									} else {
										if_result_371 = flagrt.NewMap(flagKw_kind, flagKw_symbol, flagKw_name, token, flagKw_line, line, flagKw_col, col)
									}
									if_result_372 = if_result_371
								}
								let_result_373 = if_result_372
							}
							if_result_374 = let_result_373
						} else {
							if_result_374 = flagrt.NilValue()
						}
						if_result_375 = if_result_374
					}
					if_result_376 = if_result_375
				}
				if_result_377 = if_result_376
			}
			if_result_378 = if_result_377
		}
		let_result_379 = if_result_378
	}
	return let_result_379
}

func compiler__parse_atom_token_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__parse_atom_token expects exactly 1 arguments")
	}
	return compiler__parse_atom_token_arity_1(args[0])
}

func compiler__comment_list_q_arity_1(elements flagrt.Value) flagrt.Value {
	var let_result_384 flagrt.Value
	{
		var if_result_380 flagrt.Value
		if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsEmpty(elements))) {
			if_result_380 = flagrt.NewBool(false)
		} else {
			if_result_380 = flagrt.NewBool(true)
		}
		var and_tmp = if_result_380
		var if_result_383 flagrt.Value
		if flagrt.IsTruthy(and_tmp) {
			var let_result_382 flagrt.Value
			{
				var and_tmp = flagrt.NewBool(flagrt.Eq(flagKw_symbol, flagrt.Call(flagKw_kind, flagrt.First(elements))))
				var if_result_381 flagrt.Value
				if flagrt.IsTruthy(and_tmp) {
					if_result_381 = flagrt.NewBool(flagrt.Eq(flagrt.NewString("comment"), flagrt.Call(flagKw_name, flagrt.First(elements))))
				} else {
					if_result_381 = and_tmp
				}
				let_result_382 = if_result_381
			}
			if_result_383 = let_result_382
		} else {
			if_result_383 = and_tmp
		}
		let_result_384 = if_result_383
	}
	return let_result_384
}

func compiler__comment_list_q_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__comment_list_q expects exactly 1 arguments")
	}
	return compiler__comment_list_q_arity_1(args[0])
}

func compiler__collection_node_arity_3(kind flagrt.Value, elements flagrt.Value, tok flagrt.Value) flagrt.Value {
	var if_result_385 flagrt.Value
	if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_map))) {
		if_result_385 = flagrt.NewMap(flagKw_kind, flagKw_map, flagKw_entries, elements, flagKw_line, flagrt.Call(compiler__token_line, tok), flagKw_col, flagrt.Call(compiler__token_col, tok))
	} else {
		if_result_385 = flagrt.NewMap(flagKw_kind, kind, flagKw_elements, elements, flagKw_line, flagrt.Call(compiler__token_line, tok), flagKw_col, flagrt.Call(compiler__token_col, tok))
	}
	return if_result_385
}

func compiler__collection_node_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 3 {
		panic("compiler__collection_node expects exactly 3 arguments")
	}
	return compiler__collection_node_arity_3(args[0], args[1], args[2])
}

func compiler__dispatch_error_arity_1(tok flagrt.Value) flagrt.Value {
	var if_result_386 flagrt.Value
	if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagrt.NewString("#"), flagrt.Call(flagKw_token, tok)))) {
		if_result_386 = flagrt.Call(compiler__parse_error, flagrt.Call(compiler__token_line, tok), flagrt.Call(compiler__token_col, tok), flagStr_unexpected_end_after__)
	} else {
		if_result_386 = flagrt.Call(compiler__parse_error, flagrt.Call(compiler__token_line, tok), flagrt.Call(compiler__token_col, tok), flagStr_unsupported_reader_dispa)
	}
	return if_result_386
}

func compiler__dispatch_error_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__dispatch_error expects exactly 1 arguments")
	}
	return compiler__dispatch_error_arity_1(args[0])
}

func compiler__read_string_token_arity_1(p flagrt.Value) flagrt.Value {
	var let_result_391 flagrt.Value
	{
		var tok = flagrt.Call(compiler__next_tok, p)
		var token = flagrt.Call(flagKw_token, tok)
		var line = flagrt.Call(compiler__token_line, tok)
		var col = flagrt.Call(compiler__token_col, tok)
		var if_result_390 flagrt.Value
		if flagrt.IsTruthy(flagrt.Call(compiler__triple_quoted_q, token)) {
			var if_result_387 flagrt.Value
			if flagrt.IsTruthy(flagrt.Call(compiler__triple_quoted_closed_q, token)) {
				if_result_387 = flagrt.NewMap(flagKw_kind, flagKw_string, flagKw_value, flagrt.Call(compiler__strip_edges, token, flagrt.NewLong(3), flagrt.NewLong(3)), flagKw_line, line, flagKw_col, col)
			} else {
				if_result_387 = flagrt.Call(compiler__parse_error, line, col, flagStr_unterminated_string_lite)
			}
			if_result_390 = if_result_387
		} else {
			var let_result_389 flagrt.Value
			{
				var unquoted = flagrt.Call(compiler__unquote_go_string, token)
				var if_result_388 flagrt.Value
				if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagKw_error, unquoted))) {
					if_result_388 = flagrt.Call(compiler__parse_error, line, col, flagStr_unterminated_string_lite)
				} else {
					if_result_388 = flagrt.NewMap(flagKw_kind, flagKw_string, flagKw_value, unquoted, flagKw_line, line, flagKw_col, col)
				}
				let_result_389 = if_result_388
			}
			if_result_390 = let_result_389
		}
		let_result_391 = if_result_390
	}
	return let_result_391
}

func compiler__read_string_token_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__read_string_token expects exactly 1 arguments")
	}
	return compiler__read_string_token_arity_1(args[0])
}

func compiler__read_expr_arity_1(p flagrt.Value) flagrt.Value {
	var let_result_448 flagrt.Value
	{
		var tok = flagrt.Call(compiler__peek, p)
		var if_result_447 flagrt.Value
		if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagKw_eof, tok))) {
			var let_result_392 flagrt.Value
			{
				var pos = flagrt.Call(compiler__eof_line_col, p)
				let_result_392 = flagrt.Call(compiler__parse_error, flagrt.First(pos), flagrt.First(flagrt.Rest(pos)), flagStr_unexpected_end_of_input)
			}
			if_result_447 = let_result_392
		} else {
			var let_result_446 flagrt.Value
			{
				var token = flagrt.Call(flagKw_token, tok)
				var let_result_400 flagrt.Value
				{
					var or_tmp = flagrt.NewBool(flagrt.Eq(token, flagrt.NewString("(")))
					var if_result_399 flagrt.Value
					if flagrt.IsTruthy(or_tmp) {
						if_result_399 = or_tmp
					} else {
						var let_result_398 flagrt.Value
						{
							var or_tmp = flagrt.NewBool(flagrt.Eq(token, flagrt.NewString("[")))
							var if_result_397 flagrt.Value
							if flagrt.IsTruthy(or_tmp) {
								if_result_397 = or_tmp
							} else {
								var let_result_396 flagrt.Value
								{
									var or_tmp = flagrt.NewBool(flagrt.Eq(token, flagrt.NewString("|")))
									var if_result_395 flagrt.Value
									if flagrt.IsTruthy(or_tmp) {
										if_result_395 = or_tmp
									} else {
										var let_result_394 flagrt.Value
										{
											var or_tmp = flagrt.NewBool(flagrt.Eq(token, flagrt.NewString("{")))
											var if_result_393 flagrt.Value
											if flagrt.IsTruthy(or_tmp) {
												if_result_393 = or_tmp
											} else {
												if_result_393 = flagrt.NewBool(flagrt.Eq(token, flagrt.NewString("#{")))
											}
											let_result_394 = if_result_393
										}
										if_result_395 = let_result_394
									}
									let_result_396 = if_result_395
								}
								if_result_397 = let_result_396
							}
							let_result_398 = if_result_397
						}
						if_result_399 = let_result_398
					}
					let_result_400 = if_result_399
				}
				var if_result_445 flagrt.Value
				if flagrt.IsTruthy(let_result_400) {
					var let_result_422 flagrt.Value
					{
						var if_result_405 flagrt.Value
						if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(token, flagrt.NewString("(")))) {
							if_result_405 = flagKw_list
						} else {
							var if_result_404 flagrt.Value
							if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(token, flagrt.NewString("[")))) {
								if_result_404 = flagKw_vector
							} else {
								var if_result_403 flagrt.Value
								if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(token, flagrt.NewString("|")))) {
									if_result_403 = flagKw_pipe_vector
								} else {
									var if_result_402 flagrt.Value
									if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(token, flagrt.NewString("{")))) {
										if_result_402 = flagKw_map
									} else {
										var if_result_401 flagrt.Value
										if flagrt.IsTruthy(flagrt.NewBool(true)) {
											if_result_401 = flagKw_set
										} else {
											if_result_401 = flagrt.NilValue()
										}
										if_result_402 = if_result_401
									}
									if_result_403 = if_result_402
								}
								if_result_404 = if_result_403
							}
							if_result_405 = if_result_404
						}
						var kind = if_result_405
						var if_result_409 flagrt.Value
						if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(token, flagrt.NewString("(")))) {
							if_result_409 = flagStr___10
						} else {
							var if_result_408 flagrt.Value
							if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(token, flagrt.NewString("[")))) {
								if_result_408 = flagStr___9
							} else {
								var if_result_407 flagrt.Value
								if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(token, flagrt.NewString("|")))) {
									if_result_407 = flagStr___8
								} else {
									var if_result_406 flagrt.Value
									if flagrt.IsTruthy(flagrt.NewBool(true)) {
										if_result_406 = flagStr___7
									} else {
										if_result_406 = flagrt.NilValue()
									}
									if_result_407 = if_result_406
								}
								if_result_408 = if_result_407
							}
							if_result_409 = if_result_408
						}
						var closer = if_result_409
						var if_result_413 flagrt.Value
						if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(token, flagrt.NewString("(")))) {
							if_result_413 = flagStr_missing_closing_____3
						} else {
							var if_result_412 flagrt.Value
							if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(token, flagrt.NewString("[")))) {
								if_result_412 = flagStr_missing_closing_____2
							} else {
								var if_result_411 flagrt.Value
								if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(token, flagrt.NewString("|")))) {
									if_result_411 = flagStr_missing_closing_____1
								} else {
									var if_result_410 flagrt.Value
									if flagrt.IsTruthy(flagrt.NewBool(true)) {
										if_result_410 = flagStr_missing_closing____
									} else {
										if_result_410 = flagrt.NilValue()
									}
									if_result_411 = if_result_410
								}
								if_result_412 = if_result_411
							}
							if_result_413 = if_result_412
						}
						var missing = if_result_413
						var open = flagrt.Call(compiler__next_tok, p)
						let_result_422 = func() flagrt.Value {
							var elements = flagVec
							for {
								var let_result_421 flagrt.Value
								{
									var cur = flagrt.Call(compiler__peek, p)
									var if_result_420 flagrt.Value
									if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagKw_eof, cur))) {
										if_result_420 = flagrt.Call(compiler__parse_error, flagrt.Call(compiler__token_line, open), flagrt.Call(compiler__token_col, open), missing)
									} else {
										var if_result_419 flagrt.Value
										if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(closer, flagrt.Call(flagKw_token, cur)))) {
											_ = flagrt.Call(compiler__next_tok, p)
											var let_result_415 flagrt.Value
											{
												var and_tmp = flagrt.NewBool(flagrt.Eq(kind, flagKw_list))
												var if_result_414 flagrt.Value
												if flagrt.IsTruthy(and_tmp) {
													if_result_414 = flagrt.Call(compiler__comment_list_q, elements)
												} else {
													if_result_414 = and_tmp
												}
												let_result_415 = if_result_414
											}
											var if_result_416 flagrt.Value
											if flagrt.IsTruthy(let_result_415) {
												if_result_416 = flagrt.NilValue()
											} else {
												if_result_416 = flagrt.Call(compiler__collection_node, kind, elements, open)
											}
											if_result_419 = if_result_416
										} else {
											var let_result_418 flagrt.Value
											{
												var item = compiler__read_expr_arity_1(p)
												var if_result_417 flagrt.Value
												if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsNil(item))) {
													if_result_417 = flagrt.NewRecur(elements)
												} else {
													if_result_417 = flagrt.NewRecur(flagrt.Conj(elements, item))
												}
												let_result_418 = if_result_417
											}
											if_result_419 = let_result_418
										}
										if_result_420 = if_result_419
									}
									let_result_421 = if_result_420
								}
								__loopResult := let_result_421
								if __recurValues, __isRecur := flagrt.UnwrapRecur(__loopResult); __isRecur {
									if len(__recurValues) != 1 {
										panic("internal error: recur arity mismatch")
									}
									elements = __recurValues[0]
									continue
								}
								return __loopResult
							}
						}()
					}
					if_result_445 = let_result_422
				} else {
					var if_result_444 flagrt.Value
					if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(token, flagrt.NewString("#(")))) {
						var let_result_428 flagrt.Value
						{
							var open = flagrt.Call(compiler__next_tok, p)
							let_result_428 = func() flagrt.Value {
								var elements = flagVec
								for {
									var let_result_427 flagrt.Value
									{
										var cur = flagrt.Call(compiler__peek, p)
										var if_result_426 flagrt.Value
										if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagKw_eof, cur))) {
											if_result_426 = flagrt.Call(compiler__parse_error, flagrt.Call(compiler__token_line, open), flagrt.Call(compiler__token_col, open), flagStr_missing_closing_____3)
										} else {
											var if_result_425 flagrt.Value
											if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagrt.NewString(")"), flagrt.Call(flagKw_token, cur)))) {
												_ = flagrt.Call(compiler__next_tok, p)
												if_result_425 = flagrt.NewMap(flagKw_kind, flagKw_hash_fn, flagKw_body, flagrt.NewMap(flagKw_kind, flagKw_list, flagKw_elements, elements, flagKw_line, flagrt.Call(compiler__token_line, open), flagKw_col, flagrt.Call(compiler__token_col, open)), flagKw_line, flagrt.Call(compiler__token_line, open), flagKw_col, flagrt.Call(compiler__token_col, open))
											} else {
												var let_result_424 flagrt.Value
												{
													var item = compiler__read_expr_arity_1(p)
													var if_result_423 flagrt.Value
													if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsNil(item))) {
														if_result_423 = flagrt.NewRecur(elements)
													} else {
														if_result_423 = flagrt.NewRecur(flagrt.Conj(elements, item))
													}
													let_result_424 = if_result_423
												}
												if_result_425 = let_result_424
											}
											if_result_426 = if_result_425
										}
										let_result_427 = if_result_426
									}
									__loopResult := let_result_427
									if __recurValues, __isRecur := flagrt.UnwrapRecur(__loopResult); __isRecur {
										if len(__recurValues) != 1 {
											panic("internal error: recur arity mismatch")
										}
										elements = __recurValues[0]
										continue
									}
									return __loopResult
								}
							}()
						}
						if_result_444 = let_result_428
					} else {
						var if_result_443 flagrt.Value
						if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(token, flagrt.NewString("^")))) {
							var let_result_429 flagrt.Value
							{
								var meta_tok = flagrt.Call(compiler__next_tok, p)
								var meta = compiler__read_expr_arity_1(p)
								var target = compiler__read_expr_arity_1(p)
								let_result_429 = flagrt.NewMap(flagKw_kind, flagKw_meta, flagKw_meta, meta, flagKw_target, target, flagKw_line, flagrt.Call(compiler__token_line, meta_tok), flagKw_col, flagrt.Call(compiler__token_col, meta_tok))
							}
							if_result_443 = let_result_429
						} else {
							var if_result_442 flagrt.Value
							if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(token, flagrt.NewString("'")))) {
								var let_result_433 flagrt.Value
								{
									var quote_tok = flagrt.Call(compiler__next_tok, p)
									var quoted = compiler__read_expr_arity_1(p)
									var if_result_432 flagrt.Value
									if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagKw_symbol, flagrt.Call(flagKw_kind, quoted)))) {
										if_result_432 = flagrt.NewMap(flagKw_kind, flagKw_quoted_symbol, flagKw_name, flagrt.Call(flagKw_name, quoted), flagKw_line, flagrt.Call(compiler__token_line, quote_tok), flagKw_col, flagrt.Call(compiler__token_col, quote_tok))
									} else {
										var if_result_431 flagrt.Value
										if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagKw_list, flagrt.Call(flagKw_kind, quoted)))) {
											if_result_431 = flagrt.NewMap(flagKw_kind, flagKw_quoted_list, flagKw_elements, flagrt.Call(flagKw_elements, quoted), flagKw_line, flagrt.Call(compiler__token_line, quote_tok), flagKw_col, flagrt.Call(compiler__token_col, quote_tok))
										} else {
											var if_result_430 flagrt.Value
											if flagrt.IsTruthy(flagrt.NewBool(true)) {
												if_result_430 = flagrt.Call(compiler__parse_error, flagrt.Call(compiler__token_line, quote_tok), flagrt.Call(compiler__token_col, quote_tok), flagStr_quote_currently_supports)
											} else {
												if_result_430 = flagrt.NilValue()
											}
											if_result_431 = if_result_430
										}
										if_result_432 = if_result_431
									}
									let_result_433 = if_result_432
								}
								if_result_442 = let_result_433
							} else {
								var if_result_441 flagrt.Value
								if flagrt.IsTruthy(flagrt.Call(compiler__string_token_q, tok)) {
									if_result_441 = flagrt.Call(compiler__read_string_token, p)
								} else {
									var let_result_437 flagrt.Value
									{
										var or_tmp = flagrt.NewBool(flagrt.Eq(token, flagrt.NewString(")")))
										var if_result_436 flagrt.Value
										if flagrt.IsTruthy(or_tmp) {
											if_result_436 = or_tmp
										} else {
											var let_result_435 flagrt.Value
											{
												var or_tmp = flagrt.NewBool(flagrt.Eq(token, flagrt.NewString("]")))
												var if_result_434 flagrt.Value
												if flagrt.IsTruthy(or_tmp) {
													if_result_434 = or_tmp
												} else {
													if_result_434 = flagrt.NewBool(flagrt.Eq(token, flagrt.NewString("}")))
												}
												let_result_435 = if_result_434
											}
											if_result_436 = let_result_435
										}
										let_result_437 = if_result_436
									}
									var if_result_440 flagrt.Value
									if flagrt.IsTruthy(let_result_437) {
										if_result_440 = flagrt.Call(compiler__parse_error, flagrt.Call(compiler__token_line, tok), flagrt.Call(compiler__token_col, tok), flagStr_expected_expression)
									} else {
										var if_result_439 flagrt.Value
										if flagrt.IsTruthy(flagrt.Call(compiler__starts_with_char_q, token, flagStr___11)) {
											_ = flagrt.Call(compiler__next_tok, p)
											if_result_439 = flagrt.Call(compiler__dispatch_error, tok)
										} else {
											var if_result_438 flagrt.Value
											if flagrt.IsTruthy(flagrt.NewBool(true)) {
												_ = flagrt.Call(compiler__next_tok, p)
												if_result_438 = flagrt.Call(compiler__parse_atom_token, tok)
											} else {
												if_result_438 = flagrt.NilValue()
											}
											if_result_439 = if_result_438
										}
										if_result_440 = if_result_439
									}
									if_result_441 = if_result_440
								}
								if_result_442 = if_result_441
							}
							if_result_443 = if_result_442
						}
						if_result_444 = if_result_443
					}
					if_result_445 = if_result_444
				}
				let_result_446 = if_result_445
			}
			if_result_447 = let_result_446
		}
		let_result_448 = if_result_447
	}
	return let_result_448
}

func compiler__read_expr_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__read_expr expects exactly 1 arguments")
	}
	return compiler__read_expr_arity_1(args[0])
}

func compiler__ast_node__gtcanonical_arity_1(expr flagrt.Value) flagrt.Value {
	var let_result_470 flagrt.Value
	{
		var kind = flagrt.Call(flagKw_kind, expr)
		var line = flagrt.Call(flagKw_line, expr)
		var col = flagrt.Call(flagKw_col, expr)
		var join = flagrt.NewFunction(func(args ...flagrt.Value) flagrt.Value {
			if len(args) != 1 {
				panic("fn expects exactly 1 arguments")
			}
			nodes := args[0]
			return func() flagrt.Value {
				var remaining = nodes
				var out = flagStr_
				var first_q = flagrt.NewBool(true)
				for {
					var if_result_451 flagrt.Value
					if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsEmpty(remaining))) {
						if_result_451 = out
					} else {
						var let_result_450 flagrt.Value
						{
							var piece = compiler__ast_node__gtcanonical_arity_1(flagrt.First(remaining))
							var if_result_449 flagrt.Value
							if flagrt.IsTruthy(first_q) {
								if_result_449 = flagrt.NewRecur(flagrt.Rest(remaining), piece, flagrt.NewBool(false))
							} else {
								if_result_449 = flagrt.NewRecur(flagrt.Rest(remaining), flagrt.NewString(flagrt.Str(out, " ", piece)), flagrt.NewBool(false))
							}
							let_result_450 = if_result_449
						}
						if_result_451 = let_result_450
					}
					__loopResult := if_result_451
					if __recurValues, __isRecur := flagrt.UnwrapRecur(__loopResult); __isRecur {
						if len(__recurValues) != 3 {
							panic("internal error: recur arity mismatch")
						}
						remaining = __recurValues[0]
						out = __recurValues[1]
						first_q = __recurValues[2]
						continue
					}
					return __loopResult
				}
			}()
		})
		var if_result_469 flagrt.Value
		if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_list))) {
			if_result_469 = flagrt.NewString(flagrt.Str("(:list :line ", line, " :col ", col, " :elements [", flagrt.Call(join, flagrt.Call(flagKw_elements, expr)), "])"))
		} else {
			var if_result_468 flagrt.Value
			if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_vector))) {
				if_result_468 = flagrt.NewString(flagrt.Str("(:vector :line ", line, " :col ", col, " :elements [", flagrt.Call(join, flagrt.Call(flagKw_elements, expr)), "])"))
			} else {
				var if_result_467 flagrt.Value
				if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_pipe_vector))) {
					if_result_467 = flagrt.NewString(flagrt.Str("(:pipe-vector :line ", line, " :col ", col, " :elements [", flagrt.Call(join, flagrt.Call(flagKw_elements, expr)), "])"))
				} else {
					var if_result_466 flagrt.Value
					if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_map))) {
						if_result_466 = flagrt.NewString(flagrt.Str("(:map :line ", line, " :col ", col, " :entries [", flagrt.Call(join, flagrt.Call(flagKw_entries, expr)), "])"))
					} else {
						var if_result_465 flagrt.Value
						if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_set))) {
							if_result_465 = flagrt.NewString(flagrt.Str("(:set :line ", line, " :col ", col, " :elements [", flagrt.Call(join, flagrt.Call(flagKw_elements, expr)), "])"))
						} else {
							var if_result_464 flagrt.Value
							if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_hash_fn))) {
								if_result_464 = flagrt.NewString(flagrt.Str("(:hash-fn :line ", line, " :col ", col, " :body ", compiler__ast_node__gtcanonical_arity_1(flagrt.Call(flagKw_body, expr)), ")"))
							} else {
								var if_result_463 flagrt.Value
								if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_meta))) {
									if_result_463 = flagrt.NewString(flagrt.Str("(:meta :line ", line, " :col ", col, " :meta ", compiler__ast_node__gtcanonical_arity_1(flagrt.Call(flagKw_meta, expr)), " :target ", compiler__ast_node__gtcanonical_arity_1(flagrt.Call(flagKw_target, expr)), ")"))
								} else {
									var if_result_462 flagrt.Value
									if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_symbol))) {
										if_result_462 = flagrt.NewString(flagrt.Str("(:symbol :name ", flagrt.Call(flagKw_name, expr), " :line ", line, " :col ", col, ")"))
									} else {
										var if_result_461 flagrt.Value
										if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_keyword))) {
											if_result_461 = flagrt.NewString(flagrt.Str("(:keyword :name ", flagrt.Call(flagKw_name, expr), " :line ", line, " :col ", col, ")"))
										} else {
											var if_result_460 flagrt.Value
											if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_quoted_symbol))) {
												if_result_460 = flagrt.NewString(flagrt.Str("(:quoted-symbol :name ", flagrt.Call(flagKw_name, expr), " :line ", line, " :col ", col, ")"))
											} else {
												var if_result_459 flagrt.Value
												if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_quoted_list))) {
													if_result_459 = flagrt.NewString(flagrt.Str("(:quoted-list :line ", line, " :col ", col, " :elements [", flagrt.Call(join, flagrt.Call(flagKw_elements, expr)), "])"))
												} else {
													var if_result_458 flagrt.Value
													if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_string))) {
														if_result_458 = flagrt.NewString(flagrt.Str("(:string :value ", flagrt.NewString(flagrt.Format("%q", flagrt.Call(flagKw_value, expr))), " :line ", line, " :col ", col, ")"))
													} else {
														var if_result_457 flagrt.Value
														if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_char))) {
															if_result_457 = flagrt.NewString(flagrt.Str("(:char :value ", flagrt.NewString(flagrt.Format("%q", flagrt.Call(flagKw_value, expr))), " :line ", line, " :col ", col, ")"))
														} else {
															var if_result_456 flagrt.Value
															if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_int))) {
																if_result_456 = flagrt.NewString(flagrt.Str("(:int :value ", flagrt.Call(flagKw_value, expr), " :line ", line, " :col ", col, ")"))
															} else {
																var if_result_455 flagrt.Value
																if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_bigint))) {
																	if_result_455 = flagrt.NewString(flagrt.Str("(:bigint :value ", flagrt.Call(flagKw_value, expr), " :line ", line, " :col ", col, ")"))
																} else {
																	var if_result_454 flagrt.Value
																	if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_float))) {
																		if_result_454 = flagrt.NewString(flagrt.Str("(:float :value ", flagrt.Call(flagKw_value, expr), " :raw ", flagrt.Call(flagKw_raw, expr), " :line ", line, " :col ", col, ")"))
																	} else {
																		var if_result_453 flagrt.Value
																		if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_ratio))) {
																			if_result_453 = flagrt.NewString(flagrt.Str("(:ratio :numerator ", flagrt.Call(flagKw_numerator, expr), " :denominator ", flagrt.Call(flagKw_denominator, expr), " :line ", line, " :col ", col, ")"))
																		} else {
																			var if_result_452 flagrt.Value
																			if flagrt.IsTruthy(flagrt.NewBool(true)) {
																				if_result_452 = func() flagrt.Value {
																					flagrt.Throw(flagrt.NewString(flagrt.Str("unsupported AST node ", kind)))
																					return flagrt.NilValue()
																				}()
																			} else {
																				if_result_452 = flagrt.NilValue()
																			}
																			if_result_453 = if_result_452
																		}
																		if_result_454 = if_result_453
																	}
																	if_result_455 = if_result_454
																}
																if_result_456 = if_result_455
															}
															if_result_457 = if_result_456
														}
														if_result_458 = if_result_457
													}
													if_result_459 = if_result_458
												}
												if_result_460 = if_result_459
											}
											if_result_461 = if_result_460
										}
										if_result_462 = if_result_461
									}
									if_result_463 = if_result_462
								}
								if_result_464 = if_result_463
							}
							if_result_465 = if_result_464
						}
						if_result_466 = if_result_465
					}
					if_result_467 = if_result_466
				}
				if_result_468 = if_result_467
			}
			if_result_469 = if_result_468
		}
		let_result_470 = if_result_469
	}
	return let_result_470
}

func compiler__ast_node__gtcanonical_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__ast_node__gtcanonical expects exactly 1 arguments")
	}
	return compiler__ast_node__gtcanonical_arity_1(args[0])
}

func compiler__asts__gtcanonical_arity_1(exprs flagrt.Value) flagrt.Value {
	return func() flagrt.Value {
		var remaining = exprs
		var out = flagStr_
		var first_q = flagrt.NewBool(true)
		for {
			var if_result_473 flagrt.Value
			if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsEmpty(remaining))) {
				if_result_473 = out
			} else {
				var let_result_472 flagrt.Value
				{
					var piece = flagrt.Call(compiler__ast_node__gtcanonical, flagrt.First(remaining))
					var if_result_471 flagrt.Value
					if flagrt.IsTruthy(first_q) {
						if_result_471 = flagrt.NewRecur(flagrt.Rest(remaining), piece, flagrt.NewBool(false))
					} else {
						if_result_471 = flagrt.NewRecur(flagrt.Rest(remaining), flagrt.NewString(flagrt.Str(out, "\n", piece)), flagrt.NewBool(false))
					}
					let_result_472 = if_result_471
				}
				if_result_473 = let_result_472
			}
			__loopResult := if_result_473
			if __recurValues, __isRecur := flagrt.UnwrapRecur(__loopResult); __isRecur {
				if len(__recurValues) != 3 {
					panic("internal error: recur arity mismatch")
				}
				remaining = __recurValues[0]
				out = __recurValues[1]
				first_q = __recurValues[2]
				continue
			}
			return __loopResult
		}
	}()
}

func compiler__asts__gtcanonical_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__asts__gtcanonical expects exactly 1 arguments")
	}
	return compiler__asts__gtcanonical_arity_1(args[0])
}

func compiler__drain_remaining_arity_1(in flagrt.Value) flagrt.Value {
	return func() flagrt.Value {
		var _more = flagrt.NewBool(true)
		_ = _more
		for {
			var if_result_474 flagrt.Value
			if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsNil(flagrt.Call(async__channel_receive, in)))) {
				if_result_474 = flagrt.NilValue()
			} else {
				if_result_474 = flagrt.NewRecur(flagrt.NewBool(true))
			}
			__loopResult := if_result_474
			if __recurValues, __isRecur := flagrt.UnwrapRecur(__loopResult); __isRecur {
				if len(__recurValues) != 1 {
					panic("internal error: recur arity mismatch")
				}
				_more = __recurValues[0]
				continue
			}
			return __loopResult
		}
	}()
}

func compiler__drain_remaining_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__drain_remaining expects exactly 1 arguments")
	}
	return compiler__drain_remaining_arity_1(args[0])
}

// Read SourceToken maps from in, parse top-level expressions, and send AST nodes to a new output channel.
func compiler__build_ast_from_tokens_arity_1(in flagrt.Value) flagrt.Value {
	var let_result_479 flagrt.Value
	{
		var out = flagrt.Call(async__make_channel, flagrt.NewLong(64))
		var p = flagrt.Call(compiler__make_parser, in)
		_ = flagrt.Call(async__go_run, flagrt.NewFunction(func(args ...flagrt.Value) flagrt.Value {
			if len(args) != 0 {
				panic("fn expects exactly 0 arguments")
			}
			return func() flagrt.Value {
				var __flag_try_result flagrt.Value
				defer func() {
					_ = flagrt.Call(async__channel_close, out)
					_ = flagrt.Call(compiler__drain_remaining, in)
				}()
				func() {
					defer func() {
						r := recover()
						if r == nil {
							return
						}
						__flag_thrown := flagrt.PanicValue(r)
						if flagrt.CatchMatches("Exception", __flag_thrown) {
							__flag_try_result = func() flagrt.Value {
								var e = __flag_thrown
								_ = e
								return flagrt.Call(async__channel_send, out, flagrt.NewMap(flagKw_kind, flagKw_error, flagKw_message, flagrt.ExMessage(e)))
							}()
							return
						}
						panic(r)
					}()
					__flag_try_result = func() flagrt.Value {
						var _more = flagrt.NewBool(true)
						_ = _more
						for {
							var let_result_478 flagrt.Value
							{
								var tok = flagrt.Call(compiler__peek, p)
								var if_result_477 flagrt.Value
								if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagKw_eof, tok))) {
									if_result_477 = flagrt.NilValue()
								} else {
									var let_result_476 flagrt.Value
									{
										var form = flagrt.Call(compiler__read_expr, p)
										var if_result_475 flagrt.Value
										if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsNil(form))) {
											if_result_475 = flagrt.NewRecur(flagrt.NewBool(true))
										} else {
											_ = flagrt.Call(async__channel_send, out, form)
											if_result_475 = flagrt.NewRecur(flagrt.NewBool(true))
										}
										let_result_476 = if_result_475
									}
									if_result_477 = let_result_476
								}
								let_result_478 = if_result_477
							}
							__loopResult := let_result_478
							if __recurValues, __isRecur := flagrt.UnwrapRecur(__loopResult); __isRecur {
								if len(__recurValues) != 1 {
									panic("internal error: recur arity mismatch")
								}
								_more = __recurValues[0]
								continue
							}
							return __loopResult
						}
					}()
				}()
				return __flag_try_result
			}()
		}))
		let_result_479 = out
	}
	return let_result_479
}

func compiler__build_ast_from_tokens_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__build_ast_from_tokens expects exactly 1 arguments")
	}
	return compiler__build_ast_from_tokens_arity_1(args[0])
}

func compiler__symbol_node_q_arity_1(expr flagrt.Value) flagrt.Value {
	return flagrt.NewBool(flagrt.Eq(flagKw_symbol, flagrt.Call(flagKw_kind, expr)))
}

func compiler__symbol_node_q_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__symbol_node_q expects exactly 1 arguments")
	}
	return compiler__symbol_node_q_arity_1(args[0])
}

func compiler__list_node_q_arity_1(expr flagrt.Value) flagrt.Value {
	return flagrt.NewBool(flagrt.Eq(flagKw_list, flagrt.Call(flagKw_kind, expr)))
}

func compiler__list_node_q_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__list_node_q expects exactly 1 arguments")
	}
	return compiler__list_node_q_arity_1(args[0])
}

func compiler__vector_node_q_arity_1(expr flagrt.Value) flagrt.Value {
	return flagrt.NewBool(flagrt.Eq(flagKw_vector, flagrt.Call(flagKw_kind, expr)))
}

func compiler__vector_node_q_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__vector_node_q expects exactly 1 arguments")
	}
	return compiler__vector_node_q_arity_1(args[0])
}

func compiler__pipe_vector_node_q_arity_1(expr flagrt.Value) flagrt.Value {
	return flagrt.NewBool(flagrt.Eq(flagKw_pipe_vector, flagrt.Call(flagKw_kind, expr)))
}

func compiler__pipe_vector_node_q_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__pipe_vector_node_q expects exactly 1 arguments")
	}
	return compiler__pipe_vector_node_q_arity_1(args[0])
}

func compiler__map_node_q_arity_1(expr flagrt.Value) flagrt.Value {
	return flagrt.NewBool(flagrt.Eq(flagKw_map, flagrt.Call(flagKw_kind, expr)))
}

func compiler__map_node_q_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__map_node_q expects exactly 1 arguments")
	}
	return compiler__map_node_q_arity_1(args[0])
}

func compiler__set_node_q_arity_1(expr flagrt.Value) flagrt.Value {
	return flagrt.NewBool(flagrt.Eq(flagKw_set, flagrt.Call(flagKw_kind, expr)))
}

func compiler__set_node_q_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__set_node_q expects exactly 1 arguments")
	}
	return compiler__set_node_q_arity_1(args[0])
}

func compiler__hash_fn_node_q_arity_1(expr flagrt.Value) flagrt.Value {
	return flagrt.NewBool(flagrt.Eq(flagKw_hash_fn, flagrt.Call(flagKw_kind, expr)))
}

func compiler__hash_fn_node_q_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__hash_fn_node_q expects exactly 1 arguments")
	}
	return compiler__hash_fn_node_q_arity_1(args[0])
}

func compiler__meta_node_q_arity_1(expr flagrt.Value) flagrt.Value {
	return flagrt.NewBool(flagrt.Eq(flagKw_meta, flagrt.Call(flagKw_kind, expr)))
}

func compiler__meta_node_q_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__meta_node_q expects exactly 1 arguments")
	}
	return compiler__meta_node_q_arity_1(args[0])
}

func compiler__string_node_q_arity_1(expr flagrt.Value) flagrt.Value {
	return flagrt.NewBool(flagrt.Eq(flagKw_string, flagrt.Call(flagKw_kind, expr)))
}

func compiler__string_node_q_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__string_node_q expects exactly 1 arguments")
	}
	return compiler__string_node_q_arity_1(args[0])
}

func compiler__unwrap_meta_arity_1(expr flagrt.Value) flagrt.Value {
	var if_result_480 flagrt.Value
	if flagrt.IsTruthy(flagrt.Call(compiler__meta_node_q, expr)) {
		if_result_480 = flagrt.Call(flagKw_target, expr)
	} else {
		if_result_480 = expr
	}
	return if_result_480
}

func compiler__unwrap_meta_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__unwrap_meta expects exactly 1 arguments")
	}
	return compiler__unwrap_meta_arity_1(args[0])
}

func compiler__node_children_arity_1(expr flagrt.Value) flagrt.Value {
	var if_result_481 flagrt.Value
	if flagrt.IsTruthy(flagrt.Call(compiler__map_node_q, expr)) {
		if_result_481 = flagrt.Call(flagKw_entries, expr)
	} else {
		if_result_481 = flagrt.Call(flagKw_elements, expr)
	}
	return if_result_481
}

func compiler__node_children_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__node_children expects exactly 1 arguments")
	}
	return compiler__node_children_arity_1(args[0])
}

func compiler__with_children_arity_2(expr flagrt.Value, children flagrt.Value) flagrt.Value {
	var if_result_482 flagrt.Value
	if flagrt.IsTruthy(flagrt.Call(compiler__map_node_q, expr)) {
		if_result_482 = flagrt.Assoc(expr, flagKw_entries, children)
	} else {
		if_result_482 = flagrt.Assoc(expr, flagKw_elements, children)
	}
	return if_result_482
}

func compiler__with_children_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 2 {
		panic("compiler__with_children expects exactly 2 arguments")
	}
	return compiler__with_children_arity_2(args[0], args[1])
}

func compiler__splice_node_q_arity_1(expr flagrt.Value) flagrt.Value {
	var let_result_486 flagrt.Value
	{
		var or_tmp = flagrt.Call(compiler__list_node_q, expr)
		var if_result_485 flagrt.Value
		if flagrt.IsTruthy(or_tmp) {
			if_result_485 = or_tmp
		} else {
			var let_result_484 flagrt.Value
			{
				var or_tmp = flagrt.Call(compiler__vector_node_q, expr)
				var if_result_483 flagrt.Value
				if flagrt.IsTruthy(or_tmp) {
					if_result_483 = or_tmp
				} else {
					if_result_483 = flagrt.Call(compiler__pipe_vector_node_q, expr)
				}
				let_result_484 = if_result_483
			}
			if_result_485 = let_result_484
		}
		let_result_486 = if_result_485
	}
	return let_result_486
}

func compiler__splice_node_q_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__splice_node_q expects exactly 1 arguments")
	}
	return compiler__splice_node_q_arity_1(args[0])
}

func compiler__walk_node_q_arity_1(expr flagrt.Value) flagrt.Value {
	var let_result_490 flagrt.Value
	{
		var or_tmp = flagrt.Call(compiler__splice_node_q, expr)
		var if_result_489 flagrt.Value
		if flagrt.IsTruthy(or_tmp) {
			if_result_489 = or_tmp
		} else {
			var let_result_488 flagrt.Value
			{
				var or_tmp = flagrt.Call(compiler__map_node_q, expr)
				var if_result_487 flagrt.Value
				if flagrt.IsTruthy(or_tmp) {
					if_result_487 = or_tmp
				} else {
					if_result_487 = flagrt.Call(compiler__set_node_q, expr)
				}
				let_result_488 = if_result_487
			}
			if_result_489 = let_result_488
		}
		let_result_490 = if_result_489
	}
	return let_result_490
}

func compiler__walk_node_q_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__walk_node_q expects exactly 1 arguments")
	}
	return compiler__walk_node_q_arity_1(args[0])
}

func compiler__make_symbol_arity_1(name flagrt.Value) flagrt.Value {
	return flagrt.NewMap(flagKw_kind, flagKw_symbol, flagKw_name, name)
}

func compiler__make_symbol_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__make_symbol expects exactly 1 arguments")
	}
	return compiler__make_symbol_arity_1(args[0])
}

func compiler__make_list_arity_1(elements flagrt.Value) flagrt.Value {
	return flagrt.NewMap(flagKw_kind, flagKw_list, flagKw_elements, elements)
}

func compiler__make_list_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__make_list expects exactly 1 arguments")
	}
	return compiler__make_list_arity_1(args[0])
}

func compiler__wrap_literal_arity_1(expr flagrt.Value) flagrt.Value {
	return flagrt.NewMap(flagKw_kind, flagKw_macro_literal, flagKw_inner, expr)
}

func compiler__wrap_literal_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__wrap_literal expects exactly 1 arguments")
	}
	return compiler__wrap_literal_arity_1(args[0])
}

func compiler__unwrap_literal_arity_1(expr flagrt.Value) flagrt.Value {
	var if_result_491 flagrt.Value
	if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagKw_macro_literal, flagrt.Call(flagKw_kind, expr)))) {
		if_result_491 = compiler__unwrap_literal_arity_1(flagrt.Call(flagKw_inner, expr))
	} else {
		if_result_491 = expr
	}
	return if_result_491
}

func compiler__unwrap_literal_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__unwrap_literal expects exactly 1 arguments")
	}
	return compiler__unwrap_literal_arity_1(args[0])
}

func compiler__unwrap_literal_tree_arity_1(expr flagrt.Value) flagrt.Value {
	var let_result_496 flagrt.Value
	{
		var expr = flagrt.Call(compiler__unwrap_literal, expr)
		var if_result_495 flagrt.Value
		if flagrt.IsTruthy(flagrt.Call(compiler__walk_node_q, expr)) {
			if_result_495 = flagrt.Call(compiler__with_children, expr, func() flagrt.Value {
				var remaining = flagrt.Call(compiler__node_children, expr)
				var acc = flagVec
				for {
					var if_result_492 flagrt.Value
					if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsEmpty(remaining))) {
						if_result_492 = acc
					} else {
						if_result_492 = flagrt.NewRecur(flagrt.Rest(remaining), flagrt.Conj(acc, compiler__unwrap_literal_tree_arity_1(flagrt.First(remaining))))
					}
					__loopResult := if_result_492
					if __recurValues, __isRecur := flagrt.UnwrapRecur(__loopResult); __isRecur {
						if len(__recurValues) != 2 {
							panic("internal error: recur arity mismatch")
						}
						remaining = __recurValues[0]
						acc = __recurValues[1]
						continue
					}
					return __loopResult
				}
			}())
		} else {
			var if_result_494 flagrt.Value
			if flagrt.IsTruthy(flagrt.Call(compiler__hash_fn_node_q, expr)) {
				if_result_494 = flagrt.Assoc(expr, flagKw_body, compiler__unwrap_literal_tree_arity_1(flagrt.Call(flagKw_body, expr)))
			} else {
				var if_result_493 flagrt.Value
				if flagrt.IsTruthy(flagrt.Call(compiler__meta_node_q, expr)) {
					if_result_493 = flagrt.Assoc(expr, flagKw_meta, compiler__unwrap_literal_tree_arity_1(flagrt.Call(flagKw_meta, expr)), flagKw_target, compiler__unwrap_literal_tree_arity_1(flagrt.Call(flagKw_target, expr)))
				} else {
					if_result_493 = expr
				}
				if_result_494 = if_result_493
			}
			if_result_495 = if_result_494
		}
		let_result_496 = if_result_495
	}
	return let_result_496
}

func compiler__unwrap_literal_tree_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__unwrap_literal_tree expects exactly 1 arguments")
	}
	return compiler__unwrap_literal_tree_arity_1(args[0])
}

func compiler__macro_definition_q_arity_1(expr flagrt.Value) flagrt.Value {
	var let_result_503 flagrt.Value
	{
		var and_tmp = flagrt.Call(compiler__list_node_q, expr)
		var if_result_502 flagrt.Value
		if flagrt.IsTruthy(and_tmp) {
			var let_result_501 flagrt.Value
			{
				var if_result_497 flagrt.Value
				if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsEmpty(flagrt.Call(compiler__node_children, expr)))) {
					if_result_497 = flagrt.NewBool(false)
				} else {
					if_result_497 = flagrt.NewBool(true)
				}
				var and_tmp = if_result_497
				var if_result_500 flagrt.Value
				if flagrt.IsTruthy(and_tmp) {
					var let_result_499 flagrt.Value
					{
						var and_tmp = flagrt.Call(compiler__symbol_node_q, flagrt.Call(compiler__unwrap_meta, flagrt.First(flagrt.Call(compiler__node_children, expr))))
						var if_result_498 flagrt.Value
						if flagrt.IsTruthy(and_tmp) {
							if_result_498 = flagrt.NewBool(flagrt.Eq(flagrt.NewString("defmacro"), flagrt.Call(flagKw_name, flagrt.Call(compiler__unwrap_meta, flagrt.First(flagrt.Call(compiler__node_children, expr))))))
						} else {
							if_result_498 = and_tmp
						}
						let_result_499 = if_result_498
					}
					if_result_500 = let_result_499
				} else {
					if_result_500 = and_tmp
				}
				let_result_501 = if_result_500
			}
			if_result_502 = let_result_501
		} else {
			if_result_502 = and_tmp
		}
		let_result_503 = if_result_502
	}
	return let_result_503
}

func compiler__macro_definition_q_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__macro_definition_q expects exactly 1 arguments")
	}
	return compiler__macro_definition_q_arity_1(args[0])
}

func compiler__append_all_arity_2(target flagrt.Value, items flagrt.Value) flagrt.Value {
	return func() flagrt.Value {
		var remaining = items
		var out = target
		for {
			var if_result_504 flagrt.Value
			if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsEmpty(remaining))) {
				if_result_504 = out
			} else {
				if_result_504 = flagrt.NewRecur(flagrt.Rest(remaining), flagrt.Conj(out, flagrt.First(remaining)))
			}
			__loopResult := if_result_504
			if __recurValues, __isRecur := flagrt.UnwrapRecur(__loopResult); __isRecur {
				if len(__recurValues) != 2 {
					panic("internal error: recur arity mismatch")
				}
				remaining = __recurValues[0]
				out = __recurValues[1]
				continue
			}
			return __loopResult
		}
	}()
}

func compiler__append_all_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 2 {
		panic("compiler__append_all expects exactly 2 arguments")
	}
	return compiler__append_all_arity_2(args[0], args[1])
}

func compiler__nth_node_arity_2(nodes flagrt.Value, n flagrt.Value) flagrt.Value {
	return flagrt.First(flagrt.Drop(n, nodes))
}

func compiler__nth_node_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 2 {
		panic("compiler__nth_node expects exactly 2 arguments")
	}
	return compiler__nth_node_arity_2(args[0], args[1])
}

func compiler__node_equal_q_arity_2(a flagrt.Value, b flagrt.Value) flagrt.Value {
	var let_result_506 flagrt.Value
	{
		var or_tmp = flagrt.NewBool(flagrt.IsNil(a))
		var if_result_505 flagrt.Value
		if flagrt.IsTruthy(or_tmp) {
			if_result_505 = or_tmp
		} else {
			if_result_505 = flagrt.NewBool(flagrt.IsNil(b))
		}
		let_result_506 = if_result_505
	}
	var if_result_544 flagrt.Value
	if flagrt.IsTruthy(let_result_506) {
		var let_result_508 flagrt.Value
		{
			var and_tmp = flagrt.NewBool(flagrt.IsNil(a))
			var if_result_507 flagrt.Value
			if flagrt.IsTruthy(and_tmp) {
				if_result_507 = flagrt.NewBool(flagrt.IsNil(b))
			} else {
				if_result_507 = and_tmp
			}
			let_result_508 = if_result_507
		}
		if_result_544 = let_result_508
	} else {
		var let_result_543 flagrt.Value
		{
			var kind = flagrt.Call(flagKw_kind, a)
			var if_result_509 flagrt.Value
			if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagrt.Call(flagKw_kind, b)))) {
				if_result_509 = flagrt.NewBool(false)
			} else {
				if_result_509 = flagrt.NewBool(true)
			}
			var if_result_542 flagrt.Value
			if flagrt.IsTruthy(if_result_509) {
				if_result_542 = flagrt.NewBool(false)
			} else {
				var let_result_513 flagrt.Value
				{
					var or_tmp = flagrt.NewBool(flagrt.Eq(kind, flagKw_symbol))
					var if_result_512 flagrt.Value
					if flagrt.IsTruthy(or_tmp) {
						if_result_512 = or_tmp
					} else {
						var let_result_511 flagrt.Value
						{
							var or_tmp = flagrt.NewBool(flagrt.Eq(kind, flagKw_keyword))
							var if_result_510 flagrt.Value
							if flagrt.IsTruthy(or_tmp) {
								if_result_510 = or_tmp
							} else {
								if_result_510 = flagrt.NewBool(flagrt.Eq(kind, flagKw_quoted_symbol))
							}
							let_result_511 = if_result_510
						}
						if_result_512 = let_result_511
					}
					let_result_513 = if_result_512
				}
				var if_result_541 flagrt.Value
				if flagrt.IsTruthy(let_result_513) {
					if_result_541 = flagrt.NewBool(flagrt.Eq(flagrt.Call(flagKw_name, a), flagrt.Call(flagKw_name, b)))
				} else {
					var if_result_540 flagrt.Value
					if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_string))) {
						if_result_540 = flagrt.NewBool(flagrt.Eq(flagrt.Call(flagKw_value, a), flagrt.Call(flagKw_value, b)))
					} else {
						var if_result_539 flagrt.Value
						if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_char))) {
							if_result_539 = flagrt.NewBool(flagrt.Eq(flagrt.Call(flagKw_value, a), flagrt.Call(flagKw_value, b)))
						} else {
							var if_result_538 flagrt.Value
							if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_int))) {
								if_result_538 = flagrt.NewBool(flagrt.Eq(flagrt.Call(flagKw_value, a), flagrt.Call(flagKw_value, b)))
							} else {
								var if_result_537 flagrt.Value
								if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_bigint))) {
									if_result_537 = flagrt.NewBool(flagrt.Eq(flagrt.Call(flagKw_value, a), flagrt.Call(flagKw_value, b)))
								} else {
									var if_result_536 flagrt.Value
									if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_float))) {
										var let_result_515 flagrt.Value
										{
											var and_tmp = flagrt.NewBool(flagrt.Eq(flagrt.Call(flagKw_raw, a), flagrt.Call(flagKw_raw, b)))
											var if_result_514 flagrt.Value
											if flagrt.IsTruthy(and_tmp) {
												if_result_514 = flagrt.NewBool(flagrt.Eq(flagrt.Call(flagKw_value, a), flagrt.Call(flagKw_value, b)))
											} else {
												if_result_514 = and_tmp
											}
											let_result_515 = if_result_514
										}
										if_result_536 = let_result_515
									} else {
										var if_result_535 flagrt.Value
										if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_ratio))) {
											var let_result_517 flagrt.Value
											{
												var and_tmp = flagrt.NewBool(flagrt.Eq(flagrt.Call(flagKw_numerator, a), flagrt.Call(flagKw_numerator, b)))
												var if_result_516 flagrt.Value
												if flagrt.IsTruthy(and_tmp) {
													if_result_516 = flagrt.NewBool(flagrt.Eq(flagrt.Call(flagKw_denominator, a), flagrt.Call(flagKw_denominator, b)))
												} else {
													if_result_516 = and_tmp
												}
												let_result_517 = if_result_516
											}
											if_result_535 = let_result_517
										} else {
											var if_result_534 flagrt.Value
											if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_hash_fn))) {
												if_result_534 = compiler__node_equal_q_arity_2(flagrt.Call(flagKw_body, a), flagrt.Call(flagKw_body, b))
											} else {
												var if_result_533 flagrt.Value
												if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_meta))) {
													var let_result_519 flagrt.Value
													{
														var and_tmp = compiler__node_equal_q_arity_2(flagrt.Call(flagKw_meta, a), flagrt.Call(flagKw_meta, b))
														var if_result_518 flagrt.Value
														if flagrt.IsTruthy(and_tmp) {
															if_result_518 = compiler__node_equal_q_arity_2(flagrt.Call(flagKw_target, a), flagrt.Call(flagKw_target, b))
														} else {
															if_result_518 = and_tmp
														}
														let_result_519 = if_result_518
													}
													if_result_533 = let_result_519
												} else {
													var if_result_532 flagrt.Value
													if flagrt.IsTruthy(flagrt.Call(compiler__walk_node_q, a)) {
														var let_result_524 flagrt.Value
														{
															var as = flagrt.Call(compiler__node_children, a)
															var bs = flagrt.Call(compiler__node_children, b)
															var if_result_520 flagrt.Value
															if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagrt.NewLong(int64(flagrt.Count(as))), flagrt.NewLong(int64(flagrt.Count(bs)))))) {
																if_result_520 = flagrt.NewBool(false)
															} else {
																if_result_520 = flagrt.NewBool(true)
															}
															var if_result_523 flagrt.Value
															if flagrt.IsTruthy(if_result_520) {
																if_result_523 = flagrt.NewBool(false)
															} else {
																if_result_523 = func() flagrt.Value {
																	var left = as
																	var right = bs
																	for {
																		var if_result_522 flagrt.Value
																		if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsEmpty(left))) {
																			if_result_522 = flagrt.NewBool(true)
																		} else {
																			var if_result_521 flagrt.Value
																			if flagrt.IsTruthy(compiler__node_equal_q_arity_2(flagrt.First(left), flagrt.First(right))) {
																				if_result_521 = flagrt.NewRecur(flagrt.Rest(left), flagrt.Rest(right))
																			} else {
																				if_result_521 = flagrt.NewBool(false)
																			}
																			if_result_522 = if_result_521
																		}
																		__loopResult := if_result_522
																		if __recurValues, __isRecur := flagrt.UnwrapRecur(__loopResult); __isRecur {
																			if len(__recurValues) != 2 {
																				panic("internal error: recur arity mismatch")
																			}
																			left = __recurValues[0]
																			right = __recurValues[1]
																			continue
																		}
																		return __loopResult
																	}
																}()
															}
															let_result_524 = if_result_523
														}
														if_result_532 = let_result_524
													} else {
														var if_result_531 flagrt.Value
														if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_quoted_list))) {
															var let_result_529 flagrt.Value
															{
																var as = flagrt.Call(flagKw_elements, a)
																var bs = flagrt.Call(flagKw_elements, b)
																var if_result_525 flagrt.Value
																if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagrt.NewLong(int64(flagrt.Count(as))), flagrt.NewLong(int64(flagrt.Count(bs)))))) {
																	if_result_525 = flagrt.NewBool(false)
																} else {
																	if_result_525 = flagrt.NewBool(true)
																}
																var if_result_528 flagrt.Value
																if flagrt.IsTruthy(if_result_525) {
																	if_result_528 = flagrt.NewBool(false)
																} else {
																	if_result_528 = func() flagrt.Value {
																		var left = as
																		var right = bs
																		for {
																			var if_result_527 flagrt.Value
																			if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsEmpty(left))) {
																				if_result_527 = flagrt.NewBool(true)
																			} else {
																				var if_result_526 flagrt.Value
																				if flagrt.IsTruthy(compiler__node_equal_q_arity_2(flagrt.First(left), flagrt.First(right))) {
																					if_result_526 = flagrt.NewRecur(flagrt.Rest(left), flagrt.Rest(right))
																				} else {
																					if_result_526 = flagrt.NewBool(false)
																				}
																				if_result_527 = if_result_526
																			}
																			__loopResult := if_result_527
																			if __recurValues, __isRecur := flagrt.UnwrapRecur(__loopResult); __isRecur {
																				if len(__recurValues) != 2 {
																					panic("internal error: recur arity mismatch")
																				}
																				left = __recurValues[0]
																				right = __recurValues[1]
																				continue
																			}
																			return __loopResult
																		}
																	}()
																}
																let_result_529 = if_result_528
															}
															if_result_531 = let_result_529
														} else {
															var if_result_530 flagrt.Value
															if flagrt.IsTruthy(flagrt.NewBool(true)) {
																if_result_530 = flagrt.NewBool(false)
															} else {
																if_result_530 = flagrt.NilValue()
															}
															if_result_531 = if_result_530
														}
														if_result_532 = if_result_531
													}
													if_result_533 = if_result_532
												}
												if_result_534 = if_result_533
											}
											if_result_535 = if_result_534
										}
										if_result_536 = if_result_535
									}
									if_result_537 = if_result_536
								}
								if_result_538 = if_result_537
							}
							if_result_539 = if_result_538
						}
						if_result_540 = if_result_539
					}
					if_result_541 = if_result_540
				}
				if_result_542 = if_result_541
			}
			let_result_543 = if_result_542
		}
		if_result_544 = let_result_543
	}
	return if_result_544
}

func compiler__node_equal_q_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 2 {
		panic("compiler__node_equal_q expects exactly 2 arguments")
	}
	return compiler__node_equal_q_arity_2(args[0], args[1])
}

func compiler__parse_macro_params_arity_1(params_node flagrt.Value) flagrt.Value {
	var if_result_545 flagrt.Value
	if flagrt.IsTruthy(flagrt.Call(compiler__vector_node_q, params_node)) {
		if_result_545 = flagrt.NewBool(false)
	} else {
		if_result_545 = flagrt.NewBool(true)
	}
	var if_result_559 flagrt.Value
	if flagrt.IsTruthy(if_result_545) {
		if_result_559 = func() flagrt.Value {
			flagrt.Throw(flagrt.NewMap(flagKw_message, flagrt.NewString("defmacro expects a parameter vector"), flagKw_data, flagrt.NewMap(flagKw_form, params_node)))
			return flagrt.NilValue()
		}()
	} else {
		if_result_559 = func() flagrt.Value {
			var remaining = flagrt.Call(compiler__node_children, params_node)
			var params = flagVec
			for {
				var if_result_558 flagrt.Value
				if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsEmpty(remaining))) {
					if_result_558 = flagrt.NewMap(flagKw_params, params, flagKw_rest_param, flagrt.NilValue())
				} else {
					var let_result_557 flagrt.Value
					{
						var param_node = flagrt.Call(compiler__unwrap_meta, flagrt.First(remaining))
						var if_result_546 flagrt.Value
						if flagrt.IsTruthy(flagrt.Call(compiler__symbol_node_q, param_node)) {
							if_result_546 = flagrt.NewBool(false)
						} else {
							if_result_546 = flagrt.NewBool(true)
						}
						var if_result_556 flagrt.Value
						if flagrt.IsTruthy(if_result_546) {
							if_result_556 = func() flagrt.Value {
								flagrt.Throw(flagrt.NewMap(flagKw_message, flagrt.NewString("defmacro parameters must be symbols"), flagKw_data, flagrt.NewMap(flagKw_form, params_node)))
								return flagrt.NilValue()
							}()
						} else {
							var if_result_555 flagrt.Value
							if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagrt.NewString("&"), flagrt.Call(flagKw_name, param_node)))) {
								var let_result_554 flagrt.Value
								{
									var tail = flagrt.Rest(remaining)
									var if_result_547 flagrt.Value
									if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagrt.NewLong(1), flagrt.NewLong(int64(flagrt.Count(tail)))))) {
										if_result_547 = flagrt.NewBool(false)
									} else {
										if_result_547 = flagrt.NewBool(true)
									}
									var if_result_553 flagrt.Value
									if flagrt.IsTruthy(if_result_547) {
										if_result_553 = func() flagrt.Value {
											flagrt.Throw(flagrt.NewMap(flagKw_message, flagrt.NewString("defmacro varargs must use [& name] at end"), flagKw_data, flagrt.NewMap(flagKw_form, params_node)))
											return flagrt.NilValue()
										}()
									} else {
										var let_result_552 flagrt.Value
										{
											var rest_node = flagrt.Call(compiler__unwrap_meta, flagrt.First(tail))
											var let_result_550 flagrt.Value
											{
												var if_result_548 flagrt.Value
												if flagrt.IsTruthy(flagrt.Call(compiler__symbol_node_q, rest_node)) {
													if_result_548 = flagrt.NewBool(false)
												} else {
													if_result_548 = flagrt.NewBool(true)
												}
												var or_tmp = if_result_548
												var if_result_549 flagrt.Value
												if flagrt.IsTruthy(or_tmp) {
													if_result_549 = or_tmp
												} else {
													if_result_549 = flagrt.NewBool(flagrt.Eq(flagrt.NewString("&"), flagrt.Call(flagKw_name, rest_node)))
												}
												let_result_550 = if_result_549
											}
											var if_result_551 flagrt.Value
											if flagrt.IsTruthy(let_result_550) {
												if_result_551 = func() flagrt.Value {
													flagrt.Throw(flagrt.NewMap(flagKw_message, flagrt.NewString("defmacro varargs expects symbol after &"), flagKw_data, flagrt.NewMap(flagKw_form, params_node)))
													return flagrt.NilValue()
												}()
											} else {
												if_result_551 = flagrt.NewMap(flagKw_params, params, flagKw_rest_param, flagrt.Call(flagKw_name, rest_node))
											}
											let_result_552 = if_result_551
										}
										if_result_553 = let_result_552
									}
									let_result_554 = if_result_553
								}
								if_result_555 = let_result_554
							} else {
								if_result_555 = flagrt.NewRecur(flagrt.Rest(remaining), flagrt.Conj(params, flagrt.Call(flagKw_name, param_node)))
							}
							if_result_556 = if_result_555
						}
						let_result_557 = if_result_556
					}
					if_result_558 = let_result_557
				}
				__loopResult := if_result_558
				if __recurValues, __isRecur := flagrt.UnwrapRecur(__loopResult); __isRecur {
					if len(__recurValues) != 2 {
						panic("internal error: recur arity mismatch")
					}
					remaining = __recurValues[0]
					params = __recurValues[1]
					continue
				}
				return __loopResult
			}
		}()
	}
	return if_result_559
}

func compiler__parse_macro_params_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__parse_macro_params expects exactly 1 arguments")
	}
	return compiler__parse_macro_params_arity_1(args[0])
}

func compiler__compile_multi_arity_arity_3(name flagrt.Value, children flagrt.Value, start flagrt.Value) flagrt.Value {
	var let_result_579 flagrt.Value
	{
		var arity_forms = flagrt.Drop(start, children)
		var if_result_578 flagrt.Value
		if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsEmpty(arity_forms))) {
			if_result_578 = func() flagrt.Value {
				flagrt.Throw(flagrt.NewMap(flagKw_message, flagrt.NewString("defmacro expects at least one arity"), flagKw_data, flagrt.NewMap(flagKw_name, name)))
				return flagrt.NilValue()
			}()
		} else {
			if_result_578 = func() flagrt.Value {
				var remaining = arity_forms
				var arities = flagVec
				var seen = flagMap
				var max_fixed = flagrt.NewLong(-1)
				var rest_min = flagrt.NewLong(-1)
				for {
					var if_result_577 flagrt.Value
					if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsEmpty(remaining))) {
						var let_result_561 bool
						{
							var and_tmp = flagrt.Ge(rest_min, flagrt.NewLong(0))
							var if_result_560 bool
							if and_tmp {
								if_result_560 = flagrt.Gt(max_fixed, rest_min)
							} else {
								if_result_560 = and_tmp
							}
							let_result_561 = if_result_560
						}
						var if_result_562 flagrt.Value
						if let_result_561 {
							if_result_562 = func() flagrt.Value {
								flagrt.Throw(flagrt.NewMap(flagKw_message, flagrt.NewString("defmacro rest arity must have at least as many required parameters as the largest fixed arity"), flagKw_data, flagrt.NewMap(flagKw_name, name)))
								return flagrt.NilValue()
							}()
						} else {
							if_result_562 = flagrt.NewMap(flagKw_name, name, flagKw_arities, arities)
						}
						if_result_577 = if_result_562
					} else {
						var let_result_576 flagrt.Value
						{
							var raw = flagrt.First(remaining)
							var let_result_565 flagrt.Value
							{
								var if_result_563 flagrt.Value
								if flagrt.IsTruthy(flagrt.Call(compiler__list_node_q, raw)) {
									if_result_563 = flagrt.NewBool(false)
								} else {
									if_result_563 = flagrt.NewBool(true)
								}
								var or_tmp = if_result_563
								var if_result_564 flagrt.Value
								if flagrt.IsTruthy(or_tmp) {
									if_result_564 = or_tmp
								} else {
									if_result_564 = flagrt.NewBool(flagrt.Lt(flagrt.NewLong(int64(flagrt.Count(flagrt.Call(compiler__node_children, raw)))), flagrt.NewLong(2)))
								}
								let_result_565 = if_result_564
							}
							var if_result_575 flagrt.Value
							if flagrt.IsTruthy(let_result_565) {
								if_result_575 = func() flagrt.Value {
									flagrt.Throw(flagrt.NewMap(flagKw_message, flagrt.NewString("defmacro arity expects ([params] body...)"), flagKw_data, flagrt.NewMap(flagKw_form, raw)))
									return flagrt.NilValue()
								}()
							} else {
								var let_result_574 flagrt.Value
								{
									var arity_children = flagrt.Call(compiler__node_children, raw)
									var params_expr = flagrt.First(arity_children)
									var if_result_566 flagrt.Value
									if flagrt.IsTruthy(flagrt.Call(compiler__vector_node_q, params_expr)) {
										if_result_566 = flagrt.NewBool(false)
									} else {
										if_result_566 = flagrt.NewBool(true)
									}
									var if_result_573 flagrt.Value
									if flagrt.IsTruthy(if_result_566) {
										if_result_573 = func() flagrt.Value {
											flagrt.Throw(flagrt.NewMap(flagKw_message, flagrt.NewString("defmacro arity expects a parameter vector"), flagKw_data, flagrt.NewMap(flagKw_form, raw)))
											return flagrt.NilValue()
										}()
									} else {
										var let_result_572 flagrt.Value
										{
											var param_spec = flagrt.Call(compiler__parse_macro_params, params_expr)
											var n = flagrt.NewLong(int64(flagrt.Count(flagrt.Call(flagKw_params, param_spec))))
											var rest_param = flagrt.Call(flagKw_rest_param, param_spec)
											var body_items = flagrt.Rest(arity_children)
											var if_result_567 flagrt.Value
											if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagrt.NewLong(1), flagrt.NewLong(int64(flagrt.Count(body_items)))))) {
												if_result_567 = flagrt.First(body_items)
											} else {
												if_result_567 = flagrt.Call(compiler__make_list, flagrt.Cons(flagrt.Call(compiler__make_symbol, flagStr_do), body_items))
											}
											var body = if_result_567
											var if_result_571 flagrt.Value
											if flagrt.IsTruthy(rest_param) {
												var if_result_568 flagrt.Value
												if flagrt.Ge(rest_min, flagrt.NewLong(0)) {
													if_result_568 = func() flagrt.Value {
														flagrt.Throw(flagrt.NewMap(flagKw_message, flagrt.NewString("defmacro supports only one & rest arity"), flagKw_data, flagrt.NewMap(flagKw_name, name)))
														return flagrt.NilValue()
													}()
												} else {
													if_result_568 = flagrt.NewRecur(flagrt.Rest(remaining), flagrt.Conj(arities, flagrt.NewMap(flagKw_params, flagrt.Call(flagKw_params, param_spec), flagKw_rest_param, rest_param, flagKw_body, body)), seen, max_fixed, n)
												}
												if_result_571 = if_result_568
											} else {
												var if_result_570 flagrt.Value
												if flagrt.IsTruthy(flagrt.NewBool(flagrt.Contains(seen, n))) {
													if_result_570 = func() flagrt.Value {
														flagrt.Throw(flagrt.NewMap(flagKw_message, flagrt.NewString(flagrt.Str("duplicate macro arity with ", n, " arguments")), flagKw_data, flagrt.NewMap(flagKw_name, name)))
														return flagrt.NilValue()
													}()
												} else {
													var if_result_569 flagrt.Value
													if flagrt.Gt(n, max_fixed) {
														if_result_569 = n
													} else {
														if_result_569 = max_fixed
													}
													if_result_570 = flagrt.NewRecur(flagrt.Rest(remaining), flagrt.Conj(arities, flagrt.NewMap(flagKw_params, flagrt.Call(flagKw_params, param_spec), flagKw_rest_param, flagrt.NilValue(), flagKw_body, body)), flagrt.Assoc(seen, n, flagrt.NewBool(true)), if_result_569, rest_min)
												}
												if_result_571 = if_result_570
											}
											let_result_572 = if_result_571
										}
										if_result_573 = let_result_572
									}
									let_result_574 = if_result_573
								}
								if_result_575 = let_result_574
							}
							let_result_576 = if_result_575
						}
						if_result_577 = let_result_576
					}
					__loopResult := if_result_577
					if __recurValues, __isRecur := flagrt.UnwrapRecur(__loopResult); __isRecur {
						if len(__recurValues) != 5 {
							panic("internal error: recur arity mismatch")
						}
						remaining = __recurValues[0]
						arities = __recurValues[1]
						seen = __recurValues[2]
						max_fixed = __recurValues[3]
						rest_min = __recurValues[4]
						continue
					}
					return __loopResult
				}
			}()
		}
		let_result_579 = if_result_578
	}
	return let_result_579
}

func compiler__compile_multi_arity_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 3 {
		panic("compiler__compile_multi_arity expects exactly 3 arguments")
	}
	return compiler__compile_multi_arity_arity_3(args[0], args[1], args[2])
}

func compiler__compile_defmacro_arity_1(expr flagrt.Value) flagrt.Value {
	var let_result_591 flagrt.Value
	{
		var children = flagrt.Call(compiler__node_children, expr)
		var if_result_590 flagrt.Value
		if flagrt.Lt(flagrt.NewLong(int64(flagrt.Count(children))), flagrt.NewLong(3)) {
			if_result_590 = func() flagrt.Value {
				flagrt.Throw(flagrt.NewMap(flagKw_message, flagrt.NewString("defmacro expects name, optional docstring, vector params, and body"), flagKw_data, flagrt.NewMap(flagKw_form, expr)))
				return flagrt.NilValue()
			}()
		} else {
			var let_result_589 flagrt.Value
			{
				var name_node = flagrt.Call(compiler__unwrap_meta, flagrt.Call(stdlib__second, children))
				var if_result_580 flagrt.Value
				if flagrt.IsTruthy(flagrt.Call(compiler__symbol_node_q, name_node)) {
					if_result_580 = flagrt.NewBool(false)
				} else {
					if_result_580 = flagrt.NewBool(true)
				}
				var if_result_588 flagrt.Value
				if flagrt.IsTruthy(if_result_580) {
					if_result_588 = func() flagrt.Value {
						flagrt.Throw(flagrt.NewMap(flagKw_message, flagrt.NewString("defmacro expects a macro name"), flagKw_data, flagrt.NewMap(flagKw_form, expr)))
						return flagrt.NilValue()
					}()
				} else {
					var let_result_587 flagrt.Value
					{
						var maybe_doc = flagrt.Call(stdlib__third, children)
						var has_docstring = flagrt.Call(compiler__string_node_q, maybe_doc)
						var if_result_581 flagrt.Value
						if flagrt.IsTruthy(has_docstring) {
							if_result_581 = flagrt.NewLong(3)
						} else {
							if_result_581 = flagrt.NewLong(2)
						}
						var params_index = if_result_581
						var if_result_582 flagrt.Value
						if flagrt.IsTruthy(has_docstring) {
							if_result_582 = flagrt.NewLong(4)
						} else {
							if_result_582 = flagrt.NewLong(3)
						}
						var body_index = if_result_582
						var if_result_586 flagrt.Value
						if flagrt.Ge(params_index, flagrt.NewLong(int64(flagrt.Count(children)))) {
							if_result_586 = func() flagrt.Value {
								flagrt.Throw(flagrt.NewMap(flagKw_message, flagrt.NewString("defmacro expects name, optional docstring, vector params, and body"), flagKw_data, flagrt.NewMap(flagKw_form, expr)))
								return flagrt.NilValue()
							}()
						} else {
							var if_result_585 flagrt.Value
							if flagrt.IsTruthy(flagrt.Call(compiler__list_node_q, flagrt.Call(compiler__nth_node, children, params_index))) {
								if_result_585 = flagrt.Call(compiler__compile_multi_arity, flagrt.Call(flagKw_name, name_node), children, params_index)
							} else {
								var if_result_584 flagrt.Value
								if flagrt.Ge(body_index, flagrt.NewLong(int64(flagrt.Count(children)))) {
									if_result_584 = func() flagrt.Value {
										flagrt.Throw(flagrt.NewMap(flagKw_message, flagrt.NewString("defmacro expects name, optional docstring, vector params, and body"), flagKw_data, flagrt.NewMap(flagKw_form, expr)))
										return flagrt.NilValue()
									}()
								} else {
									var let_result_583 flagrt.Value
									{
										var params_node = flagrt.Call(compiler__nth_node, children, params_index)
										var param_spec = flagrt.Call(compiler__parse_macro_params, params_node)
										let_result_583 = flagrt.NewMap(flagKw_name, flagrt.Call(flagKw_name, name_node), flagKw_params, flagrt.Call(flagKw_params, param_spec), flagKw_rest_param, flagrt.Call(flagKw_rest_param, param_spec), flagKw_body, flagrt.Call(compiler__nth_node, children, body_index), flagKw_arities, flagVec)
									}
									if_result_584 = let_result_583
								}
								if_result_585 = if_result_584
							}
							if_result_586 = if_result_585
						}
						let_result_587 = if_result_586
					}
					if_result_588 = let_result_587
				}
				let_result_589 = if_result_588
			}
			if_result_590 = let_result_589
		}
		let_result_591 = if_result_590
	}
	return let_result_591
}

func compiler__compile_defmacro_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__compile_defmacro expects exactly 1 arguments")
	}
	return compiler__compile_defmacro_arity_1(args[0])
}

func compiler__bind_macro_args_arity_2(arity flagrt.Value, args flagrt.Value) flagrt.Value {
	var let_result_601 flagrt.Value
	{
		var params = flagrt.Call(flagKw_params, arity)
		var rest_param = flagrt.Call(flagKw_rest_param, arity)
		var let_result_594 flagrt.Value
		{
			var and_tmp = flagrt.NewBool(flagrt.IsNil(rest_param))
			var if_result_593 flagrt.Value
			if flagrt.IsTruthy(and_tmp) {
				var if_result_592 flagrt.Value
				if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagrt.NewLong(int64(flagrt.Count(args))), flagrt.NewLong(int64(flagrt.Count(params)))))) {
					if_result_592 = flagrt.NewBool(false)
				} else {
					if_result_592 = flagrt.NewBool(true)
				}
				if_result_593 = if_result_592
			} else {
				if_result_593 = and_tmp
			}
			let_result_594 = if_result_593
		}
		var if_result_600 flagrt.Value
		if flagrt.IsTruthy(let_result_594) {
			if_result_600 = func() flagrt.Value {
				flagrt.Throw(flagrt.NewMap(flagKw_message, flagrt.NewString(flagrt.Str("macro expects exactly ", flagrt.NewLong(int64(flagrt.Count(params))), " arguments")), flagKw_data, flagrt.NewMap(flagKw_expected, flagrt.NewLong(int64(flagrt.Count(params))), flagKw_got, flagrt.NewLong(int64(flagrt.Count(args))))))
				return flagrt.NilValue()
			}()
		} else {
			var let_result_596 flagrt.Value
			{
				var and_tmp = rest_param
				var if_result_595 flagrt.Value
				if flagrt.IsTruthy(and_tmp) {
					if_result_595 = flagrt.NewBool(flagrt.Lt(flagrt.NewLong(int64(flagrt.Count(args))), flagrt.NewLong(int64(flagrt.Count(params)))))
				} else {
					if_result_595 = and_tmp
				}
				let_result_596 = if_result_595
			}
			var if_result_599 flagrt.Value
			if flagrt.IsTruthy(let_result_596) {
				if_result_599 = func() flagrt.Value {
					flagrt.Throw(flagrt.NewMap(flagKw_message, flagrt.NewString(flagrt.Str("macro expects at least ", flagrt.NewLong(int64(flagrt.Count(params))), " arguments")), flagKw_data, flagrt.NewMap(flagKw_expected, flagrt.NewLong(int64(flagrt.Count(params))), flagKw_got, flagrt.NewLong(int64(flagrt.Count(args))))))
					return flagrt.NilValue()
				}()
			} else {
				if_result_599 = func() flagrt.Value {
					var names = params
					var remaining_args = args
					var values = flagMap
					for {
						var if_result_598 flagrt.Value
						if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsEmpty(names))) {
							var if_result_597 flagrt.Value
							if flagrt.IsTruthy(rest_param) {
								if_result_597 = flagrt.NewMap(rest_param, remaining_args)
							} else {
								if_result_597 = flagMap
							}
							if_result_598 = flagrt.NewMap(flagKw_values, values, flagKw_rest_bindings, if_result_597)
						} else {
							if_result_598 = flagrt.NewRecur(flagrt.Rest(names), flagrt.Rest(remaining_args), flagrt.Assoc(values, flagrt.First(names), flagrt.First(remaining_args)))
						}
						__loopResult := if_result_598
						if __recurValues, __isRecur := flagrt.UnwrapRecur(__loopResult); __isRecur {
							if len(__recurValues) != 3 {
								panic("internal error: recur arity mismatch")
							}
							names = __recurValues[0]
							remaining_args = __recurValues[1]
							values = __recurValues[2]
							continue
						}
						return __loopResult
					}
				}()
			}
			if_result_600 = if_result_599
		}
		let_result_601 = if_result_600
	}
	return let_result_601
}

func compiler__bind_macro_args_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 2 {
		panic("compiler__bind_macro_args expects exactly 2 arguments")
	}
	return compiler__bind_macro_args_arity_2(args[0], args[1])
}

func compiler__literal_kind_q_arity_1(kind flagrt.Value) flagrt.Value {
	var let_result_615 flagrt.Value
	{
		var or_tmp = flagrt.NewBool(flagrt.Eq(kind, flagKw_keyword))
		var if_result_614 flagrt.Value
		if flagrt.IsTruthy(or_tmp) {
			if_result_614 = or_tmp
		} else {
			var let_result_613 flagrt.Value
			{
				var or_tmp = flagrt.NewBool(flagrt.Eq(kind, flagKw_string))
				var if_result_612 flagrt.Value
				if flagrt.IsTruthy(or_tmp) {
					if_result_612 = or_tmp
				} else {
					var let_result_611 flagrt.Value
					{
						var or_tmp = flagrt.NewBool(flagrt.Eq(kind, flagKw_char))
						var if_result_610 flagrt.Value
						if flagrt.IsTruthy(or_tmp) {
							if_result_610 = or_tmp
						} else {
							var let_result_609 flagrt.Value
							{
								var or_tmp = flagrt.NewBool(flagrt.Eq(kind, flagKw_int))
								var if_result_608 flagrt.Value
								if flagrt.IsTruthy(or_tmp) {
									if_result_608 = or_tmp
								} else {
									var let_result_607 flagrt.Value
									{
										var or_tmp = flagrt.NewBool(flagrt.Eq(kind, flagKw_bigint))
										var if_result_606 flagrt.Value
										if flagrt.IsTruthy(or_tmp) {
											if_result_606 = or_tmp
										} else {
											var let_result_605 flagrt.Value
											{
												var or_tmp = flagrt.NewBool(flagrt.Eq(kind, flagKw_float))
												var if_result_604 flagrt.Value
												if flagrt.IsTruthy(or_tmp) {
													if_result_604 = or_tmp
												} else {
													var let_result_603 flagrt.Value
													{
														var or_tmp = flagrt.NewBool(flagrt.Eq(kind, flagKw_ratio))
														var if_result_602 flagrt.Value
														if flagrt.IsTruthy(or_tmp) {
															if_result_602 = or_tmp
														} else {
															if_result_602 = flagrt.NewBool(flagrt.Eq(kind, flagKw_quoted_symbol))
														}
														let_result_603 = if_result_602
													}
													if_result_604 = let_result_603
												}
												let_result_605 = if_result_604
											}
											if_result_606 = let_result_605
										}
										let_result_607 = if_result_606
									}
									if_result_608 = let_result_607
								}
								let_result_609 = if_result_608
							}
							if_result_610 = let_result_609
						}
						let_result_611 = if_result_610
					}
					if_result_612 = let_result_611
				}
				let_result_613 = if_result_612
			}
			if_result_614 = let_result_613
		}
		let_result_615 = if_result_614
	}
	return let_result_615
}

func compiler__literal_kind_q_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__literal_kind_q expects exactly 1 arguments")
	}
	return compiler__literal_kind_q_arity_1(args[0])
}

func compiler__macro_case_clause_arity_1(expr flagrt.Value) flagrt.Value {
	var let_result_617 flagrt.Value
	{
		var or_tmp = flagrt.Call(compiler__list_node_q, expr)
		var if_result_616 flagrt.Value
		if flagrt.IsTruthy(or_tmp) {
			if_result_616 = or_tmp
		} else {
			if_result_616 = flagrt.Call(compiler__vector_node_q, expr)
		}
		let_result_617 = if_result_616
	}
	var if_result_621 flagrt.Value
	if flagrt.IsTruthy(let_result_617) {
		var let_result_620 flagrt.Value
		{
			var elems = flagrt.Call(compiler__node_children, expr)
			var if_result_619 flagrt.Value
			if flagrt.Lt(flagrt.NewLong(int64(flagrt.Count(elems))), flagrt.NewLong(2)) {
				if_result_619 = flagrt.NilValue()
			} else {
				var if_result_618 flagrt.Value
				if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagrt.NewLong(2), flagrt.NewLong(int64(flagrt.Count(elems)))))) {
					if_result_618 = flagrt.NewMap(flagKw_pattern, flagrt.First(elems), flagKw_body, flagrt.Call(stdlib__second, elems))
				} else {
					if_result_618 = flagrt.NewMap(flagKw_pattern, flagrt.First(elems), flagKw_body, flagrt.Call(compiler__make_list, flagrt.Cons(flagrt.Call(compiler__make_symbol, flagStr_do), flagrt.Rest(elems))))
				}
				if_result_619 = if_result_618
			}
			let_result_620 = if_result_619
		}
		if_result_621 = let_result_620
	} else {
		if_result_621 = flagrt.NilValue()
	}
	return if_result_621
}

func compiler__macro_case_clause_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__macro_case_clause expects exactly 1 arguments")
	}
	return compiler__macro_case_clause_arity_1(args[0])
}

func compiler__expand_defrecord_macro_arity_1(args flagrt.Value) flagrt.Value {
	var if_result_622 flagrt.Value
	if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagrt.NewLong(2), flagrt.NewLong(int64(flagrt.Count(args)))))) {
		if_result_622 = flagrt.NewBool(false)
	} else {
		if_result_622 = flagrt.NewBool(true)
	}
	var if_result_623 flagrt.Value
	if flagrt.IsTruthy(if_result_622) {
		if_result_623 = func() flagrt.Value {
			flagrt.Throw(flagrt.NewMap(flagKw_message, flagrt.NewString("macro-defrecord expects a record name and field vector"), flagKw_data, flagMap))
			return flagrt.NilValue()
		}()
	} else {
		if_result_623 = flagrt.Call(compiler__make_list, flagrt.NewArray(flagrt.Call(compiler__make_symbol, flagStr_defrecord_), flagrt.First(args), flagrt.Call(stdlib__second, args)))
	}
	return if_result_623
}

func compiler__expand_defrecord_macro_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__expand_defrecord_macro expects exactly 1 arguments")
	}
	return compiler__expand_defrecord_macro_arity_1(args[0])
}

func compiler__macro_op_arity_5(op flagrt.Value, a flagrt.Value, b flagrt.Value, c flagrt.Value, d flagrt.Value) flagrt.Value {
	var if_result_719 flagrt.Value
	if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(op, flagKw_seq))) {
		if_result_719 = func() flagrt.Value {
			var pats = a
			var remaining = b
			var bindings = c
			var rest_bindings = d
			for {
				var if_result_643 flagrt.Value
				if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsEmpty(pats))) {
					var if_result_624 flagrt.Value
					if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsEmpty(remaining))) {
						if_result_624 = flagrt.NewMap(flagKw_matched, flagrt.NewBool(true), flagKw_bindings, bindings, flagKw_rest_bindings, rest_bindings)
					} else {
						if_result_624 = flagMap_1
					}
					if_result_643 = if_result_624
				} else {
					var let_result_642 flagrt.Value
					{
						var pat = flagrt.First(pats)
						var let_result_626 flagrt.Value
						{
							var and_tmp = flagrt.Call(compiler__symbol_node_q, pat)
							var if_result_625 flagrt.Value
							if flagrt.IsTruthy(and_tmp) {
								if_result_625 = flagrt.NewBool(flagrt.Eq(flagrt.NewString("&"), flagrt.Call(flagKw_name, pat)))
							} else {
								if_result_625 = and_tmp
							}
							let_result_626 = if_result_625
						}
						var if_result_641 flagrt.Value
						if flagrt.IsTruthy(let_result_626) {
							var let_result_636 flagrt.Value
							{
								var tail = flagrt.Rest(pats)
								var if_result_627 flagrt.Value
								if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagrt.NewLong(1), flagrt.NewLong(int64(flagrt.Count(tail)))))) {
									if_result_627 = flagrt.NewBool(false)
								} else {
									if_result_627 = flagrt.NewBool(true)
								}
								var if_result_635 flagrt.Value
								if flagrt.IsTruthy(if_result_627) {
									if_result_635 = func() flagrt.Value {
										flagrt.Throw(flagrt.NewMap(flagKw_message, flagrt.NewString("macro-case rest capture must be in the penultimate position"), flagKw_data, flagMap))
										return flagrt.NilValue()
									}()
								} else {
									var let_result_634 flagrt.Value
									{
										var name_node = flagrt.First(tail)
										var let_result_632 flagrt.Value
										{
											var if_result_628 flagrt.Value
											if flagrt.IsTruthy(flagrt.Call(compiler__symbol_node_q, name_node)) {
												if_result_628 = flagrt.NewBool(false)
											} else {
												if_result_628 = flagrt.NewBool(true)
											}
											var or_tmp = if_result_628
											var if_result_631 flagrt.Value
											if flagrt.IsTruthy(or_tmp) {
												if_result_631 = or_tmp
											} else {
												var let_result_630 flagrt.Value
												{
													var or_tmp = flagrt.NewBool(flagrt.Eq(flagrt.NewString(""), flagrt.Call(flagKw_name, name_node)))
													var if_result_629 flagrt.Value
													if flagrt.IsTruthy(or_tmp) {
														if_result_629 = or_tmp
													} else {
														if_result_629 = flagrt.NewBool(flagrt.Eq(flagrt.NewString("&"), flagrt.Call(flagKw_name, name_node)))
													}
													let_result_630 = if_result_629
												}
												if_result_631 = let_result_630
											}
											let_result_632 = if_result_631
										}
										var if_result_633 flagrt.Value
										if flagrt.IsTruthy(let_result_632) {
											if_result_633 = func() flagrt.Value {
												flagrt.Throw(flagrt.NewMap(flagKw_message, flagrt.NewString("macro-case rest capture expects a symbol name"), flagKw_data, flagMap))
												return flagrt.NilValue()
											}()
										} else {
											if_result_633 = flagrt.NewMap(flagKw_matched, flagrt.NewBool(true), flagKw_bindings, bindings, flagKw_rest_bindings, flagrt.Assoc(rest_bindings, flagrt.Call(flagKw_name, name_node), remaining))
										}
										let_result_634 = if_result_633
									}
									if_result_635 = let_result_634
								}
								let_result_636 = if_result_635
							}
							if_result_641 = let_result_636
						} else {
							var if_result_640 flagrt.Value
							if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsEmpty(remaining))) {
								if_result_640 = flagMap_1
							} else {
								var let_result_639 flagrt.Value
								{
									var result = compiler__macro_op_arity_5(flagKw_pat, pat, flagrt.NewArray(flagrt.First(remaining)), bindings, rest_bindings)
									var if_result_637 flagrt.Value
									if flagrt.IsTruthy(flagrt.Call(flagKw_matched, result)) {
										if_result_637 = flagrt.NewBool(false)
									} else {
										if_result_637 = flagrt.NewBool(true)
									}
									var if_result_638 flagrt.Value
									if flagrt.IsTruthy(if_result_637) {
										if_result_638 = flagMap_1
									} else {
										if_result_638 = flagrt.NewRecur(flagrt.Rest(pats), flagrt.Rest(remaining), flagrt.Call(flagKw_bindings, result), flagrt.Call(flagKw_rest_bindings, result))
									}
									let_result_639 = if_result_638
								}
								if_result_640 = let_result_639
							}
							if_result_641 = if_result_640
						}
						let_result_642 = if_result_641
					}
					if_result_643 = let_result_642
				}
				__loopResult := if_result_643
				if __recurValues, __isRecur := flagrt.UnwrapRecur(__loopResult); __isRecur {
					if len(__recurValues) != 4 {
						panic("internal error: recur arity mismatch")
					}
					pats = __recurValues[0]
					remaining = __recurValues[1]
					bindings = __recurValues[2]
					rest_bindings = __recurValues[3]
					continue
				}
				return __loopResult
			}
		}()
	} else {
		var if_result_718 flagrt.Value
		if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(op, flagKw_pat))) {
			var let_result_675 flagrt.Value
			{
				var pat = flagrt.Call(compiler__unwrap_meta, a)
				var if_result_645 flagrt.Value
				if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsEmpty(b))) {
					if_result_645 = b
				} else {
					if_result_645 = func() flagrt.Value {
						var remaining = b
						var out = flagVec
						for {
							var if_result_644 flagrt.Value
							if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsEmpty(remaining))) {
								if_result_644 = out
							} else {
								if_result_644 = flagrt.NewRecur(flagrt.Rest(remaining), flagrt.Conj(out, flagrt.Call(compiler__unwrap_meta, flagrt.Call(compiler__unwrap_literal, flagrt.First(remaining)))))
							}
							__loopResult := if_result_644
							if __recurValues, __isRecur := flagrt.UnwrapRecur(__loopResult); __isRecur {
								if len(__recurValues) != 2 {
									panic("internal error: recur arity mismatch")
								}
								remaining = __recurValues[0]
								out = __recurValues[1]
								continue
							}
							return __loopResult
						}
					}()
				}
				var items = if_result_645
				var bindings = c
				var rest_bindings = d
				var if_result_674 flagrt.Value
				if flagrt.IsTruthy(flagrt.Call(compiler__symbol_node_q, pat)) {
					var if_result_653 flagrt.Value
					if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagrt.NewString("_"), flagrt.Call(flagKw_name, pat)))) {
						var if_result_646 flagrt.Value
						if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagrt.NewLong(1), flagrt.NewLong(int64(flagrt.Count(items)))))) {
							if_result_646 = flagrt.NewMap(flagKw_matched, flagrt.NewBool(true), flagKw_bindings, bindings, flagKw_rest_bindings, rest_bindings)
						} else {
							if_result_646 = flagMap_1
						}
						if_result_653 = if_result_646
					} else {
						var if_result_652 flagrt.Value
						if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagrt.NewString("&"), flagrt.Call(flagKw_name, pat)))) {
							if_result_652 = func() flagrt.Value {
								flagrt.Throw(flagrt.NewMap(flagKw_message, flagrt.NewString("macro-case pattern cannot use bare &"), flagKw_data, flagMap))
								return flagrt.NilValue()
							}()
						} else {
							var if_result_651 flagrt.Value
							if flagrt.IsTruthy(flagrt.NewBool(true)) {
								var if_result_647 flagrt.Value
								if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagrt.NewLong(1), flagrt.NewLong(int64(flagrt.Count(items)))))) {
									if_result_647 = flagrt.NewBool(false)
								} else {
									if_result_647 = flagrt.NewBool(true)
								}
								var if_result_650 flagrt.Value
								if flagrt.IsTruthy(if_result_647) {
									if_result_650 = flagMap_1
								} else {
									var if_result_649 flagrt.Value
									if flagrt.IsTruthy(flagrt.NewBool(flagrt.Contains(bindings, flagrt.Call(flagKw_name, pat)))) {
										var if_result_648 flagrt.Value
										if flagrt.IsTruthy(flagrt.Call(compiler__node_equal_q, flagrt.Call(flagrt.BuiltinFunction("get"), bindings, flagrt.Call(flagKw_name, pat)), flagrt.First(items))) {
											if_result_648 = flagrt.NewMap(flagKw_matched, flagrt.NewBool(true), flagKw_bindings, bindings, flagKw_rest_bindings, rest_bindings)
										} else {
											if_result_648 = flagMap_1
										}
										if_result_649 = if_result_648
									} else {
										if_result_649 = flagrt.NewMap(flagKw_matched, flagrt.NewBool(true), flagKw_bindings, flagrt.Assoc(bindings, flagrt.Call(flagKw_name, pat), flagrt.First(items)), flagKw_rest_bindings, rest_bindings)
									}
									if_result_650 = if_result_649
								}
								if_result_651 = if_result_650
							} else {
								if_result_651 = flagrt.NilValue()
							}
							if_result_652 = if_result_651
						}
						if_result_653 = if_result_652
					}
					if_result_674 = if_result_653
				} else {
					var if_result_673 flagrt.Value
					if flagrt.IsTruthy(flagrt.Call(compiler__literal_kind_q, flagrt.Call(flagKw_kind, pat))) {
						var let_result_655 flagrt.Value
						{
							var and_tmp = flagrt.NewBool(flagrt.Eq(flagrt.NewLong(1), flagrt.NewLong(int64(flagrt.Count(items)))))
							var if_result_654 flagrt.Value
							if flagrt.IsTruthy(and_tmp) {
								if_result_654 = flagrt.Call(compiler__node_equal_q, pat, flagrt.First(items))
							} else {
								if_result_654 = and_tmp
							}
							let_result_655 = if_result_654
						}
						var if_result_656 flagrt.Value
						if flagrt.IsTruthy(let_result_655) {
							if_result_656 = flagrt.NewMap(flagKw_matched, flagrt.NewBool(true), flagKw_bindings, bindings, flagKw_rest_bindings, rest_bindings)
						} else {
							if_result_656 = flagMap_1
						}
						if_result_673 = if_result_656
					} else {
						var if_result_672 flagrt.Value
						if flagrt.IsTruthy(flagrt.Call(compiler__list_node_q, pat)) {
							var let_result_658 flagrt.Value
							{
								var and_tmp = flagrt.NewBool(flagrt.Eq(flagrt.NewLong(1), flagrt.NewLong(int64(flagrt.Count(items)))))
								var if_result_657 flagrt.Value
								if flagrt.IsTruthy(and_tmp) {
									if_result_657 = flagrt.Call(compiler__list_node_q, flagrt.Call(compiler__unwrap_meta, flagrt.First(items)))
								} else {
									if_result_657 = and_tmp
								}
								let_result_658 = if_result_657
							}
							var if_result_659 flagrt.Value
							if flagrt.IsTruthy(let_result_658) {
								if_result_659 = compiler__macro_op_arity_5(flagKw_seq, flagrt.Call(compiler__node_children, pat), flagrt.Call(compiler__node_children, flagrt.Call(compiler__unwrap_meta, flagrt.First(items))), bindings, rest_bindings)
							} else {
								if_result_659 = compiler__macro_op_arity_5(flagKw_seq, flagrt.Call(compiler__node_children, pat), items, bindings, rest_bindings)
							}
							if_result_672 = if_result_659
						} else {
							var if_result_671 flagrt.Value
							if flagrt.IsTruthy(flagrt.Call(compiler__vector_node_q, pat)) {
								var let_result_661 flagrt.Value
								{
									var and_tmp = flagrt.NewBool(flagrt.Eq(flagrt.NewLong(1), flagrt.NewLong(int64(flagrt.Count(items)))))
									var if_result_660 flagrt.Value
									if flagrt.IsTruthy(and_tmp) {
										if_result_660 = flagrt.Call(compiler__vector_node_q, flagrt.Call(compiler__unwrap_meta, flagrt.First(items)))
									} else {
										if_result_660 = and_tmp
									}
									let_result_661 = if_result_660
								}
								var if_result_662 flagrt.Value
								if flagrt.IsTruthy(let_result_661) {
									if_result_662 = compiler__macro_op_arity_5(flagKw_seq, flagrt.Call(compiler__node_children, pat), flagrt.Call(compiler__node_children, flagrt.Call(compiler__unwrap_meta, flagrt.First(items))), bindings, rest_bindings)
								} else {
									if_result_662 = compiler__macro_op_arity_5(flagKw_seq, flagrt.Call(compiler__node_children, pat), items, bindings, rest_bindings)
								}
								if_result_671 = if_result_662
							} else {
								var if_result_670 flagrt.Value
								if flagrt.IsTruthy(flagrt.Call(compiler__pipe_vector_node_q, pat)) {
									var let_result_664 flagrt.Value
									{
										var and_tmp = flagrt.NewBool(flagrt.Eq(flagrt.NewLong(1), flagrt.NewLong(int64(flagrt.Count(items)))))
										var if_result_663 flagrt.Value
										if flagrt.IsTruthy(and_tmp) {
											if_result_663 = flagrt.Call(compiler__pipe_vector_node_q, flagrt.Call(compiler__unwrap_meta, flagrt.First(items)))
										} else {
											if_result_663 = and_tmp
										}
										let_result_664 = if_result_663
									}
									var if_result_665 flagrt.Value
									if flagrt.IsTruthy(let_result_664) {
										if_result_665 = compiler__macro_op_arity_5(flagKw_seq, flagrt.Call(compiler__node_children, pat), flagrt.Call(compiler__node_children, flagrt.Call(compiler__unwrap_meta, flagrt.First(items))), bindings, rest_bindings)
									} else {
										if_result_665 = compiler__macro_op_arity_5(flagKw_seq, flagrt.Call(compiler__node_children, pat), items, bindings, rest_bindings)
									}
									if_result_670 = if_result_665
								} else {
									var if_result_669 flagrt.Value
									if flagrt.IsTruthy(flagrt.NewBool(true)) {
										var let_result_667 flagrt.Value
										{
											var and_tmp = flagrt.NewBool(flagrt.Eq(flagrt.NewLong(1), flagrt.NewLong(int64(flagrt.Count(items)))))
											var if_result_666 flagrt.Value
											if flagrt.IsTruthy(and_tmp) {
												if_result_666 = flagrt.Call(compiler__node_equal_q, pat, flagrt.First(items))
											} else {
												if_result_666 = and_tmp
											}
											let_result_667 = if_result_666
										}
										var if_result_668 flagrt.Value
										if flagrt.IsTruthy(let_result_667) {
											if_result_668 = flagrt.NewMap(flagKw_matched, flagrt.NewBool(true), flagKw_bindings, bindings, flagKw_rest_bindings, rest_bindings)
										} else {
											if_result_668 = flagMap_1
										}
										if_result_669 = if_result_668
									} else {
										if_result_669 = flagrt.NilValue()
									}
									if_result_670 = if_result_669
								}
								if_result_671 = if_result_670
							}
							if_result_672 = if_result_671
						}
						if_result_673 = if_result_672
					}
					if_result_674 = if_result_673
				}
				let_result_675 = if_result_674
			}
			if_result_718 = let_result_675
		} else {
			var if_result_717 flagrt.Value
			if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(op, flagKw_case))) {
				var if_result_686 flagrt.Value
				if flagrt.Lt(flagrt.NewLong(int64(flagrt.Count(a))), flagrt.NewLong(2)) {
					if_result_686 = func() flagrt.Value {
						flagrt.Throw(flagrt.NewMap(flagKw_message, flagrt.NewString("macro-case expects a target form and at least one clause"), flagKw_data, flagMap))
						return flagrt.NilValue()
					}()
				} else {
					var let_result_685 flagrt.Value
					{
						var target_expr = flagrt.First(a)
						var if_result_676 flagrt.Value
						if flagrt.IsTruthy(flagrt.Call(compiler__list_node_q, target_expr)) {
							if_result_676 = flagrt.NewBool(false)
						} else {
							if_result_676 = flagrt.NewBool(true)
						}
						var if_result_684 flagrt.Value
						if flagrt.IsTruthy(if_result_676) {
							if_result_684 = func() flagrt.Value {
								flagrt.Throw(flagrt.NewMap(flagKw_message, flagrt.NewString("macro-case expects a list target form"), flagKw_data, flagMap))
								return flagrt.NilValue()
							}()
						} else {
							var let_result_683 flagrt.Value
							{
								var target_elems = flagrt.Call(compiler__node_children, target_expr)
								var if_result_677 flagrt.Value
								if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsEmpty(target_elems))) {
									if_result_677 = flagVec
								} else {
									if_result_677 = flagrt.Rest(target_elems)
								}
								var target = if_result_677
								let_result_683 = func() flagrt.Value {
									var clauses = flagrt.Rest(a)
									for {
										var if_result_682 flagrt.Value
										if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsEmpty(clauses))) {
											if_result_682 = func() flagrt.Value {
												flagrt.Throw(flagrt.NewMap(flagKw_message, flagrt.NewString("macro-case had no matching clause"), flagKw_data, flagMap))
												return flagrt.NilValue()
											}()
										} else {
											var let_result_681 flagrt.Value
											{
												var clause = flagrt.Call(compiler__macro_case_clause, flagrt.First(clauses))
												var if_result_680 flagrt.Value
												if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsNil(clause))) {
													if_result_680 = func() flagrt.Value {
														flagrt.Throw(flagrt.NewMap(flagKw_message, flagrt.NewString("macro-case clauses must be ([pattern] body) lists"), flagKw_data, flagMap))
														return flagrt.NilValue()
													}()
												} else {
													var let_result_679 flagrt.Value
													{
														var result = compiler__macro_op_arity_5(flagKw_pat, flagrt.Call(flagKw_pattern, clause), target, flagMap, flagMap)
														var if_result_678 flagrt.Value
														if flagrt.IsTruthy(flagrt.Call(flagKw_matched, result)) {
															if_result_678 = compiler__macro_op_arity_5(flagKw_subst, flagrt.Call(flagKw_body, clause), flagrt.Call(flagKw_bindings, result), flagrt.Call(flagKw_rest_bindings, result), flagrt.NilValue())
														} else {
															if_result_678 = flagrt.NewRecur(flagrt.Rest(clauses))
														}
														let_result_679 = if_result_678
													}
													if_result_680 = let_result_679
												}
												let_result_681 = if_result_680
											}
											if_result_682 = let_result_681
										}
										__loopResult := if_result_682
										if __recurValues, __isRecur := flagrt.UnwrapRecur(__loopResult); __isRecur {
											if len(__recurValues) != 1 {
												panic("internal error: recur arity mismatch")
											}
											clauses = __recurValues[0]
											continue
										}
										return __loopResult
									}
								}()
							}
							if_result_684 = let_result_683
						}
						let_result_685 = if_result_684
					}
					if_result_686 = let_result_685
				}
				if_result_717 = if_result_686
			} else {
				var if_result_716 flagrt.Value
				if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(op, flagKw_subst))) {
					var let_result_714 flagrt.Value
					{
						var expr = a
						var values = b
						var rest_bindings = c
						var if_result_713 flagrt.Value
						if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagKw_macro_literal, flagrt.Call(flagKw_kind, expr)))) {
							if_result_713 = expr
						} else {
							var if_result_712 flagrt.Value
							if flagrt.IsTruthy(flagrt.Call(compiler__symbol_node_q, expr)) {
								var let_result_688 flagrt.Value
								{
									var replacement = flagrt.Call(flagrt.BuiltinFunction("get"), values, flagrt.Call(flagKw_name, expr))
									var if_result_687 flagrt.Value
									if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsNil(replacement))) {
										if_result_687 = expr
									} else {
										if_result_687 = flagrt.Call(compiler__wrap_literal, replacement)
									}
									let_result_688 = if_result_687
								}
								if_result_712 = let_result_688
							} else {
								var if_result_711 flagrt.Value
								if flagrt.IsTruthy(flagrt.Call(compiler__splice_node_q, expr)) {
									var let_result_704 flagrt.Value
									{
										var out = func() flagrt.Value {
											var remaining = flagrt.Call(compiler__node_children, expr)
											var acc = flagVec
											for {
												var if_result_694 flagrt.Value
												if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsEmpty(remaining))) {
													if_result_694 = acc
												} else {
													var let_result_693 flagrt.Value
													{
														var child = flagrt.First(remaining)
														var if_result_689 flagrt.Value
														if flagrt.IsTruthy(flagrt.Call(compiler__symbol_node_q, child)) {
															if_result_689 = flagrt.Call(flagrt.BuiltinFunction("get"), rest_bindings, flagrt.Call(flagKw_name, child))
														} else {
															if_result_689 = flagrt.NilValue()
														}
														var rest_items = if_result_689
														var if_result_690 flagrt.Value
														if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsNil(rest_items))) {
															if_result_690 = flagrt.NewBool(false)
														} else {
															if_result_690 = flagrt.NewBool(true)
														}
														var if_result_692 flagrt.Value
														if flagrt.IsTruthy(if_result_690) {
															if_result_692 = flagrt.NewRecur(flagrt.Rest(remaining), func() flagrt.Value {
																var items = rest_items
																var out = acc
																for {
																	var if_result_691 flagrt.Value
																	if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsEmpty(items))) {
																		if_result_691 = out
																	} else {
																		if_result_691 = flagrt.NewRecur(flagrt.Rest(items), flagrt.Conj(out, flagrt.Call(compiler__wrap_literal, flagrt.First(items))))
																	}
																	__loopResult := if_result_691
																	if __recurValues, __isRecur := flagrt.UnwrapRecur(__loopResult); __isRecur {
																		if len(__recurValues) != 2 {
																			panic("internal error: recur arity mismatch")
																		}
																		items = __recurValues[0]
																		out = __recurValues[1]
																		continue
																	}
																	return __loopResult
																}
															}())
														} else {
															if_result_692 = flagrt.NewRecur(flagrt.Rest(remaining), flagrt.Conj(acc, compiler__macro_op_arity_5(flagKw_subst, child, values, rest_bindings, flagrt.NilValue())))
														}
														let_result_693 = if_result_692
													}
													if_result_694 = let_result_693
												}
												__loopResult := if_result_694
												if __recurValues, __isRecur := flagrt.UnwrapRecur(__loopResult); __isRecur {
													if len(__recurValues) != 2 {
														panic("internal error: recur arity mismatch")
													}
													remaining = __recurValues[0]
													acc = __recurValues[1]
													continue
												}
												return __loopResult
											}
										}()
										var let_result_699 flagrt.Value
										{
											var and_tmp = flagrt.Call(compiler__list_node_q, expr)
											var if_result_698 flagrt.Value
											if flagrt.IsTruthy(and_tmp) {
												var let_result_697 flagrt.Value
												{
													var if_result_695 flagrt.Value
													if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsEmpty(out))) {
														if_result_695 = flagrt.NewBool(false)
													} else {
														if_result_695 = flagrt.NewBool(true)
													}
													var and_tmp = if_result_695
													var if_result_696 flagrt.Value
													if flagrt.IsTruthy(and_tmp) {
														if_result_696 = flagrt.Call(compiler__symbol_node_q, flagrt.First(out))
													} else {
														if_result_696 = and_tmp
													}
													let_result_697 = if_result_696
												}
												if_result_698 = let_result_697
											} else {
												if_result_698 = and_tmp
											}
											let_result_699 = if_result_698
										}
										var if_result_703 flagrt.Value
										if flagrt.IsTruthy(let_result_699) {
											var if_result_702 flagrt.Value
											if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagrt.NewString("macro-case"), flagrt.Call(flagKw_name, flagrt.First(out))))) {
												if_result_702 = compiler__macro_op_arity_5(flagKw_case, flagrt.Rest(out), flagrt.NilValue(), flagrt.NilValue(), flagrt.NilValue())
											} else {
												var if_result_701 flagrt.Value
												if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagrt.NewString("macro-defrecord"), flagrt.Call(flagKw_name, flagrt.First(out))))) {
													if_result_701 = flagrt.Call(compiler__expand_defrecord_macro, flagrt.Rest(out))
												} else {
													var if_result_700 flagrt.Value
													if flagrt.IsTruthy(flagrt.NewBool(true)) {
														if_result_700 = flagrt.Call(compiler__with_children, expr, out)
													} else {
														if_result_700 = flagrt.NilValue()
													}
													if_result_701 = if_result_700
												}
												if_result_702 = if_result_701
											}
											if_result_703 = if_result_702
										} else {
											if_result_703 = flagrt.Call(compiler__with_children, expr, out)
										}
										let_result_704 = if_result_703
									}
									if_result_711 = let_result_704
								} else {
									var let_result_706 flagrt.Value
									{
										var or_tmp = flagrt.Call(compiler__map_node_q, expr)
										var if_result_705 flagrt.Value
										if flagrt.IsTruthy(or_tmp) {
											if_result_705 = or_tmp
										} else {
											if_result_705 = flagrt.Call(compiler__set_node_q, expr)
										}
										let_result_706 = if_result_705
									}
									var if_result_710 flagrt.Value
									if flagrt.IsTruthy(let_result_706) {
										if_result_710 = flagrt.Call(compiler__with_children, expr, func() flagrt.Value {
											var remaining = flagrt.Call(compiler__node_children, expr)
											var acc = flagVec
											for {
												var if_result_707 flagrt.Value
												if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsEmpty(remaining))) {
													if_result_707 = acc
												} else {
													if_result_707 = flagrt.NewRecur(flagrt.Rest(remaining), flagrt.Conj(acc, compiler__macro_op_arity_5(flagKw_subst, flagrt.First(remaining), values, rest_bindings, flagrt.NilValue())))
												}
												__loopResult := if_result_707
												if __recurValues, __isRecur := flagrt.UnwrapRecur(__loopResult); __isRecur {
													if len(__recurValues) != 2 {
														panic("internal error: recur arity mismatch")
													}
													remaining = __recurValues[0]
													acc = __recurValues[1]
													continue
												}
												return __loopResult
											}
										}())
									} else {
										var if_result_709 flagrt.Value
										if flagrt.IsTruthy(flagrt.Call(compiler__hash_fn_node_q, expr)) {
											if_result_709 = flagrt.Assoc(expr, flagKw_body, compiler__macro_op_arity_5(flagKw_subst, flagrt.Call(flagKw_body, expr), values, rest_bindings, flagrt.NilValue()))
										} else {
											var if_result_708 flagrt.Value
											if flagrt.IsTruthy(flagrt.Call(compiler__meta_node_q, expr)) {
												if_result_708 = flagrt.Assoc(expr, flagKw_meta, compiler__macro_op_arity_5(flagKw_subst, flagrt.Call(flagKw_meta, expr), values, rest_bindings, flagrt.NilValue()), flagKw_target, compiler__macro_op_arity_5(flagKw_subst, flagrt.Call(flagKw_target, expr), values, rest_bindings, flagrt.NilValue()))
											} else {
												if_result_708 = expr
											}
											if_result_709 = if_result_708
										}
										if_result_710 = if_result_709
									}
									if_result_711 = if_result_710
								}
								if_result_712 = if_result_711
							}
							if_result_713 = if_result_712
						}
						let_result_714 = if_result_713
					}
					if_result_716 = let_result_714
				} else {
					var if_result_715 flagrt.Value
					if flagrt.IsTruthy(flagrt.NewBool(true)) {
						if_result_715 = func() flagrt.Value {
							flagrt.Throw(flagrt.NewMap(flagKw_message, flagrt.NewString("unknown macro-op"), flagKw_data, flagrt.NewMap(flagKw_op, op)))
							return flagrt.NilValue()
						}()
					} else {
						if_result_715 = flagrt.NilValue()
					}
					if_result_716 = if_result_715
				}
				if_result_717 = if_result_716
			}
			if_result_718 = if_result_717
		}
		if_result_719 = if_result_718
	}
	return if_result_719
}

func compiler__macro_op_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 5 {
		panic("compiler__macro_op expects exactly 5 arguments")
	}
	return compiler__macro_op_arity_5(args[0], args[1], args[2], args[3], args[4])
}

func compiler__substitute_node_arity_3(expr flagrt.Value, values flagrt.Value, rest_bindings flagrt.Value) flagrt.Value {
	return flagrt.Call(compiler__macro_op, flagKw_subst, expr, values, rest_bindings, flagrt.NilValue())
}

func compiler__substitute_node_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 3 {
		panic("compiler__substitute_node expects exactly 3 arguments")
	}
	return compiler__substitute_node_arity_3(args[0], args[1], args[2])
}

func compiler__apply_macro_arity_arity_2(arity flagrt.Value, args flagrt.Value) flagrt.Value {
	var let_result_720 flagrt.Value
	{
		var bindings = flagrt.Call(compiler__bind_macro_args, arity, args)
		let_result_720 = flagrt.Call(compiler__unwrap_literal_tree, flagrt.Call(compiler__substitute_node, flagrt.Call(flagKw_body, arity), flagrt.Call(flagKw_values, bindings), flagrt.Call(flagKw_rest_bindings, bindings)))
	}
	return let_result_720
}

func compiler__apply_macro_arity_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 2 {
		panic("compiler__apply_macro_arity expects exactly 2 arguments")
	}
	return compiler__apply_macro_arity_arity_2(args[0], args[1])
}

func compiler__apply_multi_arity_arity_2(macro flagrt.Value, args flagrt.Value) flagrt.Value {
	return func() flagrt.Value {
		var remaining = flagrt.Call(flagKw_arities, macro)
		var rest_arity = flagrt.NilValue()
		for {
			var if_result_729 flagrt.Value
			if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsEmpty(remaining))) {
				var let_result_722 flagrt.Value
				{
					var and_tmp = rest_arity
					var if_result_721 flagrt.Value
					if flagrt.IsTruthy(and_tmp) {
						if_result_721 = flagrt.NewBool(flagrt.Ge(flagrt.NewLong(int64(flagrt.Count(args))), flagrt.NewLong(int64(flagrt.Count(flagrt.Call(flagKw_params, rest_arity))))))
					} else {
						if_result_721 = and_tmp
					}
					let_result_722 = if_result_721
				}
				var if_result_723 flagrt.Value
				if flagrt.IsTruthy(let_result_722) {
					if_result_723 = flagrt.Call(compiler__apply_macro_arity, rest_arity, args)
				} else {
					if_result_723 = func() flagrt.Value {
						flagrt.Throw(flagrt.NewMap(flagKw_message, flagrt.NewString("macro expects arguments"), flagKw_data, flagrt.NewMap(flagKw_got, flagrt.NewLong(int64(flagrt.Count(args))))))
						return flagrt.NilValue()
					}()
				}
				if_result_729 = if_result_723
			} else {
				var let_result_728 flagrt.Value
				{
					var arity = flagrt.First(remaining)
					var let_result_725 flagrt.Value
					{
						var and_tmp = flagrt.NewBool(flagrt.IsNil(flagrt.Call(flagKw_rest_param, arity)))
						var if_result_724 flagrt.Value
						if flagrt.IsTruthy(and_tmp) {
							if_result_724 = flagrt.NewBool(flagrt.Eq(flagrt.NewLong(int64(flagrt.Count(args))), flagrt.NewLong(int64(flagrt.Count(flagrt.Call(flagKw_params, arity))))))
						} else {
							if_result_724 = and_tmp
						}
						let_result_725 = if_result_724
					}
					var if_result_727 flagrt.Value
					if flagrt.IsTruthy(let_result_725) {
						if_result_727 = flagrt.Call(compiler__apply_macro_arity, arity, args)
					} else {
						var if_result_726 flagrt.Value
						if flagrt.IsTruthy(flagrt.Call(flagKw_rest_param, arity)) {
							if_result_726 = arity
						} else {
							if_result_726 = rest_arity
						}
						if_result_727 = flagrt.NewRecur(flagrt.Rest(remaining), if_result_726)
					}
					let_result_728 = if_result_727
				}
				if_result_729 = let_result_728
			}
			__loopResult := if_result_729
			if __recurValues, __isRecur := flagrt.UnwrapRecur(__loopResult); __isRecur {
				if len(__recurValues) != 2 {
					panic("internal error: recur arity mismatch")
				}
				remaining = __recurValues[0]
				rest_arity = __recurValues[1]
				continue
			}
			return __loopResult
		}
	}()
}

func compiler__apply_multi_arity_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 2 {
		panic("compiler__apply_multi_arity expects exactly 2 arguments")
	}
	return compiler__apply_multi_arity_arity_2(args[0], args[1])
}

func compiler__apply_macro_arity_2(macro flagrt.Value, args flagrt.Value) flagrt.Value {
	var if_result_730 flagrt.Value
	if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsEmpty(flagrt.Call(flagKw_arities, macro)))) {
		if_result_730 = flagrt.Call(compiler__apply_macro_arity, flagrt.NewMap(flagKw_params, flagrt.Call(flagKw_params, macro), flagKw_rest_param, flagrt.Call(flagKw_rest_param, macro), flagKw_body, flagrt.Call(flagKw_body, macro)), args)
	} else {
		if_result_730 = flagrt.Call(compiler__apply_multi_arity, macro, args)
	}
	return if_result_730
}

func compiler__apply_macro_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 2 {
		panic("compiler__apply_macro expects exactly 2 arguments")
	}
	return compiler__apply_macro_arity_2(args[0], args[1])
}

func compiler__expand_node_arity_3(expr flagrt.Value, macros flagrt.Value, depth flagrt.Value) flagrt.Value {
	var if_result_742 flagrt.Value
	if flagrt.Gt(depth, flagrt.NewLong(100)) {
		if_result_742 = func() flagrt.Value {
			flagrt.Throw(flagrt.NewMap(flagKw_message, flagrt.NewString("macro expansion depth exceeded"), flagKw_data, flagrt.NewMap(flagKw_expr, expr)))
			return flagrt.NilValue()
		}()
	} else {
		var if_result_741 flagrt.Value
		if flagrt.IsTruthy(flagrt.Call(compiler__list_node_q, expr)) {
			var let_result_736 flagrt.Value
			{
				var children = flagrt.Call(compiler__node_children, expr)
				var if_result_735 flagrt.Value
				if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsEmpty(children))) {
					if_result_735 = expr
				} else {
					var let_result_734 flagrt.Value
					{
						var head = flagrt.First(children)
						var if_result_731 flagrt.Value
						if flagrt.IsTruthy(flagrt.Call(compiler__symbol_node_q, head)) {
							if_result_731 = flagrt.Call(flagrt.BuiltinFunction("get"), macros, flagrt.Call(flagKw_name, head))
						} else {
							if_result_731 = flagrt.NilValue()
						}
						var macro = if_result_731
						var if_result_733 flagrt.Value
						if flagrt.IsTruthy(macro) {
							if_result_733 = compiler__expand_node_arity_3(flagrt.Call(compiler__apply_macro, macro, flagrt.Rest(children)), macros, flagrt.Add(depth, flagrt.NewLong(1)))
						} else {
							if_result_733 = flagrt.Call(compiler__with_children, expr, func() flagrt.Value {
								var remaining = children
								var out = flagVec
								for {
									var if_result_732 flagrt.Value
									if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsEmpty(remaining))) {
										if_result_732 = out
									} else {
										if_result_732 = flagrt.NewRecur(flagrt.Rest(remaining), flagrt.Conj(out, compiler__expand_node_arity_3(flagrt.First(remaining), macros, depth)))
									}
									__loopResult := if_result_732
									if __recurValues, __isRecur := flagrt.UnwrapRecur(__loopResult); __isRecur {
										if len(__recurValues) != 2 {
											panic("internal error: recur arity mismatch")
										}
										remaining = __recurValues[0]
										out = __recurValues[1]
										continue
									}
									return __loopResult
								}
							}())
						}
						let_result_734 = if_result_733
					}
					if_result_735 = let_result_734
				}
				let_result_736 = if_result_735
			}
			if_result_741 = let_result_736
		} else {
			var if_result_740 flagrt.Value
			if flagrt.IsTruthy(flagrt.Call(compiler__walk_node_q, expr)) {
				if_result_740 = flagrt.Call(compiler__with_children, expr, func() flagrt.Value {
					var remaining = flagrt.Call(compiler__node_children, expr)
					var out = flagVec
					for {
						var if_result_737 flagrt.Value
						if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsEmpty(remaining))) {
							if_result_737 = out
						} else {
							if_result_737 = flagrt.NewRecur(flagrt.Rest(remaining), flagrt.Conj(out, compiler__expand_node_arity_3(flagrt.First(remaining), macros, depth)))
						}
						__loopResult := if_result_737
						if __recurValues, __isRecur := flagrt.UnwrapRecur(__loopResult); __isRecur {
							if len(__recurValues) != 2 {
								panic("internal error: recur arity mismatch")
							}
							remaining = __recurValues[0]
							out = __recurValues[1]
							continue
						}
						return __loopResult
					}
				}())
			} else {
				var if_result_739 flagrt.Value
				if flagrt.IsTruthy(flagrt.Call(compiler__hash_fn_node_q, expr)) {
					if_result_739 = flagrt.Assoc(expr, flagKw_body, compiler__expand_node_arity_3(flagrt.Call(flagKw_body, expr), macros, depth))
				} else {
					var if_result_738 flagrt.Value
					if flagrt.IsTruthy(flagrt.Call(compiler__meta_node_q, expr)) {
						if_result_738 = flagrt.Assoc(expr, flagKw_meta, compiler__expand_node_arity_3(flagrt.Call(flagKw_meta, expr), macros, depth), flagKw_target, compiler__expand_node_arity_3(flagrt.Call(flagKw_target, expr), macros, depth))
					} else {
						if_result_738 = expr
					}
					if_result_739 = if_result_738
				}
				if_result_740 = if_result_739
			}
			if_result_741 = if_result_740
		}
		if_result_742 = if_result_741
	}
	return if_result_742
}

func compiler__expand_node_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 3 {
		panic("compiler__expand_node expects exactly 3 arguments")
	}
	return compiler__expand_node_arity_3(args[0], args[1], args[2])
}

func compiler__register_macro_arity_2(macros flagrt.Value, expr flagrt.Value) flagrt.Value {
	var let_result_743 flagrt.Value
	{
		var macro = flagrt.Call(compiler__compile_defmacro, expr)
		let_result_743 = flagrt.Assoc(macros, flagrt.Call(flagKw_name, macro), macro)
	}
	return let_result_743
}

func compiler__register_macro_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 2 {
		panic("compiler__register_macro expects exactly 2 arguments")
	}
	return compiler__register_macro_arity_2(args[0], args[1])
}

func compiler__expand_macros_arity_1(in flagrt.Value) flagrt.Value {
	var let_result_749 flagrt.Value
	{
		var out = flagrt.Call(async__make_channel, flagrt.NewLong(64))
		_ = flagrt.Call(async__go_run, flagrt.NewFunction(func(args ...flagrt.Value) flagrt.Value {
			if len(args) != 0 {
				panic("fn expects exactly 0 arguments")
			}
			return func() flagrt.Value {
				var __flag_try_result flagrt.Value
				defer func() {
					_ = flagrt.Call(async__channel_close, out)
				}()
				func() {
					defer func() {
						r := recover()
						if r == nil {
							return
						}
						__flag_thrown := flagrt.PanicValue(r)
						if flagrt.CatchMatches("Exception", __flag_thrown) {
							__flag_try_result = func() flagrt.Value {
								var e = __flag_thrown
								_ = e
								return flagrt.Call(async__channel_send, out, flagrt.NewMap(flagKw_kind, flagKw_error, flagKw_message, flagrt.ExMessage(e)))
							}()
							return
						}
						panic(r)
					}()
					__flag_try_result = func() flagrt.Value {
						var macros = flagMap
						for {
							var let_result_748 flagrt.Value
							{
								var expr = flagrt.Call(async__channel_receive, in)
								var if_result_747 flagrt.Value
								if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsNil(expr))) {
									if_result_747 = flagrt.NilValue()
								} else {
									var if_result_746 flagrt.Value
									if flagrt.IsTruthy(flagrt.Call(compiler__macro_definition_q, expr)) {
										if_result_746 = flagrt.NewRecur(flagrt.Call(compiler__register_macro, macros, expr))
									} else {
										var let_result_745 flagrt.Value
										{
											var expanded = flagrt.Call(compiler__expand_node, expr, macros, flagrt.NewLong(0))
											var if_result_744 flagrt.Value
											if flagrt.IsTruthy(flagrt.Call(compiler__macro_definition_q, expanded)) {
												if_result_744 = flagrt.NewRecur(flagrt.Call(compiler__register_macro, macros, expanded))
											} else {
												_ = flagrt.Call(async__channel_send, out, expanded)
												if_result_744 = flagrt.NewRecur(macros)
											}
											let_result_745 = if_result_744
										}
										if_result_746 = let_result_745
									}
									if_result_747 = if_result_746
								}
								let_result_748 = if_result_747
							}
							__loopResult := let_result_748
							if __recurValues, __isRecur := flagrt.UnwrapRecur(__loopResult); __isRecur {
								if len(__recurValues) != 1 {
									panic("internal error: recur arity mismatch")
								}
								macros = __recurValues[0]
								continue
							}
							return __loopResult
						}
					}()
				}()
				return __flag_try_result
			}()
		}))
		let_result_749 = out
	}
	return let_result_749
}

func compiler__expand_macros_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__expand_macros expects exactly 1 arguments")
	}
	return compiler__expand_macros_arity_1(args[0])
}

func compiler__parse_file_arity_1(path flagrt.Value) flagrt.Value {
	return flagrt.Call(compiler__expand_macros, flagrt.Call(compiler__build_ast_from_tokens, flagrt.Call(compiler__tokenize_file, path)))
}

func compiler__parse_file_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__parse_file expects exactly 1 arguments")
	}
	return compiler__parse_file_arity_1(args[0])
}

func compiler_lift__lift_node_arity_1(expr flagrt.Value) flagrt.Value {
	return expr
}

func compiler_lift__lift_node_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler_lift__lift_node expects exactly 1 arguments")
	}
	return compiler_lift__lift_node_arity_1(args[0])
}

func compiler_lift__pump_lift_bang_arity_2(in flagrt.Value, out flagrt.Value) flagrt.Value {
	var let_result_751 flagrt.Value
	{
		var expr = flagrt.Call(async__channel_receive, in)
		var if_result_750 flagrt.Value
		if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsNil(expr))) {
			if_result_750 = flagrt.Call(async__channel_close, out)
		} else {
			_ = flagrt.Call(async__channel_send, out, flagrt.Call(compiler_lift__lift_node, expr))
			if_result_750 = compiler_lift__pump_lift_bang_arity_2(in, out)
		}
		let_result_751 = if_result_750
	}
	return let_result_751
}

func compiler_lift__pump_lift_bang_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 2 {
		panic("compiler_lift__pump_lift_bang expects exactly 2 arguments")
	}
	return compiler_lift__pump_lift_bang_arity_2(args[0], args[1])
}

func compiler_lift__lift_ast_channel_arity_1(in flagrt.Value) flagrt.Value {
	var let_result_752 flagrt.Value
	{
		var out = flagrt.Call(async__make_channel, flagrt.NewLong(64))
		_ = flagrt.Call(async__go_run, flagrt.NewFunction(func(args ...flagrt.Value) flagrt.Value {
			if len(args) != 0 {
				panic("fn expects exactly 0 arguments")
			}
			return flagrt.Call(compiler_lift__pump_lift_bang, in, out)
		}))
		let_result_752 = out
	}
	return let_result_752
}

func compiler_lift__lift_ast_channel_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler_lift__lift_ast_channel expects exactly 1 arguments")
	}
	return compiler_lift__lift_ast_channel_arity_1(args[0])
}

func compiler_lift__lift_file_arity_1(path flagrt.Value) flagrt.Value {
	return flagrt.Call(compiler_lift__lift_ast_channel, flagrt.Call(compiler__parse_file, path))
}

func compiler_lift__lift_file_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler_lift__lift_file expects exactly 1 arguments")
	}
	return compiler_lift__lift_file_arity_1(args[0])
}

func compiler_ir__ir_ident_arity_1(name flagrt.Value) flagrt.Value {
	return flagrt.NewMap(flagKw_kind, flagKw_ident, flagKw_name, name)
}

func compiler_ir__ir_ident_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler_ir__ir_ident expects exactly 1 arguments")
	}
	return compiler_ir__ir_ident_arity_1(args[0])
}

func compiler_ir__ir_string_arity_1(value flagrt.Value) flagrt.Value {
	return flagrt.NewMap(flagKw_kind, flagKw_string, flagKw_value, value)
}

func compiler_ir__ir_string_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler_ir__ir_string expects exactly 1 arguments")
	}
	return compiler_ir__ir_string_arity_1(args[0])
}

func compiler_ir__ir_int_arity_1(value flagrt.Value) flagrt.Value {
	return flagrt.NewMap(flagKw_kind, flagKw_int, flagKw_value, value)
}

func compiler_ir__ir_int_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler_ir__ir_int expects exactly 1 arguments")
	}
	return compiler_ir__ir_int_arity_1(args[0])
}

func compiler_ir__ir_selector_arity_2(pkg flagrt.Value, name flagrt.Value) flagrt.Value {
	return flagrt.NewMap(flagKw_kind, flagKw_selector, flagKw_pkg, pkg, flagKw_name, name)
}

func compiler_ir__ir_selector_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 2 {
		panic("compiler_ir__ir_selector expects exactly 2 arguments")
	}
	return compiler_ir__ir_selector_arity_2(args[0], args[1])
}

func compiler_ir__ir_call_arity_2(fun flagrt.Value, args flagrt.Value) flagrt.Value {
	return flagrt.NewMap(flagKw_kind, flagKw_call, flagKw_fun, fun, flagKw_args, args)
}

func compiler_ir__ir_call_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 2 {
		panic("compiler_ir__ir_call expects exactly 2 arguments")
	}
	return compiler_ir__ir_call_arity_2(args[0], args[1])
}

func compiler_ir__ir_raw_arity_1(code flagrt.Value) flagrt.Value {
	return flagrt.NewMap(flagKw_kind, flagKw_raw, flagKw_code, code)
}

func compiler_ir__ir_raw_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler_ir__ir_raw expects exactly 1 arguments")
	}
	return compiler_ir__ir_raw_arity_1(args[0])
}

func compiler_ir__ir_index_arity_2(x flagrt.Value, index flagrt.Value) flagrt.Value {
	return flagrt.NewMap(flagKw_kind, flagKw_index, flagKw_x, x, flagKw_index, index)
}

func compiler_ir__ir_index_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 2 {
		panic("compiler_ir__ir_index expects exactly 2 arguments")
	}
	return compiler_ir__ir_index_arity_2(args[0], args[1])
}

func compiler_ir__ir_slice_arity_3(x flagrt.Value, low flagrt.Value, high flagrt.Value) flagrt.Value {
	return flagrt.NewMap(flagKw_kind, flagKw_slice, flagKw_x, x, flagKw_low, low, flagKw_high, high)
}

func compiler_ir__ir_slice_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 3 {
		panic("compiler_ir__ir_slice expects exactly 3 arguments")
	}
	return compiler_ir__ir_slice_arity_3(args[0], args[1], args[2])
}

func compiler_ir__ir_spread_arity_1(expr flagrt.Value) flagrt.Value {
	return flagrt.NewMap(flagKw_kind, flagKw_spread, flagKw_expr, expr)
}

func compiler_ir__ir_spread_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler_ir__ir_spread expects exactly 1 arguments")
	}
	return compiler_ir__ir_spread_arity_1(args[0])
}

func compiler_ir__ir_unary_arity_2(op flagrt.Value, x flagrt.Value) flagrt.Value {
	return flagrt.NewMap(flagKw_kind, flagKw_unary, flagKw_op, op, flagKw_x, x)
}

func compiler_ir__ir_unary_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 2 {
		panic("compiler_ir__ir_unary expects exactly 2 arguments")
	}
	return compiler_ir__ir_unary_arity_2(args[0], args[1])
}

func compiler_ir__ir_binary_arity_3(op flagrt.Value, left flagrt.Value, right flagrt.Value) flagrt.Value {
	return flagrt.NewMap(flagKw_kind, flagKw_binary, flagKw_op, op, flagKw_left, left, flagKw_right, right)
}

func compiler_ir__ir_binary_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 3 {
		panic("compiler_ir__ir_binary expects exactly 3 arguments")
	}
	return compiler_ir__ir_binary_arity_3(args[0], args[1], args[2])
}

func compiler_ir__ir_func_lit_arity_3(params flagrt.Value, result flagrt.Value, body flagrt.Value) flagrt.Value {
	return flagrt.NewMap(flagKw_kind, flagKw_func_lit, flagKw_params, params, flagKw_result, result, flagKw_body, body)
}

func compiler_ir__ir_func_lit_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 3 {
		panic("compiler_ir__ir_func_lit expects exactly 3 arguments")
	}
	return compiler_ir__ir_func_lit_arity_3(args[0], args[1], args[2])
}

func compiler_ir__ir_expr_stmt_arity_2(expr flagrt.Value, discard flagrt.Value) flagrt.Value {
	return flagrt.NewMap(flagKw_kind, flagKw_expr_stmt, flagKw_expr, expr, flagKw_discard, discard)
}

func compiler_ir__ir_expr_stmt_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 2 {
		panic("compiler_ir__ir_expr_stmt expects exactly 2 arguments")
	}
	return compiler_ir__ir_expr_stmt_arity_2(args[0], args[1])
}

func compiler_ir__ir_return_arity_1(expr flagrt.Value) flagrt.Value {
	return flagrt.NewMap(flagKw_kind, flagKw_return, flagKw_expr, expr)
}

func compiler_ir__ir_return_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler_ir__ir_return expects exactly 1 arguments")
	}
	return compiler_ir__ir_return_arity_1(args[0])
}

func compiler_ir__ir_defer_arity_1(expr flagrt.Value) flagrt.Value {
	return flagrt.NewMap(flagKw_kind, flagKw_defer, flagKw_expr, expr)
}

func compiler_ir__ir_defer_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler_ir__ir_defer expects exactly 1 arguments")
	}
	return compiler_ir__ir_defer_arity_1(args[0])
}

func compiler_ir__ir_go_arity_1(expr flagrt.Value) flagrt.Value {
	return flagrt.NewMap(flagKw_kind, flagKw_go, flagKw_expr, expr)
}

func compiler_ir__ir_go_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler_ir__ir_go expects exactly 1 arguments")
	}
	return compiler_ir__ir_go_arity_1(args[0])
}

func compiler_ir__ir_var_arity_3(name flagrt.Value, type_ flagrt.Value, expr flagrt.Value) flagrt.Value {
	return flagrt.NewMap(flagKw_kind, flagKw_var, flagKw_name, name, flagKw_type, type_, flagKw_expr, expr)
}

func compiler_ir__ir_var_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 3 {
		panic("compiler_ir__ir_var expects exactly 3 arguments")
	}
	return compiler_ir__ir_var_arity_3(args[0], args[1], args[2])
}

func compiler_ir__ir_assign_arity_2(name flagrt.Value, expr flagrt.Value) flagrt.Value {
	return flagrt.NewMap(flagKw_kind, flagKw_assign, flagKw_name, name, flagKw_expr, expr)
}

func compiler_ir__ir_assign_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 2 {
		panic("compiler_ir__ir_assign expects exactly 2 arguments")
	}
	return compiler_ir__ir_assign_arity_2(args[0], args[1])
}

func compiler_ir__ir_define_arity_2(names flagrt.Value, expr flagrt.Value) flagrt.Value {
	return flagrt.NewMap(flagKw_kind, flagKw_define, flagKw_names, names, flagKw_expr, expr)
}

func compiler_ir__ir_define_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 2 {
		panic("compiler_ir__ir_define expects exactly 2 arguments")
	}
	return compiler_ir__ir_define_arity_2(args[0], args[1])
}

func compiler_ir__ir_if_arity_3(cond flagrt.Value, then flagrt.Value, else_ flagrt.Value) flagrt.Value {
	return flagrt.NewMap(flagKw_kind, flagKw_if, flagKw_init, flagrt.NewString(""), flagKw_cond, cond, flagKw_then, then, flagKw_else, else_)
}

func compiler_ir__ir_if_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 3 {
		panic("compiler_ir__ir_if expects exactly 3 arguments")
	}
	return compiler_ir__ir_if_arity_3(args[0], args[1], args[2])
}

func compiler_ir__ir_if_init_arity_4(init flagrt.Value, cond flagrt.Value, then flagrt.Value, else_ flagrt.Value) flagrt.Value {
	return flagrt.NewMap(flagKw_kind, flagKw_if, flagKw_init, init, flagKw_cond, cond, flagKw_then, then, flagKw_else, else_)
}

func compiler_ir__ir_if_init_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 4 {
		panic("compiler_ir__ir_if_init expects exactly 4 arguments")
	}
	return compiler_ir__ir_if_init_arity_4(args[0], args[1], args[2], args[3])
}

func compiler_ir__ir_for_arity_1(body flagrt.Value) flagrt.Value {
	return flagrt.NewMap(flagKw_kind, flagKw_for, flagKw_body, body)
}

func compiler_ir__ir_for_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler_ir__ir_for expects exactly 1 arguments")
	}
	return compiler_ir__ir_for_arity_1(args[0])
}

func compiler_ir__ir_block_arity_1(body flagrt.Value) flagrt.Value {
	return flagrt.NewMap(flagKw_kind, flagKw_block, flagKw_body, body)
}

func compiler_ir__ir_block_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler_ir__ir_block expects exactly 1 arguments")
	}
	return compiler_ir__ir_block_arity_1(args[0])
}

func compiler_ir__ir_raw_stmt_arity_1(code flagrt.Value) flagrt.Value {
	return flagrt.NewMap(flagKw_kind, flagKw_raw_stmt, flagKw_code, code)
}

func compiler_ir__ir_raw_stmt_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler_ir__ir_raw_stmt expects exactly 1 arguments")
	}
	return compiler_ir__ir_raw_stmt_arity_1(args[0])
}

func compiler_ir__rt_call_arity_2(name flagrt.Value, args flagrt.Value) flagrt.Value {
	return flagrt.Call(compiler_ir__ir_call, flagrt.Call(compiler_ir__ir_selector, compiler_ir__runtime_alias, name), args)
}

func compiler_ir__rt_call_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 2 {
		panic("compiler_ir__rt_call expects exactly 2 arguments")
	}
	return compiler_ir__rt_call_arity_2(args[0], args[1])
}

func compiler_ir__ident_call_arity_2(name flagrt.Value, args flagrt.Value) flagrt.Value {
	return flagrt.Call(compiler_ir__ir_call, flagrt.Call(compiler_ir__ir_ident, name), args)
}

func compiler_ir__ident_call_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 2 {
		panic("compiler_ir__ident_call expects exactly 2 arguments")
	}
	return compiler_ir__ident_call_arity_2(args[0], args[1])
}

func compiler_ir__selector_call_arity_3(pkg flagrt.Value, name flagrt.Value, args flagrt.Value) flagrt.Value {
	return flagrt.Call(compiler_ir__ir_call, flagrt.Call(compiler_ir__ir_selector, pkg, name), args)
}

func compiler_ir__selector_call_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 3 {
		panic("compiler_ir__selector_call expects exactly 3 arguments")
	}
	return compiler_ir__selector_call_arity_3(args[0], args[1], args[2])
}

func compiler_ir__parse_go_fun_arity_1(go_name flagrt.Value) flagrt.Value {
	var let_result_761 flagrt.Value
	{
		var idx = flagrt.Call(flagrt.GoBind_packages_StringIndexOf, go_name, flagStr___12)
		var if_result_760 flagrt.Value
		if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsNil(idx))) {
			if_result_760 = flagrt.Call(compiler_ir__ir_ident, go_name)
		} else {
			var let_result_759 flagrt.Value
			{
				var pkg = flagrt.Call(flagrt.BuiltinFunction("subs"), go_name, flagrt.NewLong(0), idx)
				var name = flagrt.Call(flagrt.BuiltinFunction("subs"), go_name, flagrt.Call(inc, idx))
				var let_result_757 flagrt.Value
				{
					var or_tmp = flagrt.NewBool(flagrt.Eq(pkg, flagrt.NewString("")))
					var if_result_756 flagrt.Value
					if flagrt.IsTruthy(or_tmp) {
						if_result_756 = or_tmp
					} else {
						var let_result_755 flagrt.Value
						{
							var or_tmp = flagrt.NewBool(flagrt.Eq(name, flagrt.NewString("")))
							var if_result_754 flagrt.Value
							if flagrt.IsTruthy(or_tmp) {
								if_result_754 = or_tmp
							} else {
								var if_result_753 flagrt.Value
								if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsNil(flagrt.Call(flagrt.GoBind_packages_StringIndexOf, name, flagStr___12)))) {
									if_result_753 = flagrt.NewBool(false)
								} else {
									if_result_753 = flagrt.NewBool(true)
								}
								if_result_754 = if_result_753
							}
							let_result_755 = if_result_754
						}
						if_result_756 = let_result_755
					}
					let_result_757 = if_result_756
				}
				var if_result_758 flagrt.Value
				if flagrt.IsTruthy(let_result_757) {
					if_result_758 = flagrt.Call(compiler_ir__ir_ident, go_name)
				} else {
					if_result_758 = flagrt.Call(compiler_ir__ir_selector, pkg, name)
				}
				let_result_759 = if_result_758
			}
			if_result_760 = let_result_759
		}
		let_result_761 = if_result_760
	}
	return let_result_761
}

func compiler_ir__parse_go_fun_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler_ir__parse_go_fun expects exactly 1 arguments")
	}
	return compiler_ir__parse_go_fun_arity_1(args[0])
}

func compiler_ir__iife_arity_2(result flagrt.Value, body flagrt.Value) flagrt.Value {
	return flagrt.Call(compiler_ir__ir_call, flagrt.Call(compiler_ir__ir_func_lit, flagStr_, result, body), flagVec)
}

func compiler_ir__iife_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 2 {
		panic("compiler_ir__iife expects exactly 2 arguments")
	}
	return compiler_ir__iife_arity_2(args[0], args[1])
}

func compiler_ir__value_iife_arity_1(body flagrt.Value) flagrt.Value {
	return flagrt.Call(compiler_ir__iife, flagrt.NewString(flagrt.Str(compiler_ir__runtime_alias, ".Value")), body)
}

func compiler_ir__value_iife_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler_ir__value_iife expects exactly 1 arguments")
	}
	return compiler_ir__value_iife_arity_1(args[0])
}

func compiler_ir__blank_q_arity_1(s flagrt.Value) flagrt.Value {
	var let_result_763 flagrt.Value
	{
		var or_tmp = flagrt.NewBool(flagrt.IsNil(s))
		var if_result_762 flagrt.Value
		if flagrt.IsTruthy(or_tmp) {
			if_result_762 = or_tmp
		} else {
			if_result_762 = flagrt.NewBool(flagrt.Eq(flagrt.NewString(""), s))
		}
		let_result_763 = if_result_762
	}
	return let_result_763
}

func compiler_ir__blank_q_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler_ir__blank_q expects exactly 1 arguments")
	}
	return compiler_ir__blank_q_arity_1(args[0])
}

func compiler_ir__empty_seq_q_arity_1(xs flagrt.Value) flagrt.Value {
	var let_result_765 flagrt.Value
	{
		var or_tmp = flagrt.NewBool(flagrt.IsNil(xs))
		var if_result_764 flagrt.Value
		if flagrt.IsTruthy(or_tmp) {
			if_result_764 = or_tmp
		} else {
			if_result_764 = flagrt.NewBool(flagrt.IsEmpty(xs))
		}
		let_result_765 = if_result_764
	}
	return let_result_765
}

func compiler_ir__empty_seq_q_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler_ir__empty_seq_q expects exactly 1 arguments")
	}
	return compiler_ir__empty_seq_q_arity_1(args[0])
}

func compiler_ir__join_names_arity_1(names flagrt.Value) flagrt.Value {
	return func() flagrt.Value {
		var remaining = names
		var out = flagStr_
		var first_q = flagrt.NewBool(true)
		for {
			var if_result_768 flagrt.Value
			if flagrt.IsTruthy(flagrt.Call(compiler_ir__empty_seq_q, remaining)) {
				if_result_768 = out
			} else {
				var let_result_767 flagrt.Value
				{
					var piece = flagrt.Str(flagrt.First(remaining))
					var if_result_766 flagrt.Value
					if flagrt.IsTruthy(first_q) {
						if_result_766 = flagrt.NewRecur(flagrt.Rest(remaining), flagrt.NewString(piece), flagrt.NewBool(false))
					} else {
						if_result_766 = flagrt.NewRecur(flagrt.Rest(remaining), flagrt.NewString(flagrt.Str(out, ", ", piece)), flagrt.NewBool(false))
					}
					let_result_767 = if_result_766
				}
				if_result_768 = let_result_767
			}
			__loopResult := if_result_768
			if __recurValues, __isRecur := flagrt.UnwrapRecur(__loopResult); __isRecur {
				if len(__recurValues) != 3 {
					panic("internal error: recur arity mismatch")
				}
				remaining = __recurValues[0]
				out = __recurValues[1]
				first_q = __recurValues[2]
				continue
			}
			return __loopResult
		}
	}()
}

func compiler_ir__join_names_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler_ir__join_names expects exactly 1 arguments")
	}
	return compiler_ir__join_names_arity_1(args[0])
}

func compiler_ir__render_ir_node_arity_3(mode flagrt.Value, node flagrt.Value, indent flagrt.Value) flagrt.Value {
	var if_result_819 flagrt.Value
	if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(mode, flagKw_stmts))) {
		if_result_819 = func() flagrt.Value {
			var remaining = node
			var out = flagStr_
			for {
				var if_result_769 flagrt.Value
				if flagrt.IsTruthy(flagrt.Call(compiler_ir__empty_seq_q, remaining)) {
					if_result_769 = out
				} else {
					if_result_769 = flagrt.NewRecur(flagrt.Rest(remaining), flagrt.NewString(flagrt.Str(out, compiler_ir__render_ir_node_arity_3(flagKw_stmt, flagrt.First(remaining), indent))))
				}
				__loopResult := if_result_769
				if __recurValues, __isRecur := flagrt.UnwrapRecur(__loopResult); __isRecur {
					if len(__recurValues) != 2 {
						panic("internal error: recur arity mismatch")
					}
					remaining = __recurValues[0]
					out = __recurValues[1]
					continue
				}
				return __loopResult
			}
		}()
	} else {
		var if_result_818 flagrt.Value
		if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(mode, flagKw_expr))) {
			var let_result_792 flagrt.Value
			{
				var kind = flagrt.Call(flagKw_kind, node)
				var if_result_791 flagrt.Value
				if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_ident))) {
					if_result_791 = flagrt.Call(flagKw_name, node)
				} else {
					var if_result_790 flagrt.Value
					if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_string))) {
						if_result_790 = flagrt.NewString(flagrt.Format("%q", flagrt.Call(flagKw_value, node)))
					} else {
						var if_result_789 flagrt.Value
						if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_int))) {
							if_result_789 = flagrt.NewString(flagrt.Str(flagrt.Call(flagKw_value, node)))
						} else {
							var if_result_788 flagrt.Value
							if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_selector))) {
								var if_result_770 flagrt.Value
								if flagrt.IsTruthy(flagrt.Call(compiler_ir__blank_q, flagrt.Call(flagKw_pkg, node))) {
									if_result_770 = flagrt.Call(flagKw_name, node)
								} else {
									if_result_770 = flagrt.NewString(flagrt.Str(flagrt.Call(flagKw_pkg, node), ".", flagrt.Call(flagKw_name, node)))
								}
								if_result_788 = if_result_770
							} else {
								var if_result_787 flagrt.Value
								if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_call))) {
									if_result_787 = func() flagrt.Value {
										var remaining = flagrt.Call(flagKw_args, node)
										var out = flagStr_
										var first_q = flagrt.NewBool(true)
										for {
											var if_result_774 flagrt.Value
											if flagrt.IsTruthy(flagrt.Call(compiler_ir__empty_seq_q, remaining)) {
												if_result_774 = flagrt.NewString(flagrt.Str(compiler_ir__render_ir_node_arity_3(flagKw_expr, flagrt.Call(flagKw_fun, node), flagStr_), "(", out, ")"))
											} else {
												var let_result_773 flagrt.Value
												{
													var arg = flagrt.First(remaining)
													var if_result_771 flagrt.Value
													if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagKw_spread, flagrt.Call(flagKw_kind, arg)))) {
														if_result_771 = flagrt.NewString(flagrt.Str(compiler_ir__render_ir_node_arity_3(flagKw_expr, flagrt.Call(flagKw_expr, arg), flagStr_), "..."))
													} else {
														if_result_771 = compiler_ir__render_ir_node_arity_3(flagKw_expr, arg, flagStr_)
													}
													var piece = if_result_771
													var if_result_772 flagrt.Value
													if flagrt.IsTruthy(first_q) {
														if_result_772 = flagrt.NewRecur(flagrt.Rest(remaining), piece, flagrt.NewBool(false))
													} else {
														if_result_772 = flagrt.NewRecur(flagrt.Rest(remaining), flagrt.NewString(flagrt.Str(out, ", ", piece)), flagrt.NewBool(false))
													}
													let_result_773 = if_result_772
												}
												if_result_774 = let_result_773
											}
											__loopResult := if_result_774
											if __recurValues, __isRecur := flagrt.UnwrapRecur(__loopResult); __isRecur {
												if len(__recurValues) != 3 {
													panic("internal error: recur arity mismatch")
												}
												remaining = __recurValues[0]
												out = __recurValues[1]
												first_q = __recurValues[2]
												continue
											}
											return __loopResult
										}
									}()
								} else {
									var if_result_786 flagrt.Value
									if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_index))) {
										if_result_786 = flagrt.NewString(flagrt.Str(compiler_ir__render_ir_node_arity_3(flagKw_expr, flagrt.Call(flagKw_x, node), flagStr_), "[", compiler_ir__render_ir_node_arity_3(flagKw_expr, flagrt.Call(flagKw_index, node), flagStr_), "]"))
									} else {
										var if_result_785 flagrt.Value
										if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_slice))) {
											var let_result_777 string
											{
												var if_result_775 flagrt.Value
												if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsNil(flagrt.Call(flagKw_low, node)))) {
													if_result_775 = flagStr_
												} else {
													if_result_775 = compiler_ir__render_ir_node_arity_3(flagKw_expr, flagrt.Call(flagKw_low, node), flagStr_)
												}
												var low = if_result_775
												var if_result_776 flagrt.Value
												if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsNil(flagrt.Call(flagKw_high, node)))) {
													if_result_776 = flagStr_
												} else {
													if_result_776 = compiler_ir__render_ir_node_arity_3(flagKw_expr, flagrt.Call(flagKw_high, node), flagStr_)
												}
												var high = if_result_776
												let_result_777 = flagrt.Str(compiler_ir__render_ir_node_arity_3(flagKw_expr, flagrt.Call(flagKw_x, node), flagStr_), "[", low, ":", high, "]")
											}
											if_result_785 = flagrt.NewString(let_result_777)
										} else {
											var if_result_784 flagrt.Value
											if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_unary))) {
												var if_result_778 string
												if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagrt.NewString("!"), flagrt.Call(flagKw_op, node)))) {
													if_result_778 = flagrt.Str("!(", compiler_ir__render_ir_node_arity_3(flagKw_expr, flagrt.Call(flagKw_x, node), flagStr_), ")")
												} else {
													if_result_778 = flagrt.Str(flagrt.Call(flagKw_op, node), compiler_ir__render_ir_node_arity_3(flagKw_expr, flagrt.Call(flagKw_x, node), flagStr_))
												}
												if_result_784 = flagrt.NewString(if_result_778)
											} else {
												var if_result_783 flagrt.Value
												if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_binary))) {
													if_result_783 = flagrt.NewString(flagrt.Str(compiler_ir__render_ir_node_arity_3(flagKw_expr, flagrt.Call(flagKw_left, node), flagStr_), " ", flagrt.Call(flagKw_op, node), " ", compiler_ir__render_ir_node_arity_3(flagKw_expr, flagrt.Call(flagKw_right, node), flagStr_)))
												} else {
													var if_result_782 flagrt.Value
													if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_func_lit))) {
														var let_result_780 string
														{
															var result = flagrt.Call(flagKw_result, node)
															var if_result_779 string
															if flagrt.IsTruthy(flagrt.Call(compiler_ir__blank_q, result)) {
																if_result_779 = flagrt.Str("func(", flagrt.Call(flagKw_params, node), ")")
															} else {
																if_result_779 = flagrt.Str("func(", flagrt.Call(flagKw_params, node), ") ", result)
															}
															var head = if_result_779
															let_result_780 = flagrt.Str(head, " {\n", compiler_ir__render_ir_node_arity_3(flagKw_stmts, flagrt.Call(flagKw_body, node), flagStr___2), "}")
														}
														if_result_782 = flagrt.NewString(let_result_780)
													} else {
														var if_result_781 flagrt.Value
														if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_raw))) {
															if_result_781 = flagrt.Call(flagKw_code, node)
														} else {
															if_result_781 = func() flagrt.Value {
																flagrt.Throw(flagrt.NewMap(flagKw_message, flagrt.NewString("unsupported IR expression"), flagKw_data, flagrt.NewMap(flagKw_node, node)))
																return flagrt.NilValue()
															}()
														}
														if_result_782 = if_result_781
													}
													if_result_783 = if_result_782
												}
												if_result_784 = if_result_783
											}
											if_result_785 = if_result_784
										}
										if_result_786 = if_result_785
									}
									if_result_787 = if_result_786
								}
								if_result_788 = if_result_787
							}
							if_result_789 = if_result_788
						}
						if_result_790 = if_result_789
					}
					if_result_791 = if_result_790
				}
				let_result_792 = if_result_791
			}
			if_result_818 = let_result_792
		} else {
			var if_result_817 flagrt.Value
			if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(mode, flagKw_stmt))) {
				var let_result_816 flagrt.Value
				{
					var kind = flagrt.Call(flagKw_kind, node)
					var if_result_815 flagrt.Value
					if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_raw_stmt))) {
						if_result_815 = flagrt.Call(flagKw_code, node)
					} else {
						var if_result_814 flagrt.Value
						if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_expr_stmt))) {
							var if_result_793 string
							if flagrt.IsTruthy(flagrt.Call(flagKw_discard, node)) {
								if_result_793 = flagrt.Str(indent, "_ = ", compiler_ir__render_ir_node_arity_3(flagKw_expr, flagrt.Call(flagKw_expr, node), flagStr_), "\n")
							} else {
								if_result_793 = flagrt.Str(indent, compiler_ir__render_ir_node_arity_3(flagKw_expr, flagrt.Call(flagKw_expr, node), flagStr_), "\n")
							}
							if_result_814 = flagrt.NewString(if_result_793)
						} else {
							var if_result_813 flagrt.Value
							if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_return))) {
								var if_result_794 string
								if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsNil(flagrt.Call(flagKw_expr, node)))) {
									if_result_794 = flagrt.Str(indent, "return\n")
								} else {
									if_result_794 = flagrt.Str(indent, "return ", compiler_ir__render_ir_node_arity_3(flagKw_expr, flagrt.Call(flagKw_expr, node), flagStr_), "\n")
								}
								if_result_813 = flagrt.NewString(if_result_794)
							} else {
								var if_result_812 flagrt.Value
								if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_defer))) {
									if_result_812 = flagrt.NewString(flagrt.Str(indent, "defer ", compiler_ir__render_ir_node_arity_3(flagKw_expr, flagrt.Call(flagKw_expr, node), flagStr_), "\n"))
								} else {
									var if_result_811 flagrt.Value
									if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_go))) {
										if_result_811 = flagrt.NewString(flagrt.Str(indent, "go ", compiler_ir__render_ir_node_arity_3(flagKw_expr, flagrt.Call(flagKw_expr, node), flagStr_), "\n"))
									} else {
										var if_result_810 flagrt.Value
										if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_var))) {
											var let_result_801 string
											{
												var name = flagrt.Call(flagKw_name, node)
												var type_ = flagrt.Call(flagKw_type, node)
												var expr = flagrt.Call(flagKw_expr, node)
												var let_result_797 flagrt.Value
												{
													var if_result_795 flagrt.Value
													if flagrt.IsTruthy(flagrt.Call(compiler_ir__blank_q, type_)) {
														if_result_795 = flagrt.NewBool(false)
													} else {
														if_result_795 = flagrt.NewBool(true)
													}
													var and_tmp = if_result_795
													var if_result_796 flagrt.Value
													if flagrt.IsTruthy(and_tmp) {
														if_result_796 = flagrt.NewBool(flagrt.IsNil(expr))
													} else {
														if_result_796 = and_tmp
													}
													let_result_797 = if_result_796
												}
												var if_result_800 string
												if flagrt.IsTruthy(let_result_797) {
													if_result_800 = flagrt.Str(indent, "var ", name, " ", type_, "\n")
												} else {
													var if_result_798 flagrt.Value
													if flagrt.IsTruthy(flagrt.Call(compiler_ir__blank_q, type_)) {
														if_result_798 = flagrt.NewBool(false)
													} else {
														if_result_798 = flagrt.NewBool(true)
													}
													var if_result_799 string
													if flagrt.IsTruthy(if_result_798) {
														if_result_799 = flagrt.Str(indent, "var ", name, " ", type_, " = ", compiler_ir__render_ir_node_arity_3(flagKw_expr, expr, flagStr_), "\n")
													} else {
														if_result_799 = flagrt.Str(indent, "var ", name, " = ", compiler_ir__render_ir_node_arity_3(flagKw_expr, expr, flagStr_), "\n")
													}
													if_result_800 = if_result_799
												}
												let_result_801 = if_result_800
											}
											if_result_810 = flagrt.NewString(let_result_801)
										} else {
											var if_result_809 flagrt.Value
											if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_assign))) {
												if_result_809 = flagrt.NewString(flagrt.Str(indent, flagrt.Call(flagKw_name, node), " = ", compiler_ir__render_ir_node_arity_3(flagKw_expr, flagrt.Call(flagKw_expr, node), flagStr_), "\n"))
											} else {
												var if_result_808 flagrt.Value
												if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_define))) {
													if_result_808 = flagrt.NewString(flagrt.Str(indent, flagrt.Call(compiler_ir__join_names, flagrt.Call(flagKw_names, node)), " := ", compiler_ir__render_ir_node_arity_3(flagKw_expr, flagrt.Call(flagKw_expr, node), flagStr_), "\n"))
												} else {
													var if_result_807 flagrt.Value
													if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_if))) {
														var let_result_804 string
														{
															var init = flagrt.Call(flagKw_init, node)
															var if_result_802 string
															if flagrt.IsTruthy(flagrt.Call(compiler_ir__blank_q, init)) {
																if_result_802 = flagrt.Str(indent, "if ", compiler_ir__render_ir_node_arity_3(flagKw_expr, flagrt.Call(flagKw_cond, node), flagStr_), " {\n")
															} else {
																if_result_802 = flagrt.Str(indent, "if ", init, "; ", compiler_ir__render_ir_node_arity_3(flagKw_expr, flagrt.Call(flagKw_cond, node), flagStr_), " {\n")
															}
															var head = if_result_802
															var then_block = compiler_ir__render_ir_node_arity_3(flagKw_stmts, flagrt.Call(flagKw_then, node), flagrt.NewString(flagrt.Str(indent, "\t")))
															var if_result_803 string
															if flagrt.IsTruthy(flagrt.Call(compiler_ir__empty_seq_q, flagrt.Call(flagKw_else, node))) {
																if_result_803 = ""
															} else {
																if_result_803 = flagrt.Str(indent, "} else {\n", compiler_ir__render_ir_node_arity_3(flagKw_stmts, flagrt.Call(flagKw_else, node), flagrt.NewString(flagrt.Str(indent, "\t"))))
															}
															var else_part = if_result_803
															let_result_804 = flagrt.Str(head, then_block, else_part, indent, "}\n")
														}
														if_result_807 = flagrt.NewString(let_result_804)
													} else {
														var if_result_806 flagrt.Value
														if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_for))) {
															if_result_806 = flagrt.NewString(flagrt.Str(indent, "for {\n", compiler_ir__render_ir_node_arity_3(flagKw_stmts, flagrt.Call(flagKw_body, node), flagrt.NewString(flagrt.Str(indent, "\t"))), indent, "}\n"))
														} else {
															var if_result_805 flagrt.Value
															if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_block))) {
																if_result_805 = flagrt.NewString(flagrt.Str(indent, "{\n", compiler_ir__render_ir_node_arity_3(flagKw_stmts, flagrt.Call(flagKw_body, node), flagrt.NewString(flagrt.Str(indent, "\t"))), indent, "}\n"))
															} else {
																if_result_805 = func() flagrt.Value {
																	flagrt.Throw(flagrt.NewMap(flagKw_message, flagrt.NewString("unsupported IR statement"), flagKw_data, flagrt.NewMap(flagKw_node, node)))
																	return flagrt.NilValue()
																}()
															}
															if_result_806 = if_result_805
														}
														if_result_807 = if_result_806
													}
													if_result_808 = if_result_807
												}
												if_result_809 = if_result_808
											}
											if_result_810 = if_result_809
										}
										if_result_811 = if_result_810
									}
									if_result_812 = if_result_811
								}
								if_result_813 = if_result_812
							}
							if_result_814 = if_result_813
						}
						if_result_815 = if_result_814
					}
					let_result_816 = if_result_815
				}
				if_result_817 = let_result_816
			} else {
				if_result_817 = func() flagrt.Value {
					flagrt.Throw(flagrt.NewMap(flagKw_message, flagrt.NewString("unsupported IR render mode"), flagKw_data, flagrt.NewMap(flagKw_mode, mode)))
					return flagrt.NilValue()
				}()
			}
			if_result_818 = if_result_817
		}
		if_result_819 = if_result_818
	}
	return if_result_819
}

func compiler_ir__render_ir_node_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 3 {
		panic("compiler_ir__render_ir_node expects exactly 3 arguments")
	}
	return compiler_ir__render_ir_node_arity_3(args[0], args[1], args[2])
}

func compiler_ir__render_ir_arity_1(node flagrt.Value) flagrt.Value {
	return flagrt.Call(compiler_ir__render_ir_node, flagKw_expr, node, flagStr_)
}

func compiler_ir__render_ir_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler_ir__render_ir expects exactly 1 arguments")
	}
	return compiler_ir__render_ir_arity_1(args[0])
}

func compiler_ir__render_ir_stmt_arity_2(node flagrt.Value, indent flagrt.Value) flagrt.Value {
	return flagrt.Call(compiler_ir__render_ir_node, flagKw_stmt, node, indent)
}

func compiler_ir__render_ir_stmt_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 2 {
		panic("compiler_ir__render_ir_stmt expects exactly 2 arguments")
	}
	return compiler_ir__render_ir_stmt_arity_2(args[0], args[1])
}

func compiler_ir__render_ir_stmts_arity_2(stmts flagrt.Value, indent flagrt.Value) flagrt.Value {
	return flagrt.Call(compiler_ir__render_ir_node, flagKw_stmts, stmts, indent)
}

func compiler_ir__render_ir_stmts_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 2 {
		panic("compiler_ir__render_ir_stmts expects exactly 2 arguments")
	}
	return compiler_ir__render_ir_stmts_arity_2(args[0], args[1])
}

func compiler_codegen__symbol_node_q_arity_1(expr flagrt.Value) flagrt.Value {
	return flagrt.NewBool(flagrt.Eq(flagKw_symbol, flagrt.Call(flagKw_kind, expr)))
}

func compiler_codegen__symbol_node_q_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler_codegen__symbol_node_q expects exactly 1 arguments")
	}
	return compiler_codegen__symbol_node_q_arity_1(args[0])
}

func compiler_codegen__list_node_q_arity_1(expr flagrt.Value) flagrt.Value {
	return flagrt.NewBool(flagrt.Eq(flagKw_list, flagrt.Call(flagKw_kind, expr)))
}

func compiler_codegen__list_node_q_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler_codegen__list_node_q expects exactly 1 arguments")
	}
	return compiler_codegen__list_node_q_arity_1(args[0])
}

func compiler_codegen__emit_ir_expr_arity_1(expr flagrt.Value) flagrt.Value {
	var if_result_828 flagrt.Value
	if flagrt.IsTruthy(flagrt.Call(compiler_codegen__symbol_node_q, expr)) {
		if_result_828 = flagrt.Call(compiler_ir__ir_ident, flagrt.Call(flagKw_name, expr))
	} else {
		var if_result_827 flagrt.Value
		if flagrt.IsTruthy(flagrt.Call(compiler_codegen__list_node_q, expr)) {
			var let_result_826 flagrt.Value
			{
				var children = flagrt.Call(flagKw_elements, expr)
				var if_result_820 flagrt.Value
				if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagrt.NewLong(3), flagrt.NewLong(int64(flagrt.Count(children)))))) {
					if_result_820 = flagrt.NewBool(false)
				} else {
					if_result_820 = flagrt.NewBool(true)
				}
				var if_result_825 flagrt.Value
				if flagrt.IsTruthy(if_result_820) {
					if_result_825 = func() flagrt.Value {
						flagrt.Throw(flagrt.NewMap(flagKw_message, flagrt.NewString("unsupported list shape for codegen"), flagKw_data, flagrt.NewMap(flagKw_expr, expr)))
						return flagrt.NilValue()
					}()
				} else {
					var let_result_824 flagrt.Value
					{
						var op = flagrt.First(children)
						var lhs = flagrt.Call(stdlib__second, children)
						var rhs = flagrt.Call(stdlib__third, children)
						var let_result_822 flagrt.Value
						{
							var and_tmp = flagrt.Call(compiler_codegen__symbol_node_q, op)
							var if_result_821 flagrt.Value
							if flagrt.IsTruthy(and_tmp) {
								if_result_821 = flagrt.NewBool(flagrt.Contains(compiler_codegen__binary_ops, flagrt.Call(flagKw_name, op)))
							} else {
								if_result_821 = and_tmp
							}
							let_result_822 = if_result_821
						}
						var if_result_823 flagrt.Value
						if flagrt.IsTruthy(let_result_822) {
							if_result_823 = flagrt.Call(compiler_ir__ir_binary, flagrt.Call(flagKw_name, op), compiler_codegen__emit_ir_expr_arity_1(lhs), compiler_codegen__emit_ir_expr_arity_1(rhs))
						} else {
							if_result_823 = func() flagrt.Value {
								flagrt.Throw(flagrt.NewMap(flagKw_message, flagrt.NewString("unsupported list operator for codegen"), flagKw_data, flagrt.NewMap(flagKw_expr, expr)))
								return flagrt.NilValue()
							}()
						}
						let_result_824 = if_result_823
					}
					if_result_825 = let_result_824
				}
				let_result_826 = if_result_825
			}
			if_result_827 = let_result_826
		} else {
			if_result_827 = func() flagrt.Value {
				flagrt.Throw(flagrt.NewMap(flagKw_message, flagrt.NewString("unsupported AST node for codegen"), flagKw_data, flagrt.NewMap(flagKw_expr, expr)))
				return flagrt.NilValue()
			}()
		}
		if_result_828 = if_result_827
	}
	return if_result_828
}

func compiler_codegen__emit_ir_expr_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler_codegen__emit_ir_expr expects exactly 1 arguments")
	}
	return compiler_codegen__emit_ir_expr_arity_1(args[0])
}

func compiler_codegen__emit_go_expr_arity_1(expr flagrt.Value) flagrt.Value {
	var let_result_830 flagrt.Value
	{
		var node = flagrt.Call(compiler_codegen__emit_ir_expr, expr)
		var if_result_829 flagrt.Value
		if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagKw_binary, flagrt.Call(flagKw_kind, node)))) {
			if_result_829 = flagrt.NewString(flagrt.Str("(", flagrt.Call(compiler_ir__render_ir, node), ")"))
		} else {
			if_result_829 = flagrt.Call(compiler_ir__render_ir, node)
		}
		let_result_830 = if_result_829
	}
	return let_result_830
}

func compiler_codegen__emit_go_expr_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler_codegen__emit_go_expr expects exactly 1 arguments")
	}
	return compiler_codegen__emit_go_expr_arity_1(args[0])
}

func compiler_codegen__emit_go_loop_bang_arity_2(in flagrt.Value, out flagrt.Value) flagrt.Value {
	var let_result_832 flagrt.Value
	{
		var expr = flagrt.Call(async__channel_receive, in)
		var if_result_831 flagrt.Value
		if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsNil(expr))) {
			if_result_831 = flagrt.Call(async__channel_close, out)
		} else {
			_ = flagrt.Call(async__channel_send, out, flagrt.Call(compiler_codegen__emit_go_expr, expr))
			if_result_831 = compiler_codegen__emit_go_loop_bang_arity_2(in, out)
		}
		let_result_832 = if_result_831
	}
	return let_result_832
}

func compiler_codegen__emit_go_loop_bang_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 2 {
		panic("compiler_codegen__emit_go_loop_bang expects exactly 2 arguments")
	}
	return compiler_codegen__emit_go_loop_bang_arity_2(args[0], args[1])
}

func compiler_codegen__emit_go_channel_arity_1(in flagrt.Value) flagrt.Value {
	var let_result_833 flagrt.Value
	{
		var out = flagrt.Call(async__make_channel, flagrt.NewLong(64))
		_ = flagrt.Call(async__go_run, flagrt.NewFunction(func(args ...flagrt.Value) flagrt.Value {
			if len(args) != 0 {
				panic("fn expects exactly 0 arguments")
			}
			return flagrt.Call(compiler_codegen__emit_go_loop_bang, in, out)
		}))
		let_result_833 = out
	}
	return let_result_833
}

func compiler_codegen__emit_go_channel_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler_codegen__emit_go_channel expects exactly 1 arguments")
	}
	return compiler_codegen__emit_go_channel_arity_1(args[0])
}

func compiler_codegen__compile_file_arity_1(path flagrt.Value) flagrt.Value {
	return flagrt.Call(compiler_codegen__emit_go_channel, flagrt.Call(compiler_lift__lift_file, path))
}

func compiler_codegen__compile_file_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler_codegen__compile_file expects exactly 1 arguments")
	}
	return compiler_codegen__compile_file_arity_1(args[0])
}

func compiler__literal_rt_call_arity_2(name flagrt.Value, args flagrt.Value) flagrt.Value {
	return flagrt.NewMap(flagKw_kind, flagKw_call, flagKw_fun, flagrt.NewMap(flagKw_kind, flagKw_selector, flagKw_pkg, flagrt.NewString("flagrt"), flagKw_name, name), flagKw_args, args)
}

func compiler__literal_rt_call_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 2 {
		panic("compiler__literal_rt_call expects exactly 2 arguments")
	}
	return compiler__literal_rt_call_arity_2(args[0], args[1])
}

func compiler__ast_node_to_ir_arity_1(node flagrt.Value) flagrt.Value {
	var let_result_851 flagrt.Value
	{
		var kind = flagrt.Call(flagKw_kind, node)
		var if_result_850 flagrt.Value
		if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_string))) {
			if_result_850 = flagrt.NewMap(flagKw_kind, flagKw_string, flagKw_value, flagrt.Call(flagKw_value, node))
		} else {
			var if_result_849 flagrt.Value
			if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_char))) {
				if_result_849 = flagrt.Call(compiler__literal_rt_call, flagStr_NewString, flagrt.NewArray(flagrt.NewMap(flagKw_kind, flagKw_string, flagKw_value, flagrt.Call(flagKw_value, node))))
			} else {
				var if_result_848 flagrt.Value
				if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_int))) {
					if_result_848 = flagrt.Call(compiler__literal_rt_call, flagStr_NewLong, flagrt.NewArray(flagrt.NewMap(flagKw_kind, flagKw_int, flagKw_value, flagrt.Call(flagKw_value, node))))
				} else {
					var if_result_847 flagrt.Value
					if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_bigint))) {
						if_result_847 = flagrt.Call(compiler__literal_rt_call, flagStr_NewBigIntFromString, flagrt.NewArray(flagrt.NewMap(flagKw_kind, flagKw_string, flagKw_value, flagrt.Call(flagKw_value, node))))
					} else {
						var if_result_846 flagrt.Value
						if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_ratio))) {
							if_result_846 = flagrt.Call(compiler__literal_rt_call, flagStr_NewRatio, flagrt.NewArray(flagrt.NewMap(flagKw_kind, flagKw_int, flagKw_value, flagrt.Call(flagKw_numerator, node)), flagrt.NewMap(flagKw_kind, flagKw_int, flagKw_value, flagrt.Call(flagKw_denominator, node))))
						} else {
							var if_result_845 flagrt.Value
							if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_float))) {
								var let_result_837 flagrt.Value
								{
									var raw = flagrt.Call(flagKw_raw, node)
									var let_result_835 flagrt.Value
									{
										var or_tmp = flagrt.NewBool(flagrt.IsNil(raw))
										var if_result_834 flagrt.Value
										if flagrt.IsTruthy(or_tmp) {
											if_result_834 = or_tmp
										} else {
											if_result_834 = flagrt.NewBool(flagrt.Eq(flagrt.NewString(""), raw))
										}
										let_result_835 = if_result_834
									}
									var if_result_836 flagrt.Value
									if flagrt.IsTruthy(let_result_835) {
										if_result_836 = flagrt.NewString(flagrt.Format("%g", flagrt.Call(flagKw_value, node)))
									} else {
										if_result_836 = raw
									}
									var code = if_result_836
									let_result_837 = flagrt.Call(compiler__literal_rt_call, flagStr_NewDouble, flagrt.NewArray(flagrt.NewMap(flagKw_kind, flagKw_raw, flagKw_code, code)))
								}
								if_result_845 = let_result_837
							} else {
								var if_result_844 flagrt.Value
								if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_keyword))) {
									if_result_844 = flagrt.Call(compiler__literal_rt_call, flagStr_NewKeyword, flagrt.NewArray(flagrt.NewMap(flagKw_kind, flagKw_string, flagKw_value, flagrt.Call(flagKw_name, node))))
								} else {
									var if_result_843 flagrt.Value
									if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_quoted_symbol))) {
										if_result_843 = flagrt.Call(compiler__literal_rt_call, flagStr_NewSymbol, flagrt.NewArray(flagrt.NewMap(flagKw_kind, flagKw_string, flagKw_value, flagrt.Call(flagKw_name, node))))
									} else {
										var if_result_842 flagrt.Value
										if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_symbol))) {
											var let_result_841 flagrt.Value
											{
												var name = flagrt.Call(flagKw_name, node)
												var if_result_840 flagrt.Value
												if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(name, flagrt.NewString("true")))) {
													if_result_840 = flagrt.Call(compiler__literal_rt_call, flagStr_NewBool, flagVec_2)
												} else {
													var if_result_839 flagrt.Value
													if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(name, flagrt.NewString("false")))) {
														if_result_839 = flagrt.Call(compiler__literal_rt_call, flagStr_NewBool, flagVec_3)
													} else {
														var if_result_838 flagrt.Value
														if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(name, flagrt.NewString("nil")))) {
															if_result_838 = flagrt.Call(compiler__literal_rt_call, flagStr_NilValue, flagVec)
														} else {
															if_result_838 = func() flagrt.Value {
																flagrt.Throw(flagrt.NewMap(flagKw_message, flagrt.NewString("unsupported symbol for literal lowering"), flagKw_data, flagrt.NewMap(flagKw_node, node)))
																return flagrt.NilValue()
															}()
														}
														if_result_839 = if_result_838
													}
													if_result_840 = if_result_839
												}
												let_result_841 = if_result_840
											}
											if_result_842 = let_result_841
										} else {
											if_result_842 = func() flagrt.Value {
												flagrt.Throw(flagrt.NewMap(flagKw_message, flagrt.NewString("unsupported AST node for literal lowering"), flagKw_data, flagrt.NewMap(flagKw_node, node)))
												return flagrt.NilValue()
											}()
										}
										if_result_843 = if_result_842
									}
									if_result_844 = if_result_843
								}
								if_result_845 = if_result_844
							}
							if_result_846 = if_result_845
						}
						if_result_847 = if_result_846
					}
					if_result_848 = if_result_847
				}
				if_result_849 = if_result_848
			}
			if_result_850 = if_result_849
		}
		let_result_851 = if_result_850
	}
	return let_result_851
}

func compiler__ast_node_to_ir_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__ast_node_to_ir expects exactly 1 arguments")
	}
	return compiler__ast_node_to_ir_arity_1(args[0])
}

func compiler__quoted_ast_to_ir_node_arity_2(mode flagrt.Value, node flagrt.Value) flagrt.Value {
	var if_result_874 flagrt.Value
	if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(mode, flagKw_seq))) {
		if_result_874 = func() flagrt.Value {
			var remaining = node
			var out = flagVec
			for {
				var let_result_853 flagrt.Value
				{
					var or_tmp = flagrt.NewBool(flagrt.IsNil(remaining))
					var if_result_852 flagrt.Value
					if flagrt.IsTruthy(or_tmp) {
						if_result_852 = or_tmp
					} else {
						if_result_852 = flagrt.NewBool(flagrt.IsEmpty(remaining))
					}
					let_result_853 = if_result_852
				}
				var if_result_854 flagrt.Value
				if flagrt.IsTruthy(let_result_853) {
					if_result_854 = out
				} else {
					if_result_854 = flagrt.NewRecur(flagrt.Rest(remaining), flagrt.Conj(out, compiler__quoted_ast_to_ir_node_arity_2(flagKw_node, flagrt.First(remaining))))
				}
				__loopResult := if_result_854
				if __recurValues, __isRecur := flagrt.UnwrapRecur(__loopResult); __isRecur {
					if len(__recurValues) != 2 {
						panic("internal error: recur arity mismatch")
					}
					remaining = __recurValues[0]
					out = __recurValues[1]
					continue
				}
				return __loopResult
			}
		}()
	} else {
		var if_result_873 flagrt.Value
		if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(mode, flagKw_node))) {
			var let_result_872 flagrt.Value
			{
				var kind = flagrt.Call(flagKw_kind, node)
				var if_result_871 flagrt.Value
				if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_int))) {
					if_result_871 = flagrt.Call(compiler__ast_node_to_ir, node)
				} else {
					var if_result_870 flagrt.Value
					if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_bigint))) {
						if_result_870 = flagrt.Call(compiler__ast_node_to_ir, node)
					} else {
						var if_result_869 flagrt.Value
						if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_ratio))) {
							if_result_869 = flagrt.Call(compiler__ast_node_to_ir, node)
						} else {
							var if_result_868 flagrt.Value
							if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_float))) {
								if_result_868 = flagrt.Call(compiler__ast_node_to_ir, node)
							} else {
								var if_result_867 flagrt.Value
								if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_keyword))) {
									if_result_867 = flagrt.Call(compiler__ast_node_to_ir, node)
								} else {
									var if_result_866 flagrt.Value
									if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_quoted_symbol))) {
										if_result_866 = flagrt.Call(compiler__ast_node_to_ir, node)
									} else {
										var if_result_865 flagrt.Value
										if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_symbol))) {
											if_result_865 = flagrt.Call(compiler__literal_rt_call, flagStr_NewSymbol, flagrt.NewArray(flagrt.NewMap(flagKw_kind, flagKw_string, flagKw_value, flagrt.Call(flagKw_name, node))))
										} else {
											var if_result_864 flagrt.Value
											if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_quoted_list))) {
												if_result_864 = flagrt.Call(compiler__literal_rt_call, flagStr_NewList, compiler__quoted_ast_to_ir_node_arity_2(flagKw_seq, flagrt.Call(flagKw_elements, node)))
											} else {
												var if_result_863 flagrt.Value
												if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_list))) {
													if_result_863 = flagrt.Call(compiler__literal_rt_call, flagStr_NewList, compiler__quoted_ast_to_ir_node_arity_2(flagKw_seq, flagrt.Call(flagKw_elements, node)))
												} else {
													var if_result_862 flagrt.Value
													if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_vector))) {
														if_result_862 = flagrt.Call(compiler__literal_rt_call, flagStr_NewArray, compiler__quoted_ast_to_ir_node_arity_2(flagKw_seq, flagrt.Call(flagKw_elements, node)))
													} else {
														var if_result_861 flagrt.Value
														if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_pipe_vector))) {
															if_result_861 = flagrt.Call(compiler__literal_rt_call, flagStr_NewVector, compiler__quoted_ast_to_ir_node_arity_2(flagKw_seq, flagrt.Call(flagKw_elements, node)))
														} else {
															var if_result_860 flagrt.Value
															if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_map))) {
																var let_result_858 flagrt.Value
																{
																	var entries = flagrt.Call(flagKw_entries, node)
																	var if_result_855 flagrt.Value
																	if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsNil(entries))) {
																		if_result_855 = flagrt.NewLong(0)
																	} else {
																		if_result_855 = flagrt.NewLong(int64(flagrt.Count(entries)))
																	}
																	var n = if_result_855
																	var if_result_856 flagrt.Value
																	if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagrt.NewLong(0), flagrt.Mod(n, flagrt.NewLong(2))))) {
																		if_result_856 = flagrt.NewBool(false)
																	} else {
																		if_result_856 = flagrt.NewBool(true)
																	}
																	var if_result_857 flagrt.Value
																	if flagrt.IsTruthy(if_result_856) {
																		if_result_857 = func() flagrt.Value {
																			flagrt.Throw(flagrt.NewMap(flagKw_message, flagrt.NewString("quoted map literal expects an even number of forms"), flagKw_data, flagrt.NewMap(flagKw_node, node)))
																			return flagrt.NilValue()
																		}()
																	} else {
																		if_result_857 = flagrt.Call(compiler__literal_rt_call, flagStr_NewMap, compiler__quoted_ast_to_ir_node_arity_2(flagKw_seq, entries))
																	}
																	let_result_858 = if_result_857
																}
																if_result_860 = let_result_858
															} else {
																var if_result_859 flagrt.Value
																if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_set))) {
																	if_result_859 = flagrt.Call(compiler__literal_rt_call, flagStr_NewSet, compiler__quoted_ast_to_ir_node_arity_2(flagKw_seq, flagrt.Call(flagKw_elements, node)))
																} else {
																	if_result_859 = func() flagrt.Value {
																		flagrt.Throw(flagrt.NewMap(flagKw_message, flagrt.NewString("unsupported quoted literal"), flagKw_data, flagrt.NewMap(flagKw_node, node)))
																		return flagrt.NilValue()
																	}()
																}
																if_result_860 = if_result_859
															}
															if_result_861 = if_result_860
														}
														if_result_862 = if_result_861
													}
													if_result_863 = if_result_862
												}
												if_result_864 = if_result_863
											}
											if_result_865 = if_result_864
										}
										if_result_866 = if_result_865
									}
									if_result_867 = if_result_866
								}
								if_result_868 = if_result_867
							}
							if_result_869 = if_result_868
						}
						if_result_870 = if_result_869
					}
					if_result_871 = if_result_870
				}
				let_result_872 = if_result_871
			}
			if_result_873 = let_result_872
		} else {
			if_result_873 = func() flagrt.Value {
				flagrt.Throw(flagrt.NewMap(flagKw_message, flagrt.NewString("unsupported quoted lowering mode"), flagKw_data, flagrt.NewMap(flagKw_mode, mode)))
				return flagrt.NilValue()
			}()
		}
		if_result_874 = if_result_873
	}
	return if_result_874
}

func compiler__quoted_ast_to_ir_node_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 2 {
		panic("compiler__quoted_ast_to_ir_node expects exactly 2 arguments")
	}
	return compiler__quoted_ast_to_ir_node_arity_2(args[0], args[1])
}

func compiler__quoted_ast_to_ir_arity_1(node flagrt.Value) flagrt.Value {
	return flagrt.Call(compiler__quoted_ast_to_ir_node, flagKw_node, node)
}

func compiler__quoted_ast_to_ir_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__quoted_ast_to_ir expects exactly 1 arguments")
	}
	return compiler__quoted_ast_to_ir_arity_1(args[0])
}

func compiler__call_to_ir_arity_2(callee flagrt.Value, args flagrt.Value) flagrt.Value {
	return func() flagrt.Value {
		var remaining = args
		var out = flagrt.NewArray(callee)
		for {
			var let_result_876 flagrt.Value
			{
				var or_tmp = flagrt.NewBool(flagrt.IsNil(remaining))
				var if_result_875 flagrt.Value
				if flagrt.IsTruthy(or_tmp) {
					if_result_875 = or_tmp
				} else {
					if_result_875 = flagrt.NewBool(flagrt.IsEmpty(remaining))
				}
				let_result_876 = if_result_875
			}
			var if_result_877 flagrt.Value
			if flagrt.IsTruthy(let_result_876) {
				if_result_877 = flagrt.Call(compiler__literal_rt_call, flagStr_Call, out)
			} else {
				if_result_877 = flagrt.NewRecur(flagrt.Rest(remaining), flagrt.Conj(out, flagrt.First(remaining)))
			}
			__loopResult := if_result_877
			if __recurValues, __isRecur := flagrt.UnwrapRecur(__loopResult); __isRecur {
				if len(__recurValues) != 2 {
					panic("internal error: recur arity mismatch")
				}
				remaining = __recurValues[0]
				out = __recurValues[1]
				continue
			}
			return __loopResult
		}
	}()
}

func compiler__call_to_ir_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 2 {
		panic("compiler__call_to_ir expects exactly 2 arguments")
	}
	return compiler__call_to_ir_arity_2(args[0], args[1])
}

func compiler__call_ast_to_ir_node_arity_2(mode flagrt.Value, node flagrt.Value) flagrt.Value {
	var if_result_913 flagrt.Value
	if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(mode, flagKw_seq))) {
		if_result_913 = func() flagrt.Value {
			var remaining = node
			var out = flagVec
			for {
				var let_result_879 flagrt.Value
				{
					var or_tmp = flagrt.NewBool(flagrt.IsNil(remaining))
					var if_result_878 flagrt.Value
					if flagrt.IsTruthy(or_tmp) {
						if_result_878 = or_tmp
					} else {
						if_result_878 = flagrt.NewBool(flagrt.IsEmpty(remaining))
					}
					let_result_879 = if_result_878
				}
				var if_result_880 flagrt.Value
				if flagrt.IsTruthy(let_result_879) {
					if_result_880 = out
				} else {
					if_result_880 = flagrt.NewRecur(flagrt.Rest(remaining), flagrt.Conj(out, compiler__call_ast_to_ir_node_arity_2(flagKw_eval, flagrt.First(remaining))))
				}
				__loopResult := if_result_880
				if __recurValues, __isRecur := flagrt.UnwrapRecur(__loopResult); __isRecur {
					if len(__recurValues) != 2 {
						panic("internal error: recur arity mismatch")
					}
					remaining = __recurValues[0]
					out = __recurValues[1]
					continue
				}
				return __loopResult
			}
		}()
	} else {
		var if_result_912 flagrt.Value
		if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(mode, flagKw_eval))) {
			var let_result_902 flagrt.Value
			{
				var kind = flagrt.Call(flagKw_kind, node)
				var if_result_901 flagrt.Value
				if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_string))) {
					if_result_901 = flagrt.Call(compiler__ast_node_to_ir, node)
				} else {
					var if_result_900 flagrt.Value
					if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_char))) {
						if_result_900 = flagrt.Call(compiler__ast_node_to_ir, node)
					} else {
						var if_result_899 flagrt.Value
						if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_int))) {
							if_result_899 = flagrt.Call(compiler__ast_node_to_ir, node)
						} else {
							var if_result_898 flagrt.Value
							if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_bigint))) {
								if_result_898 = flagrt.Call(compiler__ast_node_to_ir, node)
							} else {
								var if_result_897 flagrt.Value
								if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_ratio))) {
									if_result_897 = flagrt.Call(compiler__ast_node_to_ir, node)
								} else {
									var if_result_896 flagrt.Value
									if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_float))) {
										if_result_896 = flagrt.Call(compiler__ast_node_to_ir, node)
									} else {
										var if_result_895 flagrt.Value
										if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_keyword))) {
											if_result_895 = flagrt.Call(compiler__ast_node_to_ir, node)
										} else {
											var if_result_894 flagrt.Value
											if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_quoted_symbol))) {
												if_result_894 = flagrt.Call(compiler__ast_node_to_ir, node)
											} else {
												var if_result_893 flagrt.Value
												if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_symbol))) {
													var let_result_886 flagrt.Value
													{
														var name = flagrt.Call(flagKw_name, node)
														var let_result_884 flagrt.Value
														{
															var or_tmp = flagrt.NewBool(flagrt.Eq(name, flagrt.NewString("true")))
															var if_result_883 flagrt.Value
															if flagrt.IsTruthy(or_tmp) {
																if_result_883 = or_tmp
															} else {
																var let_result_882 flagrt.Value
																{
																	var or_tmp = flagrt.NewBool(flagrt.Eq(name, flagrt.NewString("false")))
																	var if_result_881 flagrt.Value
																	if flagrt.IsTruthy(or_tmp) {
																		if_result_881 = or_tmp
																	} else {
																		if_result_881 = flagrt.NewBool(flagrt.Eq(name, flagrt.NewString("nil")))
																	}
																	let_result_882 = if_result_881
																}
																if_result_883 = let_result_882
															}
															let_result_884 = if_result_883
														}
														var if_result_885 flagrt.Value
														if flagrt.IsTruthy(let_result_884) {
															if_result_885 = flagrt.Call(compiler__ast_node_to_ir, node)
														} else {
															if_result_885 = func() flagrt.Value {
																flagrt.Throw(flagrt.NewMap(flagKw_message, flagrt.NewString("unresolved symbol for call lowering"), flagKw_data, flagrt.NewMap(flagKw_node, node)))
																return flagrt.NilValue()
															}()
														}
														let_result_886 = if_result_885
													}
													if_result_893 = let_result_886
												} else {
													var if_result_892 flagrt.Value
													if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_quoted_list))) {
														if_result_892 = flagrt.Call(compiler__quoted_ast_to_ir, node)
													} else {
														var if_result_891 flagrt.Value
														if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_vector))) {
															if_result_891 = flagrt.Call(compiler__quoted_ast_to_ir, node)
														} else {
															var if_result_890 flagrt.Value
															if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_pipe_vector))) {
																if_result_890 = flagrt.Call(compiler__quoted_ast_to_ir, node)
															} else {
																var if_result_889 flagrt.Value
																if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_map))) {
																	if_result_889 = flagrt.Call(compiler__quoted_ast_to_ir, node)
																} else {
																	var if_result_888 flagrt.Value
																	if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_set))) {
																		if_result_888 = flagrt.Call(compiler__quoted_ast_to_ir, node)
																	} else {
																		var if_result_887 flagrt.Value
																		if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_list))) {
																			if_result_887 = compiler__call_ast_to_ir_node_arity_2(flagKw_call, node)
																		} else {
																			if_result_887 = func() flagrt.Value {
																				flagrt.Throw(flagrt.NewMap(flagKw_message, flagrt.NewString("unsupported AST node for call lowering"), flagKw_data, flagrt.NewMap(flagKw_node, node)))
																				return flagrt.NilValue()
																			}()
																		}
																		if_result_888 = if_result_887
																	}
																	if_result_889 = if_result_888
																}
																if_result_890 = if_result_889
															}
															if_result_891 = if_result_890
														}
														if_result_892 = if_result_891
													}
													if_result_893 = if_result_892
												}
												if_result_894 = if_result_893
											}
											if_result_895 = if_result_894
										}
										if_result_896 = if_result_895
									}
									if_result_897 = if_result_896
								}
								if_result_898 = if_result_897
							}
							if_result_899 = if_result_898
						}
						if_result_900 = if_result_899
					}
					if_result_901 = if_result_900
				}
				let_result_902 = if_result_901
			}
			if_result_912 = let_result_902
		} else {
			var if_result_911 flagrt.Value
			if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(mode, flagKw_call))) {
				var let_result_910 flagrt.Value
				{
					var elements = flagrt.Call(flagKw_elements, node)
					var let_result_904 flagrt.Value
					{
						var or_tmp = flagrt.NewBool(flagrt.IsNil(elements))
						var if_result_903 flagrt.Value
						if flagrt.IsTruthy(or_tmp) {
							if_result_903 = or_tmp
						} else {
							if_result_903 = flagrt.NewBool(flagrt.IsEmpty(elements))
						}
						let_result_904 = if_result_903
					}
					var if_result_909 flagrt.Value
					if flagrt.IsTruthy(let_result_904) {
						if_result_909 = func() flagrt.Value {
							flagrt.Throw(flagrt.NewMap(flagKw_message, flagrt.NewString("unsupported form"), flagKw_data, flagrt.NewMap(flagKw_node, node)))
							return flagrt.NilValue()
						}()
					} else {
						var let_result_908 flagrt.Value
						{
							var head = flagrt.First(elements)
							var let_result_906 flagrt.Value
							{
								var and_tmp = flagrt.NewBool(flagrt.Eq(flagKw_symbol, flagrt.Call(flagKw_kind, head)))
								var if_result_905 flagrt.Value
								if flagrt.IsTruthy(and_tmp) {
									if_result_905 = flagrt.NewBool(flagrt.Contains(compiler__special_form_names, flagrt.Call(flagKw_name, head)))
								} else {
									if_result_905 = and_tmp
								}
								let_result_906 = if_result_905
							}
							var if_result_907 flagrt.Value
							if flagrt.IsTruthy(let_result_906) {
								if_result_907 = func() flagrt.Value {
									flagrt.Throw(flagrt.NewMap(flagKw_message, flagrt.NewString("special form not handled by call lowering"), flagKw_data, flagrt.NewMap(flagKw_node, node)))
									return flagrt.NilValue()
								}()
							} else {
								if_result_907 = flagrt.Call(compiler__call_to_ir, compiler__call_ast_to_ir_node_arity_2(flagKw_eval, head), compiler__call_ast_to_ir_node_arity_2(flagKw_seq, flagrt.Rest(elements)))
							}
							let_result_908 = if_result_907
						}
						if_result_909 = let_result_908
					}
					let_result_910 = if_result_909
				}
				if_result_911 = let_result_910
			} else {
				if_result_911 = func() flagrt.Value {
					flagrt.Throw(flagrt.NewMap(flagKw_message, flagrt.NewString("unsupported call lowering mode"), flagKw_data, flagrt.NewMap(flagKw_mode, mode)))
					return flagrt.NilValue()
				}()
			}
			if_result_912 = if_result_911
		}
		if_result_913 = if_result_912
	}
	return if_result_913
}

func compiler__call_ast_to_ir_node_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 2 {
		panic("compiler__call_ast_to_ir_node expects exactly 2 arguments")
	}
	return compiler__call_ast_to_ir_node_arity_2(args[0], args[1])
}

func compiler__call_ast_to_ir_arity_1(node flagrt.Value) flagrt.Value {
	return flagrt.Call(compiler__call_ast_to_ir_node, flagKw_call, node)
}

func compiler__call_ast_to_ir_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__call_ast_to_ir expects exactly 1 arguments")
	}
	return compiler__call_ast_to_ir_arity_1(args[0])
}

func compiler__ctor_to_ir_arity_2(name flagrt.Value, args flagrt.Value) flagrt.Value {
	return flagrt.Call(compiler__literal_rt_call, name, args)
}

func compiler__ctor_to_ir_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 2 {
		panic("compiler__ctor_to_ir expects exactly 2 arguments")
	}
	return compiler__ctor_to_ir_arity_2(args[0], args[1])
}

func compiler__runtime_call_to_ir_arity_2(name flagrt.Value, args flagrt.Value) flagrt.Value {
	return flagrt.Call(compiler__literal_rt_call, name, args)
}

func compiler__runtime_call_to_ir_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 2 {
		panic("compiler__runtime_call_to_ir expects exactly 2 arguments")
	}
	return compiler__runtime_call_to_ir_arity_2(args[0], args[1])
}

func compiler__fold_call_to_ir_arity_2(name flagrt.Value, args flagrt.Value) flagrt.Value {
	var let_result_917 flagrt.Value
	{
		var or_tmp = flagrt.NewBool(flagrt.IsNil(args))
		var if_result_916 flagrt.Value
		if flagrt.IsTruthy(or_tmp) {
			if_result_916 = or_tmp
		} else {
			var let_result_915 flagrt.Value
			{
				var or_tmp = flagrt.NewBool(flagrt.IsEmpty(args))
				var if_result_914 flagrt.Value
				if flagrt.IsTruthy(or_tmp) {
					if_result_914 = or_tmp
				} else {
					if_result_914 = flagrt.NewBool(flagrt.IsEmpty(flagrt.Rest(args)))
				}
				let_result_915 = if_result_914
			}
			if_result_916 = let_result_915
		}
		let_result_917 = if_result_916
	}
	var if_result_921 flagrt.Value
	if flagrt.IsTruthy(let_result_917) {
		if_result_921 = func() flagrt.Value {
			flagrt.Throw(flagrt.NewMap(flagKw_message, flagrt.NewString("fold expects at least two arguments"), flagKw_data, flagrt.NewMap(flagKw_name, name)))
			return flagrt.NilValue()
		}()
	} else {
		if_result_921 = func() flagrt.Value {
			var acc = flagrt.First(args)
			var remaining = flagrt.Rest(args)
			for {
				var let_result_919 flagrt.Value
				{
					var or_tmp = flagrt.NewBool(flagrt.IsNil(remaining))
					var if_result_918 flagrt.Value
					if flagrt.IsTruthy(or_tmp) {
						if_result_918 = or_tmp
					} else {
						if_result_918 = flagrt.NewBool(flagrt.IsEmpty(remaining))
					}
					let_result_919 = if_result_918
				}
				var if_result_920 flagrt.Value
				if flagrt.IsTruthy(let_result_919) {
					if_result_920 = acc
				} else {
					if_result_920 = flagrt.NewRecur(flagrt.Call(compiler__literal_rt_call, name, flagrt.NewArray(acc, flagrt.First(remaining))), flagrt.Rest(remaining))
				}
				__loopResult := if_result_920
				if __recurValues, __isRecur := flagrt.UnwrapRecur(__loopResult); __isRecur {
					if len(__recurValues) != 2 {
						panic("internal error: recur arity mismatch")
					}
					acc = __recurValues[0]
					remaining = __recurValues[1]
					continue
				}
				return __loopResult
			}
		}()
	}
	return if_result_921
}

func compiler__fold_call_to_ir_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 2 {
		panic("compiler__fold_call_to_ir expects exactly 2 arguments")
	}
	return compiler__fold_call_to_ir_arity_2(args[0], args[1])
}

func compiler__qualified_name_q_arity_1(name flagrt.Value) flagrt.Value {
	return func() flagrt.Value {
		var s = name
		for {
			var let_result_923 flagrt.Value
			{
				var or_tmp = flagrt.NewBool(flagrt.IsNil(s))
				var if_result_922 flagrt.Value
				if flagrt.IsTruthy(or_tmp) {
					if_result_922 = or_tmp
				} else {
					if_result_922 = flagrt.NewBool(flagrt.Eq(flagrt.NewLong(0), flagrt.NewLong(int64(flagrt.Count(s)))))
				}
				let_result_923 = if_result_922
			}
			var if_result_925 flagrt.Value
			if flagrt.IsTruthy(let_result_923) {
				if_result_925 = flagrt.NewBool(false)
			} else {
				var if_result_924 flagrt.Value
				if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagrt.NewString("/"), flagrt.Call(flagrt.BuiltinFunction("subs"), s, flagrt.NewLong(0), flagrt.NewLong(1))))) {
					if_result_924 = flagrt.NewBool(true)
				} else {
					if_result_924 = flagrt.NewRecur(flagrt.Call(flagrt.BuiltinFunction("subs"), s, flagrt.NewLong(1)))
				}
				if_result_925 = if_result_924
			}
			__loopResult := if_result_925
			if __recurValues, __isRecur := flagrt.UnwrapRecur(__loopResult); __isRecur {
				if len(__recurValues) != 1 {
					panic("internal error: recur arity mismatch")
				}
				s = __recurValues[0]
				continue
			}
			return __loopResult
		}
	}()
}

func compiler__qualified_name_q_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__qualified_name_q expects exactly 1 arguments")
	}
	return compiler__qualified_name_q_arity_1(args[0])
}

func compiler__ident_ir_arity_2(ident flagrt.Value, k flagrt.Value) flagrt.Value {
	var let_result_928 flagrt.Value
	{
		var ir = flagrt.NewMap(flagKw_kind, flagKw_ident, flagKw_name, ident)
		var if_result_927 flagrt.Value
		if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(k, flagKw_string))) {
			if_result_927 = flagrt.NewMap(flagKw_ir, ir, flagKw_expr_kind, flagKw_string)
		} else {
			var if_result_926 flagrt.Value
			if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(k, flagKw_bool))) {
				if_result_926 = flagrt.NewMap(flagKw_ir, ir, flagKw_expr_kind, flagKw_bool)
			} else {
				if_result_926 = flagrt.NewMap(flagKw_ir, ir, flagKw_expr_kind, flagKw_value)
			}
			if_result_927 = if_result_926
		}
		let_result_928 = if_result_927
	}
	return let_result_928
}

func compiler__ident_ir_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 2 {
		panic("compiler__ident_ir expects exactly 2 arguments")
	}
	return compiler__ident_ir_arity_2(args[0], args[1])
}

func compiler__symbol_to_ir_arity_3(name flagrt.Value, ident flagrt.Value, ctx flagrt.Value) flagrt.Value {
	var let_result_983 flagrt.Value
	{
		var locals = flagrt.Call(flagKw_locals, ctx)
		var globals = flagrt.Call(flagKw_globals, ctx)
		var functions = flagrt.Call(flagKw_functions, ctx)
		var module = flagrt.Call(flagKw_module, ctx)
		var go_fns = flagrt.Call(flagKw_go_fns, ctx)
		var self_name = flagrt.Call(flagKw_self_function_name, ctx)
		var self_var = flagrt.Call(flagKw_self_variadic_name, ctx)
		var slash_q = flagrt.Call(compiler__qualified_name_q, name)
		var let_result_939 flagrt.Value
		{
			var if_result_929 flagrt.Value
			if flagrt.IsTruthy(slash_q) {
				if_result_929 = flagrt.NewBool(false)
			} else {
				if_result_929 = flagrt.NewBool(true)
			}
			var and_tmp = if_result_929
			var if_result_938 flagrt.Value
			if flagrt.IsTruthy(and_tmp) {
				var let_result_937 flagrt.Value
				{
					var and_tmp = ident
					var if_result_936 flagrt.Value
					if flagrt.IsTruthy(and_tmp) {
						var let_result_935 flagrt.Value
						{
							var if_result_930 flagrt.Value
							if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagrt.NewString(""), ident))) {
								if_result_930 = flagrt.NewBool(false)
							} else {
								if_result_930 = flagrt.NewBool(true)
							}
							var and_tmp = if_result_930
							var if_result_934 flagrt.Value
							if flagrt.IsTruthy(and_tmp) {
								var let_result_933 flagrt.Value
								{
									var if_result_931 flagrt.Value
									if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsNil(locals))) {
										if_result_931 = flagrt.NewBool(false)
									} else {
										if_result_931 = flagrt.NewBool(true)
									}
									var and_tmp = if_result_931
									var if_result_932 flagrt.Value
									if flagrt.IsTruthy(and_tmp) {
										if_result_932 = flagrt.NewBool(flagrt.Contains(locals, ident))
									} else {
										if_result_932 = and_tmp
									}
									let_result_933 = if_result_932
								}
								if_result_934 = let_result_933
							} else {
								if_result_934 = and_tmp
							}
							let_result_935 = if_result_934
						}
						if_result_936 = let_result_935
					} else {
						if_result_936 = and_tmp
					}
					let_result_937 = if_result_936
				}
				if_result_938 = let_result_937
			} else {
				if_result_938 = and_tmp
			}
			let_result_939 = if_result_938
		}
		var if_result_982 flagrt.Value
		if flagrt.IsTruthy(let_result_939) {
			if_result_982 = flagrt.Call(compiler__ident_ir, ident, flagrt.Call(flagrt.BuiltinFunction("get"), locals, ident))
		} else {
			var let_result_944 flagrt.Value
			{
				var and_tmp = flagrt.NewBool(flagrt.Eq(name, self_name))
				var if_result_943 flagrt.Value
				if flagrt.IsTruthy(and_tmp) {
					var let_result_942 flagrt.Value
					{
						var and_tmp = self_var
						var if_result_941 flagrt.Value
						if flagrt.IsTruthy(and_tmp) {
							var if_result_940 flagrt.Value
							if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagrt.NewString(""), self_var))) {
								if_result_940 = flagrt.NewBool(false)
							} else {
								if_result_940 = flagrt.NewBool(true)
							}
							if_result_941 = if_result_940
						} else {
							if_result_941 = and_tmp
						}
						let_result_942 = if_result_941
					}
					if_result_943 = let_result_942
				} else {
					if_result_943 = and_tmp
				}
				let_result_944 = if_result_943
			}
			var if_result_981 flagrt.Value
			if flagrt.IsTruthy(let_result_944) {
				if_result_981 = flagrt.NewMap(flagKw_ir, flagrt.Call(compiler__literal_rt_call, flagStr_NewFunction, flagrt.NewArray(flagrt.NewMap(flagKw_kind, flagKw_ident, flagKw_name, self_var))), flagKw_expr_kind, flagKw_value)
			} else {
				var let_result_947 flagrt.Value
				{
					var if_result_945 flagrt.Value
					if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsNil(module))) {
						if_result_945 = flagrt.NewBool(false)
					} else {
						if_result_945 = flagrt.NewBool(true)
					}
					var and_tmp = if_result_945
					var if_result_946 flagrt.Value
					if flagrt.IsTruthy(and_tmp) {
						if_result_946 = flagrt.NewBool(flagrt.Contains(module, name))
					} else {
						if_result_946 = and_tmp
					}
					let_result_947 = if_result_946
				}
				var if_result_980 flagrt.Value
				if flagrt.IsTruthy(let_result_947) {
					var let_result_956 flagrt.Value
					{
						var go_ident = flagrt.Call(flagrt.BuiltinFunction("get"), module, name)
						var let_result_950 flagrt.Value
						{
							var if_result_948 flagrt.Value
							if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsNil(globals))) {
								if_result_948 = flagrt.NewBool(false)
							} else {
								if_result_948 = flagrt.NewBool(true)
							}
							var and_tmp = if_result_948
							var if_result_949 flagrt.Value
							if flagrt.IsTruthy(and_tmp) {
								if_result_949 = flagrt.NewBool(flagrt.Contains(globals, go_ident))
							} else {
								if_result_949 = and_tmp
							}
							let_result_950 = if_result_949
						}
						var if_result_955 flagrt.Value
						if flagrt.IsTruthy(let_result_950) {
							if_result_955 = flagrt.Call(compiler__ident_ir, go_ident, flagrt.Call(flagrt.BuiltinFunction("get"), globals, go_ident))
						} else {
							var let_result_953 flagrt.Value
							{
								var if_result_951 flagrt.Value
								if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsNil(functions))) {
									if_result_951 = flagrt.NewBool(false)
								} else {
									if_result_951 = flagrt.NewBool(true)
								}
								var and_tmp = if_result_951
								var if_result_952 flagrt.Value
								if flagrt.IsTruthy(and_tmp) {
									if_result_952 = flagrt.NewBool(flagrt.Contains(functions, go_ident))
								} else {
									if_result_952 = and_tmp
								}
								let_result_953 = if_result_952
							}
							var if_result_954 flagrt.Value
							if flagrt.IsTruthy(let_result_953) {
								if_result_954 = flagrt.NewMap(flagKw_ir, flagrt.Call(compiler__literal_rt_call, flagStr_NewFunction, flagrt.NewArray(flagrt.NewMap(flagKw_kind, flagKw_ident, flagKw_name, go_ident))), flagKw_expr_kind, flagKw_value)
							} else {
								if_result_954 = flagrt.NewMap(flagKw_ir, flagrt.NewMap(flagKw_kind, flagKw_ident, flagKw_name, go_ident), flagKw_expr_kind, flagKw_value)
							}
							if_result_955 = if_result_954
						}
						let_result_956 = if_result_955
					}
					if_result_980 = let_result_956
				} else {
					var let_result_964 flagrt.Value
					{
						var and_tmp = ident
						var if_result_963 flagrt.Value
						if flagrt.IsTruthy(and_tmp) {
							var let_result_962 flagrt.Value
							{
								var if_result_957 flagrt.Value
								if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagrt.NewString(""), ident))) {
									if_result_957 = flagrt.NewBool(false)
								} else {
									if_result_957 = flagrt.NewBool(true)
								}
								var and_tmp = if_result_957
								var if_result_961 flagrt.Value
								if flagrt.IsTruthy(and_tmp) {
									var let_result_960 flagrt.Value
									{
										var if_result_958 flagrt.Value
										if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsNil(globals))) {
											if_result_958 = flagrt.NewBool(false)
										} else {
											if_result_958 = flagrt.NewBool(true)
										}
										var and_tmp = if_result_958
										var if_result_959 flagrt.Value
										if flagrt.IsTruthy(and_tmp) {
											if_result_959 = flagrt.NewBool(flagrt.Contains(globals, ident))
										} else {
											if_result_959 = and_tmp
										}
										let_result_960 = if_result_959
									}
									if_result_961 = let_result_960
								} else {
									if_result_961 = and_tmp
								}
								let_result_962 = if_result_961
							}
							if_result_963 = let_result_962
						} else {
							if_result_963 = and_tmp
						}
						let_result_964 = if_result_963
					}
					var if_result_979 flagrt.Value
					if flagrt.IsTruthy(let_result_964) {
						if_result_979 = flagrt.Call(compiler__ident_ir, ident, flagrt.Call(flagrt.BuiltinFunction("get"), globals, ident))
					} else {
						var let_result_972 flagrt.Value
						{
							var and_tmp = ident
							var if_result_971 flagrt.Value
							if flagrt.IsTruthy(and_tmp) {
								var let_result_970 flagrt.Value
								{
									var if_result_965 flagrt.Value
									if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagrt.NewString(""), ident))) {
										if_result_965 = flagrt.NewBool(false)
									} else {
										if_result_965 = flagrt.NewBool(true)
									}
									var and_tmp = if_result_965
									var if_result_969 flagrt.Value
									if flagrt.IsTruthy(and_tmp) {
										var let_result_968 flagrt.Value
										{
											var if_result_966 flagrt.Value
											if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsNil(functions))) {
												if_result_966 = flagrt.NewBool(false)
											} else {
												if_result_966 = flagrt.NewBool(true)
											}
											var and_tmp = if_result_966
											var if_result_967 flagrt.Value
											if flagrt.IsTruthy(and_tmp) {
												if_result_967 = flagrt.NewBool(flagrt.Contains(functions, ident))
											} else {
												if_result_967 = and_tmp
											}
											let_result_968 = if_result_967
										}
										if_result_969 = let_result_968
									} else {
										if_result_969 = and_tmp
									}
									let_result_970 = if_result_969
								}
								if_result_971 = let_result_970
							} else {
								if_result_971 = and_tmp
							}
							let_result_972 = if_result_971
						}
						var if_result_978 flagrt.Value
						if flagrt.IsTruthy(let_result_972) {
							if_result_978 = flagrt.NewMap(flagKw_ir, flagrt.Call(compiler__literal_rt_call, flagStr_NewFunction, flagrt.NewArray(flagrt.NewMap(flagKw_kind, flagKw_ident, flagKw_name, ident))), flagKw_expr_kind, flagKw_value)
						} else {
							var if_result_977 flagrt.Value
							if flagrt.IsTruthy(flagrt.NewBool(flagrt.Contains(compiler__builtin_function_names, name))) {
								if_result_977 = flagrt.NewMap(flagKw_ir, flagrt.Call(compiler__literal_rt_call, flagStr_BuiltinFunction, flagrt.NewArray(flagrt.NewMap(flagKw_kind, flagKw_string, flagKw_value, name))), flagKw_expr_kind, flagKw_value)
							} else {
								var let_result_975 flagrt.Value
								{
									var if_result_973 flagrt.Value
									if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsNil(go_fns))) {
										if_result_973 = flagrt.NewBool(false)
									} else {
										if_result_973 = flagrt.NewBool(true)
									}
									var and_tmp = if_result_973
									var if_result_974 flagrt.Value
									if flagrt.IsTruthy(and_tmp) {
										if_result_974 = flagrt.NewBool(flagrt.Contains(go_fns, name))
									} else {
										if_result_974 = and_tmp
									}
									let_result_975 = if_result_974
								}
								var if_result_976 flagrt.Value
								if flagrt.IsTruthy(let_result_975) {
									if_result_976 = flagrt.NewMap(flagKw_ir, flagrt.NewMap(flagKw_kind, flagKw_selector, flagKw_pkg, flagrt.NewString("flagrt"), flagKw_name, flagrt.Call(flagrt.BuiltinFunction("get"), go_fns, name)), flagKw_expr_kind, flagKw_value)
								} else {
									if_result_976 = func() flagrt.Value {
										flagrt.Throw(flagrt.NewMap(flagKw_message, flagrt.NewString(flagrt.Str("unknown symbol \"", name, "\"")), flagKw_data, flagrt.NewMap(flagKw_name, name)))
										return flagrt.NilValue()
									}()
								}
								if_result_977 = if_result_976
							}
							if_result_978 = if_result_977
						}
						if_result_979 = if_result_978
					}
					if_result_980 = if_result_979
				}
				if_result_981 = if_result_980
			}
			if_result_982 = if_result_981
		}
		let_result_983 = if_result_982
	}
	return let_result_983
}

func compiler__symbol_to_ir_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 3 {
		panic("compiler__symbol_to_ir expects exactly 3 arguments")
	}
	return compiler__symbol_to_ir_arity_3(args[0], args[1], args[2])
}

func compiler__eval_ast_to_ir_node_arity_3(mode flagrt.Value, node flagrt.Value, ctx flagrt.Value) flagrt.Value {
	var if_result_1031 flagrt.Value
	if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(mode, flagKw_seq))) {
		if_result_1031 = func() flagrt.Value {
			var remaining = node
			var out = flagVec
			for {
				var let_result_985 flagrt.Value
				{
					var or_tmp = flagrt.NewBool(flagrt.IsNil(remaining))
					var if_result_984 flagrt.Value
					if flagrt.IsTruthy(or_tmp) {
						if_result_984 = or_tmp
					} else {
						if_result_984 = flagrt.NewBool(flagrt.IsEmpty(remaining))
					}
					let_result_985 = if_result_984
				}
				var if_result_986 flagrt.Value
				if flagrt.IsTruthy(let_result_985) {
					if_result_986 = out
				} else {
					if_result_986 = flagrt.NewRecur(flagrt.Rest(remaining), flagrt.Conj(out, compiler__eval_ast_to_ir_node_arity_3(flagKw_eval, flagrt.First(remaining), ctx)))
				}
				__loopResult := if_result_986
				if __recurValues, __isRecur := flagrt.UnwrapRecur(__loopResult); __isRecur {
					if len(__recurValues) != 2 {
						panic("internal error: recur arity mismatch")
					}
					remaining = __recurValues[0]
					out = __recurValues[1]
					continue
				}
				return __loopResult
			}
		}()
	} else {
		var if_result_1030 flagrt.Value
		if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(mode, flagKw_eval))) {
			var let_result_1012 flagrt.Value
			{
				var kind = flagrt.Call(flagKw_kind, node)
				var if_result_1011 flagrt.Value
				if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_string))) {
					if_result_1011 = flagrt.Call(compiler__ast_node_to_ir, node)
				} else {
					var if_result_1010 flagrt.Value
					if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_char))) {
						if_result_1010 = flagrt.Call(compiler__ast_node_to_ir, node)
					} else {
						var if_result_1009 flagrt.Value
						if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_int))) {
							if_result_1009 = flagrt.Call(compiler__ast_node_to_ir, node)
						} else {
							var if_result_1008 flagrt.Value
							if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_bigint))) {
								if_result_1008 = flagrt.Call(compiler__ast_node_to_ir, node)
							} else {
								var if_result_1007 flagrt.Value
								if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_ratio))) {
									if_result_1007 = flagrt.Call(compiler__ast_node_to_ir, node)
								} else {
									var if_result_1006 flagrt.Value
									if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_float))) {
										if_result_1006 = flagrt.Call(compiler__ast_node_to_ir, node)
									} else {
										var if_result_1005 flagrt.Value
										if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_keyword))) {
											if_result_1005 = flagrt.Call(compiler__ast_node_to_ir, node)
										} else {
											var if_result_1004 flagrt.Value
											if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_quoted_symbol))) {
												if_result_1004 = flagrt.Call(compiler__ast_node_to_ir, node)
											} else {
												var if_result_1003 flagrt.Value
												if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_symbol))) {
													var let_result_992 flagrt.Value
													{
														var name = flagrt.Call(flagKw_name, node)
														var let_result_990 flagrt.Value
														{
															var or_tmp = flagrt.NewBool(flagrt.Eq(name, flagrt.NewString("true")))
															var if_result_989 flagrt.Value
															if flagrt.IsTruthy(or_tmp) {
																if_result_989 = or_tmp
															} else {
																var let_result_988 flagrt.Value
																{
																	var or_tmp = flagrt.NewBool(flagrt.Eq(name, flagrt.NewString("false")))
																	var if_result_987 flagrt.Value
																	if flagrt.IsTruthy(or_tmp) {
																		if_result_987 = or_tmp
																	} else {
																		if_result_987 = flagrt.NewBool(flagrt.Eq(name, flagrt.NewString("nil")))
																	}
																	let_result_988 = if_result_987
																}
																if_result_989 = let_result_988
															}
															let_result_990 = if_result_989
														}
														var if_result_991 flagrt.Value
														if flagrt.IsTruthy(let_result_990) {
															if_result_991 = flagrt.Call(compiler__ast_node_to_ir, node)
														} else {
															if_result_991 = flagrt.Call(flagKw_ir, flagrt.Call(compiler__symbol_to_ir, name, name, ctx))
														}
														let_result_992 = if_result_991
													}
													if_result_1003 = let_result_992
												} else {
													var if_result_1002 flagrt.Value
													if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_quoted_list))) {
														if_result_1002 = flagrt.Call(compiler__quoted_ast_to_ir, node)
													} else {
														var if_result_1001 flagrt.Value
														if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_vector))) {
															if_result_1001 = flagrt.Call(compiler__ctor_to_ir, flagStr_NewArray, compiler__eval_ast_to_ir_node_arity_3(flagKw_seq, flagrt.Call(flagKw_elements, node), ctx))
														} else {
															var if_result_1000 flagrt.Value
															if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_pipe_vector))) {
																if_result_1000 = flagrt.Call(compiler__ctor_to_ir, flagStr_NewVector, compiler__eval_ast_to_ir_node_arity_3(flagKw_seq, flagrt.Call(flagKw_elements, node), ctx))
															} else {
																var if_result_999 flagrt.Value
																if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_set))) {
																	if_result_999 = flagrt.Call(compiler__ctor_to_ir, flagStr_NewSet, compiler__eval_ast_to_ir_node_arity_3(flagKw_seq, flagrt.Call(flagKw_elements, node), ctx))
																} else {
																	var if_result_998 flagrt.Value
																	if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_map))) {
																		var let_result_996 flagrt.Value
																		{
																			var entries = flagrt.Call(flagKw_entries, node)
																			var if_result_993 flagrt.Value
																			if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsNil(entries))) {
																				if_result_993 = flagrt.NewLong(0)
																			} else {
																				if_result_993 = flagrt.NewLong(int64(flagrt.Count(entries)))
																			}
																			var n = if_result_993
																			var if_result_994 flagrt.Value
																			if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagrt.NewLong(0), flagrt.Mod(n, flagrt.NewLong(2))))) {
																				if_result_994 = flagrt.NewBool(false)
																			} else {
																				if_result_994 = flagrt.NewBool(true)
																			}
																			var if_result_995 flagrt.Value
																			if flagrt.IsTruthy(if_result_994) {
																				if_result_995 = func() flagrt.Value {
																					flagrt.Throw(flagrt.NewMap(flagKw_message, flagrt.NewString("map literal expects key/value pairs"), flagKw_data, flagrt.NewMap(flagKw_node, node)))
																					return flagrt.NilValue()
																				}()
																			} else {
																				if_result_995 = flagrt.Call(compiler__ctor_to_ir, flagStr_NewMap, compiler__eval_ast_to_ir_node_arity_3(flagKw_seq, entries, ctx))
																			}
																			let_result_996 = if_result_995
																		}
																		if_result_998 = let_result_996
																	} else {
																		var if_result_997 flagrt.Value
																		if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_list))) {
																			if_result_997 = compiler__eval_ast_to_ir_node_arity_3(flagKw_call, node, ctx)
																		} else {
																			if_result_997 = func() flagrt.Value {
																				flagrt.Throw(flagrt.NewMap(flagKw_message, flagrt.NewString("unsupported AST node for eval lowering"), flagKw_data, flagrt.NewMap(flagKw_node, node)))
																				return flagrt.NilValue()
																			}()
																		}
																		if_result_998 = if_result_997
																	}
																	if_result_999 = if_result_998
																}
																if_result_1000 = if_result_999
															}
															if_result_1001 = if_result_1000
														}
														if_result_1002 = if_result_1001
													}
													if_result_1003 = if_result_1002
												}
												if_result_1004 = if_result_1003
											}
											if_result_1005 = if_result_1004
										}
										if_result_1006 = if_result_1005
									}
									if_result_1007 = if_result_1006
								}
								if_result_1008 = if_result_1007
							}
							if_result_1009 = if_result_1008
						}
						if_result_1010 = if_result_1009
					}
					if_result_1011 = if_result_1010
				}
				let_result_1012 = if_result_1011
			}
			if_result_1030 = let_result_1012
		} else {
			var if_result_1029 flagrt.Value
			if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(mode, flagKw_call))) {
				var let_result_1028 flagrt.Value
				{
					var elements = flagrt.Call(flagKw_elements, node)
					var let_result_1014 flagrt.Value
					{
						var or_tmp = flagrt.NewBool(flagrt.IsNil(elements))
						var if_result_1013 flagrt.Value
						if flagrt.IsTruthy(or_tmp) {
							if_result_1013 = or_tmp
						} else {
							if_result_1013 = flagrt.NewBool(flagrt.IsEmpty(elements))
						}
						let_result_1014 = if_result_1013
					}
					var if_result_1027 flagrt.Value
					if flagrt.IsTruthy(let_result_1014) {
						if_result_1027 = func() flagrt.Value {
							flagrt.Throw(flagrt.NewMap(flagKw_message, flagrt.NewString("unsupported form"), flagKw_data, flagrt.NewMap(flagKw_node, node)))
							return flagrt.NilValue()
						}()
					} else {
						var let_result_1026 flagrt.Value
						{
							var head = flagrt.First(elements)
							var args = flagrt.Rest(elements)
							var if_result_1025 flagrt.Value
							if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagKw_symbol, flagrt.Call(flagKw_kind, head)))) {
								var let_result_1024 flagrt.Value
								{
									var op = flagrt.Call(flagKw_name, head)
									var unary = flagrt.Call(flagrt.BuiltinFunction("get"), compiler__unary_runtime_ops, op)
									var fold = flagrt.Call(flagrt.BuiltinFunction("get"), compiler__fold_runtime_ops, op)
									var binary = flagrt.Call(flagrt.BuiltinFunction("get"), compiler__binary_runtime_ops, op)
									var variadic = flagrt.Call(flagrt.BuiltinFunction("get"), compiler__variadic_runtime_ops, op)
									var if_result_1023 flagrt.Value
									if flagrt.IsTruthy(unary) {
										var if_result_1015 flagrt.Value
										if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagrt.NewLong(1), flagrt.NewLong(int64(flagrt.Count(args)))))) {
											if_result_1015 = flagrt.NewBool(false)
										} else {
											if_result_1015 = flagrt.NewBool(true)
										}
										var if_result_1016 flagrt.Value
										if flagrt.IsTruthy(if_result_1015) {
											if_result_1016 = func() flagrt.Value {
												flagrt.Throw(flagrt.NewMap(flagKw_message, flagrt.NewString(flagrt.Str(op, " expects exactly one argument")), flagKw_data, flagrt.NewMap(flagKw_node, node)))
												return flagrt.NilValue()
											}()
										} else {
											if_result_1016 = flagrt.Call(compiler__runtime_call_to_ir, unary, compiler__eval_ast_to_ir_node_arity_3(flagKw_seq, args, ctx))
										}
										if_result_1023 = if_result_1016
									} else {
										var if_result_1022 flagrt.Value
										if flagrt.IsTruthy(fold) {
											if_result_1022 = flagrt.Call(compiler__fold_call_to_ir, fold, compiler__eval_ast_to_ir_node_arity_3(flagKw_seq, args, ctx))
										} else {
											var if_result_1021 flagrt.Value
											if flagrt.IsTruthy(binary) {
												var if_result_1017 flagrt.Value
												if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagrt.NewLong(2), flagrt.NewLong(int64(flagrt.Count(args)))))) {
													if_result_1017 = flagrt.NewBool(false)
												} else {
													if_result_1017 = flagrt.NewBool(true)
												}
												var if_result_1018 flagrt.Value
												if flagrt.IsTruthy(if_result_1017) {
													if_result_1018 = func() flagrt.Value {
														flagrt.Throw(flagrt.NewMap(flagKw_message, flagrt.NewString(flagrt.Str(op, " expects exactly two arguments")), flagKw_data, flagrt.NewMap(flagKw_node, node)))
														return flagrt.NilValue()
													}()
												} else {
													if_result_1018 = flagrt.Call(compiler__runtime_call_to_ir, binary, compiler__eval_ast_to_ir_node_arity_3(flagKw_seq, args, ctx))
												}
												if_result_1021 = if_result_1018
											} else {
												var if_result_1020 flagrt.Value
												if flagrt.IsTruthy(variadic) {
													if_result_1020 = flagrt.Call(compiler__runtime_call_to_ir, variadic, compiler__eval_ast_to_ir_node_arity_3(flagKw_seq, args, ctx))
												} else {
													var if_result_1019 flagrt.Value
													if flagrt.IsTruthy(flagrt.NewBool(flagrt.Contains(compiler__special_form_names, op))) {
														if_result_1019 = func() flagrt.Value {
															flagrt.Throw(flagrt.NewMap(flagKw_message, flagrt.NewString("special form not handled by eval lowering"), flagKw_data, flagrt.NewMap(flagKw_node, node)))
															return flagrt.NilValue()
														}()
													} else {
														if_result_1019 = flagrt.Call(compiler__call_to_ir, compiler__eval_ast_to_ir_node_arity_3(flagKw_eval, head, ctx), compiler__eval_ast_to_ir_node_arity_3(flagKw_seq, args, ctx))
													}
													if_result_1020 = if_result_1019
												}
												if_result_1021 = if_result_1020
											}
											if_result_1022 = if_result_1021
										}
										if_result_1023 = if_result_1022
									}
									let_result_1024 = if_result_1023
								}
								if_result_1025 = let_result_1024
							} else {
								if_result_1025 = flagrt.Call(compiler__call_to_ir, compiler__eval_ast_to_ir_node_arity_3(flagKw_eval, head, ctx), compiler__eval_ast_to_ir_node_arity_3(flagKw_seq, args, ctx))
							}
							let_result_1026 = if_result_1025
						}
						if_result_1027 = let_result_1026
					}
					let_result_1028 = if_result_1027
				}
				if_result_1029 = let_result_1028
			} else {
				if_result_1029 = func() flagrt.Value {
					flagrt.Throw(flagrt.NewMap(flagKw_message, flagrt.NewString("unsupported eval lowering mode"), flagKw_data, flagrt.NewMap(flagKw_mode, mode)))
					return flagrt.NilValue()
				}()
			}
			if_result_1030 = if_result_1029
		}
		if_result_1031 = if_result_1030
	}
	return if_result_1031
}

func compiler__eval_ast_to_ir_node_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 3 {
		panic("compiler__eval_ast_to_ir_node expects exactly 3 arguments")
	}
	return compiler__eval_ast_to_ir_node_arity_3(args[0], args[1], args[2])
}

func compiler__eval_ast_to_ir_arity_2(node flagrt.Value, ctx flagrt.Value) flagrt.Value {
	return flagrt.Call(compiler__eval_ast_to_ir_node, flagKw_eval, node, ctx)
}

func compiler__eval_ast_to_ir_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 2 {
		panic("compiler__eval_ast_to_ir expects exactly 2 arguments")
	}
	return compiler__eval_ast_to_ir_arity_2(args[0], args[1])
}

func compiler__stmt_list_arity_1(stmts flagrt.Value) flagrt.Value {
	var let_result_1033 flagrt.Value
	{
		var or_tmp = flagrt.NewBool(flagrt.IsNil(stmts))
		var if_result_1032 flagrt.Value
		if flagrt.IsTruthy(or_tmp) {
			if_result_1032 = or_tmp
		} else {
			if_result_1032 = flagrt.NewBool(flagrt.IsEmpty(stmts))
		}
		let_result_1033 = if_result_1032
	}
	var if_result_1034 flagrt.Value
	if flagrt.IsTruthy(let_result_1033) {
		if_result_1034 = flagVec
	} else {
		if_result_1034 = stmts
	}
	return if_result_1034
}

func compiler__stmt_list_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__stmt_list expects exactly 1 arguments")
	}
	return compiler__stmt_list_arity_1(args[0])
}

func compiler__append_stmt_arity_2(stmts flagrt.Value, stmt flagrt.Value) flagrt.Value {
	return flagrt.Conj(flagrt.Call(compiler__stmt_list, stmts), stmt)
}

func compiler__append_stmt_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 2 {
		panic("compiler__append_stmt expects exactly 2 arguments")
	}
	return compiler__append_stmt_arity_2(args[0], args[1])
}

func compiler__if_to_ir_arity_7(name flagrt.Value, type_ flagrt.Value, cond flagrt.Value, then_stmts flagrt.Value, then_expr flagrt.Value, else_stmts flagrt.Value, else_expr flagrt.Value) flagrt.Value {
	return flagrt.NewArray(flagrt.NewMap(flagKw_kind, flagKw_var, flagKw_name, name, flagKw_type, type_, flagKw_expr, flagrt.NilValue()), flagrt.NewMap(flagKw_kind, flagKw_if, flagKw_init, flagrt.NewString(""), flagKw_cond, cond, flagKw_then, flagrt.Call(compiler__append_stmt, then_stmts, flagrt.NewMap(flagKw_kind, flagKw_assign, flagKw_name, name, flagKw_expr, then_expr)), flagKw_else, flagrt.Call(compiler__append_stmt, else_stmts, flagrt.NewMap(flagKw_kind, flagKw_assign, flagKw_name, name, flagKw_expr, else_expr))))
}

func compiler__if_to_ir_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 7 {
		panic("compiler__if_to_ir expects exactly 7 arguments")
	}
	return compiler__if_to_ir_arity_7(args[0], args[1], args[2], args[3], args[4], args[5], args[6])
}

func compiler__concat_stmts_arity_2(left flagrt.Value, right flagrt.Value) flagrt.Value {
	return func() flagrt.Value {
		var out = flagrt.Call(compiler__stmt_list, left)
		var remaining = flagrt.Call(compiler__stmt_list, right)
		for {
			var let_result_1036 flagrt.Value
			{
				var or_tmp = flagrt.NewBool(flagrt.IsNil(remaining))
				var if_result_1035 flagrt.Value
				if flagrt.IsTruthy(or_tmp) {
					if_result_1035 = or_tmp
				} else {
					if_result_1035 = flagrt.NewBool(flagrt.IsEmpty(remaining))
				}
				let_result_1036 = if_result_1035
			}
			var if_result_1037 flagrt.Value
			if flagrt.IsTruthy(let_result_1036) {
				if_result_1037 = out
			} else {
				if_result_1037 = flagrt.NewRecur(flagrt.Conj(out, flagrt.First(remaining)), flagrt.Rest(remaining))
			}
			__loopResult := if_result_1037
			if __recurValues, __isRecur := flagrt.UnwrapRecur(__loopResult); __isRecur {
				if len(__recurValues) != 2 {
					panic("internal error: recur arity mismatch")
				}
				out = __recurValues[0]
				remaining = __recurValues[1]
				continue
			}
			return __loopResult
		}
	}()
}

func compiler__concat_stmts_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 2 {
		panic("compiler__concat_stmts expects exactly 2 arguments")
	}
	return compiler__concat_stmts_arity_2(args[0], args[1])
}

func compiler__let_to_ir_arity_5(name flagrt.Value, type_ flagrt.Value, binding_stmts flagrt.Value, body_stmts flagrt.Value, body_expr flagrt.Value) flagrt.Value {
	return flagrt.NewArray(flagrt.NewMap(flagKw_kind, flagKw_var, flagKw_name, name, flagKw_type, type_, flagKw_expr, flagrt.NilValue()), flagrt.NewMap(flagKw_kind, flagKw_block, flagKw_body, flagrt.Call(compiler__append_stmt, flagrt.Call(compiler__concat_stmts, binding_stmts, body_stmts), flagrt.NewMap(flagKw_kind, flagKw_assign, flagKw_name, name, flagKw_expr, body_expr))))
}

func compiler__let_to_ir_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 5 {
		panic("compiler__let_to_ir expects exactly 5 arguments")
	}
	return compiler__let_to_ir_arity_5(args[0], args[1], args[2], args[3], args[4])
}

func compiler_ast_source__ast__gtsource_arity_1(expr flagrt.Value) flagrt.Value {
	var let_result_1066 flagrt.Value
	{
		var kind = flagrt.Call(flagKw_kind, expr)
		var join = flagrt.NewFunction(func(args ...flagrt.Value) flagrt.Value {
			if len(args) != 1 {
				panic("fn expects exactly 1 arguments")
			}
			nodes := args[0]
			return func() flagrt.Value {
				var remaining = nodes
				var out = flagStr_
				var first_q = flagrt.NewBool(true)
				for {
					var if_result_1040 flagrt.Value
					if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsEmpty(remaining))) {
						if_result_1040 = out
					} else {
						var let_result_1039 flagrt.Value
						{
							var piece = compiler_ast_source__ast__gtsource_arity_1(flagrt.First(remaining))
							var if_result_1038 flagrt.Value
							if flagrt.IsTruthy(first_q) {
								if_result_1038 = flagrt.NewRecur(flagrt.Rest(remaining), piece, flagrt.NewBool(false))
							} else {
								if_result_1038 = flagrt.NewRecur(flagrt.Rest(remaining), flagrt.NewString(flagrt.Str(out, " ", piece)), flagrt.NewBool(false))
							}
							let_result_1039 = if_result_1038
						}
						if_result_1040 = let_result_1039
					}
					__loopResult := if_result_1040
					if __recurValues, __isRecur := flagrt.UnwrapRecur(__loopResult); __isRecur {
						if len(__recurValues) != 3 {
							panic("internal error: recur arity mismatch")
						}
						remaining = __recurValues[0]
						out = __recurValues[1]
						first_q = __recurValues[2]
						continue
					}
					return __loopResult
				}
			}()
		})
		var if_result_1065 flagrt.Value
		if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_symbol))) {
			if_result_1065 = flagrt.Call(flagKw_name, expr)
		} else {
			var if_result_1064 flagrt.Value
			if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_keyword))) {
				if_result_1064 = flagrt.NewString(flagrt.Str(":", flagrt.Call(flagKw_name, expr)))
			} else {
				var if_result_1063 flagrt.Value
				if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_quoted_symbol))) {
					if_result_1063 = flagrt.NewString(flagrt.Str("'", flagrt.Call(flagKw_name, expr)))
				} else {
					var if_result_1062 flagrt.Value
					if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_int))) {
						if_result_1062 = flagrt.NewString(flagrt.Str(flagrt.Call(flagKw_value, expr)))
					} else {
						var if_result_1061 flagrt.Value
						if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_bigint))) {
							if_result_1061 = flagrt.NewString(flagrt.Str(flagrt.Call(flagKw_value, expr), "N"))
						} else {
							var if_result_1060 flagrt.Value
							if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_float))) {
								if_result_1060 = flagrt.Call(flagKw_raw, expr)
							} else {
								var if_result_1059 flagrt.Value
								if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_ratio))) {
									if_result_1059 = flagrt.NewString(flagrt.Str(flagrt.Call(flagKw_numerator, expr), "/", flagrt.Call(flagKw_denominator, expr)))
								} else {
									var if_result_1058 flagrt.Value
									if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_string))) {
										if_result_1058 = flagrt.NewString(flagrt.Format("%q", flagrt.Call(flagKw_value, expr)))
									} else {
										var if_result_1057 flagrt.Value
										if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_char))) {
											var let_result_1045 flagrt.Value
											{
												var ch = flagrt.Call(flagKw_value, expr)
												var if_result_1044 flagrt.Value
												if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(ch, flagrt.NewString(" ")))) {
													if_result_1044 = flagStr__space
												} else {
													var if_result_1043 flagrt.Value
													if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(ch, flagrt.NewString("\n")))) {
														if_result_1043 = flagStr__newline
													} else {
														var if_result_1042 flagrt.Value
														if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(ch, flagrt.NewString("\t")))) {
															if_result_1042 = flagStr__tab
														} else {
															var if_result_1041 flagrt.Value
															if flagrt.IsTruthy(flagrt.NewBool(true)) {
																if_result_1041 = flagrt.NewString(flagrt.Str("\\", ch))
															} else {
																if_result_1041 = flagrt.NilValue()
															}
															if_result_1042 = if_result_1041
														}
														if_result_1043 = if_result_1042
													}
													if_result_1044 = if_result_1043
												}
												let_result_1045 = if_result_1044
											}
											if_result_1057 = let_result_1045
										} else {
											var if_result_1056 flagrt.Value
											if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_list))) {
												if_result_1056 = flagrt.NewString(flagrt.Str("(", flagrt.Call(join, flagrt.Call(flagKw_elements, expr)), ")"))
											} else {
												var if_result_1055 flagrt.Value
												if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_vector))) {
													if_result_1055 = flagrt.NewString(flagrt.Str("[", flagrt.Call(join, flagrt.Call(flagKw_elements, expr)), "]"))
												} else {
													var if_result_1054 flagrt.Value
													if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_pipe_vector))) {
														var let_result_1047 string
														{
															var inner = flagrt.Call(join, flagrt.Call(flagKw_elements, expr))
															var if_result_1046 string
															if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsEmpty(inner))) {
																if_result_1046 = "| |"
															} else {
																if_result_1046 = flagrt.Str("| ", inner, " |")
															}
															let_result_1047 = if_result_1046
														}
														if_result_1054 = flagrt.NewString(let_result_1047)
													} else {
														var if_result_1053 flagrt.Value
														if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_map))) {
															if_result_1053 = flagrt.NewString(flagrt.Str("{", flagrt.Call(join, flagrt.Call(flagKw_entries, expr)), "}"))
														} else {
															var if_result_1052 flagrt.Value
															if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_set))) {
																if_result_1052 = flagrt.NewString(flagrt.Str("#{", flagrt.Call(join, flagrt.Call(flagKw_elements, expr)), "}"))
															} else {
																var if_result_1051 flagrt.Value
																if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_quoted_list))) {
																	if_result_1051 = flagrt.NewString(flagrt.Str("'(", flagrt.Call(join, flagrt.Call(flagKw_elements, expr)), ")"))
																} else {
																	var if_result_1050 flagrt.Value
																	if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_hash_fn))) {
																		if_result_1050 = flagrt.NewString(flagrt.Str("#(", compiler_ast_source__ast__gtsource_arity_1(flagrt.Call(flagKw_body, expr)), ")"))
																	} else {
																		var if_result_1049 flagrt.Value
																		if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(kind, flagKw_meta))) {
																			if_result_1049 = flagrt.NewString(flagrt.Str("^", compiler_ast_source__ast__gtsource_arity_1(flagrt.Call(flagKw_meta, expr)), " ", compiler_ast_source__ast__gtsource_arity_1(flagrt.Call(flagKw_target, expr))))
																		} else {
																			var if_result_1048 flagrt.Value
																			if flagrt.IsTruthy(flagrt.NewBool(true)) {
																				if_result_1048 = func() flagrt.Value {
																					flagrt.Throw(flagrt.NewMap(flagKw_message, flagrt.NewString("unsupported AST node"), flagKw_data, flagrt.NewMap(flagKw_expr, expr)))
																					return flagrt.NilValue()
																				}()
																			} else {
																				if_result_1048 = flagrt.NilValue()
																			}
																			if_result_1049 = if_result_1048
																		}
																		if_result_1050 = if_result_1049
																	}
																	if_result_1051 = if_result_1050
																}
																if_result_1052 = if_result_1051
															}
															if_result_1053 = if_result_1052
														}
														if_result_1054 = if_result_1053
													}
													if_result_1055 = if_result_1054
												}
												if_result_1056 = if_result_1055
											}
											if_result_1057 = if_result_1056
										}
										if_result_1058 = if_result_1057
									}
									if_result_1059 = if_result_1058
								}
								if_result_1060 = if_result_1059
							}
							if_result_1061 = if_result_1060
						}
						if_result_1062 = if_result_1061
					}
					if_result_1063 = if_result_1062
				}
				if_result_1064 = if_result_1063
			}
			if_result_1065 = if_result_1064
		}
		let_result_1066 = if_result_1065
	}
	return let_result_1066
}

func compiler_ast_source__ast__gtsource_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler_ast_source__ast__gtsource expects exactly 1 arguments")
	}
	return compiler_ast_source__ast__gtsource_arity_1(args[0])
}

func compiler_ast_source__asts__gtsource_arity_1(exprs flagrt.Value) flagrt.Value {
	return func() flagrt.Value {
		var remaining = exprs
		var out = flagStr_
		var first_q = flagrt.NewBool(true)
		for {
			var if_result_1069 flagrt.Value
			if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsEmpty(remaining))) {
				if_result_1069 = out
			} else {
				var let_result_1068 flagrt.Value
				{
					var piece = flagrt.Call(compiler_ast_source__ast__gtsource, flagrt.First(remaining))
					var if_result_1067 flagrt.Value
					if flagrt.IsTruthy(first_q) {
						if_result_1067 = flagrt.NewRecur(flagrt.Rest(remaining), piece, flagrt.NewBool(false))
					} else {
						if_result_1067 = flagrt.NewRecur(flagrt.Rest(remaining), flagrt.NewString(flagrt.Str(out, "\n", piece)), flagrt.NewBool(false))
					}
					let_result_1068 = if_result_1067
				}
				if_result_1069 = let_result_1068
			}
			__loopResult := if_result_1069
			if __recurValues, __isRecur := flagrt.UnwrapRecur(__loopResult); __isRecur {
				if len(__recurValues) != 3 {
					panic("internal error: recur arity mismatch")
				}
				remaining = __recurValues[0]
				out = __recurValues[1]
				first_q = __recurValues[2]
				continue
			}
			return __loopResult
		}
	}()
}

func compiler_ast_source__asts__gtsource_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler_ast_source__asts__gtsource expects exactly 1 arguments")
	}
	return compiler_ast_source__asts__gtsource_arity_1(args[0])
}

func compiler_tokenizer__drain_tokens_into_arity_2(ch flagrt.Value, acc flagrt.Value) flagrt.Value {
	var let_result_1071 flagrt.Value
	{
		var token = flagrt.Call(async__channel_receive, ch)
		var if_result_1070 flagrt.Value
		if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsNil(token))) {
			if_result_1070 = acc
		} else {
			if_result_1070 = compiler_tokenizer__drain_tokens_into_arity_2(ch, flagrt.Conj(acc, token))
		}
		let_result_1071 = if_result_1070
	}
	return let_result_1071
}

func compiler_tokenizer__drain_tokens_into_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 2 {
		panic("compiler_tokenizer__drain_tokens_into expects exactly 2 arguments")
	}
	return compiler_tokenizer__drain_tokens_into_arity_2(args[0], args[1])
}

func compiler_tokenizer__drain_tokens_arity_1(ch flagrt.Value) flagrt.Value {
	return flagrt.Call(compiler_tokenizer__drain_tokens_into, ch, flagVec)
}

func compiler_tokenizer__drain_tokens_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler_tokenizer__drain_tokens expects exactly 1 arguments")
	}
	return compiler_tokenizer__drain_tokens_arity_1(args[0])
}

func compiler_tokenizer__tokenize_file_to_array_arity_1(path flagrt.Value) flagrt.Value {
	return flagrt.Call(compiler_tokenizer__drain_tokens, flagrt.Call(compiler__tokenize_file, path))
}

func compiler_tokenizer__tokenize_file_to_array_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler_tokenizer__tokenize_file_to_array expects exactly 1 arguments")
	}
	return compiler_tokenizer__tokenize_file_to_array_arity_1(args[0])
}

func compiler_tokenizer__parse_file_to_array_arity_1(path flagrt.Value) flagrt.Value {
	return flagrt.Call(compiler_tokenizer__drain_tokens, flagrt.Call(compiler__parse_file, path))
}

func compiler_tokenizer__parse_file_to_array_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler_tokenizer__parse_file_to_array expects exactly 1 arguments")
	}
	return compiler_tokenizer__parse_file_to_array_arity_1(args[0])
}

func compiler_tokenizer__build_ast_file_to_array_arity_1(path flagrt.Value) flagrt.Value {
	return flagrt.Call(compiler_tokenizer__drain_tokens, flagrt.Call(compiler__build_ast_from_tokens, flagrt.Call(compiler__tokenize_file, path)))
}

func compiler_tokenizer__build_ast_file_to_array_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler_tokenizer__build_ast_file_to_array expects exactly 1 arguments")
	}
	return compiler_tokenizer__build_ast_file_to_array_arity_1(args[0])
}

func compiler_tokenizer__build_ast_source_to_array_arity_1(source flagrt.Value) flagrt.Value {
	return flagrt.Call(compiler_tokenizer__drain_tokens, flagrt.Call(compiler__build_ast_from_tokens, flagrt.Call(compiler__tokenize_source, source)))
}

func compiler_tokenizer__build_ast_source_to_array_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler_tokenizer__build_ast_source_to_array expects exactly 1 arguments")
	}
	return compiler_tokenizer__build_ast_source_to_array_arity_1(args[0])
}

func compiler_tokenizer__expand_file_to_array_arity_1(path flagrt.Value) flagrt.Value {
	return flagrt.Call(compiler_tokenizer__drain_tokens, flagrt.Call(compiler__parse_file, path))
}

func compiler_tokenizer__expand_file_to_array_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler_tokenizer__expand_file_to_array expects exactly 1 arguments")
	}
	return compiler_tokenizer__expand_file_to_array_arity_1(args[0])
}

func compiler_tokenizer__expand_source_to_array_arity_1(source flagrt.Value) flagrt.Value {
	return flagrt.Call(compiler_tokenizer__drain_tokens, flagrt.Call(compiler__expand_macros, flagrt.Call(compiler__build_ast_from_tokens, flagrt.Call(compiler__tokenize_source, source))))
}

func compiler_tokenizer__expand_source_to_array_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler_tokenizer__expand_source_to_array expects exactly 1 arguments")
	}
	return compiler_tokenizer__expand_source_to_array_arity_1(args[0])
}

func compiler_tokenizer__expand_source_arity_1(source flagrt.Value) flagrt.Value {
	return flagrt.Call(compiler_ast_source__asts__gtsource, flagrt.Call(compiler_tokenizer__expand_source_to_array, source))
}

func compiler_tokenizer__expand_source_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler_tokenizer__expand_source expects exactly 1 arguments")
	}
	return compiler_tokenizer__expand_source_arity_1(args[0])
}

func compiler_tokenizer__slurp_text_arity_1(path flagrt.Value) flagrt.Value {
	return func() flagrt.Value {
		var lines = flagrt.FileToStrings(path)
		var acc = flagrt.NilValue()
		for {
			var if_result_1075 flagrt.Value
			if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsEmpty(lines))) {
				var if_result_1072 flagrt.Value
				if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsNil(acc))) {
					if_result_1072 = flagStr_
				} else {
					if_result_1072 = acc
				}
				if_result_1075 = if_result_1072
			} else {
				var let_result_1074 string
				{
					var line = flagrt.Str(flagrt.First(lines))
					var if_result_1073 string
					if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsNil(acc))) {
						if_result_1073 = line
					} else {
						if_result_1073 = flagrt.Str(acc, "\n", line)
					}
					let_result_1074 = if_result_1073
				}
				if_result_1075 = flagrt.NewRecur(flagrt.Rest(lines), flagrt.NewString(let_result_1074))
			}
			__loopResult := if_result_1075
			if __recurValues, __isRecur := flagrt.UnwrapRecur(__loopResult); __isRecur {
				if len(__recurValues) != 2 {
					panic("internal error: recur arity mismatch")
				}
				lines = __recurValues[0]
				acc = __recurValues[1]
				continue
			}
			return __loopResult
		}
	}()
}

func compiler_tokenizer__slurp_text_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler_tokenizer__slurp_text expects exactly 1 arguments")
	}
	return compiler_tokenizer__slurp_text_arity_1(args[0])
}

func compiler_tokenizer__expand_fixture_arity_1(name flagrt.Value) flagrt.Value {
	return flagrt.Call(compiler_tokenizer__expand_source, flagrt.Call(compiler_tokenizer__slurp_text, flagrt.NewString(flagrt.Str("examples/compiler_tokenizer/macros/", name, ".in"))))
}

func compiler_tokenizer__expand_fixture_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler_tokenizer__expand_fixture expects exactly 1 arguments")
	}
	return compiler_tokenizer__expand_fixture_arity_1(args[0])
}

func compiler_tokenizer__expected_fixture_arity_1(name flagrt.Value) flagrt.Value {
	return flagrt.Call(compiler_tokenizer__slurp_text, flagrt.NewString(flagrt.Str("examples/compiler_tokenizer/macros/", name, ".expected")))
}

func compiler_tokenizer__expected_fixture_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler_tokenizer__expected_fixture expects exactly 1 arguments")
	}
	return compiler_tokenizer__expected_fixture_arity_1(args[0])
}

func compiler_tokenizer__lift_file_to_array_arity_1(path flagrt.Value) flagrt.Value {
	return flagrt.Call(compiler_tokenizer__drain_tokens, flagrt.Call(compiler_lift__lift_file, path))
}

func compiler_tokenizer__lift_file_to_array_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler_tokenizer__lift_file_to_array expects exactly 1 arguments")
	}
	return compiler_tokenizer__lift_file_to_array_arity_1(args[0])
}

func compiler_tokenizer__compile_file_to_go_exprs_arity_1(path flagrt.Value) flagrt.Value {
	return flagrt.Call(compiler_tokenizer__drain_tokens, flagrt.Call(compiler_codegen__compile_file, path))
}

func compiler_tokenizer__compile_file_to_go_exprs_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler_tokenizer__compile_file_to_go_exprs expects exactly 1 arguments")
	}
	return compiler_tokenizer__compile_file_to_go_exprs_arity_1(args[0])
}

func compiler_tokenizer__ast_file_to_source_arity_1(path flagrt.Value) flagrt.Value {
	return flagrt.Call(compiler_ast_source__asts__gtsource, flagrt.Call(compiler_tokenizer__build_ast_file_to_array, path))
}

func compiler_tokenizer__ast_file_to_source_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler_tokenizer__ast_file_to_source expects exactly 1 arguments")
	}
	return compiler_tokenizer__ast_file_to_source_arity_1(args[0])
}

func compiler_tokenizer__ast_node_to_source_arity_1(node flagrt.Value) flagrt.Value {
	return flagrt.Call(compiler_ast_source__ast__gtsource, node)
}

func compiler_tokenizer__ast_node_to_source_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler_tokenizer__ast_node_to_source expects exactly 1 arguments")
	}
	return compiler_tokenizer__ast_node_to_source_arity_1(args[0])
}

func compiler_tokenizer__nth_token_arity_2(tokens flagrt.Value, n flagrt.Value) flagrt.Value {
	return flagrt.First(flagrt.Drop(n, tokens))
}

func compiler_tokenizer__nth_token_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 2 {
		panic("compiler_tokenizer__nth_token expects exactly 2 arguments")
	}
	return compiler_tokenizer__nth_token_arity_2(args[0], args[1])
}

func compiler_tokenizer__node_children_arity_1(node flagrt.Value) flagrt.Value {
	var if_result_1076 flagrt.Value
	if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagKw_map, flagrt.Call(flagKw_kind, node)))) {
		if_result_1076 = flagrt.Call(flagKw_entries, node)
	} else {
		if_result_1076 = flagrt.Call(flagKw_elements, node)
	}
	return if_result_1076
}

func compiler_tokenizer__node_children_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler_tokenizer__node_children expects exactly 1 arguments")
	}
	return compiler_tokenizer__node_children_arity_1(args[0])
}

func compiler_tokenizer__nth_child_arity_2(node flagrt.Value, n flagrt.Value) flagrt.Value {
	return flagrt.Call(compiler_tokenizer__nth_token, flagrt.Call(compiler_tokenizer__node_children, node), n)
}

func compiler_tokenizer__nth_child_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 2 {
		panic("compiler_tokenizer__nth_child expects exactly 2 arguments")
	}
	return compiler_tokenizer__nth_child_arity_2(args[0], args[1])
}

var inc = flagrt.NewFunction(inc_variadic)
var dec = flagrt.NewFunction(dec_variadic)
var close = flagrt.NewFunction(close_variadic)
var peek = flagrt.BuiltinFunction("first")
var pop = flagrt.BuiltinFunction("rest")
var identity = flagrt.NewFunction(identity_variadic)
var constantly = flagrt.NewFunction(constantly_variadic)
var second = flagrt.NewFunction(second_variadic)
var third = flagrt.NewFunction(third_variadic)
var fourth = flagrt.NewFunction(fourth_variadic)
var fifth = flagrt.NewFunction(fifth_variadic)
var sixth = flagrt.NewFunction(sixth_variadic)
var seventh = flagrt.NewFunction(seventh_variadic)
var eighth = flagrt.NewFunction(eighth_variadic)
var ninth = flagrt.NewFunction(ninth_variadic)
var tenth = flagrt.NewFunction(tenth_variadic)
var val = flagrt.NewFunction(val_variadic)
var some_q = flagrt.NewFunction(some_q_variadic)
var true_q = flagrt.NewFunction(true_q_variadic)
var false_q = flagrt.NewFunction(false_q_variadic)
var boolean = flagrt.NewFunction(boolean_variadic)
var number_q = flagrt.NewFunction(number_q_variadic)
var int_q = flagrt.NewFunction(int_q_variadic)
var float_q = flagrt.NewFunction(float_q_variadic)
var string_q = flagrt.NewFunction(string_q_variadic)
var symbol_q = flagrt.NewFunction(symbol_q_variadic)
var keyword_q = flagrt.NewFunction(keyword_q_variadic)
var map_q = flagrt.NewFunction(map_q_variadic)
var vector_q = flagrt.NewFunction(vector_q_variadic)
var set_q = flagrt.NewFunction(set_q_variadic)
var sequential_q = flagrt.NewFunction(sequential_q_variadic)
var coll_q = flagrt.NewFunction(coll_q_variadic)
var empty = flagrt.NewFunction(empty_variadic)
var zero_q = flagrt.NewFunction(zero_q_variadic)
var pos_q = flagrt.NewFunction(pos_q_variadic)
var neg_q = flagrt.NewFunction(neg_q_variadic)
var even_q = flagrt.NewFunction(even_q_variadic)
var odd_q = flagrt.NewFunction(odd_q_variadic)
var not_empty_q = flagrt.NewFunction(not_empty_q_variadic)
var not_any_q = flagrt.NewFunction(not_any_q_variadic)
var remove = flagrt.NewFunction(remove_variadic)
var update = flagrt.NewFunction(update_variadic)
var get_in = flagrt.NewFunction(get_in_variadic)
var update_in = flagrt.NewFunction(update_in_variadic)
var assoc_in = flagrt.NewFunction(assoc_in_variadic)
var dissoc_in = flagrt.NewFunction(dissoc_in_variadic)
var juxt = flagrt.NewFunction(juxt_variadic)
var partial = flagrt.NewFunction(partial_variadic)
var complement = flagrt.NewFunction(complement_variadic)
var fnil = flagrt.NewFunction(fnil_variadic)
var comp = flagrt.NewFunction(comp_variadic)
var iterate = flagrt.NewFunction(iterate_variadic)
var range_ = flagrt.NewFunction(range__variadic)
var repeatedly = flagrt.NewFunction(repeatedly_variadic)
var take_while = flagrt.NewFunction(take_while_variadic)
var drop_while = flagrt.NewFunction(drop_while_variadic)
var take_last = flagrt.NewFunction(take_last_variadic)
var drop_last = flagrt.NewFunction(drop_last_variadic)
var take_nth = flagrt.NewFunction(take_nth_variadic)
var keep = flagrt.NewFunction(keep_variadic)
var map_indexed = flagrt.NewFunction(map_indexed_variadic)
var keep_indexed = flagrt.NewFunction(keep_indexed_variadic)
var reduce_kv = flagrt.NewFunction(reduce_kv_variadic)
var every_q = flagrt.NewFunction(every_q_variadic)
var every_pred = flagrt.NewFunction(every_pred_variadic)
var some_fn = flagrt.NewFunction(some_fn_variadic)
var max_key = flagrt.NewFunction(max_key_variadic)
var zipmap = flagrt.NewFunction(zipmap_variadic)
var group_by = flagrt.NewFunction(group_by_variadic)
var select_keys = flagrt.NewFunction(select_keys_variadic)
var merge = flagrt.NewFunction(merge_variadic)
var merge_with = flagrt.NewFunction(merge_with_variadic)
var mapcat = flagrt.NewFunction(mapcat_variadic)
var distinct = flagrt.NewFunction(distinct_variadic)
var flatten = flagrt.NewFunction(flatten_variadic)
var interpose = flagrt.NewFunction(interpose_variadic)
var interleave = flagrt.NewFunction(interleave_variadic)
var partition = flagrt.NewFunction(partition_variadic)
var partition_all = flagrt.NewFunction(partition_all_variadic)
var partition_by = flagrt.NewFunction(partition_by_variadic)
var sort = flagrt.NewFunction(sort_variadic)
var async__channel_map = flagrt.GoBind_async_PipeMap
var async__channel_reduce = flagrt.GoBind_async_PipeReduce
var async__deref = flagrt.GoBind_async_Deref
var async__swap_bang = flagrt.GoBind_async_Swap
var async__channel_close = flagrt.GoBind_async_ChannelClose
var async__channel_filter = flagrt.GoBind_async_PipeFilter
var async__atom = flagrt.GoBind_async_Atom
var async__future_run = flagrt.GoBind_async_FutureRun
var async__future_piped_run = flagrt.GoBind_async_FuturePipeRun
var async__reset_bang = flagrt.GoBind_async_Reset
var async__make_channel = flagrt.GoBind_async_MakeChannel
var async__channel_send = flagrt.GoBind_async_ChannelSend
var async__channel_receive = flagrt.GoBind_async_ChannelReceive
var async__channel_lines = flagrt.GoBind_async_LinesPipe
var async__go_run = flagrt.GoBind_async_GoRun
var async__sleep = flagrt.GoBind_async_Sleep
var async__select_ = flagrt.GoBind_async_Select
var async__channel_every_q = flagrt.NewFunction(async__channel_every_q_variadic)
var async__channel_some_q = flagrt.NewFunction(async__channel_some_q_variadic)
var stdlib__second = flagrt.NewFunction(stdlib__second_variadic)
var stdlib__third = flagrt.NewFunction(stdlib__third_variadic)
var compiler__whitespace_chars = flagSet
var compiler__delimiter_chars = flagSet_1
var compiler__token_eof = flagrt.NewLong(0)
var compiler__token_error = flagrt.NewLong(1)
var compiler__token_list_open = flagrt.NewLong(2)
var compiler__token_list_close = flagrt.NewLong(3)
var compiler__token_vector_open = flagrt.NewLong(4)
var compiler__token_vector_close = flagrt.NewLong(5)
var compiler__token_map_open = flagrt.NewLong(6)
var compiler__token_map_close = flagrt.NewLong(7)
var compiler__token_metadata = flagrt.NewLong(8)
var compiler__token_quote = flagrt.NewLong(9)
var compiler__token_dispatch_set_open = flagrt.NewLong(10)
var compiler__token_dispatch_fn_open = flagrt.NewLong(11)
var compiler__token_string = flagrt.NewLong(12)
var compiler__token_atom = flagrt.NewLong(13)
var compiler__token_pipe = flagrt.NewLong(14)
var compiler__whitespace_char_q = flagrt.NewFunction(compiler__whitespace_char_q_variadic)
var compiler__delimiter_char_q = flagrt.NewFunction(compiler__delimiter_char_q_variadic)
var compiler____gtSourceToken = flagrt.NewFunction(compiler____gtSourceToken_variadic)
var compiler__map__gtSourceToken = flagrt.NewFunction(compiler__map__gtSourceToken_variadic)
var compiler____gtTokenState = flagrt.NewFunction(compiler____gtTokenState_variadic)
var compiler__map__gtTokenState = flagrt.NewFunction(compiler__map__gtTokenState_variadic)
var compiler____gtParseToken = flagrt.NewFunction(compiler____gtParseToken_variadic)
var compiler__map__gtParseToken = flagrt.NewFunction(compiler__map__gtParseToken_variadic)
var compiler__emit_token_bang = flagrt.NewFunction(compiler__emit_token_bang_variadic)
var compiler__make_state = flagrt.NewFunction(compiler__make_state_variadic)
var compiler__empty_token_state = flagrt.Call(compiler__make_state, flagStr_, flagrt.NewLong(0), flagrt.NewLong(0), flagrt.NewBool(false), flagrt.NewBool(false), flagrt.NewBool(false))
var compiler__starts_triple_quote_q = flagrt.NewFunction(compiler__starts_triple_quote_q_variadic)
var compiler__starts_dispatch_set_q = flagrt.NewFunction(compiler__starts_dispatch_set_q_variadic)
var compiler__starts_dispatch_fn_q = flagrt.NewFunction(compiler__starts_dispatch_fn_q_variadic)
var compiler__strip_quoted_token = flagrt.NewFunction(compiler__strip_quoted_token_variadic)
var compiler__token__gtstring_value = flagrt.NewFunction(compiler__token__gtstring_value_variadic)
var compiler__source_token__gtparse_token = flagrt.NewFunction(compiler__source_token__gtparse_token_variadic)
var compiler__token_end_col = flagrt.NewFunction(compiler__token_end_col_variadic)
var compiler__tokenize_line_step_bang = flagrt.NewFunction(compiler__tokenize_line_step_bang_variadic)
var compiler__tokenize_lines_bang = flagrt.NewFunction(compiler__tokenize_lines_bang_variadic)
var compiler__split_lines = flagrt.NewFunction(compiler__split_lines_variadic)
var compiler__tokenize_source = flagrt.NewFunction(compiler__tokenize_source_variadic)
var compiler__tokenize_file = flagrt.NewFunction(compiler__tokenize_file_variadic)
var compiler__tokenize_file_parse_tokens = flagrt.NewFunction(compiler__tokenize_file_parse_tokens_variadic)
var compiler__parse_error = flagrt.NewFunction(compiler__parse_error_variadic)
var compiler__token_line = flagrt.NewFunction(compiler__token_line_variadic)
var compiler__token_col = flagrt.NewFunction(compiler__token_col_variadic)
var compiler__make_parser = flagrt.NewFunction(compiler__make_parser_variadic)
var compiler__peek = flagrt.NewFunction(compiler__peek_variadic)
var compiler__next_tok = flagrt.NewFunction(compiler__next_tok_variadic)
var compiler__eof_line_col = flagrt.NewFunction(compiler__eof_line_col_variadic)
var compiler__starts_with_char_q = flagrt.NewFunction(compiler__starts_with_char_q_variadic)
var compiler__ends_with_char_q = flagrt.NewFunction(compiler__ends_with_char_q_variadic)
var compiler__contains_float_mark_q = flagrt.NewFunction(compiler__contains_float_mark_q_variadic)
var compiler__slash_count = flagrt.NewFunction(compiler__slash_count_variadic)
var compiler__split_once = flagrt.NewFunction(compiler__split_once_variadic)
var compiler__unescape_char = flagrt.NewFunction(compiler__unescape_char_variadic)
var compiler__unquote_go_string = flagrt.NewFunction(compiler__unquote_go_string_variadic)
var compiler__string_token_q = flagrt.NewFunction(compiler__string_token_q_variadic)
var compiler__strip_edges = flagrt.NewFunction(compiler__strip_edges_variadic)
var compiler__triple_quoted_q = flagrt.NewFunction(compiler__triple_quoted_q_variadic)
var compiler__triple_quoted_closed_q = flagrt.NewFunction(compiler__triple_quoted_closed_q_variadic)
var compiler__parse_char_token = flagrt.NewFunction(compiler__parse_char_token_variadic)
var compiler__signed_decimal_int_q = flagrt.NewFunction(compiler__signed_decimal_int_q_variadic)
var compiler__parse_atom_token = flagrt.NewFunction(compiler__parse_atom_token_variadic)
var compiler__comment_list_q = flagrt.NewFunction(compiler__comment_list_q_variadic)
var compiler__collection_node = flagrt.NewFunction(compiler__collection_node_variadic)
var compiler__dispatch_error = flagrt.NewFunction(compiler__dispatch_error_variadic)
var compiler__read_string_token = flagrt.NewFunction(compiler__read_string_token_variadic)
var compiler__read_expr = flagrt.NewFunction(compiler__read_expr_variadic)
var compiler__ast_node__gtcanonical = flagrt.NewFunction(compiler__ast_node__gtcanonical_variadic)
var compiler__asts__gtcanonical = flagrt.NewFunction(compiler__asts__gtcanonical_variadic)
var compiler__drain_remaining = flagrt.NewFunction(compiler__drain_remaining_variadic)
var compiler__build_ast_from_tokens = flagrt.NewFunction(compiler__build_ast_from_tokens_variadic)
var compiler__symbol_node_q = flagrt.NewFunction(compiler__symbol_node_q_variadic)
var compiler__list_node_q = flagrt.NewFunction(compiler__list_node_q_variadic)
var compiler__vector_node_q = flagrt.NewFunction(compiler__vector_node_q_variadic)
var compiler__pipe_vector_node_q = flagrt.NewFunction(compiler__pipe_vector_node_q_variadic)
var compiler__map_node_q = flagrt.NewFunction(compiler__map_node_q_variadic)
var compiler__set_node_q = flagrt.NewFunction(compiler__set_node_q_variadic)
var compiler__hash_fn_node_q = flagrt.NewFunction(compiler__hash_fn_node_q_variadic)
var compiler__meta_node_q = flagrt.NewFunction(compiler__meta_node_q_variadic)
var compiler__string_node_q = flagrt.NewFunction(compiler__string_node_q_variadic)
var compiler__unwrap_meta = flagrt.NewFunction(compiler__unwrap_meta_variadic)
var compiler__node_children = flagrt.NewFunction(compiler__node_children_variadic)
var compiler__with_children = flagrt.NewFunction(compiler__with_children_variadic)
var compiler__splice_node_q = flagrt.NewFunction(compiler__splice_node_q_variadic)
var compiler__walk_node_q = flagrt.NewFunction(compiler__walk_node_q_variadic)
var compiler__make_symbol = flagrt.NewFunction(compiler__make_symbol_variadic)
var compiler__make_list = flagrt.NewFunction(compiler__make_list_variadic)
var compiler__wrap_literal = flagrt.NewFunction(compiler__wrap_literal_variadic)
var compiler__unwrap_literal = flagrt.NewFunction(compiler__unwrap_literal_variadic)
var compiler__unwrap_literal_tree = flagrt.NewFunction(compiler__unwrap_literal_tree_variadic)
var compiler__macro_definition_q = flagrt.NewFunction(compiler__macro_definition_q_variadic)
var compiler__append_all = flagrt.NewFunction(compiler__append_all_variadic)
var compiler__nth_node = flagrt.NewFunction(compiler__nth_node_variadic)
var compiler__node_equal_q = flagrt.NewFunction(compiler__node_equal_q_variadic)
var compiler__parse_macro_params = flagrt.NewFunction(compiler__parse_macro_params_variadic)
var compiler__compile_multi_arity = flagrt.NewFunction(compiler__compile_multi_arity_variadic)
var compiler__compile_defmacro = flagrt.NewFunction(compiler__compile_defmacro_variadic)
var compiler__bind_macro_args = flagrt.NewFunction(compiler__bind_macro_args_variadic)
var compiler__literal_kind_q = flagrt.NewFunction(compiler__literal_kind_q_variadic)
var compiler__macro_case_clause = flagrt.NewFunction(compiler__macro_case_clause_variadic)
var compiler__expand_defrecord_macro = flagrt.NewFunction(compiler__expand_defrecord_macro_variadic)
var compiler__macro_op = flagrt.NewFunction(compiler__macro_op_variadic)
var compiler__substitute_node = flagrt.NewFunction(compiler__substitute_node_variadic)
var compiler__apply_macro_arity = flagrt.NewFunction(compiler__apply_macro_arity_variadic)
var compiler__apply_multi_arity = flagrt.NewFunction(compiler__apply_multi_arity_variadic)
var compiler__apply_macro = flagrt.NewFunction(compiler__apply_macro_variadic)
var compiler__expand_node = flagrt.NewFunction(compiler__expand_node_variadic)
var compiler__register_macro = flagrt.NewFunction(compiler__register_macro_variadic)
var compiler__expand_macros = flagrt.NewFunction(compiler__expand_macros_variadic)
var compiler__parse_file = flagrt.NewFunction(compiler__parse_file_variadic)
var compiler_lift__lift_node = flagrt.NewFunction(compiler_lift__lift_node_variadic)
var compiler_lift__pump_lift_bang = flagrt.NewFunction(compiler_lift__pump_lift_bang_variadic)
var compiler_lift__lift_ast_channel = flagrt.NewFunction(compiler_lift__lift_ast_channel_variadic)
var compiler_lift__lift_file = flagrt.NewFunction(compiler_lift__lift_file_variadic)
var compiler_ir__runtime_alias = flagStr_flagrt
var compiler_ir__ir_ident = flagrt.NewFunction(compiler_ir__ir_ident_variadic)
var compiler_ir__ir_string = flagrt.NewFunction(compiler_ir__ir_string_variadic)
var compiler_ir__ir_int = flagrt.NewFunction(compiler_ir__ir_int_variadic)
var compiler_ir__ir_selector = flagrt.NewFunction(compiler_ir__ir_selector_variadic)
var compiler_ir__ir_call = flagrt.NewFunction(compiler_ir__ir_call_variadic)
var compiler_ir__ir_raw = flagrt.NewFunction(compiler_ir__ir_raw_variadic)
var compiler_ir__ir_index = flagrt.NewFunction(compiler_ir__ir_index_variadic)
var compiler_ir__ir_slice = flagrt.NewFunction(compiler_ir__ir_slice_variadic)
var compiler_ir__ir_spread = flagrt.NewFunction(compiler_ir__ir_spread_variadic)
var compiler_ir__ir_unary = flagrt.NewFunction(compiler_ir__ir_unary_variadic)
var compiler_ir__ir_binary = flagrt.NewFunction(compiler_ir__ir_binary_variadic)
var compiler_ir__ir_func_lit = flagrt.NewFunction(compiler_ir__ir_func_lit_variadic)
var compiler_ir__ir_expr_stmt = flagrt.NewFunction(compiler_ir__ir_expr_stmt_variadic)
var compiler_ir__ir_return = flagrt.NewFunction(compiler_ir__ir_return_variadic)
var compiler_ir__ir_defer = flagrt.NewFunction(compiler_ir__ir_defer_variadic)
var compiler_ir__ir_go = flagrt.NewFunction(compiler_ir__ir_go_variadic)
var compiler_ir__ir_var = flagrt.NewFunction(compiler_ir__ir_var_variadic)
var compiler_ir__ir_assign = flagrt.NewFunction(compiler_ir__ir_assign_variadic)
var compiler_ir__ir_define = flagrt.NewFunction(compiler_ir__ir_define_variadic)
var compiler_ir__ir_if = flagrt.NewFunction(compiler_ir__ir_if_variadic)
var compiler_ir__ir_if_init = flagrt.NewFunction(compiler_ir__ir_if_init_variadic)
var compiler_ir__ir_for = flagrt.NewFunction(compiler_ir__ir_for_variadic)
var compiler_ir__ir_block = flagrt.NewFunction(compiler_ir__ir_block_variadic)
var compiler_ir__ir_raw_stmt = flagrt.NewFunction(compiler_ir__ir_raw_stmt_variadic)
var compiler_ir__rt_call = flagrt.NewFunction(compiler_ir__rt_call_variadic)
var compiler_ir__ident_call = flagrt.NewFunction(compiler_ir__ident_call_variadic)
var compiler_ir__selector_call = flagrt.NewFunction(compiler_ir__selector_call_variadic)
var compiler_ir__parse_go_fun = flagrt.NewFunction(compiler_ir__parse_go_fun_variadic)
var compiler_ir__iife = flagrt.NewFunction(compiler_ir__iife_variadic)
var compiler_ir__value_iife = flagrt.NewFunction(compiler_ir__value_iife_variadic)
var compiler_ir__blank_q = flagrt.NewFunction(compiler_ir__blank_q_variadic)
var compiler_ir__empty_seq_q = flagrt.NewFunction(compiler_ir__empty_seq_q_variadic)
var compiler_ir__join_names = flagrt.NewFunction(compiler_ir__join_names_variadic)
var compiler_ir__render_ir_node = flagrt.NewFunction(compiler_ir__render_ir_node_variadic)
var compiler_ir__render_ir = flagrt.NewFunction(compiler_ir__render_ir_variadic)
var compiler_ir__render_ir_stmt = flagrt.NewFunction(compiler_ir__render_ir_stmt_variadic)
var compiler_ir__render_ir_stmts = flagrt.NewFunction(compiler_ir__render_ir_stmts_variadic)
var compiler_codegen__binary_ops = flagSet_2
var compiler_codegen__symbol_node_q = flagrt.NewFunction(compiler_codegen__symbol_node_q_variadic)
var compiler_codegen__list_node_q = flagrt.NewFunction(compiler_codegen__list_node_q_variadic)
var compiler_codegen__emit_ir_expr = flagrt.NewFunction(compiler_codegen__emit_ir_expr_variadic)
var compiler_codegen__emit_go_expr = flagrt.NewFunction(compiler_codegen__emit_go_expr_variadic)
var compiler_codegen__emit_go_loop_bang = flagrt.NewFunction(compiler_codegen__emit_go_loop_bang_variadic)
var compiler_codegen__emit_go_channel = flagrt.NewFunction(compiler_codegen__emit_go_channel_variadic)
var compiler_codegen__compile_file = flagrt.NewFunction(compiler_codegen__compile_file_variadic)
var compiler__literal_rt_call = flagrt.NewFunction(compiler__literal_rt_call_variadic)
var compiler__ast_node_to_ir = flagrt.NewFunction(compiler__ast_node_to_ir_variadic)
var compiler__quoted_ast_to_ir_node = flagrt.NewFunction(compiler__quoted_ast_to_ir_node_variadic)
var compiler__quoted_ast_to_ir = flagrt.NewFunction(compiler__quoted_ast_to_ir_variadic)
var compiler__special_form_names = flagSet_3
var compiler__call_to_ir = flagrt.NewFunction(compiler__call_to_ir_variadic)
var compiler__call_ast_to_ir_node = flagrt.NewFunction(compiler__call_ast_to_ir_node_variadic)
var compiler__call_ast_to_ir = flagrt.NewFunction(compiler__call_ast_to_ir_variadic)
var compiler__ctor_to_ir = flagrt.NewFunction(compiler__ctor_to_ir_variadic)
var compiler__runtime_call_to_ir = flagrt.NewFunction(compiler__runtime_call_to_ir_variadic)
var compiler__fold_call_to_ir = flagrt.NewFunction(compiler__fold_call_to_ir_variadic)
var compiler__builtin_function_names = flagSet_4
var compiler__unary_runtime_ops = flagMap_4
var compiler__fold_runtime_ops = flagMap_5
var compiler__binary_runtime_ops = flagMap_6
var compiler__variadic_runtime_ops = flagMap_7
var compiler__qualified_name_q = flagrt.NewFunction(compiler__qualified_name_q_variadic)
var compiler__ident_ir = flagrt.NewFunction(compiler__ident_ir_variadic)
var compiler__symbol_to_ir = flagrt.NewFunction(compiler__symbol_to_ir_variadic)
var compiler__eval_ast_to_ir_node = flagrt.NewFunction(compiler__eval_ast_to_ir_node_variadic)
var compiler__eval_ast_to_ir = flagrt.NewFunction(compiler__eval_ast_to_ir_variadic)
var compiler__stmt_list = flagrt.NewFunction(compiler__stmt_list_variadic)
var compiler__append_stmt = flagrt.NewFunction(compiler__append_stmt_variadic)
var compiler__if_to_ir = flagrt.NewFunction(compiler__if_to_ir_variadic)
var compiler__concat_stmts = flagrt.NewFunction(compiler__concat_stmts_variadic)
var compiler__let_to_ir = flagrt.NewFunction(compiler__let_to_ir_variadic)
var compiler_ast_source__ast__gtsource = flagrt.NewFunction(compiler_ast_source__ast__gtsource_variadic)
var compiler_ast_source__asts__gtsource = flagrt.NewFunction(compiler_ast_source__asts__gtsource_variadic)
var compiler_tokenizer__drain_tokens_into = flagrt.NewFunction(compiler_tokenizer__drain_tokens_into_variadic)
var compiler_tokenizer__drain_tokens = flagrt.NewFunction(compiler_tokenizer__drain_tokens_variadic)
var compiler_tokenizer__tokenize_file_to_array = flagrt.NewFunction(compiler_tokenizer__tokenize_file_to_array_variadic)
var compiler_tokenizer__parse_file_to_array = flagrt.NewFunction(compiler_tokenizer__parse_file_to_array_variadic)
var compiler_tokenizer__build_ast_file_to_array = flagrt.NewFunction(compiler_tokenizer__build_ast_file_to_array_variadic)
var compiler_tokenizer__build_ast_source_to_array = flagrt.NewFunction(compiler_tokenizer__build_ast_source_to_array_variadic)
var compiler_tokenizer__expand_file_to_array = flagrt.NewFunction(compiler_tokenizer__expand_file_to_array_variadic)
var compiler_tokenizer__expand_source_to_array = flagrt.NewFunction(compiler_tokenizer__expand_source_to_array_variadic)
var compiler_tokenizer__expand_source = flagrt.NewFunction(compiler_tokenizer__expand_source_variadic)
var compiler_tokenizer__slurp_text = flagrt.NewFunction(compiler_tokenizer__slurp_text_variadic)
var compiler_tokenizer__expand_fixture = flagrt.NewFunction(compiler_tokenizer__expand_fixture_variadic)
var compiler_tokenizer__expected_fixture = flagrt.NewFunction(compiler_tokenizer__expected_fixture_variadic)
var compiler_tokenizer__lift_file_to_array = flagrt.NewFunction(compiler_tokenizer__lift_file_to_array_variadic)
var compiler_tokenizer__compile_file_to_go_exprs = flagrt.NewFunction(compiler_tokenizer__compile_file_to_go_exprs_variadic)
var compiler_tokenizer__ast_file_to_source = flagrt.NewFunction(compiler_tokenizer__ast_file_to_source_variadic)
var compiler_tokenizer__ast_node_to_source = flagrt.NewFunction(compiler_tokenizer__ast_node_to_source_variadic)
var compiler_tokenizer__nth_token = flagrt.NewFunction(compiler_tokenizer__nth_token_variadic)
var compiler_tokenizer__node_children = flagrt.NewFunction(compiler_tokenizer__node_children_variadic)
var compiler_tokenizer__nth_child = flagrt.NewFunction(compiler_tokenizer__nth_child_variadic)
var flagSet = flagrt.NewSet(flagrt.NewString(" "), flagrt.NewString("\t"), flagrt.NewString("\n"), flagrt.NewString("\r"), flagrt.NewString(","))
var flagSet_1 = flagrt.NewSet(flagrt.NewString("("), flagrt.NewString(")"), flagrt.NewString("["), flagrt.NewString("]"), flagrt.NewString("{"), flagrt.NewString("}"), flagrt.NewString("^"), flagrt.NewString("'"), flagrt.NewString("|"))
var flagKw_token = flagrt.NewKeyword("token")
var flagKw_line = flagrt.NewKeyword("line")
var flagKw_offset = flagrt.NewKeyword("offset")
var flagKw_start_line = flagrt.NewKeyword("start-line")
var flagKw_start_offset = flagrt.NewKeyword("start-offset")
var flagKw_in_string = flagrt.NewKeyword("in-string")
var flagKw_triple = flagrt.NewKeyword("triple")
var flagKw_escaped = flagrt.NewKeyword("escaped")
var flagKw_kind = flagrt.NewKeyword("kind")
var flagKw_lexeme = flagrt.NewKeyword("lexeme")
var flagKw_string = flagrt.NewKeyword("string")
var flagKw_message = flagrt.NewKeyword("message")
var flagKw_col = flagrt.NewKeyword("col")
var flagStr_ = flagrt.NewString("")
var flagStr_unexpected_end_after__ = flagrt.NewString("unexpected end after #")
var flagStr____ = flagrt.NewString("\"\"\"")
var flagStr__ = flagrt.NewString("\"")
var flagStr___ = flagrt.NewString("#{")
var flagStr____1 = flagrt.NewString("#(")
var flagVec = flagrt.NewArray()
var flagKw_in = flagrt.NewKeyword("in")
var flagKw_la = flagrt.NewKeyword("la")
var flagKw_none = flagrt.NewKeyword("none")
var flagKw_last = flagrt.NewKeyword("last")
var flagKw_eof = flagrt.NewKeyword("eof")
var flagVec_1 = flagrt.NewArray(flagrt.NewLong(1), flagrt.NewLong(1))
var flagStr___1 = flagrt.NewString("\r")
var flagStr___2 = flagrt.NewString("\t")
var flagStr___3 = flagrt.NewString("\n")
var flagKw_error = flagrt.NewKeyword("error")
var flagKw_char = flagrt.NewKeyword("char")
var flagKw_value = flagrt.NewKeyword("value")
var flagStr_unsupported_character_li = flagrt.NewString("unsupported character literal")
var flagStr___4 = flagrt.NewString(":")
var flagKw_keyword = flagrt.NewKeyword("keyword")
var flagKw_name = flagrt.NewKeyword("name")
var flagStr___5 = flagrt.NewString("\\")
var flagStr___6 = flagrt.NewString("/")
var flagStr_ratio_denominator_cannot = flagrt.NewString("ratio denominator cannot be zero")
var flagKw_ratio = flagrt.NewKeyword("ratio")
var flagKw_numerator = flagrt.NewKeyword("numerator")
var flagKw_denominator = flagrt.NewKeyword("denominator")
var flagKw_symbol = flagrt.NewKeyword("symbol")
var flagStr_N = flagrt.NewString("N")
var flagKw_bigint = flagrt.NewKeyword("bigint")
var flagKw_int = flagrt.NewKeyword("int")
var flagKw_float = flagrt.NewKeyword("float")
var flagKw_raw = flagrt.NewKeyword("raw")
var flagKw_map = flagrt.NewKeyword("map")
var flagKw_entries = flagrt.NewKeyword("entries")
var flagKw_elements = flagrt.NewKeyword("elements")
var flagStr_unsupported_reader_dispa = flagrt.NewString("unsupported reader dispatch")
var flagStr_unterminated_string_lite = flagrt.NewString("unterminated string literal")
var flagStr_unexpected_end_of_input = flagrt.NewString("unexpected end of input")
var flagKw_list = flagrt.NewKeyword("list")
var flagKw_vector = flagrt.NewKeyword("vector")
var flagKw_pipe_vector = flagrt.NewKeyword("pipe-vector")
var flagKw_set = flagrt.NewKeyword("set")
var flagStr___7 = flagrt.NewString("}")
var flagStr___8 = flagrt.NewString("|")
var flagStr___9 = flagrt.NewString("]")
var flagStr___10 = flagrt.NewString(")")
var flagStr_missing_closing____ = flagrt.NewString("missing closing \"}\"")
var flagStr_missing_closing_____1 = flagrt.NewString("missing closing \"|\"")
var flagStr_missing_closing_____2 = flagrt.NewString("missing closing \"]\"")
var flagStr_missing_closing_____3 = flagrt.NewString("missing closing \")\"")
var flagKw_hash_fn = flagrt.NewKeyword("hash-fn")
var flagKw_body = flagrt.NewKeyword("body")
var flagKw_meta = flagrt.NewKeyword("meta")
var flagKw_target = flagrt.NewKeyword("target")
var flagKw_quoted_symbol = flagrt.NewKeyword("quoted-symbol")
var flagKw_quoted_list = flagrt.NewKeyword("quoted-list")
var flagStr_quote_currently_supports = flagrt.NewString("quote currently supports symbols and lists")
var flagStr_expected_expression = flagrt.NewString("expected expression")
var flagStr___11 = flagrt.NewString("#")
var flagKw_macro_literal = flagrt.NewKeyword("macro-literal")
var flagKw_inner = flagrt.NewKeyword("inner")
var flagKw_form = flagrt.NewKeyword("form")
var flagKw_data = flagrt.NewKeyword("data")
var flagKw_params = flagrt.NewKeyword("params")
var flagKw_rest_param = flagrt.NewKeyword("rest-param")
var flagMap = flagrt.NewMap()
var flagKw_arities = flagrt.NewKeyword("arities")
var flagStr_do = flagrt.NewString("do")
var flagKw_expected = flagrt.NewKeyword("expected")
var flagKw_got = flagrt.NewKeyword("got")
var flagKw_values = flagrt.NewKeyword("values")
var flagKw_rest_bindings = flagrt.NewKeyword("rest-bindings")
var flagKw_pattern = flagrt.NewKeyword("pattern")
var flagStr_defrecord_ = flagrt.NewString("defrecord*")
var flagKw_seq = flagrt.NewKeyword("seq")
var flagKw_matched = flagrt.NewKeyword("matched")
var flagKw_bindings = flagrt.NewKeyword("bindings")
var flagMap_1 = flagrt.NewMap(flagKw_matched, flagrt.NewBool(false))
var flagKw_pat = flagrt.NewKeyword("pat")
var flagKw_case = flagrt.NewKeyword("case")
var flagKw_subst = flagrt.NewKeyword("subst")
var flagKw_op = flagrt.NewKeyword("op")
var flagKw_expr = flagrt.NewKeyword("expr")
var flagStr_flagrt = flagrt.NewString("flagrt")
var flagKw_ident = flagrt.NewKeyword("ident")
var flagKw_selector = flagrt.NewKeyword("selector")
var flagKw_pkg = flagrt.NewKeyword("pkg")
var flagKw_call = flagrt.NewKeyword("call")
var flagKw_fun = flagrt.NewKeyword("fun")
var flagKw_args = flagrt.NewKeyword("args")
var flagKw_code = flagrt.NewKeyword("code")
var flagKw_index = flagrt.NewKeyword("index")
var flagKw_x = flagrt.NewKeyword("x")
var flagKw_slice = flagrt.NewKeyword("slice")
var flagKw_low = flagrt.NewKeyword("low")
var flagKw_high = flagrt.NewKeyword("high")
var flagKw_spread = flagrt.NewKeyword("spread")
var flagKw_unary = flagrt.NewKeyword("unary")
var flagKw_binary = flagrt.NewKeyword("binary")
var flagKw_left = flagrt.NewKeyword("left")
var flagKw_right = flagrt.NewKeyword("right")
var flagKw_func_lit = flagrt.NewKeyword("func-lit")
var flagKw_result = flagrt.NewKeyword("result")
var flagKw_expr_stmt = flagrt.NewKeyword("expr-stmt")
var flagKw_discard = flagrt.NewKeyword("discard")
var flagKw_return = flagrt.NewKeyword("return")
var flagKw_defer = flagrt.NewKeyword("defer")
var flagKw_go = flagrt.NewKeyword("go")
var flagKw_var = flagrt.NewKeyword("var")
var flagKw_type = flagrt.NewKeyword("type")
var flagKw_assign = flagrt.NewKeyword("assign")
var flagKw_define = flagrt.NewKeyword("define")
var flagKw_names = flagrt.NewKeyword("names")
var flagKw_if = flagrt.NewKeyword("if")
var flagKw_init = flagrt.NewKeyword("init")
var flagKw_cond = flagrt.NewKeyword("cond")
var flagKw_then = flagrt.NewKeyword("then")
var flagKw_else = flagrt.NewKeyword("else")
var flagKw_for = flagrt.NewKeyword("for")
var flagKw_block = flagrt.NewKeyword("block")
var flagKw_raw_stmt = flagrt.NewKeyword("raw-stmt")
var flagStr___12 = flagrt.NewString(".")
var flagKw_stmts = flagrt.NewKeyword("stmts")
var flagKw_stmt = flagrt.NewKeyword("stmt")
var flagKw_node = flagrt.NewKeyword("node")
var flagKw_mode = flagrt.NewKeyword("mode")
var flagSet_2 = flagrt.NewSet(flagrt.NewString("+"), flagrt.NewString("-"), flagrt.NewString("*"), flagrt.NewString("/"))
var flagStr_NewString = flagrt.NewString("NewString")
var flagStr_NewLong = flagrt.NewString("NewLong")
var flagStr_NewBigIntFromString = flagrt.NewString("NewBigIntFromString")
var flagStr_NewRatio = flagrt.NewString("NewRatio")
var flagStr_NewDouble = flagrt.NewString("NewDouble")
var flagStr_NewKeyword = flagrt.NewString("NewKeyword")
var flagStr_NewSymbol = flagrt.NewString("NewSymbol")
var flagStr_NewBool = flagrt.NewString("NewBool")
var flagMap_2 = flagrt.NewMap(flagKw_kind, flagKw_raw, flagKw_code, flagrt.NewString("true"))
var flagVec_2 = flagrt.NewArray(flagMap_2)
var flagMap_3 = flagrt.NewMap(flagKw_kind, flagKw_raw, flagKw_code, flagrt.NewString("false"))
var flagVec_3 = flagrt.NewArray(flagMap_3)
var flagStr_NilValue = flagrt.NewString("NilValue")
var flagStr_NewList = flagrt.NewString("NewList")
var flagStr_NewArray = flagrt.NewString("NewArray")
var flagStr_NewVector = flagrt.NewString("NewVector")
var flagStr_NewMap = flagrt.NewString("NewMap")
var flagStr_NewSet = flagrt.NewString("NewSet")
var flagSet_3 = flagrt.NewSet(flagrt.NewString("+"), flagrt.NewString("*"), flagrt.NewString("-"), flagrt.NewString("/"), flagrt.NewString("%"), flagrt.NewString("mod"), flagrt.NewString("quot"), flagrt.NewString("rem"), flagrt.NewString("compare"), flagrt.NewString("=="), flagrt.NewString("="), flagrt.NewString("<"), flagrt.NewString("<="), flagrt.NewString(">"), flagrt.NewString(">="), flagrt.NewString("max"), flagrt.NewString("min"), flagrt.NewString("bit-and"), flagrt.NewString("bit-or"), flagrt.NewString("bit-xor"), flagrt.NewString("bit-not"), flagrt.NewString("bit-shift-left"), flagrt.NewString("bit-shift-right"), flagrt.NewString("unsigned-bit-shift-right"), flagrt.NewString("bit-test"), flagrt.NewString("bit-set"), flagrt.NewString("bit-clear"), flagrt.NewString("bit-flip"), flagrt.NewString("str"), flagrt.NewString("println"), flagrt.NewString("testing"), flagrt.NewString("ex-info"), flagrt.NewString("ex-message"), flagrt.NewString("ex-data"), flagrt.NewString("ex-cause"), flagrt.NewString("throw"), flagrt.NewString("try"), flagrt.NewString("catch"), flagrt.NewString("finally"), flagrt.NewString("is"), flagrt.NewString("expect-exception"), flagrt.NewString("if"), flagrt.NewString("do"), flagrt.NewString("doto"), flagrt.NewString("for"), flagrt.NewString("doseq"), flagrt.NewString("let"), flagrt.NewString("loop"), flagrt.NewString("recur"), flagrt.NewString("defer"), flagrt.NewString("update!"), flagrt.NewString("symbol"), flagrt.NewString("name"), flagrt.NewString("keyword"), flagrt.NewString("first"), flagrt.NewString("fist"), flagrt.NewString("rest"), flagrt.NewString("next"), flagrt.NewString("last"), flagrt.NewString("reverse"), flagrt.NewString("cons"), flagrt.NewString("take"), flagrt.NewString("drop"), flagrt.NewString("not-empty"), flagrt.NewString("seq"), flagrt.NewString("empty?"), flagrt.NewString("nil?"), flagrt.NewString("type-of"), flagrt.NewString("count"), flagrt.NewString("double"), flagrt.NewString("into"), flagrt.NewString("format"), flagrt.NewString("hash-map"), flagrt.NewString("list"), flagrt.NewString("array"), flagrt.NewString("map"), flagrt.NewString("concat"), flagrt.NewString("sort-by"), flagrt.NewString("apply"), flagrt.NewString("pmap"), flagrt.NewString("filter"), flagrt.NewString("reduce"), flagrt.NewString("doall"), flagrt.NewString("some"), flagrt.NewString("seq?"), flagrt.NewString("set"), flagrt.NewString("vec"), flagrt.NewString("conj"), flagrt.NewString("contains?"), flagrt.NewString("line-seq"), flagrt.NewString("repeat"), flagrt.NewString("rand-int"), flagrt.NewString("assoc"), flagrt.NewString("dissoc"), flagrt.NewString("to-json"), flagrt.NewString("from-json"), flagrt.NewString("open-file"), flagrt.NewString("file-to-strings"), flagrt.NewString("go-fn"), flagrt.NewString("go-fn-args"), flagrt.NewString("fn"))
var flagStr_Call = flagrt.NewString("Call")
var flagKw_eval = flagrt.NewKeyword("eval")
var flagSet_4 = flagrt.NewSet(flagrt.NewString("+"), flagrt.NewString("-"), flagrt.NewString("*"), flagrt.NewString("/"), flagrt.NewString("%"), flagrt.NewString("quot"), flagrt.NewString("rem"), flagrt.NewString("mod"), flagrt.NewString("compare"), flagrt.NewString("=="), flagrt.NewString("="), flagrt.NewString("<"), flagrt.NewString("<="), flagrt.NewString(">"), flagrt.NewString(">="), flagrt.NewString("max"), flagrt.NewString("min"), flagrt.NewString("bit-and"), flagrt.NewString("bit-or"), flagrt.NewString("bit-xor"), flagrt.NewString("bit-not"), flagrt.NewString("bit-shift-left"), flagrt.NewString("bit-shift-right"), flagrt.NewString("unsigned-bit-shift-right"), flagrt.NewString("bit-test"), flagrt.NewString("bit-set"), flagrt.NewString("bit-clear"), flagrt.NewString("bit-flip"), flagrt.NewString("first"), flagrt.NewString("fist"), flagrt.NewString("rest"), flagrt.NewString("next"), flagrt.NewString("last"), flagrt.NewString("reverse"), flagrt.NewString("cons"), flagrt.NewString("take"), flagrt.NewString("drop"), flagrt.NewString("nth"), flagrt.NewString("slow-nth"), flagrt.NewString("map"), flagrt.NewString("concat"), flagrt.NewString("sort-by"), flagrt.NewString("apply"), flagrt.NewString("pmap"), flagrt.NewString("filter"), flagrt.NewString("reduce"), flagrt.NewString("get"), flagrt.NewString("keys"), flagrt.NewString("vals"), flagrt.NewString("find"), flagrt.NewString("hash-map"), flagrt.NewString("list"), flagrt.NewString("array"), flagrt.NewString("not-empty"), flagrt.NewString("empty?"), flagrt.NewString("nil?"), flagrt.NewString("type-of"), flagrt.NewString("count"), flagrt.NewString("double"), flagrt.NewString("numerator"), flagrt.NewString("denominator"), flagrt.NewString("format"), flagrt.NewString("subs"), flagrt.NewString("keyword"), flagrt.NewString("into"), flagrt.NewString("doall"), flagrt.NewString("dorun"), flagrt.NewString("line-seq"), flagrt.NewString("some"), flagrt.NewString("seq"), flagrt.NewString("seq?"), flagrt.NewString("set"), flagrt.NewString("vec"), flagrt.NewString("conj"), flagrt.NewString("contains?"), flagrt.NewString("assoc"), flagrt.NewString("dissoc"), flagrt.NewString("open-file"), flagrt.NewString("close-file"), flagrt.NewString("close-channel"), flagrt.NewString("file-to-strings"), flagrt.NewString("rand-int"), flagrt.NewString("rand"), flagrt.NewString("rand-nth"), flagrt.NewString("shuffle"), flagrt.NewString("repeat"), flagrt.NewString("union"), flagrt.NewString("intersection"), flagrt.NewString("difference"), flagrt.NewString("subset?"), flagrt.NewString("superset?"), flagrt.NewString("disjoint?"), flagrt.NewString("rename-keys"), flagrt.NewString("map-invert"), flagrt.NewString("select"), flagrt.NewString("project"), flagrt.NewString("rename"), flagrt.NewString("go-fn"), flagrt.NewString("go-fn-args"), flagrt.NewString("re-pattern"), flagrt.NewString("re-matches"), flagrt.NewString("ex-message"), flagrt.NewString("ex-data"), flagrt.NewString("ex-cause"))
var flagMap_4 = flagrt.NewMap(flagrt.NewString("first"), flagrt.NewString("First"), flagrt.NewString("fist"), flagrt.NewString("First"), flagrt.NewString("rest"), flagrt.NewString("Rest"), flagrt.NewString("next"), flagrt.NewString("Next"), flagrt.NewString("last"), flagrt.NewString("Last"), flagrt.NewString("reverse"), flagrt.NewString("Reverse"), flagrt.NewString("seq?"), flagrt.NewString("SeqPredicate"), flagrt.NewString("double"), flagrt.NewString("Double"), flagrt.NewString("bit-not"), flagrt.NewString("BitNot"))
var flagMap_5 = flagrt.NewMap(flagrt.NewString("+"), flagrt.NewString("Add"), flagrt.NewString("*"), flagrt.NewString("Mul"), flagrt.NewString("-"), flagrt.NewString("Sub"), flagrt.NewString("/"), flagrt.NewString("Div"))
var flagMap_6 = flagrt.NewMap(flagrt.NewString("%"), flagrt.NewString("Mod"), flagrt.NewString("mod"), flagrt.NewString("Mod"), flagrt.NewString("quot"), flagrt.NewString("Quot"), flagrt.NewString("rem"), flagrt.NewString("Rem"), flagrt.NewString("bit-shift-left"), flagrt.NewString("BitShiftLeft"), flagrt.NewString("bit-shift-right"), flagrt.NewString("BitShiftRight"), flagrt.NewString("unsigned-bit-shift-right"), flagrt.NewString("UnsignedBitShiftRight"), flagrt.NewString("bit-test"), flagrt.NewString("BitTest"), flagrt.NewString("bit-set"), flagrt.NewString("BitSet"), flagrt.NewString("bit-clear"), flagrt.NewString("BitClear"), flagrt.NewString("bit-flip"), flagrt.NewString("BitFlip"))
var flagMap_7 = flagrt.NewMap(flagrt.NewString("bit-and"), flagrt.NewString("BitAnd"), flagrt.NewString("bit-or"), flagrt.NewString("BitOr"), flagrt.NewString("bit-xor"), flagrt.NewString("BitXor"), flagrt.NewString("max"), flagrt.NewString("Max"), flagrt.NewString("min"), flagrt.NewString("Min"))
var flagKw_ir = flagrt.NewKeyword("ir")
var flagKw_expr_kind = flagrt.NewKeyword("expr-kind")
var flagKw_bool = flagrt.NewKeyword("bool")
var flagKw_locals = flagrt.NewKeyword("locals")
var flagKw_globals = flagrt.NewKeyword("globals")
var flagKw_functions = flagrt.NewKeyword("functions")
var flagKw_module = flagrt.NewKeyword("module")
var flagKw_go_fns = flagrt.NewKeyword("go-fns")
var flagKw_self_function_name = flagrt.NewKeyword("self-function-name")
var flagKw_self_variadic_name = flagrt.NewKeyword("self-variadic-name")
var flagStr_NewFunction = flagrt.NewString("NewFunction")
var flagStr_BuiltinFunction = flagrt.NewString("BuiltinFunction")
var flagStr__tab = flagrt.NewString("\\tab")
var flagStr__newline = flagrt.NewString("\\newline")
var flagStr__space = flagrt.NewString("\\space")

func main() {
	_ = strings.HasSuffix
}
