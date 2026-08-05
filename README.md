![flag-lang banner](./flag-lang.jpg)

# flag-lang

`flag-lang` is a Clojure-inspired language that compiles to Go.

The goal is to keep the expressiveness of Lisp/Clojure while delivering a faster, lighter runtime model for production workloads.

**Documentation:** [The FLAG Book](docs/flag-book.md) — table of contents for the language manual, modules, Go libraries, and async.

## Project goals

`flag-lang` is being built to be a **better Clojure** in key operational areas:

- **Fast startup:** no JVM warmup cost for small tools/services.
- **Lower memory usage:** compact runtime representation for common values.
- **Higher performance:** unboxed values and arrays, plus stack allocation where possible (for example, temporaries in map/filter style sequences).
- **Small standalone binaries:** around **10MB** for small programs, no VM needed.
- **Small docker container images:** no VM required, enabling minimal images.
- **Native interop:** easy linkage to highly optimized **C**, **Go**, and **Rust** code.

## Current state

The repository currently includes:

- a compiler CLI (`flag-lang`)
- a parser that reads Clojure-like source into an AST
- Go code generation for an initial subset (`defn`, arithmetic, calls, printing)
- a runtime with tagged values, numerics, ratios, lists, and arrays
- an interactive REPL path using Yaegi evaluation
- libraries under `libraries/` (e.g. `burp.lib`, `csv.lib`, `async.lib`) imported like any module
- concurrency via `async.lib` ([docs/async.md](docs/async.md))
- Go interop policy: pure Go packages + adapters ([docs/go-libraries.md](docs/go-libraries.md))

## Quick start

```bash
go run ./cmd/flag-lang repl
```

Inside the REPL:

```clojure
:import ["async.lib" :refer [future sleep]]
:load "examples/concurrency/main.flag"
```

```bash
go run ./cmd/flag-lang compile examples/hello/src/main.flag -o hello.go
go build -o /dev/null hello.go
go run ./hello.go
```

```bash
go run ./cmd/flag-lang build examples/hello -o hello
./hello
```

```bash
./build
flag --help
```

This builds `bin/flag-lang` and installs it as `flag` into `/usr/local/bin`
(override with `INSTALL_DIR=/path/to/bin` if needed).

```bash
go run ./cmd/flag-lang test examples/FRS
```
