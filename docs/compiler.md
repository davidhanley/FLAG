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
`fold-call-to-ir`, `eq-to-ir`, `nil-pred-to-ir`, `empty-pred-to-ir`,
`count-to-ir`, `str-to-ir`, `ex-info-to-ir`, `assoc-to-ir`, `conj-to-ir`,
`seq-to-ir`, `drop-to-ir`, `take-to-ir`, `cons-to-ir`, `get-to-ir`, `nth-to-ir`,
`not-empty-to-ir`, `type-of-to-ir`, `vec-to-ir`, `dissoc-to-ir`,
`contains-pred-to-ir`, `into-to-ir`, `concat-to-ir`, `ex-message-to-ir`,
`ex-data-to-ir`, `ex-cause-to-ir`, `println-to-ir`, `compare-to-ir`,
`numeric-eq-to-ir`, `comparison-to-ir`, `symbol-ctor-to-ir`, `name-to-ir`,
`keyword-ctor-to-ir`, `format-to-ir`, `map-to-ir`, `sort-by-to-ir`,
`apply-to-ir`, `pmap-to-ir`, `filter-to-ir`, `reduce-to-ir`, `doall-to-ir`,
`some-to-ir`, `set-to-ir`, `line-seq-to-ir`, `repeat-to-ir`, `rand-int-to-ir`,
`to-json-to-ir`, `from-json-to-ir`, `open-file-to-ir`, `file-to-strings-to-ir`,
`file-to-strings-path-to-ir`, `go-fn-to-ir`, `go-fn-args-to-ir`,
`symbol-to-ir`, `eval-ast-to-ir`, `if-to-ir`, `let-to-ir`,
`do-prelude-stmts`, `do-defer-to-ir`, `loop-to-ir`, `recur-to-ir`,
`defer-to-ir`, `throw-to-ir`, `future-to-ir`, `doto-to-ir`,
`update-bang-to-ir`, `catch-handler-to-ir`, `try-to-ir`, `doseq-to-ir`,
`doseq-body-to-ir`, `mapcat-binding-to-ir`, `fn-to-ir`, `ns-to-ir`,
`def-to-ir`, `defn-binding-to-ir`, `defn-to-ir`, `defn-multi-to-ir`,
`write-to-ir`, `ends-with-to-ir`, `testing-to-ir`, `is-to-ir`,
`expect-exception-to-ir`, `ast-form-source`, `hash-fn-placeholder-index`,
`map-destructure-key-ir`, `go-ident`, `bind-params-to-ir`,
`compile-form-to-ir`, `compile-forms-to-ir`, `form-defer-result`,
`forms-have-defer?`, `list-head-name`, `catch-type-name`,
`parse-catch-clause`, and `split-try-args`.

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

`eq-to-ir` is `(= a b c …)`: at least two arguments, string args wrapped
with `NewString`, adjacent values chained as `flagrt.Eq` joined with `&&`,
then `flagrt.NewBool`. Production `compile-form-to-ir` uses this for `=`.
Self-calls use the same direct Go ident as production (`goName_arity_N` or
the variadic wrapper), not `flagrt.Call`.

`nil-pred-to-ir` / `empty-pred-to-ir` wrap `flagrt.IsNil` / `IsEmpty` in
`NewBool`. `count-to-ir` is `NewLong(int64(Count(x)))`. Each takes one
argument; string args become `NewString`. Production `compile-form-to-ir`
handles `nil?`, `empty?`, `count`, `throw` (existing `throw-to-ir` IIFE),
and `loop`/`recur` (symbol bindings, `:mutable-value` locals,
`:loop-binding-names` for arity). Nested `fn` clears loop names so `recur`
cannot escape.

`str-to-ir` is `(str …)`: `flagrt.Str` of the raw argument expressions (string
literals stay Go strings, not `NewString`). The result kind is `:string`.
`ex-info-to-ir` builds `NewMap` of `:message` (wrapped `NewString`), `:data`,
and optional `:cause`. Production `compile-form-to-ir` requires a string
message and coerces data/cause to Value. Keywords are inline `NewKeyword`
(Go still interns them when `constants` is set). `assoc-to-ir` /
`conj-to-ir` are `flagrt.Assoc` / `Conj`. `assoc` needs a collection plus
odd-length key/value pairs and rejects non-Value args (no string wrap).
`conj` needs a collection plus at least one item and wraps strings.

