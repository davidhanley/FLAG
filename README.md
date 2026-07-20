# flag-lang

`flag-lang` is the starting point for a new Lisp that aims to stay close to Clojure at the source level while compiling to Go instead of the JVM.

This repository currently provides:

- a small `flag-lang` compiler CLI
- a minimal compiler package
- support for a tiny starter subset of forms
- an example source program

## Current language subset

The initial scaffold intentionally keeps the language surface small while the project structure settles. The compiler currently understands:

- `(ns some.namespace)` at the top of a file
- `(println "...")`
- `(print "...")`
- `(+ a b ...)` numeric expressions
- `(defn name[params] body)` with numeric params/body expressions
- function calls such as `(name 47)`

Unsupported forms return explicit compiler errors.

Parsing is now AST-based and Clojure-like: whitespace/newlines are insignificant, `;` comments are supported, and a source file can contain multiple top-level expressions.

## Project layout

```text
cmd/flag-lang/          CLI entrypoint
internal/compiler/      FLAG source to Go code generation
runtime/numerics.go     numeric Value runtime
runtime/list.go         linked-list Value runtime
examples/               Sample FLAG programs
```

`runtime.Value` now supports:
- tagged long and double numbers
- tagged linked lists backed by Go's built-in `container/list` package (pointer in `p`, length stored in numeric field)

## Quick start

```bash
go run ./cmd/flag-lang compile examples/hello.flag -o hello.go
go run ./hello.go
```

## Near-term direction

The scaffold is set up to grow toward:

- richer Clojure-compatible reading and parsing
- an intermediate representation for analysis and transforms
- Go code generation beyond print forms
- namespaces, vars, functions, collections, and control flow

## Runtime benchmark prototypes

The repository also includes a small benchmark package for comparing two `Value` dispatch strategies for numeric operations:

- vtable-style double dispatch
- tag-byte dispatch with `switch`

Both variants model numeric values with a `d` field plus an `unsafe.Pointer` field so you can experiment with the layout you described for `Value`.

Run the benchmarks with:

```bash
go test ./internal/runtimebench -bench . -benchmem
```

The benchmark set covers:

- scalar `long + long`
- scalar `long + double`
- reducing an array of numeric values
- reducing a mixed array of `long` and `double` values
