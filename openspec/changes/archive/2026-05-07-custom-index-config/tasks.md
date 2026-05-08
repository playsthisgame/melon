# Tasks

## 1. Manifest Schema

- [x] 1.1 Add `IndexConfig` struct to `internal/manifest/schema.go` with `URL string` and `Exclusive bool` fields
- [x] 1.2 Add `Index *IndexConfig` field (pointer, omitempty) to `Manifest` struct
- [x] 1.3 Add manifest schema test covering round-trip parse/marshal of the `index` block

## 2. Index Package

- [x] 2.1 Change `index.Fetch()` signature to `Fetch(url string) ([]Entry, error)` and remove the hardcoded `IndexURL` constant usage inside it
- [x] 2.2 Export `DefaultIndexURL` constant (rename from `IndexURL`) so callers can reference the public default
- [x] 2.3 Update existing `index` package tests to pass a URL to `Fetch()`

## 3. CLI — Search Command

- [x] 3.1 In `search_cmd.go`, load `melon.yaml` from the working directory (ignore error if absent)
- [x] 3.2 Resolve active index URLs: if manifest has `index.url`, add it; if `exclusive` is false or absent, also add `DefaultIndexURL`
- [x] 3.3 Fetch and merge results: custom index entries first, public index entries appended with same-name duplicates suppressed
- [x] 3.4 Update search tests to cover: no manifest (public only), custom + public (merged, deduped), exclusive (custom only)

## 4. CLI — Info Command

- [x] 4.1 Apply the same index resolution logic from task 3.1–3.2 to `info_cmd.go`
- [x] 4.2 Update info tests to cover the custom index path

## 5. Documentation

- [x] 5.1 Add the `index` block to the `melon.yaml` schema example in `README.md` (near the `vendor` field, around line 139)
- [x] 5.2 Update the `melon search` command description in `README.md` to mention that search respects the `index` block for custom/private registries

## 6. Validation

- [x] 5.1 Run `go test ./...` and confirm all tests pass
- [ ] 5.2 Manual smoke test: add an `index` block to a local `melon.yaml` and verify `melon search` returns results from the custom URL
