# HAMT maps and sets

Part of **[The FLAG Book](flag-book.md)**.

Large FLAG maps and sets are a 32-way **bitmap HAMT** in `runtime/hamt`.
This is FLAG-owned. It does not use `github.com/benbjohnson/immutable`.

Maps of at most 8 unique keys stay array-maps (`docs/value.md`). The 9th unique
key promotes to this HAMT. `type-of` is still `:map`. Sets are always HAMTs
(`type-of` `:set`).

## Shape

Each node is a packed bitmap of up to 32 slots (5 hash bits per level). A slot
is either a key/value leaf or a child node. Keys with the same 64-bit `Hash`
share a collision node. `Get` / `assoc` / `dissoc` copy O(log₃₂ n) nodes.

The type is generic: `hamt.Map[K, V]` and `hamt.Set[K]` (`Set` is
`Map[K, struct{}]`). FLAG passes `Hash` and `valuesHamtEq` (identity, then `Eq`)
so numeric keys like `1` and `1.0` still match.

## Persistent API

`NewMap(hash, eq)`, `Assoc`, `Dissoc`, `Get`, `Len`, `Range`.
`NewSet(hash, eq)`, `Add`, `Delete`, `Has`, `Len`, `Range`.

Assoc/dissoc return a new map/set. The previous value is unchanged.

## Transients

`Map.Transient()` / `Set.Transient()` start a session. The first time that
session touches a node it copies the node, stamps it dirty, and later updates
mutate in place. `Persistent()` sets the session to dead so those nodes freeze
without walking the tree. Using the transient after that panics
(`transient used after persistent!`).

FLAG `(transient {})` / `(transient #{})` wrap these sessions on a
`Value` (`:transient-map` / `:transient-set`). See [transients.md](transients.md).
`NewMap` of many pairs, HAMT `assoc` of several keys, `NewSet`, `union` /
`intersection` / `difference`, `select`, `project`, `rename`, and persistent
`conj` onto HAMT maps/sets also use transients internally then freeze.
