package runtime

import (
	packages "flag-lang/runtime/packages"
)

func adaptGoBind_packages_StringIncludes(args ...Value) Value {
	goArgArityExact("str/includes?", args, 2)
	a0 := goArgString("str/includes?", 0, args[0])
	a1 := goArgString("str/includes?", 1, args[1])
	return goRetBool(packages.StringIncludes(a0, a1))
}

// GoBind_packages_StringIncludes dispatches str/includes?, string/includes?.
var GoBind_packages_StringIncludes = NewFunction(adaptGoBind_packages_StringIncludes)

func adaptGoBind_packages_StringIndexOf(args ...Value) Value {
	goArgArityAtLeast("str/index-of", args, 2)
	if len(args) > 3 {
		panic("str/index-of expects 2 or 3 arguments")
	}
	a0 := goArgString("str/index-of", 0, args[0])
	a1 := goArgString("str/index-of", 1, args[1])
	rest := make([]int64, 0, len(args)-2)
	for i := 2; i < len(args); i++ {
		rest = append(rest, goArgInt64("str/index-of", i, args[i]))
	}
	return goRetAny(packages.StringIndexOf(a0, a1, rest...))
}

// GoBind_packages_StringIndexOf dispatches str/index-of, string/index-of.
var GoBind_packages_StringIndexOf = NewFunction(adaptGoBind_packages_StringIndexOf)

func adaptGoBind_packages_StringLastIndexOf(args ...Value) Value {
	goArgArityAtLeast("str/last-index-of", args, 2)
	if len(args) > 3 {
		panic("str/last-index-of expects 2 or 3 arguments")
	}
	a0 := goArgString("str/last-index-of", 0, args[0])
	a1 := goArgString("str/last-index-of", 1, args[1])
	rest := make([]int64, 0, len(args)-2)
	for i := 2; i < len(args); i++ {
		rest = append(rest, goArgInt64("str/last-index-of", i, args[i]))
	}
	return goRetAny(packages.StringLastIndexOf(a0, a1, rest...))
}

// GoBind_packages_StringLastIndexOf dispatches str/last-index-of, string/last-index-of.
var GoBind_packages_StringLastIndexOf = NewFunction(adaptGoBind_packages_StringLastIndexOf)

func adaptGoBind_packages_StringLowerCase(args ...Value) Value {
	goArgArityExact("str/lower-case", args, 1)
	a0 := goArgString("str/lower-case", 0, args[0])
	return goRetString(packages.StringLowerCase(a0))
}

// GoBind_packages_StringLowerCase dispatches str/lower-case, string/lower-case.
var GoBind_packages_StringLowerCase = NewFunction(adaptGoBind_packages_StringLowerCase)

func adaptGoBind_packages_StringReplaceFirst(args ...Value) Value {
	goArgArityExact("str/replace-first", args, 3)
	a0 := goArgString("str/replace-first", 0, args[0])
	a1 := goArgString("str/replace-first", 1, args[1])
	a2 := goArgString("str/replace-first", 2, args[2])
	return goRetString(packages.StringReplaceFirst(a0, a1, a2))
}

// GoBind_packages_StringReplaceFirst dispatches str/replace-first, string/replace-first.
var GoBind_packages_StringReplaceFirst = NewFunction(adaptGoBind_packages_StringReplaceFirst)

func adaptGoBind_packages_StringReverse(args ...Value) Value {
	goArgArityExact("str/reverse", args, 1)
	a0 := goArgString("str/reverse", 0, args[0])
	return goRetString(packages.StringReverse(a0))
}

// GoBind_packages_StringReverse dispatches str/reverse, string/reverse.
var GoBind_packages_StringReverse = NewFunction(adaptGoBind_packages_StringReverse)

func adaptGoBind_packages_StringSplitLines(args ...Value) Value {
	goArgArityExact("str/split-lines", args, 1)
	a0 := goArgString("str/split-lines", 0, args[0])
	return goRetStrings(packages.StringSplitLines(a0))
}

// GoBind_packages_StringSplitLines dispatches str/split-lines, string/split-lines.
var GoBind_packages_StringSplitLines = NewFunction(adaptGoBind_packages_StringSplitLines)

func adaptGoBind_packages_StringTrimNewline(args ...Value) Value {
	goArgArityExact("str/trim-newline", args, 1)
	a0 := goArgString("str/trim-newline", 0, args[0])
	return goRetString(packages.StringTrimNewline(a0))
}

// GoBind_packages_StringTrimNewline dispatches str/trim-newline, string/trim-newline.
var GoBind_packages_StringTrimNewline = NewFunction(adaptGoBind_packages_StringTrimNewline)

func adaptGoBind_packages_StringTriml(args ...Value) Value {
	goArgArityExact("str/triml", args, 1)
	a0 := goArgString("str/triml", 0, args[0])
	return goRetString(packages.StringTriml(a0))
}

// GoBind_packages_StringTriml dispatches str/triml, string/triml.
var GoBind_packages_StringTriml = NewFunction(adaptGoBind_packages_StringTriml)

func adaptGoBind_packages_StringTrimr(args ...Value) Value {
	goArgArityExact("str/trimr", args, 1)
	a0 := goArgString("str/trimr", 0, args[0])
	return goRetString(packages.StringTrimr(a0))
}

// GoBind_packages_StringTrimr dispatches str/trimr, string/trimr.
var GoBind_packages_StringTrimr = NewFunction(adaptGoBind_packages_StringTrimr)
