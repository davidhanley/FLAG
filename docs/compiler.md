# FLAG compiler libraries

Part of **[The FLAG Book](flag-book.md)**.

The self-hosted compiler lives in `libraries/compiler/`. `parse-file` is:

`tokenize-file` → `build-ast-from-tokens` → `expand-macros`.

The Go compiler embeds those libraries as `*_flag_gen.go`. Regenerate in order
after lowering changes (flattening `if`/`let`, etc.):

```
go run flag-lang/internal/compiler/tokengen
go run flag-lang/internal/compiler/astgen
go run flag-lang/internal/compiler/expandmacrosgen
go run flag-lang/internal/compiler/lowergen
```

`CompileProgram` (inlined imports) is what the generators run; they do not use
package-mode emit. Interned constants are prefixed per file (`flag*`,
`astflag*`, `macflag*`, `lowflag*`) so the four packages can share one Go
package. Tokenizer keeps referenced helpers (`inc`, `stdlib__second` /
`stdlib__third`), the `SourceToken` / `TokenState` / `ParseToken` records, and
renames Go-predeclared prelude names (`close` → `prologue_close`) so they do
not shadow builtins.

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
`:for`, `:block`, `:func-decl`, `:raw-stmt`. An IIFE is a `:call` of a `:func-lit` with no args.
`:binary` does not add parentheses (same as Go).

`compiler/lower.lib` exports `ast-node-to-ir`, `quoted-ast-to-ir`,
`call-to-ir`, `call-ast-to-ir`, `ctor-to-ir`, `runtime-call-to-ir`,
`fold-call-to-ir`, `symbol-to-ir`, `eval-ast-to-ir`, `if-to-ir`, `let-to-ir`,
`do-prelude-stmts`, `do-defer-to-ir`, `loop-to-ir`, `recur-to-ir`,
`defer-to-ir`, `throw-to-ir`, `future-to-ir`, `doto-to-ir`,
`update-bang-to-ir`, `catch-handler-to-ir`, `try-to-ir`, `doseq-to-ir`,
`doseq-body-to-ir`, `mapcat-binding-to-ir`, `fn-to-ir`, `ns-to-ir`,
`def-to-ir`, `defn-binding-to-ir`, `defn-to-ir`, and `defn-multi-to-ir`.

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

`let-to-ir` emits a result variable plus a Go `{ }` block that holds
bindings, body prelude statements, and `name = body-expr`. Production
`letExprToGo` uses this instead of a value IIFE, so nested `let` (and
`and`/`or`, which expand to `let`) no longer wrap in `func() T { ... }()`.
The block scopes bindings so they can shadow parameters and outer lets
without renaming. `let` still wraps in an IIFE when the body contains
`defer` (Go `defer` is function-scoped; `with-open` relies on that).
`for` / `doseq` MapCat callbacks now run the rest expression's statements
inside the binding function, so a body like `(or cell "")` (a flattened
`let`) cannot mention `cell` before it is bound.

`do-prelude-stmts` turns compiled `do` forms into flatten prelude: earlier
forms become discarded statements, the last form's statements stay with the
result expression. Production `doExprToGo` uses this unless a form is
`defer`. `do-defer-to-ir` wraps the same forms in an IIFE so Go `defer`
runs when that thunk returns (`with-open`). Empty `(do)` is still
`NilValue` in Go.

`loop-to-ir` emits a value IIFE: init statements, `var` bindings, then
`for { body; __loopResult := body-expr; if unwrap-recur { assign; continue }; return }`.
Production `loopExprToGo` still compiles bindings (symbols only) and the
body via `do`, then calls this. `recur-to-ir` is `flagrt.NewRecur(args...)`;
Go still checks `recur` is inside `loop` and arity matches.

`defer-to-ir` is `defer flagrt.Call(thunk)`. Production still compiles the
zero-arg function and keeps its statements on the `goExpr`. `throw-to-ir`
wraps `flagrt.Throw` plus `return NilValue` in a value IIFE so `throw` is an
expression. `future-to-ir` is `flagrt.NewFuture(func() flagrt.Value { body })`.

`doto-to-ir` binds `__doto` and `Call`s each step on it. `update-bang-to-ir`
assigns then returns the mutable binding. `try-to-ir` / `catch-handler-to-ir`
emit the recover/`CatchMatches` IIFE; Go still parses `catch` types and
compiles bodies. `doseq-to-ir` wraps `DoAll`; `mapcat-binding-to-ir` is the
`for`/`doseq` MapCat callback. `fn-to-ir` wraps `NewFunction` around a
variadic `func(args ...flagrt.Value)`: arity panic (`exactly` vs rest
`at least`), `p := args[i]` or `_ = args[i]`, init/body statements, then
`return`. Production `compileLambda` still binds parameters (including `&`
and destructure) and compiles the body; `#()` still rewrites placeholders
in Go, then calls `compileLambda`.

`ns-to-ir` is the `// Source namespace:` comment. `def-to-ir` is a
top-level `var name = expr`. `defn-to-ir` emits named `:func-decl`s: the
typed `goName_arity_N` function plus a variadic wrapper, or one rest
variadic. `defn-multi-to-ir` emits each arity function and a `len(args)`
dispatch. `defn-binding-to-ir` is `var name = NewFunction(variadic)`.
Production still parses names, docs, `&`/destructure, and compiles bodies;
FLAG assembles the declarations. `deftest` and `go-interface` still use
the Go printer. Go still compiles subforms (`exprToGo`); FLAG assembles
the trees.

`and` / `or` are prologue macros, not compiler special forms. They expand
to `let` plus `if` so each clause is evaluated at most once and later
clauses short-circuit: `(or a b)` becomes `(let [or-tmp a] (if or-tmp
or-tmp b))`. Empty `(or)` is `nil`; empty `(and)` is `true`. Two-argument
clauses are explicit so `(or nil [])` keeps `[]` (a lone `& rest` splice
would turn the empty vector into `(or)` → `nil`). Flattened `let` means
those expansions are also result-var / block, not IIFEs.

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
`go`, `var`/`:=`/`=`, `if`, `for`, `{ }` blocks, and preformatted raw lines.
`renderIRExpr` / `renderIRStmt` are the only Go-string printers.

Expression lowering is on IR: literals, calls, collections, `if`/`do`/`let`/
`loop`/`recur`/`try`/`throw`/`defer`/`go`/`doto`/`update!`, `for`/`doseq` MapCat
IIFEs, destructure bindings, and `future`. `if`/`let`/`do`/`loop`/`recur`/`defer`/`throw`/`future`/`doto`/`update!`/
`try`/`doseq`/`for` MapCat assembly is FLAG. Remaining string concat
is the Go file printer (top-level `func`/`var` emission), not per-form
lowering.

## Package emit (`flag-lang build`)

`CompileProgram` still inlines the import graph into one `package main` (used
by `compile`, `test`, REPL, and code generators). `flag-lang build` uses
`CompileProgramPackages` / `WriteProgramPackages`: one Go package per FLAG
module, entry as `package main`, libraries imported as `flagbuild/<pkg>`.
Exported names are capitalized (`add` → `math.Add`). The prelude is
`flagbuild/prologue` (`inc` → `prologue.Inc`), not copied into importers.
Each library package also emits a `.flagi` (header, exported macros,
`(declare …)`). See [modules.md](modules.md).
