# The FLAG Book

A guided path through the FLAG language documentation. Each chapter is a standalone
doc; this file is the table of contents and suggested reading order.

FLAG is a Clojure-inspired Lisp that **compiles to Go** and ships as small native
binaries—not a JVM image with an always-on evaluator.

---

## How to use this book

| Goal | Start here |
|------|------------|
| Install / run something | [Chapter 1 — Introduction](#1-introduction) |
| Language reference | [Chapter 2 — Language manual](#2-language-manual) |
| Interactive workflow | [REPL guide](#repl-guide) |
| Multi-file programs & imports | [Chapter 3 — Modules](#3-modules) |
| Wrapping Go packages | [Chapter 4 — Go libraries](#4-go-libraries) |
| Goroutines, futures, channels | [Chapter 5 — Async](#5-async) |
| Worked examples | [Appendices](#appendices--examples) |

---

## Chapters

### 1. Introduction

**[README.md](../README.md)**

Project goals, current state, and quick start (`repl`, `compile`, `build`, `test`).

Themes that run through the rest of the book:

- Fast startup, low memory, small standalone binaries (~10MB class)
- No VM in the deploy artifact
- Expressiveness of Lisp with Go’s operational model

---

### 2. Language manual

**[flag-lang.md](../flag-lang.md)**

What is implemented **today**: CLI commands, core syntax, special forms, data
literals, destructuring, builtins, numerics, sequences, testing, and pointers into
modules / async / libraries.

Read this for day-to-day “how do I write FLAG?”

---

## REPL guide

**[repl.md](repl.md)**

Interactive usage, including:

- evaluating forms in a live session
- `:import` for libraries and modules
- `:load` for loading code from files

---

### 3. Modules

**[modules.md](modules.md)**

File-based modules (not Clojure classpath `ns` soup):

- Header map: `:namespace`, `:exports`, `:imports`
- Qualified imports by default; `:as` and `:refer`
- Private by default; exports are an allowlist
- Macro exports and `libraries/*.lib`
- Build / import graph vs legacy file merge

---

### 4. Go libraries

**[go-libraries.md](go-libraries.md)**

How FLAG wraps native Go without open host interop:

- Three layers: pure Go → adapter (box/unbox) → `.lib` surface
- No `any` kitchen-sink APIs in pure packages
- `:go-exports` and static adapters
- CSV as the reference library; options maps (e.g. `:fields-per-record`)

---

### 5. Async

**[async.md](async.md)**

Concurrency as a **library**, not core:

- Import `async.lib` (`:refer` or `async/…`)
- `go`, `future` (macros)
- `sleep`, channels, `select`
- Design notes (callable futures, non-blocking select, small core)

---

## Appendices — examples

| Example | Doc | What it shows |
|---------|-----|----------------|
| Hello | [examples/hello](../examples/hello) | Minimal program |
| Modules | [examples/modules/README.md](../examples/modules/README.md) | Multi-file import, `:as`, `:refer` |
| Concurrency | [examples/concurrency/README.md](../examples/concurrency/README.md) | `async.lib` end-to-end + tests |
| FRS | [examples/FRS](../examples/FRS) | Larger real app (CSV, HTML, scoring) |

Run examples with:

```bash
go run ./cmd/flag-lang build path/to/main.flag -o /tmp/app
go run ./cmd/flag-lang test path/to/example   # when tests exist
```

---

## Related (not chapters)

| File | Role |
|------|------|
| [AGENTS.md](../AGENTS.md) | Instructions for coding agents working in this repo |
| [LICENSE](../LICENSE) | License |

---

## Suggested paths

**New to FLAG (half day)**  
1 → 2 (skim) → hello example → 3 (header + import) → modules example.

**Shipping a tool that uses CSV / HTML**  
2 → 3 → 4 → FRS or a thin app with `csv.lib` / `burp.lib`.

**Concurrent services**  
2 → 3 → 5 → concurrency example → 4 if you wrap more Go libs (e.g. Kafka).

**Contributing to the compiler / runtime**  
All chapters, then the tree under `internal/compiler` and `runtime/`.

---

## Status

This book tracks the **current** repo. FLAG is evolving; when behavior and docs
diverge, prefer the implementation and update the chapter.

Last structure update: chapters linked to existing docs as of the async / modules
library work.
