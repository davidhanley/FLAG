package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

type benchRow struct {
	name     string
	ns       int64
	checksum string
	why      string
}

var benchWhy = map[string]string{
	"eager-range-sum": "int64 range cursor; each yielded long is still a boxed Value",
	"lazy-range-sum":  "infinite :range cursor; Clojure still chunks 32",
	"lazy-map-inc":    "lazy map over (range) — FRS fractions/scores path",
	"ratio-scores":    "FRS (map #(/ 5 (+ 5 %)) (range)) then * base; boxed ratios",
	"keyword-gets":    "keyword lookup on persistent maps (every FRS athlete field)",
	"nested-keyword":  "(:race-name (:header m)) nested gets",
	"assoc-scores":    "assoc of :header/:points-scored/:overall-rank (add-scores-and-rank)",
	"group-by-name":   "group-by :name (compute-sex-results)",
	"mapcat-filter":   "mapcat + filter by :sex",
	"sort-by-points":  "sort-by on maps",
	"juxt-keys":       "(juxt :name :age) distinct-athletes keys",
	"string-pipeline": "trim + upper-case (CSV cleaning)",
	"apply-max":       "(apply max (range n)): boxed Values vs Clojure primitive LongRange",
	"frs-score-shape": "lazy scores + assoc + group-by — FRS inner loop without I/O",
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "lispbench: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	root, err := repoRoot()
	if err != nil {
		return err
	}
	example := filepath.Join(root, "examples", "lispbench")
	flagBin := filepath.Join(os.TempDir(), "flag-lispbench")

	fmt.Fprintln(os.Stderr, "lispbench: building FLAG binary")
	build := exec.Command("go", "run", "./cmd/flag-lang", "build", filepath.Join(example, "main.flag"), "-o", flagBin)
	build.Dir = root
	build.Stderr = os.Stderr
	if out, err := build.Output(); err != nil {
		return fmt.Errorf("flag build: %w\n%s", err, out)
	}
	defer os.Remove(flagBin)

	fmt.Fprintln(os.Stderr, "lispbench: running FLAG")
	flagRows, err := runBenches(exec.Command(flagBin))
	if err != nil {
		return fmt.Errorf("flag run: %w", err)
	}

	clj, err := clojureCmd(example)
	if err != nil {
		return err
	}
	fmt.Fprintln(os.Stderr, "lispbench: running Clojure")
	cljRows, err := runBenches(clj)
	if err != nil {
		return fmt.Errorf("clojure run: %w", err)
	}

	printTable(flagRows, cljRows)
	return nil
}

func clojureCmd(example string) (*exec.Cmd, error) {
	script := filepath.Join(example, "lispbench.clj")
	if _, err := exec.LookPath("clojure"); err == nil {
		return exec.Command("clojure", "-M", "-i", script), nil
	}
	if _, err := exec.LookPath("clj"); err == nil {
		return exec.Command("clj", "-M", "-i", script), nil
	}
	return nil, fmt.Errorf("clojure/clj not on PATH")
}

func runBenches(cmd *exec.Cmd) (map[string]benchRow, error) {
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%w\n%s", err, strings.TrimSpace(string(out)))
	}
	rows := map[string]benchRow{}
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "BENCH ") {
			continue
		}
		fields := strings.SplitN(line, " ", 4)
		if len(fields) != 4 {
			return nil, fmt.Errorf("bad BENCH line: %q", line)
		}
		ns, err := strconv.ParseInt(fields[2], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("bad ns in %q: %w", line, err)
		}
		name := fields[1]
		rows[name] = benchRow{name: name, ns: ns, checksum: fields[3], why: benchWhy[name]}
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("no BENCH lines in output:\n%s", out)
	}
	return rows, nil
}

func printTable(flagRows, cljRows map[string]benchRow) {
	names := make([]string, 0, len(flagRows))
	for name := range flagRows {
		names = append(names, name)
	}
	sort.Slice(names, func(i, j int) bool {
		return orderIndex(names[i]) < orderIndex(names[j])
	})

	fmt.Printf("%-18s %10s %10s %9s  %s\n", "bench", "flag_ms", "clj_ms", "flag/clj", "why (FRS-related)")
	fmt.Println(strings.Repeat("-", 110))
	var flagSum, cljSum float64
	for _, name := range names {
		fr, ok := flagRows[name]
		if !ok {
			fmt.Printf("%-18s missing FLAG row\n", name)
			continue
		}
		cr, ok := cljRows[name]
		if !ok {
			fmt.Printf("%-18s missing Clojure row\n", name)
			continue
		}
		if fr.checksum != cr.checksum {
			fmt.Printf("%-18s checksum mismatch flag=%s clj=%s\n", name, fr.checksum, cr.checksum)
		}
		flagMs := float64(fr.ns) / 1e6
		cljMs := float64(cr.ns) / 1e6
		flagSum += flagMs
		cljSum += cljMs
		ratio := 0.0
		if cljMs > 0 {
			ratio = flagMs / cljMs
		}
		fmt.Printf("%-18s %10.1f %10.1f %8.1fx  %s\n", name, flagMs, cljMs, ratio, fr.why)
	}
	ratio := 0.0
	if cljSum > 0 {
		ratio = flagSum / cljSum
	}
	fmt.Println(strings.Repeat("-", 110))
	fmt.Printf("%-18s %10.1f %10.1f %8.1fx  total (same work, in-process; JVM startup excluded)\n",
		"TOTAL", flagSum, cljSum, ratio)
}

func orderIndex(name string) int {
	order := []string{
		"eager-range-sum", "lazy-range-sum", "lazy-map-inc", "ratio-scores",
		"keyword-gets", "nested-keyword", "assoc-scores", "group-by-name",
		"mapcat-filter", "sort-by-points", "juxt-keys", "string-pipeline",
		"apply-max", "frs-score-shape",
	}
	for i, n := range order {
		if n == name {
			return i
		}
	}
	return 1000
}

func repoRoot() (string, error) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		wd, err := os.Getwd()
		return wd, err
	}
	dir := filepath.Dir(file)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found from %s", file)
		}
		dir = parent
	}
}
