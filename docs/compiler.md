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
  prologue macros such as `cond`, `->`, and `defrecord` expand the same way.
- Expansion walks lists, vectors, maps, sets, pipe-vectors, hash-fn bodies, and
  metadata. Quoted lists are left unchanged. Depth limit is 100.
- Substituted arguments are treated as literals so names like `rest` inside
  `(mapcat rest)` are not rewritten by `->>`.
- Throws become `{:kind :error :message ...}` on the output channel.

Canned fixture tests are `examples/compiler_tokenizer/macros/*.in` vs
`*.expected`, driven by `expand_macros_test.flag`. Helpers `slurp-text`,
`expand-fixture`, and `expected-fixture` live in
`examples/compiler_tokenizer/main.flag`.
