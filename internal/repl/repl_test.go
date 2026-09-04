package repl

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunCollectionPrinting(t *testing.T) {
	cases := []struct {
		expr string
		want string
	}{
		{expr: "(list 1 2 3)", want: "(1 2 3)"},
		{expr: "'(1 2 3)", want: "(1 2 3)"},
		{expr: "(list)", want: "()"},
		{expr: "(array 1 2 3)", want: "[1 2 3]"},
		{expr: "[1 2 3]", want: "[1 2 3]"},
		{expr: "(array)", want: "[]"},
		{expr: "(vector/vector 1 2 3)", want: "| 1 2 3 |"},
		{expr: "| 1 2 3 |", want: "| 1 2 3 |"},
		{expr: "| |", want: "| |"},
		{expr: "{:a 1}", want: "{:a 1}"},
		{expr: "#{1}", want: "#{1}"},
		{expr: "true", want: "true"},
		{expr: `"hi"`, want: "hi"},
	}

	exprs := make([]string, 0, len(cases))
	wants := make([]string, 0, len(cases))
	for _, tc := range cases {
		exprs = append(exprs, tc.expr)
		wants = append(wants, tc.want)
	}

	got := replEval(t, exprs...)
	if len(got) != len(wants) {
		t.Fatalf("expected %d printed results, got %d:\n%q", len(wants), len(got), got)
	}
	for i, want := range wants {
		if got[i] != want {
			t.Fatalf("expr %s: want %q, got %q", exprs[i], want, got[i])
		}
	}
}

func TestRunVectorPrintsPipeSyntax(t *testing.T) {
	got := replEval(t,
		"(def v (vector/vector 1 2 3))",
		"v",
		"| 1 2 3 |",
		"(nth v 1)",
		"(= [1 2 3] v)",
	)
	want := []string{"| 1 2 3 |", "| 1 2 3 |", "| 1 2 3 |", "2", "false"}
	if len(got) != len(want) {
		t.Fatalf("expected %d printed results, got %d:\n%q", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("result %d: want %q, got %q", i, want[i], got[i])
		}
	}
}

func replEval(t *testing.T, forms ...string) []string {
	t.Helper()
	input := strings.NewReader(strings.Join(append(append([]string{}, forms...), ":quit"), "\n") + "\n")
	var output bytes.Buffer
	if err := Run(input, &output); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	got := output.String()
	if strings.Contains(got, "error:") {
		t.Fatalf("expected no REPL errors, got:\n%s", got)
	}
	parts := strings.Split(got, "flag> ")
	if len(parts) > 0 && parts[0] == "" {
		parts = parts[1:]
	}
	if len(parts) > 0 && strings.TrimSpace(parts[len(parts)-1]) == "" {
		parts = parts[:len(parts)-1]
	}
	results := make([]string, 0, len(parts))
	for _, part := range parts {
		results = append(results, strings.TrimSuffix(part, "\n"))
	}
	return results
}

func TestRunDefAssocMap(t *testing.T) {
	input := strings.NewReader("(def a {:a 1 :b 2})\n(def b (assoc a :c 3))\nb\n:quit\n")
	var output bytes.Buffer

	if err := Run(input, &output); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	got := output.String()
	if strings.Contains(got, "error:") {
		t.Fatalf("expected no REPL errors, got:\n%s", got)
	}
	if !strings.Contains(got, "{:a 1 :b 2 :c 3}") {
		t.Fatalf("expected assoc result in output, got:\n%s", got)
	}
}

func TestRunJSONFunctions(t *testing.T) {
	input := strings.NewReader("(to-json {:a 1 :b [2 3]})\n(from-json \"{\\\"a\\\":1}\")\n:quit\n")
	var output bytes.Buffer

	if err := Run(input, &output); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	got := output.String()
	if strings.Contains(got, "error:") {
		t.Fatalf("expected no REPL errors, got:\n%s", got)
	}
	if !strings.Contains(got, "\"a\":1") || !strings.Contains(got, "\"b\":[2,3]") {
		t.Fatalf("expected JSON output in REPL response, got:\n%s", got)
	}
	if !strings.Contains(got, "{:a 1}") {
		t.Fatalf("expected parsed map output in REPL response, got:\n%s", got)
	}
}

