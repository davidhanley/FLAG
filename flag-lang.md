# flag-lang manual (current state)

`flag-lang` is a Clojure-inspired Lisp that compiles to Go.

This document describes what is implemented **right now** in this repo.

Part of **[The FLAG Book](docs/flag-book.md)** (chapter 2).

## Running and compiling

Start REPL:

```bash
go run ./cmd/flag-lang repl
```

The REPL accepts multi-line forms and waits until delimiters balance before evaluating.

Compile `.flag` source to Go (resolves module `:imports` from the file path):

```bash
go run ./cmd/flag-lang compile examples/hello/src/main.flag -o hello.go
go build -o /dev/null hello.go
go run ./hello.go
```

Build a native binary directly from `.flag`:

```bash
go run ./cmd/flag-lang build examples/hello -o hello
./hello
```

Build from a source directory:

- If `main.flag` (or `main.clj` / `main.cljc`) exists, it is the modular entry and imports are followed.
- Otherwise all `.flag` / `.clj` / `.cljc` files are merged in lexical order (legacy).

```bash
go run ./cmd/flag-lang build path/to/src -o app
./app
```

Run tests from a source directory:

```bash
go run ./cmd/flag-lang test path/to/src
```

## Modules and namespaces

Modular projects use a **header map** as the first form of each file. See [docs/modules.md](docs/modules.md) for the full design.

```clojure
{:namespace "chess"
 :exports   [move legal?]
 :imports   ["board.flag"
             ["csv.flaglib" :as "csv"]
             ["util.flag" :refer [trim]]]}
```

Summary:

- Private by default; only `:exports` are importable
- Bare import → `provider-ns/name` (e.g. `chess/move`)
- `:as "c"` → `c/move`
- `:refer [move]` → unqualified `move` in the **current** module
- Legacy `(ns my.namespace)` still works as a display-only namespace (no export/import)

## Core syntax and forms

Implemented top-level forms:

- Module header map (`:namespace`, `:exports`, `:imports`) — preferred
- `(ns my.namespace)` — legacy
- `(def name expr)`
- `(def name "doc" expr)` optional docstring
- `(defn fname "doc" [args] body)` optional docstring
- `(defn fname ([args] body) ([args2] body) ...)` multiple fixed arities
- `(defmacro name "doc" [args] body)` optional docstring
- `(defmacro name ([args] body) ([args2] body) ...)` multiple arities, same `()` style as `defn`
- `(deftest name body...)` runs during build/repl compilation
- `(defrecord Name [fields])` — Go struct + `->Name` / `map->Name` constructors
- expression forms at top level (evaluated in `main`; entry module only when using imports)

Every compiled program gets **`internal/compiler/prologue.flag`** (macros and FLAG functions). Unqualified **runtime builtins** live in `runtime/builtins.go`. Namespaced hosts (`str/…`, `io/…`, `vector/…`, …) are compile-time Go adapters (`goFnBindings`).

Implemented special forms:

- `(if test then [else])`
- `(do expr1 expr2 ... exprN)`
- `(let [bindings...] body...)`
- `(loop [bindings...] body...)` with `(recur args...)` only in **tail position** of the loop body (not nested in `let` / `if`)
- `(for [bindings...] body)` list comprehension (eager array)
- `(doseq [bindings...] body)` sequential side effects (currently lazy `mapcat`; do not rely on it to launch `go` — use `loop`)
- `(or …)` / `(and …)` short-circuit
- `(doto obj form…)` thread `obj` as first argument
- `(update! name expr)` mutate a `^{:volatile true}` let binding
- `(defer f)` — Go `defer`: evaluate `f` now, call it with no args when the enclosing compiled function returns (LIFO). Use in `do` / `let` / `defn` bodies, e.g. `(defer (fn [] (close chan)))`. Yields `nil` if it is the last body form.
- `(fn [args] body)`
- `#(...)` shorthand function literals (`%`, `%1`, `%2`, ...)
- `(throw x)` / `(ex-info msg map)` / `(ex-info msg map cause)`
- `(try expr* catch-clause* finally-clause?)` with `(catch Type name expr*)` and `(finally expr*)`
- `(ex-message e)` / `(ex-data e)` / `(ex-cause e)`
- `(symbol x)` / `(name x)` / `(keyword x)` / `(str …)` / `(println …)` / `(format fmt args…)`
- `_` and names starting with `_` are intentionally unused bindings (`fn`/`defn`/`let`/`loop`/`for`/`doseq`/destructuring). Multiple `_` are allowed. Prefixed names such as `_k` can still be referenced.
- `(comment ...)` form comments, which the parser discards entirely
- `(testing "label" body...)` test grouping
- `(is expr)` / `(is expr "message")` test assertion with optional message
- `(expect-exception body…)` test helper

