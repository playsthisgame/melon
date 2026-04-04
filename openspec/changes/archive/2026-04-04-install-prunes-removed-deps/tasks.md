## 1. Store: add Remove helper

- [x] 1.1 Add `store.Remove(projectDir string, dep resolver.ResolvedDep) error` to `internal/store/store.go` that calls `os.RemoveAll` on `InstalledPath(projectDir, dep)`

## 2. Placer: add Unplace helper

- [x] 2.1 Add `placer.Unplace(deps []lockfile.LockedDep, m manifest.Manifest, projectDir string, out io.Writer) error` to `internal/placer/placer.go` that derives agent target bases the same way `Place` does, then removes the skill symlink for each dep from each target base (silently ignoring not-exist errors)

## 3. Wire pruning into mln install

- [x] 3.1 In `cmd/mln/install_cmd.go`, after writing the new `melon.lock` and before calling `placer.Place`, iterate `diff.Removed` and call `placer.Unplace` (skip if `flagNoPlace`) then `store.Remove` for each removed dep
- [x] 3.2 Confirm the existing `printLockDiff` output already displays `- <dep>@<version>` for removed entries (no change needed if it does)

## 4. Tests

- [x] 4.1 Add a unit test for `store.Remove`: verify the cache directory is deleted and that calling it on a non-existent path returns no error
- [x] 4.2 Add a unit test for `placer.Unplace`: verify the skill symlink is removed from the agent directory and that a missing symlink does not return an error
- [x] 4.3 Add an integration-style test in `install_cmd` or a helper: run install with a manifest containing two deps, then remove one dep from the manifest, run install again, and assert the removed dep's symlink and cache directory are gone while the remaining dep's are intact
