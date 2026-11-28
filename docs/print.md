# `print`

Part of **[The FLAG Book](flag-book.md)** (chapter 2 companion).

`(print …)` is Clojure `print`: write arguments to stdout for humans, with
no trailing newline, and return `nil`.

```clojure
(print 1 2 3)          ;; writes 1 2 3
(print "hi" nil :k)    ;; writes hi nil :k
(print)                ;; writes nothing
(print "" 1)           ;; writes  1  (leading space)
```

## Behavior

- Variadic, including zero arguments.
- A single space between arguments (even when an argument prints as empty).
- No trailing newline (`println` still adds one).
- Always returns `nil`.
- Strings print unquoted (`(print "hi")` writes `hi`, not `"hi"`).
- `nil` prints as `nil` (unlike `(str nil)`, which is `""`).
- Booleans print as `true` / `false`; keywords as `:name`.
- Usable as a value: `(apply print [1 2 3])`, `(let [p print] (p 7))`.

`pr` / `prn` / `print-str` / `*out*` are not implemented. Nested collections
use FLAG `String()` for their contents, so strings inside a list still show
Go-quoted (`(print '("a"))` writes `("a")`, not Clojure’s `(a)`).

## Compiler / runtime

Expression and top-level `(print …)` lower to `flagrt.Print(...)`. The
function value is `flagrt.FnPrint`.
