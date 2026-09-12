package compiler

import flagrt "flag-lang/runtime"

func ParseFile(source string) (FileAST, error) {
	return fileASTFromFLAGTokens(flagrt.Call(compiler__tokenize_source, flagrt.NewString(source)))
}

// ParseSourceToChannel tokenizes source and streams top-level AST forms.
func ParseSourceToChannel(source string) <-chan ASTForm {
	return drainFLAGAST(flagrt.Call(compiler__build_ast_from_tokens, flagrt.Call(compiler__tokenize_source, flagrt.NewString(source))))
}

// ParseFileToChannel tokenizes a source file and streams top-level AST forms.
func ParseFileToChannel(path string) <-chan ASTForm {
	return drainFLAGAST(flagrt.Call(compiler__build_ast_from_tokens, flagrt.Call(compiler__tokenize_file, flagrt.NewString(path))))
}
