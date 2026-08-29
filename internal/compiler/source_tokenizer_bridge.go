package compiler

import (
	flagrt "flag-lang/runtime"
)

// TokenizeFileToChannel tokenizes a source file through the FLAG tokenizer.
func TokenizeFileToChannel(path string) <-chan SourceToken {
	return drainFLAGSourceTokens(flagrt.Call(compiler__tokenize_file, flagrt.NewString(path)))
}

// TokenizeSourceToChannel tokenizes an in-memory source string through the FLAG tokenizer.
func TokenizeSourceToChannel(source string) <-chan SourceToken {
	return drainFLAGSourceTokens(flagrt.Call(compiler__tokenize_source, flagrt.NewString(source)))
}

func drainFLAGSourceTokens(flagTokens flagrt.Value) <-chan SourceToken {
	out := make(chan SourceToken, 32)
	go func() {
		defer close(out)
		for {
			sourceToken := flagrt.ChannelReceive(flagTokens)
			if flagrt.IsNil(sourceToken) {
				return
			}
			out <- flagrt.GoStruct[SourceToken](sourceToken)
		}
	}()
	return out
}
