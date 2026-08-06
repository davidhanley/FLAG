package runtime

import (
	"testing"
)

// makeTestPipeChan returns a closed, pre-filled channel Value ready for pipeline tests.
func makeTestPipeChan(vals ...Value) Value {
	ch := MakeChannel(NewLong(int64(len(vals) + 1)))
	for _, v := range vals {
		ChannelSend(ch, v)
	}
	ChannelClose(ch)
	return ch
}

// collectPipe drains a pipeline output channel into a slice.
func collectPipe(ch Value) []Value {
	var out []Value
	for {
		v := ChannelReceive(ch)
		if v.tag == TagNil {
			return out
		}
		out = append(out, v)
	}
}

func TestChannelClose_Idempotent(t *testing.T) {
	ch := MakeChannel(NewLong(1))
	ChannelClose(ch)
	ChannelClose(ch) // must not panic
}

func TestChannelReceive_ClosedReturnsNil(t *testing.T) {
	ch := MakeChannel(NewLong(1))
	ChannelSend(ch, NewLong(42))
	ChannelClose(ch)
	v1 := ChannelReceive(ch)
	if v1.tag != TagLong || v1.Long() != 42 {
		t.Fatalf("expected 42, got %v", v1)
	}
	v2 := ChannelReceive(ch)
	if v2.tag != TagNil {
		t.Fatalf("expected nil on closed channel, got tag %v", v2.tag)
	}
}

func TestPipeMap(t *testing.T) {
	inch := makeTestPipeChan(NewLong(1), NewLong(2), NewLong(3))
	double := NewFunction(func(args ...Value) Value { return NewLong(args[0].Long() * 2) })
	got := collectPipe(ChannelPipeMap(double, inch))
	if len(got) != 3 || got[0].Long() != 2 || got[1].Long() != 4 || got[2].Long() != 6 {
		t.Fatalf("pipe-map unexpected: %v", got)
	}
}

func TestPipeFilter(t *testing.T) {
	inch := makeTestPipeChan(NewLong(1), NewLong(2), NewLong(3), NewLong(4))
	even := NewFunction(func(args ...Value) Value { return NewBool(args[0].Long()%2 == 0) })
	got := collectPipe(ChannelPipeFilter(even, inch))
	if len(got) != 2 || got[0].Long() != 2 || got[1].Long() != 4 {
		t.Fatalf("pipe-filter unexpected: %v", got)
	}
}

func TestPipeReduce(t *testing.T) {
	inch := makeTestPipeChan(NewLong(1), NewLong(2), NewLong(3), NewLong(4))
	add := NewFunction(func(args ...Value) Value { return NewLong(args[0].Long() + args[1].Long()) })
	result := ChannelPipeReduce(add, NewLong(0), inch)
	if result.Long() != 10 {
		t.Fatalf("pipe-reduce expected 10, got %v", result.Long())
	}
}

func TestPipeEvery_AllTrue(t *testing.T) {
	inch := makeTestPipeChan(NewLong(2), NewLong(4), NewLong(6))
	even := NewFunction(func(args ...Value) Value { return NewBool(args[0].Long()%2 == 0) })
	r := ChannelPipeEvery(even, inch)
	if r.tag != TagBool || !r.Bool() {
		t.Fatal("expected true")
	}
}

func TestPipeEvery_ShortCircuit(t *testing.T) {
	inch := makeTestPipeChan(NewLong(2), NewLong(3), NewLong(4))
	even := NewFunction(func(args ...Value) Value { return NewBool(args[0].Long()%2 == 0) })
	r := ChannelPipeEvery(even, inch)
	if r.tag != TagBool || r.Bool() {
		t.Fatal("expected false")
	}
}

func TestPipeSome_Found(t *testing.T) {
	inch := makeTestPipeChan(NewLong(1), NewLong(3), NewLong(4), NewLong(5))
	even := NewFunction(func(args ...Value) Value { return NewBool(args[0].Long()%2 == 0) })
	r := ChannelPipeSome(even, inch)
	if r.tag != TagLong || r.Long() != 4 {
		t.Fatalf("expected 4, got %v", r)
	}
}

func TestPipeSome_NotFound(t *testing.T) {
	inch := makeTestPipeChan(NewLong(1), NewLong(3), NewLong(5))
	even := NewFunction(func(args ...Value) Value { return NewBool(args[0].Long()%2 == 0) })
	r := ChannelPipeSome(even, inch)
	if r.tag != TagNil {
		t.Fatalf("expected nil, got tag %v", r.tag)
	}
}

func TestLinesPipe_FromPath(t *testing.T) {
	// Use an absolute path to a known project file.
	path := "/Users/davidhanley/projects/FLAG/runtime/channel_pipe.go"
	ch := ChannelLinesPipe(NewString(path))
	got := collectPipe(ch)
	if len(got) == 0 {
		t.Fatal("lines-pipe returned no lines for channel_pipe.go")
	}
	if got[0].tag != TagString || got[0].StringValue() != "package runtime" {
		t.Fatalf("unexpected first line: %q", got[0].StringValue())
	}
}
