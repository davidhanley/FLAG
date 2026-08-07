package runtime

import (
	"fmt"
	"sync"
	"unsafe"
)

// ChannelObject is an opaque FLAG channel.
// data carries stream values; done is the termination/cancellation signal.
type ChannelObject struct {
	data chan Value
	done chan struct{}
	once sync.Once
}

// closeDone closes the termination signal exactly once; safe to call many times.
func (c *ChannelObject) closeDone() {
	c.once.Do(func() { close(c.done) })
}

// MakeChannel creates a channel.
//   - no args: unbuffered
//   - one non-negative long: buffered with that capacity
func MakeChannel(args ...Value) Value {
	capacity := 0
	switch len(args) {
	case 0:
		// unbuffered
	case 1:
		if !isNumericTag(args[0].tag) {
			panic("make-channel capacity must be a non-negative integer")
		}
		n := args[0].Long()
		if n < 0 {
			panic("make-channel capacity must be non-negative")
		}
		if n > int64(^uint(0)>>1) {
			panic("make-channel capacity too large")
		}
		capacity = int(n)
	default:
		panic("make-channel expects 0 or 1 arguments")
	}
	return newChannelValue(make(chan Value, capacity))
}

func newChannelValue(ch chan Value) Value {
	return Value{p: unsafe.Pointer(&ChannelObject{data: ch, done: make(chan struct{})}), tag: TagChannel}
}

func (v Value) ChannelObject() *ChannelObject {
	if v.tag != TagChannel {
		panic("ChannelObject called on non-channel Value")
	}
	if v.p == nil {
		panic("channel Value does not contain channel pointer")
	}
	return (*ChannelObject)(v.p)
}

// ChannelSend sends value on ch.
// Returns true if the value was sent, false if the channel was terminated.
func ChannelSend(ch Value, value Value) Value {
	if ch.tag != TagChannel {
		panic(fmt.Sprintf("channel-send expects a channel, got %s", ValueToString(ch)))
	}
	c := ch.ChannelObject()
	select {
	case <-c.done:
		return NewBool(false)
	default:
	}
	select {
	case c.data <- value:
		return NewBool(true)
	case <-c.done:
		return NewBool(false)
	}
}

// ChannelClose terminates ch; safe to call multiple times (idempotent).
// This is the single close signal for FLAG channels.
func ChannelClose(ch Value) Value {
	if ch.tag != TagChannel {
		panic(fmt.Sprintf("channel-close expects a channel, got %s", ValueToString(ch)))
	}
	ch.ChannelObject().closeDone()
	return NilValue()
}

// recvChannelValue receives one value, preferring queued data over termination.
// Once the done signal is closed and no data remains, it returns nil.
func recvChannelValue(c *ChannelObject) (Value, bool) {
	select {
	case v := <-c.data:
		return v, true
	case <-c.done:
		select {
		case v := <-c.data:
			return v, true
		default:
			return NilValue(), false
		}
	}
}

// ChannelReceive receives one value from ch.
// Returns nil when the channel has been terminated and no buffered values remain.
func ChannelReceive(ch Value) Value {
	if ch.tag != TagChannel {
		panic(fmt.Sprintf("channel-receive expects a channel, got %s", ValueToString(ch)))
	}
	v, ok := recvChannelValue(ch.ChannelObject())
	if !ok {
		return NilValue()
	}
	return v
}

// ChannelSelect takes alternating channel and handler pairs:
//
//	(select ch1 f1 ch2 f2 ...)
//
// For each pair, if a value is ready on the channel (non-blocking try-receive),
// the handler is called with that value. Handlers may be functions or other
// callable Values (via Call). Returns the number of channels that had a value
// ready and were processed. Does not wait for empty channels.
func ChannelSelect(args ...Value) Value {
	if len(args) == 0 {
		return NewLong(0)
	}
	if len(args)%2 != 0 {
		panic("select expects alternating channel and handler pairs")
	}
	count := int64(0)
	for i := 0; i < len(args); i += 2 {
		ch := args[i]
		handler := args[i+1]
		if ch.tag != TagChannel {
			panic(fmt.Sprintf("select expects a channel at position %d, got %s", i+1, ValueToString(ch)))
		}
		select {
		case v := <-ch.ChannelObject().data:
			_ = Call(handler, v)
			count++
		default:
			// not ready
		}
	}
	return NewLong(count)
}
