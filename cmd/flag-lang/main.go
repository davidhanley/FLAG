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

	outDir, err := compilePackagesDir(inputPath, outputPath)
	if err != nil {
		return err
	}
	if err := removeStaleFatGo(inputPath, outputPath); err != nil {
		return err
	}
	if err := os.RemoveAll(outDir); err != nil {
		return fmt.Errorf("clear generated Go dir: %w", err)
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("create generated Go dir: %w", err)
	}
	pkgMode, err := writePackagesOrInline(outDir, inputPath, nil)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "flag-lang: wrote generated Go to %s\n", outDir)

	if err := goBuildGenerated(outDir, os.DevNull, pkgMode); err != nil {
		return err
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

	if outputPath == "" {
		base := filepath.Base(inputPath)
		outputPath = strings.TrimSuffix(base, filepath.Ext(base))
	}
	absOut, err := filepath.Abs(outputPath)
	if err != nil {
		return fmt.Errorf("resolve output path: %w", err)
	}

	entryPath, ok, err := resolvePackageBuildEntry(inputPath)
	if err != nil {
		return err
	}
	if ok {
		inspectDir, err := generatedInspectDir(inputPath)
		if err != nil {
			return err
		}
		if err := os.RemoveAll(inspectDir); err != nil {
			return fmt.Errorf("clear generated Go dir: %w", err)
		}
		if err := os.MkdirAll(inspectDir, 0o755); err != nil {
			return fmt.Errorf("create generated Go dir: %w", err)
		}
		if err := removeStaleFatGo(inputPath, ""); err != nil {
			return err
		}
		pkgMode, err := writePackagesOrInline(inspectDir, entryPath, nil)
		if err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "flag-lang: wrote generated Go to %s\n", inspectDir)
		if err := goBuildGenerated(inspectDir, absOut, pkgMode); err != nil {
			return err
		}
		return nil
	}

	tempDir, err := os.MkdirTemp(".", ".flag-build-*")
	if err != nil {
		return fmt.Errorf("create temp build dir: %w", err)
	}
	defer os.RemoveAll(tempDir)

	goSource, err := compileBuildInput(inputPath)
	if err != nil {
		return err
	}

	if err := writeLegacyInspectGo(inputPath, goSource); err != nil {
		return err
	}

	goPath := filepath.Join(tempDir, "main.go")
	if err := os.WriteFile(goPath, goSource, 0o644); err != nil {
		return fmt.Errorf("write generated Go: %w", err)
	}

	cmd := exec.Command("go", "build", "-o", absOut, goPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("go build failed: %w\n%s", err, strings.TrimSpace(string(output)))
	}

	return nil
}

// resolvePackageBuildEntry returns the FLAG entry file for one-package-per-module
// emit. ok is false for legacy directory concatenations.
func resolvePackageBuildEntry(inputPath string) (string, bool, error) {
	stat, err := os.Stat(inputPath)
	if err != nil {
		return "", false, fmt.Errorf("stat %s: %w", inputPath, err)
	}
	if !stat.IsDir() {
		return inputPath, true, nil
	}
	for _, name := range []string{"main.flag", "main.clj", "main.cljc"} {
		candidate := filepath.Join(inputPath, name)
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate, true, nil
		}
	}
	sourceFiles, err := findBuildSourceFiles(inputPath)
	if err != nil {
		return "", false, err
	}
	if len(sourceFiles) == 1 {
		return sourceFiles[0], true, nil
	}
	return "", false, nil
}

// generatedInspectDir is the persistent directory for `flag-lang build` output.
// A leading dot keeps `go test ./...` from treating it as a package.
func generatedInspectDir(inputPath string) (string, error) {
	stat, err := os.Stat(inputPath)
	if err != nil {
		return "", fmt.Errorf("stat %s: %w", inputPath, err)
	}
	dir := inputPath
	if !stat.IsDir() {
		dir = filepath.Dir(inputPath)
	}
	return filepath.Join(dir, ".flag-build"), nil
}

