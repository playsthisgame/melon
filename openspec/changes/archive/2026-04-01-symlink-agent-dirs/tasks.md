## 1. Update placer implementation

- [x] 1.1 Rewrite `Place` in `internal/placer/placer.go` to use `os.Symlink` instead of `copyDir` — compute a relative symlink target from the skill slot directory to `store.InstalledPath`
- [x] 1.2 Remove the existing entry at the skill slot before symlinking (use `os.RemoveAll` to handle stale copies or broken symlinks)
- [x] 1.3 Ensure the parent directory (`<agent-dir>/skills/`) is created with `os.MkdirAll` before calling `os.Symlink`
- [x] 1.4 Update the progress message printed to `out` to say "linked" instead of "placed"
- [x] 1.5 Delete the `copyDir` and `copyFile` helper functions (dead code after the rewrite)
- [x] 1.6 Update the package doc comment to describe symlinking rather than copying

## 2. Tests

- [x] 2.1 Add/update test in `internal/placer/` that verifies the skill slot is a symlink after `Place` runs
- [x] 2.2 Add test that verifies the symlink target resolves to the correct `.melon/` cache path
- [x] 2.3 Add test that verifies idempotency — calling `Place` twice replaces the existing symlink without error
- [x] 2.4 Add test that verifies a stale directory (from old copy-based placement) is replaced by a symlink