### Concurrency (`async.lib`)

Not in the language core. Full reference: **[docs/async.md](docs/async.md)**.

```clojure
{:imports [["async.lib" :refer [go future sleep
                                make-channel channel-send channel-receive
                                select]]]}
```

| Name | Kind | Role |
|------|------|------|
| `go` / `future` | macros | Async body; future returns a 0-arg fn `(f)` for the result |
| `sleep` | function | Pause current goroutine (milliseconds) |
| `make-channel` / `channel-send` / `channel-receive` | functions | FLAG-value channels |
| `atom` / `deref` / `reset!` / `swap!` | functions | Atoms (no watches) |
| `select` | function | Non-blocking multi-receive + handlers; returns count |
| `channel-every?` / `channel-some?` | FLAG fns | Close the input on short-circuit |

Example: [`examples/concurrency`](examples/concurrency).

Implemented macros (from `prologue.flag`):

- `when` / `when-not`
- `inc` / `dec` (macros, not first-class functions — do not pass to `swap!`)
- `not` / `not=`
- `cond`
- `case` (constant match/expr pairs; optional final default)
- `condp` (pred + expr; test/result pairs, `test :>> result-fn`, leftover or `:else`)
- `->` / `->>` / `some->` / `some->>` / `cond->` / `cond->>` / `as->`
- `dotimes` (eager `0 .. n-1`; returns `nil`)
- `when-let` / `if-let` / `if-not` / `if-some` / `when-some` (`if-some`/`when-some` bind when not `nil`, so `false` is kept)
- `with-open` — bind resources and `(defer (fn [] (close name)))` each; LIFO close
- `with-channel` — same as `with-open`, for channels (`(with-channel [ch (make-channel)] ...)`)

There is no `while`. Prefer `loop` / `for` / `doseq` / `dotimes` over open-ended imperative loops.

### `as->`

Named-binding thread. Bind `name` to `expr`, then to each successive form.

```clojure
(as-> 0 n
  (inc n)
  (+ n 5)
  (/ n 2))
;; => 3

(as-> 5 n
  (- 10 n)
  (* n 3))
;; => 15
```

`(as-> x name)` with no forms is `x`.

### `condp`

`(condp pred expr & clauses)` tests `(pred test expr)` for each clause.

- `test result` — if the pred call is truthy, return `result`
- `test :>> result-fn` — if truthy, call `(result-fn pred-result)`
- leftover form or `:else result` — default
- no match — throw (`"No matching clause: …"`)

```clojure
(condp = 2
  1 :a
  2 :b
  3 :c)
;; => :b

(condp = 9
  1 :a
  :z)
;; => :z

(condp = 9
  1 :a
  :else :z)
;; => :z

(condp (fn [want x] (if (= want x) x nil)) 4
  1 :>> inc
  4 :>> dec)
;; => 3
```

`pred` is spliced into each test (same as `case` for the expression).

### `dotimes`

`(dotimes [name n] body…)` runs `body` with `name` bound to `0 .. n-1`. Eager (`doseq` over `(range 0 n)`). Returns `nil`. Negative or zero `n` does nothing.

```clojure
(let [^{:volatile true} acc 0]
  (dotimes [i 5]
    (update! acc (+ acc i)))
  acc)
;; => 10
```

### `try` / `catch` / `finally`

Clojure-shaped. Compiles to Go `defer`/`recover`. `(throw x)` panics the FLAG value; runtime panics (strings) become strings in `catch`.

```clojure
(try
  expr*
  (catch ExceptionInfo e expr*)
  (catch Exception e expr*)
  (finally expr*))
```

Catch types (first match wins):

| Type | Matches |
|------|---------|
| `ExceptionInfo` | maps from `(ex-info msg data)` / `(ex-info msg data cause)` |
| `Exception` / `Throwable` / `:default` | any recovered panic |

`(ex-message e)` / `(ex-data e)` / `(ex-cause e)` follow Clojure (`nil` when absent). Thrown strings yield themselves from `ex-message`.

