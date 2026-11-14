# FLAG vs Clojure lispbench

Part of **[The FLAG Book](flag-book.md)**.

FRS is a lispy scoring pipeline (lazy `range`, ratios, keyword maps, `group-by`,
`assoc`). The full FRS program is often ~4× Clojure (I/O dilutes CPU). These
kernels drop I/O so the runtime gap shows up. After lazy `range`, CPU totals
are ~2× Clojure; maps/`assoc`/`group-by` still lead. Huge **ratio** sums can
favor FLAG.

## Run

```bash
go run ./cmd/lispbench
```

Requires `clojure` or `clj` on `PATH`. The harness builds
`examples/lispbench/main.flag`, runs it, runs `examples/lispbench/lispbench.clj`,
and prints `flag_ms`, `clj_ms`, and `flag/clj`. Timing is **in-process**
(`time/nano-time` vs `System/nanoTime`); JVM startup is excluded.

Correctness of the work functions (small *n*):

```bash
go run ./cmd/flag-lang test examples/lispbench/work.flag
```

Do not `flag test` the whole directory: `main.flag` is the timed driver.

## `time/nano-time`

`(time/nano-time)` → long nanoseconds (wall clock). Use for elapsed time:

```clojure
(let [t0 (time/nano-time)
      _ (do-work)
      t1 (time/nano-time)]
  (- t1 t0))
```

No-arg; not a date. Prefer this over `t/now` for benchmarks.

## Why FRS is slow (what the rows highlight)

| Bench | FRS analog | Likely FLAG cost |
|-------|------------|------------------|
| `lazy-range-sum`, `lazy-map-inc` | `(def fractions (map #(/ 5 …) (range)))` | `(range)` is a lazy list with O(1) `first`/`rest`. Clojure still **chunks 32**. |
| `ratio-scores`, `frs-score-shape` | `get-scores-list` | Boxed `Value` + ratio objects vs Clojure ratios after JIT. |
| `eager-range-sum`, `apply-max` | `reduce +`, `(apply max …)` | Every long is a tagged `Value`; `+` type-switches. Clojure `LongRange` + primitive combine. |
| `keyword-gets`, `nested-keyword` | `(:name athlete)`, `(:race-name (:header r))` | Persistent map lookup through `Value` keys. |
| `assoc-scores` | `add-scores-and-rank` | `assoc` of several keys per athlete. |
| `group-by-name`, `juxt-keys` | `group-by :name`, `(juxt :name :age)` | HAMT `assoc` in a FLAG `reduce`. |
| `mapcat-filter`, `sort-by-points` | `compute-sex-results` | Extra seq copies; no chunking. |
| `string-pipeline` | CSV `trim` / `upper-case` | Usually closer to Clojure; not the FRS 4× by itself. |

If `keyword-gets` / `nested-keyword` / `group-by-*` lead, boxed maps (not
`range`) are the next place to optimize.