// compilePackagesDir is the package-emit directory for `flag-lang compile`.
// A `-o file.go` path still means “inspect dir next to that file” so we never
// write a fat inlined program.
func compilePackagesDir(inputPath, outputPath string) (string, error) {
	if outputPath == "" {
		return generatedInspectDir(inputPath)
	}
	if strings.EqualFold(filepath.Ext(outputPath), ".go") {
		return filepath.Join(filepath.Dir(outputPath), ".flag-build"), nil
	}
	return outputPath, nil
}

// removeStaleFatGo deletes old inlined compile output (`compile -o foo.go` or
// `<base>.go` next to the source) so inspect never shows prologue__inc in a
// leftover sidecar.
func removeStaleFatGo(inputPath, outputPath string) error {
	remove := func(path string) error {
		if path == "" {
			return nil
		}
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove stale %s: %w", path, err)
		}
		return nil
	}
	if strings.EqualFold(filepath.Ext(outputPath), ".go") {
		if err := remove(outputPath); err != nil {
			return err
		}
	}
	ext := filepath.Ext(inputPath)
	sidecar := filepath.Join(filepath.Dir(inputPath), strings.TrimSuffix(filepath.Base(inputPath), ext)+".go")
	return remove(sidecar)
}

func duplicateGoPackageError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "duplicate Go package name")
}

// writePackagesOrInline writes package-mode Go under dir. Files that share a
// :namespace (e.g. libraries/compiler/*.lib) cannot be separate Go packages
// yet; those fall back to one inlined main.go.
func writePackagesOrInline(dir, entryPath string, testPaths []string) (packageMode bool, err error) {
	if len(testPaths) > 0 {
		err = compiler.WriteProgramPackagesWithTests(dir, entryPath, testPaths)
	} else {
		err = compiler.WriteProgramPackages(dir, entryPath)
	}
	if err == nil {
		return true, nil
	}
	if !duplicateGoPackageError(err) {
		return false, err
	}
	var src []byte
	if len(testPaths) > 0 {
		src, err = compiler.CompileProgramWithTests(entryPath, testPaths)
	} else {
		src, err = compiler.CompileProgram(entryPath)
	}
	if err != nil {
		return false, err
	}
	if err := os.WriteFile(filepath.Join(dir, "main.go"), src, 0o644); err != nil {
		return false, fmt.Errorf("write generated Go: %w", err)
	}
	return false, nil
}

func goBuildGenerated(dir, output string, packageMode bool) error {
	var cmd *exec.Cmd
	if packageMode {
		cmd = exec.Command("go", "build", "-o", output)
		cmd.Dir = dir
	} else {
		cmd = exec.Command("go", "build", "-o", output, filepath.Join(dir, "main.go"))
	}
	if combined, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("go build failed: %w\n%s", err, strings.TrimSpace(string(combined)))
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

	entryPath, testPaths, pkgMode, err := packageTestInput(inputPath)
	if err != nil {
		return err
	}

	tempDir, err := os.MkdirTemp(".", ".flag-test-*")
	if err != nil {
		return fmt.Errorf("create temp test dir: %w", err)
	}
	defer os.RemoveAll(tempDir)

	if outputPath == "" {
		outputPath = filepath.Join(tempDir, "flag-test")
	}
	absOut, err := filepath.Abs(outputPath)
	if err != nil {
		return fmt.Errorf("resolve test binary path: %w", err)
	}

	// Harness is built in tempDir. Inspect output is package-mode `.flag-build/`
	// (prologue linked, not copied). Never write a fat sidecar next to source.
	if err := writeInspectPackages(inputPath); err != nil {
		return err
	}

	sourceOffset := 0
	if pkgMode {
		linked, err := writePackagesOrInline(tempDir, entryPath, testPaths)
		if err != nil {
			return err
		}
		if err := goBuildGenerated(tempDir, absOut, linked); err != nil {
			return err
		}
	} else {
		goSource, offset, err := compileTestInput(inputPath)
		if err != nil {
			return err
		}
		sourceOffset = offset
		tempMainPath := filepath.Join(tempDir, "main.go")
		if err := os.WriteFile(tempMainPath, goSource, 0o644); err != nil {
			return fmt.Errorf("write generated Go: %w", err)
		}
		cmd := exec.Command("go", "build", "-o", absOut, tempMainPath)
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("go build failed: %w\n%s", err, strings.TrimSpace(string(output)))
		}
	}

	runCmd := exec.Command(absOut)
	runOutput, err := runCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("flag test failed: %w\n%s", err, strings.TrimSpace(remapTestRuntimeOutput(string(runOutput), sourceOffset)))
	}
	if remapped := remapTestRuntimeOutput(string(runOutput), sourceOffset); len(strings.TrimSpace(remapped)) > 0 {
		fmt.Fprint(os.Stdout, remapped)
	}

	return nil
}

