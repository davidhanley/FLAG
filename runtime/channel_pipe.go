package runtime

import (
	"bufio"
	"os"
)

const pipeBufferSize = 32

// ChannelPipeMap applies fn to each value on inch, sending results to a new
// buffered output channel. Non-blocking: returns the output channel immediately.
func ChannelPipeMap(fn, inch Value) Value {
	if inch.tag != TagChannel {
		panic("channel-map: second argument must be a channel")
	}
	out := newChannelValue(make(chan Value, pipeBufferSize))
	go func() {
		for {
			v := ChannelReceive(inch)
			if v.tag == TagNil {
				ChannelClose(out)
				return
			}
			if !ChannelSend(out, Call(fn, v)).Bool() {
				return
			}
		}
	}()
	return out
}

// ChannelPipeFilter sends values from inch to a new output channel only when
// pred returns truthy. Non-blocking: returns the output channel immediately.
func ChannelPipeFilter(pred, inch Value) Value {
	if inch.tag != TagChannel {
		panic("channel-filter: second argument must be a channel")
	}
	out := newChannelValue(make(chan Value, pipeBufferSize))
	go func() {
		for {
			v := ChannelReceive(inch)
			if v.tag == TagNil {
				ChannelClose(out)
				return
			}
			if IsTruthy(Call(pred, v)) && !ChannelSend(out, v).Bool() {
				return
			}
		}
	}()
	return out
}

// ChannelPipeReduce drains inch, folding with fn(acc, v).
// Blocking: returns the final accumulated value.
func ChannelPipeReduce(fn, init, inch Value) Value {
	if inch.tag != TagChannel {
		panic("channel-reduce: third argument must be a channel")
	}
	acc := init
	for {
		v := ChannelReceive(inch)
		if v.tag == TagNil {
			return acc
		}
		acc = Call(fn, acc, v)
	}
}

// ChannelPipeEvery drains inch, returning true if pred is truthy for every
// value, false as soon as any value fails.
// Blocking: returns a bool Value.
func ChannelPipeEvery(pred, inch Value) Value {
	if inch.tag != TagChannel {
		panic("channel-every?: second argument must be a channel")
	}
	for {
		v := ChannelReceive(inch)
		if v.tag == TagNil {
			return NewBool(true)
		}
		if !IsTruthy(Call(pred, v)) {
			ChannelClose(inch)
			return NewBool(false)
		}
	}
}

// ChannelPipeSome drains inch, returning the first value for which pred is
// truthy, or nil if none.
// Blocking: returns the matching Value or nil.
func ChannelPipeSome(pred, inch Value) Value {
	if inch.tag != TagChannel {
		panic("channel-some?: second argument must be a channel")
	}
	for {
		v := ChannelReceive(inch)
		if v.tag == TagNil {
			return NilValue()
		}
		if IsTruthy(Call(pred, v)) {
			ChannelClose(inch)
			return v
		}
	}
}

// ChannelLinesPipe reads lines from source (a string path or TagFile value)
// and sends each line as a string Value on a new buffered channel.
// The channel is terminated when all lines have been read.
// Non-blocking: returns the output channel immediately.
func ChannelLinesPipe(source Value) Value {
	out := newChannelValue(make(chan Value, pipeBufferSize))
	switch source.tag {
	case TagString:
		path := source.StringValue()
		f, err := os.Open(path)
		if err != nil {
			panic("channel-lines: cannot open file: " + err.Error())
		}
		go func() {
			defer f.Close()
			scanner := bufio.NewScanner(f)
			for scanner.Scan() {
				if !ChannelSend(out, NewString(scanner.Text())).Bool() {
					return
				}
			}
			ChannelClose(out)
		}()
	case TagFile:
		fo := source.FileObject()
		go func() {
			fo.mu.Lock()
			if fo.closed || fo.File == nil {
				fo.mu.Unlock()
				panic("channel-lines: file is closed")
			}
			scanner := bufio.NewScanner(fo.File)
			fo.mu.Unlock()
			for scanner.Scan() {
				if !ChannelSend(out, NewString(scanner.Text())).Bool() {
					return
				}
			}
			ChannelClose(out)
		}()
	default:
		panic("channel-lines: source must be a string path or file")
	}
	return out
}
