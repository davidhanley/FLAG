# Sequence extras (`take-while`, `drop-while`, `take-last`, `drop-last`, `take-nth`, `group-by`)

Part of **[The FLAG Book](flag-book.md)** (chapter 2 companion).

These are prologue functions with Clojure’s collection arities (no transducers).
End-relative helpers walk with `drop` / `next`; `n` must be an integer (`pos?`
for “how many”). Compare with `vec` when you only care about values.

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
