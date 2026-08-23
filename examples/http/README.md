# http example

Small end-to-end demo of the FLAG HTTP library.

It starts a local server, routes requests by path and method, then calls back
into that server with the client-side `http/get`, `http/post`, and `http/send`
functions.

Full API reference: **[docs/http.md](../../docs/http.md)**.

## Run the demo

```bash
go run ./cmd/flag-lang build examples/http/main.flag -o /tmp/flag-http
/tmp/flag-http
```

Or as a directory (uses `main.flag`):

```bash
go run ./cmd/flag-lang build examples/http -o /tmp/flag-http
/tmp/flag-http
```

## Tests

FLAG unit tests live in `main_test.flag`. Run them with:

```bash
go run ./cmd/flag-lang test examples/http
```

## Expected demo output

```text
server: 127.0.0.1:NNNNN
GET / -> 200 home
GET /hello/:name -> 200 hello flag from FLAG
POST /echo -> 201 echo payload method=POST
GET /static/*path -> 200 static css/site.css
DELETE /echo -> 405 method not allowed
```