// packageTestInput is the package-mode entry for `flag-lang test`: one program
// file plus its tests. ok is false for legacy multi-file concatenations.
func packageTestInput(inputPath string) (entry string, tests []string, ok bool, err error) {
	stat, err := os.Stat(inputPath)
	if err != nil {
		return "", nil, false, fmt.Errorf("stat %s: %w", inputPath, err)
	}
	if !stat.IsDir() {
		sourcePath := inputPath
		testPath := ""
		if isTestSourceFile(inputPath) {
			sourcePath = siblingSourceFile(inputPath)
			if sourcePath == "" {
				return "", nil, false, fmt.Errorf("could not find source file for %s", inputPath)
			}
			testPath = inputPath
		} else {
			testPath = siblingTestFile(inputPath)
		}
		if testPath == "" {
			return "", nil, false, fmt.Errorf("could not find test file for %s", inputPath)
		}
		if _, err := os.Stat(testPath); err != nil {
			return "", nil, false, fmt.Errorf("stat %s: %w", testPath, err)
		}
		hasHeader, err := compiler.ModuleHasHeader(sourcePath)
		if err != nil {
			return "", nil, false, err
		}
		if !hasHeader {
			return "", nil, false, nil
		}
		return sourcePath, []string{testPath}, true, nil
	}

	sourceFiles, testFiles, err := collectSourceAndTestFiles(inputPath)
	if err != nil {
		return "", nil, false, err
	}
	if len(testFiles) == 0 {
		return "", nil, false, fmt.Errorf("no test files found in %s", inputPath)
	}
	if len(sourceFiles) == 0 {
		return "", nil, false, fmt.Errorf("no source files found in %s", inputPath)
	}
	if len(sourceFiles) == 1 {
		hasHeader, err := compiler.ModuleHasHeader(sourceFiles[0])
		if err != nil {
			return "", nil, false, err
		}
		if !hasHeader {
			return "", nil, false, nil
		}
		return sourceFiles[0], testFiles, true, nil
	}
	return "", nil, false, nil
}

// compileTestInput builds Go for `flag-lang test`. Modular programs with
// :imports use CompileProgramWithTests; legacy projects still merge sources.
func compileTestInput(inputPath string) ([]byte, int, error) {
	stat, err := os.Stat(inputPath)
	if err != nil {
		return nil, 0, fmt.Errorf("stat %s: %w", inputPath, err)
	}

	if !stat.IsDir() {
		sourcePath := inputPath
		testPath := ""
		if isTestSourceFile(inputPath) {
			sourcePath = siblingSourceFile(inputPath)
			if sourcePath == "" {
				return nil, 0, fmt.Errorf("could not find source file for %s", inputPath)
			}
			testPath = inputPath
		} else {
			testPath = siblingTestFile(inputPath)
		}
		hasImports, err := compiler.ModuleHasImports(sourcePath)
		if err != nil {
			return nil, 0, err
		}
		if hasImports {
			var tests []string
			if testPath != "" {
				if _, err := os.Stat(testPath); err == nil {
					tests = []string{testPath}
				}
			}
			goSource, err := compiler.CompileProgramWithTests(sourcePath, tests)
			return goSource, 0, err
		}
		source, sourceOffset, err := readTestSource(inputPath)
		if err != nil {
			return nil, 0, err
		}
		goSource, err := compiler.Compile(string(source))
		if err != nil {
			return nil, sourceOffset, remapCompileError(err, sourceOffset)
		}
		return goSource, sourceOffset, nil
	}

	sourceFiles, testFiles, err := collectSourceAndTestFiles(inputPath)
	if err != nil {
		return nil, 0, err
	}
	if len(sourceFiles) == 0 {
		return nil, 0, fmt.Errorf("no source files found in %s", inputPath)
	}
	// Prefer modular compile when there is a single program file with imports
	// (e.g. examples/FRS/frs.flag importing csv.lib / burp.lib).
	if len(sourceFiles) == 1 {
		hasImports, err := compiler.ModuleHasImports(sourceFiles[0])
		if err != nil {
			return nil, 0, err
		}
		if hasImports {
			goSource, err := compiler.CompileProgramWithTests(sourceFiles[0], testFiles)
			return goSource, 0, err
		}
	}

	source, sourceOffset, err := readTestSource(inputPath)
	if err != nil {
		return nil, 0, err
	}
	goSource, err := compiler.Compile(string(source))
	if err != nil {
		return nil, sourceOffset, remapCompileError(err, sourceOffset)
	}
	return goSource, sourceOffset, nil
}

