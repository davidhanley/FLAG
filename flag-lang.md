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
- `(throw x)` / `(ex-info msg map)`
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
- `->` / `->>` / `some->` / `some->>` / `cond->`
- `when-let`
- `with-open` — bind resources and `(defer (fn [] (close name)))` each; LIFO close
- `with-channel` — same as `with-open`, for channels (`(with-channel [ch (make-channel)] ...)`)

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

- `+`, `-`, `*`, `/`, `%`
- `=`, `<`, `<=`, `>`, `>=`
- `max` / `min` (at least one argument)
- `rand-int` (`(rand-int n)` → `[0, n)`)
- `double` (coerce to float)

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
- **R** `range` (0-arg infinite from 0; 1-arg infinite from *n*; 2-arg `[start, end)`; large 2-arg may be lazy)
- **R** `repeat` (`(repeat x)` infinite lazy; `(repeat n x)`)
- **R** `some` (first truthy `(pred x)`, else `nil`)
- **R** `doall` (realize lazy seq, return it) / `dorun` (realize, return `nil`)
- **R** `line-seq` (lazy lines from a file)
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
| `str/` | `trim`, `replace`, `escape`, `split`, `join`, `blank?`, `starts-with?`, `ends-with?`, `upper-case`, `capitalize` |
| `io/` | `reader`, `writer`, `readline`, `scan-directory` |
| `vector/` | `vector`, `get`, `set`, `append`, `prepend`, `pop`, `insert`, `remove` (FLAG vectors only) |
| `json/` | `read`, `read-str` |
| `math/` | `abs` |
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

In REPL, standard-library Yaegi symbols are pre-registered for lookup (for example `fmt.Println`).

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

## Sequences and laziness

- `range` with one arg returns a lazy sequence.
- `range` with no args starts at 0 and returns a lazy sequence.
- large two-arg ranges can be lazy.
- `iterate` and `(repeatedly f)` are lazy (built on infinite `range`).
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