```clojure
(try
  (throw (ex-info "boom" {:a 1}))
  (catch ExceptionInfo e
    (ex-data e))
  (finally
    (println "done")))
;; => {:a 1}
```

`catch` / `finally` are only legal inside `try`. There are no Java exception classes.

## Data literals

- integers: `1`
- floats: `2.0`
- ratios: `5/6`
- strings: `"hello"`
- multiline strings: `"""hello
  world"""`
- booleans: `true`, `false`
- nil: `nil`
- symbols: `'abc`
- keywords: `:kw`
- lists: `'(1 2 3)` or `(list 1 2 3)`
- arrays: `[1 2 3]` or `(array 1 2 3)`
- vectors: `| 1 2 3 |` or `(vector/vector 1 2 3)`
- maps: `{:a 1 :b 2}`
- sets: `#{1 2 3}`
- characters: `\a`, `\newline`, `\space`, `\tab` (and similar reader chars)

## Functions and calling

Function calls are Lisp-style:

```clojure
(f 1 2)
```

`defn` currently lowers to:

- a direct arity function (`name_arity_N`) per arity
- a variadic wrapper (`name_variadic`) that dispatches on argument count
- a function value var (`name`)

`defmacro` uses the same multiple-arity lists. `macro-case` clauses are `([pattern] body)` lists (vectors still work).

Multiple arities use Clojure-style lists after the name:

```clojure
(defn sort
  ([coll] (sort-by identity coll))
  ([comp coll] (sort-by identity comp coll)))
```

Self-recursive same-arity calls are compiled to direct arity calls for speed. Compiler flags
for direct non-self calls are not exposed yet.

## Destructuring (implemented)

Supported in both `let` and function argument vectors (`defn` / `fn`). Bindings named `_` or starting with `_` do not fail Go unused-variable checks.

### Sequential/vector destructuring

- positional: `[a b c]`
- rest: `[a b & rest]`
- alias: `[a b :as all]`
- nesting supported

### Map destructuring

- explicit key bindings: `{:a a :b b}`
- `:keys [a b]`
- `:syms [x y]`
- `:strs ["k"]`
- defaults: `:or {a 1}`
- alias: `:as m`
- nesting supported

Note: `:strs` currently maps via symbol-key lookup (runtime does not yet have a first-class string `Value` key type).

## Builtin functions

Source of truth: `runtime/builtins.go` (Go) and `internal/compiler/prologue.flag` (FLAG). Marks: **R** = runtime builtin, **P** = prologue.

### Numeric and comparison — R

- `+`, `-`, `*`, `/`, `%` (`%` is Clojure `mod`)
- `quot` / `rem` / `mod` (Clojure: truncating quotient; remainder with sign of dividend; modulus with sign of divisor)
- `=`, `==` (numeric equality; `(== 1 1.0)` is true; non-numbers throw), `<`, `<=`, `>`, `>=`
- `compare` (returns `-1`/`0`/`1`; numbers, strings, and `nil`)
- `max` / `min` (at least one argument)
- `rand-int` (`(rand-int n)` → integer `[0, n)`)
- `rand` (`(rand)` → float `[0, 1)`; `(rand n)` → float `[0, n)`)
- `rand-nth` (random element; empty collection throws)
- `shuffle` (random permutation as an array; does not mutate the input)
- `double` (coerce to float)
- `numerator` / `denominator` (ratios and integers; denominator is always positive; floats throw)
- `bit-and` / `bit-or` / `bit-xor` (variadic, ≥2 args) / `bit-not`
- `bit-shift-left` / `bit-shift-right` / `unsigned-bit-shift-right`
- `bit-test` / `bit-set` / `bit-clear` / `bit-flip`

### Sequence operations

