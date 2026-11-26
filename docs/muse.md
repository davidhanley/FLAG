# MUSE

Part of **[The FLAG Book](flag-book.md)**.

MUSE is FLAG’s build file. A `.muse` file is a FLAG map. Today it only lists
git libraries; later it will also name extra sources and build resources.

```clojure
{:libraries ["csv"]}
```

`flag-lang compile`, `build`, and `test` accept a `.muse` path the same way they
accept a directory. A FLAG file or directory next to a `.muse` file also
triggers a fetch.

## Libraries

Each `:libraries` entry is:

- a short name (`csv`) — resolved to `$MUSE_LIB_CSV`, then a sibling
  checkout, then [github.com/davidhanley/CSV](https://github.com/davidhanley/CSV)
- a git URL (`https://…`, `git@…`, `file://…`)
- a local git repository path

MUSE clones into `$MUSE_CACHE` (default `~/.cache/flag-muse`) and copies the
library’s `.lib` file into `<project>/libraries/` so `:imports ["csv.lib"]`
resolves without shipping the library inside flag-lang.

A library repo should contain `libraries/<name>.lib` (or `<name>.lib` at the
root).

## Example

```text
examples/FRS/frs.muse     {:libraries ["csv"]}
examples/FRS/frs.flag     {:imports ["csv.lib" "burp.lib"]}
```

```bash
go run ./cmd/flag-lang test examples/FRS
go run ./cmd/flag-lang test examples/FRS/frs.muse
```

`burp.lib` is still an in-tree language library. CSV is not.
