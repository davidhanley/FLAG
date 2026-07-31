package runtime

import runtimepackages "flag-lang/runtime/packages"

func init() {
	runtimepackages.RegisterString(RegisterGoFunction)
	runtimepackages.RegisterLong(RegisterGoFunction)
	runtimepackages.RegisterDate(RegisterGoFunction)
	runtimepackages.RegisterDateTime(RegisterGoFunction)
	runtimepackages.RegisterCharacter(RegisterGoFunction)
	RegisterGoFunction("regex/compile", RegexCompile)
	RegisterGoFunction("re-pattern", RegexCompile)
	RegisterGoFunction("re-matches", RegexMatches)
	RegisterGoFunction("io/reader", OpenFile)
	RegisterGoFunction("clojure.java.io/reader", OpenFile)
	RegisterGoFunction("io/readline", ReadLine)
	RegisterGoFunction("io/writer", OpenWriter)
	RegisterGoFunction("io/scan-directory", ScanDirectory)
	RegisterGoFunction("json/read", JSONRead)
	RegisterGoFunction("json/read-str", JSONReadStr)
	RegisterGoFunction("line-seq", LineSeq)
	RegisterGoFunction("math/abs", Abs)
	RegisterGoFunction("t/now", TimeNow)
	RegisterGoFunction("t/years", TimeYears)
	RegisterGoFunction("t/minus", TimeMinus)
	RegisterGoFunction("t/after?", TimeAfter)
	RegisterGoFunction("datetime/now", TimeNow)
	RegisterGoFunction("dateTime/now", TimeNow)
}