func TestRunOpenFileAndFileToStrings(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.flag")
	if err := os.WriteFile(path, []byte("line-a\nline-b\n"), 0o600); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	input := strings.NewReader(fmt.Sprintf("(def f (open-file %q))\n(first (file-to-strings f))\n:quit\n", path))
	var output bytes.Buffer

	if err := Run(input, &output); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	got := output.String()
	if strings.Contains(got, "error:") {
		t.Fatalf("expected no REPL errors, got:\n%s", got)
	}
	if !strings.Contains(got, "line-a") {
		t.Fatalf("expected first file line in output, got:\n%s", got)
	}
}

func TestRunGoFunctionLookup(t *testing.T) {
	input := strings.NewReader(`(def p (go-fn "fmt.Sprintf"))
(p "dave is %d years old" 23)
(go-fn-args "fmt.Sprintf")
:quit
`)
	var output bytes.Buffer

	if err := Run(input, &output); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	got := output.String()
	if strings.Contains(got, "error:") {
		t.Fatalf("expected no REPL errors, got:\n%s", got)
	}
	if !strings.Contains(got, "dave is 23 years old") {
		t.Fatalf("expected formatted go function output, got:\n%s", got)
	}
	if !strings.Contains(got, ":params") || !strings.Contains(got, ":returns") {
		t.Fatalf("expected go function metadata in output, got:\n%s", got)
	}
}

func TestRunMultiLineForm(t *testing.T) {
	input := strings.NewReader("(defn add1 [x]\n  (+ x 1))\n(add1 2)\n:quit\n")
	var output bytes.Buffer

	if err := Run(input, &output); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	got := output.String()
	if strings.Contains(got, "error:") {
		t.Fatalf("expected no REPL errors, got:\n%s", got)
	}
	if !strings.Contains(got, "3") {
		t.Fatalf("expected multiline form result in output, got:\n%s", got)
	}
}

func TestRunDefrecord(t *testing.T) {
	input := strings.NewReader("(defrecord Food [weight calories])\n(:weight (->Food 10 200))\n:quit\n")
	var output bytes.Buffer

	if err := Run(input, &output); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	got := output.String()
	if strings.Contains(got, "error:") {
		t.Fatalf("expected no REPL errors, got:\n%s", got)
	}
	if !strings.Contains(got, "10") {
		t.Fatalf("expected record field lookup in output, got:\n%s", got)
	}
}

func TestRunDefnDashRejected(t *testing.T) {
	input := strings.NewReader("(defn- hidden-helper [x] (+ x 1))\n:quit\n")
	var output bytes.Buffer

	if err := Run(input, &output); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	got := output.String()
	if !strings.Contains(got, "defn- is not supported") {
		t.Fatalf("expected defn- rejection error, got:\n%s", got)
	}
}

func TestRunImportLibraryIntoRepl(t *testing.T) {
	input := strings.NewReader(":import [\"async.lib\" :refer [future sleep]]\n(def f (future (do (sleep 10) 42)))\n(f)\n:quit\n")
	var output bytes.Buffer

	if err := Run(input, &output); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	got := output.String()
	if strings.Contains(got, "error:") {
		t.Fatalf("expected no REPL errors, got:\n%s", got)
	}
	if !strings.Contains(got, "42") {
		t.Fatalf("expected imported async future result in output, got:\n%s", got)
	}
}

func TestRunImportQualifiedLibraryNames(t *testing.T) {
	input := strings.NewReader(":import \"async.lib\"\n(def ch (async/future :piped? true 42))\n(async/channel-receive ch)\n:quit\n")
	var output bytes.Buffer

	if err := Run(input, &output); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	got := output.String()
	if strings.Contains(got, "error:") {
		t.Fatalf("expected no REPL errors, got:\n%s", got)
	}
	if !strings.Contains(got, "42") {
		t.Fatalf("expected qualified async future result in output, got:\n%s", got)
	}
}

func TestRunLoadFileWithImports(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "session.flag")
	source := `{:namespace "session"
 :imports [["async.lib" :refer [future sleep]]]}

(defn answer []
  ((future (do (sleep 10) 42))))
`
	if err := os.WriteFile(path, []byte(source), 0o600); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	input := strings.NewReader(fmt.Sprintf(":load %q\n(answer)\n:quit\n", path))
	var output bytes.Buffer

	if err := Run(input, &output); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	got := output.String()
	if strings.Contains(got, "error:") {
		t.Fatalf("expected no REPL errors, got:\n%s", got)
	}
	if !strings.Contains(got, "42") {
		t.Fatalf("expected loaded file function result in output, got:\n%s", got)
	}
}
