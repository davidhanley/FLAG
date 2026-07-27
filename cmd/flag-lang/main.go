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
	case "test":
		return runTest(args[1:])
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
		base := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))
		outputPath = filepath.Join(filepath.Dir(inputPath), base+".go")
	}

	if err := os.WriteFile(outputPath, goSource, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", outputPath, err)
	}

	cmd := exec.Command("go", "build", "-o", os.DevNull, outputPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("go build failed: %w\n%s", err, strings.TrimSpace(string(output)))
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

func runTest(args []string) error {
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
				return usageError("test expects exactly one input file or directory")
			}
			inputPath = arg
		}
	}

	if inputPath == "" {
		return usageError("test expects exactly one input file or directory")
	}

	source, sourceOffset, err := readTestSource(inputPath)
	if err != nil {
		return err
	}

	goSource, err := compiler.Compile(string(source))
	if err != nil {
		return remapCompileError(err, sourceOffset)
	}

	tempDir, err := os.MkdirTemp(".", ".flag-test-*")
	if err != nil {
		return fmt.Errorf("create temp test dir: %w", err)
	}
	defer os.RemoveAll(tempDir)

	if outputPath == "" {
		outputPath = filepath.Join(tempDir, "flag-test")
	}

	tempMainPath := filepath.Join(tempDir, "main.go")
	if err := os.WriteFile(tempMainPath, goSource, 0o644); err != nil {
		return fmt.Errorf("write generated Go: %w", err)
	}

	cmd := exec.Command("go", "build", "-o", outputPath, tempMainPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("go build failed: %w\n%s", err, strings.TrimSpace(string(output)))
	}

	runCmd := exec.Command(outputPath)
	runOutput, err := runCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("flag test failed: %w\n%s", err, strings.TrimSpace(remapTestRuntimeOutput(string(runOutput), sourceOffset)))
	}
	if remapped := remapTestRuntimeOutput(string(runOutput), sourceOffset); len(strings.TrimSpace(remapped)) > 0 {
		fmt.Fprint(os.Stdout, remapped)
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

func readTestSource(inputPath string) ([]byte, int, error) {
	stat, err := os.Stat(inputPath)
	if err != nil {
		return nil, 0, fmt.Errorf("stat %s: %w", inputPath, err)
	}
	if !stat.IsDir() {
		files, err := associatedTestFiles(inputPath)
		if err != nil {
			return nil, 0, err
		}
		if len(files) < 2 {
			return nil, 0, fmt.Errorf("no associated source and test files found for %s", inputPath)
		}
		sourceMerged, err := mergeSourceFiles(files[:1])
		if err != nil {
			return nil, 0, err
		}
		testMerged, err := mergeTestSources(files[1:])
		if err != nil {
			return nil, 0, err
		}
		merged := append(append([]byte{}, sourceMerged...), []byte("\n\n")...)
		merged = append(merged, testMerged...)
		return merged, countLines(sourceMerged) + 2, nil
	}

	sourceFiles, testFiles, err := collectSourceAndTestFiles(inputPath)
	if err != nil {
		return nil, 0, err
	}
	if len(testFiles) == 0 {
		return nil, 0, fmt.Errorf("no test files found in %s", inputPath)
	}
	merged, err := mergeSourceFiles(sourceFiles)
	if err != nil {
		return nil, 0, err
	}
	testMerged, err := mergeTestSources(testFiles)
	if err != nil {
		return nil, 0, err
	}
	out := append(append([]byte{}, merged...), []byte("\n\n")...)
	out = append(out, testMerged...)
	return out, countLines(merged) + 2, nil
}

func countLines(source []byte) int {
	if len(source) == 0 {
		return 1
	}
	return strings.Count(string(source), "\n") + 1
}

func remapCompileError(err error, offset int) error {
	if offset <= 0 {
		return err
	}

	msg := err.Error()
	const marker = " at "
	idx := strings.LastIndex(msg, marker)
	if idx < 0 {
		return err
	}
	var line, col int
	if _, scanErr := fmt.Sscanf(msg[idx+len(marker):], "%d:%d", &line, &col); scanErr != nil {
		return err
	}
	if line <= offset {
		return err
	}
	line -= offset
	return fmt.Errorf("%s at %d:%d", msg[:idx], line, col)
}

func remapTestRuntimeOutput(output string, offset int) string {
	if offset <= 0 || output == "" {
		return output
	}

	lines := strings.Split(output, "\n")
	for i, line := range lines {
		idx := strings.LastIndex(line, "at ")
		if idx < 0 {
			continue
		}
		rest := line[idx+3:]
		var n, col, consumed int
		if _, err := fmt.Sscanf(rest, "%d:%d%n", &n, &col, &consumed); err == nil && n > offset {
			lines[i] = fmt.Sprintf("%sat %d:%d%s", line[:idx], n-offset, col, rest[consumed:])
			continue
		}
		consumed = 0
		if _, err := fmt.Sscanf(rest, "%d%n", &n, &consumed); err == nil && n > offset {
			lines[i] = fmt.Sprintf("%sat %d%s", line[:idx], n-offset, rest[consumed:])
		}
	}
	return strings.Join(lines, "\n")
}

func collectSourceAndTestFiles(root string) ([]string, []string, error) {
	sourceFiles := make([]string, 0, 16)
	testFiles := make([]string, 0, 16)
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".flag" && ext != ".clj" && ext != ".cljc" {
			return nil
		}
		if isTestSourceFile(path) {
			testFiles = append(testFiles, path)
			return nil
		}
		sourceFiles = append(sourceFiles, path)
		return nil
	})
	if err != nil {
		return nil, nil, fmt.Errorf("walk %s: %w", root, err)
	}
	sort.Strings(sourceFiles)
	sort.Strings(testFiles)
	return sourceFiles, testFiles, nil
}

