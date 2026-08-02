# FLAG modules and namespaces

Part of **[The FLAG Book](flag-book.md)** (chapter 3).

FLAG modules are **file-based** and **explicit**. This is intentionally different from Clojure’s classpath-wide `ns` / `:require` model.

## Goals

- One file ≈ one module
- Private by default; public API is an allowlist (`:exports`)
- Imports are explicit paths or library names
- **No flat imports by default** — imported names are qualified
- Short aliases via `:as`; optional bare names via `:refer`

## Module header

The **first form** of a module file is a map:

```clojure
{:namespace "chess"
 :exports   [move legal?]
 :imports   ["board.flag"
             ["csv.flaglib" :as "csv"]
             ["util.flag" :refer [trim]]]}
```

| Key | Required | Default | Meaning |
|-----|----------|---------|---------|
| `:namespace` | yes | — | Canonical module name (string). Default import prefix. |
| `:exports` | no | `[]` | Local def names visible to importers (symbols). |
| `:imports` | no | `[]` | Modules this file depends on. |
| `:go-exports` | no | `{}` | Map of local export name → host bind key for native libraries (see below). |

Unknown keys are currently rejected so the schema stays intentional.

### Body

Everything after the header is ordinary FLAG: `def`, `defn`, `defmacro`, `deftest`, top-level expressions, etc.

Names defined in the body are **private** unless listed in `:exports`.

```clojure
{:namespace "board"
 :exports   [empty-board place]}

(defn empty-board [] ...)
(defn place [b sq piece] ...)
(defn internal-index [sq] ...)   ;; not importable
```

## Import specs

Each `:imports` entry is either a string or a vector:

```clojure
;; Bare path / library name
"chess.flag"
"csv.flaglib"

;; Alias the prefix
["chess.flag" :as "c"]

;; Bind selected exports as unqualified names in *this* module
["chess.flag" :refer [move legal?]]

;; Both
["chess.flag" :as "c" :refer [move]]
```

### Path resolution

- Relative paths resolve against the **importing file’s directory** first
- Then against the process working directory
- Then against **`libraries/`** under the project root (directory containing `libraries/` or `go.mod`, walking up from the importer and cwd)
- Extensions: `.flag`, `.lib`, `.flaglib` (extension may be omitted)

Examples:

```clojure
:imports ["board.flag"]     ;; next to this file
:imports ["burp.lib"]       ;; libraries/burp.lib
:imports ["csv.lib"]        ;; libraries/csv.lib
```

### What a bare import binds

If `chess.flag` has `:namespace "chess"` and exports `move`, `legal?`:

```clojure
:imports ["chess.flag"]
```

usable names in the importer:

- `chess/move`
- `chess/legal?`

Not usable: bare `move`, or non-exported names.

### `:as`

```clojure
:imports [["chess.flag" :as "c"]]
```

usable:

- `c/move`
- `c/legal?`

The alias **replaces** the default prefix for that import-spec. The canonical `chess/…` prefix is **not** also registered unless you import again without `:as` (v1: one prefix per import-spec).

### `:refer`

```clojure
:imports [["chess.flag" :refer [move]]
          ["async.lib" :refer [go future sleep]]]
```

usable:

- `move` / `go` / `sleep` — unqualified in the **current** module
- and the qualified names (`chess/move`, `async/go-run`, …) for all exports of that import

`:refer` works for **functions and macros** listed in the provider’s `:exports`.  
Macros (e.g. `async/go`) are only available unqualified if referred (or as `async/go` via the package prefix).

Referring an unknown or private name is a compile error.

### Combining `:as` and `:refer`

```clojure
:imports [["chess.flag" :as "c" :refer [move]]]
```

- `c/move`, `c/legal?` (qualified under alias)
- bare `move` (referred)

## Name resolution order

When resolving a symbol `S` in module `M`:

