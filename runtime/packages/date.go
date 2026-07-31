package packages

import "time"

func RegisterDate(register func(string, any)) {
	register("date/from-string", DateFromString)
	register("c/from-string", DateFromString)
}

func DateFromString(value string) any {
	layouts := []string{
		"2006-1-2",
		"2006-01-02",
	}
	for _, layout := range layouts {
		parsed, err := time.ParseInLocation(layout, value, time.UTC)
		if err == nil {
			return parsed
		}
	}
	return nil
}
