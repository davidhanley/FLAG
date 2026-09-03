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
- `(defmacro name "doc" [args] body)` optional docstring
- `(deftest name body...)` runs during build/repl compilation
- expression forms at top level (evaluated in `main`; entry module only when using imports)


Implemented special forms:

- `(if test then [else])`
- `(do expr1 expr2 ... exprN)`
- `(let [bindings...] body...)`
- `(fn [args] body)`
- `#(...)` shorthand function literals (`%`, `%1`, `%2`, ...)
- `(comment ...)` form comments, which the parser discards entirely
- `(testing "label" body...)` test grouping
- `(is expr)` / `(is expr "message")` test assertion with optional message

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
| `select` | function | Non-blocking multi-receive + handlers; returns count |

Example: [`examples/concurrency`](examples/concurrency).

Implemented macros (from standard macros file):

- `when`
- `not=`
- `cond`
- `->`
- `->>`
- `some->`

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
- lists: `'(1 2 3)`
- vectors: `[1 2 3]`
- maps: `{:a 1 :b 2}`
- sets: `#{1 2 3}`

## Functions and calling

Function calls are Lisp-style:

```clojure
(f 1 2)
```

`defn` currently lowers to:

- a direct arity function (`name_arity_N`)
- a variadic wrapper (`name_variadic`)
- a function value var (`name`)

Self-recursive same-arity calls are compiled to direct arity calls for speed. Compiler flags
for direct non-self calls are not exposed yet.

## Destructuring (implemented)

Supported in both `let` and function argument vectors (`defn` / `fn`):

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

### Numeric and comparison

- `+`, `-`, `*`, `/`, `%`
- `=`, `<`, `>`

### Sequence operations

- `first` (also `fist` alias)
- `rest`
- `take`
- `drop` (usage: `(drop n coll)`)
- `map`
- `pmap` (parallel map; worker count = `NumCPU()*2`, capped by item count)
- `filter`
- `reduce`
- `range`

### Collections

- `get`
- `assoc`
- `dissoc`
- `keys`
- `vals`

### Symbols/strings/printing

- `symbol`
- `name`
- `str`
- `println`
- `print`

### JSON

- `to-json`
- `from-json`

### File I/O

- `open-file`
- `file-to-strings`

`file-to-strings` is lazy: it opens/reads on demand as elements are consumed.

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
- map/filter/reduce/take/drop work across list/array/lazy-list values.
- `map` returns a lazy sequence when every input sequence is lazy.
- `pmap` currently materializes input tuples, computes in parallel, and returns an eager array while preserving order.

## Current limitations

- Language coverage is partial (not full Clojure yet).
- Some semantics intentionally differ while runtime/data model is still evolving.
- Error messages are improving but still lower-level in some paths.
- Module `:imports` work for file entry points; directory builds without `main.flag` still use legacy file merging.
- Host packages such as `str/…` and `io/…` may still resolve without an import during migration; `burp` and `csv` require `libraries/*.lib` imports.
