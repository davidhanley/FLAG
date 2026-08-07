# concurrency example

Demonstrates `libraries/async.lib` imported with `:refer`.

Full API reference: **[docs/async.md](../../docs/async.md)**.

```clojure
{:imports [["async.lib" :refer [go future sleep
                                make-channel channel-send channel-receive
                                channel-close select
                                channel-some? channel-every?]]]}
```

- `(go body...)` / `(future body...)` — macros; future values also answer `(f :ready?)` and may be piped with `:piped? true`
- `(make-channel)`, `(channel-send)`, `(channel-receive)`, `(channel-close)`, `(sleep ms)` — functions
- `(select ch f ch f…)` — non-blocking multi-receive; calls each ready handler; returns count
- `(channel-some?)` / `(channel-every?)` — short-circuiting stream reducers that terminate their input

## Run the demo

```bash
go run ./cmd/flag-lang build examples/concurrency/main.flag -o /tmp/flag-concurrency
/tmp/flag-concurrency
```

Or as a directory (uses `main.flag`):

```bash
go run ./cmd/flag-lang build examples/concurrency -o /tmp/flag-concurrency
/tmp/flag-concurrency
```

## Tests

FLAG unit tests live in `main_test.flag` (fib, sleep, go). Run them with:

```bash
go run ./cmd/flag-lang test examples/concurrency
```

These are also run from the Go acceptance suite (`cmd/flag-lang` tests).

## Expected demo output (`go:` line order may vary)

```
main: start
main: future started (not waiting yet)
future: computing fib(20)...
main: future result = 6765
main: future again = 6765
main: kicked off work
go: computing fib(28)...
go: fib(28) = 317811
go: later task done
main: end
```

`(sleep 500)` at the end of `main` keeps the process alive for fire-and-forget
`(go …)` work. The future result is obtained with `(f)`, not a special deref,
and `(f :ready?)` checks readiness without blocking. Use `(future :piped? true
...)` when you want a channel instead of a callable result function.