`seq-to-ir` is `flagrt.Seq` of one Value (strings wrapped). `drop-to-ir` /
`take-to-ir` / `cons-to-ir` are two-arg `Drop` / `Take` / `Cons` with no
string wrap. `get-to-ir` / `nth-to-ir` are `Get` / `Nth` with two or three
args and string wrapping. Compile-form mode `:seq-op` is the `seq` special
form; `:seq` still means “compile an argument list”.

`not-empty-to-ir` is strict-Value `NotEmpty`. `type-of-to-ir` / `vec-to-ir`
wrap strings like `seq`. `dissoc-to-ir` is `MapDissoc` (≥1 arg, no string
wrap). `contains-pred-to-ir` is `NewBool(Contains)` (exactly two Values).
`(list)` / `(array)` / `(hash-map)` use `ctor-to-ir` (`NewList` / `NewArray` /
`NewMap`) and wrap strings; `hash-map` needs even key/value pairs. `into-to-ir`
is two collection Values. `concat-to-ir` wraps strings. `ex-message-to-ir` /
`ex-data-to-ir` / `ex-cause-to-ir` wrap strings. `println-to-ir` is `str`
args (`Println`) with result kind `:value`.

`compare-to-ir` is two-arg `Compare` (strings wrapped). `numeric-eq-to-ir` is
`NewBool(NumericEq(...))` with no string wrap. `comparison-to-ir` chains
`Lt`/`Le`/`Gt`/`Ge` with `&&`. Compile-form wraps that in `NewBool` so `let`/`and`
stay Values (Go’s `exprToGo` path still uses a raw `:bool`). `symbol-ctor-to-ir` /
`name-to-ir` / `keyword-ctor-to-ir` take a string or Value (`Name` is
`:string`). `format-to-ir` is `NewString(Format(fmt, args…))` with a string
format. `map`/`sort-by`/`apply`/`pmap`/`filter`/`reduce`/`some` are Value-only
runtime calls. `doall`/`set`/`line-seq`/`rand-int` are unary strict Values.
`repeat` wraps strings. `to-json-to-ir` is unary strict with result `:string`;
`from-json-to-ir` takes a string. `open-file`/`go-fn`/`go-fn-args` pass strings
through or wrap Values with `Name`. `file-to-strings` uses `FileToStringsPath`
for strings and `FileToStrings` for Values.

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
Production `compile-form-to-ir` compiles bindings (symbols only) and the
body via `do`, then calls this; Go `loopExprToGo` is the fallback. `recur-to-ir`
is `flagrt.NewRecur(args...)`; FLAG and Go both require `recur` inside `loop`
with matching arity.

`defer-to-ir` is `defer flagrt.Call(thunk)`. Production `compile-form-to-ir`
compiles the zero-arg function (strings wrapped), sets `:defer?` / `:stmt`,
and keeps thunk statements on the form. A `do` / `:forms` body that contains
`defer` wraps with `do-defer-to-ir` so the thunk runs when that IIFE returns.
`throw-to-ir` wraps `flagrt.Throw` plus `return NilValue` in a value IIFE so
`throw` is an expression. `future-to-ir` is
`flagrt.NewFuture(func() flagrt.Value { body })`.

