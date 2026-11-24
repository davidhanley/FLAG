# Testing with Smoke

Smoke is the FLAG testing system.

Part of **[The FLAG Book](flag-book.md)**.

## Running tests

Use the `test` command:

```bash
go run ./cmd/flag-lang test path/to/src
go run ./cmd/flag-lang test examples/FRS
go run ./cmd/flag-lang test examples/lispbench/work.flag
go test ./...
```

GitHub Actions (`.github/workflows/ci.yml`) runs `go build ./...`, `go test ./...`, and
`go run ./cmd/flag-lang test tests` on pull requests and on `main`/`master`. That
check is free on a public repo (and is a few minutes of the included quota on a
private one). Lispbench is not in CI (it needs Clojure).

To **block merge** until it passes: repo **Settings → Rules → Rulesets** (or
**Branches → branch protection**) on the default branch. Require status check
`CI / test`. The workflow must have run once on that branch before the check
name appears.

FLAG vs Clojure timings: `go run ./cmd/lispbench` ([lispbench.md](lispbench.md)).

`examples/FRS` is a compiler stress test: modular `:imports`, ratios, CSV/JSON, and prologue (`select-keys`, `apply max`). Generated inspect Go is package-mode `.flag-build/` (prologue linked, not inlined). `go test ./...` skips that dotted directory.

You can point it at either:

- a source file such as `math.flag`
- a matching test file such as `math_test.flag`
- a directory containing `.flag` source and test files

## Test files

Smoke looks for test files whose basename ends in `_test`:

- `math.flag` pairs with `math_test.flag`
- directory runs collect normal `.flag` source files plus every `*_test.flag` file

For single-file and legacy directory runs, Smoke merges the program source with the test source before compiling.

For modular programs that use a header map with `:imports`, Smoke appends test forms to the entry module so tests can see private definitions from that module. In that mode, test modules must not declare their own `:imports`.

## Test forms

Smoke uses three core forms:

```clojure
(deftest add1-test
  (testing "increments once"
    (is (= (add1 2) 3)))
  (testing "optional custom message"
    (is (= (add1 0) 1) "add1 should increment zero")))
```

- `deftest` defines a test
- `testing` groups related assertions under a label
- `is` checks an expression and optionally accepts a custom failure message

When an `is` assertion fails, Smoke reports the source location of the assertion.

## How Smoke runs

Smoke compiles a temporary test harness, builds it, and runs the resulting binary. The harness is temporary. FLAG writes the program (no tests) to `.flag-build/` next to the source for inspection: one Go package per module, prologue linked rather than copied.
