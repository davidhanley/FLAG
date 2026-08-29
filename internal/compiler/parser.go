package compiler

func ParseFile(source string) (FileAST, error) {
	return ParseTokenChannel(TokenizeSourceToChannel(source))
}

func isSignedDecimalInteger(token string) bool {
	if token == "" {
		return false
	}
	start := 0
	if token[0] == '+' || token[0] == '-' {
		if len(token) == 1 {
			return false
		}
		start = 1
	}
	for i := start; i < len(token); i++ {
		if token[i] < '0' || token[i] > '9' {
			return false
		}
	}
	return true
}
