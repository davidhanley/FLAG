# `range`

Part of **[The FLAG Book](flag-book.md)** (chapter 2 companion).

`range` is a runtime builtin. Non-empty results are a **range seq** (`:range`):
the current `int64` plus exclusive `end` and `step`. Stepping is native integer
add/compare in Go (no lazy-list nodes or mutex). The end is exclusive. Empty
ranges are `[]`.

| Form | Result |
|------|--------|
| `(range)` | lazy `0, 1, 2, …` |
| `(range end)` | `0 .. end-1` |
| `(range start end)` | `start .. end-1` |
| `(range start end step)` | `start`, `start+step`, … while the value is still on the way to `end` |

```clojure
(range 4)           ;; (0 1 2 3)  — not an infinite seq from 4
(range 0)           ;; ()
(range -2)          ;; ()
(range 1 4)         ;; (1 2 3)
(range 5 1)         ;; ()
(range 0 8 2)       ;; (0 2 4 6)
(range 5 1 -1)      ;; (5 4 3 2)
(take 2 (range 7 9 0))  ;; (7 7)
```

Zero `step` with `start = end` is empty; otherwise it repeats `start` (lazy
list). Finite ranges are counted (`count` / `drop` / `slow-nth` / `last` are
O(1)). `(range)` is infinite from 0. `reduce` / `vec` / `take` walk with a
cursor that only bumps the current long.

`(map f (range n))` is eager (an array). `(map f (range))` stays a lazy list so
infinite maps do not hang until forced. Two equal ranges compare equal
(`(= (range 3) (range 3))`). The compiler emits `flagrt.Range(...)`;
`(apply range xs)` uses `flagrt.FnRange`.
