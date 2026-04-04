## Why

When a dependency is removed from `melon.yml` and `mln install` is run, the old `.melon/` cache entry and its agent directory symlinks are left behind. This leaves stale artifacts on disk that can confuse agents and waste space.

## What Changes

- After resolving and fetching the new dependency set, `mln install` SHALL compute which deps were in the previous lock but are absent from the new one
- For each removed dep, delete its symlink(s) from all agent directories
- For each removed dep, delete its directory from `.melon/`
- Print removed entries in the existing lock diff output (already tracked by `lockfile.Diff`)

## Capabilities

### New Capabilities

- `install-pruning`: `mln install` removes stale `.melon/` cache entries and agent symlinks for dependencies that are no longer in `melon.yml`

### Modified Capabilities

- `skill-placement`: the placement step now also removes symlinks for deps that are no longer in the resolved set (previously only created/replaced links)

## Impact

- `cmd/mln/install_cmd.go`: add pruning step after the new lock is computed
- `internal/placer/placer.go`: extend (or add a sibling function) to handle symlink removal for stale deps
- `internal/store/store.go`: add a `Remove` function to delete a dep's cache directory
- No new dependencies required
