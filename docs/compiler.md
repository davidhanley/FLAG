# FLAG compiler libraries

Part of **[The FLAG Book](flag-book.md)**.

The self-hosted compiler lives in `libraries/compiler/`. `parse-file` is:

`tokenize-file` → `build-ast-from-tokens` → `expand-macros`.

## `expand-macros`

`compiler/expand-macros.lib` exports `expand-macros`, which takes a channel of AST
maps and returns a channel of expanded ASTs.

A background `go` loop:

1. Receives the next form (`nil` closes the output channel).
2. If the form is `defmacro` (or expands to one), compiles it and stores it in
   the expander’s macro table. Definitions are not forwarded.
3. Otherwise expands the form with macros defined so far and sends the result.

The Go compiler embeds this library (`expand_macros_flag_gen.go`) and runs it
after parsing: prologue and imported `defmacro` forms are seeded, then module
forms are expanded. `defmacro` is not compiled to Go; it only feeds later
expansion.

Behavior:

- Single-arity `(defmacro name "doc?" [params] body)` and multi-arity
  `(defmacro name "doc?" ([params] body...) ...)`. Extra arity body forms are
  wrapped in `do`. `& rest` parameters splice into lists, vectors, and
  pipe-vectors (not maps, sets, or quoted lists).
- `macro-case` and `macro-defrecord` run during template substitution, so
  prologue macros such as `cond`, `and`, `or`, `->`, and `defrecord` expand
  the same way.
- Expansion walks lists, vectors, maps, sets, pipe-vectors, hash-fn bodies, and
  metadata. Quoted lists are left unchanged. Depth limit is 100.
- Substituted arguments are treated as literals so names like `rest` inside
  `(mapcat rest)` are not rewritten by `->>`.
- Throws become `{:kind :error :message ...}` on the output channel.

Canned fixture tests are `examples/compiler_tokenizer/macros/*.in` vs
`*.expected`, driven by `expand_macros_test.flag`. Helpers `slurp-text`,
`expand-fixture`, and `expected-fixture` live in
`examples/compiler_tokenizer/main.flag`.

## FLAG IR (`compiler/ir.lib`)

`libraries/compiler/ir.lib` is the FLAG form of `internal/compiler/ir.go`. Nodes
are maps with a `:kind` keyword. `render-ir` / `render-ir-stmt` /
`render-ir-stmts` must match `renderIRExpr` / `renderIRStmt` on the same trees.

Constructors (`ir-ident`, `ir-call`, `rt-call`, `iife`, …) build those maps.
Go can ingest them with `flagValueToIRExpr` / `flagValueToIRStmt`.

Expression kinds: `:ident`, `:string`, `:int`, `:selector`, `:call`, `:index`,
`:slice`, `:spread`, `:unary`, `:binary`, `:func-lit`, `:raw`. Statement kinds:
`:expr-stmt`, `:return`, `:defer`, `:go`, `:var`, `:assign`, `:define`, `:if`,
`:for`, `:raw-stmt`. An IIFE is a `:call` of a `:func-lit` with no args.
`:binary` does not add parentheses (same as Go).

`compiler/lower.lib` exports `ast-node-to-ir`, `quoted-ast-to-ir`,
`call-to-ir`, `call-ast-to-ir`, `ctor-to-ir`, `runtime-call-to-ir`,
`fold-call-to-ir`, `symbol-to-ir`, `eval-ast-to-ir`, and `if-to-ir`.

`ast-node-to-ir` lowers FLAG AST literal maps (`:int`, `:string`, `:char`,
`:bigint`, `:ratio`, `:float`, `:keyword`, `:quoted-symbol`, and symbols
`true`/`false`/`nil`) to IR maps. The Go compiler calls this from `exprToGo`
(`flagLiteralToIR`) and still hoists keywords, chars, bigints, ratios, and
quoted symbols with `internIR`. String literals stay `:string`
(`exprKindString`); other literals become `flagrt.New*` calls.

`quoted-ast-to-ir` lowers quoted collections: `:quoted-list` / `:list` →
`NewList`, `:vector` → `NewArray`, `:pipe-vector` → `NewVector`, `:map` →
`NewMap`, `:set` → `NewSet`. Nested symbols become `NewSymbol` (including
`true`/`false`/`nil`). Go `quotedLiteralToIR` calls this; quoted lists are
still interned as `List`. Odd-length maps error.

