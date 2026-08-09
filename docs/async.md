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
                                  channel-close select
                                  channel-map channel-filter channel-reduce
                                  channel-every? channel-some? channel-lines]]]}
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
| `future` | macro | Body on a goroutine; returns a callable result fn, or a channel when piped |
| `go-run` | function | Engine for `go`: runs a 0-arg function asynchronously |
| `future-run` | function | Engine for `future`: runs a 0-arg function; returns a result fn |
| `future-piped-run` | function | Engine for piped futures: runs a 0-arg function; returns a channel |
| `sleep` | function | Pause the current goroutine for *n* milliseconds |
| `make-channel` | function | Create a channel of FLAG values (optional buffer size) |
| `channel-send` | function | Blocking send; returns `true` on success, `false` if terminated |
| `channel-receive` | function | Blocking receive; returns `nil` after termination once buffered values are drained |
| `channel-close` | function | Terminate a channel (idempotent) |
| `select` | function | Non-blocking multi-receive; call handlers for ready channels |
| `channel-map` | function | Non-blocking: apply fn to each value; return output channel |
| `channel-filter` | function | Non-blocking: filter by pred; return output channel |
| `channel-reduce` | function | Blocking: fold channel into a single value |
| `channel-every?` | function | Blocking: true if pred holds for all values; terminates input on false |
| `channel-some?` | function | Blocking: first value matching pred, or nil; terminates input on success |
| `channel-lines` | function | Non-blocking: read file lines into a channel |

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
- Returns a **callable result function** (not a special “future object”) by default
- **Call** that function to get the result: `(f)` — no `@` / `deref`
- Call it with `:ready?` to check readiness without blocking: `(f :ready?)`
- If you pass `:piped? true`, it returns a **channel** instead of a function:
  `(future :piped? true ...)` → `(channel-receive ch)`
- First call **blocks** until the body finishes; later calls return the **cached** value without waiting
- If the body panics, the panic is **re-raised** when you call the result function

```clojure
(let [f (future (fib 40))]
  (println "started")
  (println (f :ready?))  ;; false while computing
  (println (f))    ;; blocks until ready
  (println (f)))   ;; cached

(let [ch (future :piped? true (fib 40))]
  (println (channel-receive ch)))
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
- Returns `true` if sent, `false` if the channel has been terminated

### `channel-receive`

```clojure
(channel-receive ch)
```

- Blocking receive from `ch`
- Returns the received value, or **`nil`** once the channel has been terminated and no buffered values remain

```clojure
(let [ch (make-channel)]
  (go (channel-send ch 42))
  (println (channel-receive ch)))   ;; 42
```

### `channel-close`

```clojure
(channel-close ch)
```

- Terminates `ch`; safe to call multiple times (idempotent)
- Returns `nil`

Use `channel-close` to signal end-of-stream or cancellation.
`channel-send` returns `false` if it is blocked when termination arrives.

---

## Channel pipelines

These operations treat a channel as a lazy stream.  All default to a buffer
size of 32 on output channels.  Import from `async.lib`:

```clojure
{:imports [["async.lib" :refer [channel-map channel-filter channel-reduce
                                 channel-every? channel-some? channel-lines
                                 channel-close]]]}
```

| Name | Blocking? | Returns |
|------|-----------|---------|
| `channel-map` | No | output channel |
| `channel-filter` | No | output channel |
| `channel-reduce` | Yes | final accumulated value |
| `channel-every?` | Yes | bool |
| `channel-some?` | Yes | first match or `nil` |
| `channel-lines` | No | output channel |

### `channel-map`

```clojure
(channel-map f ch)
```

- Non-blocking: spawns a goroutine and returns a new buffered output channel
- Sends `(f v)` for each value `v` received from `ch`
- Output channel is terminated when `ch` is exhausted

```clojure
(let [nums (make-channel 4)]
  (channel-send nums 1)
  (channel-send nums 2)
  (channel-close nums)
  (let [doubled (channel-map (fn [n] (* n 2)) nums)]
    (println (channel-receive doubled))  ;; 2
    (println (channel-receive doubled)))) ;; 4
```

### `channel-filter`

```clojure
(channel-filter pred ch)
```

- Non-blocking: spawns a goroutine and returns a new buffered output channel
- Forwards only values from `ch` for which `(pred v)` is truthy
- Output channel is terminated when `ch` is exhausted

```clojure
(let [evens (channel-filter even? source-ch)]
  ...)
```

### `channel-reduce`

```clojure
(channel-reduce f init ch)
```

- **Blocking**: drains `ch` completely, folding with `(f acc v)` starting from `init`
- Returns the final accumulated value

```clojure
(channel-reduce + 0 numbers-ch)   ;; sum of all values
```

### `channel-every?`

```clojure
(channel-every? pred ch)
```

- **Blocking**: returns `true` if `(pred v)` is truthy for every value in `ch`
- Short-circuits on the first falsy result and terminates `ch`
- Returns `false` as soon as any value fails

```clojure
(channel-every? pos? numbers-ch)
```

### `channel-some?`

```clojure
(channel-some? pred ch)
```

- **Blocking**: returns the first value for which `(pred v)` is truthy, or `nil` if none
- Short-circuits after the first match and terminates `ch`

```clojure
(channel-some? even? numbers-ch)   ;; first even value or nil
```

### `channel-lines`

```clojure
(channel-lines path-or-file)
```

- Accepts a **string path** or an open **file** value
- Non-blocking: returns a buffered channel of string values (one per line, no newline)
- The channel is terminated when all lines have been read
- For string paths, the file is opened synchronously (errors panic at the call site)

```clojure
(let [lines (channel-lines "/etc/hosts")]
  (let [first-line (channel-receive lines)]
    (println first-line)))

;; chain with channel-filter:
(let [ch (channel-filter (fn [l] (not (str/starts-with? l "#")))
                         (channel-lines "/etc/hosts"))]
  ...)
```

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
| `channel-close` idempotent | Safe to call from multiple goroutines; closes the termination signal |
| Pipelines use buffer 32 | Avoids head-of-line blocking without unbounded memory |
| `channel-reduce`/`channel-every?`/`channel-some?` blocking | Aggregation always needs the full stream; callers expect a value |
| Short-circuit closes input | Lets upstream senders stop without panic or background draining |
| `channel-lines` opens file synchronously | Error surfaces at call site, not in a distant goroutine |
| No eval / full image | Async does not force a compiler into app binaries |

---

## Related

- Example: [examples/concurrency](../examples/concurrency)
- Modules / `:refer`: [modules.md](modules.md)
- Go interop policy: [go-libraries.md](go-libraries.md)
- Manual overview: [flag-lang.md](../flag-lang.md)
