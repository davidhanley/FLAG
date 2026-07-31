package packages

import (
	"testing"
	"time"
)

func TestDateTimeFormatter(t *testing.T) {
	got, err := DateTimeFormatter("yyyy-MM-dd")
	if err != nil {
		t.Fatalf("DateTimeFormatter returned error: %v", err)
	}
	if got != "2006-01-02" {
		t.Fatalf("unexpected layout: %q", got)
	}
}

func TestDateTimeFormatterRejectsUnsupportedTokens(t *testing.T) {
	if _, err := DateTimeFormatter("yyyy-MM-dd z"); err == nil {
		t.Fatal("expected unsupported token error")
	}
}

func TestDateTimeUnparse(t *testing.T) {
	layout, err := DateTimeFormatter("yyyy-MM-dd")
	if err != nil {
		t.Fatalf("DateTimeFormatter returned error: %v", err)
	}
	got, err := DateTimeUnparse(layout, time.Date(2026, time.March, 8, 9, 30, 0, 0, time.FixedZone("x", -4*60*60)))
	if err != nil {
		t.Fatalf("DateTimeUnparse returned error: %v", err)
	}
	if got != "2026-03-08" {
		t.Fatalf("unexpected formatted date: %q", got)
	}
}
