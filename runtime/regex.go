package runtime

import runtimepackages "flag-lang/runtime/packages"

func RegexCompile(pattern any, options ...any) Value {
	if len(options) > 1 {
		panic("regex/compile expects pattern and optional flags or value")
	}
	if len(options) == 1 {
		if flags, ok := regexFlags(options[0]); ok {
			matcher := runtimepackages.CompileRegex(regexPattern(pattern), flags)
			return regexMatcherValue(matcher)
		}
		return NewBool(RegexMatches(pattern, regexValue(options[0])))
	}
	matcher := runtimepackages.CompileRegex(regexPattern(pattern))
	return regexMatcherValue(matcher)
}

func regexMatcherValue(matcher func(string) bool) Value {
	return NewFunction(func(args ...Value) Value {
		if len(args) != 1 {
			panic("regex matcher expects exactly one argument")
		}
		return NewBool(matcher(valueAsString(args[0])))
	})
}

func regexPattern(pattern any) string {
	switch compiled := pattern.(type) {
	case Value:
		if compiled.tag != TagString {
			panic("regex/compile expects string pattern")
		}
		return compiled.StringValue()
	case string:
		return compiled
	default:
		panic("regex/compile expects string pattern")
	}
}

func regexFlags(raw any) (int64, bool) {
	switch v := raw.(type) {
	case int:
		return int64(v), true
	case int8:
		return int64(v), true
	case int16:
		return int64(v), true
	case int32:
		return int64(v), true
	case int64:
		return v, true
	case uint:
		return int64(v), true
	case uint8:
		return int64(v), true
	case uint16:
		return int64(v), true
	case uint32:
		return int64(v), true
	case Value:
		if v.tag != TagLong {
			return 0, false
		}
		return v.Long(), true
	default:
		return 0, false
	}
}

func regexValue(raw any) string {
	switch v := raw.(type) {
	case string:
		return v
	case Value:
		return valueAsString(v)
	default:
		panic("regex/compile expects string match value")
	}
}

func RegexMatches(pattern any, value string) bool {
	switch compiled := pattern.(type) {
	case Value:
		if compiled.tag != TagFunction {
			panic("re-matches expects a function or pattern")
		}
		return IsTruthy(Call(compiled, NewString(value)))
	case string:
		return runtimepackages.MatchesRegex(runtimepackages.CompileRegex(compiled), value)
	default:
		panic("re-matches expects a function or pattern")
	}
}

func valueAsString(v Value) string {
	switch v.tag {
	case TagString:
		return v.StringValue()
	case TagSymbol:
		return v.SymbolObject().Name
	default:
		return ValueToString(v)
	}
}
