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
