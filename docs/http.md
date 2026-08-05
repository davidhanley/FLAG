# HTTP for FLAG

Part of **[The FLAG Book](flag-book.md)**.

FLAG’s HTTP layer is a thin wrapper over Go’s `net/http`. It is designed as a
normal imported library, not a compiler feature.

## Import

```clojure
{:namespace "app"
 :imports   ["http.lib"]}
```

## Surface

| Function | Purpose |
|----------|---------|
| `http/request` | Build a request map |
| `http/header` | Add a header to a request map |
| `http/body` | Set request body bytes/text |
| `http/send` | Send a request map |
| `http/get` | Convenience GET |
| `http/post` | Convenience POST |

## Request map keys

Use kebab-case keywords:

```clojure
{:method "POST"
 :url "https://example.com"
 :headers {:x-test "yes"}
 :body "hello"
 :timeout-ms 1000
 :follow-redirects true}
```

Headers may be strings, symbols, or collections of values. Response values come
back as maps with `:status`, `:status-text`, `:url`, `:method`, `:ok`,
`:headers`, and `:body`.
