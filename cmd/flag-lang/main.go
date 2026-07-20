package main

import (
	"fmt"
	"os"

	"flag-lang/internal/compiler"
	"flag-lang/internal/repl"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "flag-lang: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return usageError("missing command")
	}

	switch args[0] {
	case "compile":
		return runCompile(args[1:])
	case "repl":
		return runRepl(args[1:])
	case "help", "-h", "--help":
		printUsage()
		return nil
	default:
		return usageError("unknown command %q", args[0])
	}
}

func runRepl(args []string) error {
	if len(args) > 0 {
		return usageError("repl does not take arguments")
	}
	return repl.Run(os.Stdin, os.Stdout)
}

func runCompile(args []string) error {
	outputPath := ""
	inputPath := ""

	for index := 0; index < len(args); index++ {
		arg := args[index]
		switch arg {
		case "-h", "--help", "help":
			printUsage()
			return nil
		case "-o", "--output":
			index++
			if index >= len(args) {
				return usageError("missing value for %s", arg)
			}
			outputPath = args[index]
		default:
			if len(arg) > 0 && arg[0] == '-' {
				return usageError("unknown flag %q", arg)
			}
			if inputPath != "" {
				return usageError("compile expects exactly one input file")
			}
			inputPath = arg
		}
	}

	if inputPath == "" {
		return usageError("compile expects exactly one input file")
	}

	source, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("read %s: %w", inputPath, err)
	}

	goSource, err := compiler.Compile(string(source))
	if err != nil {
		return err
	}

	if outputPath == "" {
		_, err = os.Stdout.Write(goSource)
		return err
	}

	if err := os.WriteFile(outputPath, goSource, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", outputPath, err)
	}

	return nil
}

func usageError(format string, args ...any) error {
	printUsage()
	return fmt.Errorf(format, args...)
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintln(os.Stderr, "  flag-lang compile [-o output.go] <input.flag>")
	fmt.Fprintln(os.Stderr, "  flag-lang repl")
}
