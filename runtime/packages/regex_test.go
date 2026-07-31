package packages

import "testing"

func TestRegexCompile(t *testing.T) {
	matcher := CompileRegex("^he.*o$")
	if !matcher("hello") {
		t.Fatal("expected regex to match hello")
	}
	if matcher("bye") {
		t.Fatal("expected regex to reject bye")
	}
}

func TestRegexCompileWithFlags(t *testing.T) {
	caseInsensitive := CompileRegex("^hello$", 0x02)
	if !caseInsensitive("HELLO") {
		t.Fatal("expected case-insensitive regex to match HELLO")
	}

	literal := CompileRegex("a.b", 0x10)
	if !literal("a.b") {
		t.Fatal("expected literal regex to match exact text")
	}
	if literal("acb") {
		t.Fatal("expected literal regex to reject wildcard-style match")
	}
}

func TestRegexMatches(t *testing.T) {
	matcher := CompileRegex("^he.*o$")
	if !MatchesRegex(matcher, "hello") {
		t.Fatal("expected regex matcher to accept hello")
	}
	if MatchesRegex(matcher, "bye") {
		t.Fatal("expected regex matcher to reject bye")
	}
	if !MatchesRegex(CompileRegex("^he.*o$"), "hello") {
		t.Fatal("expected string pattern to match hello")
	}
}
