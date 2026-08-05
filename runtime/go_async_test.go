package runtime

import (
	"testing"
	"time"
)

func TestFutureReadyPredicate(t *testing.T) {
	f := NewFuture(func() Value {
		time.Sleep(30 * time.Millisecond)
		return NewLong(42)
	})

	if got := Call(f, NewKeyword("ready?")); got.tag != TagBool || got.Bool() {
		t.Fatalf("expected future to report not ready, got %#v", got)
	}

	if got := Call(f); got.tag != TagLong || got.Long() != 42 {
		t.Fatalf("expected future result 42, got %#v", got)
	}

	if got := Call(f, NewKeyword("ready?")); got.tag != TagBool || !got.Bool() {
		t.Fatalf("expected future to report ready, got %#v", got)
	}
}

func TestFuturePipeRun(t *testing.T) {
	ch := FuturePipeRun(NewFunction(func(args ...Value) Value {
		if len(args) != 0 {
			t.Fatalf("expected no args, got %d", len(args))
		}
		return NewString("ok")
	}))

	if got := ChannelReceive(ch); got.tag != TagString || got.StringValue() != "ok" {
		t.Fatalf("expected piped future value 'ok', got %#v", got)
	}
}
