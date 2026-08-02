package repl

import (
	"bufio"
	"fmt"
	"io"
	"reflect"
	"strings"

	"flag-lang/internal/compiler"
	flagrt "flag-lang/runtime"

	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

func Run(input io.Reader, output io.Writer) error {
	i := interp.New(interp.Options{})
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
		if line == ":quit" || line == ":exit" {
			return nil
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
		if compiled.Setup != "" {
			setupParts := strings.Split(compiled.Setup, ";;")
			setupFailed := false
			for _, setupPart := range setupParts {
				setupPart = strings.TrimSpace(setupPart)
				if setupPart == "" {
					continue
				}
				if _, err := i.Eval(setupPart); err != nil {
					fmt.Fprintf(output, "error: %v\n", err)
					setupFailed = true
					break
				}
			}
			if setupFailed {
				continue
			}
		}
		if compiled.ResultExpr == "" {
			continue
		}

		fnName := fmt.Sprintf("__flagEval%d", counter)
		counter++
		if _, err := i.Eval(fmt.Sprintf("func %s() any { return %s }", fnName, compiled.ResultExpr)); err != nil {
			fmt.Fprintf(output, "error: %v\n", err)
			continue
		}

		result, err := i.Eval(fnName + "()")
		if err != nil {
			fmt.Fprintf(output, "error: %v\n", err)
			continue
		}
		fmt.Fprintln(output, flagrt.Str(result.Interface()))
	}
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
	symbols := map[string]map[string]reflect.Value{
		"flagrt/flagrt": {
			"Value":                      reflect.ValueOf((*flagrt.Value)(nil)),
			"NewLong":                    reflect.ValueOf(flagrt.NewLong),
			"NewDouble":                  reflect.ValueOf(flagrt.NewDouble),
			"NewRatio":                   reflect.ValueOf(flagrt.NewRatio),
			"NewRatioFromRat":            reflect.ValueOf(flagrt.NewRatioFromRat),
			"NewBool":                    reflect.ValueOf(flagrt.NewBool),
			"NewString":                  reflect.ValueOf(flagrt.NewString),
			"NewSymbol":                  reflect.ValueOf(flagrt.NewSymbol),
			"NewKeyword":                 reflect.ValueOf(flagrt.NewKeyword),
			"NewFunction":                reflect.ValueOf(flagrt.NewFunction),
			"NewMap":                     reflect.ValueOf(flagrt.NewMap),
			"NewSet":                     reflect.ValueOf(flagrt.NewSet),
			"Add":                        reflect.ValueOf(flagrt.Add),
			"Eq":                         reflect.ValueOf(flagrt.Eq),
			"Lt":                         reflect.ValueOf(flagrt.Lt),
			"Gt":                         reflect.ValueOf(flagrt.Gt),
			"IsTruthy":                   reflect.ValueOf(flagrt.IsTruthy),
			"NilValue":                   reflect.ValueOf(flagrt.NilValue),
			"Symbol":                     reflect.ValueOf(flagrt.Symbol),
			"Name":                       reflect.ValueOf(flagrt.Name),
			"Call":                       reflect.ValueOf(flagrt.Call),
			"BuiltinFunction":            reflect.ValueOf(flagrt.BuiltinFunction),
			"Assoc":                      reflect.ValueOf(flagrt.Assoc),
			"Dissoc":                     reflect.ValueOf(flagrt.Dissoc),
			"Get":                        reflect.ValueOf(flagrt.Get),
			"MapAssoc":                   reflect.ValueOf(flagrt.MapAssoc),
			"MapDissoc":                  reflect.ValueOf(flagrt.MapDissoc),
			"Str":                        reflect.ValueOf(flagrt.Str),
			"Println":                    reflect.ValueOf(flagrt.Println),
			"ValueToString":              reflect.ValueOf(flagrt.ValueToString),
			"Sub":                        reflect.ValueOf(flagrt.Sub),
			"Mul":                        reflect.ValueOf(flagrt.Mul),
			"Div":                        reflect.ValueOf(flagrt.Div),
			"Mod":                        reflect.ValueOf(flagrt.Mod),
			"ValueToAny":                 reflect.ValueOf(flagrt.ValueToAny),
			"NewList":                    reflect.ValueOf(flagrt.NewList),
			"ListCons":                   reflect.ValueOf(flagrt.ListCons),
			"ListRest":                   reflect.ValueOf(flagrt.ListRest),
			"ListAppend":                 reflect.ValueOf(flagrt.ListAppend),
			"First":                      reflect.ValueOf(flagrt.First),
			"Rest":                       reflect.ValueOf(flagrt.Rest),
			"SeqFirst":                   reflect.ValueOf(flagrt.SeqFirst),
			"SeqRest":                    reflect.ValueOf(flagrt.SeqRest),
			"Take":                       reflect.ValueOf(flagrt.Take),
			"Drop":                       reflect.ValueOf(flagrt.Drop),
			"Map":                        reflect.ValueOf(flagrt.Map),
			"PMap":                       reflect.ValueOf(flagrt.PMap),
			"Filter":                     reflect.ValueOf(flagrt.Filter),
			"Reduce":                     reflect.ValueOf(flagrt.Reduce),
			"Range":                      reflect.ValueOf(flagrt.Range),
			"RandInt":                    reflect.ValueOf(flagrt.RandInt),
			"OpenFile":                   reflect.ValueOf(flagrt.OpenFile),
			"FileToStrings":              reflect.ValueOf(flagrt.FileToStrings),
			"FileToStringsPath":          reflect.ValueOf(flagrt.FileToStringsPath),
			"RegisterGoFunction":         reflect.ValueOf(flagrt.RegisterGoFunction),
			"RegisterGoSymbols":          reflect.ValueOf(flagrt.RegisterGoSymbols),
			"RegisterGoPackageFunctions": reflect.ValueOf(flagrt.RegisterGoPackageFunctions),
			"GoFunction":                 reflect.ValueOf(flagrt.GoFunction),
			"GoFunctionArgs":             reflect.ValueOf(flagrt.GoFunctionArgs),
			"ToJSON":                     reflect.ValueOf(flagrt.ToJSON),
			"FromJSON":                   reflect.ValueOf(flagrt.FromJSON),
			"NewArray":                   reflect.ValueOf(flagrt.NewArray),
			"ArrayGet":                   reflect.ValueOf(flagrt.ArrayGet),
			"ArrayAssoc":                 reflect.ValueOf(flagrt.ArrayAssoc),
			"ArrayAppend":                reflect.ValueOf(flagrt.ArrayAppend),
			"ArrayRest":                  reflect.ValueOf(flagrt.ArrayRest),
		},
	}
	// Merge the generated static Go-function adapters (GoBind_*), which the
	// compiler emits for namespaced calls such as (str/trim s).
	for name, value := range flagrt.GoBindSymbols() {
		symbols["flagrt/flagrt"][name] = value
	}
	return symbols
}