func associatedTestFiles(inputPath string) ([]string, error) {
	if isTestSourceFile(inputPath) {
		paired := siblingSourceFile(inputPath)
		files := make([]string, 0, 2)
		if paired == "" {
			return nil, fmt.Errorf("could not find source file for %s", inputPath)
		}
		if _, err := os.Stat(paired); err != nil {
			return nil, fmt.Errorf("stat %s: %w", paired, err)
		}
		files = append(files, paired)
		files = append(files, inputPath)
		return files, nil
	}

	paired := siblingTestFile(inputPath)
	if paired == "" {
		return nil, fmt.Errorf("could not find test file for %s", inputPath)
	}
	if _, err := os.Stat(paired); err != nil {
		return nil, fmt.Errorf("stat %s: %w", paired, err)
	}
	return []string{inputPath, paired}, nil
}

func mergeSourceFiles(files []string) ([]byte, error) {
	return mergeFiles(files, false)
}

func mergeTestSources(files []string) ([]byte, error) {
	return mergeFiles(files, true)
}

func mergeFiles(files []string, stripNamespace bool) ([]byte, error) {
	var merged strings.Builder
	for i, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", file, err)
		}
		if stripNamespace {
			content, err = stripLeadingNamespaceForm(content)
			if err != nil {
				return nil, fmt.Errorf("strip namespace from %s: %w", file, err)
			}
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

func siblingTestFile(path string) string {
	ext := filepath.Ext(path)
	base := strings.TrimSuffix(filepath.Base(path), ext)
	if strings.HasSuffix(base, "_test") {
		return ""
	}
	return filepath.Join(filepath.Dir(path), base+"_test"+ext)
}

func siblingSourceFile(path string) string {
	ext := filepath.Ext(path)
	base := strings.TrimSuffix(filepath.Base(path), ext)
	if !strings.HasSuffix(base, "_test") {
		return ""
	}
	return filepath.Join(filepath.Dir(path), strings.TrimSuffix(base, "_test")+ext)
}

func isTestSourceFile(path string) bool {
	ext := filepath.Ext(path)
	base := strings.TrimSuffix(filepath.Base(path), ext)
	return strings.HasSuffix(base, "_test")
}

func stripLeadingNamespaceForm(source []byte) ([]byte, error) {
	i := 0
	for i < len(source) && isWhitespaceByte(source[i]) {
		i++
	}
	if i+4 > len(source) || source[i] != '(' || !strings.HasPrefix(string(source[i+1:i+4]), "ns ") && !(i+3 < len(source) && source[i+1] == 'n' && source[i+2] == 's' && isDelimiterByte(source[i+3])) {
		return source, nil
	}

	depth := 0
	inString := false
	escaped := false
	for j := i; j < len(source); j++ {
		ch := source[j]
		if inString {
			if escaped {
				escaped = false
				continue
			}
			if ch == '\\' {
				escaped = true
				continue
			}
			if ch == '"' {
				inString = false
			}
			continue
		}
		if ch == '"' {
			inString = true
			continue
		}
		if ch == ';' {
			for j < len(source) && source[j] != '\n' {
				j++
			}
			continue
		}
		switch ch {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return append([]byte{}, source[j+1:]...), nil
			}
		}
	}
	return nil, fmt.Errorf("unterminated namespace form")
}

func isWhitespaceByte(ch byte) bool {
	switch ch {
	case ' ', '\t', '\n', '\r':
		return true
	default:
		return false
	}
}

func isDelimiterByte(ch byte) bool {
	switch ch {
	case '(', ')', '[', ']', '{', '}', '"', '\'', ';', ',':
		return true
	default:
		return isWhitespaceByte(ch)
	}
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
	fmt.Fprintln(os.Stderr, "  flag-lang test [-o test-bin] <input.flag|source-dir>")
	fmt.Fprintln(os.Stderr, "  flag-lang repl")
}
