# Math (`math/…`)

Part of **[The FLAG Book](flag-book.md)** (chapter 2 companion).

FLAG’s `math/` package is the Clojure `clojure.math` / `java.lang.Math` surface.
Functions are ambient (no module import). Arguments may be any FLAG number
(`long`, `double`, `ratio`, `bigint`); non-numbers throw.

## `math/abs`

Preserves the input numeric type (unlike the rest of this package):

```clojure
(math/abs -3)     ;; 3
(math/abs -2.5)   ;; 2.5
(math/abs -3/2)   ;; 3/2
```

`abs` of `Long/MIN_VALUE` promotes to bigint.

## Double-valued functions

These coerce to `double` and return a float (`:float`):

| Form | Meaning | Example |
|------|---------|---------|
| `(math/sqrt x)` | nonnegative square root | `(math/sqrt 4)` → `2.0` |
| `(math/pow a b)` | `a` raised to `b` | `(math/pow 2 3)` → `8.0` |
| `(math/exp x)` | `e^x` | `(math/exp 0)` → `1.0` |
| `(math/log x)` | natural log | `(math/log 1)` → `0.0` |
| `(math/log10 x)` | log base 10 | `(math/log10 100)` → `2.0` |
| `(math/sin x)` | sine (radians) | `(math/sin 0)` → `0.0` |
| `(math/cos x)` | cosine (radians) | `(math/cos 0)` → `1.0` |
| `(math/tan x)` | tangent (radians) | `(math/tan 0)` → `0.0` |
| `(math/floor x)` | greatest integer ≤ `x` as double | `(math/floor -2.3)` → `-3.0` |
| `(math/ceil x)` | least integer ≥ `x` as double | `(math/ceil -2.3)` → `-2.0` |
| `(math/IEEE-remainder x y)` | IEEE 754 remainder | `(math/IEEE-remainder 5 3)` → `-1.0` |

IEEE special values follow Go `math` / Java `Math` (NaN, ±Inf). Domain errors
such as `(math/sqrt -1)` yield NaN rather than throwing.

`IEEE-remainder` is **not** `rem` or `mod`. It is `f1 − f2 × n` where `n` is
the integer closest to `f1/f2` (even `n` on a tie).

## `math/round`

Returns a **long** (`:int`), matching `clojure.math/round` / `Math.round`:

- nearest integer
- **ties round toward +∞** (`(math/round 2.5)` → `3`, `(math/round -1.5)` → `-1`)
- NaN → `0`
- ±Inf saturate at `Long/MAX_VALUE` / `Long/MIN_VALUE`

```clojure
(math/round 1.4)    ;; 1
(math/round 1.5)    ;; 2
(math/round 2.5)    ;; 3
(math/round -1.5)   ;; -1
```

This differs from Go’s `math.Round` (ties away from zero).

## Core bitwise ops (`bit-*`)

These are `clojure.core` builtins (not `math/…`). They operate on 64-bit longs
like Java/`clojure.core`: integers and in-range bigints only; shift counts are
`n & 63`. See [flag-lang.md](../flag-lang.md) for the full table.

| Form | Meaning |
|------|---------|
| `(bit-and x y & more)` | bitwise and |
| `(bit-or x y & more)` | bitwise or |
| `(bit-xor x y & more)` | bitwise xor |
| `(bit-not x)` | bitwise complement |
| `(bit-shift-left x n)` | `x << n` |
| `(bit-shift-right x n)` | arithmetic `x >> n` |
| `(unsigned-bit-shift-right x n)` | logical `x >>> n` |
| `(bit-test x n)` | true if bit `n` is set |
| `(bit-set x n)` | set bit `n` |
| `(bit-clear x n)` | clear bit `n` |
| `(bit-flip x n)` | toggle bit `n` |
