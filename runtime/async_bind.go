package runtime

// FLAG-callable adapters for libraries/async.lib (:go-exports).
// Pure concurrency primitives live in go_async.go / channel.go / time_helpers.go;
// this file only boxes them as Function Values for the module system.

var (
	GoBind_async_GoRun          = NewFunction(adaptAsyncGoRun)
	GoBind_async_FutureRun      = NewFunction(adaptAsyncFutureRun)
	GoBind_async_FuturePipeRun  = NewFunction(adaptAsyncFuturePipeRun)
	GoBind_async_Sleep          = NewFunction(adaptAsyncSleep)
	GoBind_async_MakeChannel    = NewFunction(adaptAsyncMakeChannel)
	GoBind_async_ChannelSend    = NewFunction(adaptAsyncChannelSend)
	GoBind_async_ChannelReceive = NewFunction(adaptAsyncChannelReceive)
	GoBind_async_Select         = NewFunction(adaptAsyncSelect)
)

func adaptAsyncGoRun(args ...Value) Value {
	goArgArityExact("async/go-run", args, 1)
	return GoRun(args[0])
}

func adaptAsyncFutureRun(args ...Value) Value {
	goArgArityExact("async/future-run", args, 1)
	return FutureRun(args[0])
}

func adaptAsyncFuturePipeRun(args ...Value) Value {
	goArgArityExact("async/future-piped-run", args, 1)
	return FuturePipeRun(args[0])
}

func adaptAsyncSleep(args ...Value) Value {
	goArgArityExact("async/sleep", args, 1)
	return Sleep(args[0])
}

func adaptAsyncMakeChannel(args ...Value) Value {
	return MakeChannel(args...)
}

func adaptAsyncChannelSend(args ...Value) Value {
	goArgArityExact("async/channel-send", args, 2)
	return ChannelSend(args[0], args[1])
}

func adaptAsyncChannelReceive(args ...Value) Value {
	goArgArityExact("async/channel-receive", args, 1)
	return ChannelReceive(args[0])
}

func adaptAsyncSelect(args ...Value) Value {
	return ChannelSelect(args...)
}
