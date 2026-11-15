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
    get(v Value, key Value, missing Value) Value
    contains(v Value, key Value) bool
    find(v Value, key Value) Value
    keys(v Value) Value
    vals(v Value) Value
    first(v Value) Value
    rest(v Value) Value
    next(v Value) Value
    last(v Value) Value
    conj(v Value, entries ...Value) Value
    count(v Value) int
    isEmpty(v Value) bool
    notEmpty(v Value) Value
    seq(v Value) Value
    nth(v Value, i int, missing Value, hasMissing bool) Value
    slowNth(v Value, i int, missing Value, hasMissing bool) Value
    drop(v Value, n int) Value
    cons(v Value, item Value) Value
    seqQ(v Value) bool
    seqCursor(v Value) seqCursor
    eq(v Value, other Value) bool
    isTruthy(v Value) bool
    isNil(v Value) bool
    compare(v Value, other Value) Value
    toAny(v Value) any
    toNative(v Value) any
    asString(v Value) string
    asIdent(v Value) (string, bool)
    toJSON(v Value) (any, error)
    name(v Value) string
    exMessage(v Value) Value
    goSeq(v Value) []Value
    isNumeric(v Value) bool
    add(v, other Value) Value
    sub(v, other Value) Value
    mul(v, other Value) Value
    div(v, other Value) Value
    quot(v, other Value) Value
    cmpNum(v, other Value) int
    abs(v Value) Value
    toDouble(v Value) Value
    numerator(v Value) Value
    denominator(v Value) Value
    toFloat64(v Value) float64
    toRat(v Value) *big.Rat
    toBigInt(v Value) *big.Int
    bitLong(v Value, name string) int64
    addLong(v Value, n int64) Value
    addDouble(v Value, n float64) Value
    addRatio(v Value, n *big.Rat) Value
    addBigInt(v Value, n *big.Int) Value
    // sub/mul/div/quot/cmp + Long/Double/Ratio/BigInt similarly
}

type Value struct {
    d    float64
    p    unsafe.Pointer
    tag  ValueTag
    impl impl
}
```

`hash` / `typ` / `String` / `assoc` / `dissoc` / `get` / `contains` take the
`Value` so longs and bools can still read `d`. Constructors go through
`makeValue`, which always fills `impl`. `TypeOf` and hashing call `v.impl`
directly. Printing is `v.String()` → `v.impl.String(v)` (`fmt.Stringer`). There
is no `ValueToString`. `Assoc` is `v.impl.assoc(v, entries...)`. Map, array, and
vector override `assoc`. `Dissoc` and `MapDissoc` are `v.impl.dissoc(v, keys...)`;
map and nil override `dissoc`. `Get` is `v.impl.get(v, key, missing)` after arity
checks; map, set, array, vector, string, date, record, and nil override `get`.
`Contains` is `v.impl.contains(v, key)`; map, set, date, record, and nil override
`contains` (`nil` is `false`). `Find` is `v.impl.find(v, key)`; map, date, record,
and nil override `find` (`nil` is `nil`). `Keys` / `Vals` are `v.impl.keys(v)` and
`v.impl.vals(v)`; map, date, and nil override them (`nil` is `nil`). Other types
inherit `defaultImpl`, which panics. `First` is `v.impl.first(v)`; nil, list, array,
vector, lazy-list, set, and string override it. `Rest` is `v.impl.rest(v)` on the
same types except nil (which panics; `SeqRest` wraps nil as `()`). `Next` is
`v.impl.next(v)` (nil is `nil`; empty rest is `nil`). `Last` is `v.impl.last(v)`
on list, array, vector, and lazy-list. `Conj` is `v.impl.conj(v, entries...)` after
the arity check; list, array, vector, set, and map override it. `Count` /
`IsEmpty` / `NotEmpty` / `Seq` dispatch the same way. Nil counts as 0 and is
empty; string `seq` is `vec` of runes. `Nth` / `SlowNth` / `Drop` / `Cons` /
`seq?` dispatch the same way (`seq?` is true for list, array, vector, lazy-list).
`newSeqCursor` is `v.impl.seqCursor(v)`; cursor kinds are not `ValueTag`. `Eq` /
`IsTruthy` / `IsNil` / `Compare` dispatch through `impl`. Numeric `=` uses
`isNumeric` then `cmpNum`. FFI conversion is the same: `ValueToAny` is
`toAny`, Go/native unbox is `toNative`, JSON is `toJSON`, `name` is `name`,
`ex-message` is `exMessage`, and string/symbol names use `asIdent`/`asString`.
Array `toAny` is `[]any`; file `toNative` is the host `*os.File`.

Binary math is double-dispatch, not a tag switch. `(+ a b)` is `flagrt.Add(a, b)`
→ `a.impl.add(a, b)`. If `a` is a long it unboxes and calls
`b.impl.addLong(b, a.Long())` (same pattern for double/ratio/bigint and for
`-` `*` `/` `quot` `cmpNum`). Variadic `(+ 1 2 3)` and `(apply + xs)` fold that
binary `Add`; the two-arg path is not itself variadic. `(< a b c d)` is chained
binary compares, like Clojure. A `Value` with a nil `impl` is uninitialized and
those calls panic.

`value_impl.go` holds `defaultImpl` (panic stubs for optional ops) and the
`implByTag` table. Required methods (`hash`, `typ`, `String`) live on each
type’s file (`string.go`, `long.go`, `double.go`, …). Type-specific overrides
such as `assoc`/`dissoc` stay in that same file.

Do not put every FLAG function on `impl` yet. When types are known at compile
time, `(+ a b)` can skip the first hop (`b.addLong(a.longValue)`) or become a
native add.
