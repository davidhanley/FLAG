package packages

import (
	"fmt"
	"strings"
	"time"
	"unicode"
)

func RegisterDateTime(register func(string, any)) {
	register("datetime/formatter", DateTimeFormatter)
	register("dateTime/formatter", DateTimeFormatter)
	register("datetime/unparse", DateTimeUnparse)
	register("dateTime/unparse", DateTimeUnparse)
}

func DateTimeFormatter(pattern string) (string, error) {
	if strings.TrimSpace(pattern) == "" {
		return "", fmt.Errorf("datetime formatter pattern cannot be empty")
	}

	type token struct {
		java string
		goof string
	}
	tokens := []token{
		{java: "yyyy", goof: "2006"},
		{java: "yy", goof: "06"},
		{java: "MM", goof: "01"},
		{java: "M", goof: "1"},
		{java: "dd", goof: "02"},
		{java: "d", goof: "2"},
		{java: "HH", goof: "15"},
		{java: "H", goof: "15"},
		{java: "mm", goof: "04"},
		{java: "m", goof: "4"},
		{java: "ss", goof: "05"},
		{java: "s", goof: "5"},
	}

	var layout strings.Builder
	for i := 0; i < len(pattern); {
		if pattern[i] == '\'' {
			if i+1 < len(pattern) && pattern[i+1] == '\'' {
				layout.WriteByte('\'')
				i += 2
				continue
			}
			end := i + 1
			for end < len(pattern) && pattern[end] != '\'' {
				end++
			}
			if end >= len(pattern) {
				return "", fmt.Errorf("datetime formatter has unmatched quote")
			}
			layout.WriteString(pattern[i+1 : end])
			i = end + 1
			continue
		}

		matched := false
		for _, tok := range tokens {
			if strings.HasPrefix(pattern[i:], tok.java) {
				layout.WriteString(tok.goof)
				i += len(tok.java)
				matched = true
				break
			}
		}
		if matched {
			continue
		}

		ch := rune(pattern[i])
		if unicode.IsLetter(ch) {
			return "", fmt.Errorf("unsupported datetime formatter token %q", string(ch))
		}
		layout.WriteByte(pattern[i])
		i++
	}

	return layout.String(), nil
}

func DateTimeUnparse(formatter string, value time.Time) (string, error) {
	if formatter == "" {
		return "", fmt.Errorf("datetime/unparse expects a formatter produced by datetime/formatter")
	}
	return value.UTC().Format(formatter), nil
}
