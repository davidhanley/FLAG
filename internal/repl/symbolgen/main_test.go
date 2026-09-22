package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestExportedRuntimeNamesIncludesCompilerTargets(t *testing.T) {
	root, err := findRepoRoot()
	if err != nil {
		t.Fatal(err)
	}
	names, err := exportedRuntimeNames(filepath.Join(root, "runtime"))
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]exportKind{}
	for _, item := range names {
		got[item.name] = item.kind
	}
	required := map[string]exportKind{
		"MapCat":                 exportFunc,
		"DoAll":                  exportFunc,
		"BitAnd":                 exportFunc,
		"UnsignedBitShiftRight":  exportFunc,
		"Value":                  exportType,
		"GoBind_async_FutureRun": exportVar,
	}
	for name, kind := range required {
		gotKind, ok := got[name]
		if !ok {
			t.Errorf("missing exported runtime name %s", name)
			continue
		}
		if gotKind != kind {
			t.Errorf("%s: kind %v, want %v", name, gotKind, kind)
		}
	}
	if _, ok := got["GoStruct"]; ok {
		t.Error("generic GoStruct should not be registered with Yaegi")
	}
	for _, item := range names {
		if strings.HasSuffix(item.name, "_test") {
			t.Errorf("unexpected test-only export %s", item.name)
		}
	}
}

func TestRenderSymbolsRegistersTypesAsPointers(t *testing.T) {
	src, err := renderSymbols([]export{
		{name: "MapCat", kind: exportFunc},
		{name: "Value", kind: exportType},
		{name: "GoBind_async_FutureRun", kind: exportVar},
	})
	if err != nil {
		t.Fatal(err)
	}
	text := compactSpace(string(src))
	want := []string{
		`"MapCat": reflect.ValueOf(flagrt.MapCat)`,
		`"Value": reflect.ValueOf((*flagrt.Value)(nil))`,
		`"GoBind_async_FutureRun": reflect.ValueOf(flagrt.GoBind_async_FutureRun)`,
	}
	for _, snippet := range want {
		if !strings.Contains(text, compactSpace(snippet)) {
			t.Errorf("generated source missing %s\n%s", snippet, src)
		}
	}
}

func compactSpace(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
