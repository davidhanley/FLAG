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

Unsupported forms return explicit compiler errors.

## Project layout

```text
cmd/flag-lang/          CLI entrypoint
internal/compiler/      FLAG source to Go code generation
examples/               Sample FLAG programs
```

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
