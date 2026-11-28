# `loop` / `fn` / `recur`

Part of **[The FLAG Book](flag-book.md)** (chapter 2 companion).

FLAG `recur` matches Clojure’s tail-call rebinding: it is a `continue` of the
innermost `loop` or function, not a boxed value and not a stack-growing call.

## Where `recur` is legal

`recur` must be in **tail position** of a `loop`, `fn`, `defn`, or `#()` body.

Tail position includes:

- the body of `loop` / `fn` / `defn` / `#()`
- either branch of `if` (not the test)
- the last form of `do` / `let`
- the last form of macros that expand to those (`when`, `cond`, `or`, `and`, …)

Not tail (compile error: `can only recur from tail position`):

- `if` test, `let`/`loop` binding inits, non-last `do` forms
- arguments of a call, or elements of a literal
- `(when … (recur …))` followed by another body form

`recur` cannot jump out of `try` / `catch` / `finally` (`cannot recur across try`).
A `loop` or `fn` *inside* `try` may still `recur` to itself.

Outside any `loop`/`fn`: `recur can only be used within loop or fn`.
Wrong arity: `recur expects N arguments`. Empty bindings `(loop [] …)` are
allowed (`N` is 0).

## Innermost target

`recur` always targets the innermost `loop` or function. An inner `fn` does not
`recur` to an outer `loop`. Nested `loop` forms work the same way.

Variadic functions take fixed parameters plus one rest argument. The last
`recur` value becomes the new rest collection. `(seq nil)` is `nil`, so the
usual Clojure pattern `(if (seq xs) (recur (inc n) (next xs)) n)` terminates.

Destructuring parameters re-run each iteration from the undestructured
parameter (`__argN` or the rest box). Recur with the undestructured arity,
not the destructured names.

## Examples

```clojure
(loop [n 0]
  (if (= n 3)
    n
    (recur (inc n))))          ;; 3

(defn countdown [n]
  (if (= n 0) n (recur (dec n))))

((fn [n acc]
   (if (= n 0) acc (recur (dec n) (+ acc n))))
 3 0)                          ;; 6

(loop [n 0]
  (cond
    (< n 3) (recur (inc n))
    :else n))

(loop [i 0]
  (if (= i 2)
    i
    (do
      (loop [j 0]
        (if (= j 2) j (recur (inc j))))
      (recur (inc i)))))
```
