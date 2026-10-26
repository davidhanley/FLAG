# `range`

Part of **[The FLAG Book](flag-book.md)** (chapter 2 companion).

`range` is a runtime builtin. Non-empty results are **lazy lists** (`:lazy-list`);
the end is exclusive. Empty ranges are `[]`.

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

Zero `step` with `start = end` is empty; otherwise it repeats `start` (lazy).
Finite ranges stay lazy until `vec` / `doall` / `reduce` / `take`. `(range)` is
an infinite lazy list from 0. `first`/`rest` on lazy lists are O(1) per step.
`(count (range n))` is O(1) via a remaining-length hint. The compiler emits
`flagrt.Range(...)`; `(apply range xs)` uses `flagrt.FnRange`.