- **R** `first` (also `fist`) / `rest` / `next` (`next` is `nil` on empty) / `last` / `reverse` / `cons`
- **P** `peek` = `first`, `pop` = `rest`
- **P** `second` … `tenth`, `val` (second of a pair)
- **R** `take` / `drop` (`(drop n coll)`)
- **R** `nth` (fast random-access; arrays/strings/vectors; optional not-found)
- **R** `slow-nth` (sequential; lists/lazy seqs; optional not-found)
- **R** `map` (lazy when every input is lazy) / `concat` / `filter` / `reduce` (2- or 3-arg) / `apply`
- **R** `pmap` (parallel map; workers = `NumCPU()*2`, capped by item count; eager array, order preserved)
- **R** `sort-by` (`(sort-by keyfn coll)` or `(sort-by keyfn comp coll)`; array)
- **P** `sort` (`(sort coll)` or `(sort comp coll)`; default `<`)
- **P** `range` (Clojure: `(range)` infinite from 0; `(range end)` is `0 .. end-1`; `(range start end)` / `(range start end step)`; empty when the interval is vacant; step `0` repeats `start`)
- **R** `repeat` (`(repeat x)` infinite lazy; `(repeat n x)`)
- **R** `some` (first truthy `(pred x)`, else `nil`)
- **R** `doall` (realize lazy seq, return it) / `dorun` (realize, return `nil`)
- **R** `line-seq` (lazy lines from a file)
- **P** `take-while` / `drop-while` / `take-last` / `drop-last` / `take-nth` (Clojure collection arities; `take-nth` with `n <= 0` repeats the first item)
- **P** `keep` / `mapcat` / `map-indexed` (`(map f (range) coll)`) / `keep-indexed`
- **P** `reduce-kv` (maps: `f acc k v`; arrays/vectors: `f acc idx v`; `nil` → init)
- **P** `distinct` (first occurrence, input order; array)
- **P** `flatten` (nested sequential; maps/sets are leaves)
- **P** `interpose` / `interleave` (arrays)
- **P** `partition` (`n`, optional `step`/`pad`; drop short tail unless padded)
- **P** `partition-all` (`n`, optional `step`; keep short final group)
- **P** `partition-by` (new group when `f` changes)
- **P** `iterate` (lazy: `x`, `(f x)`, …) / `repeatedly` (`(repeatedly f)` infinite; `(repeatedly n f)`)
- **P** `remove` (`filter` of complement) / `not-any?` / `every?`

### Collections

- **R** `list` / `array` / `hash-map` (constructors; evaluate arguments)
- **R** `set` (from a seq) / `vec` (from a seq)
- **R** `conj` (collection + items) / `into` / `contains?`
- **R** `seq` (`nil` if empty) / `seq?` / `empty?` / `not-empty` (coll or `nil`) / `count`
- **P** `not-empty?` (boolean) / `empty` (same-type empty; `nil` for non-collections)
- **R** `get` (optional default) / `assoc` (maps by key; arrays and FLAG vectors by index, append at `count`) / `dissoc`
- **R** `keys` / `vals` / `find` (entry pair or `nil`)
- **P** `update` / `get-in` / `update-in` / `assoc-in` / `dissoc-in`
- **P** `zipmap` / `group-by` / `select-keys` / `merge` / `merge-with` / `max-key`

### Sets and relations — R

- `union` / `intersection` / `difference`
- `subset?` / `superset?` / `disjoint?`
- `rename-keys` (map) / `map-invert`
- `select` (predicate + set) / `project` / `rename` (relation set of maps)

### Predicates — P except `nil?` **R**

- `true?` / `false?` (only booleans `true` / `false`)
- `boolean` (truthy → `true`, `nil`/`false` → `false`)
- `nil?` / `some?`
- `string?` / `symbol?` / `keyword?`
- `map?` (`:map` / `:record`)
- `vector?` (FLAG vectors `| … |` only; arrays are not vectors)
- `set?` / `sequential?` (`:list` / `:array` / `:vector` / `:lazy-list`) / `coll?`
- `number?` (`:int` / `:float` / `:bigint` / `:ratio`) / `int?` / `float?`
- `zero?` / `pos?` / `neg?` (throw on non-numbers)
- `even?` / `odd?` (integers including bigint; throw otherwise)

### Functional combinators — P

- `identity` / `constantly`
- `partial` / `juxt`
- `comp` (rightmost applied to all args, then unary wrapping; `(comp)` is `identity`)
- `complement` / `fnil` (1–3 defaulted leading args)
- `every-pred` / `some-fn`

### Types — R

- `type-of` → `:int`, `:float`, `:bigint`, `:ratio`, `:bool`, `:string`, `:keyword`, `:symbol`, `:nil`, `:list`, `:array`, `:vector`, `:map`, `:set`, `:fn`, `:date`, `:file`, `:lazy-list`, `:channel`, `:atom`, `:record`

### Symbols / strings / printing

- **R/special** `symbol` / `name` / `keyword` / `str` / `println` / `format` (Go `fmt.Sprintf`)
- **R** `subs` (`(subs s start)` or `(subs s start end)`; rune indices, exclusive end; panics out of range)
- **R** `re-pattern` / `re-matches`

### JSON

