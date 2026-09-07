# FLAG TODO (Clojure feature backlog)

## High priority

- [x] **Builtinize `merge`** (it exists in `internal/compiler/prologue.flag` today). Make it a runtime builtin so it is always available, faster, and easier to optimize/document consistently.
- [x] Add **`assoc-in, dissoc-in`** for nested map updates.
- [x] Add **`update-in`** for nested key-path transforms.
- [x] Add **`merge-with`** for key conflict resolution.
- [x] Expand `assoc` to match Clojure behavior on **vectors** (`(assoc [a b] 1 x)`), not only maps.
- [x] Add **`vals`** (map value collection).
- [x] Add **`find`** (map entry lookup as pair / nil).
- [x] Add **`dissoc` parity behaviors** around nil/no-op edge cases to match Clojure exactly.

## Sequence + collection APIs

- [x] Add **`nth`** (`not-found` arity included).
- [x] Add **`peek` / `pop`** as FLAG aliases of `first` / `rest`.
- [x] Add **`interpose`**.
- [x] Add **`interleave`**.
- [x] Add **`partition`** and **`partition-all`**.
- [x] Add **`partition-by`**.
- [x] Add **`distinct`**.
- [x] Add **`flatten`** (or `tree-seq` + `flatten` strategy).
- [x] Add **`map-indexed`** and **`keep-indexed`**.
- [x] Add **`reduce-kv`** for map/vector keyed reduction.
- [x] Add **`sort`** (plain comparator form; `sort-by` already exists).
- [ ] Add vector-specialized helpers **`mapv`** and **`filterv`**.

## Predicates + type/core helpers

- [x] Add **`true?`** and **`false?`**.
- [x] Add **`zero?`**, **`pos?`**, **`neg?`**, **`number?`**, **`int?`**, **`float?`**.
- [x] Add **`string?`**, **`keyword?`**, **`symbol?`**, **`map?`**, **`vector?`**, **`set?`**, **`sequential?`**, **`coll?`**.
- [x] Add **`even?`** / **`odd?`** as core helpers (common enough to be built in, not user-defined per project).
- [x] Add **`empty`** (return same-type empty collection).

## Functional combinators

- [x] Add **`complement`**.
- [x] Add **`fnil`**.
- [x] Add **`every-pred`**.
- [x] Add **`some-fn`**.
- [x] Add **`iterate`**.
- [x] Add **`repeatedly`**.
- [x] Add **`comp` function parity** (current prologue macro version is limited to unary composition).

## Sets

- [x] Add **`clojure.set`-style ops**: `union`, `intersection`, `difference`, `subset?`, `superset?`, `rename-keys`.

## Reader/language-level parity

- [ ] Add **syntax-quote / unquote / unquote-splicing** (`` ` ``, `~`, `~@`) for macro ergonomics.
- [ ] Add **gensym (`x#`) support** in syntax-quoted forms.
- [ ] Add **namespaced map literal** support (`#:user{:id 1}`) if feasible.
- [ ] Add **metadata round-trip semantics** (`with-meta`, `meta`) if value model supports it.

## State model

- [x] Add **atom-like refs** (`atom`, `deref`, `reset!`, `swap!`) as core state primitive.
- [ ] Add **`compare-and-set!`** behavior for lock-free coordination.
- [ ] Clarify interaction of `update!` (mutable let bindings) vs Clojure-style state refs and document migration guidance.

## Runtime/perf follow-ups

- [x] channel-some? / channel-every? close the input on short-circuit (FLAG in async.lib).
- [ ] Records: make keyword lookup fast (no reflection). Today `(:field rec)` walks struct fields and `flag` tags on every get.
- [ ] Typed defrecord constructors must reject wrong field types consistently (constructor + map->record paths).

## Docs/tests parity tasks

- [ ] Update `flag-lang.md` builtin lists to match actual implemented surface (currently stale in places).
- [ ] Add language tests for each new core fn/macro and edge-case parity tests vs Clojure where behavior intentionally matches.
- [ ] Add a "Clojure parity matrix" page showing: implemented, partial, planned, and intentionally different semantics.