// compileBuildInput compiles a file via the module graph, or a directory via
// main.flag / a single modular source / legacy merge.
func compileBuildInput(inputPath string) ([]byte, error) {
	stat, err := os.Stat(inputPath)
	if err != nil {
		return nil, fmt.Errorf("stat %s: %w", inputPath, err)
	}
	if !stat.IsDir() {
		return compiler.CompileProgram(inputPath)
	}
	// Directory: prefer main.flag / main.clj as a modular entry when present.
	for _, name := range []string{"main.flag", "main.clj", "main.cljc"} {
		candidate := filepath.Join(inputPath, name)
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return compiler.CompileProgram(candidate)
		}
	}
	// Single modular source with :imports (e.g. examples/FRS/frs.flag).
	sourceFiles, err := findBuildSourceFiles(inputPath)
	if err != nil {
		return nil, err
	}
	if len(sourceFiles) == 1 {
		hasImports, err := compiler.ModuleHasImports(sourceFiles[0])
		if err != nil {
			return nil, err
		}
		if hasImports {
			return compiler.CompileProgram(sourceFiles[0])
		}
	}
	source, err := readBuildSource(inputPath)
	if err != nil {
		return nil, err
	}
	return compiler.Compile(string(source))
}

// writeInspectPackages writes package-mode Go under `.flag-build/` next to the
// program (prologue linked, not copied). Failures only warn so a successful
// test run is not failed by inspect emit.
func writeInspectPackages(inputPath string) error {
	path := inputPath
	if !isDir(inputPath) && isTestSourceFile(inputPath) {
		if sibling := siblingSourceFile(inputPath); sibling != "" {
			path = sibling
		}
	}
	entryPath, ok, err := resolvePackageBuildEntry(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "flag-lang: skipped writing generated Go: %v\n", err)
		return nil
	}
	inspectDir, err := generatedInspectDir(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "flag-lang: skipped writing generated Go: %v\n", err)
		return nil
	}
	if !ok {
		goSource, err := compileBuildInput(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "flag-lang: skipped writing generated Go: %v\n", err)
			return nil
		}
		return writeLegacyInspectGo(path, goSource)
	}
	if err := os.RemoveAll(inspectDir); err != nil {
		return fmt.Errorf("clear generated Go dir: %w", err)
	}
	if err := os.MkdirAll(inspectDir, 0o755); err != nil {
		return fmt.Errorf("create generated Go dir: %w", err)
	}
	if _, err := writePackagesOrInline(inspectDir, entryPath, nil); err != nil {
		fmt.Fprintf(os.Stderr, "flag-lang: skipped writing generated Go: %v\n", err)
		return nil
	}
	fmt.Fprintf(os.Stderr, "flag-lang: wrote generated Go to %s\n", inspectDir)
	return nil
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// readProgramSource returns the merged program source for inputPath, excluding
// test files. For a file it returns the file itself (or its sibling source file
// when a *_test file was named); for a directory it merges all non-test sources.
func readProgramSource(inputPath string) ([]byte, error) {
	stat, err := os.Stat(inputPath)
	if err != nil {
		return nil, fmt.Errorf("stat %s: %w", inputPath, err)
	}
	if !stat.IsDir() {
		path := inputPath
		if isTestSourceFile(inputPath) {
			path = siblingSourceFile(inputPath)
			if path == "" {
				return nil, fmt.Errorf("could not find source file for %s", inputPath)
			}
		}
		source, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", path, err)
		}
		return source, nil
	}

	sourceFiles, _, err := collectSourceAndTestFiles(inputPath)
	if err != nil {
		return nil, err
	}
	if len(sourceFiles) == 0 {
		return nil, fmt.Errorf("no non-test source files found in %s", inputPath)
	}
	return mergeSourceFiles(sourceFiles)
}

