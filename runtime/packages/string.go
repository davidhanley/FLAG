package packages

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

func RegisterString(register func(string, any)) {
	register("string/trim", StringTrim)
	register("str/trim", StringTrim)
	register("string/replace", StringReplace)
	register("str/replace", StringReplace)
	register("string/escape", StringEscape)
	register("str/escape", StringEscape)
	register("string/split", StringSplit)
	register("str/split", StringSplit)
	register("string/join", StringJoin)
	register("str/join", StringJoin)
	register("string/blank?", StringBlank)
	register("str/blank?", StringBlank)
	register("string/starts-with?", StringStartsWith)
	register("str/starts-with?", StringStartsWith)
	register("string/ends-with?", StringEndsWith)
	register("str/ends-with?", StringEndsWith)
	register("string/upper-case", StringUpperCase)
	register("str/upper-case", StringUpperCase)
	register("string/capitalize", StringCapitalize)
	register("str/capitalize", StringCapitalize)
	register("string/lower-case", StringLowerCase)
	register("str/lower-case", StringLowerCase)
	register("string/triml", StringTriml)
	register("str/triml", StringTriml)
	register("string/trimr", StringTrimr)
	register("str/trimr", StringTrimr)
	register("string/trim-newline", StringTrimNewline)
	register("str/trim-newline", StringTrimNewline)
	register("string/includes?", StringIncludes)
	register("str/includes?", StringIncludes)
	register("string/index-of", StringIndexOf)
	register("str/index-of", StringIndexOf)
	register("string/last-index-of", StringLastIndexOf)
	register("str/last-index-of", StringLastIndexOf)
	register("string/replace-first", StringReplaceFirst)
	register("str/replace-first", StringReplaceFirst)
	register("string/split-lines", StringSplitLines)
	register("str/split-lines", StringSplitLines)
	register("string/reverse", StringReverse)
	register("str/reverse", StringReverse)
}

func StringTrim(value string) string {
	return strings.TrimSpace(value)
}

func StringReplace(value, old, new string) string {
	return strings.ReplaceAll(value, old, new)
}

func StringEscape(value string, cmap map[any]any) string {
	if value == "" || len(cmap) == 0 {
		return value
	}

	var builder strings.Builder
	for _, ch := range value {
		if replacement, ok := stringEscapeReplacement(cmap, ch); ok {
			builder.WriteString(replacement)
			continue
		}
		builder.WriteRune(ch)
	}
	return builder.String()
}

func stringEscapeReplacement(cmap map[any]any, ch rune) (string, bool) {
	if replacement, ok := cmap[string(ch)]; ok && replacement != nil {
		return fmt.Sprint(replacement), true
	}
	if replacement, ok := cmap[ch]; ok && replacement != nil {
		return fmt.Sprint(replacement), true
	}
	if replacement, ok := cmap[int(ch)]; ok && replacement != nil {
		return fmt.Sprint(replacement), true
	}
	if replacement, ok := cmap[int64(ch)]; ok && replacement != nil {
		return fmt.Sprint(replacement), true
	}
	return "", false
}

func StringSplit(value, sep string, limit ...int64) []string {
	if len(limit) == 0 {
		return strings.Split(value, sep)
	}
	if limit[0] <= 0 {
		return strings.Split(value, sep)
	}
	return strings.SplitN(value, sep, int(limit[0]))
}

func StringJoin(args ...any) string {
	switch len(args) {
	case 1:
		return strings.Join(stringJoinValues(args[0]), "")
	case 2:
		sep, ok := args[0].(string)
		if !ok {
			panic("str/join expects separator string")
		}
		return strings.Join(stringJoinValues(args[1]), sep)
	default:
		panic("str/join expects collection or separator and collection")
	}
}

func stringJoinValues(raw any) []string {
	if raw == nil {
		return []string{}
	}
	switch values := raw.(type) {
	case []string:
		return values
	case []any:
		out := make([]string, len(values))
		for i, value := range values {
			if value == nil {
				out[i] = ""
				continue
			}
			out[i] = fmt.Sprint(value)
		}
		return out
	default:
		panic("str/join expects a collection of strings")
	}
}

func StringBlank(value string) bool {
	return strings.TrimSpace(value) == ""
}

func StringStartsWith(value, prefix string) bool {
	return strings.HasPrefix(value, prefix)
}

func StringEndsWith(value, suffix string) bool {
	return strings.HasSuffix(value, suffix)
}

func StringUpperCase(value string) string {
	return strings.ToUpper(value)
}

func StringCapitalize(value string) string {
	if value == "" {
		return ""
	}
	r, size := utf8.DecodeRuneInString(value)
	if r == utf8.RuneError && size == 0 {
		return ""
	}
	return string(unicode.ToUpper(r)) + strings.ToLower(value[size:])
}

func StringLowerCase(value string) string {
	return strings.ToLower(value)
}

func StringTriml(value string) string {
	return strings.TrimLeftFunc(value, unicode.IsSpace)
}

func StringTrimr(value string) string {
	return strings.TrimRightFunc(value, unicode.IsSpace)
}

func StringTrimNewline(value string) string {
	return strings.TrimRight(value, "\n\r")
}

func StringIncludes(value, substr string) bool {
	return strings.Contains(value, substr)
}

func StringIndexOf(value, substr string, from ...int64) any {
	if len(from) > 1 {
		panic("str/index-of expects 2 or 3 arguments")
	}
	runes := []rune(value)
	needle := []rune(substr)
	start := 0
	if len(from) == 1 {
		if from[0] > 0 {
			start = int(from[0])
		}
	}
	n := len(runes)
	if start > n {
		if len(needle) == 0 {
			return int64(n)
		}
		return nil
	}
	if len(needle) == 0 {
		return int64(start)
	}
	limit := n - len(needle)
	for i := start; i <= limit; i++ {
		if runeSliceEqual(runes[i:i+len(needle)], needle) {
			return int64(i)
		}
	}
	return nil
}

func StringLastIndexOf(value, substr string, from ...int64) any {
	if len(from) > 1 {
		panic("str/last-index-of expects 2 or 3 arguments")
	}
	runes := []rune(value)
	needle := []rune(substr)
	n := len(runes)
	fromIdx := n
	if len(from) == 1 {
		if from[0] < 0 {
			return nil
		}
		fromIdx = int(from[0])
	}
	if len(needle) == 0 {
		if fromIdx > n {
			fromIdx = n
		}
		return int64(fromIdx)
	}
	maxStart := n - len(needle)
	if fromIdx < maxStart {
		maxStart = fromIdx
	}
	if maxStart < 0 {
		return nil
	}
	for i := maxStart; i >= 0; i-- {
		if runeSliceEqual(runes[i:i+len(needle)], needle) {
			return int64(i)
		}
	}
	return nil
}

func StringReplaceFirst(value, old, new string) string {
	return strings.Replace(value, old, new, 1)
}

func StringSplitLines(value string) []string {
	normalized := strings.ReplaceAll(value, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	parts := strings.Split(normalized, "\n")
	for len(parts) > 1 && parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}
	if len(parts) == 1 && parts[0] == "" && value != "" {
		return []string{}
	}
	return parts
}

func StringReverse(value string) string {
	runes := []rune(value)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func runeSliceEqual(a, b []rune) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
