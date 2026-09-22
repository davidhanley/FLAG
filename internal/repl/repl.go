package repl

//go:generate go run flag-lang/internal/repl/symbolgen

import (
	"bufio"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"

	"flag-lang/internal/compiler"
	flagrt "flag-lang/runtime"

	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

func Run(input io.Reader, output io.Writer) error {
	i := interp.New(interp.Options{Stderr: io.Discard})
	i.Use(stdlib.Symbols)
	flagrt.RegisterGoSymbols(stdlib.Symbols)
	if err := i.Use(runtimeSymbols()); err != nil {
		return fmt.Errorf("load runtime symbols: %w", err)
	}
	flagrt.RegisterGoSymbols(runtimeSymbols())
	if _, err := i.Eval(`import flagrt "flagrt/flagrt"`); err != nil {
		return fmt.Errorf("import runtime symbols: %w", err)
	}

	lineCompiler := compiler.NewReplCompiler()
	scanner := bufio.NewScanner(input)
	counter := 0
	if err := evalCompiled(i, output, lineCompiler.PrologueSetup(), &counter); err != nil {
		return fmt.Errorf("load compiler prologue: %w", err)
	}
	var bufferedSource strings.Builder
	for {
		fmt.Fprint(output, "flag> ")
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return err
			}
			return nil
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if bufferedSource.Len() == 0 && strings.HasPrefix(line, ":") {
			handled, err := runCommand(lineCompiler, i, output, scanner.Text(), &counter)
			if err != nil {
				fmt.Fprintf(output, "error: %v\n", err)
			}
			if handled {
				if line == ":quit" || line == ":exit" {
					return nil
				}
				continue
			}
		}

		if bufferedSource.Len() > 0 {
			bufferedSource.WriteByte('\n')
		}
		bufferedSource.WriteString(scanner.Text())

		complete, err := replInputComplete(bufferedSource.String())
		if err != nil {
			fmt.Fprintf(output, "error: %v\n", err)
			bufferedSource.Reset()
			continue
		}
		if !complete {
			continue
		}

		source := bufferedSource.String()
		bufferedSource.Reset()

		compiled, err := lineCompiler.CompileLine(strings.TrimSpace(source))
		if err != nil {
			fmt.Fprintf(output, "error: %v\n", err)
			continue
		}
		if compiled.Setup == "" && compiled.ResultExpr == "" {
			continue
		}
		if err := evalCompiled(i, output, compiled, &counter); err != nil {
			fmt.Fprintf(output, "error: %v\n", err)
			continue
		}
	}
}

func runCommand(lineCompiler *compiler.ReplCompiler, i *interp.Interpreter, output io.Writer, rawLine string, counter *int) (bool, error) {
	line := strings.TrimSpace(rawLine)
	switch {
	case line == ":quit" || line == ":exit":
		return true, nil
	case line == ":help":
		fmt.Fprintln(output, ":import <spec>  import a module or library into the REPL")
		fmt.Fprintln(output, ":load <path>    load and evaluate FLAG code from a file")
		fmt.Fprintln(output, ":quit           exit the REPL")
		return true, nil
	case strings.HasPrefix(line, ":import "):
		compiled, err := lineCompiler.ImportSpec(strings.TrimSpace(rawLine[len(":import "):]), "")
		if err != nil {
			return true, err
		}
		executeCompiledBatch(i, output, compiled, counter)
		return true, nil
	case line == ":import":
		return true, fmt.Errorf(":import expects a module spec")
	case strings.HasPrefix(line, ":load "):
		path, err := parseLoadPath(strings.TrimSpace(rawLine[len(":load "):]))
		if err != nil {
			return true, err
		}
		compiled, err := lineCompiler.LoadFile(path)
		if err != nil {
			return true, err
		}
		executeCompiledBatch(i, output, compiled, counter)
		return true, nil
	case line == ":load":
		return true, fmt.Errorf(":load expects a file path")
	default:
		return false, nil
	}
}

func executeCompiledBatch(i *interp.Interpreter, output io.Writer, compiled []compiler.ReplCompiled, counter *int) bool {
	for _, step := range compiled {
		if err := evalCompiled(i, output, step, counter); err != nil {
			fmt.Fprintf(output, "error: %v\n", err)
			return false
		}
	}
	return true
}

func evalCompiled(i *interp.Interpreter, output io.Writer, compiled compiler.ReplCompiled, counter *int) error {
	if compiled.Setup != "" {
		setupParts := strings.Split(compiled.Setup, ";;")
		for _, setupPart := range setupParts {
			setupPart = strings.TrimSpace(setupPart)
			if setupPart == "" {
				continue
			}
			if _, err := i.Eval(setupPart); err != nil {
				return err
			}
		}
	}
	if compiled.ResultExpr == "" {
		return nil
	}

	fnName := fmt.Sprintf("__flagEval%d", *counter)
	*counter = *counter + 1
	if _, err := i.Eval(fmt.Sprintf("func %s() any { return %s }", fnName, compiled.ResultExpr)); err != nil {
		return err
	}

	result, err := i.Eval(fnName + "()")
	if err != nil {
		return err
	}
	fmt.Fprintln(output, flagrt.Str(result.Interface()))
	return nil
}

func parseLoadPath(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("empty load path")
	}
	if !strings.HasPrefix(raw, "\"") {
		return raw, nil
	}
	unquoted, err := strconv.Unquote(raw)
	if err != nil {
		return "", err
	}
	return unquoted, nil
}

func replInputComplete(source string) (bool, error) {
	_, err := compiler.ParseFile(source)
	if err == nil {
		return true, nil
	}

	msg := err.Error()
	switch {
	case strings.Contains(msg, "missing closing"),
		strings.Contains(msg, "unterminated string literal"),
		strings.Contains(msg, "unexpected end of input"),
		strings.Contains(msg, "unexpected end after quote"),
		strings.Contains(msg, "unexpected end after #"):
		return false, nil
	default:
		return false, err
	}
}

func runtimeSymbols() map[string]map[string]reflect.Value {
	return map[string]map[string]reflect.Value{
		"flagrt/flagrt": generatedRuntimeSymbols(),
	}
}