- `to-json` / `from-json`
- also `json/read` / `json/read-str` (namespaced)

### File I/O

- **R** `open-file` / `close-file` (idempotent) / `close-channel` (idempotent) / `file-to-strings` (lazy)
- **P** `close` — `(type-of x)` then `close-file` or `close-channel`; throw otherwise
- `(.write file content)` method
- `with-open` / `with-channel` call `(close name)` from a `defer` thunk

### Namespaced runtime packages

Canonical names (aliases such as `string/…`, `datetime/…` also bind):

| Namespace | Functions |
|-----------|-----------|
| `str/` | `trim`, `triml`, `trimr`, `trim-newline`, `replace`, `replace-first`, `escape`, `split`, `split-lines`, `join`, `blank?`, `includes?`, `index-of`, `last-index-of`, `starts-with?`, `ends-with?`, `upper-case`, `lower-case`, `capitalize`, `reverse` |
| `io/` | `reader`, `writer`, `readline`, `scan-directory` |
| `vector/` | `vector`, `get`, `set`, `append`, `prepend`, `pop`, `insert`, `remove` (FLAG vectors only) |
| `json/` | `read`, `read-str` |
| `math/` | `abs`, `sqrt`, `pow`, `exp`, `log`, `log10`, `sin`, `cos`, `tan`, `floor`, `ceil`, `round`, `IEEE-remainder` |
| `regex/` | `compile` (`re-pattern` wraps this) |
| `date/` `dateTime/` `t/` | `from-string`, `formatter`, `now`, `unparse`, `after?`, `minus`, `years` |
| `character/` | `toUpperCase` |
| `long/` | `parse` |

`(.endsWith s suffix)` is a compiler method (string suffix).

### Prelude aliases (not extra implementations)

- `peek` / `pop` — see sequences
- `close` — see File I/O

### Go interop (early)

- `go-fn`
- `go-fn-args`

`go-fn` resolves a registered Go function by name and returns a FLAG-callable function value.
`go-fn-args` returns argument/return metadata for a registered Go function.

In REPL, standard-library Yaegi symbols are pre-registered for lookup (for example `fmt.Println`). Runtime `flagrt` symbols are generated from exported `runtime` identifiers (`go generate ./internal/repl`) so the REPL can evaluate the same compiler output as `flag-lang build`.

Example:

```clojure
(def println (go-fn "fmt.Println"))
(println "hello from go interop")

(go-fn-args "fmt.Println")
;; => {:name fmt.Println
;;     :variadic true
;;     :params [any...]
;;     :returns [int error]}
```

Name resolution accepts common forms for registered symbols (for example `fmt.Println`,
`fmt/fmt.Println`, and package-qualified keys used by the registry).

### Burp HTML rendering

Import the library module (searched under `libraries/`):

```clojure
{:namespace "app"
 :imports   ["burp.lib"]}
```

Then:

- `burp/html`
- `burp/html5`
- `burp/escape`
- `burp/raw`

Burp is a Hiccup-style HTML renderer (`libraries/burp` Go code + `libraries/burp.lib` FLAG exports).

```clojure
(burp/html [:div#app.hero {:data-role "main"} [:span "Hello"]])
```

### CSV reading

See [docs/go-libraries.md](docs/go-libraries.md) for the pure-Go + adapter policy.

```clojure
{:namespace "app"
 :imports   ["csv.lib"]}
```

| FLAG | Pure Go |
|------|---------|
| `(csv/read-csv-path path)` / `(… path opts)` | `csv.ReadFile` |
| `(csv/read-csv-reader rdr)` / `(… rdr opts)` | `csv.ReadAll` |
| `(csv/read-csv-lines lines)` / `(… lines opts)` | `csv.ReadLines` |
| `(csv/read-csv x)` / `(… x opts)` | adapter dispatches to path / reader / lines |

Optional **opts** map (keywords → Go `csv.Options`; defaults: `fields-per-record -1`, `lazy-quotes true`):

```clojure
(csv/read-csv-path "f.csv" {:fields-per-record -1
                            :lazy-quotes true
                            :trim-leading-space false})
```

Keys: `:fields-per-record`, `:lazy-quotes`, `:trim-leading-space`, `:comma`, `:comment`.

## Numerics

Runtime numeric tags include:

- `long` (int64)
- `double` (float64)
- `ratio` (`big.Rat`)
- `bigint` (`big.Int`)

Arithmetic promotes as needed across numeric types.