1. Local binding (params, `let`, etc.)
2. Unqualified name defined or `:refer`’d in `M`
3. Qualified `prefix/name` against `M`’s import table (prefix from bare import’s `:namespace` or from `:as`)
4. Language prelude / builtins (`map`, `+`, `if`, …)
5. Built-in package bindings that remain globally registered during migration (e.g. some `str/…` adapters) — see “Stdlib” below
6. Else: compile error

Only **exported** names are visible across the import boundary.

## Qualification rules (summary)

| Mechanism | Names in importer |
|-----------|-------------------|
| bare import of `chess` | `chess/export` |
| `:as "c"` | `c/export` |
| `:refer [a b]` | unqualified `a`, `b` in current module |
| not in `:exports` | never importable |

There is **no** Clojure-style `:refer :all` in v1.

## Stdlib and non-default packages

Language prelude (core special forms and builtins) needs no import.

Project libraries live under **`libraries/`** as `.lib` modules. Import them explicitly:

```clojure
{:namespace "app"
 :imports   ["csv.lib" "burp.lib"]}

(csv/read-csv path)
(burp/html [:div "hi"])
```

### Native libraries (`:go-exports`)

See **[go-libraries.md](go-libraries.md)** for the full pure-Go + adapter policy.
Concurrency lives in **[async.md](async.md)** (`libraries/async.lib`).

Some libraries wrap Go implementations. They declare `:go-exports` so the compiler binds exported names to static runtime adapters (no ambient global registration):

```clojure
;; libraries/burp.lib
{:namespace "burp"
 :exports   [html html5 escape raw]
 :go-exports {html   "burp/html"
              html5  "burp/html5"
              escape "burp/escape"
              raw    "burp/raw"}}
```

Without `:imports ["burp.lib"]`, `burp/html` is a compile error.

Other host packages (`str/…`, `io/…`, …) may still resolve without an import during migration; prefer making them libraries over time.

## Legacy `(ns …)`

`(ns some.name)` remains accepted for older files: it records a display namespace only and does **not** enable the export/import system.

Prefer the header map for all new modules.

```clojure
;; legacy
(ns hello.core)

;; preferred
{:namespace "hello"}
```

## Build model

```text
flag-lang build path/to/main.flag
```

1. Read the entry file header
2. Load each import (DFS), detect cycles
3. Compile each module once with its import environment
4. Emit one Go program (unique Go identifiers per `namespace` + local name)
5. `go build`

Directory builds should use an entry file (or a designated main module). Concatenating every file in a directory without regard to imports is the old model and will be phased out for modular projects.

## Errors (v1)

- Missing `:namespace` in a module header
- Import path not found
- Circular imports
- Use of non-exported name via import
- `:refer` of a name not in provider `:exports`
- Duplicate `:as` prefix in one module
- Unqualified `:refer` colliding with a local def or another refer
- Export list naming a symbol not defined in the module
- Imports present but compiled without a file path (string-only compile)

## Example

```clojure
;; board.flag
{:namespace "board"
 :exports   [empty-board place]}

(defn empty-board [] {})
(defn place [b sq piece] (assoc b sq piece))
```

```clojure
;; chess.flag
{:namespace "chess"
 :exports   [move legal?]
 :imports   ["board.flag"]}

(defn move [game from to]
  (assoc game :board (board/place (:board game) to :piece)))

(defn legal? [game] true)
```

```clojure
;; main.flag
{:namespace "main"
 :imports   [["chess.flag" :as "ch"]
             ["chess.flag" :refer [legal?]]]}

(defn flag_main []
  (let [g {:board (board/empty-board)}]  ;; error: board not imported here
    ...))
```

Correct `main` if it needs board:

```clojure
{:namespace "main"
 :imports   ["board.flag"
             ["chess.flag" :as "ch"]]}

(println (ch/legal? {:board (board/empty-board)}))
```

## Future extensions (not v1)

- `:refer :all`, `:exclude`, `:rename`
- Versioned libraries (`csv@1`)
- Multiple files sharing one `:namespace`
- Re-export helpers
- `:main true` entry metadata
```
