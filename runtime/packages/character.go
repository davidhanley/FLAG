package packages

import "strings"

func RegisterCharacter(register func(string, any)) {
	register("character/toUppercase", ToUppercase)
	register("character/toUpperCase", ToUppercase)
	register("Character/toUpperCase", ToUppercase)
}

func ToUppercase(value string) string {
	return strings.ToUpper(value)
}
