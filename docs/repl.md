# REPL (`flag-lang repl`)

The REPL is the fastest way to try FLAG forms without building a binary.

Start it with:

```bash
go run ./cmd/flag-lang repl
```

At the prompt, ordinary FLAG forms are compiled and evaluated in a live session,
so `def`, `defn`, and `defmacro` stay available for later input.

## Commands

### `:import`

Import a library or module into the current REPL session using the same import
spec syntax as module headers:

```clojure
:import "async.lib"
:import ["async.lib" :refer [future sleep]]
:import ["chess.flag" :as "ch" :refer [move]]
```

- String imports register qualified names like `async/future`
- `:refer` brings selected names into the local REPL scope
- `:as` changes the qualified prefix

### `:load`

Load and evaluate FLAG code from a file:

```clojure
:load "examples/concurrency/main.flag"
```

Loaded files run in the current REPL session, so their defs and functions remain
available after the command finishes.

If the file has a module header with `:imports`, those imports are resolved and
loaded first.

### `:help`, `:quit`, `:exit`

- `:help` prints the REPL commands
- `:quit` and `:exit` leave the REPL

## Example

```clojure
:import ["async.lib" :refer [future sleep]]
(def f (future (do (sleep 10) 42)))
(f)
```

You can also load code from disk and call what it defines:

```clojure
:load "path/to/file.flag"
(my-function)
```

## Compiled code vs Yaegi

The REPL compiles FLAG with the same lowering as `flag-lang build`, then evaluates
the generated Go in [Yaegi](https://github.com/traefik/yaegi). Yaegi only sees
runtime identifiers that are registered as `flagrt` symbols.

Those symbols are generated from every exported `runtime` func, var, and type
(`internal/repl/runtime_symbols_gen.go`), except generic functions that cannot
be registered with `reflect.ValueOf`. After adding a runtime API that the
compiler may emit as `flagrt.Name`, regenerate:

```bash
go generate ./internal/repl
```

That keeps forms such as `doseq` / `for` (`flagrt.MapCat`) and bit ops
(`flagrt.BitAnd`, …) working in the REPL without a hand-maintained allowlist.
