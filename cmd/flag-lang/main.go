package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

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
	case "build":
		return runBuild(args[1:])
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

func runBuild(args []string) error {
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
				return usageError("build expects exactly one input file or directory")
			}
			inputPath = arg
		}
	}

	if inputPath == "" {
		return usageError("build expects exactly one input file or directory")
	}

	source, err := readBuildSource(inputPath)
	if err != nil {
		return err
	}
	goSource, err := compiler.Compile(string(source))
	if err != nil {
		return err
	}

	if outputPath == "" {
		base := filepath.Base(inputPath)
		outputPath = strings.TrimSuffix(base, filepath.Ext(base))
	}

	tempDir, err := os.MkdirTemp(".", ".flag-build-*")
	if err != nil {
		return fmt.Errorf("create temp build dir: %w", err)
	}
	defer os.RemoveAll(tempDir)

	tempMainPath := filepath.Join(tempDir, "main.go")
	if err := os.WriteFile(tempMainPath, goSource, 0o644); err != nil {
		return fmt.Errorf("write generated Go: %w", err)
	}

	cmd := exec.Command("go", "build", "-o", outputPath, tempMainPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("go build failed: %w\n%s", err, strings.TrimSpace(string(output)))
	}

	return nil
}

func readBuildSource(inputPath string) ([]byte, error) {
	stat, err := os.Stat(inputPath)
	if err != nil {
		return nil, fmt.Errorf("stat %s: %w", inputPath, err)
	}
	if !stat.IsDir() {
		source, err := os.ReadFile(inputPath)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", inputPath, err)
		}
		return source, nil
	}

	flagFiles, err := findBuildSourceFiles(inputPath)
	if err != nil {
		return nil, err
	}
	if len(flagFiles) == 0 {
		return nil, fmt.Errorf("no source files found in %s (expected .flag, .clj, or .cljc)", inputPath)
	}

	var merged strings.Builder
	for i, file := range flagFiles {
		content, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", file, err)
		}
		if i > 0 {
			merged.WriteString("\n\n")
		}
		merged.WriteString(";; file: ")
		merged.WriteString(file)
		merged.WriteByte('\n')
		merged.Write(content)
	}
	return []byte(merged.String()), nil
}

func findBuildSourceFiles(root string) ([]string, error) {
	files := make([]string, 0, 16)
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".flag" || ext == ".clj" || ext == ".cljc" {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk %s: %w", root, err)
	}
	sort.Strings(files)
	return files, nil
}

func usageError(format string, args ...any) error {
	printUsage()
	return fmt.Errorf(format, args...)
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintln(os.Stderr, "  flag-lang compile [-o output.go] <input.flag>")
	fmt.Fprintln(os.Stderr, "  flag-lang build [-o output-bin] <input.flag|source-dir>")
	fmt.Fprintln(os.Stderr, "  flag-lang repl")
}
