package runtime

import (
	"io"
	"os"
	"testing"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdout
	os.Stdout = w
	fn()
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	os.Stdout = old
	out, err := io.ReadAll(r)
	if err := r.Close(); err != nil {
		t.Fatal(err)
	}
	if err != nil {
		t.Fatal(err)
	}
	return string(out)
}

func TestPrintEmpty(t *testing.T) {
	got := captureStdout(t, func() {
		if Print() != NilValue {
			t.Fatal("print of no args should return nil")
		}
	})
	if got != "" {
		t.Fatalf("empty print wrote %q", got)
	}
}

func TestPrintSpacesNoNewline(t *testing.T) {
	got := captureStdout(t, func() {
		if Print(NewLong(1), NewLong(2), NewLong(3)) != NilValue {
			t.Fatal("print should return nil")
		}
	})
	if got != "1 2 3" {
		t.Fatalf("got %q", got)
	}
}

func TestPrintUnquotedStringsAndNil(t *testing.T) {
	got := captureStdout(t, func() {
		Print(NewString("hi"), NilValue, NewString("there"))
	})
	if got != "hi nil there" {
		t.Fatalf("got %q", got)
	}
}

func TestPrintBoolKeywordEmptyString(t *testing.T) {
	got := captureStdout(t, func() {
		Print(NewString(""), TrueValue, FalseValue, NewKeyword("k"))
	})
	if got != " true false :k" {
		t.Fatalf("got %q", got)
	}
}

func TestPrintCollections(t *testing.T) {
	got := captureStdout(t, func() {
		Print(NewArray(NewLong(1), NewLong(2)), NewList(NewLong(3)))
	})
	if got != "[1 2] (3)" {
		t.Fatalf("got %q", got)
	}
}

func TestPrintGoStringLiteral(t *testing.T) {
	got := captureStdout(t, func() {
		Print("a", NewLong(1))
	})
	if got != "a 1" {
		t.Fatalf("got %q", got)
	}
}

func TestFnPrintAndBuiltin(t *testing.T) {
	got := captureStdout(t, func() {
		if FnPrint(NewLong(9), NewString("x")) != NilValue {
			t.Fatal("FnPrint should return nil")
		}
	})
	if got != "9 x" {
		t.Fatalf("FnPrint got %q", got)
	}
	got = captureStdout(t, func() {
		Call(BuiltinFunction("print"), NewString("ok"), NewLong(2))
	})
	if got != "ok 2" {
		t.Fatalf("builtin print got %q", got)
	}
	got = captureStdout(t, func() {
		Apply(BuiltinFunction("print"), NewArray(NewLong(1), NewLong(2), NewLong(3)))
	})
	if got != "1 2 3" {
		t.Fatalf("apply print got %q", got)
	}
}
