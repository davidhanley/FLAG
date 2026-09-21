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

- [x] Update `flag-lang.md` builtin lists to match actual implemented surface (currently stale in places).
- [ ] Add language tests for each new core fn/macro and edge-case parity tests vs Clojure where behavior intentionally matches.
- [ ] Add a "Clojure parity matrix" page showing: implemented, partial, planned, and intentionally different semantics.

## Control-flow macros a Clojure programmer will type

Already in prologue: `when`, `when-not`, `when-let`, `if-let`, `if-not`, `if-some`, `when-some`, `cond`, `case`, `condp`, `->`, `->>`, `some->`, `some->>`, `cond->`, `cond->>`, `as->`, `dotimes`.

- [x] Add **`if-let`** / **`if-not`** / **`if-some`** / **`when-some`**.
- [x] Add **`cond->>`** (thread last only when the test is truthy).
- [x] Add **`as->`** (named-binding thread).
- [x] Add **`condp`**.
- [x] Add **`dotimes`** (not `while` — open-ended imperative loops are discouraged; use `loop`/`for`/`doseq`).
- [ ] Add **`comment`** is done; add **`declare`**, **`defonce`**.

## Math (`clojure.core` / `Math`)

Have: `+` `-` `*` `/` `%` `quot` `rem` `mod` `max` `min` `rand-int` `rand` `rand-nth` `shuffle` `double` `numerator` `denominator` `bit-and` `bit-or` `bit-xor` `bit-not` `bit-shift-left` `bit-shift-right` `unsigned-bit-shift-right` `bit-test` `bit-set` `bit-clear` `bit-flip` `math/abs` `math/sqrt` `math/pow` `math/exp` `math/log` `math/log10` `math/sin` `math/cos` `math/tan` `math/floor` `math/ceil` `math/round` `math/IEEE-remainder` `inc`/`dec` `zero?` `pos?` `neg?` `even?` `odd?` `compare` `==`.

- [x] Add **`quot`**, **`rem`**, **`mod`** (Clojure names; FLAG `%` is `mod`).
- [x] Add **`compare`** and numeric **`==`**.
- [x] Add **`rand`**, **`rand-nth`**, **`shuffle`**.
- [x] Add **`numerator`** / **`denominator`** for ratios.
- [x] Add bit ops: **`bit-and`**, **`bit-or`**, **`bit-xor`**, **`bit-not`**, **`bit-shift-left`**, **`bit-shift-right`**, **`unsigned-bit-shift-right`**, **`bit-test`**, **`bit-set`**, **`bit-clear`**, **`bit-flip`**.
- [x] Add common `Math` surface as `math/…` (or core aliases): **`sqrt`**, **`pow`**, **`exp`**, **`log`**, **`log10`**, **`sin`/`cos`/`tan`**, **`floor`**, **`ceil`**, **`round`**, **`IEEE-remainder`**.
- [x] Make **`inc` / `dec` first-class functions** (or add function variants) so `(map inc xs)` and `(swap! a inc)` work. Today they are macros.

## Strings (`clojure.core` / `clojure.string`)

Have: `str`, `format`, `subs`, `str/trim`, `str/triml`, `str/trimr`, `str/trim-newline`, `str/replace`, `str/replace-first`, `str/escape`, `str/split`, `str/split-lines`, `str/join`, `str/blank?`, `str/includes?`, `str/index-of`, `str/last-index-of`, `str/starts-with?`, `str/ends-with?`, `str/upper-case`, `str/lower-case`, `str/capitalize`, `str/reverse`.

- [x] Add **`subs`**.
- [x] Add **`str/lower-case`**, **`str/triml`**, **`str/trimr`**, **`str/trim-newline`**.
- [x] Add **`str/includes?`**, **`str/index-of`**, **`str/last-index-of`**.
- [x] Add **`str/replace-first`**, **`str/split-lines`**, **`str/reverse`**.
- [ ] Add regex seq helpers: **`re-find`**, **`re-seq`**, **`re-find`** groups / **`re-matches`** already exists.

## Sequence extras

- [ ] Add **`mapv`** / **`filterv`** (still open above).
- [x] Add **`take-while`**, **`drop-while`**, **`take-last`**, **`drop-last`**, **`take-nth`**.
- [ ] Add **`split-at`**, **`split-with`**.
- [ ] Add **`cycle`**, **`reductions`**, **`frequencies`**, **`butlast`**.
- [ ] Add **`ffirst`**, **`nfirst`**, **`nnext`**, **`fnnext`**.
- [ ] Add **`lazy-seq`** / **`lazy-cat`** (or document that FLAG lazy lists are the substitute).
- [ ] Add **`reduced`** / **`ensure-reduced`** / **`reduced?`** so `reduce` can short-circuit.

## Exceptions / `try`

- [x] Add **`try` / `catch` / `finally`**. Today: `throw`, `ex-info`, `defer`, tests-only `expect-exception`.
- [x] Add **`ex-message`**, **`ex-data`**.

## Unintended incompatibilities (fix or document loudly)

These surprise Clojure programmers and are **not** intentional FLAG design (unlike `future` returning a 0-arg fn, atoms without watches, channels in `async.lib`, `| |` vectors vs `[ ]` arrays).

- [x] **`(range n)`** is infinite starting at `n`. Clojure’s `(range n)` is `0 .. n-1`. Either match Clojure or make `flag-lang.md` shout this; current wording is easy to misread.
- [x] **`inc` / `dec` are macros**, so they cannot be passed to `map`, `apply`, `swap!`, etc.
- [ ] **`recur` is only legal in the strict tail of `loop`**, not inside nested `let` / `if` (Clojure allows those when they are in tail position). Either extend tail analysis or document with examples.
- [x] **`doseq` is lazy `mapcat`**: side effects (including `go`) may never run. Clojure `doseq` is eager. Make FLAG `doseq` eager.
- [ ] **`peek` / `pop` are `first` / `rest`**, not Clojure vector stack ops (end of vector). Rename, split array vs vector, or document as intentional — today it looks like a bug.
- [ ] **`line-seq` / `file-to-strings` yield symbols**, not strings. Clojure yields strings.
- [ ] **Quoted lists** such as `'(1 2 3)` are not always usable as seqs (`cons`/`slow-nth` / `tenth` can fail). Clojure quoted lists are proper lists.
- [ ] **No `@` deref reader**; must write `(deref a)`. Easy to add if atoms stay.
- [ ] **`:strs` destructuring** looks up symbol keys, not string keys.
- [ ] **`print` is not implemented** (only `println` / `str`).
- [ ] ** Arrays supporting push to the end and having size in the box breaks immutability


## Intentionally different (do not “fix” to Clojure)
- `future` returns a callable 0-arg function (or a channel when `:piped? true`), not a `deref`-able IDeref.
- Atoms have no watches.
- Concurrency lives in `async.lib`, not `clojure.core` / `core.async`.
- `[1 2 3]` is an array; FLAG vectors are `| 1 2 3 |`.
- Modules use a header map (`:namespace` / `:exports` / `:imports`), not `(ns … :require …)`.