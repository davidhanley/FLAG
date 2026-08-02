package packages

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	regexFlagCaseInsensitive = 0x02
	regexFlagMultiline       = 0x08
	regexFlagDotAll          = 0x20
	regexFlagLiteral         = 0x10
	regexSupportedFlags      = regexFlagCaseInsensitive | regexFlagMultiline | regexFlagDotAll | regexFlagLiteral
)

func CompileRegex(pattern string, flags ...int64) func(string) bool {
	if len(flags) > 1 {
		panic("regex/compile expects pattern and optional flags")
	}
	if len(flags) == 1 {
		pattern = applyRegexFlags(pattern, flags[0])
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		panic("regex/compile failed: " + err.Error())
	}
	return re.MatchString
}

func applyRegexFlags(pattern string, flags int64) string {
	unsupported := flags &^ int64(regexSupportedFlags)
	if unsupported != 0 {
		panic(fmt.Sprintf("regex/compile unsupported flags: %d", unsupported))
	}
	if flags&regexFlagLiteral != 0 {
		pattern = regexp.QuoteMeta(pattern)
	}
	modes := make([]string, 0, 3)
	if flags&regexFlagCaseInsensitive != 0 {
		modes = append(modes, "i")
	}
	if flags&regexFlagMultiline != 0 {
		modes = append(modes, "m")
	}
	if flags&regexFlagDotAll != 0 {
		modes = append(modes, "s")
	}
	if len(modes) == 0 {
		return pattern
	}
	return "(?" + strings.Join(modes, "") + ")" + pattern
}

func MatchesRegex(matcher func(string) bool, value string) bool {
	if matcher == nil {
		panic("re-matches expects a compiled regex")
	}
	return matcher(value)
}
