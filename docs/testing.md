# Testing with Smoke

Smoke is the FLAG testing system.

Part of **[The FLAG Book](flag-book.md)**.

## Running tests

Use the `test` command:

```bash
go run ./cmd/flag-lang test path/to/src
```

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

Smoke compiles a temporary test harness, builds it, and runs the resulting binary. The test harness itself is temporary, but FLAG still writes the generated program Go file next to the source for inspection.
