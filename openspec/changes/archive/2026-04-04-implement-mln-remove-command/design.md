## Context

`mln remove` is stubbed in `cmd/mln/main.go` with a no-op handler. The `mln add` command (in `add_cmd.go`) provides the pattern to follow: mutate `melon.yml` then call `runInstall`. The `install-prunes-removed-deps` change landed first, so `runInstall` now automatically calls `placer.Unplace` and `store.Remove` for any dep that drops out of the lock — `remove_cmd.go` does not need to handle cleanup explicitly.

## Goals / Non-Goals

**Goals:**

- Delete the named dependency from `melon.yml`
- Regenerate `melon.lock`, remove the agent symlink, and prune the `.melon/` cache entry by running the install pipeline on the updated manifest
- Print a clear error if the dependency does not exist in `melon.yml`

**Non-Goals:**

- Removing transitive-only dependencies that are no longer needed — a full `mln install` handles the lock correctly
- Interactive confirmation prompts

## Decisions

### Reuse `runInstall` for cleanup, lock regeneration, and placement

After removing the dep from `melon.yml`, call `runInstall` exactly as `runAdd` does. Because `runInstall` now prunes removed deps (symlinks + cache), no explicit cleanup code is needed in `remove_cmd.go`.

_Alternative considered_: Manually unplace and remove the cache entry before calling install. Rejected — `runInstall`'s pruning step already does this based on the diff between old and new lock, so doing it twice would be redundant.

### Error on unknown dep
If `<name>` is not a key in `melon.yml`, return an error immediately rather than silently succeeding. This prevents confusing output when users mistype a dep name.

## Risks / Trade-offs

- **Cleanup tied to install pruning**: Symlink and cache removal happen inside `runInstall` via the lock diff. If the manifest has no remaining deps, `runInstall` exits early before computing a diff — the removed dep's symlink and cache will still be cleaned up because install's early-exit path should also handle pruning. (Verify this edge case in tests.)

## Migration Plan

No migration needed — this is a new command implementation with no schema or data format changes.
