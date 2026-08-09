package runtime

import (
	"fmt"
	"os"
)

// ReportGoPanic logs a panic that escaped a (go ...) / go-run body.
func ReportGoPanic(r any) {
	fmt.Fprintf(os.Stderr, "flag-lang: panic in (go): %v\n", r)
}

// GoRun runs a zero-argument FLAG function on a new goroutine and returns nil
// immediately. Used by the async/go macro: (async/go-run (fn [] body...)).
func GoRun(fn Value) Value {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				ReportGoPanic(r)
			}
		}()
		_ = Call(fn)
	}()
	return NilValue()
}

// NewFuture runs compute on a new goroutine and returns a zero-argument FLAG
// function. Calling that function blocks until the body finishes, then returns
// its value (or re-panics). Further calls return the same cached result without
// blocking. Signaling uses a closed done channel.
func NewFuture(compute func() Value) Value {
	if compute == nil {
		panic("NewFuture expects non-nil function")
	}
	done := make(chan struct{})
	var (
		result   Value
		panicVal any
	)
	go func() {
		defer close(done)
		defer func() {
			if r := recover(); r != nil {
				panicVal = r
			}
		}()
		result = compute()
	}()

	return NewFunction(func(args ...Value) Value {
		if len(args) == 1 && isKeyword(args[0], "ready?") {
			select {
			case <-done:
				return NewBool(true)
			default:
				return NewBool(false)
			}
		}
		if len(args) != 0 {
			panic("future result expects 0 arguments or :ready?")
		}
		<-done
		if panicVal != nil {
			panic(panicVal)
		}
		return result
	})
}

// FutureRun is the FLAG-callable form of NewFuture: takes a zero-arg function.
// Used by the async/future macro: (async/future-run (fn [] body...)).
func FutureRun(fn Value) Value {
	return NewFuture(func() Value {
		return Call(fn)
	})
}

// FuturePipeRun is the FLAG-callable form of a piped future: it runs a zero-arg
// function and returns a channel that receives the result once.
func FuturePipeRun(fn Value) Value {
	if fn.tag != TagFunction {
		panic("FuturePipeRun expects a function Value")
	}
	ch := newChannelValue(make(chan Value, 1))
	go func() {
		defer func() {
			if r := recover(); r != nil {
				ReportGoPanic(r)
				_ = ChannelSend(ch, NilValue())
				ChannelClose(ch)
			}
		}()
		_ = ChannelSend(ch, Call(fn))
		ChannelClose(ch)
	}()
	return ch
}

func isKeyword(v Value, name string) bool {
	return v.tag == TagSymbol && v.SymbolObject().IsKeyword && v.SymbolObject().Name == name
}
