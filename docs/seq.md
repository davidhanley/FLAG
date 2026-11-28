# Sequence extras (`take-while`, `drop-while`, `split-at`, `split-with`, `cycle`, `reductions`, `frequencies`, `butlast`, `take-last`, `drop-last`, `take-nth`, `group-by`, `reduced`, `peek`, `pop`)

Part of **[The FLAG Book](flag-book.md)** (chapter 2 companion).

Most of these are prologue functions with Clojure’s collection arities (no
transducers). `reduced` / `reduced?` / `ensure-reduced` / `unreduced` are
runtime builtins. End-relative helpers walk with `drop` / `next`; `n` must be
an integer (`pos?` for “how many”). Compare with `vec` when you only care
about values.

## `take-while`

`(take-while pred coll)` keeps a prefix while `(pred x)` is truthy, then stops.

```clojure
(take-while pos? [1 2 -1 3])   ;; [1 2]
(take-while pos? [-1 2])       ;; []
(take-while pos? nil)          ;; []
```

Stops on a lazy input once `pred` fails, so `(take-while (fn [x] (< x 3)) (range))`
is finite.

## `drop-while`

`(drop-while pred coll)` skips a prefix while `(pred x)` is truthy and returns
the rest (same collection kind when possible). `nil` stays `nil`.

```clojure
(drop-while pos? [1 2 -1 3])   ;; [-1 3]
(drop-while pos? [1 2])        ;; []
(drop-while pos? nil)          ;; nil
```

## `split-at`

`(split-at n coll)` is `[(take n coll) (drop n coll)]`. `n` must be a
non-negative integer (same as `take` / `drop`).

```clojure
(split-at 2 [1 2 3 4 5])   ;; [[1 2] [3 4 5]]
(split-at 0 [1 2 3])       ;; [[] [1 2 3]]
(split-at 10 [1 2 3])      ;; [[1 2 3] []]
(split-at 0 nil)           ;; [[] nil]
```

`take` always yields an array. `drop` keeps the remaining collection kind.
`coll` must be a list, array, vector, lazy-list, or range — the same kinds
`drop` accepts (not strings). `n > 0` with `nil` panics, like `drop`.

## `split-with`

`(split-with pred coll)` is `[(take-while pred coll) (drop-while pred coll)]`.

```clojure
(split-with pos? [1 2 -1 3])   ;; [[1 2] [-1 3]]
(split-with pos? [-1 2])       ;; [[] [-1 2]]
(split-with pos? nil)          ;; [[] nil]
```

Stops on a lazy input once `pred` fails, so
`(split-with (fn [x] (< x 3)) (range))` is finite on the left.

## `cycle`

`(cycle coll)` repeats `coll` forever. Empty/`nil` is `[]`. Consume with
`take` — the result is infinite when `coll` is non-empty.

```clojure
(take 5 (cycle [1 2 3]))   ;; [1 2 3 1 2]
(cycle [])                 ;; []
(cycle nil)                ;; []
```

## `reductions`

`(reductions f coll)` is the intermediate values of `(reduce f coll)`.
`(reductions f init coll)` starts from `init`. Empty `coll` without `init`
is `[(f)]` (Clojure: `(reductions + [])` is `(0)`).

```clojure
(reductions + [1 2 3 4])       ;; [1 3 6 10]
(reductions + 0 [1 2 3 4])     ;; [0 1 3 6 10]
(reductions + [])              ;; [0]
(take 4 (reductions + (range))) ;; [0 1 3 6]
```

## `frequencies`

`(frequencies coll)` maps each distinct item to its count. Empty/`nil` is `{}`.

```clojure
(frequencies [:a :b :a])   ;; {:a 2 :b 1}
(frequencies nil)          ;; {}
```

## `butlast`

`(butlast coll)` is all but the last item, or `nil` when `coll` has fewer
than two items (unlike `drop-last`, which yields `[]`).

```clojure
(butlast [1 2 3 4])   ;; [1 2 3]
(butlast [1])         ;; nil
(butlast [])          ;; nil
(butlast nil)         ;; nil
```

## `peek` / `pop`

Clojure stack ops. Lists grow at the front; FLAG arrays and vectors grow at the
end.

- `(peek coll)` — list: `first`; array/vector: `last`. Empty or `nil` → `nil`.
- `(pop coll)` — list without first, or array/vector without last. Same kind.
  Empty or `nil` panics. Not `rest` / `butlast`.

