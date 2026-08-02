package packages

import (
	"strconv"
	"strings"
)

func RegisterLong(register func(string, any)) {
	register("long/parse", LongParse)
}

func LongParse(value string) any {
	parsed, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	if err != nil {
		return nil
	}
	return parsed
}
