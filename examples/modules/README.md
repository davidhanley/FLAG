# modules example

Small multi-file FLAG program that exercises the module system:

| File | Namespace | Role |
|------|-----------|------|
| `math.flag` | `math` | exports `add`, `double` |
| `greeter.flag` | `greeter` | imports `math`, exports `greet`, `shout` |
| `main.flag` | `main` | entry; uses bare import, `:as`, and `:refer` |

## Build and run

```bash
go run ./cmd/flag-lang build examples/modules/main.flag -o /tmp/flag-modules
/tmp/flag-modules
```

Expected output:

```
5
Hello, world x2 = 10
HI FLAG
```

Also works as a directory entry (uses `main.flag`):

```bash
go run ./cmd/flag-lang build examples/modules -o /tmp/flag-modules
```