```clojure
(peek '(1 2 3))     ;; 1
(pop '(1 2 3))      ;; (2 3)
(peek [1 2 3])      ;; 3
(pop [1 2 3])       ;; [1 2]
(peek | 1 2 3 |)    ;; 3
(pop | 1 2 3 |)     ;; | 1 2 |
(peek [])           ;; nil
(pop [])            ;; panic
```

Maps, sets, strings, ranges, and lazy seqs are not stacks.

## `reduced` / `reduced?` / `ensure-reduced` / `unreduced`

`(reduced x)` boxes `x`. When `f` returns a reduced value, `reduce` unwraps it
and stops. Empty coll with a reduced init stays reduced. `(reduced? x)` is
true only for that box. `(ensure-reduced x)` boxes unless already reduced.
`(unreduced x)` unwraps once, or returns `x`. Equality is identity.
`(type-of (reduced 1))` is `:reduced`.

```clojure
(reduce (fn [acc x]
          (if (>= (+ acc x) 6)
            (reduced (+ acc x))
            (+ acc x)))
        0
        (range))          ;; 6
(reduced? (reduced 1))    ;; true
(unreduced (reduced 1))   ;; 1
(reduce + (reduced 1) []) ;; still reduced
```

`reduce-kv` honors the same stop: return `(reduced acc)` from `f`.

FLAG `deref` is atoms only — use `unreduced`, not `@`.

## `ffirst` / `nfirst` / `nnext` / `fnnext`

Clojure’s `first`/`next` shortcuts. `next` (not `rest`) means a missing tail is
`nil`, not `[]`.

```clojure
(ffirst [[1 2] [3]])    ;; 1          ; (first (first x))
(nfirst [[1 2 3] [4]])  ;; [2 3]      ; (next (first x))
(nnext [1 2 3 4])       ;; [3 4]      ; (next (next x))
(fnnext [1 2 3 4])      ;; 3          ; (first (nnext x))
(nnext [1 2])           ;; nil
(fnnext nil)            ;; nil
```

## `take-last`

`(take-last n coll)` is the last `n` items, or the whole seq if shorter. Empty
or `nil` coll, and `n <= 0`, yield `nil` (Clojure).

```clojure
(take-last 2 [1 2 3 4])   ;; (3 4)
(take-last 10 [1 2 3])    ;; (1 2 3)
(take-last 2 [])          ;; nil
(take-last 0 [1 2])       ;; nil
```

## `drop-last`

`(drop-last coll)` drops one item; `(drop-last n coll)` drops `n`. `n <= 0`
returns `coll`. `nil` is empty.

```clojure
(drop-last [1 2 3 4])     ;; [1 2 3]
(drop-last 2 [1 2 3 4])   ;; [1 2]
(drop-last 10 [1 2])      ;; []
(drop-last 0 [1 2 3])     ;; [1 2 3]
```

## `take-nth`

`(take-nth n coll)` is every `n`th item, starting with the first. If `n <= 0`
and `coll` is non-empty, Clojure repeats the first item forever (`repeat`).

## `group-by`

`(group-by f coll)` is a map from `(f x)` to an array of the `x`s in encounter
order. Empty/`nil` coll is `{}`. Implemented in Go (`flagrt.GroupBy`): it walks
a seq cursor, grows per-key slices, then freezes arrays. A keyword `f` uses
`KeywordGet` (no `Call`).

```clojure
(group-by identity [1 2 2])   ;; {1 [1] 2 [2 2]}
(group-by :k [{:k :a} {:k :a}])  ;; {:a [{:k :a} {:k :a}]}
(group-by even? nil)          ;; {}
```

## `juxt`

`(juxt f g ...)` returns a function that applies each `f` to the same arguments
and returns an array of the results. Keyword fns with one argument use
`KeywordGet`.

```clojure
((juxt first second) [1 2 3])           ;; [1 2]
((juxt :name :age) {:name "ADA" :age 40})  ;; ["ADA" 40]
((juxt +) 1 2 3)                        ;; [6]
```

```clojure
(take-nth 2 [1 2 3 4 5])           ;; [1 3 5]
(take-nth 10 [1 2 3])              ;; [1]
(take 3 (take-nth 0 [7 8 9]))      ;; [7 7 7]
```
