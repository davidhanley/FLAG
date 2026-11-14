# `Value` and `impl`

Part of **[The FLAG Book](flag-book.md)**.

Runtime `Value` is still a tagged immediate (`d`, `p`, `tag`) plus a small
per-type vtable:

```go
type impl interface {
    hash(v Value) string // map/set identity key
    typ(v Value) Value   // type-of keyword (`type` is reserved in Go)
    String(v Value) string
    assoc(v Value, entries ...Value) Value
    dissoc(v Value, keys ...Value) Value
}

type Value struct {
    d    float64
    p    unsafe.Pointer
    tag  ValueTag
    impl impl
}
```

`hash` / `typ` / `String` / `assoc` / `dissoc` take the `Value` so longs and
bools can still read `d`. Constructors go through `makeValue`, which always fills
`impl`. `TypeOf` and hashing call `v.impl` directly. Printing is `v.String()` →
`v.impl.String(v)` (`fmt.Stringer`). There is no `ValueToString`. `Assoc` is
`v.impl.assoc(v, entries...)`. Map, array, and vector override `assoc`. `Dissoc`
and `MapDissoc` are `v.impl.dissoc(v, keys...)`; map and nil override `dissoc`.
Other types inherit `defaultImpl`, which panics. A `Value` with a nil `impl` is
uninitialized and those calls panic.

More builtins (`get`, `first`, …) will move onto `impl` the same way. Do not put
every FLAG function on `impl` yet. Numerics stay on tags until math is optimized.
