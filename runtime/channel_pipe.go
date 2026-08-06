package runtime

import (
	"bufio"
	"os"
)

const pipeBufferSize = 32

// ChannelPipeMap applies fn to each value on inch, sending results to a new
// buffered output channel.  Non-blocking: returns the output channel immediately.
// The output channel is closed when inch is exhausted or closed.
func ChannelPipeMap(fn, inch Value) Value {
	if inch.tag != TagChannel {
		panic("pipe-map: second argument must be a channel")
	}
	out := make(chan Value, pipeBufferSize)
	go func() {
		defer close(out)
		for {
			v, ok := <-inch.ChannelObject().ch
			if !ok {
				return
			}
			out <- Call(fn, v)
		}
	}()
	return newChannelValue(out)
}

// ChannelPipeFilter sends values from inch to a new output channel only when
// pred returns truthy.  Non-blocking: returns the output channel immediately.
func ChannelPipeFilter(pred, inch Value) Value {
	if inch.tag != TagChannel {
		panic("pipe-filter: second argument must be a channel")
	}
	out := make(chan Value, pipeBufferSize)
	go func() {
		defer close(out)
		for {
			v, ok := <-inch.ChannelObject().ch
			if !ok {
				return
			}
			if IsTruthy(Call(pred, v)) {
				out <- v
			}
		}
	}()
	return newChannelValue(out)
}

// ChannelPipeReduce drains inch, folding with fn(acc, v).
// Blocking: returns the final accumulated value.
func ChannelPipeReduce(fn, init, inch Value) Value {
	if inch.tag != TagChannel {
		panic("pipe-reduce: third argument must be a channel")
	}
	acc := init
	for {
		v, ok := <-inch.ChannelObject().ch
		if !ok {
			return acc
		}
		acc = Call(fn, acc, v)
	}
}

// ChannelPipeEvery drains inch, returning true if pred is truthy for every
// value, false as soon as any value fails.  Remaining values are drained in a
// background goroutine so the sender is not blocked.
// Blocking: returns a bool Value.
func ChannelPipeEvery(pred, inch Value) Value {
	if inch.tag != TagChannel {
		panic("pipe-every?: second argument must be a channel")
	}
	ch := inch.ChannelObject().ch
	for {
		v, ok := <-ch
		if !ok {
			return NewBool(true)
		}
		if !IsTruthy(Call(pred, v)) {
			go func() {
				for range ch {
				}
			}()
			return NewBool(false)
		}
	}
}

// ChannelPipeSome drains inch, returning the first value for which pred is
// truthy, or nil if none.  Remaining values are drained in a background goroutine.
// Blocking: returns the matching Value or nil.
func ChannelPipeSome(pred, inch Value) Value {
	if inch.tag != TagChannel {
		panic("pipe-some?: second argument must be a channel")
	}
	ch := inch.ChannelObject().ch
	for {
		v, ok := <-ch
		if !ok {
			return NilValue()
		}
		if IsTruthy(Call(pred, v)) {
			go func() {
				for range ch {
				}
			}()
			return v
		}
	}
}

// ChannelLinesPipe reads lines from source (a string path or TagFile value)
// and sends each line as a string Value on a new buffered channel.
// The channel is closed when all lines have been read.
// Non-blocking: returns the output channel immediately.
// File open errors are reported synchronously (panic before the goroutine starts).
func ChannelLinesPipe(source Value) Value {
	out := make(chan Value, pipeBufferSize)
	switch source.tag {
	case TagString:
		path := source.StringValue()
		f, err := os.Open(path)
		if err != nil {
			panic("lines-pipe: cannot open file: " + err.Error())
		}
		go func() {
			defer close(out)
			defer f.Close()
			scanner := bufio.NewScanner(f)
			for scanner.Scan() {
				out <- NewString(scanner.Text())
			}
		}()
	case TagFile:
		fo := source.FileObject()
		go func() {
			defer close(out)
			fo.mu.Lock()
			if fo.closed || fo.File == nil {
				fo.mu.Unlock()
				panic("lines-pipe: file is closed")
			}
			scanner := bufio.NewScanner(fo.File)
			fo.mu.Unlock()
			for scanner.Scan() {
				out <- NewString(scanner.Text())
			}
		}()
	default:
		panic("lines-pipe: source must be a string path or file")
	}
	return newChannelValue(out)
}
