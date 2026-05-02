## 1. Manifest Schema

- [x] 1.1 Add `Vendor *bool` field to `Manifest` struct in `internal/manifest/schema.go` with yaml tag `vendor,omitempty`
- [x] 1.2 Add `IsVendored()` helper method on `Manifest` that returns `true` when `Vendor` is nil or `true`
- [x] 1.3 Add manifest round-trip test: `vendor: false` parses correctly and `vendor: true` / absent both report `IsVendored() == true`

## 2. gitignore Package

- [x] 2.1 Create `internal/gitignore/gitignore.go` with `EnsureEntries(path string, entries []string) (added []string, err error)` — reads file (or starts empty), appends missing entries under a melon comment block, writes back
- [x] 2.2 Add `RemoveEntries(path string, entries []string) error` — removes matching lines from the file
- [x] 2.3 Add `ContainsEntry(path string, entry string) (bool, error)` — reports whether an entry exists
- [x] 2.4 Write table-driven unit tests for all three functions covering: file does not exist, entry already present (idempotent), multiple entries, entry removal, comment block written on first write

## 3. melon init

- [x] 3.1 Add vendoring prompt to TTY init flow in `internal/cli/init_cmd.go` (after tool compat selection, default yes)
- [x] 3.2 Add vendoring prompt to non-TTY / `--yes` flow — `--yes` defaults to true (no `vendor` field written)
- [x] 3.3 Update `generateManifestYAML` to accept a `vendor bool` param and conditionally emit `vendor: false` line
- [x] 3.4 Update init model in `internal/cli/init_model.go` to include vendoring field in result struct
- [x] 3.5 Add/update init tests to cover `vendor: false` written when user opts out

## 4. Install gitignore Sync

- [x] 4.1 After the place step in `internal/cli/install_cmd.go`, check `manifest.IsVendored()`; if false, collect all managed symlink paths and call `gitignore.EnsureEntries`
- [x] 4.2 Also ensure `.melon/` is in the entries list
- [x] 4.3 Print the "Tip: run `git rm --cached`" hint for any entries that were newly added
- [x] 4.4 Write integration test: install with `vendor: false` creates/updates `.gitignore` with expected entries

## 5. Add Command gitignore Sync

- [x] 5.1 After the install pipeline in `internal/cli/add_cmd.go`, check `manifest.IsVendored()`; if false, call `gitignore.EnsureEntries` with the new dep's symlink path(s)
- [x] 5.2 Write test: `mln add` with `vendor: false` appends symlink path to `.gitignore`

## 6. Remove Command gitignore Cleanup

- [x] 6.1 After the install pipeline in `internal/cli/remove_cmd.go`, check `manifest.IsVendored()`; if false, call `gitignore.RemoveEntries` with the removed dep's symlink path(s)
- [x] 6.2 Write test: `mln remove` with `vendor: false` removes symlink path from `.gitignore`

## 7. README

- [x] 7.1 Update the "How it works" file table to show both vendor modes (committed vs gitignored) for `.melon/` and skill symlink dirs
- [x] 7.2 Update the Manifest Reference section to document the `vendor` field
- [x] 7.3 Update the CI paragraph that currently says `.melon/` and symlinks are always committed to reflect that this depends on the `vendor` setting

## 8. End-to-End Verification

- [x] 8.1 Run full test suite (`go test ./...`) and fix any failures
- [x] 8.2 Manual smoke test: `mln init` (opt out of vendoring) → `mln add` → verify `.gitignore` → `mln remove` → verify `.gitignore` cleaned up
