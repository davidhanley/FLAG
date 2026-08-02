package runtime

import "testing"

func TestChannelSendReceive(t *testing.T) {
	ch := MakeChannel()
	done := make(chan struct{})
	go func() {
		defer close(done)
		ChannelSend(ch, NewLong(42))
	}()
	got := ChannelReceive(ch)
	<-done
	if got.tag != TagLong || got.Long() != 42 {
		t.Fatalf("got %v", ValueToString(got))
	}
}

func TestBufferedChannel(t *testing.T) {
	ch := MakeChannel(NewLong(1))
	ChannelSend(ch, NewString("hi")) // must not block
	got := ChannelReceive(ch)
	if got.tag != TagString || got.StringValue() != "hi" {
		t.Fatalf("got %v", ValueToString(got))
	}
}

func TestChannelSelect(t *testing.T) {
	a := MakeChannel(NewLong(1))
	b := MakeChannel(NewLong(1))
	empty := MakeChannel(NewLong(1))
	ChannelSend(a, NewLong(1))
	ChannelSend(b, NewLong(2))

	var seen []int64
	handler := NewFunction(func(args ...Value) Value {
		seen = append(seen, args[0].Long())
		return NilValue()
	})
	n := ChannelSelect(a, handler, empty, handler, b, handler)
	if n.tag != TagLong || n.Long() != 2 {
		t.Fatalf("expected 2 ready, got %v", ValueToString(n))
	}
	if len(seen) != 2 || seen[0] != 1 || seen[1] != 2 {
		t.Fatalf("handlers saw %v", seen)
	}
	// Nothing left ready.
	n2 := ChannelSelect(a, handler, b, handler, empty, handler)
	if n2.Long() != 0 {
		t.Fatalf("expected 0, got %d", n2.Long())
	}
}