`doto-to-ir` binds `__doto` and `Call`s each step on it. Production compiles
the target and each step to Value. `update-bang-to-ir` assigns then returns
the mutable binding. `try-to-ir` / `catch-handler-to-ir` emit the
recover/`CatchMatches` IIFE. Production parses `catch` types
(`ExceptionInfo`, `Exception`, `Throwable`, `:default`), compiles bodies, and
skips the IIFE when there is no `catch`/`finally`. `doseq-to-ir` wraps
`DoAll`; `mapcat-binding-to-ir` is the `for`/`doseq` MapCat callback.
Production `doseq`/`for` take a binding vector of symbol/collection pairs
(strings wrapped); empty `for` bindings are `NewArray(body)` and empty
`doseq` bindings discard the body then `DoAll`. `fn-to-ir` wraps `NewFunction` around a
variadic `func(args ...flagrt.Value)`: arity panic (`exactly` vs rest
`at least`), `p := args[i]` or `_ = args[i]`, init/body statements, then
`return`. `#()` rewrites `%` / `%N` placeholders (nested `#()` kept
intact) to `__pN` and lowers as `fn`. `.write` is `flagrt.Write` (file
Value, stringish content). `.endsWith` is `strings.HasSuffix` with
result kind `:bool`. Other `.Foo` methods stay unhandled. `testing`
discards its body in a `NilValue` IIFE. `is` panics unless the test is
truthy (optional string message; default is `ast-form-source`).
`expect-exception` recovers in a deferred thunk and optionally
`strings.Contains` the panic text.

`go-ident` maps FLAG names to Go identifiers (`-` → `_`, `?` → `_q`,
`!` → `_bang`, `>` → `_gt`, `<` → `_lt`, `main` → `flag_main`, keywords
get a trailing `_`). `bind-params-to-ir` binds `fn`/`defn` parameter
vectors: symbols, unused `_`, `& rest` (`NewArray`, or `NewMap` for a
map rest pattern), vector destructure (`SeqFirst` / `SeqRest`), and map
destructure (`Get` with `:as` / `:or` / `:keys` / `:syms` / `:strs`).
`:or` defaults must be literals (`map-destructure-key-ir` uses
`NewKeyword`, not interned idents). `let` uses the same patterns.

`compile-form-to-ir` / `compile-forms-to-ir` lower a form or `do` body:
literals, collections, symbols (including `pkg/name` via the module
table), `if`, `let` (with `^{:volatile true}` and vector/map
destructure), `fn`, `#()`, `update!`, `=`,
`nil?`, `empty?`, `count`, `throw`, `loop`/`recur`, `str`, `ex-info`,
`assoc`, `conj`, `seq`, `drop`/`take`/`cons`, `get`/`nth`, `not-empty`,
`type-of`, `vec`, `dissoc`, `contains?`, `list`/`array`/`hash-map`, `into`,
`concat`, `ex-message`/`ex-data`/`ex-cause`, `println`, `compare`/`==`/`<`/`>` ,
`symbol`/`name`/`keyword`, `format`, `map`/`filter`/`reduce`/`apply`/`pmap`/
`sort-by`/`some`/`set`/`doall`, `line-seq`/`repeat`/`rand-int`, `to-json`/
`from-json`, `open-file`/`file-to-strings`/`go-fn`/`go-fn-args`, `defer`,
`doto`, `try`/`catch`/`finally`, `doseq`, `for`, `.write` / `.endsWith`,
`testing` / `is` / `expect-exception`, folded `+ - * /`, and other
runtime ops already in FLAG.
Unknown `.method` calls still throw `not handled by FLAG`.
`list-head-name`, `catch-type-name`, `parse-catch-clause`, and `split-try-args`
parse `try` clauses; `form-defer-result` / `forms-have-defer?` mark `defer`
forms so `:forms` can wrap them.
Production `compileDefn` / `compileLambda` / `bindLambdaParamStmts` try
FLAG first and fall back to Go on those errors. FLAG preserves
`:mutable-value` locals across nested `fn` parameter binding so `update!`
still sees volatile lets.

`ns-to-ir` is the `// Source namespace:` comment. `def-to-ir` is a
top-level `var name = expr`. `defn-to-ir` emits named `:func-decl`s: the
typed `goName_arity_N` function plus a variadic wrapper, or one rest
variadic. `defn-multi-to-ir` emits each arity function and a `len(args)`
dispatch. `defn-binding-to-ir` is `var name = NewFunction(variadic)`.
Production still parses names, docs, and `&`/destructure. `deftest` and
`go-interface` still use the Go printer.

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
