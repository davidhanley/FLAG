# File lines (`line-seq`, `file-to-strings`)

Part of **[The FLAG Book](flag-book.md)** (chapter 2 companion).

`line-seq` and `file-to-strings` are lazy sequences of **strings**, matching
Clojure `line-seq` (not symbols). Newlines are stripped.

```clojure
(first (line-seq (open-file "tests/language.flag")))
;; => "(ns language.core)"   ; :string

(first (file-to-strings "tests/language.flag"))
;; same
```

- `(line-seq file)` — open file Value (`open-file` / `io/reader`)
- `(file-to-strings file-or-path)` — file Value, or a path string/symbol

`io/readline` already returned strings; these helpers now do too.
