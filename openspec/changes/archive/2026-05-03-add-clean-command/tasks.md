## 1. Core Implementation

- [x] 1.1 Add a `store.DirName` exported helper (or expose logic) so `clean` can match raw dir names against lock entries without relying on lossy name reconstruction
- [x] 1.2 Create `internal/cli/clean_cmd.go` with `runClean` implementing: load lock → list store → compute orphans → remove cache dirs → remove symlinks
- [x] 1.3 Register `cleanCmd` in `cli.go`'s `Run` function alongside the existing subcommands

## 2. Symlink Removal

- [x] 2.1 In `runClean`, load `melon.yaml` (best-effort); if absent, warn and skip symlink step
- [x] 2.2 Build a `[]lockfile.LockedDep` slice from orphaned entries and call `placer.Unplace` to remove their symlinks

## 3. Edge Cases & Output

- [x] 3.1 If `melon.lock` is absent, print "No melon.lock found. Run 'melon install' first." and return nil
- [x] 3.2 If `.melon/` is absent or empty, print "Nothing to clean." and return nil
- [x] 3.3 Print a per-entry removal line (using `removeStyle`) and a final count summary

## 4. Tests

- [x] 4.1 Write `internal/cli/clean_cmd_test.go` covering: no lock file, nothing to clean, orphaned entry removed, symlink removed alongside cache entry

## 5. Documentation

- [x] 5.1 Update `README.md` to document the `clean` command under the commands reference section
