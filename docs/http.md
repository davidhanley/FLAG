# HTTP for FLAG

Part of **[The FLAG Book](flag-book.md)**.

FLAG’s HTTP layer is a normal imported library built on top of Go’s `net/http`.
That is enough for both client requests and small local servers; incoming
requests already run on the per-request goroutines provided by `net/http`.

## Import

```clojure
{:namespace "app"
 :imports   ["http.lib"]}
```

## Surface

### Client

| Function | Purpose |
|----------|---------|
| `http/request` | Build a request map |
| `http/header` | Add a header to a request map |
| `http/body` | Set request body bytes/text |
| `http/send` | Send a request map |
| `http/get` | Convenience GET |
| `http/post` | Convenience POST |

### Server

| Function | Purpose |
|----------|---------|
| `http/listen` | Start a local HTTP server and return a server handle |
| `http/stop` | Stop a server |
| `http/address` | Read the bound address from a server handle |
| `http/route` | Build a route entry |
| `http/router` | Build a router map from route entries |
| `http/get-route` etc. | Method-specific route helpers |

## Client request / response maps

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

## Starting a server

The main entry point is `http/listen`:

```clojure
(def server
  (http/listen
    {:port 8080
     :router
     (http/router
       (http/get-route "/" (fn [_] "hello"))
       (http/get-route "/users/:id"
         (fn [req]
           {:status 200
            :headers {:content-type "text/plain; charset=utf-8"}
            :body (:id (:params req))}))}))
```

Options:

| Key | Meaning |
|-----|---------|
| `:addr` | Exact bind address, e.g. `"127.0.0.1:8080"` |
| `:host` | Host to bind when `:addr` is omitted (defaults to `127.0.0.1`) |
| `:port` | Port to bind when `:addr` is omitted (defaults to `0`) |
| `:handler` | A single FLAG handler function |
| `:router` | A router created with `http/router` |
| `:routes` | A raw collection of route maps |

If neither `:addr` nor `:port` is fixed, the server binds to an ephemeral local
port on `127.0.0.1`.

Read the actual address with:

```clojure
(http/address server)
```

Stop it with:

```clojure
(http/stop server)
```

## Routing

Routes are just maps with `:method`, `:path`, and `:handler`. The helper
functions make that nicer:

```clojure
(http/router
  (http/get-route "/" home)
  (http/post-route "/submit" submit)
  (http/any-route "/health" health))
```

Supported path matching:

- exact paths, such as `"/about"`
- named segments, such as `"/users/:id"`
- trailing wildcards, such as `"/static/*path"`

When a path matches but the method does not, the router returns `405 Method Not
Allowed`. Unknown paths return `404 Not Found`.

## Incoming request map

Each handler receives one request map with keys such as:

```clojure
{:method "GET"
 :url "http://127.0.0.1:8080/users/42?lang=flag"
 :path "/users/42"
 :raw-path "/users/42"
 :query-string "lang=flag"
 :query {:lang "flag"}
 :headers {:accept "*/*"}
 :body ""
 :host "127.0.0.1:8080"
 :remote-addr "127.0.0.1:54012"
 :scheme "http"
 :params {:id "42"}}
```

Header values and query values become arrays when the same key appears more than
once.

## Handler return values

Handlers may return either:

1. a response map
2. a plain value, which becomes a `200` response body via string conversion

Response maps support:

```clojure
{:status 202
 :headers {:content-type "text/plain; charset=utf-8"
           :x-test "yes"}
 :body "accepted"}
```

If `:status` is omitted it defaults to `200`. If a body is present and no
`content-type` header is set, the server defaults to plain UTF-8 text.
