package compiler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCompileProgramLibraryImportCompilerTokenizer(t *testing.T) {
	dir := t.TempDir()
	main := filepath.Join(dir, "main.flag")
	if err := os.WriteFile(main, []byte(`
{:namespace "main"
 :imports [["compiler.lib" :as "shadow"]
           ["async.lib" :refer [channel-receive]]]}
(defn drain [ch]
  (let [v (channel-receive ch)]
    (if (nil? v)
      []
      (conj (drain ch) v))))
(println (count (drain (shadow/tokenize-file "sample.flag"))))
`), 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := CompileProgram(main)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			t.Skipf("libraries/ not resolvable from test cwd: %v", err)
		}
		t.Fatalf("CompileProgram: %v", err)
	}
	got := string(out)
	for _, want := range []string{
		"func compiler__tokenize_file_arity_1",
		"var compiler__tokenize_file =",
		"flagrt.Call(compiler__tokenize_file",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in:\n%s", want, got)
		}
	}
}