Recent optimization: numeric comparisons have fast paths for common integer cases (`long/long`, `long/bigint`, `bigint/bigint`), significantly reducing overhead in hot recursive numeric code.

### `quot`, `rem`, `mod` (and `%`)

Clojure-style two-argument ops on ints, ratios, and floats. Integer (or ratio) divide-by-zero throws; float zero follows IEEE.

| Form | Meaning | Sign / rounding | Example |
|------|---------|-----------------|---------|
| `(quot num div)` | integer quotient | toward zero | `(quot -10 3)` → `-3` |
| `(rem num div)` | remainder | sign of **dividend** | `(rem -10 3)` → `-1` |
| `(mod num div)` | modulus | sign of **divisor** | `(mod -10 3)` → `2` |
| `(% num div)` | same as `mod` | sign of **divisor** | `(% -5 3)` → `1` |

Identity: `(+ (* (quot n d) d) (rem n d))` equals `n` (when `d` is nonzero).

```clojure
(quot 10 3)       ;; 3
(quot 10 -3)      ;; -3
(quot 5/2 1/2)    ;; 5
(quot 10.0 3)     ;; 3.0

(rem 10 3)        ;; 1
(rem -10 -3)      ;; -1

(mod 10 3)        ;; 1
(mod -10 3)       ;; 2
(mod 10 -3)       ;; -2
(mod -10 -3)      ;; -1
```

### Bitwise ops (clojure.core)

64-bit two’s-complement longs, matching Clojure / Java `long` ops. Integers
and in-range bigints are accepted; floats, ratios, and oversized bigints throw.
Shift counts are taken mod 64 (`n & 63`). `bit-shift-right` is arithmetic
(sign-extending); `unsigned-bit-shift-right` is logical.

```clojure
(bit-and 1 3 7)                      ;; 1
(bit-or 1 2 4)                       ;; 7
(bit-xor 5 3)                        ;; 6
(bit-not 0)                          ;; -1
(bit-shift-left 1 2)                 ;; 4
(bit-shift-right -8 2)               ;; -2
(unsigned-bit-shift-right -8 2)      ;; 4611686018427387902
(bit-test 2 1)                       ;; true
(bit-set 0 1)                        ;; 2
(bit-clear 3 0)                      ;; 2
(bit-flip 0 0)                       ;; 1
```

### `math/…` (clojure.math)

See **[docs/math.md](docs/math.md)**. `abs` keeps the input type; the rest coerce
to double. `round` returns a long and ties toward +∞ (Java `Math.round`).

```clojure
(math/sqrt 4)                 ;; 2.0
(math/pow 2 3)                ;; 8.0
(math/floor -2.3)             ;; -3.0
(math/round 2.5)              ;; 3
(math/IEEE-remainder 5 3)     ;; -1.0
```

### `==` and `compare`

- `==` is numeric-only (ints, ratios, floats). Zero or one argument is `true`. Non-numbers throw.
- `=` is value equality (and also treats `1` and `1.0` as equal).
- `compare` returns `-1`, `0`, or `1`. `nil` is smaller than any other value.

```clojure
(== 1 1.0 1)          ;; true
(== 1 2)              ;; false
(compare 1 2)         ;; -1
(compare "a" "b")     ;; -1
(compare nil 1)       ;; -1
```

## Sequences and laziness

- `(range)` is a lazy sequence `0, 1, 2, …`.
- `(range n)` is `0 .. n-1` (Clojure), not an infinite sequence from `n`.
- `(range start end)` / `(range start end step)` are exclusive of `end`; negative `step` counts down.
- `iterate` and `(repeatedly f)` are lazy.
- map/filter/reduce/take/drop work across list/array/vector/lazy-list values.
- Vectors are distinct from arrays: `| 1 2 3 |` vs `[1 2 3]`. `vector/*` only accepts vectors.
- `map` returns a lazy sequence when every input sequence is lazy.
- `pmap` currently materializes input tuples, computes in parallel, and returns an eager array while preserving order.

## Current limitations

- Language coverage is partial (not full Clojure yet).
- Some semantics intentionally differ while runtime/data model is still evolving.
- Error messages are improving but still lower-level in some paths.
- Module `:imports` work for file entry points; directory builds without `main.flag` still use legacy file merging.
- Host packages such as `str/…`, `io/…`, and `vector/…` may still resolve without an import during migration; `burp` and `csv` require `libraries/*.lib` imports.