// writeLegacyInspectGo writes a merged (non-package) program into `.flag-build/main.go`.
func writeLegacyInspectGo(inputPath string, goSource []byte) error {
	inspectDir, err := generatedInspectDir(inputPath)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(inspectDir, 0o755); err != nil {
		return fmt.Errorf("create generated Go dir: %w", err)
	}
	goPath := filepath.Join(inspectDir, "main.go")
	if err := os.WriteFile(goPath, goSource, 0o644); err != nil {
		return fmt.Errorf("write generated Go: %w", err)
	}
	fmt.Fprintf(os.Stderr, "flag-lang: wrote generated Go to %s\n", inspectDir)
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

// stripLeadingNamespaceForm removes a legacy (ns ...) form or a module header
// map ({:namespace ...}) from the start of a source buffer so test files can be
// merged into a program that already declared its module.
func stripLeadingNamespaceForm(source []byte) ([]byte, error) {
	i := 0
	for i < len(source) && isWhitespaceByte(source[i]) {
		i++
	}
	if i >= len(source) {
		return source, nil
	}

	// Module header map: first non-whitespace form starts with '{'.
	if source[i] == '{' {
		end, err := scanBalanced(source, i, '{', '}')
		if err != nil {
			return nil, err
		}
		return append([]byte{}, source[end:]...), nil
	}

	// Legacy (ns ...)
	if i+4 > len(source) || source[i] != '(' || !strings.HasPrefix(string(source[i+1:i+4]), "ns ") && !(i+3 < len(source) && source[i+1] == 'n' && source[i+2] == 's' && isDelimiterByte(source[i+3])) {
		return source, nil
	}

	end, err := scanBalanced(source, i, '(', ')')
	if err != nil {
		return nil, fmt.Errorf("unterminated namespace form")
	}
	return append([]byte{}, source[end:]...), nil
}

// scanBalanced returns the index just past a balanced open/close region starting at start.
func scanBalanced(source []byte, start int, open, close byte) (int, error) {
	depth := 0
	inString := false
	escaped := false
	for j := start; j < len(source); j++ {
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
		// Nested collections inside a map/list.
		switch ch {
		case open:
			depth++
		case close:
			depth--
			if depth == 0 {
				return j + 1, nil
			}
		case '(':
			if open != '(' {
				// Track nested lists inside maps via a simple paren depth only when
				// scanning maps? Braces inside strings already ignored; lists use
				// different delimiters so '{}' map scan is independent of ().
			}
		}
		// When scanning a map, also ignore balanced nested [] () that don't affect {}.
		// The depth counter only tracks open/close of the target pair, which is correct
		// for maps containing vectors/lists as long as those don't contain unescaped
		// top-level braces (they shouldn't in module headers).
	}
	return 0, fmt.Errorf("unterminated form starting with %q", string(open))
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
		if ext != ".flag" && ext != ".clj" && ext != ".cljc" {
			return nil
		}
		if isTestSourceFile(path) {
			return nil
		}
		files = append(files, path)
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
	fmt.Fprintln(os.Stderr, "  flag-lang compile [-o dir] <input.flag>")
	fmt.Fprintln(os.Stderr, "  flag-lang build [-o output-bin] <input.flag|source-dir>")
	fmt.Fprintln(os.Stderr, "  flag-lang test [-o test-bin] <input.flag|source-dir>")
	fmt.Fprintln(os.Stderr, "  flag-lang repl")
}
