# Wrapping Go libraries for FLAG

Part of **[The FLAG Book](flag-book.md)** (chapter 4).

FLAG ships native binaries via Go. Third-party and project libraries should be
**ordinary Go source**, not a dynamic host reflection surface. This document is
the policy for how FLAG talks to Go — deliberately different from Clojure’s
open Java interop.

## Goals

- **Pure Go packages** with no knowledge of FLAG (`Value`, keywords, runtime).
- **Explicit adapters** that unbox FLAG values, call Go, and box results.
- **Source in, binary out** — libraries are Go modules/packages; `go build`
  compiles and caches for the local OS/arch. No custom `.o` distribution.
- **Importable FLAG surface** via `libraries/*.lib` (see [modules.md](modules.md)).

## Three layers

```text
FLAG user code
    │  import "csv.lib" / "kafka.lib"
    ▼
FLAG module  (libraries/*.lib)     — :exports, optional thin FLAG wrappers
    │
    ▼
Adapter  (runtime or hand-written binds)  — only layer that imports flag-lang/runtime
    │  unbox Value → concrete Go types
    │  call pure API
    │  box result / normalize errors
    ▼
Pure Go package  (libraries/<name>/ or external module)
    — concrete types, idiomatic errors, no `any` for FLAG convenience
```

**Rule:** only the adapter may import `flag-lang/runtime`.  
The pure package must be usable from a normal Go program with no FLAG in the tree.

## Pure Go API rules

1. **Concrete types** — `string`, `[]byte`, `io.Reader`, structs, `[][]string`, etc.
2. **Idiomatic errors** — prefer `(T, error)`. Adapters decide panic vs FLAG exceptions.
3. **No `any` / `interface{}` as an interop kitchen sink.**  
   If several *Go* shapes are valid, use **separate functions** (or a small owned sum type).
4. **Options as structs**, not option maps in pure Go:
   ```go
   type Options struct {
       FieldsPerRecord  int
       LazyQuotes       bool
       // ...
   }
   func DefaultOptions() Options
   func ReadFile(path string, opt Options) ([][]string, error)
   ```
5. **Host objects** (Kafka readers, DB pools) stay as Go types; adapters may later
   expose them as opaque FLAG handles. Field-as-map visibility can come later.

## Adapter rules

Adapters:

| Do | Don’t |
|----|--------|
| Fixed arities with clear names | Reflect over arbitrary objects |
| Explicit `Value` tag switches | Business / domain logic |
| Map FLAG option maps → Go `Options` | Accept `any` “because Lisp” |
| Call one pure function | Re-implement the library |

Example shape:

```go
func GoBind_csv_ReadFile(args ...Value) Value {
    // 1 or 2 args: path, optional opts map
    path := goArgString(...)
    opt := csvOptionsFromValue(...) // keywords → Options
    rows, err := csvlib.ReadFile(path, opt)
    if err != nil { panic(...) }
    return goRetStringMatrix(rows)
}
```

### FLAG option maps

User code may pass a map of **kebab-case keywords** that mirror Go option fields:

```clojure
(csv/read-csv-path "f.csv" {:fields-per-record -1
                            :lazy-quotes true})
```

The adapter owns the mapping (`:fields-per-record` → `Options.FieldsPerRecord`).  
Unknown keys should error clearly. Defaults come from `DefaultOptions()` when the
map is omitted or nil.

Do **not** put option sniffing in the pure Go package.

### Polymorphic “one function” at the FLAG edge

If FLAG wants `(csv/read-csv x)` for path *or* reader *or* lines, implement that
**in the adapter** (or in FLAG `cond`), by dispatching on `Value` tags / types and
calling the matching pure function. The pure package still has only typed entry points.

## FLAG `.lib` surface

```clojure
{:namespace "csv"
 :exports   [read-csv read-csv-path read-csv-reader read-csv-lines]
 :go-exports {read-csv        "csv/read-csv"
              read-csv-path   "csv/read-csv-path"
              ...}}
```

- No import → no `csv/…` symbols (not ambient globals).
- `:go-exports` values are host bind keys resolved to static `GoBind_*` adapters.
- Prefer **more exported functions** with clear names over one mega-entry.

## Layout

```text
libraries/
  csv/
    read.go           # pure Go
    options.go
    read_test.go
  csv.lib             # FLAG exports only
  kafka/
    go/               # or flat package — pure Go wrapper around kafka-go
    kafka.lib
```

Adapters for CSV live under `runtime/` today (with other static binds). Hand-written
adapters are preferred when signatures involve options maps or multi-arity
dispatch; simple unary string/int APIs may stay generated.

HTTP follows the same pattern: `http.lib` re-exports a thin wrapper over
`net/http`, while `runtime/http_bind.go` handles the FLAG value conversions.

## Opaque handles and configs (later)

Stateful objects (`*kafka.Reader`, open files) may appear as opaque FLAG values:

- Construct / use / close only via FLAG functions.
- No open field poking from FLAG by default.
- A later feature may project selected fields as a map for debugging/config
  visibility — that is **adapter or FLAG sugar**, not the pure Go API changing
  to return maps.

## Anti-patterns (Clojure-style traps)

| Avoid | Prefer |
|-------|--------|
| `obj.method` on unknown host types | Fixed exported FLAG fns |
| Reflection by default | Static adapters |
| One `any` parameter “for flexibility” | Separate pure functions + FLAG dispatch |
| Host null/exception soup in user code | Adapter-normalized errors |
| Registering every Go symbol globally | Explicit `.lib` imports |

## Escape hatch

`go-fn` / reflection may remain for REPL exploration. Product libraries and
AOT programs should use static adapters and `.lib` imports only.

## Checklist for a new Go-backed library

1. Write pure Go package + Go tests (no `flag-lang/runtime`).
2. Define `Options` / constructors with sensible defaults.
3. Write adapters (unbox / call / box); map FLAG option keywords.
4. Add `libraries/<name>.lib` with `:exports` and `:go-exports`.
5. Document FLAG arities and option keys in `flag-lang.md` or the lib README.
6. Ensure nothing is ambient without import.
```
