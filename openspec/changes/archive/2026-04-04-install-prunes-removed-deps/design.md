## Context

`mln install` currently only adds or updates: `placer.Place` creates/replaces symlinks for the deps that are in the resolved graph, and fetcher downloads them into `.melon/`. When a dep is removed from `melon.yml`, the old `.melon/<name>@<version>/` directory and all agent symlinks pointing to it are left on disk. The `lockfile.Diff` function already computes `Removed` entries (deps in the old lock but absent from the new one) — this diff is only used for display output today.

## Goals / Non-Goals

**Goals:**
- After computing the new lock, delete the `.melon/` cache directory for every dep in `diff.Removed`
- Remove the agent symlink for every dep in `diff.Removed` from all agent directories (same target derivation as `placer.Place`)
- Keep the diff output that already prints `- <dep>@<version>` for removed entries

**Non-Goals:**
- Pruning deps that were never in the lock but happen to be present in `.melon/` (orphan cleanup is a separate `mln prune` concern)
- Removing cache entries when `--no-place` is set — cache cleanup should still happen; only symlink removal should be skipped

## Decisions

### Use `lockfile.Diff` as the source of truth for what to prune
The diff between `oldLock` (loaded at the start of install) and `newLock` (just computed) already identifies exactly which deps were removed. Reusing this avoids re-deriving the removed set from the manifest or store listing.

_Alternative_: Compare `store.List()` against the new lock. Rejected because it would also catch unrelated orphan entries (e.g. from manual edits or other projects sharing a store), which is out of scope.

### Add `store.Remove(projectDir, dep)` for cache deletion
A single helper in the `store` package that calls `os.RemoveAll` on `store.InstalledPath(projectDir, dep)` keeps deletion symmetric with `InstalledPath` and `List`. `install_cmd.go` calls it for each dep in `diff.Removed`.

### Add `placer.Unplace(deps, m, projectDir)` for symlink removal
Symmetric with `placer.Place` — iterates the same target bases and removes the symlink at `<agent-dir>/<skill-name>`. Missing symlinks are silently ignored (already-absent is not an error). This mirrors the `os.RemoveAll` idiom already used inside `Place` for idempotent replacement.

### Pruning happens after the new lock is written, before placement
Order of operations in `runInstall`:
1. Resolve → fetch → build `newLock`
2. Write `melon.lock`
3. **Prune: unplace + remove store for `diff.Removed`**
4. Place: create/replace symlinks for current deps

This ensures the lock on disk always reflects what's actually in the store and agent dirs by the time install returns.

### `--no-place` suppresses symlink removal but not cache removal
Users who pass `--no-place` intend to skip all filesystem placement. Symlink removal is considered part of placement; cache removal is not. So: always prune the store; only unplace if `!flagNoPlace`.

## Risks / Trade-offs

- **Race condition**: If two `mln install` processes run concurrently they could both try to delete the same entry. Mitigation: `os.RemoveAll` on a non-existent path is a no-op, so the second call is harmless.
- **Symlink pointing into deleted cache**: During the window between unplace and delete (or if a previous run left a dangling symlink), agent tooling could observe a broken link. Mitigation: unplace before store removal so the dangling state is never observable after a successful install.