`call-to-ir` builds `flagrt.Call(callee, args...)` from IR maps. Go
`callExprToGo` still resolves the callee and arguments (locals, interned
keywords, self-calls stay direct Go idents), then uses `call-to-ir` for the
generic Call node. `call-ast-to-ir` lowers a `:list` whose head is not a
special form when every subform is a literal, collection, or nested generic
call; unresolved symbols and special forms (`if`, `+`, `fn`, …) error so Go
can keep those paths.

`ctor-to-ir` builds `NewArray` / `NewMap` / `NewVector` / `NewSet` (and
`(list)` / `(array)` ctors). Go still lowers each element, wraps strings,
and interns constant collections; FLAG emits the call node.

`runtime-call-to-ir` / `fold-call-to-ir` emit simple runtime ops. Go keeps
arity checks; FLAG emits `flagrt.First`, folded `flagrt.Add`, `BitAnd`, etc.

`symbol-to-ir` takes the FLAG name, Go ident (from `toGoIdentifier`), and a
context map (`:locals`, `:globals`, `:functions`, `:module`, `:go-fns`,
`:self-function-name`, `:self-variadic-name`). It matches Go resolution:
locals, self `NewFunction`, module table, globals, functions,
`BuiltinFunction`, then Go bindings. `eval-ast-to-ir` uses that context to
lower literals, evaluated collections, runtime ops, and generic calls.

`if-to-ir` emits a result variable plus a Go `if` that assigns each branch:
`var name T; if cond { name = then } else { name = else }`. Nested branch
statements go inside the corresponding arm; condition statements run first.
Production `ifExprToGo` uses this instead of a value IIFE, so nested `if`
forms no longer wrap each other in `func() T { ... }()`. Result-var
statements travel with the expression (`goExpr.stmts`) and are emitted
before the use site (including `=`, `is`, and `expect-exception`). `do`
flattens to prelude + last expression unless it contains `defer`, which
still wraps in an IIFE so the thunk runs when `do` returns.

`and` / `or` are prologue macros, not compiler special forms. They expand
to `let` plus `if` so each clause is evaluated at most once and later
clauses short-circuit: `(or a b)` becomes `(let [or-tmp a] (if or-tmp
or-tmp b))`. Empty `(or)` is `nil`; empty `(and)` is `true`. Two-argument
clauses are explicit so `(or nil [])` keeps `[]` (a lone `& rest` splice
would turn the empty vector into `(or)` → `nil`). Because `let` is still
an IIFE, nested `and`/`or` still wrap in `func() T { ... }()`.

`compiler/codegen.lib` already lowers its toy `+ - * /` subset to IR maps, then
renders (wrapping binaries in `()` because FLAG arithmetic needs grouping).

Golden tests live in `examples/compiler_tokenizer/main_test.flag` next to the
Go tests in `internal/compiler/ir_test.go`.

## Go lowering IR

After expansion, the Go compiler still lowers FLAG forms to Go. Expression
nodes are maps in FLAG (`compiler/ir.lib`) and structs in Go (`ir.go`). The
goal is for FLAG lowering to emit the same trees instead of concatenating Go
source.

`IRExpr` nodes today: ident, string, int, `pkg.Name` selector, call, index,
slice, spread (`expr...`), unary `!(x)`, binary `left op right`, anonymous
`func` literals (optional parameter list), and `IRRaw`. An IIFE is a call of a
func literal with no arguments. `IRStmt` covers `_ = expr`, `return`, `defer`,
`go`, `var`/`:=`/`=`, `if`, `for`, and preformatted raw lines.
`renderIRExpr` / `renderIRStmt` are the only Go-string printers.

Expression lowering is on IR: literals, calls, collections, `if`/`do`/`let`/
`loop`/`try`/`throw`/`defer`/`go`/`doto`/`update!`, `for`/`doseq` MapCat
IIFEs, destructure bindings, and `future`. Remaining string concat
is the Go file printer (top-level `func`/`var` emission), not per-form
lowering.
