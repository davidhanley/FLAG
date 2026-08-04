# Async concurrency (`libraries/async.lib`)

Part of **[The FLAG Book](flag-book.md)** (chapter 5).

Concurrency is **not** part of the FLAG core language. It lives in
`libraries/async.lib` and is pulled in with a normal module import. That keeps
application binaries small unless you use these features.

Implementation notes (for maintainers): pure Go runtime helpers in
`runtime/go_async.go`, `runtime/channel.go`, `runtime/time_helpers.go`; FLAG
adapters in `runtime/async_bind.go`; surface and macros in `libraries/async.lib`.
Policy for Go-backed libs: [go-libraries.md](go-libraries.md). Modules/imports:
[modules.md](modules.md).

## Import

```clojure
{:namespace "app"
 :imports   [["async.lib" :refer [go future sleep
                                  make-channel channel-send channel-receive
                                  select]]]}
```

With a bare import (no `:refer`), names are qualified:

```clojure
{:imports ["async.lib"]}

(async/go (println "hi"))
(async/sleep 100)
```

`:refer` also works for the **macros** `go` and `future`.

Demo and tests: [`examples/concurrency`](../examples/concurrency).

---

## Summary

| Name | Kind | Role |
|------|------|------|
| `go` | macro | Fire-and-forget body on a new goroutine; returns `nil` |
| `future` | macro | Body on a goroutine; returns a callable result fn |
| `go-run` | function | Engine for `go`: runs a 0-arg function asynchronously |
| `future-run` | function | Engine for `future`: runs a 0-arg function; returns a result fn |
| `sleep` | function | Pause the current goroutine for *n* milliseconds |
| `make-channel` | function | Create a channel of FLAG values (optional buffer size) |
| `channel-send` | function | Blocking send |
| `channel-receive` | function | Blocking receive |
| `select` | function | Non-blocking multi-receive; call handlers for ready channels |

Usually you only `:refer` the macros plus the channel/sleep API. `go-run` and
`future-run` are exported so macro expansions can call `async/go-run` /
`async/future-run` without requiring those names in `:refer`.

---

## `go` (macro)

```clojure
(go body...)
;; expands roughly to:
;; (async/go-run (fn [] (do body...)))
```

- Evaluates `body...` (like `do`) on a **new goroutine**
- Returns **`nil` immediately** to the caller (fire-and-forget)
- Panics in the body are **recovered** and printed to stderr:
  `flag-lang: panic in (go): …`
  The process is not killed by default

```clojure
(go (println "fib:" (fib 40)))
(println "started")   ;; typically prints first
(sleep 1000)          ;; keep process alive if needed
```

---

## `future` (macro)

```clojure
(future body...)
;; expands roughly to:
;; (async/future-run (fn [] (do body...)))
```

- Starts `body...` on a **new goroutine**
- Returns a **callable result function** (not a special “future object”)
- **Call** that function to get the result: `(f)` — no `@` / `deref`
- Call it with `:ready?` to check readiness without blocking: `(f :ready?)`
- First call **blocks** until the body finishes; later calls return the **cached** value without waiting
- If the body panics, the panic is **re-raised** when you call the result function

```clojure
(let [f (future (fib 40))]
  (println "started")
  (println (f :ready?))  ;; false while computing
  (println (f))    ;; blocks until ready
  (println (f)))   ;; cached
```

### Implementation note (done-channel)

Futures use a closed `chan struct{}` as a completion signal and store the result
in outer variables. Multiple waiters and repeated calls are safe. This can be
changed later without changing the FLAG API.

---

## `go-run` / `future-run` (functions)

Low-level engines used by the macros:

| Form | Meaning |
|------|---------|
| `(go-run f)` | Run zero-arg function `f` on a new goroutine; return `nil` |
| `(future-run f)` | Run zero-arg `f` asynchronously; return a 0-arg result function |

You can call them directly if you already have a thunk:

```clojure
(go-run (fn [] (println "hi")))
(let [f (future-run (fn [] 42))] (f))
```

---

## `sleep`

```clojure
(sleep ms)
```

- Pauses the **current** goroutine for `ms` milliseconds (`ms` is a non-negative number)
- Returns `nil`
- Does not start a new goroutine

Useful for demos, pacing, and giving fire-and-forget work time to finish before
process exit.

---

## Channels

Channels carry **FLAG values** (`Value`), not raw Go types. They are opaque
handles printed as `#<channel>`.

### `make-channel`

```clojure
(make-channel)     ;; unbuffered
(make-channel n)   ;; buffered with capacity n (non-negative integer)
```

Returns a new channel.

### `channel-send`

```clojure
(channel-send ch val)
```

- Blocking send of `val` on `ch`
- Unbuffered: waits until a receiver is ready
- Buffered: waits only if the buffer is full
- Returns `nil`

### `channel-receive`

```clojure
(channel-receive ch)
```

- Blocking receive from `ch`
- Returns the received value

```clojure
(let [ch (make-channel)]
  (go (channel-send ch 42))
  (println (channel-receive ch)))   ;; 42
```

There is no `close` / closed-channel receive yet.

---

## `select`

Non-blocking multi-channel poll (not Go’s “wait for one of many”).

```clojure
(select ch1 handler1
        ch2 handler2
        ...)
```

- Arguments are **pairs**: channel, then handler (lambda or any callable)
- For each pair, **try-receive** (non-blocking):
  - if a value is ready → call `(handler value)` and count it
  - if not ready → skip (do not wait)
- Returns a **long**: how many channels had a ready value and were processed
- Odd number of arguments is an error

```clojure
(let [a (make-channel 1)
      b (make-channel 1)
      empty (make-channel 1)
      ^{:volatile true} sum 0]
  (channel-send a 10)
  (channel-send b 20)
  (let [n (select a (fn [v] (update! sum (+ sum v)))
                  empty (fn [_] nil)
                  b (fn [v] (update! sum (+ sum v))))]
    ;; n = 2, sum = 30
    ))
```

Handlers may be `#(...)`, `(fn [v] …)`, or any value `Call` accepts as a function.

---

## Design choices (short)

| Choice | Rationale |
|--------|-----------|
| Library, not core | Keep default binaries small; explicit import graph |
| `go` / `future` as macros | Body must not evaluate before scheduling; expands to thunks |
| Future = callable, not `@` | No extra deref syntax; just `(f)` |
| Future done-channel | Simple, multi-waiter, multi-call; swap later if needed |
| `select` non-blocking | Matches “process whatever is ready now”; returns a count |
| No eval / full image | Async does not force a compiler into app binaries |

---

## Related

- Example: [examples/concurrency](../examples/concurrency)
- Modules / `:refer`: [modules.md](modules.md)
- Go interop policy: [go-libraries.md](go-libraries.md)
- Manual overview: [flag-lang.md](../flag-lang.md)
